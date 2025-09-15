<template>
  <div :class="panelClasses" :style="panelStyles" ref="panelRef">
    <!-- Panel Header -->
    <div 
      v-if="!isCollapsed || showHeaderWhenCollapsed"
      class="panel-header"
      :class="headerClasses"
    >
      <div class="panel-header-content">
        <div class="panel-title-section">
          <component 
            v-if="icon && !isCollapsed" 
            :is="icon" 
            class="panel-icon"
          />
          <h3 
            v-if="title && !isCollapsed" 
            class="panel-title"
          >
            {{ title }}
          </h3>
        </div>
        
        <div class="panel-header-actions">
          <slot name="header-actions" />
          
          <button 
            v-if="collapsible"
            class="panel-collapse-btn"
            @click="handleToggle"
            :title="isCollapsed ? 'Expand panel' : 'Collapse panel'"
            :aria-label="`${isCollapsed ? 'Expand' : 'Collapse'} ${title || 'panel'}`"
          >
            <ChevronLeftIcon 
              class="collapse-icon"
              :class="{ 'rotate-180': isCollapsed }"
            />
          </button>
        </div>
      </div>
    </div>

    <!-- Panel Content -->
    <div 
      v-if="!isCollapsed"
      class="panel-content"
      :class="contentClasses"
    >
      <div 
        v-if="scrollable"
        ref="scrollContainer"
        class="panel-scrollable-content"
        :class="scrollClasses"
        @scroll="handleScroll"
      >
        <LoadingState v-if="loading" />
        <ErrorState v-else-if="error" :error="error" @retry="$emit('retry')" />
        <slot v-else />
      </div>
      
      <div v-else class="panel-static-content">
        <LoadingState v-if="loading" />
        <ErrorState v-else-if="error" :error="error" @retry="$emit('retry')" />
        <slot v-else />
      </div>
    </div>

    <!-- Panel Footer -->
    <div 
      v-if="$slots.footer && !isCollapsed"
      class="panel-footer"
      :class="footerClasses"
    >
      <slot name="footer" />
    </div>

    <!-- Resize Handle -->
    <div 
      v-if="resizable && !isCollapsed"
      class="panel-resize-handle"
      :class="resizeHandleClasses"
      @mousedown="startResize"
    >
      <div class="resize-indicator" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ChevronLeftIcon } from 'lucide-vue-next'
import LoadingState from './LoadingState.vue'
import ErrorState from './ErrorState.vue'

// Types
interface PanelError {
  message: string
  code?: string
}

// Props
interface Props {
  title?: string
  icon?: any
  collapsible?: boolean
  isCollapsed?: boolean
  scrollable?: boolean
  loading?: boolean
  error?: PanelError | null
  resizable?: boolean
  resizeDirection?: 'right' | 'left'
  minWidth?: number
  maxWidth?: number
  width?: number
  variant?: 'primary' | 'secondary' | 'elevated'
  size?: 'sm' | 'md' | 'lg' | 'xl'
  showHeaderWhenCollapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  collapsible: true,
  isCollapsed: false,
  scrollable: true,
  loading: false,
  error: null,
  resizable: true,
  resizeDirection: 'right',
  minWidth: 280,
  maxWidth: 600,
  width: 320,
  variant: 'primary',
  size: 'md',
  showHeaderWhenCollapsed: false
})

// Emits
const emit = defineEmits<{
  toggle: [collapsed: boolean]
  resize: [width: number]
  scroll: [event: Event]
  retry: []
}>()

// Refs
const panelRef = ref<HTMLElement>()
const scrollContainer = ref<HTMLElement>()

// State
const isResizing = ref(false)
const currentWidth = ref(props.width)

// Computed
const panelClasses = computed(() => [
  'base-panel',
  `panel-variant-${props.variant}`,
  `panel-size-${props.size}`,
  {
    'panel-collapsed': props.isCollapsed,
    'panel-resizable': props.resizable,
    'panel-resizing': isResizing.value
  }
])

const panelStyles = computed(() => ({
  '--panel-width': `${currentWidth.value}px`,
  '--panel-min-width': `${props.minWidth}px`,
  '--panel-max-width': `${props.maxWidth}px`,
  width: props.isCollapsed ? '48px' : `${currentWidth.value}px`,
  minWidth: props.isCollapsed ? '48px' : `${props.minWidth}px`,
  maxWidth: props.isCollapsed ? '48px' : `${props.maxWidth}px`
}))

const headerClasses = computed(() => [
  {
    'panel-header-collapsed': props.isCollapsed
  }
])

const contentClasses = computed(() => [
  {
    'panel-content-scrollable': props.scrollable,
    'panel-content-loading': props.loading,
    'panel-content-error': props.error
  }
])

const scrollClasses = computed(() => [
  'custom-scrollbar',
  'smooth-scrollbar',
  {
    'scroll-enabled': props.scrollable
  }
])

const footerClasses = computed(() => [
  'panel-footer-default'
])

const resizeHandleClasses = computed(() => [
  `resize-handle-${props.resizeDirection}`
])

// Methods
function handleToggle() {
  emit('toggle', !props.isCollapsed)
}

function handleScroll(event: Event) {
  emit('scroll', event)
}

