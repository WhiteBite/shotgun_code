import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { defineStore } from 'pinia'
import { createStoreWithDependencies, type StoreDependencies } from '@/stores/StoreDependencyContainer'
import { throttle } from '@/utils/performance'
import { APP_CONFIG } from '@/config/app-config'

// Use centralized configuration instead of hardcoded constants
const MAX_CONCURRENT_TOOLTIPS = APP_CONFIG.ui.tooltips.MAX_CONCURRENT_TOOLTIPS
const CACHE_SIZE = APP_CONFIG.ui.tooltips.CACHE_SIZE
const THROTTLE_DELAY = APP_CONFIG.ui.tooltips.THROTTLE_DELAY
const DEBOUNCE_DELAY = APP_CONFIG.ui.tooltips.DEBOUNCE_DELAY
const VIRTUALIZATION_THRESHOLD = APP_CONFIG.ui.tooltips.VIRTUALIZATION_THRESHOLD

export type TooltipPosition = 'top' | 'bottom' | 'left' | 'right' | 'auto'
export type TooltipTrigger = 'hover' | 'focus' | 'click' | 'manual'
export type TooltipSize = 'sm' | 'md' | 'lg'

export interface TooltipConfig {
  id: string
  content: string | (() => string)
  position: TooltipPosition
  trigger: TooltipTrigger
  delay: number
  hideDelay: number
  disabled: boolean
  allowHTML: boolean
  size: TooltipSize
  maxWidth: number
  zIndex: number
  className?: string
  interactive: boolean
  arrow: boolean
  persistent: boolean
}

export interface TooltipInstance {
  id: string
  element: HTMLElement
  config: TooltipConfig
  isVisible: boolean
  position: { x: number; y: number }
  computedPosition: TooltipPosition
  content: string
  lastUpdated: number
  priority: number
  cached: boolean
}

const defaultConfig: Partial<TooltipConfig> = {
  position: 'auto',
  trigger: 'hover',
  delay: APP_CONFIG.ui.tooltips.DEFAULT_DELAY,
  hideDelay: APP_CONFIG.ui.tooltips.HIDE_DELAY,
  disabled: false,
  allowHTML: false,
  size: 'md',
  maxWidth: APP_CONFIG.ui.tooltips.MAX_WIDTH,
  zIndex: 1020,
  interactive: false,
  arrow: true,
  persistent: false
}

