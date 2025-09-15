<template>
  <div class="streaming-context-viewer">
    <!-- Memory usage indicator and streaming info -->
    <div class="streaming-header">
      <div v-if="streamingContext" class="streaming-info">
        <span class="status-indicator" :class="statusClass">
          {{ streamingContext.status }}
        </span>
        <span class="file-count">{{ streamingContext.files.length }} files</span>
        <span class="size-info">{{ formatSize(streamingContext.totalCharacters) }}</span>
        <span v-if="streamingProgress" class="progress-info">
          {{ streamingProgress.filesProcessed }}/{{ streamingProgress.totalFiles }} processed
        </span>
      </div>
      
      <div 
        v-if="showMemoryInfo && memoryUsage > 0" 
        :class="['memory-indicator', memoryStatusClass]"
        :title="memoryTooltip"
      >
        Memory: {{ formatMemoryUsage }} MB
      </div>
    </div>
    
    <!-- Streaming content display - MEMORY OPTIMIZED -->
    <div class="content-container">
      <!-- Loading indicator -->
      <div v-if="isLoading" class="loading-indicator">
        <div class="spinner"></div>
        <p>{{ progressMessage || 'Loading context...' }}</p>
        <div v-if="streamingProgress" class="progress-details">
          <div class="progress-bar">
            <div 
              class="progress-fill" 
              :style="{ width: progressPercentage + '%' }"
            ></div>
          </div>
          <p class="progress-text">{{ progressText }}</p>
        </div>
      </div>
      
      <!-- Error display -->
      <div v-else-if="error" class="error-display">
        <p>Error: {{ error }}</p>
        <button class="retry-button" @click="retryLoad">Retry</button>
      </div>
      
      <!-- Paginated streaming content -->
      <div v-else-if="currentChunk" class="streaming-viewer">
        <!-- Stream controls -->
        <div class="stream-controls">
          <button 
            :disabled="currentPage <= 1" 
            class="nav-button"
            @click="loadPreviousChunk"
          >
            ← Previous
          </button>
          
          <div class="stream-info">
            <span class="page-info">
              Chunk {{ currentPage }} of {{ totalPages }}
            </span>
            <span class="line-info">
              Lines {{ currentChunk.startLine + 1 }}-{{ currentChunk.endLine + 1 }}
            </span>
          </div>
          
          <button 
            :disabled="!hasMore" 
            class="nav-button"
            @click="loadNextChunk"
          >
            Next →
          </button>
        </div>
        
        <!-- Streaming content with line numbers -->
        <pre class="stream-content">
          <code 
            v-for="(line, index) in currentChunk.lines" 
            :key="currentChunk.startLine + index"
            class="stream-line"
            :data-line="currentChunk.startLine + index + 1"
          >{{ line }}</code>
        </pre>
        
        <!-- Stream progress indicator -->
        <div class="stream-progress">
          <div 
            class="stream-progress-fill" 
            :style="{ width: progress + '%' }"
          ></div>
        </div>
      </div>
      
      <!-- Empty state -->
      <div v-else class="empty-state">
        <p>No streaming context available</p>
        <p class="text-sm text-gray-500">Start streaming to view content</p>
      </div>
    </div>
    
    <!-- Controls -->
    <div class="viewer-controls">
      <div class="left-controls">
        <button 
          v-if="!isStreaming && !streamingContext" 
          :disabled="isLoading || !contextId"
          class="control-button start-button"
          @click="startStreaming"
        >
          {{ isLoading ? 'Starting...' : 'Start Streaming' }}
        </button>
        
        <button 
          v-if="isStreaming" 
          class="control-button stop-button"
          @click="stopStreaming"
        >
          Stop Streaming
        </button>
        
        <button 
          v-if="currentChunk" 
          class="control-button"
          @click="searchInStream"
        >
          Search
        </button>
      </div>
      
      <div class="right-controls">
        <button 
          v-if="currentChunk" 
          :disabled="isLoading"
          class="control-button copy-button"
          @click="copyCurrentChunk"
        >
          Copy Chunk
        </button>
        
        <button 
          v-if="streamingContext" 
          :disabled="isLoading"
          class="control-button"
          @click="downloadStream"
        >
          Download
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { usePaginatedContext } from '@/composables/usePaginatedContext'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import type { StreamingContext, StreamingProgress } from '@/domain/entities/ContextSummary'

// MEMORY OPTIMIZED: StreamingContextViewer using paginated approach
interface Props {
  contextId?: string
  containerHeight?: number
  linesPerPage?: number
  showMemoryInfo?: boolean
  autoStart?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  contextId: undefined,
  containerHeight: 400,
  linesPerPage: 150, // Slightly larger chunks for streaming
  showMemoryInfo: true,
  autoStart: false
})

