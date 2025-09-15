/**
 * Layout Management Policy
 * 
 * Domain service that encapsulates layout management business logic,
 * panel configurations, and responsive design rules following DDD principles.
 */

import { APP_CONFIG } from '@/config/app-config';
import type { LocalStorageService } from './LocalStorageService';
import { defaultLocalStorageService } from './LocalStorageService';

export type WorkspaceMode = 'manual' | 'autonomous';
export type PanelDockPosition = 'left' | 'right' | 'top' | 'bottom' | 'float';
export type LayoutPreset = 'default' | 'focus' | 'debug' | 'presentation' | 'minimal' | 'custom';
export type ViewportSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'xxl';

export interface PanelState {
  id: string;
  visible: boolean;
  position: { x: number; y: number; width: number; height: number };
  docked: PanelDockPosition;
  collapsed: boolean;
  minimized: boolean;
  zIndex: number;
  persistent: boolean;
}

export interface LayoutState {
  // Legacy support
  contextPanelWidth: number;
  resultsPanelWidth: number;
  panelVisibility: Record<string, boolean>;
  headerBarVisible: boolean;
  statusBarVisible: boolean;
  
  // Enhanced layout
  panels: Record<string, PanelState>;
  currentPreset: LayoutPreset;
  customLayouts: Record<string, unknown>;
  splitViewEnabled: boolean;
  gridLayout: boolean;
  adaptiveLayout: boolean;
}

export interface PanelConfiguration {
  minWidth: number;
  maxWidth: number;
  defaultWidth: number;
  minHeight?: number;
  maxHeight?: number;
  defaultHeight?: number;
}

export interface LayoutValidationResult {
  isValid: boolean;
  violations: string[];
  adjustedLayout?: Partial<LayoutState>;
  suggestions: string[];
}

export interface ResponsiveBreakpoints {
  xs: number;
  sm: number;
  md: number;
  lg: number;
  xl: number;
  xxl: number;
}

export class LayoutManagementPolicy {
  private storageService: LocalStorageService;
  private breakpoints: ResponsiveBreakpoints;

  constructor(storageService: LocalStorageService) {
    this.storageService = storageService;
    this.breakpoints = APP_CONFIG.ui.responsive.BREAKPOINTS;
  }

  /**
   * Get panel configuration based on centralized config
   */
  getPanelConfiguration(): Record<string, PanelConfiguration> {
    return {
      contextArea: {
        minWidth: APP_CONFIG.ui.panels.CONTEXT_PANEL_MIN_WIDTH,
        maxWidth: APP_CONFIG.ui.panels.CONTEXT_PANEL_MAX_WIDTH,
        defaultWidth: APP_CONFIG.ui.workspace.DEFAULT_CONTEXT_PANEL_WIDTH
      },
      resultsArea: {
        minWidth: APP_CONFIG.ui.panels.RESULTS_PANEL_MIN_WIDTH,
        maxWidth: APP_CONFIG.ui.panels.RESULTS_PANEL_MAX_WIDTH,
        defaultWidth: APP_CONFIG.ui.workspace.DEFAULT_RESULTS_PANEL_WIDTH
      }
    };
  }

  /**
   * Get adaptive panel configuration based on viewport size
   */
  getAdaptivePanelConfiguration(viewportSize: ViewportSize): Record<string, PanelConfiguration> {
    const baseConfig = this.getPanelConfiguration();

    switch (viewportSize) {
      case 'xs':
      case 'sm':
        return {
          contextArea: {
            ...baseConfig.contextArea,
            minWidth: 250,
            maxWidth: 400,
            defaultWidth: 300
          },
          resultsArea: {
            ...baseConfig.resultsArea,
            minWidth: 300,
            maxWidth: 500,
            defaultWidth: 400
          }
        };

      case 'md':
        return {
          contextArea: {
            ...baseConfig.contextArea,
            minWidth: 280,
            maxWidth: 500,
            defaultWidth: 350
          },
          resultsArea: {
            ...baseConfig.resultsArea,
            minWidth: 350,
            maxWidth: 700,
            defaultWidth: 450
          }
        };

      default:
        return baseConfig;
    }
  }

