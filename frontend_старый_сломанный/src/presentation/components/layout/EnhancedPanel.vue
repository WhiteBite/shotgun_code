<template>
  <div
    ref="panelRef"
    class="enhanced-panel"
    :class="panelClasses"
    :style="panelStyles"
    @mousedown="handlePanelMouseDown"
    @contextmenu="handleContextMenu"
  >
    <!-- Panel Header -->
    <div
      v-if="showHeader"
      ref="headerRef"
      class="panel-header"
      :class="headerClasses"
      @mousedown="handleHeaderMouseDown"
      @dblclick="handleHeaderDoubleClick"
    >
      <!-- Drag Handle -->
      <div 
        v-if="draggable && !isMaximized"
        v-tooltip="'Drag to move panel'"
        class="drag-handle"
      >
        <Bars3Icon class="w-4 h-4" />
      </div>

      <!-- Panel Icon -->
      <div v-if="icon" class="panel-icon">
        <component :is="icon" class="w-4 h-4" />
      </div>

      <!-- Panel Title -->
      <div class="panel-title">
        <h3 v-if="title" class="title-text">{{ title }}</h3>
        <span v-if="subtitle" class="subtitle-text">{{ subtitle }}</span>
      </div>

      <!-- Badge/Status -->
      <div v-if="badge" class="panel-badge">
        <span class="badge-text">{{ badge }}</span>
      </div>

      <!-- Header Actions -->
      <div class="panel-actions">
        <slot name="header-actions" />
        
        <!-- Built-in Actions -->
        <button
          v-if="collapsible"
          v-tooltip="isCollapsed ? 'Expand panel' : 'Collapse panel'"
          class="action-btn"
          @click="toggleCollapse"
        >
          <ChevronUpIcon v-if="!isCollapsed" class="w-4 h-4" />
          <ChevronDownIcon v-else class="w-4 h-4" />
        </button>

        <button
          v-if="maximizable"
          v-tooltip="isMaximized ? 'Restore panel' : 'Maximize panel'"
          class="action-btn"
          @click="toggleMaximize"
        >
          <ArrowsPointingInIcon v-if="isMaximized" class="w-4 h-4" />
          <ArrowsPointingOutIcon v-else class="w-4 h-4" />
        </button>

        <button
          v-if="closable"
          v-tooltip="'Close panel'"
          class="action-btn close-btn"
          @click="handleClose"
        >
          <XMarkIcon class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Panel Content -->
    <div
      ref="contentRef"
      class="panel-content"
      :class="contentClasses"
      :style="contentStyles"
    >
      <!-- Loading Overlay -->
      <div v-if="loading" class="panel-loading">
        <div class="loading-spinner">
          <div class="spinner" />
        </div>
        <p class="loading-text">{{ loadingText || 'Loading...' }}</p>
      </div>

      <!-- Content Slot -->
      <div v-else class="content-wrapper" :class="{ 'collapsed': isCollapsed }">
        <slot />
      </div>

      <!-- Error State -->
      <div v-if="error" class="panel-error">
        <ExclamationTriangleIcon class="w-6 h-6 text-red-400" />
        <p class="error-text">{{ error }}</p>
        <button v-if="retryable" class="retry-btn" @click="$emit('retry')">
          Retry
        </button>
      </div>
    </div>

    <!-- Resize Handles -->
    <template v-if="resizable && !isMaximized">
      <div
        v-for="handle in activeResizeHandles"
        :key="handle"
        v-tooltip="`Resize ${handle}`"
        :class="`resize-handle resize-${handle}`"
        @mousedown="startResize($event, handle)"
      />
    </template>

    <!-- Dock Indicator -->
    <div
      v-if="showDockIndicator"
      class="dock-indicator"
      :class="`dock-${dockPreview}`"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useElementBounding, useWindowSize } from '@vueuse/core'
import {
  Bars3Icon,
  ChevronUpIcon,
  ChevronDownIcon,
  ArrowsPointingInIcon,
  ArrowsPointingOutIcon,
  XMarkIcon,
  ExclamationTriangleIcon
} from '@/presentation/components/icons'
import { APP_CONFIG } from '@/config/app-config'
import { useEnhancedPanelService } from '@/composables/useEnhancedPanelService'

// Types
export type PanelPosition = {
  x: number
  y: number
  width: number
  height: number
}

