import type { App, DirectiveBinding } from 'vue'
import { useTooltip, type TooltipConfig } from '@/composables/useTooltip'
import { APP_CONFIG } from '@/config/app-config'

interface TooltipElement extends HTMLElement {
  _tooltipId?: string
  _tooltipStore?: ReturnType<typeof useTooltip>
}

interface TooltipBinding {
  content: string
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
}

function parseBinding(binding: DirectiveBinding<string | TooltipBinding>): TooltipBinding {
  if (typeof binding.value === 'string') {
    return {
      content: binding.value,
      ...binding.modifiers
    }
  }
  
  return binding.value || { content: '' }
}

function updateTooltipConfig(element: TooltipElement, binding: DirectiveBinding) {
  const config = parseBinding(binding)
  
  if (!config.content) {
    console.warn('Tooltip directive requires content')
    return
  }

  // Get or create tooltip store instance
  if (!element._tooltipStore) {
    element._tooltipStore = useTooltip()
  }

  const store = element._tooltipStore

  // Update existing tooltip or create new one
  if (element._tooltipId) {
    store.updateTooltip(element._tooltipId, config)
  } else {
    element._tooltipId = store.registerTooltip(element, config)
    element.setAttribute('data-tooltip', element._tooltipId)
    
    // Add ARIA attributes for accessibility
    element.setAttribute('aria-describedby', `tooltip-${element._tooltipId}`)
    
    // Make element focusable if it's not already
    if (!element.hasAttribute('tabindex') && !['BUTTON', 'INPUT', 'SELECT', 'TEXTAREA', 'A'].includes(element.tagName)) {
      element.setAttribute('tabindex', '0')
    }
  }
}

function removeTooltip(element: TooltipElement) {
  if (element._tooltipId && element._tooltipStore) {
    element._tooltipStore.unregisterTooltip(element._tooltipId)
    element.removeAttribute('data-tooltip')
    element.removeAttribute('aria-describedby')
    delete element._tooltipId
    delete element._tooltipStore
  }
}

export const vTooltip = {
  mounted(element: TooltipElement, binding: DirectiveBinding) {
    updateTooltipConfig(element, binding)
  },

  updated(element: TooltipElement, binding: DirectiveBinding) {
    updateTooltipConfig(element, binding)
  },

  beforeUnmount(element: TooltipElement) {
    removeTooltip(element)
  }
}

// Plugin installer
export default {
  install(app: App) {
    app.directive('tooltip', vTooltip)
  }
}

// Usage examples:
/*
<!-- Simple tooltip -->
<button v-tooltip="'This is a tooltip'">Button</button>

<!-- With options object -->
<button v-tooltip="{ 
  content: 'Complex tooltip', 
  position: 'top', 
  delay: APP_CONFIG.ui.tooltips.DEFAULT_DELAY, // {APP_CONFIG.ui.tooltips.DEFAULT_DELAY}ms
  interactive: true 
}">
  Button
</button>

<!-- With modifiers (for simple cases) -->
<button v-tooltip.top.interactive="'Tooltip with modifiers'">Button</button>

<!-- Dynamic content -->
<button v-tooltip="{ content: dynamicTooltipContent }">Button</button>

<!-- HTML content -->
<button v-tooltip="{ 
  content: '<strong>Bold</strong> tooltip', 
  allowHTML: true 
}">
  Button
</button>

<!-- Disabled tooltip -->
<button v-tooltip="{ 
  content: 'This tooltip is disabled', 
  disabled: isTooltipDisabled 
}">
  Button
</button>
*/