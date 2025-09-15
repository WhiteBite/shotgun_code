<template>
  <div class="split-pane" :class="`split-pane-${direction}`">
    <div 
      ref="leftPanel" 
      class="split-pane-panel split-pane-left"
      :style="leftPanelStyle"
    >
      <slot name="left" />
    </div>
    
    <div 
      class="split-pane-divider" 
      :class="{ 'divider-resizing': isResizing }"
      @mousedown="startResize"
    >
      <div class="divider-handle"></div>
    </div>
    
    <div 
      ref="rightPanel" 
      class="split-pane-panel split-pane-right"
      :style="rightPanelStyle"
    >
      <slot name="right" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

// Props
interface Props {
  initialRatio?: number
  direction?: 'horizontal' | 'vertical'
  minSize?: number
  storageKey?: string
}

const props = withDefaults(defineProps<Props>(), {
  initialRatio: 0.5,
  direction: 'horizontal',
  minSize: 100,
  storageKey: ''
})

// Emits
const emit = defineEmits<{
  resize: [ratio: number]
}>()

// Refs
const leftPanel = ref<HTMLElement>()
const rightPanel = ref<HTMLElement>()
const container = ref<HTMLElement>()

// State
const ratio = ref(props.initialRatio)
const isResizing = ref(false)

// Computed
const leftPanelStyle = computed(() => {
  if (props.direction === 'vertical') {
    return {
      height: `${ratio.value * 100}%`
    }
  }
  return {
    width: `${ratio.value * 100}%`
  }
})

const rightPanelStyle = computed(() => {
  if (props.direction === 'vertical') {
    return {
      height: `${(1 - ratio.value) * 100}%`
    }
  }
  return {
    width: `${(1 - ratio.value) * 100}%`
  }
})

// Methods
function startResize(event: MouseEvent) {
  event.preventDefault()
  isResizing.value = true
  
  const startX = event.clientX
  const startY = event.clientY
  const startRatio = ratio.value
  
  const containerRect = container.value?.getBoundingClientRect()
  if (!containerRect) return
  
  const onMouseMove = (e: MouseEvent) => {
    if (!containerRect) return
    
    if (props.direction === 'vertical') {
      const deltaY = e.clientY - startY
      const containerHeight = containerRect.height
      const newRatio = Math.max(
        props.minSize / containerHeight,
        Math.min(1 - props.minSize / containerHeight, startRatio + deltaY / containerHeight)
      )
      ratio.value = newRatio
    } else {
      const deltaX = e.clientX - startX
      const containerWidth = containerRect.width
      const newRatio = Math.max(
        props.minSize / containerWidth,
        Math.min(1 - props.minSize / containerWidth, startRatio + deltaX / containerWidth)
      )
      ratio.value = newRatio
    }
    
    emit('resize', ratio.value)
  }
  
  const onMouseUp = () => {
    isResizing.value = false
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
    
    // Save to localStorage if storageKey is provided
    if (props.storageKey) {
      localStorage.setItem(props.storageKey, ratio.value.toString())
    }
  }
  
  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

// Load saved ratio from localStorage
onMounted(() => {
  if (props.storageKey) {
    const savedRatio = localStorage.getItem(props.storageKey)
    if (savedRatio) {
      const parsedRatio = parseFloat(savedRatio)
      if (!isNaN(parsedRatio) && parsedRatio > 0 && parsedRatio < 1) {
        ratio.value = parsedRatio
      }
    }
  }
})
</script>

<style scoped>
.split-pane {
  display: flex;
  height: 100%;
  width: 100%;
  position: relative;
}

.split-pane-horizontal {
  flex-direction: row;
}

.split-pane-vertical {
  flex-direction: column;
}

.split-pane-panel {
  overflow: hidden;
  position: relative;
}

.split-pane-left {
  flex: 0 0 auto;
}

.split-pane-right {
  flex: 1 1 auto;
}

.split-pane-divider {
  position: relative;
  background: rgba(148, 163, 184, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s ease;
  z-index: 10;
}

.split-pane-horizontal .split-pane-divider {
  width: 4px;
  cursor: col-resize;
}

.split-pane-vertical .split-pane-divider {
  height: 4px;
  cursor: row-resize;
}

.split-pane-divider:hover {
  background: rgba(59, 130, 246, 0.5);
}

.divider-resizing {
  background: rgba(59, 130, 246, 0.8) !important;
  z-index: 1000;
}

.divider-handle {
  background: rgba(203, 213, 225, 0.8);
  border-radius: 2px;
  transition: background-color 0.2s ease;
}

.split-pane-horizontal .divider-handle {
  width: 2px;
  height: 40px;
}

.split-pane-vertical .divider-handle {
  width: 40px;
  height: 2px;
}

.split-pane-divider:hover .divider-handle {
  background: white;
}
</style>