export type PanelState = 'normal' | 'minimized' | 'maximized' | 'docked'
export type DockPosition = 'left' | 'right' | 'top' | 'bottom' | 'center'
export type ResizeHandle = 'n' | 's' | 'e' | 'w' | 'ne' | 'nw' | 'se' | 'sw'

// Props
interface Props {
  // Basic properties
  title?: string
  subtitle?: string
  icon?: any
  badge?: string | number
  
  // State management
  modelValue?: PanelPosition
  state?: PanelState
  zIndex?: number
  
  // Capabilities
  draggable?: boolean
  resizable?: boolean
  collapsible?: boolean
  maximizable?: boolean
  closable?: boolean
  
  // Resize configuration
  minWidth?: number
  minHeight?: number
  maxWidth?: number
  maxHeight?: number
  aspectRatio?: number
  
  // Appearance
  variant?: 'default' | 'glass' | 'solid' | 'outlined'
  size?: 'sm' | 'md' | 'lg'
  showHeader?: boolean
  
  // Content state
  loading?: boolean
  loadingText?: string
  error?: string
  retryable?: boolean
  
  // Advanced features
  snapToGrid?: boolean
  gridSize?: number
  magneticEdges?: boolean
  dockable?: boolean
  
  // Styling
  className?: string
  headerClass?: string
  contentClass?: string
}

const props = withDefaults(defineProps<Props>(), {
  state: 'normal',
  zIndex: 100,
  draggable: true,
  resizable: true,
  collapsible: false,
  maximizable: true,
  closable: true,
  minWidth: APP_CONFIG.ui.panels.MIN_WIDTH,
  minHeight: APP_CONFIG.ui.panels.MIN_HEIGHT,
  variant: 'default',
  size: 'md',
  showHeader: true,
  snapToGrid: false,
  gridSize: 10,
  magneticEdges: true,
  dockable: true
})

// Emits
interface Emits {
  'update:modelValue': [position: PanelPosition]
  'update:state': [state: PanelState]
  'move': [position: { x: number; y: number }]
  'resize': [size: { width: number; height: number }]
  'close': []
  'maximize': []
  'minimize': []
  'restore': []
  'dock': [position: DockPosition]
  'undock': []
  'retry': []
}

const emit = defineEmits<Emits>()

// Use the service
const { enhancedPanelService } = useEnhancedPanelService()

// Template refs
const panelRef = ref<HTMLElement>()
const headerRef = ref<HTMLElement>()
const contentRef = ref<HTMLElement>()

// Position and size tracking
const { width: windowWidth, height: windowHeight } = useWindowSize()

// Computed properties
const isMaximized = computed(() => props.state === 'maximized')
const isMinimized = computed(() => props.state === 'minimized')
const isDocked = computed(() => props.state === 'docked')

const isCollapsed = computed(() => enhancedPanelService.getIsCollapsed())
const isDragging = computed(() => enhancedPanelService.getIsDragging())
const isResizing = computed(() => enhancedPanelService.getIsResizing())
const resizeHandle = computed(() => enhancedPanelService.getResizeHandle())
const showDockIndicator = computed(() => enhancedPanelService.getShowDockIndicator())
const dockPreview = computed(() => enhancedPanelService.getDockPreview())

const panelClasses = computed(() => enhancedPanelService.getPanelClasses(props, isMaximized.value, isMinimized.value, isDocked.value))
const headerClasses = computed(() => enhancedPanelService.getHeaderClasses(props, isMaximized.value))
const contentClasses = computed(() => enhancedPanelService.getContentClasses(props))
const panelStyles = computed(() => enhancedPanelService.getPanelStyles(props, isMaximized.value))
const contentStyles = computed(() => enhancedPanelService.getContentStyles())
const activeResizeHandles = computed(() => enhancedPanelService.getActiveResizeHandles(props, isMaximized.value))

// Methods
const toggleCollapse = () => {
  enhancedPanelService.toggleCollapse()
}

const toggleMaximize = () => {
  const newState = isMaximized.value ? 'normal' : 'maximized'
  emit('update:state', newState)
  
  if (newState === 'maximized') {
    emit('maximize')
  } else {
    emit('restore')
  }
}

const handleClose = () => {
  emit('close')
}

// Drag functionality
const handleHeaderMouseDown = (event: MouseEvent) => {
  if (!props.draggable || isMaximized.value) return
  startDrag(event)
}

