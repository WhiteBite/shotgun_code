import { ref, computed, watch } from 'vue'
import { defineStore } from 'pinia'
import { useWindowSize } from '@vueuse/core'
import { defaultLocalStorageService } from '@/domain/services/LocalStorageService'
import { APP_CONFIG } from '@/config/app-config'

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
  contextPanelWidth: number
  resultsPanelWidth: number
  panelVisibility: Record<string, boolean>
  headerBarVisible: boolean
  statusBarVisible: boolean
  panels: Record<string, PanelState>
  currentPreset: LayoutPreset
  customLayouts: Record<string, unknown>
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
  
  // Layout Management using centralized config
  const layout = ref<LayoutState>({
    contextPanelWidth: APP_CONFIG.ui.workspace.DEFAULT_CONTEXT_PANEL_WIDTH,
    resultsPanelWidth: APP_CONFIG.ui.workspace.DEFAULT_RESULTS_PANEL_WIDTH,
    panelVisibility: {
      contextArea: true,
      resultsArea: true,
      console: false,
      settings: false
    },
    headerBarVisible: true,
    statusBarVisible: true,
    panels: {},
    currentPreset: APP_CONFIG.ui.workspace.DEFAULT_LAYOUT_PRESET as LayoutPreset,
    customLayouts: {},
    splitViewEnabled: false,
    gridLayout: false,
    adaptiveLayout: true
  })
  
  // User Preferences with defaultLocalStorageService
  const preferences = ref<WorkspacePreferences>({
    autoContextBuild: true,
    smartSuggestions: true,
    autoApplyChanges: false,
    showTokenCount: true,
    showEstimatedCost: true,
    enableAnimations: APP_CONFIG.ui.workspace.ENABLE_ANIMATIONS,
    compactMode: false,
    enableTooltips: true,
    tooltipDelay: APP_CONFIG.ui.tooltips.DEFAULT_DELAY,
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
  
  // Load preferences from storage on initialization
  const loadPreferences = () => {
    const savedPreferences = defaultLocalStorageService.getItem<Partial<WorkspacePreferences>>('workspace_preferences')
    if (savedPreferences) {
      preferences.value = { ...preferences.value, ...savedPreferences }
    }
  }
  
  // Save preferences to storage
  const persistPreferences = () => {
    defaultLocalStorageService.setItem('workspace_preferences', preferences.value)
  }
  
  // Enhanced responsive breakpoints from centralized config
  const breakpoints = APP_CONFIG.ui.responsive.BREAKPOINTS
  
  // Current viewport size
  const currentViewport = computed((): ViewportSize => {
    const width = windowWidth.value
    if (width >= breakpoints.XL) return 'xl'
    if (width >= breakpoints.LG) return 'lg'
    if (width >= breakpoints.MD) return 'md'
    if (width >= breakpoints.SM) return 'sm'
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
  
  const availableWorkspaceArea = computed(() => ({
    width: windowWidth.value - (layout.value.contextPanelWidth + layout.value.resultsPanelWidth),
    height: windowHeight.value - (layout.value.headerBarVisible ? 60 : 0) - (layout.value.statusBarVisible ? 30 : 0)
  }))
  
  const activePanels = computed(() => {
    return Object.entries(layout.value.panelVisibility)
      .filter(([name, visible]) => visible)
      .map(([panelName]) => panelName)
  })
  
  const dockedPanels = computed(() => {
    return Object.entries(layout.value.panels)
      .filter(([name, panel]) => panel.docked !== 'float')
      .map(([panelId]) => panelId)
  })
  
  const panelConfiguration = computed(() => ({
    contextArea: {
      name: 'Context Area',
      defaultWidth: isManualMode.value ? 400 : 400,
      minWidth: APP_CONFIG.ui.panels.CONTEXT_PANEL_MIN_WIDTH,
      maxWidth: APP_CONFIG.ui.panels.CONTEXT_PANEL_MAX_WIDTH,
      visible: layout.value.panelVisibility.contextArea
    },
    resultsArea: {
      name: 'Results Area',
      defaultWidth: isManualMode.value ? 500 : 600,
      minWidth: APP_CONFIG.ui.panels.RESULTS_PANEL_MIN_WIDTH,
      maxWidth: APP_CONFIG.ui.panels.RESULTS_PANEL_MAX_WIDTH,
      visible: layout.value.panelVisibility.resultsArea
    }
  }))
  
  const adaptivePanelConfiguration = computed(() => {
    const viewport = currentViewport.value
    const config = { ...panelConfiguration.value }
    
    if (viewport === 'xs' || viewport === 'sm') {
      config.contextArea.defaultWidth = windowWidth.value - 40
      config.resultsArea.defaultWidth = windowWidth.value - 40
    } else if (viewport === 'md') {
      config.contextArea.defaultWidth = Math.min(350, windowWidth.value * 0.35)
      config.resultsArea.defaultWidth = Math.min(450, windowWidth.value * 0.45)
    }
    
    return config
  })
  
  // Persistence methods using LocalStorageService
  const persistLayout = () => {
    const layoutKey = `layout_${currentMode.value}`
    defaultLocalStorageService.setItem(layoutKey, layout.value)
  }
  
  const persistWindowState = () => {
    defaultLocalStorageService.setItem('window_state', windowState.value)
  }
  
  const persistMode = () => {
    defaultLocalStorageService.setItem('workspace_mode', currentMode.value)
  }
  
  const loadStoredData = () => {
    const savedMode = defaultLocalStorageService.getItem<WorkspaceMode>('workspace_mode')
    if (savedMode && (savedMode === 'manual' || savedMode === 'autonomous')) {
      currentMode.value = savedMode
    }
    
    const layoutKey = `layout_${currentMode.value}`
    const savedLayout = defaultLocalStorageService.getItem<Partial<LayoutState>>(layoutKey)
    if (savedLayout) {
      layout.value = { ...layout.value, ...savedLayout }
    }
    
    const savedWindowState = defaultLocalStorageService.getItem<Partial<WindowState>>('window_state')
    if (savedWindowState) {
      windowState.value = { ...windowState.value, ...savedWindowState }
    }
    
    // Load preferences
    loadPreferences()
  }

  // Actions
  const setMode = async (mode: WorkspaceMode) => {
    if (isTransitioning.value || currentMode.value === mode) return
    
    isTransitioning.value = true
    persistLayout()
    currentMode.value = mode
    persistMode()
    
    const layoutKey = `layout_${mode}`
    const savedLayout = defaultLocalStorageService.getItem<Partial<LayoutState>>(layoutKey)
    if (savedLayout) {
      layout.value = { ...layout.value, ...savedLayout }
    }
    
    await new Promise(resolve => setTimeout(resolve, 200))
    isTransitioning.value = false
  }
  
  const toggleMode = async () => {
    const newMode = currentMode.value === 'manual' ? 'autonomous' : 'manual'
    await setMode(newMode)
  }
  
  const toggleFullscreen = () => {
    windowState.value.isFullscreen = !windowState.value.isFullscreen
    persistWindowState()
  }
  
  const toggleMaximize = () => {
    windowState.value.isMaximized = !windowState.value.isMaximized
    persistWindowState()
  }
  
  const toggleFocusMode = () => {
    windowState.value.isFocusMode = !windowState.value.isFocusMode
    persistWindowState()
  }
  
  const setAlwaysOnTop = (alwaysOnTop: boolean) => {
    windowState.value.isAlwaysOnTop = alwaysOnTop
    persistWindowState()
  }
  
  const setLayout = (newLayout: Partial<LayoutState>) => {
    layout.value = { ...layout.value, ...newLayout }
    persistLayout()
  }
  
  const updateLayout = (updates: Partial<LayoutState>) => {
    Object.assign(layout.value, updates)
    persistLayout()
  }
  
  const updatePreferences = (newPreferences: Partial<WorkspacePreferences>) => {
    preferences.value = { ...preferences.value, ...newPreferences }
    persistPreferences()
  }
  
  const togglePanelVisibility = (panelName: string) => {
    layout.value.panelVisibility[panelName] = !layout.value.panelVisibility[panelName]
    persistLayout()
  }
  
  const setPanelWidth = (panel: 'context' | 'results', width: number) => {
    const config = panelConfiguration.value
    
    if (panel === 'context') {
      const minWidth = config.contextArea.minWidth
      const maxWidth = config.contextArea.maxWidth
      layout.value.contextPanelWidth = Math.max(minWidth, Math.min(maxWidth, width))
    } else {
      const minWidth = config.resultsArea.minWidth
      const maxWidth = config.resultsArea.maxWidth
      layout.value.resultsPanelWidth = Math.max(minWidth, Math.min(maxWidth, width))
    }
    
    persistLayout()
  }
  
  const autoArrangePanels = () => {
    if (!layout.value.adaptiveLayout) return
    
    const viewport = currentViewport.value
    const adaptiveConfig = adaptivePanelConfiguration.value
    
    layout.value.contextPanelWidth = adaptiveConfig.contextArea.defaultWidth
    layout.value.resultsPanelWidth = adaptiveConfig.resultsArea.defaultWidth
    
    if (viewport === 'xs' || viewport === 'sm') {
      layout.value.statusBarVisible = false
      layout.value.splitViewEnabled = false
    } else {
      layout.value.statusBarVisible = true
    }
    
    persistLayout()
  }
  
  const resetLayout = () => {
    const defaultLayout: LayoutState = {
      contextPanelWidth: APP_CONFIG.ui.workspace.DEFAULT_CONTEXT_PANEL_WIDTH,
      resultsPanelWidth: APP_CONFIG.ui.workspace.DEFAULT_RESULTS_PANEL_WIDTH,
      panelVisibility: {
        contextArea: true,
        resultsArea: true,
        console: false,
        settings: false
      },
      headerBarVisible: true,
      statusBarVisible: true,
      panels: {},
      currentPreset: APP_CONFIG.ui.workspace.DEFAULT_LAYOUT_PRESET as LayoutPreset,
      customLayouts: {},
      splitViewEnabled: false,
      gridLayout: false,
      adaptiveLayout: true
    }
    
    layout.value = defaultLayout
    defaultLocalStorageService.remove(`layout_${currentMode.value}`)
  }
  
  const initializeStore = () => {
    loadStoredData()
    if (layout.value.adaptiveLayout) {
      autoArrangePanels()
    }
  }
  
  const getBreakpoint = (width: number): string => {
    if (width >= breakpoints.XL) return 'xl'
    if (width >= breakpoints.LG) return 'lg'
    if (width >= breakpoints.MD) return 'md'
    if (width >= breakpoints.SM) return 'sm'
    return 'xs'
  }
  
  const isResponsiveLayout = (breakpoint: string): boolean => {
    const width = window.innerWidth
    return getBreakpoint(width) === breakpoint
  }
  
  // Watch for window size changes and adjust layout
  watch([windowWidth, windowHeight], () => {
    if (layout.value.adaptiveLayout) {
      setTimeout(() => {
        autoArrangePanels();
      }, 50);
    }
    
    const totalWidth = windowWidth.value;
    const minCenterWidth = 300;
    const maxTotalSidePanelsWidth = totalWidth - minCenterWidth;
    
    if (layout.value.contextPanelWidth + layout.value.resultsPanelWidth > maxTotalSidePanelsWidth) {
      const ratio = maxTotalSidePanelsWidth / (layout.value.contextPanelWidth + layout.value.resultsPanelWidth);
      layout.value.contextPanelWidth = Math.floor(layout.value.contextPanelWidth * ratio);
      layout.value.resultsPanelWidth = Math.floor(layout.value.resultsPanelWidth * ratio);
      persistLayout();
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
    
    // Actions
    setMode,
    toggleMode,
    toggleFullscreen,
    toggleMaximize,
    toggleFocusMode,
    setAlwaysOnTop,
    setLayout,
    updateLayout,
    updatePreferences,
    togglePanelVisibility,
    setPanelWidth,
    resetLayout,
    autoArrangePanels,
    initializeStore,
    persistMode,
    persistLayout,
    persistWindowState,
    persistPreferences,
    loadStoredData,
    loadPreferences,
    getBreakpoint,
    isResponsiveLayout
  }
})