const emit = defineEmits<{
  (e: 'copy'): void
  (e: 'download'): void
  (e: 'error', error: string): void
  (e: 'stream-started'): void
  (e: 'stream-stopped'): void
}>()

const contextStore = useContextBuilderStore()

// Use paginated context for memory safety
const {
  currentChunk,
  isLoading: isPaginating,
  // error: paginationError,
  currentPage,
  totalPages,
  progress,
  hasMore,
  loadNextChunk,
  loadPreviousChunk,
  searchInChunk
} = usePaginatedContext(props.contextId, props.linesPerPage)

// Streaming-specific state
const isStreaming = ref(false)
const streamingContext = ref<StreamingContext | null>(null)
const streamingProgress = ref<StreamingProgress | null>(null)
const progressMessage = ref('')
const error = ref<string | null>(null)
const memoryUsage = ref(0)

// Computed properties
const isLoading = computed(() => isPaginating.value || (isStreaming.value && !currentChunk.value))

const statusClass = computed(() => {
  if (!streamingContext.value) return ''
  switch (streamingContext.value.status) {
    case 'streaming': return 'status-streaming'
    case 'ready': return 'status-ready'
    case 'error': return 'status-error'
    default: return 'status-idle'
  }
})

const progressPercentage = computed(() => {
  if (!streamingProgress.value) return 0
  return Math.round((streamingProgress.value.filesProcessed / streamingProgress.value.totalFiles) * 100)
})

const progressText = computed(() => {
  if (!streamingProgress.value) return ''
  const { filesProcessed, totalFiles, bytesProcessed, totalBytes } = streamingProgress.value
  const mbProcessed = (bytesProcessed / (1024 * 1024)).toFixed(1)
  const mbTotal = (totalBytes / (1024 * 1024)).toFixed(1)
  return `${filesProcessed}/${totalFiles} files, ${mbProcessed}/${mbTotal} MB`
})

const formatMemoryUsage = computed(() => memoryUsage.value.toFixed(1))
const memoryStatusClass = computed(() => {
  if (memoryUsage.value > 60) return 'bg-red-600 text-white'
  if (memoryUsage.value > 40) return 'bg-orange-600 text-white'
  if (memoryUsage.value > 25) return 'bg-yellow-600 text-white'
  return 'bg-green-600 text-white'
})

const memoryTooltip = computed(() => {
  return `Memory usage: ${formatMemoryUsage.value} MB`
})

// Methods
const formatSize = (bytes: number): string => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

const startStreaming = async () => {
  if (!props.contextId) {
    error.value = 'No context ID provided'
    return
  }
  
  try {
    error.value = null
    progressMessage.value = 'Initializing streaming context...'
    
    // Create streaming context via store
    const stream = await contextStore.createStreamingContext(
      contextStore.currentProject?.path || '',
      contextStore.selectedFiles || []
    )
    
    streamingContext.value = stream
    isStreaming.value = true
    
    emit('stream-started')
    progressMessage.value = ''
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'Failed to start streaming'
    error.value = errorMessage
    progressMessage.value = ''
    emit('error', errorMessage)
  }
}

const stopStreaming = async () => {
  try {
    if (streamingContext.value) {
      await contextStore.closeStreamingContext()
    }
    
    isStreaming.value = false
    streamingContext.value = null
    streamingProgress.value = null
    
    emit('stream-stopped')
  } catch (err) {
    console.error('Error stopping stream:', err)
  }
}

const retryLoad = async () => {
  error.value = null
  if (isStreaming.value) {
    await stopStreaming()
  }
  await startStreaming()
}

const copyCurrentChunk = async () => {
  if (!currentChunk.value) return
  
  try {
    const content = currentChunk.value.lines.join('\n')
    await navigator.clipboard.writeText(content)
    emit('copy')
  } catch (error) {
    console.error('Failed to copy chunk:', error)
  }
}

const downloadStream = () => {
  emit('download')
}

const searchInStream = () => {
  // Implement search functionality for streaming context
  if (currentChunk.value) {
    const query = prompt('Search in current chunk:')
    if (query) {
      const results = searchInChunk(query)
      if (results.length > 0) {
        alert(`Found ${results.length} matches in current chunk`)
      } else {
        alert('No matches found in current chunk')
      }
    }
  }
}

// Memory monitoring
const updateMemoryUsage = () => {
  if ('performance' in window && 'memory' in (performance as unknown)) {
    const memory = (performance as unknown).memory
    memoryUsage.value = memory.usedJSHeapSize / 1024 / 1024 // MB
  }
}

// Watch for streaming context changes
watch(() => contextStore.streamingContext, (newContext) => {
  if (newContext) {
    streamingContext.value = newContext
    isStreaming.value = true
  }
}, { immediate: true })