const handlePanelMouseDown = (event: MouseEvent) => {
  // Bring panel to front
  if (panelRef.value) {
    panelRef.value.style.zIndex = String(props.zIndex + 1000)
  }
}

const startDrag = (event: MouseEvent) => {
  if (!props.modelValue) return
  
  enhancedPanelService.setIsDragging(true)
  const startX = event.clientX - props.modelValue.x
  const startY = event.clientY - props.modelValue.y

  const handleMouseMove = (e: MouseEvent) => {
    if (!props.modelValue) return

    let newX = e.clientX - startX
    let newY = e.clientY - startY

    // Snap to grid
    if (props.snapToGrid) {
      newX = Math.round(newX / props.gridSize) * props.gridSize
      newY = Math.round(newY / props.gridSize) * props.gridSize
    }

    // Magnetic edges
    if (props.magneticEdges) {
      const margin = APP_CONFIG.ui.panels.RESIZE_INDICATOR_HEIGHT
      if (Math.abs(newX) < margin) newX = 0
      if (Math.abs(newY) < margin) newY = 0
      if (Math.abs(newX + props.modelValue.width - windowWidth.value) < margin) {
        newX = windowWidth.value - props.modelValue.width
      }
      if (Math.abs(newY + props.modelValue.height - windowHeight.value) < margin) {
        newY = windowHeight.value - props.modelValue.height
      }
    }

    // Update position
    const newPosition: PanelPosition = {
      ...props.modelValue,
      x: newX,
      y: newY
    }

    emit('update:modelValue', newPosition)
    emit('move', { x: newX, y: newY })

    // Show dock preview
    if (props.dockable) {
      enhancedPanelService.updateDockPreview(e.clientX, e.clientY, windowWidth.value, windowHeight.value)
    }
  }

  const handleMouseUp = () => {
    enhancedPanelService.setIsDragging(false)
    enhancedPanelService.setShowDockIndicator(false)
    
    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseup', handleMouseUp)
    
    // Check for docking
    if (props.dockable && dockPreview.value) {
      emit('dock', dockPreview.value)
      enhancedPanelService.setDockPreview(undefined)
    }
  }

  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
}

// Resize functionality
const startResize = (event: MouseEvent, handle: ResizeHandle) => {
  if (!props.modelValue) return
  
  event.stopPropagation()
  enhancedPanelService.setIsResizing(true)
  enhancedPanelService.setResizeHandle(handle)
  
  const startPos = props.modelValue
  const startMouse = { x: event.clientX, y: event.clientY }

  const handleMouseMove = (e: MouseEvent) => {
    if (!props.modelValue) return

    const deltaX = e.clientX - startMouse.x
    const deltaY = e.clientY - startMouse.y

    const newPosition = enhancedPanelService.calculateNewSize(startPos, deltaX, deltaY, handle, props)
    
    emit('update:modelValue', newPosition)
    emit('resize', { width: newPosition.width, height: newPosition.height })
  }

  const handleMouseUp = () => {
    enhancedPanelService.setIsResizing(false)
    enhancedPanelService.setResizeHandle(undefined)
    
    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseup', handleMouseUp)
  }

  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
}

const handleHeaderDoubleClick = () => {
  if (props.maximizable) {
    toggleMaximize()
  }
}

const handleContextMenu = (event: MouseEvent) => {
  event.preventDefault()
  // Could emit context menu event for custom menu
}

// Lifecycle
onMounted(() => {
  // Set initial focus
  if (panelRef.value) {
    panelRef.value.focus()
  }
})
</script>

<style scoped>
.enhanced-panel {
  @apply absolute bg-gray-800 border border-gray-600 rounded-lg shadow-xl;
  @apply transition-all duration-200 ease-in-out;
  @apply focus:outline-none focus:ring-2 focus:ring-blue-500/50;
  min-width: v-bind('APP_CONFIG.ui.panels.MIN_WIDTH + "px"');
  min-height: v-bind('APP_CONFIG.ui.panels.MIN_HEIGHT + "px"');
}

/* Panel variants */
.panel-glass {
  @apply bg-gray-800/80 backdrop-blur-md border-gray-500/50;
}

.panel-solid {
  @apply bg-gray-800 border-gray-600;
}

.panel-outlined {
  @apply bg-transparent border-2 border-gray-500;
}

/* Panel sizes */
.panel-sm {
  @apply text-sm;
}

