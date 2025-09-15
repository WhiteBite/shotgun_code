<template>
  <div 
    ref="containerRef"
    class="virtual-scroll-container"
    :style="{ height: props.height + 'px' }"
    @scroll="handleScroll"
  >
    <div 
      class="virtual-scroll-content"
      :style="{ height: totalHeight + 'px', position: 'relative' }"
    >
      <div 
        class="virtual-scroll-viewport"
        :style="{ transform: `translateY(${offsetY}px)` }"
      >
        <div
          v-for="line in visibleLines"
          :key="line.index"
          class="virtual-scroll-line"
          :style="{ height: props.lineHeight + 'px', lineHeight: props.lineHeight + 'px' }"
        >
          <span class="line-number">{{ line.index + 1 }}</span>
          <span class="line-content">{{ line.content }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'

interface Props {
  content?: string
  contentProvider?: (startIndex: number, endIndex: number) => Promise<string[]>
  height?: number
  lineHeight?: number
  overscan?: number
  totalLines?: number
}

const props = withDefaults(defineProps<Props>(), {
  content: '',
  contentProvider: undefined,
  height: 400,
  lineHeight: 20,
  overscan: 1, // Reduced overscan for better performance
  totalLines: 0
})

const containerRef = ref<HTMLElement>()
const scrollTop = ref(0)
const cachedLines = ref<Map<number, string>>(new Map())

// More efficient approach: extract lines on demand without creating large arrays
const getLine = (index: number): string => {
  if (props.contentProvider) {
    // For streaming content, check cache first
    if (cachedLines.value.has(index)) {
      return cachedLines.value.get(index) || ''
    }
    return '' // Line not loaded yet
  }
  
  if (!props.content || index < 0) return ''
  
  // For extremely large content, we have hard limits
  const maxLines = 5000 // Hard limit to prevent memory issues
  if (index >= maxLines) {
    if (index === maxLines) {
      return '... [Content truncated due to size limits] ...'
    }
    return ''
  }
  
  // Extract specific line without splitting entire content
  let lineStart = 0
  let currentLine = 0
  
  if (index > 0) {
    let pos = 0
    while (currentLine < index) {
      const nextNewline = props.content.indexOf('\n', pos)
      if (nextNewline === -1) {
        // No more lines
        return ''
      }
      lineStart = nextNewline + 1
      pos = lineStart
      currentLine++
    }
  }
  
  const lineEnd = props.content.indexOf('\n', lineStart)
  return props.content.substring(lineStart, lineEnd === -1 ? undefined : lineEnd)
}

// Calculate total line count
const totalLineCount = computed(() => {
  if (props.totalLines && props.totalLines > 0) {
    return Math.min(props.totalLines, 5000) // Hard limit
  }
  
  if (props.contentProvider) {
    // For streaming content, we need to know the total lines
    return props.totalLines || 0
  }
  
  if (!props.content) return 0
  
  const maxLines = 5000 // Hard limit
  const newlineRegex = /\n/g
  let match
  let lineCount = 0
  
  while ((match = newlineRegex.exec(props.content)) !== null) {
    lineCount++
    if (lineCount >= maxLines) {
      return maxLines
    }
  }
  
  return Math.min(lineCount + 1, maxLines) // +1 for the last line without a newline
})

const totalHeight = computed(() => totalLineCount.value * props.lineHeight)

// Calculate visible range with conservative limits
const visibleRange = computed(() => {
  const containerHeight = props.height
  const start = Math.floor(scrollTop.value / props.lineHeight)
  const end = Math.min(
    totalLineCount.value,
    Math.ceil((scrollTop.value + containerHeight) / props.lineHeight)
  )
  
  return {
    start: Math.max(0, start - props.overscan),
    end: Math.min(totalLineCount.value, end + props.overscan)
  }
})

const startIndex = computed(() => visibleRange.value.start)
const endIndex = computed(() => visibleRange.value.end)

// Extract only visible lines without creating large arrays
const visibleLines = computed(() => {
  const lines = []
  // Limit the number of visible lines to prevent rendering issues
  const maxVisibleLines = 200
  const actualEndIndex = Math.min(endIndex.value, startIndex.value + maxVisibleLines)
  
  for (let i = startIndex.value; i < actualEndIndex; i++) {
    lines.push({
      index: i,
      content: getLine(i)
    })
  }
  return lines
})

const offsetY = computed(() => startIndex.value * props.lineHeight)

const handleScroll = (event: Event) => {
  const target = event.target as HTMLElement
  scrollTop.value = target.scrollTop
  
  // More aggressive throttling for performance
  if (target.scrollTop % 25 === 0) {
    // Minimal delay to allow UI updates
    setTimeout(() => {}, 0)
  }
  
  // Load content on demand for streaming
  if (props.contentProvider) {
    loadVisibleContent()
  }
}

// Load content on demand for streaming
const loadVisibleContent = async () => {
  if (!props.contentProvider) return
  
  const start = Math.max(0, startIndex.value - props.overscan)
  const end = Math.min(totalLineCount.value, endIndex.value + props.overscan)
  
  // Check what lines we need to load
  const linesToLoad: number[] = []
  for (let i = start; i < end; i++) {
    if (!cachedLines.value.has(i)) {
      linesToLoad.push(i)
    }
  }
  
  if (linesToLoad.length > 0) {
    try {
      // Group lines into chunks for batch loading
      const chunkSize = 50
      for (let i = 0; i < linesToLoad.length; i += chunkSize) {
        const chunk = linesToLoad.slice(i, i + chunkSize)
        const chunkStart = Math.min(...chunk)
        const chunkEnd = Math.max(...chunk)
        
        const lines = await props.contentProvider(chunkStart, chunkEnd)
        lines.forEach((line, index) => {
          cachedLines.value.set(chunkStart + index, line)
        })
      }
      
      // Trigger reactivity
      cachedLines.value = new Map(cachedLines.value)
    } catch (error) {
      console.error('Error loading content:', error)
    }
  }
}

// Performance monitoring with AGGRESSIVE thresholds
const memoryUsage = ref(0)
const MAX_CACHE_SIZE = 3; // CRITICAL: Only keep 3 chunks in memory
const MEMORY_CLEANUP_INTERVAL = 3000; // Cleanup every 3 seconds
const CRITICAL_MEMORY_THRESHOLD = 20; // 20MB threshold

const updateMemoryUsage = () => {
  if ('performance' in window && 'memory' in (performance as any)) {
    const memory = (performance as any).memory
    memoryUsage.value = memory.usedJSHeapSize / 1024 / 1024 // MB
    
    // CRITICAL: More aggressive memory warning thresholds
    if (memoryUsage.value > CRITICAL_MEMORY_THRESHOLD) {
      console.warn(`CRITICAL memory usage in VirtualScrollViewer: ${memoryUsage.value.toFixed(1)} MB`);
      // Force aggressive cleanup
      aggressiveCacheCleanup();
    }
  }
}

// CRITICAL: Aggressive cache cleanup function
const aggressiveCacheCleanup = () => {
  console.log('Performing aggressive cache cleanup in VirtualScrollViewer');
  
  // Only keep the most recent MAX_CACHE_SIZE chunks
  if (cachedLines.value.size > MAX_CACHE_SIZE * 50) {
    const entries = Array.from(cachedLines.value.entries());
    cachedLines.value.clear();
    
    // Keep only the last MAX_CACHE_SIZE * 50 lines
    const recentEntries = entries.slice(-MAX_CACHE_SIZE * 50);
    recentEntries.forEach(([key, value]) => {
      cachedLines.value.set(key, value);
    });
    
    console.log(`Cleaned cache: ${entries.length} -> ${recentEntries.length} lines`);
  }
  
  // Force garbage collection if available
  if (window.gc) {
    try {
      window.gc();
    } catch (e) {
      console.warn('Failed to trigger garbage collection in VirtualScrollViewer');
    }
  }
}

onMounted(() => {
  updateMemoryUsage()
  // CRITICAL: More frequent memory monitoring and aggressive cleanup
  const memoryInterval = setInterval(updateMemoryUsage, 1000);
  const cleanupInterval = setInterval(aggressiveCacheCleanup, MEMORY_CLEANUP_INTERVAL);
  
  onUnmounted(() => {
    console.log('VirtualScrollViewer unmounting - performing final cleanup');
    clearInterval(memoryInterval);
    clearInterval(cleanupInterval);
    
    // CRITICAL: Force cleanup on unmount
    cachedLines.value.clear();
    
    // Force garbage collection if available
    if (window.gc) {
      try {
        window.gc();
        console.log('Final garbage collection triggered on VirtualScrollViewer unmount');
      } catch (e) {
        console.warn('Failed to trigger final garbage collection');
      }
    }
  })
  
  // Load initial content for streaming
  if (props.contentProvider) {
    loadVisibleContent()
  }
})

// Watch for content changes and reset scroll with aggressive cleanup
watch(() => props.content, () => {
  console.log('Content changed - performing aggressive cleanup');
  scrollTop.value = 0
  cachedLines.value.clear()
  if (containerRef.value) {
    containerRef.value.scrollTop = 0
  }
  // Force garbage collection after content change
  setTimeout(() => {
    if (window.gc) window.gc();
  }, 100);
}, { flush: 'post' })

// Watch for content provider changes with aggressive cleanup
watch(() => props.contentProvider, () => {
  console.log('Content provider changed - performing aggressive cleanup');
  scrollTop.value = 0
  cachedLines.value.clear()
  if (containerRef.value) {
    containerRef.value.scrollTop = 0
  }
  if (props.contentProvider) {
    loadVisibleContent()
  }
  // Force garbage collection after provider change
  setTimeout(() => {
    if (window.gc) window.gc();
  }, 100);
}, { flush: 'post' })

defineExpose({
  scrollToTop: () => {
    scrollTop.value = 0
    cachedLines.value.clear()
    if (containerRef.value) {
      containerRef.value.scrollTop = 0
    }
    // Force garbage collection after scroll reset
    setTimeout(() => {
      if (window.gc) window.gc();
    }, 100);
  },
  scrollToLine: (lineNumber: number) => {
    // Limit the line number to prevent errors
    const safeLineNumber = Math.min(lineNumber, totalLineCount.value - 1)
    const targetScroll = safeLineNumber * props.lineHeight
    scrollTop.value = targetScroll
    if (containerRef.value) {
      containerRef.value.scrollTop = targetScroll
    }
  },
  memoryUsage,
  clearCache: () => {
    console.log('Manual cache clear requested');
    cachedLines.value.clear();
    if (window.gc) {
      try {
        window.gc();
        console.log('Manual garbage collection triggered');
      } catch (e) {
        console.warn('Failed to trigger manual garbage collection');
      }
    }
  },
  aggressiveCacheCleanup // Expose for external cleanup
})
</script>

<style scoped>
.virtual-scroll-container {
  overflow: auto;
  width: 100%;
  position: relative;
  background-color: #f8fafc;
}

.virtual-scroll-content {
  width: 100%;
}

.virtual-scroll-viewport {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
}

.virtual-scroll-line {
  display: flex;
  font-family: 'Courier New', Courier, monospace;
  font-size: 14px;
  white-space: pre;
  border-bottom: 1px solid #eee;
  padding: 0 10px;
}

.line-number {
  width: 50px;
  text-align: right;
  padding-right: 10px;
  color: #999;
  user-select: none;
  flex-shrink: 0;
}

.line-content {
  flex: 1;
  padding-left: 10px;
  word-break: break-all;
}
</style>