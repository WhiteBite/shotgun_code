import { ref, computed, onMounted, onUnmounted } from 'vue'

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
      return { width: '48px', minWidth: '48px', maxWidth: '48px' }
    }
    
    if (isMobile.value) {
      return { width: '100vw' }
    }
    
    // Allow wider panels for context panel specifically
    const maxWidth = panelId === 'context' ? 800 : 480
    
    return {
      width: `${width.value}px`,
      minWidth: '280px',
      maxWidth: `${maxWidth}px`
    }
  })
  
  // Methods
  const toggleCollapse = () => {
    isCollapsed.value = !isCollapsed.value
    saveState()
  }
  
  const setWidth = (newWidth: number) => {
    const maxWidth = panelId === 'context' ? 800 : 480
    const clampedWidth = Math.max(280, Math.min(maxWidth, newWidth))
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
    isMobile.value = window.innerWidth < 768
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