  /**
   * Determine viewport size based on window width
   */
  getViewportSize(windowWidth: number): ViewportSize {
    if (windowWidth >= this.breakpoints.xxl) return 'xxl';
    if (windowWidth >= this.breakpoints.xl) return 'xl';
    if (windowWidth >= this.breakpoints.lg) return 'lg';
    if (windowWidth >= this.breakpoints.md) return 'md';
    if (windowWidth >= this.breakpoints.sm) return 'sm';
    return 'xs';
  }

  /**
   * Check if layout is mobile-friendly
   */
  isMobileLayout(viewportSize: ViewportSize): boolean {
    return ['xs', 'sm'].includes(viewportSize);
  }

  /**
   * Check if layout is tablet-friendly
   */
  isTabletLayout(viewportSize: ViewportSize): boolean {
    return viewportSize === 'md';
  }

  /**
   * Check if layout is desktop-friendly
   */
  isDesktopLayout(viewportSize: ViewportSize): boolean {
    return ['lg', 'xl', 'xxl'].includes(viewportSize);
  }

  /**
   * Validate layout configuration
   */
  validateLayout(layout: LayoutState, windowWidth: number, windowHeight: number): LayoutValidationResult {
    const violations: string[] = [];
    const suggestions: string[] = [];
    const adjustedLayout: Partial<LayoutState> = {};

    const viewportSize = this.getViewportSize(windowWidth);
    const config = this.getAdaptivePanelConfiguration(viewportSize);

    // Validate panel widths
    if (layout.contextPanelWidth < config.contextArea.minWidth) {
      violations.push(`Context panel width ${layout.contextPanelWidth}px is below minimum ${config.contextArea.minWidth}px`);
      adjustedLayout.contextPanelWidth = config.contextArea.minWidth;
      suggestions.push('Context panel width adjusted to minimum allowed value');
    }

    if (layout.contextPanelWidth > config.contextArea.maxWidth) {
      violations.push(`Context panel width ${layout.contextPanelWidth}px exceeds maximum ${config.contextArea.maxWidth}px`);
      adjustedLayout.contextPanelWidth = config.contextArea.maxWidth;
      suggestions.push('Context panel width adjusted to maximum allowed value');
    }

    if (layout.resultsPanelWidth < config.resultsArea.minWidth) {
      violations.push(`Results panel width ${layout.resultsPanelWidth}px is below minimum ${config.resultsArea.minWidth}px`);
      adjustedLayout.resultsPanelWidth = config.resultsArea.minWidth;
      suggestions.push('Results panel width adjusted to minimum allowed value');
    }

    if (layout.resultsPanelWidth > config.resultsArea.maxWidth) {
      violations.push(`Results panel width ${layout.resultsPanelWidth}px exceeds maximum ${config.resultsArea.maxWidth}px`);
      adjustedLayout.resultsPanelWidth = config.resultsArea.maxWidth;
      suggestions.push('Results panel width adjusted to maximum allowed value');
    }

    // Validate total layout doesn't exceed window width
    const minCenterWidth = APP_CONFIG.ui.layout.MIN_CENTER_AREA_WIDTH;
    const totalPanelWidth = layout.contextPanelWidth + layout.resultsPanelWidth;
    const maxAllowedPanelWidth = windowWidth - minCenterWidth;

    if (totalPanelWidth > maxAllowedPanelWidth) {
      violations.push(`Total panel width ${totalPanelWidth}px exceeds available space ${maxAllowedPanelWidth}px`);
      
      // Proportionally scale down panels
      const scale = maxAllowedPanelWidth / totalPanelWidth;
      adjustedLayout.contextPanelWidth = Math.floor(layout.contextPanelWidth * scale);
      adjustedLayout.resultsPanelWidth = Math.floor(layout.resultsPanelWidth * scale);
      suggestions.push('Panel widths scaled down proportionally to fit available space');
    }

    // Validate mobile layout constraints
    if (this.isMobileLayout(viewportSize)) {
      if (layout.splitViewEnabled) {
        violations.push('Split view is not recommended on mobile devices');
        adjustedLayout.splitViewEnabled = false;
        suggestions.push('Consider disabling split view on mobile for better usability');
      }

      if (Object.values(layout.panelVisibility).filter(Boolean).length > 2) {
        violations.push('Too many panels visible on mobile layout');
        suggestions.push('Consider hiding some panels on mobile for better performance');
      }
    }

    return {
      isValid: violations.length === 0,
      violations,
      adjustedLayout: Object.keys(adjustedLayout).length > 0 ? adjustedLayout : undefined,
      suggestions
    };
  }

