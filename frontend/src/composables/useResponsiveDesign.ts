import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useWorkspaceStore } from '@/stores/workspace.store'

export interface ResponsiveBreakpoint {
  name: string
  minWidth: number
  maxWidth?: number
  columns: number
  panelSizes: {
    contextPanel: number
    resultsPanel: number
  }
  features: {
    showContextArea: boolean
    showResultsArea: boolean
    showStatusBar: boolean
    collapsiblePanels: boolean
    stackedLayout: boolean
  }
}

export const BREAKPOINTS: ResponsiveBreakpoint[] = [
  {
    name: 'xs',
    minWidth: 0,
    maxWidth: 640,
    columns: 1,
    panelSizes: {
      contextPanel: 300,
      resultsPanel: 300
    },
    features: {
      showContextArea: false,
      showResultsArea: false,
      showStatusBar: true,
      collapsiblePanels: true,
      stackedLayout: true
    }
  },
  {
    name: 'sm',
    minWidth: 640,
    maxWidth: 768,
    columns: 1,
    panelSizes: {
      contextPanel: 320,
      resultsPanel: 320
    },
    features: {
      showContextArea: true,
      showResultsArea: false,
      showStatusBar: true,
      collapsiblePanels: true,
      stackedLayout: true
    }
  },
  {
    name: 'md',
    minWidth: 768,
    maxWidth: 1024,
    columns: 2,
    panelSizes: {
      contextPanel: 350,
      resultsPanel: 400
    },
    features: {
      showContextArea: true,
      showResultsArea: true,
      showStatusBar: true,
      collapsiblePanels: true,
      stackedLayout: false
    }
  },
  {
    name: 'lg',
    minWidth: 1024,
    maxWidth: 1280,
    columns: 3,
    panelSizes: {
      contextPanel: 400,
      resultsPanel: 500
    },
    features: {
      showContextArea: true,
      showResultsArea: true,
      showStatusBar: true,
      collapsiblePanels: false,
      stackedLayout: false
    }
  },
  {
    name: 'xl',
    minWidth: 1280,
    maxWidth: 1536,
    columns: 3,
    panelSizes: {
      contextPanel: 450,
      resultsPanel: 550
    },
    features: {
      showContextArea: true,
      showResultsArea: true,
      showStatusBar: true,
      collapsiblePanels: false,
      stackedLayout: false
    }
  },
  {
    name: '2xl',
    minWidth: 1536,
    columns: 3,
    panelSizes: {
      contextPanel: 500,
      resultsPanel: 600
    },
    features: {
      showContextArea: true,
      showResultsArea: true,
      showStatusBar: true,
      collapsiblePanels: false,
      stackedLayout: false
    }
  }
]