.panel-lg {
  @apply text-lg;
}

/* Panel states */
.panel-maximized {
  @apply fixed inset-0 rounded-none;
  z-index: 1000;
}

.panel-minimized {
  @apply h-auto;
}

.panel-docked {
  @apply rounded-none;
}

/* Header */
.panel-header {
  @apply flex items-center px-4 py-2 border-b border-gray-600;
  @apply bg-gray-800 rounded-t-lg;
  user-select: none;
}

.panel-header.draggable {
  @apply cursor-move;
}

.drag-handle {
  @apply text-gray-400 mr-2 cursor-move;
}

.panel-icon {
  @apply text-gray-300 mr-2;
}

.panel-title {
  @apply flex-1 min-w-0;
}

.title-text {
  @apply font-medium text-gray-200 truncate;
}

.subtitle-text {
  @apply text-xs text-gray-400 block truncate;
}

.panel-badge {
  @apply mr-2;
}

.badge-text {
  @apply px-2 py-1 text-xs bg-blue-600 text-white rounded-full;
}

.panel-actions {
  @apply flex items-center space-x-1;
}

.action-btn {
  @apply p-1 text-gray-400 hover:text-gray-200 hover:bg-gray-600 rounded;
  @apply transition-colors duration-150;
}

.close-btn:hover {
  @apply text-red-400 bg-red-900/30;
}

/* Content */
.panel-content {
  @apply flex-1 relative overflow-hidden;
}

.content-wrapper {
  @apply h-full overflow-auto;
}

.content-wrapper.collapsed {
  @apply hidden;
}

/* Loading state */
.panel-loading {
  @apply absolute inset-0 flex flex-col items-center justify-center;
  @apply bg-gray-800/90 backdrop-blur-sm;
}

.loading-spinner {
  @apply mb-4;
}

.spinner {
  @apply w-8 h-8 border-2 border-gray-600 border-t-blue-500 rounded-full animate-spin;
}

.loading-text {
  @apply text-gray-300 text-sm;
}

/* Error state */
.panel-error {
  @apply absolute inset-0 flex flex-col items-center justify-center p-8;
  @apply bg-red-900/20 backdrop-blur-sm;
}

.error-text {
  @apply text-red-300 text-center mt-2 mb-4;
}

.retry-btn {
  @apply px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded;
  @apply transition-colors duration-150;
}

/* Resize handles */
.resize-handle {
  @apply absolute bg-transparent;
  @apply hover:bg-blue-500/30 transition-colors duration-150;
}

.resize-n {
  @apply top-0 left-2 right-2 h-2 cursor-n-resize;
}

.resize-s {
  @apply bottom-0 left-2 right-2 h-2 cursor-s-resize;
}

.resize-e {
  @apply right-0 top-2 bottom-2 w-2 cursor-e-resize;
}

.resize-w {
  @apply left-0 top-2 bottom-2 w-2 cursor-w-resize;
}

.resize-ne {
  @apply top-0 right-0 w-4 h-4 cursor-ne-resize;
}

.resize-nw {
  @apply top-0 left-0 w-4 h-4 cursor-nw-resize;
}

.resize-se {
  @apply bottom-0 right-0 w-4 h-4 cursor-se-resize;
}

.resize-sw {
  @apply bottom-0 left-0 w-4 h-4 cursor-sw-resize;
}

/* Dock indicator */
.dock-indicator {
  @apply fixed border-2 border-dashed border-blue-400 bg-blue-400/20;
  @apply pointer-events-none transition-all duration-200;
  z-index: 10000;
}

.dock-left {
  @apply left-0 top-0 w-1/2 h-full;
}

.dock-right {
  @apply right-0 top-0 w-1/2 h-full;
}

.dock-top {
  @apply left-0 top-0 w-full h-1/2;
}

.dock-bottom {
  @apply left-0 bottom-0 w-full h-1/2;
}

/* Animation states */
.dragging {
  @apply cursor-move;
  transition: none !important;
}

.resizing {
  transition: none !important;
}

/* Responsive adjustments */
@media (max-width: v-bind('APP_CONFIG.ui.responsive.BREAKPOINTS.SM + "px"')) {
  .enhanced-panel {
    @apply text-sm;
  }
  
  .panel-header {
    @apply px-2 py-1;
  }
  
  .resize-handle {
    @apply opacity-100;
  }
}
</style>