export const useTooltipStore = defineStore('tooltip', () => {
  return createStoreWithDependencies('tooltip', (dependencies: StoreDependencies) => {
    // Inject StorageRepository (even though not currently used, for consistency)
    const { localStorageService } = dependencies;

    // State
    const tooltips = ref<Map<string, TooltipInstance>>(new Map())
    const activeTooltip = ref<string | null>(null)
    const hoveredTooltip = ref<string | null>(null)
    const isGloballyDisabled = ref(false)
    
    // Performance optimization state
    const contentCache = ref<Map<string, { content: string; timestamp: number }>>(new Map())
    const positionCache = ref<Map<string, { position: { x: number; y: number }; timestamp: number }>>(new Map())
    const renderQueue = ref<Set<string>>(new Set())
    const isProcessingQueue = ref(false)
    
    // Timers for delayed show/hide
    const showTimers = ref<Map<string, NodeJS.Timeout>>(new Map())
    const hideTimers = ref<Map<string, NodeJS.Timeout>>(new Map())

    // Performance and caching utilities
    const getCachedContent = (id: string, contentFunc: () => string): string => {
      const cacheKey = `${id}-content`
      const cached = contentCache.value.get(cacheKey)
      
      if (cached && Date.now() - cached.timestamp < 30000) { // 30s cache
        return cached.content
      }
      
      const content = contentFunc()
      contentCache.value.set(cacheKey, { content, timestamp: Date.now() })
      
      // Clean old cache entries if cache is full
      if (contentCache.value.size > CACHE_SIZE) {
        const oldestKey = contentCache.value.keys().next().value
        if (oldestKey) {
          contentCache.value.delete(oldestKey)
        }
      }
      
      return content
    }
    
    const addToRenderQueue = (id: string) => {
      renderQueue.value.add(id)
      if (!isProcessingQueue.value) {
        processRenderQueue()
      }
    }
    
    const processRenderQueue = async () => {
      if (isProcessingQueue.value || renderQueue.value.size === 0) return
      
      isProcessingQueue.value = true
      
      // Process queue in batches for better performance
      const batch = Array.from(renderQueue.value).slice(0, 5)
      renderQueue.value.clear()
      
      for (const id of batch) {
        const instance = tooltips.value.get(id)
        if (instance && instance.isVisible) {
          await nextTick()
          updateTooltipPosition(instance)
        }
      }
      
      isProcessingQueue.value = false
      
      // Process remaining items if any
      if (renderQueue.value.size > 0) {
        setTimeout(processRenderQueue, 16) // Next animation frame
      }
    }
    
    // Handle scroll events to update tooltip positions
    const handleScroll = () => {
      visibleTooltips.value.forEach(tooltip => {
        updateTooltipPosition(tooltip)
      })
    }
    
    // Throttled version of scroll handler
    const throttledHandleScroll = throttle(handleScroll, THROTTLE_DELAY)

    // Computed
    const activeTooltipInstance = computed(() => {
      return activeTooltip.value ? tooltips.value.get(activeTooltip.value) : null
    })

    const visibleTooltips = computed(() => {
      const visible = Array.from(tooltips.value.values()).filter(tooltip => tooltip.isVisible)
      // Sort by priority for rendering optimization
      return visible.sort((a, b) => b.priority - a.priority)
    })
    
    // const shouldVirtualize = computed(() => {
    //   return tooltips.value.size > VIRTUALIZATION_THRESHOLD
    // })

    // Actions
    const registerTooltip = (
      element: HTMLElement, 
      config: Partial<TooltipConfig> & { content: string }
    ): string => {
      const id = generateTooltipId()
      const fullConfig = { ...defaultConfig, ...config, id } as TooltipConfig
      
      const instance: TooltipInstance = {
        id,
        element,
        config: fullConfig,
        isVisible: false,
        position: { x: 0, y: 0 },
        computedPosition: fullConfig.position,
        content: typeof config.content === 'function' ? 
          getCachedContent(id, config.content) : 
          config.content,
        lastUpdated: Date.now(),
        priority: calculatePriority(element, fullConfig),
        cached: false
      }

      tooltips.value.set(id, instance)
      attachEventListeners(instance)
      
      return id
    }

    const unregisterTooltip = (id: string) => {
      const instance = tooltips.value.get(id)
      if (instance) {
        detachEventListeners(instance)
        hideTooltip(id, true)
        tooltips.value.delete(id)
      }
    }

    const updateTooltip = (id: string, updates: Partial<TooltipConfig>) => {
      const instance = tooltips.value.get(id)
      if (instance) {
        instance.config = { ...instance.config, ...updates }
        instance.lastUpdated = Date.now()
        
        // Clear cached content if content changed
        const cacheKey = `${id}-content`
        contentCache.value.delete(cacheKey)
        
        if (typeof updates.content === 'function') {
          instance.content = getCachedContent(id, updates.content)
        } else {
          instance.content = updates.content
        }
        
        // Recalculate priority if configuration changed
        instance.priority = calculatePriority(instance.element, instance.config)
      }
    }

    const showTooltip = async (id: string, force = false) => {
      if (isGloballyDisabled.value && !force) return
      
      const instance = tooltips.value.get(id)
      if (!instance || instance.config.disabled) return

      // Check concurrent tooltip limit
      const currentVisible = visibleTooltips.value.length
      if (currentVisible >= MAX_CONCURRENT_TOOLTIPS && !force) {
        // Hide lowest priority tooltip
        const lowestPriority = visibleTooltips.value[visibleTooltips.value.length - 1]
        if (lowestPriority && instance.priority > lowestPriority.priority) {
          hideTooltip(lowestPriority.id, true)
        } else {
          return // Don't show if current tooltip has lower priority
        }
      }

      // Clear any existing hide timer
      const hideTimer = hideTimers.value.get(id)
      if (hideTimer) {
        clearTimeout(hideTimer)
        hideTimers.value.delete(id)
      }

      // Hide other tooltips if not persistent
      if (!instance.config.persistent) {
        hideAllTooltips(id)
      }

      const showDelayed = () => {
        addToRenderQueue(id)
        instance.isVisible = true
        instance.lastUpdated = Date.now()
        activeTooltip.value = id
        
        // Remove show timer
        showTimers.value.delete(id)
      }

      if (force || instance.config.delay === 0) {
        showDelayed()
      } else {
        // Clear existing show timer
        const existingTimer = showTimers.value.get(id)
        if (existingTimer) {
          clearTimeout(existingTimer)
        }
        
        // Set new show timer
        const timer = setTimeout(showDelayed, instance.config.delay) as NodeJS.Timeout
        showTimers.value.set(id, timer)
      }
    }

    const hideTooltip = (id: string, force = false) => {
      const instance = tooltips.value.get(id)
      if (!instance) return

      // Clear any existing show timer
      const showTimer = showTimers.value.get(id)
      if (showTimer) {
        clearTimeout(showTimer)
        showTimers.value.delete(id)
      }

      const hideDelayed = () => {
        instance.isVisible = false
        if (activeTooltip.value === id) {
          activeTooltip.value = null
        }
        hideTimers.value.delete(id)
      }

      if (force || instance.config.hideDelay === 0) {
        hideDelayed()
      } else {
        const timer = setTimeout(hideDelayed, instance.config.hideDelay) as NodeJS.Timeout
        hideTimers.value.set(id, timer)
      }
    }

    const hideAllTooltips = (except?: string) => {
      tooltips.value.forEach((instance, id) => {
        if (id !== except && !instance.config.persistent) {
          hideTooltip(id, true)
        }
      })
    }

    const toggleTooltip = (id: string) => {
      const instance = tooltips.value.get(id)
      if (instance) {
        if (instance.isVisible) {
          hideTooltip(id)
        } else {
          showTooltip(id)
        }
      }
    }

    const setGlobalDisabled = (disabled: boolean) => {
      isGloballyDisabled.value = disabled
      if (disabled) {
        hideAllTooltips()
      }
    }

    // Position calculation
    const updateTooltipPosition = (instance: TooltipInstance) => {
      const element = instance.element
      const rect = element.getBoundingClientRect()
      const config = instance.config
      
      let position = config.position
      let x = 0
      let y = 0

      // Auto-calculate position if set to 'auto'
      if (position === 'auto') {
        position = calculateOptimalPosition(rect)
        instance.computedPosition = position
      }

      const offset = 12 // Distance from element
      const arrowSize = config.arrow ? 8 : 0

      switch (position) {
        case 'top':
          x = rect.left + rect.width / 2
          y = rect.top - offset - arrowSize
          break
        case 'bottom':
          x = rect.left + rect.width / 2
          y = rect.bottom + offset + arrowSize
          break
        case 'left':
          x = rect.left - offset - arrowSize
          y = rect.top + rect.height / 2
          break
        case 'right':
          x = rect.right + offset + arrowSize
          y = rect.top + rect.height / 2
          break
      }

      instance.position = { x, y }
    }

    const calculateOptimalPosition = (rect: DOMRect): TooltipPosition => {
      const viewport = {
        width: window.innerWidth,
        height: window.innerHeight
      }

      const spaceTop = rect.top
      const spaceBottom = viewport.height - rect.bottom
      const spaceLeft = rect.left
      const spaceRight = viewport.width - rect.right

      // Prioritize top/bottom over left/right
      if (spaceTop > spaceBottom && spaceTop > 100) {
        return 'top'
      } else if (spaceBottom > 100) {
        return 'bottom'
      } else if (spaceRight > spaceLeft && spaceRight > 200) {
        return 'right'
      } else if (spaceLeft > 200) {
        return 'left'
      } else {
        return 'bottom' // Fallback
      }
    }

    // Event handlers
    const attachEventListeners = (instance: TooltipInstance) => {
      const element = instance.element
      const config = instance.config

      if (config.trigger === 'hover') {
        element.addEventListener('mouseenter', () => handleMouseEnter(instance.id))
        element.addEventListener('mouseleave', () => handleMouseLeave(instance.id))
      }

      if (config.trigger === 'focus' || config.trigger === 'hover') {
        element.addEventListener('focus', () => handleFocus(instance.id))
        element.addEventListener('blur', () => handleBlur(instance.id))
      }

      if (config.trigger === 'click') {
        element.addEventListener('click', () => handleClick(instance.id))
      }

      // Keyboard accessibility
      element.addEventListener('keydown', (e) => handleKeydown(e, instance.id))
    }

    const detachEventListeners = (instance: TooltipInstance) => {
      const element = instance.element
      // Remove all possible event listeners
      element.removeEventListener('mouseenter', () => handleMouseEnter(instance.id))
      element.removeEventListener('mouseleave', () => handleMouseLeave(instance.id))
      element.removeEventListener('focus', () => handleFocus(instance.id))
      element.removeEventListener('blur', () => handleBlur(instance.id))
      element.removeEventListener('click', () => handleClick(instance.id))
      element.removeEventListener('keydown', (e) => handleKeydown(e, instance.id))
    }

    const handleMouseEnter = (id: string) => {
      hoveredTooltip.value = id
      showTooltip(id)
    }

    const handleMouseLeave = (id: string) => {
      hoveredTooltip.value = null
      const instance = tooltips.value.get(id)
      if (instance && !instance.config.interactive) {
        hideTooltip(id)
      }
    }

    const handleFocus = (id: string) => {
      showTooltip(id)
    }

    const handleBlur = (id: string) => {
      hideTooltip(id)
    }

    const handleClick = (id: string) => {
      toggleTooltip(id)
    }

    const handleKeydown = (event: KeyboardEvent, id: string) => {
      if (event.key === 'Escape') {
        hideTooltip(id, true)
      }
    }

    // Utility functions
    const generateTooltipId = (): string => {
      return `tooltip-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    }
    
    // Calculate tooltip priority based on element type and configuration
    const calculatePriority = (element: HTMLElement, config: TooltipConfig): number => {
      let priority = 0
      
      // Base priority by element type
      const tagName = element.tagName.toLowerCase()
      switch (tagName) {
        case 'button':
          priority += 10
          break
        case 'input':
        case 'textarea':
        case 'select':
          priority += 8
          break
        case 'a':
          priority += 6
          break
        default:
          priority += 3
      }
      
      // Priority by trigger type
      switch (config.trigger) {
        case 'focus':
          priority += 5
          break
        case 'click':
          priority += 3
          break
        case 'hover':
          priority += 1
          break
      }
      
      // Higher priority for interactive tooltips
      if (config.interactive) priority += 2
      
      // Higher priority for persistent tooltips
      if (config.persistent) priority += 3
      
      // Priority based on viewport position (elements in center get higher priority)
      const rect = element.getBoundingClientRect()
      const centerX = window.innerWidth / 2
      const centerY = window.innerHeight / 2
      const distance = Math.sqrt(
        Math.pow(rect.left + rect.width / 2 - centerX, 2) +
        Math.pow(rect.top + rect.height / 2 - centerY, 2)
      )
      const maxDistance = Math.sqrt(centerX * centerX + centerY * centerY)
      priority += Math.max(0, 5 - (distance / maxDistance) * 5)
      
      return Math.round(priority)
    }

    // Cleanup
    const cleanup = () => {
      // Clear all timers
      showTimers.value.forEach(timer => clearTimeout(timer))
      hideTimers.value.forEach(timer => clearTimeout(timer))
      showTimers.value.clear()
      hideTimers.value.clear()

      // Clear caches
      contentCache.value.clear()
      positionCache.value.clear()
      renderQueue.value.clear()

      // Hide all tooltips
      hideAllTooltips()
    }
    
    // Performance monitoring
    const getPerformanceStats = () => {
      return {
        totalTooltips: tooltips.value.size,
        visibleTooltips: visibleTooltips.value.length,
        cacheSize: contentCache.value.size,
        positionCacheSize: positionCache.value.size,
        queueSize: renderQueue.value.size,
        isProcessing: isProcessingQueue.value
      }
    }

    // Global event listeners with performance optimization
    onMounted(() => {
      // Throttled scroll and resize handlers
      window.addEventListener('scroll', throttledHandleScroll, { passive: true })
      window.addEventListener('resize', throttledHandleScroll, { passive: true })
      
      // Hide tooltips on outside click with delegation
      document.addEventListener('click', (e) => {
        const target = e.target as HTMLElement
        if (!target.closest('[data-tooltip]') && !target.closest('.tooltip-content')) {
          hideAllTooltips()
        }
      }, { passive: true })
      
      // Performance monitoring in development
      if (process.env.NODE_ENV === 'development') {
        const logStats = () => {
          const stats = getPerformanceStats()
          if (stats.totalTooltips > 0) {
            console.debug('Tooltip Performance Stats:', stats)
          }
        }
        setInterval(logStats, 10000) // Log every 10 seconds
      }
    })

    onUnmounted(() => {
      cleanup()
    })

    return {
      // State
      tooltips: computed(() => tooltips.value),
      activeTooltip: computed(() => activeTooltip.value),
      activeTooltipInstance,
      visibleTooltips,
      isGloballyDisabled: computed(() => isGloballyDisabled.value),

      // Actions
      registerTooltip,
      unregisterTooltip,
      updateTooltip,
      showTooltip,
      hideTooltip,
      hideAllTooltips,
      toggleTooltip,
      setGlobalDisabled,
      cleanup
    }
  })
})

// Composable hook
export function useTooltip() {
  return useTooltipStore()
}