// Auto-start streaming if requested
watch(() => props.contextId, (newContextId) => {
  if (newContextId && props.autoStart) {
    startStreaming()
  }
}, { immediate: true })

// Lifecycle
let memoryMonitorInterval: number | null = null

onMounted(() => {
  updateMemoryUsage()
  memoryMonitorInterval = window.setInterval(updateMemoryUsage, 3000)
})

onUnmounted(() => {
  if (memoryMonitorInterval) {
    clearInterval(memoryMonitorInterval)
  }
  if (isStreaming.value) {
    stopStreaming()
  }
})

defineExpose({
  startStreaming,
  stopStreaming,
  currentChunk: computed(() => currentChunk.value),
  isStreaming: computed(() => isStreaming.value),
  streamingContext: computed(() => streamingContext.value)
})
</script>

<style scoped>
.streaming-context-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  overflow: hidden;
  background-color: #ffffff;
}

.streaming-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem;
  background-color: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.streaming-info {
  display: flex;
  gap: 1rem;
  align-items: center;
  font-size: 0.875rem;
}

.status-indicator {
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.status-streaming {
  background-color: #dbeafe;
  color: #1d4ed8;
}

.status-ready {
  background-color: #dcfce7;
  color: #166534;
}

.status-error {
  background-color: #fee2e2;
  color: #dc2626;
}

.status-idle {
  background-color: #f1f5f9;
  color: #64748b;
}

.file-count, .size-info, .progress-info {
  color: #64748b;
  font-weight: 500;
}

.memory-indicator {
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 0.25rem;
}

.content-container {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 2rem;
  color: #64748b;
}

.spinner {
  width: 2rem;
  height: 2rem;
  border: 2px solid #e2e8f0;
  border-top: 2px solid #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.progress-details {
  width: 100%;
  max-width: 300px;
  margin-top: 1rem;
}

.progress-bar {
  width: 100%;
  height: 4px;
  background-color: #e2e8f0;
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 0.5rem;
}

.progress-fill {
  height: 100%;
  background-color: #3b82f6;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 0.75rem;
  text-align: center;
  color: #64748b;
}

.error-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #dc2626;
  padding: 2rem;
  text-align: center;
}

.retry-button {
  margin-top: 1rem;
  padding: 0.5rem 1rem;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.retry-button:hover {
  background-color: #2563eb;
}

.streaming-viewer {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.stream-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background-color: #f1f5f9;
  border-bottom: 1px solid #e2e8f0;
  font-size: 0.875rem;
}

.stream-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
}

.page-info {
  font-weight: 600;
  color: #374151;
}

.line-info {
  font-size: 0.75rem;
  color: #6b7280;
}

.nav-button {
  padding: 0.375rem 0.75rem;
  background-color: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
}

.nav-button:hover:not(:disabled) {
  background-color: #f8fafc;
  border-color: #94a3b8;
}

.nav-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.stream-content {
  flex: 1;
  margin: 0;
  padding: 1rem;
  overflow: auto;
  white-space: pre;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
  background-color: #fafafa;
  color: #1e293b;
}

.stream-line {
  display: block;
  position: relative;
  padding-left: 3rem;
}

.stream-line::before {
  content: attr(data-line);
  position: absolute;
  left: 0;
  width: 2.5rem;
  text-align: right;
  color: #94a3b8;
  font-size: 0.75rem;
  user-select: none;
}

.stream-progress {
  height: 2px;
  background-color: #e2e8f0;
  overflow: hidden;
}

.stream-progress-fill {
  height: 100%;
  background-color: #10b981;
  transition: width 0.3s ease;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #94a3b8;
  padding: 2rem;
}

.viewer-controls {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem;
  border-top: 1px solid #e2e8f0;
  background-color: #f8fafc;
}

.left-controls, .right-controls {
  display: flex;
  gap: 0.5rem;
}

.control-button {
  padding: 0.375rem 0.75rem;
  font-size: 0.875rem;
  border-radius: 0.25rem;
  border: 1px solid #cbd5e1;
  background-color: #ffffff;
  cursor: pointer;
  transition: all 0.2s;
}

.control-button:hover:not(:disabled) {
  background-color: #f8fafc;
  border-color: #94a3b8;
}

.control-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.start-button {
  background-color: #10b981;
  color: white;
  border-color: #10b981;
}

.start-button:hover:not(:disabled) {
  background-color: #059669;
  border-color: #059669;
}

.stop-button {
  background-color: #ef4444;
  color: white;
  border-color: #ef4444;
}

.stop-button:hover:not(:disabled) {
  background-color: #dc2626;
  border-color: #dc2626;
}

.copy-button {
  background-color: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.copy-button:hover:not(:disabled) {
  background-color: #2563eb;
  border-color: #2563eb;
}
</style>