import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useWorkspaceStore } from '@/stores/workspace.store'
import { APP_CONFIG } from '@/config/app-config'

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
    maxWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.XS,
    columns: 1,
    panelSizes: {
      contextPanel: APP_CONFIG.ui.panels.CONTEXT_PANEL_MIN_WIDTH,
      resultsPanel: APP_CONFIG.ui.panels.RESULTS_PANEL_MIN_WIDTH
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
    minWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.XS,
    maxWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.SM,
    columns: 1,
    panelSizes: {
      contextPanel: APP_CONFIG.ui.panels.CONTEXT_PANEL_MIN_WIDTH + 20,
      resultsPanel: APP_CONFIG.ui.panels.RESULTS_PANEL_MIN_WIDTH - 80
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
    minWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.SM,
    maxWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.MD,
    columns: 2,
    panelSizes: {
      contextPanel: APP_CONFIG.ui.panels.CONTEXT_PANEL_MIN_WIDTH + 50,
      resultsPanel: APP_CONFIG.ui.panels.RESULTS_PANEL_MIN_WIDTH
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
    minWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.MD,
    maxWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.LG,
    columns: 3,
    panelSizes: {
      contextPanel: APP_CONFIG.ui.workspace.DEFAULT_CONTEXT_PANEL_WIDTH,
      resultsPanel: APP_CONFIG.ui.workspace.DEFAULT_RESULTS_PANEL_WIDTH
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
    minWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.LG,
    maxWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.XL,
    columns: 3,
    panelSizes: {
      contextPanel: APP_CONFIG.ui.workspace.DEFAULT_CONTEXT_PANEL_WIDTH + 50,
      resultsPanel: APP_CONFIG.ui.workspace.DEFAULT_RESULTS_PANEL_WIDTH + 50
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
    minWidth: APP_CONFIG.ui.responsive.BREAKPOINTS.XL,
    columns: 3,
    panelSizes: {
      contextPanel: APP_CONFIG.ui.panels.CONTEXT_PANEL_MAX_WIDTH - 100,
      resultsPanel: APP_CONFIG.ui.panels.RESULTS_PANEL_MAX_WIDTH - 100
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
    const responsiveMargin = APP_CONFIG.ui.layout.RESPONSIVE_MARGIN
    
    if (bp.features.stackedLayout) {
      // In mobile/stacked mode, panels take full width
      return {
        contextPanel: Math.min(availableWidth - responsiveMargin, APP_CONFIG.ui.panels.CONTEXT_PANEL_MAX_WIDTH),
        resultsPanel: Math.min(availableWidth - responsiveMargin, APP_CONFIG.ui.panels.RESULTS_PANEL_MAX_WIDTH)
      }
    }
    
    // Calculate optimal sizes based on available space
    const minCentralPanel = APP_CONFIG.ui.layout.MIN_CENTER_AREA_WIDTH
    const totalPanelWidth = bp.panelSizes.contextPanel + bp.panelSizes.resultsPanel
    const remainingWidth = availableWidth - totalPanelWidth - minCentralPanel
    
    if (remainingWidth < 0) {
      // Need to shrink panels
      const scale = (availableWidth - minCentralPanel) / totalPanelWidth
      return {
        contextPanel: Math.max(APP_CONFIG.ui.panels.CONTEXT_PANEL_MIN_WIDTH, Math.floor(bp.panelSizes.contextPanel * scale)),
        resultsPanel: Math.max(APP_CONFIG.ui.panels.RESULTS_PANEL_MIN_WIDTH, Math.floor(bp.panelSizes.resultsPanel * scale))
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
    workspaceStore.setLayout({
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
      if (workspaceStore.layout.panelVisibility.contextArea && workspaceStore.layout.panelVisibility.resultsArea) {
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
    resizeTimeout = setTimeout(applyResponsiveLayout, APP_CONFIG.events.DEBOUNCE_DELAY)
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