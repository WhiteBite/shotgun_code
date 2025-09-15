import { ref, computed, watch } from 'vue'
import { defineStore } from 'pinia'
import { useLocalStorage, useWindowSize } from '@vueuse/core'

export type WorkspaceMode = 'manual' | 'autonomous'
export type PanelDockPosition = 'left' | 'right' | 'top' | 'bottom' | 'float'
export type LayoutPreset = 'default' | 'focus' | 'debug' | 'presentation' | 'minimal' | 'custom'
export type ViewportSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'xxl'

export interface PanelState {
  id: string
  visible: boolean
  position: { x: number; y: number; width: number; height: number }
  docked: PanelDockPosition
  collapsed: boolean
  minimized: boolean
  zIndex: number
  persistent: boolean
}

export interface WindowState {
  isMaximized: boolean
  isFullscreen: boolean
  isFocusMode: boolean
  isMinimized: boolean
  isAlwaysOnTop: boolean
}

export interface LayoutState {
  // Legacy support
  contextPanelWidth: number
  resultsPanelWidth: number
  panelVisibility: Record<string, boolean>
  headerBarVisible: boolean
  statusBarVisible: boolean
  
  // Enhanced layout
  panels: Record<string, PanelState>
  currentPreset: LayoutPreset
  customLayouts: Record<string, any>
  splitViewEnabled: boolean
  gridLayout: boolean
  adaptiveLayout: boolean
}

export interface WorkspacePreferences {
  autoContextBuild: boolean
  smartSuggestions: boolean
  autoApplyChanges: boolean
  showTokenCount: boolean
  showEstimatedCost: boolean
  enableAnimations: boolean
  compactMode: boolean
  
  // Enhanced preferences
  enableTooltips: boolean
  tooltipDelay: number
  enableSounds: boolean
  autoSaveInterval: number
  enablePanelSnapping: boolean
  enableMagneticEdges: boolean
  gridSnapping: boolean
  gridSize: number
  panelTransparency: number
  enableBlur: boolean
  highContrastMode: boolean
  reducedMotion: boolean
  enableAutoLayout: boolean
  rememberPanelStates: boolean
}