  /**
   * Calculate optimal layout for current viewport
   */
  calculateOptimalLayout(
    currentLayout: LayoutState,
    windowWidth: number,
    windowHeight: number,
    mode: WorkspaceMode
  ): LayoutState {
    const viewportSize = this.getViewportSize(windowWidth);
    const config = this.getAdaptivePanelConfiguration(viewportSize);

    let optimalLayout: LayoutState = { ...currentLayout };

    // Apply mode-specific defaults
    if (mode === 'autonomous') {
      optimalLayout.contextPanelWidth = config.contextArea.defaultWidth;
      optimalLayout.resultsPanelWidth = Math.max(config.resultsArea.defaultWidth, 600);
    } else {
      optimalLayout.contextPanelWidth = config.contextArea.defaultWidth;
      optimalLayout.resultsPanelWidth = config.resultsArea.defaultWidth;
    }

    // Apply viewport-specific optimizations
    if (this.isMobileLayout(viewportSize)) {
      optimalLayout.splitViewEnabled = false;
      optimalLayout.gridLayout = false;
      optimalLayout.panelVisibility = {
        ...optimalLayout.panelVisibility,
        contextArea: true,
        resultsArea: true,
        console: false,
        settings: false
      };
    } else if (this.isTabletLayout(viewportSize)) {
      optimalLayout.contextPanelWidth = Math.min(windowWidth * 0.4, config.contextArea.maxWidth);
      optimalLayout.resultsPanelWidth = Math.min(windowWidth * 0.5, config.resultsArea.maxWidth);
    }

    // Validate and adjust if necessary
    const validation = this.validateLayout(optimalLayout, windowWidth, windowHeight);
    if (validation.adjustedLayout) {
      optimalLayout = { ...optimalLayout, ...validation.adjustedLayout };
    }

    return optimalLayout;
  }

  /**
   * Clamp panel width to valid range
   */
  clampPanelWidth(panelType: 'context' | 'results', width: number, viewportSize: ViewportSize): number {
    const config = this.getAdaptivePanelConfiguration(viewportSize);
    const panelConfig = panelType === 'context' ? config.contextArea : config.resultsArea;

    return Math.min(Math.max(width, panelConfig.minWidth), panelConfig.maxWidth);
  }

  /**
   * Get default layout for mode
   */
  getDefaultLayoutForMode(mode: WorkspaceMode, viewportSize: ViewportSize): LayoutState {
    const config = this.getAdaptivePanelConfiguration(viewportSize);

    const baseLayout: LayoutState = {
      contextPanelWidth: config.contextArea.defaultWidth,
      resultsPanelWidth: config.resultsArea.defaultWidth,
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
    };

    // Mode-specific adjustments
    if (mode === 'autonomous') {
      baseLayout.resultsPanelWidth = Math.max(baseLayout.resultsPanelWidth, 600);
      baseLayout.panelVisibility.console = true;
    }

    return baseLayout;
  }

  /**
   * Save layout to storage
   */
  async saveLayout(mode: WorkspaceMode, layout: LayoutState): Promise<boolean> {
    const key = `layout_${mode}`;
    return this.storageService.set(key, layout);
  }

