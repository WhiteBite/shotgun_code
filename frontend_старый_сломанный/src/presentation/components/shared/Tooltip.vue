<template>
  <div
    ref="triggerRef"
    class="tooltip-wrapper"
    :class="wrapperClass"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
    @focus="handleFocus"
    @blur="handleBlur"
    @click="handleClick"
    @keydown="handleKeydown"
  >
    <slot :is-visible="isVisible" :show="show" :hide="hide" :toggle="toggle" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useTooltip, type TooltipConfig } from '@/composables/useTooltip'

// Props
interface Props {
  content: string | (() => string)
  position?: TooltipConfig['position']
  trigger?: TooltipConfig['trigger']
  delay?: number
  hideDelay?: number
  disabled?: boolean
  allowHTML?: boolean
  size?: TooltipConfig['size']
  maxWidth?: number
  interactive?: boolean
  arrow?: boolean
  persistent?: boolean
  className?: string
  wrapperClass?: string
}

const props = withDefaults(defineProps<Props>(), {
  position: 'auto',
  trigger: 'hover',
  delay: 500,
  hideDelay: 200,
  disabled: false,
  allowHTML: false,
  size: 'md',
  maxWidth: 300,
  interactive: false,
  arrow: true,
  persistent: false,
  className: '',
  wrapperClass: ''
})

// Emits
interface Emits {
  show: []
  hide: []
  toggle: [visible: boolean]
}

const emit = defineEmits<Emits>()

// Setup
const tooltipStore = useTooltip()
const triggerRef = ref<HTMLElement>()
const tooltipId = ref<string>()

// Computed
const isVisible = computed(() => {
  return tooltipId.value ? 
    tooltipStore.tooltips.get(tooltipId.value)?.isVisible || false : 
    false
})

const tooltipConfig = computed((): Partial<TooltipConfig> => ({
  content: props.content,
  position: props.position,
  trigger: props.trigger,
  delay: props.delay,
  hideDelay: props.hideDelay,
  disabled: props.disabled,
  allowHTML: props.allowHTML,
  size: props.size,
  maxWidth: props.maxWidth,
  interactive: props.interactive,
  arrow: props.arrow,
  persistent: props.persistent,
  className: props.className
}))

// Methods
const show = (force = false) => {
  if (tooltipId.value) {
    tooltipStore.showTooltip(tooltipId.value, force)
    emit('show')
  }
}

const hide = (force = false) => {
  if (tooltipId.value) {
    tooltipStore.hideTooltip(tooltipId.value, force)
    emit('hide')
  }
}

const toggle = () => {
  if (tooltipId.value) {
    tooltipStore.toggleTooltip(tooltipId.value)
    emit('toggle', isVisible.value)
  }
}

// Event handlers
const handleMouseEnter = () => {
  if (props.trigger === 'hover') {
    show()
    emit('show')
  }
}

const handleMouseLeave = () => {
  if (props.trigger === 'hover' && !props.interactive) {
    hide()
    emit('hide')
  }
}

const handleFocus = () => {
  if (props.trigger === 'focus' || props.trigger === 'hover') {
    show()
    emit('show')
  }
}

const handleBlur = () => {
  if (props.trigger === 'focus' || (props.trigger === 'hover' && !props.interactive)) {
    hide()
    emit('hide')
  }
}

const handleClick = () => {
  if (props.trigger === 'click') {
    toggle()
  }
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && isVisible.value) {
    hide(true)
    emit('hide')
  }
}

// Lifecycle
onMounted(() => {
  if (triggerRef.value) {
    tooltipId.value = tooltipStore.registerTooltip(triggerRef.value, tooltipConfig.value)
    
    // Add accessibility attributes
    triggerRef.value.setAttribute('aria-describedby', `tooltip-${tooltipId.value}`)
    
    // Make focusable if needed
    if (!triggerRef.value.hasAttribute('tabindex') && 
        !['BUTTON', 'INPUT', 'SELECT', 'TEXTAREA', 'A'].includes(triggerRef.value.tagName)) {
      triggerRef.value.setAttribute('tabindex', '0')
    }
  }
})

onUnmounted(() => {
  if (tooltipId.value) {
    tooltipStore.unregisterTooltip(tooltipId.value)
  }
})

// Watch for prop changes
watch(
  tooltipConfig,
  (newConfig) => {
    if (tooltipId.value) {
      tooltipStore.updateTooltip(tooltipId.value, newConfig)
    }
  },
  { deep: true }
)

// Expose methods for parent access
defineExpose({
  show,
  hide,
  toggle,
  isVisible
})
</script>

<style scoped>
.tooltip-wrapper {
  display: inline-block;
  position: relative;
}

.tooltip-wrapper:focus {
  outline: 2px solid var(--accent-primary, hsl(210, 100%, 65%));
  outline-offset: 2px;
  border-radius: var(--radius-sm, 0.25rem);
}

.tooltip-wrapper:focus:not(:focus-visible) {
  outline: none;
}
</style>