export const useWorkspaceStore = defineStore('workspace', () => {
  // Core State
  const currentMode = ref<WorkspaceMode>('manual')
  const isTransitioning = ref(false)
  
  // Window Management
  const { width: windowWidth, height: windowHeight } = useWindowSize()
  const windowState = ref<WindowState>({
    isMaximized: false,
    isFullscreen: false,
    isFocusMode: false,
    isMinimized: false,
    isAlwaysOnTop: false
  })
  
  // Layout Management
  const layout = ref<LayoutState>({
    // Legacy support
    contextPanelWidth: 400,
    resultsPanelWidth: 500,
    panelVisibility: {
      contextArea: true,
      resultsArea: true,
      console: false,
      settings: false
    },
    headerBarVisible: true,
    statusBarVisible: true,
    
    // Enhanced layout
    panels: {},
    currentPreset: 'default',
    customLayouts: {},
    splitViewEnabled: false,
    gridLayout: false,
    adaptiveLayout: true
  })
  
  // User Preferences with LocalStorage
  const preferences = useLocalStorage<WorkspacePreferences>('workspace_preferences', {
    autoContextBuild: true,
    smartSuggestions: true,
    autoApplyChanges: false,
    showTokenCount: true,
    showEstimatedCost: true,
    enableAnimations: true,
    compactMode: false,
    
    // Enhanced preferences
    enableTooltips: true,
    tooltipDelay: 500,
    enableSounds: false,
    autoSaveInterval: 30000,
    enablePanelSnapping: true,
    enableMagneticEdges: true,
    gridSnapping: false,
    gridSize: 20,
    panelTransparency: 0.95,
    enableBlur: true,
    highContrastMode: false,
    reducedMotion: false,
    enableAutoLayout: true,
    rememberPanelStates: true
  })
  
  // Enhanced responsive breakpoints
  const breakpoints = {
    xs: 480,
    sm: 768,
    md: 1024,
    lg: 1200,
    xl: 1400,
    xxl: 1600
  }
  
  // Current viewport size
  const currentViewport = computed((): ViewportSize => {
    const width = windowWidth.value
    if (width >= breakpoints.xxl) return 'xxl'
    if (width >= breakpoints.xl) return 'xl'
    if (width >= breakpoints.lg) return 'lg'
    if (width >= breakpoints.md) return 'md'
    if (width >= breakpoints.sm) return 'sm'
    return 'xs'
  })
  
  // Enhanced Computed Properties
  const isManualMode = computed(() => currentMode.value === 'manual')
  const isAutonomousMode = computed(() => currentMode.value === 'autonomous')
  const isMobileLayout = computed(() => ['xs', 'sm'].includes(currentViewport.value))
  const isTabletLayout = computed(() => currentViewport.value === 'md')
  const isDesktopLayout = computed(() => ['lg', 'xl', 'xxl'].includes(currentViewport.value))
  
  const layoutClasses = computed(() => ({
    'workspace-manual': isManualMode.value,
    'workspace-autonomous': isAutonomousMode.value,
    'workspace-transitioning': isTransitioning.value,
    'workspace-compact': preferences.value.compactMode,
    'workspace-animated': preferences.value.enableAnimations,
    'workspace-focus': windowState.value.isFocusMode,
    'workspace-fullscreen': windowState.value.isFullscreen,
    'workspace-maximized': windowState.value.isMaximized,
    'workspace-mobile': isMobileLayout.value,
    'workspace-tablet': isTabletLayout.value,
    'workspace-desktop': isDesktopLayout.value,
    'workspace-high-contrast': preferences.value.highContrastMode,
    'workspace-reduced-motion': preferences.value.reducedMotion,
    'workspace-grid': layout.value.gridLayout,
    'workspace-split': layout.value.splitViewEnabled
  }))
  
  // Missing computed properties
  const availableWorkspaceArea = computed(() => ({
    width: windowWidth.value - (layout.value.contextPanelWidth + layout.value.resultsPanelWidth),
    height: windowHeight.value - (layout.value.headerBarVisible ? 60 : 0) - (layout.value.statusBarVisible ? 30 : 0)
  }))
  
  const activePanels = computed(() => {
    return Object.entries(layout.value.panelVisibility)
      .filter(([_, visible]) => visible)
      .map(([panelName]) => panelName)
  })
  
  const dockedPanels = computed(() => {
    return Object.entries(layout.value.panels)
      .filter(([_, panel]) => panel.docked !== 'float')
      .map(([panelId]) => panelId)
  })
  
  const panelConfiguration = computed(() => ({
    contextArea: {
      minWidth: 300,
      maxWidth: 800,
      defaultWidth: 400
    },
    resultsArea: {
      minWidth: 400,
      maxWidth: 1000,
      defaultWidth: 500
    }
  }))
  
  const adaptivePanelConfiguration = computed(() => {
    const config = panelConfiguration.value
    const viewport = currentViewport.value
    
    // Adjust configuration based on viewport
    if (viewport === 'xs' || viewport === 'sm') {
      return {
        contextArea: {
          ...config.contextArea,
          minWidth: 250,
          maxWidth: 400
        },
        resultsArea: {
          ...config.resultsArea,
          minWidth: 300,
          maxWidth: 500
        }
      }
    }
    
    return config
  })
  
  // Window State Actions
  const toggleFullscreen = () => {
    windowState.value.isFullscreen = !windowState.value.isFullscreen
    if (windowState.value.isFullscreen) {
      windowState.value.isMaximized = false
    }
  }
  
  const toggleMaximize = () => {
    windowState.value.isMaximized = !windowState.value.isMaximized
    if (windowState.value.isMaximized) {
      windowState.value.isFullscreen = false
    }
  }
  
  const toggleFocusMode = () => {
    windowState.value.isFocusMode = !windowState.value.isFocusMode
  }
  
  const setAlwaysOnTop = (enabled: boolean) => {
    windowState.value.isAlwaysOnTop = enabled
  }
  
  // Panel arrangement function
  const autoArrangePanels = () => {
    if (!layout.value.adaptiveLayout) return
    
    const config = panelConfiguration.value
    
    if (isMobileLayout.value) {
      // On mobile, stack panels vertically
      updateLayout({
        contextPanelWidth: Math.min(windowWidth.value * 0.9, config.contextArea.maxWidth),
        resultsPanelWidth: Math.min(windowWidth.value * 0.9, config.resultsArea.maxWidth)
      })
    } else if (isTabletLayout.value) {
      // On tablet, reduce panel widths
      updateLayout({
        contextPanelWidth: Math.min(windowWidth.value * 0.4, config.contextArea.maxWidth),
        resultsPanelWidth: Math.min(windowWidth.value * 0.5, config.resultsArea.maxWidth)
      })
    }
    // Desktop layouts use default values
  }

  // Actions
  const setMode = async (mode: WorkspaceMode) => {
    if (currentMode.value === mode || isTransitioning.value) return
    
    isTransitioning.value = true
    
    try {
      // Store current layout state before switching
      const currentLayoutKey = `layout_${currentMode.value}`
      localStorage.setItem(currentLayoutKey, JSON.stringify(layout.value))
      
      // Switch mode
      currentMode.value = mode
      
      // Restore layout for new mode or use defaults
      const newLayoutKey = `layout_${mode}`
      const savedLayout = localStorage.getItem(newLayoutKey)
      
      if (savedLayout) {
        layout.value = { ...layout.value, ...JSON.parse(savedLayout) }
      } else {
        // Apply default layout for the new mode
        if (mode === 'autonomous') {
          layout.value.contextPanelWidth = 400
          layout.value.resultsPanelWidth = 600
        } else {
          layout.value.contextPanelWidth = 400
          layout.value.resultsPanelWidth = 500
        }
      }
      
      // Animate transition if enabled
      if (preferences.value.enableAnimations) {
        await new Promise(resolve => setTimeout(resolve, 300))
      }
      
    } finally {
      isTransitioning.value = false
    }
  }
  
  const toggleMode = async () => {
    const newMode = currentMode.value === 'manual' ? 'autonomous' : 'manual'
    await setMode(newMode)
  }
  
  const setLayout = (newLayout: Partial<LayoutState>) => {
    // Fix for panel sizing when resizing the window
    layout.value = {
      ...layout.value,
      ...newLayout
    };
    
    // Ensure panel widths don't exceed available space
    const totalWidth = windowWidth.value;
    const minCenterWidth = 300; // Minimum width for center panel
    const maxTotalSidePanelsWidth = totalWidth - minCenterWidth;
    
    if (layout.value.contextPanelWidth + layout.value.resultsPanelWidth > maxTotalSidePanelsWidth) {
      // Scale down both panels proportionally
      const ratio = maxTotalSidePanelsWidth / (layout.value.contextPanelWidth + layout.value.resultsPanelWidth);
      layout.value.contextPanelWidth = Math.floor(layout.value.contextPanelWidth * ratio);
      layout.value.resultsPanelWidth = Math.floor(layout.value.resultsPanelWidth * ratio);
    }
    
    // Persist layout changes
    const layoutKey = `layout_${currentMode.value}`;
    localStorage.setItem(layoutKey, JSON.stringify(layout.value));
  };
  

  const updatePreferences = (updates: Partial<WorkspacePreferences>) => {
    preferences.value = { ...preferences.value, ...updates }
    localStorage.setItem('workspace_preferences', JSON.stringify(preferences.value))
  }
  
  const togglePanelVisibility = (panelName: string) => {
    const newVisibility = {
      ...layout.value.panelVisibility,
      [panelName]: !layout.value.panelVisibility[panelName]
    }
    
    updateLayout({ panelVisibility: newVisibility })
  }
  
  const setPanelWidth = (panel: 'context' | 'results', width: number) => {
    const config = panelConfiguration.value
    const panelConfig = panel === 'context' ? config.contextArea : config.resultsArea
    
    // Clamp width to min/max values
    const clampedWidth = Math.min(
      Math.max(width, panelConfig.minWidth), 
      panelConfig.maxWidth
    )
    
    if (panel === 'context') {
      updateLayout({ contextPanelWidth: clampedWidth })
    } else {
      updateLayout({ resultsPanelWidth: clampedWidth })
    }
  }
  
  const resetLayout = () => {
    const defaultLayout: LayoutState = {
      contextPanelWidth: 400,
      resultsPanelWidth: 500,
      panelVisibility: {
        contextArea: true,
        resultsArea: true,
        console: false,
        settings: false
      },
      headerBarVisible: true,
      statusBarVisible: true,
      // Add missing properties
      panels: {},
      currentPreset: 'default',
      customLayouts: {},
      splitViewEnabled: false,
      gridLayout: false,
      adaptiveLayout: true
    }
    
    layout.value = defaultLayout
    localStorage.removeItem(`layout_${currentMode.value}`)
  }
  
  // Enhanced initialization
  const initializeStore = () => {
    // Load saved mode
    const savedMode = localStorage.getItem('workspace_mode') as WorkspaceMode
    if (savedMode && (savedMode === 'manual' || savedMode === 'autonomous')) {
      currentMode.value = savedMode
    }
    
    // Load saved layout for current mode
    const savedLayout = localStorage.getItem(`layout_${currentMode.value}`)
    if (savedLayout) {
      try {
        const parsedLayout = JSON.parse(savedLayout)
        layout.value = { ...layout.value, ...parsedLayout }
      } catch (error) {
        console.warn('Failed to parse saved layout:', error)
      }
    }
    
    // Load window state
    const savedWindowState = localStorage.getItem('window_state')
    if (savedWindowState) {
      try {
        const parsedWindowState = JSON.parse(savedWindowState)
        windowState.value = { ...windowState.value, ...parsedWindowState }
      } catch (error) {
        console.warn('Failed to parse saved window state:', error)
      }
    }
    
    // Apply adaptive layout if enabled
    if (layout.value.adaptiveLayout) {
      autoArrangePanels()
    }
  }
  
  // Persistence helper
  const persistMode = () => {
    localStorage.setItem('workspace_mode', currentMode.value)
  }
  
  // Responsive helpers
  const getBreakpoint = (width: number): string => {
    if (width >= breakpoints.xl) return 'xl'
    if (width >= breakpoints.lg) return 'lg'
    if (width >= breakpoints.md) return 'md'
    if (width >= breakpoints.sm) return 'sm'
    return 'xs'
  }
  
  const isResponsiveLayout = (breakpoint: string): boolean => {
    const width = window.innerWidth
    return getBreakpoint(width) === breakpoint
  }
  
  // Watch for window size changes and adjust layout
  watch([windowWidth, windowHeight], () => {
    if (layout.value.adaptiveLayout) {
      // Add a small delay to ensure DOM has updated
      setTimeout(() => {
        autoArrangePanels();
      }, 50);
    }
    
    // Check if panels are too wide for current window
    const totalWidth = windowWidth.value;
    const minCenterWidth = 300; // Minimum width for center panel
    const maxTotalSidePanelsWidth = totalWidth - minCenterWidth;
    
    if (layout.value.contextPanelWidth + layout.value.resultsPanelWidth > maxTotalSidePanelsWidth) {
      // Scale down both panels proportionally
      const ratio = maxTotalSidePanelsWidth / (layout.value.contextPanelWidth + layout.value.resultsPanelWidth);
      layout.value.contextPanelWidth = Math.floor(layout.value.contextPanelWidth * ratio);
      layout.value.resultsPanelWidth = Math.floor(layout.value.resultsPanelWidth * ratio);
      
      // Persist changes
      const layoutKey = `layout_${currentMode.value}`;
      localStorage.setItem(layoutKey, JSON.stringify(layout.value));
    }
  }, { immediate: true });

  return {
    // State
    currentMode,
    isTransitioning,
    layout,
    preferences,
    breakpoints,
    windowState,
    windowWidth,
    windowHeight,
    currentViewport,
    
    // Computed
    isManualMode,
    isAutonomousMode,
    isMobileLayout,
    isTabletLayout,
    isDesktopLayout,
    layoutClasses,
    panelConfiguration,
    adaptivePanelConfiguration,
    availableWorkspaceArea,
    activePanels,
    dockedPanels,
    
    // Actions - Mode Management
    setMode,
    toggleMode,
    
    // Actions - Window Management
    toggleFullscreen,
    toggleMaximize,
    toggleFocusMode,
    setAlwaysOnTop,
    
    // Actions - Panel Management
    setLayout,
    updatePreferences,
    togglePanelVisibility,
    setPanelWidth,
    resetLayout,
    
    // Actions - Layout Management
    autoArrangePanels,
    initializeStore,
    persistMode,
    
    // Actions - Persistence
    getBreakpoint,
    isResponsiveLayout
  }
})