  /**
   * Load layout from storage
   */
  async loadLayout(mode: WorkspaceMode, viewportSize: ViewportSize): Promise<LayoutState> {
    const key = `layout_${mode}`;
    const savedLayout = this.storageService.get<LayoutState>(key);

    if (savedLayout) {
      return savedLayout;
    }

    // Return default layout if none saved
    return this.getDefaultLayoutForMode(mode, viewportSize);
  }

  /**
   * Get layout CSS classes based on state
   */
  getLayoutClasses(
    mode: WorkspaceMode,
    layout: LayoutState,
    viewportSize: ViewportSize,
    isTransitioning: boolean,
    windowState: {
      isMaximized: boolean;
      isFullscreen: boolean;
      isFocusMode: boolean;
    },
    preferences: {
      compactMode: boolean;
      enableAnimations: boolean;
      highContrastMode: boolean;
      reducedMotion: boolean;
    }
  ): Record<string, boolean> {
    return {
      'workspace-manual': mode === 'manual',
      'workspace-autonomous': mode === 'autonomous',
      'workspace-transitioning': isTransitioning,
      'workspace-compact': preferences.compactMode,
      'workspace-animated': preferences.enableAnimations && !preferences.reducedMotion,
      'workspace-focus': windowState.isFocusMode,
      'workspace-fullscreen': windowState.isFullscreen,
      'workspace-maximized': windowState.isMaximized,
      'workspace-mobile': this.isMobileLayout(viewportSize),
      'workspace-tablet': this.isTabletLayout(viewportSize),
      'workspace-desktop': this.isDesktopLayout(viewportSize),
      'workspace-high-contrast': preferences.highContrastMode,
      'workspace-reduced-motion': preferences.reducedMotion,
      'workspace-grid': layout.gridLayout,
      'workspace-split': layout.splitViewEnabled,
      'workspace-adaptive': layout.adaptiveLayout
    };
  }

  /**
   * Calculate available workspace area
   */
  calculateAvailableWorkspaceArea(
    windowWidth: number,
    windowHeight: number,
    layout: LayoutState
  ): { width: number; height: number } {
    const usedWidth = layout.contextPanelWidth + layout.resultsPanelWidth;
    const usedHeight = (layout.headerBarVisible ? 60 : 0) + (layout.statusBarVisible ? 30 : 0);

    return {
      width: windowWidth - usedWidth,
      height: windowHeight - usedHeight
    };
  }

  /**
   * Auto-arrange panels based on viewport
   */
  autoArrangePanels(
    currentLayout: LayoutState,
    windowWidth: number,
    windowHeight: number
  ): LayoutState {
    if (!currentLayout.adaptiveLayout) {
      return currentLayout;
    }

    const viewportSize = this.getViewportSize(windowWidth);
    return this.calculateOptimalLayout(currentLayout, windowWidth, windowHeight, 'manual');
  }

  /**
   * Handle panel resize with constraints
   */
  handlePanelResize(
    panelType: 'context' | 'results',
    newWidth: number,
    currentLayout: LayoutState,
    windowWidth: number,
    viewportSize: ViewportSize
  ): { width: number; isValid: boolean; reason?: string } {
    const clampedWidth = this.clampPanelWidth(panelType, newWidth, viewportSize);
    
    // Check if resize would violate layout constraints
    const testLayout = { ...currentLayout };
    if (panelType === 'context') {
      testLayout.contextPanelWidth = clampedWidth;
    } else {
      testLayout.resultsPanelWidth = clampedWidth;
    }

    const validation = this.validateLayout(testLayout, windowWidth, 0);
    
    if (!validation.isValid) {
      const adjustedWidth = validation.adjustedLayout?.[`${panelType}PanelWidth`] || clampedWidth;
      return {
        width: adjustedWidth,
        isValid: false,
        reason: validation.violations[0]
      };
    }

    return {
      width: clampedWidth,
      isValid: true
    };
  }
}

// Default instance for dependency injection
export const createLayoutManagementPolicy = (storageService: LocalStorageService) => 
  new LayoutManagementPolicy(storageService);

// Add default instance export
export const defaultLayoutManagementPolicy = new LayoutManagementPolicy(defaultLocalStorageService);