function startResize(event: MouseEvent) {
  if (!props.resizable) return
  
  isResizing.value = true
  const startX = event.clientX
  const startWidth = currentWidth.value

  const onMouseMove = (e: MouseEvent) => {
    const deltaX = props.resizeDirection === 'right' ? e.clientX - startX : startX - e.clientX
    const newWidth = Math.max(
      props.minWidth,
      Math.min(props.maxWidth, startWidth + deltaX)
    )
    currentWidth.value = newWidth
    emit('resize', newWidth)
  }

  const onMouseUp = () => {
    isResizing.value = false
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }

  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

// Watch for external width changes
watch(() => props.width, (newWidth) => {
  currentWidth.value = newWidth
}, { immediate: true })

onMounted(() => {
  // Initialize width
  currentWidth.value = props.width
  
  // Ensure scroll container has proper styles
  if (scrollContainer.value) {
    scrollContainer.value.style.overflowY = props.scrollable ? 'auto' : 'hidden'
    scrollContainer.value.style.overflowX = 'hidden'
  }
})
</script>

<style scoped>
.base-panel {
  /* Glass morphism background */
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
  
  /* Layout */
  display: flex;
  flex-direction: column;
  height: 100%;
  position: relative;
  
  /* Transitions */
  transition: 
    width var(--transition-normal),
    min-width var(--transition-normal),
    background var(--transition-fast),
    backdrop-filter var(--transition-fast);
  
  /* Shadow */
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.panel-variant-primary {
  background: rgba(15, 23, 42, 0.85);
  border-color: rgba(148, 163, 184, 0.1);
}

.panel-variant-secondary {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(148, 163, 184, 0.15);
}

.panel-variant-elevated {
  background: rgba(51, 65, 85, 0.9);
  border-color: rgba(148, 163, 184, 0.2);
  box-shadow: 
    0 12px 40px rgba(0, 0, 0, 0.4),
    inset 0 1px 0 rgba(255, 255, 255, 0.15);
}

.panel-collapsed {
  min-width: 48px !important;
  max-width: 48px !important;
  width: 48px !important;
}

.panel-header {
  background: linear-gradient(145deg, 
    rgba(51, 65, 85, 0.95), 
    rgba(30, 41, 59, 0.9)
  );
  backdrop-filter: blur(8px);
  border-bottom: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 11px 11px 0 0;
  padding: 16px;
  flex-shrink: 0;
  position: relative;
}

.panel-header-collapsed {
  padding: 12px 8px;
  border-radius: 11px;
}

.panel-header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.panel-title-section {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.panel-icon {
  width: 20px;
  height: 20px;
  color: rgb(168, 85, 247);
  flex-shrink: 0;
}

.panel-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: rgb(248, 250, 252);
  margin: 0;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.panel-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.panel-collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: rgba(51, 65, 85, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  color: rgb(203, 213, 225);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.panel-collapse-btn:hover {
  background: rgba(59, 130, 246, 0.2);
  border-color: rgba(59, 130, 246, 0.4);
  color: rgb(147, 197, 253);
}

.collapse-icon {
  width: 16px;
  height: 16px;
  transition: transform var(--transition-normal);
}

.panel-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  position: relative;
}

.panel-scrollable-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 16px;
  min-height: 0;
  
  /* Ensure proper scrolling */
  scrollbar-width: thin;
  scrollbar-color: rgba(59, 130, 246, 0.5) rgba(30, 41, 59, 0.4);
}

.panel-scrollable-content::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.panel-scrollable-content::-webkit-scrollbar-track {
  background: rgba(30, 41, 59, 0.4);
  border-radius: 4px;
}

.panel-scrollable-content::-webkit-scrollbar-thumb {
  background: rgba(59, 130, 246, 0.5);
  border-radius: 4px;
  transition: background 0.2s ease;
}

.panel-scrollable-content::-webkit-scrollbar-thumb:hover {
  background: rgba(59, 130, 246, 0.8);
}

.panel-static-content {
  flex: 1;
  padding: 16px;
  overflow: hidden;
}

.panel-footer {
  background: rgba(30, 41, 59, 0.8);
  border-top: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 0 0 11px 11px;
  padding: 16px;
  flex-shrink: 0;
}

.panel-resize-handle {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 4px;
  cursor: col-resize;
  background: transparent;
  transition: background var(--transition-fast);
  z-index: 10;
}

.resize-handle-right {
  right: -2px;
}

.resize-handle-left {
  left: -2px;
}

.panel-resize-handle:hover,
.panel-resizing .panel-resize-handle {
  background: rgba(59, 130, 246, 0.3);
}

.panel-resize-handle:hover .resize-indicator,
.panel-resizing .panel-resize-handle .resize-indicator {
  background: rgb(59, 130, 246);
  opacity: 1;
}

.resize-indicator {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 2px;
  height: 24px;
  background: rgba(59, 130, 246, 0.5);
  border-radius: 1px;
  opacity: 0;
  transition: all var(--transition-fast);
}

/* Responsive design */
@media (max-width: 768px) {
  .base-panel {
    border-radius: 8px;
  }
  
  .panel-header {
    padding: 12px;
    border-radius: 7px 7px 0 0;
  }
  
  .panel-scrollable-content,
  .panel-static-content,
  .panel-footer {
    padding: 12px;
  }
}

/* Animation states */
.panel-content-loading {
  pointer-events: none;
}

.panel-resizing {
  user-select: none;
  pointer-events: none;
}

.panel-resizing * {
  cursor: col-resize !important;
}

/* Scroll specific classes */
.scroll-enabled {
  overflow-y: auto !important;
  overflow-x: hidden !important;
}
</style>