import { App, type DirectiveBinding } from 'vue'
import { useContextualTooltip } from '@/composables/useContextualTooltip'
import { useTooltip } from '@/composables/useTooltip'
import { APP_CONFIG } from '@/config/app-config'

export interface SmartTooltipOptions {
  content: string | (() => string)
  type?: 'file' | 'mode' | 'project' | 'ui' | 'action'
  action?: string
  position?: 'top' | 'bottom' | 'left' | 'right' | 'auto'
  trigger?: 'hover' | 'focus' | 'click' | 'manual'
  delay?: number
  maxWidth?: number
  adaptive?: boolean
  showShortcuts?: boolean
  showStats?: boolean
  showTips?: boolean
  allowHTML?: boolean
  interactive?: boolean
  size?: 'sm' | 'md' | 'lg'
}

// Element type detection based on tag name, classes, and attributes
const detectElementType = (el: HTMLElement): string => {
  const tagName = el.tagName.toLowerCase()
  const classes = el.className
  const role = el.getAttribute('role')
  const ariaLabel = el.getAttribute('aria-label')
  
  // Check for specific patterns
  if (classes.includes('file-') || ariaLabel?.includes('file')) return 'file'
  if (classes.includes('mode-') || ariaLabel?.includes('mode')) return 'mode'
  if (classes.includes('project-') || ariaLabel?.includes('project')) return 'project'
  if (tagName === 'button' || role === 'button') return 'action'
  if (tagName === 'input' || tagName === 'textarea' || tagName === 'select') return 'input'
  if (classes.includes('panel-') || role === 'tabpanel') return 'panel'
  if (classes.includes('nav-') || tagName === 'nav') return 'navigation'
  
  return 'ui'
}

// Action detection based on common patterns
const detectAction = (el: HTMLElement): string | undefined => {
  const ariaLabel = el.getAttribute('aria-label')?.toLowerCase()
  const title = el.getAttribute('title')?.toLowerCase()
  const textContent = el.textContent?.toLowerCase()
  const classes = el.className.toLowerCase()
  
  const actionKeywords = {
    'save': ['save', 'persist'],
    'export': ['export', 'download'],
    'import': ['import', 'upload'],
    'delete': ['delete', 'remove', 'trash'],
    'edit': ['edit', 'modify'],
    'settings': ['settings', 'preferences', 'config'],
    'help': ['help', 'info', 'guide'],
    'search': ['search', 'find'],
    'generate': ['generate', 'create', 'build'],
    'toggle-mode': ['mode', 'toggle'],
    'new-file': ['new', 'create file'],
    'open': ['open', 'load']
  }
  
  for (const [action, keywords] of Object.entries(actionKeywords)) {
    if (keywords.some(keyword => 
      ariaLabel?.includes(keyword) || 
      title?.includes(keyword) || 
      textContent?.includes(keyword) ||
      classes.includes(keyword.replace(' ', '-'))
    )) {
      return action
    }
  }
  
  return undefined
}

// Smart tooltip directive
const smartTooltipDirective = {
  mounted(el: HTMLElement, binding: DirectiveBinding<SmartTooltipOptions | string>) {
    const options: SmartTooltipOptions = binding.value || { content: binding.arg || '' }
    
    // Auto-detect element type and action if not specified
    const elementType = options.type || detectElementType(el)
    const action = options.action || detectAction(el)
    
    // Initialize tooltip services
    const contextualTooltip = useContextualTooltip()
    const tooltipService = useTooltip()
    
    // Generate contextual content
    const generateContent = () => {
      const baseContent = typeof options.content === 'function' 
        ? options.content() 
        : options.content
      
      if (!options.adaptive) {
        return baseContent
      }
      
      // Use appropriate contextual generator
      switch (elementType) {
        case 'file':
          return contextualTooltip.createFileTooltip(baseContent, action)
        case 'mode':
          return contextualTooltip.createModeTooltip(baseContent, action)
        case 'project':
          return contextualTooltip.createProjectTooltip(baseContent, action)
        default:
          return contextualTooltip.createUITooltip(baseContent, action)
      }
    }
    
    // Register tooltip with enhanced configuration
    const tooltipId = tooltipService.registerTooltip(el, {
      content: generateContent,
      position: options.position || 'auto',
      trigger: options.trigger || 'hover',
      delay: options.delay || APP_CONFIG.ui.tooltips.DEFAULT_DELAY,
      maxWidth: options.maxWidth || APP_CONFIG.ui.tooltips.MAX_WIDTH,
      allowHTML: options.allowHTML !== undefined ? options.allowHTML : APP_CONFIG.ui.tooltips.DEFAULT_ALLOW_HTML,
      interactive: options.interactive !== undefined ? options.interactive : APP_CONFIG.ui.tooltips.DEFAULT_INTERACTIVE,
      size: options.size || APP_CONFIG.ui.tooltips.DEFAULT_SIZE
    })
    
    // Store tooltip ID for cleanup
    el._smartTooltipId = tooltipId
    
    // Add element type class for styling
    el.classList.add(`tooltip-type-${elementType}`)
    
    // Add accessibility attributes
    if (!el.hasAttribute('aria-describedby')) {
      el.setAttribute('aria-describedby', `tooltip-${tooltipId}`)
    }
  },
  
  updated(el: HTMLElement, binding: DirectiveBinding<SmartTooltipOptions | string>) {
    const options: SmartTooltipOptions = binding.value || { content: binding.arg || '' }
    const tooltipService = useTooltip()
    
    if (el._smartTooltipId) {
      // Update tooltip content
      const elementType = options.type || detectElementType(el)
      const action = options.action || detectAction(el)
      const contextualTooltip = useContextualTooltip()
      
      const generateContent = () => {
        const baseContent = typeof options.content === 'function' 
          ? options.content() 
          : options.content
        
        if (!options.adaptive) {
          return baseContent
        }
        
        switch (elementType) {
          case 'file':
            return contextualTooltip.createFileTooltip(baseContent, action)
          case 'mode':
            return contextualTooltip.createModeTooltip(baseContent, action)
          case 'project':
            return contextualTooltip.createProjectTooltip(baseContent, action)
          default:
            return contextualTooltip.createUITooltip(baseContent, action)
        }
      }
      
      tooltipService.updateTooltip(el._smartTooltipId, {
        content: generateContent
      })
    }
  },
  
  unmounted(el: HTMLElement) {
    const tooltipService = useTooltip()
    
    if (el._smartTooltipId) {
      tooltipService.unregisterTooltip(el._smartTooltipId)
      delete el._smartTooltipId
    }
    
    // Clean up classes
    el.classList.forEach(className => {
      if (className.startsWith('tooltip-type-')) {
        el.classList.remove(className)
      }
    })
  }
}

// Extend HTMLElement type to include our custom property
declare global {
  interface HTMLElement {
    _smartTooltipId?: string
  }
}

// Plugin installation
export default {
  install(app: App) {
    app.directive('smart-tooltip', smartTooltipDirective)
    
    // Also register as v-tooltip-smart for backwards compatibility
    app.directive('tooltip-smart', smartTooltipDirective)
  }
}

// Export directive for manual use
export { smartTooltipDirective }