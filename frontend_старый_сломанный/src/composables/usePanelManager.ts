import { ref, computed, onMounted, onUnmounted } from 'vue'
import { APP_CONFIG } from '@/config/app-config'

export interface PanelState {
  isCollapsed: boolean
  width: number
  isResizing: boolean
}

export function usePanelManager(panelId: string, defaultWidth = 320) {
  const isCollapsed = ref(false)
  const width = ref(defaultWidth)
  const isResizing = ref(false)
  const isMobile = ref(false)
  
  // Load saved state from localStorage
  const savedState = localStorage.getItem(`panel-${panelId}`)
  if (savedState) {
    try {
      const parsed = JSON.parse(savedState)
      isCollapsed.value = parsed.isCollapsed || false
      width.value = parsed.width || defaultWidth
    } catch (e) {
      console.warn('Failed to parse saved panel state:', e)
    }
  }
  
  // Reactive state
  const panelState = computed((): PanelState => ({
    isCollapsed: isCollapsed.value,
    width: width.value,
    isResizing: isResizing.value
  }))
  
  // Panel classes
  const panelClasses = computed(() => ({
    'panel-container': true,
    'collapsed': isCollapsed.value,
    'responsive': !isCollapsed.value,
    'mobile-full': isMobile.value && !isCollapsed.value
  }))
  
  // Panel style
  const panelStyle = computed(() => {
    if (isCollapsed.value) {
      return { 
        width: `${APP_CONFIG.ui.panels.COLLAPSED_WIDTH}px`, 
        minWidth: `${APP_CONFIG.ui.panels.COLLAPSED_WIDTH}px`, 
        maxWidth: `${APP_CONFIG.ui.panels.COLLAPSED_WIDTH}px` 
      }
    }
    
    if (isMobile.value) {
      return { width: '100vw' }
    }
    
    // Allow wider panels for context panel specifically
    const maxWidth = panelId === 'context' ? APP_CONFIG.ui.panels.CONTEXT_PANEL_MAX_WIDTH : APP_CONFIG.ui.panels.RESULTS_PANEL_MAX_WIDTH
    
    return {
      width: `${width.value}px`,
      minWidth: `${APP_CONFIG.ui.panels.MIN_PANEL_WIDTH}px`,
      maxWidth: `${maxWidth}px`
    }
  })
  
  // Methods
  const toggleCollapse = () => {
    isCollapsed.value = !isCollapsed.value
    saveState()
  }
  
  const setWidth = (newWidth: number) => {
    const maxWidth = panelId === 'context' ? APP_CONFIG.ui.panels.CONTEXT_PANEL_MAX_WIDTH : APP_CONFIG.ui.panels.RESULTS_PANEL_MAX_WIDTH
    const clampedWidth = Math.max(APP_CONFIG.ui.panels.MIN_PANEL_WIDTH, Math.min(maxWidth, newWidth))
    width.value = clampedWidth
    saveState()
  }
  
  const startResize = () => {
    isResizing.value = true
  }
  
  const stopResize = () => {
    isResizing.value = false
  }
  
  const saveState = () => {
    const state = {
      isCollapsed: isCollapsed.value,
      width: width.value
    }
    localStorage.setItem(`panel-${panelId}`, JSON.stringify(state))
  }
  
  // Responsive handling
  const checkScreenSize = () => {
    isMobile.value = window.innerWidth < APP_CONFIG.ui.responsive.BREAKPOINTS.SM
  }
  
  onMounted(() => {
    checkScreenSize()
    window.addEventListener('resize', checkScreenSize)
  })
  
  onUnmounted(() => {
    window.removeEventListener('resize', checkScreenSize)
  })
  
  return {
    // State
    panelState,
    isCollapsed: computed(() => isCollapsed.value),
    width: computed(() => width.value),
    isResizing: computed(() => isResizing.value),
    isMobile: computed(() => isMobile.value),
    
    // Computed
    panelClasses,
    panelStyle,
    
    // Methods
    toggleCollapse,
    setWidth,
    startResize,
    stopResize,
    saveState
  }
}

export default usePanelManager