export function useResponsiveDesign() {
  const workspaceStore = useWorkspaceStore()
  
  // Reactive window dimensions
  const windowWidth = ref(window.innerWidth)
  const windowHeight = ref(window.innerHeight)
  
  // Current breakpoint
  const currentBreakpoint = computed(() => {
    const width = windowWidth.value
    return BREAKPOINTS.find(bp => {
      if (bp.maxWidth) {
        return width >= bp.minWidth && width < bp.maxWidth
      }
      return width >= bp.minWidth
    }) || BREAKPOINTS[BREAKPOINTS.length - 1]
  })
  
  // Responsive utilities
  const isMobile = computed(() => currentBreakpoint.value.name === 'xs' || currentBreakpoint.value.name === 'sm')
  const isTablet = computed(() => currentBreakpoint.value.name === 'md')
  const isDesktop = computed(() => currentBreakpoint.value.name === 'lg' || currentBreakpoint.value.name === 'xl' || currentBreakpoint.value.name === '2xl')
  
  const shouldShowPanel = computed(() => ({
    context: currentBreakpoint.value.features.showContextArea,
    results: currentBreakpoint.value.features.showResultsArea,
    statusBar: currentBreakpoint.value.features.showStatusBar
  }))
  
  const layoutMode = computed(() => {
    if (currentBreakpoint.value.features.stackedLayout) {
      return 'stacked'
    }
    return currentBreakpoint.value.columns === 3 ? 'three-column' : 'two-column'
  })
  
  // Panel size calculations
  const calculateOptimalPanelSizes = () => {
    const availableWidth = windowWidth.value
    const bp = currentBreakpoint.value
    
    if (bp.features.stackedLayout) {
      // In mobile/stacked mode, panels take full width
      return {
        contextPanel: Math.min(availableWidth - 40, 400), // 40px for margins
        resultsPanel: Math.min(availableWidth - 40, 400)
      }
    }
    
    // Calculate optimal sizes based on available space
    const minCentralPanel = 600
    const totalPanelWidth = bp.panelSizes.contextPanel + bp.panelSizes.resultsPanel
    const remainingWidth = availableWidth - totalPanelWidth - minCentralPanel
    
    if (remainingWidth < 0) {
      // Need to shrink panels
      const scale = (availableWidth - minCentralPanel) / totalPanelWidth
      return {
        contextPanel: Math.max(250, Math.floor(bp.panelSizes.contextPanel * scale)),
        resultsPanel: Math.max(300, Math.floor(bp.panelSizes.resultsPanel * scale))
      }
    }
    
    return bp.panelSizes
  }
  
  const optimalPanelSizes = computed(() => calculateOptimalPanelSizes())
  
  // Auto-adjust workspace layout based on breakpoint
  const applyResponsiveLayout = () => {
    const bp = currentBreakpoint.value
    const sizes = optimalPanelSizes.value
    
    // Update workspace store with responsive settings
    workspaceStore.updateLayout({
      contextPanelWidth: sizes.contextPanel,
      resultsPanelWidth: sizes.resultsPanel,
      panelVisibility: {
        contextArea: bp.features.showContextArea && !workspaceStore.isTransitioning,
        resultsArea: bp.features.showResultsArea && !workspaceStore.isTransitioning,
        console: workspaceStore.layout.panelVisibility.console,
        settings: workspaceStore.layout.panelVisibility.settings
      },
      headerBarVisible: true, // Always visible
      statusBarVisible: bp.features.showStatusBar
    })
    
    // Update preferences for mobile optimization
    if (isMobile.value) {
      workspaceStore.updatePreferences({
        compactMode: true,
        enableAnimations: false, // Disable for performance
        autoContextBuild: false, // Prevent automatic operations on mobile
        smartSuggestions: false
      })
    } else {
      workspaceStore.updatePreferences({
        compactMode: false,
        enableAnimations: true,
        autoContextBuild: true,
        smartSuggestions: true
      })
    }
  }
  
  // Media query helpers
  const matchesMediaQuery = (query: string) => {
    return window.matchMedia(query).matches
  }
  
  const isLandscape = computed(() => windowWidth.value > windowHeight.value)
  const isPortrait = computed(() => windowWidth.value <= windowHeight.value)
  
  // Device detection
  const deviceType = computed(() => {
    if (isMobile.value) return 'mobile'
    if (isTablet.value) return 'tablet'
    return 'desktop'
  })
  
  // Panel collapse helpers for small screens
  const getCollapsedPanels = () => {
    if (!currentBreakpoint.value.features.collapsiblePanels) return []
    
    const collapsed = []
    if (isMobile.value) {
      // On mobile, only show one panel at a time
      if (workspaceStore.layout.panelVisibility.contexArea && workspaceStore.layout.panelVisibility.resultsArea) {
        collapsed.push('resultsArea')
      }
    }
    return collapsed
  }
  
  // Adaptive component sizing
  const getAdaptiveSize = (baseSize: number, scaleFactor = 1) => {
    const bp = currentBreakpoint.value
    if (bp.name === 'xs') return Math.floor(baseSize * 0.8 * scaleFactor)
    if (bp.name === 'sm') return Math.floor(baseSize * 0.9 * scaleFactor)
    if (bp.name === 'md') return Math.floor(baseSize * 0.95 * scaleFactor)
    return Math.floor(baseSize * scaleFactor)
  }
  
  // Font scaling for better readability
  const getFontScale = () => {
    if (isMobile.value) return 0.9
    if (isTablet.value) return 0.95
    return 1
  }
  
  // Touch-friendly sizing
  const getTouchTargetSize = (baseSize: number) => {
    // Ensure minimum touch target size of 44px on mobile
    if (isMobile.value) return Math.max(44, baseSize)
    return baseSize
  }
  
  // Event handlers
  const handleResize = () => {
    windowWidth.value = window.innerWidth
    windowHeight.value = window.innerHeight
    
    // Apply responsive layout after a short delay to avoid too frequent updates
    clearTimeout(resizeTimeout)
    resizeTimeout = setTimeout(applyResponsiveLayout, 150)
  }
  
  let resizeTimeout: NodeJS.Timeout
  
  // Lifecycle
  onMounted(() => {
    window.addEventListener('resize', handleResize, { passive: true })
    window.addEventListener('orientationchange', handleResize, { passive: true })
    
    // Apply initial responsive layout
    applyResponsiveLayout()
  })
  
  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
    window.removeEventListener('orientationchange', handleResize)
    clearTimeout(resizeTimeout)
  })
  
  return {
    // Reactive properties
    windowWidth,
    windowHeight,
    currentBreakpoint,
    
    // Computed properties
    isMobile,
    isTablet,
    isDesktop,
    isLandscape,
    isPortrait,
    deviceType,
    layoutMode,
    shouldShowPanel,
    optimalPanelSizes,
    
    // Methods
    applyResponsiveLayout,
    getCollapsedPanels,
    getAdaptiveSize,
    getFontScale,
    getTouchTargetSize,
    matchesMediaQuery,
    
    // Constants
    BREAKPOINTS
  }
}