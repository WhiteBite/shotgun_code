/**
 * Workspace View Model
 * Manages overall workspace state and layout
 */

import { ref, computed, reactive, watch } from 'vue';
import { useProjectStore } from '../stores/project.clean.store';
import { WorkspaceMode } from '../domain/value-objects';
import { WorkspaceState } from '../domain/entities';

export interface PanelState {
  id: string;
  isVisible: boolean;
  isCollapsed: boolean;
  width: number;
  minWidth: number;
  maxWidth: number;
  position: 'left' | 'right' | 'top' | 'bottom';
}

export interface WorkspaceViewModel {
  // State
  readonly currentMode: WorkspaceMode;
  readonly isProjectLoaded: boolean;
  readonly panels: Record<string, PanelState>;
  readonly layoutConfig: {
    leftPanelWidth: number;
    rightPanelWidth: number;
    topSectionHeight: number;
  };
  
  // Actions
  switchMode(mode: WorkspaceMode): void;
  togglePanel(panelId: string): void;
  resizePanel(panelId: string, width: number): void;
  resetLayout(): void;
  savePanelState(panelId: string, state: Partial<PanelState>): void;
  loadPanelState(panelId: string): PanelState | null;
  isValidForMode(mode: WorkspaceMode): boolean;
}

export function createWorkspaceViewModel(): WorkspaceViewModel {
  const projectStore = useProjectStore();
  
  // Local state
  const currentMode = ref<WorkspaceMode>(WorkspaceMode.FILE_EXPLORER);
  
  const panels = reactive<Record<string, PanelState>>({
    files: {
      id: 'files',
      isVisible: true,
      isCollapsed: false,
      width: 320,
      minWidth: 280,
      maxWidth: 480,
      position: 'left'
    },
    context: {
      id: 'context',
      isVisible: true,
      isCollapsed: false,
      width: 400,
      minWidth: 320,
      maxWidth: 600,
      position: 'right'
    },
    reports: {
      id: 'reports',
      isVisible: true,
      isCollapsed: false,
      width: 380,
      minWidth: 320,
      maxWidth: 500,
      position: 'right'
    },
    autonomous: {
      id: 'autonomous',
      isVisible: false,
      isCollapsed: false,
      width: 350,
      minWidth: 300,
      maxWidth: 450,
      position: 'right'
    }
  });

  const layoutConfig = reactive({
    leftPanelWidth: 320,
    rightPanelWidth: 380,
    topSectionHeight: 60
  });

  // Computed
  const isProjectLoaded = computed(() => projectStore.hasProject);

  // Actions
  const switchMode = (mode: WorkspaceMode): void => {
    if (!isValidForMode(mode)) {
      console.warn(`Cannot switch to mode ${mode}: validation failed`);
      return;
    }

    currentMode.value = mode;
    projectStore.switchWorkspaceMode(mode);
    
    // Adjust panel visibility based on mode
    switch (mode) {
      case WorkspaceMode.FILE_EXPLORER:
        panels.files.isVisible = true;
        panels.context.isVisible = false;
        panels.reports.isVisible = false;
        panels.autonomous.isVisible = false;
        break;
        
      case WorkspaceMode.CONTEXT_BUILDER:
        panels.files.isVisible = true;
        panels.context.isVisible = true;
        panels.reports.isVisible = false;
        panels.autonomous.isVisible = false;
        break;
        
      case WorkspaceMode.CODE_GENERATOR:
        panels.files.isVisible = true;
        panels.context.isVisible = true;
        panels.reports.isVisible = true;
        panels.autonomous.isVisible = false;
        break;
        
      case WorkspaceMode.TASK_COMPOSER:
        panels.files.isVisible = true;
        panels.context.isVisible = true;
        panels.reports.isVisible = true;
        panels.autonomous.isVisible = true;
        break;
    }
  };

  const togglePanel = (panelId: string): void => {
    const panel = panels[panelId];
    if (panel) {
      panel.isCollapsed = !panel.isCollapsed;
      savePanelState(panelId, { isCollapsed: panel.isCollapsed });
    }
  };

  const resizePanel = (panelId: string, width: number): void => {
    const panel = panels[panelId];
    if (panel) {
      const clampedWidth = Math.max(
        panel.minWidth,
        Math.min(panel.maxWidth, width)
      );
      
      panel.width = clampedWidth;
      
      // Update layout config based on panel position
      if (panel.position === 'left') {
        layoutConfig.leftPanelWidth = clampedWidth;
      } else if (panel.position === 'right') {
        layoutConfig.rightPanelWidth = clampedWidth;
      }
      
      savePanelState(panelId, { width: clampedWidth });
    }
  };

  const resetLayout = (): void => {
    // Reset all panels to default state
    Object.keys(panels).forEach(panelId => {
      const panel = panels[panelId];
      panel.isCollapsed = false;
      
      switch (panelId) {
        case 'files':
          panel.width = 320;
          break;
        case 'context':
          panel.width = 400;
          break;
        case 'reports':
          panel.width = 380;
          break;
        case 'autonomous':
          panel.width = 350;
          break;
      }
    });

    // Reset layout config
    layoutConfig.leftPanelWidth = 320;
    layoutConfig.rightPanelWidth = 380;
    layoutConfig.topSectionHeight = 60;

    // Clear saved state
    Object.keys(panels).forEach(panelId => {
      try {
        localStorage.removeItem(`panel_${panelId}`);
      } catch (error) {
        console.warn(`Failed to clear panel state for ${panelId}:`, error);
      }
    });
  };

  const savePanelState = (panelId: string, state: Partial<PanelState>): void => {
    try {
      const panel = panels[panelId];
      if (panel) {
        Object.assign(panel, state);
        localStorage.setItem(`panel_${panelId}`, JSON.stringify(panel));
      }
    } catch (error) {
      console.warn(`Failed to save panel state for ${panelId}:`, error);
    }
  };

  const loadPanelState = (panelId: string): PanelState | null => {
    try {
      const stored = localStorage.getItem(`panel_${panelId}`);
      if (stored) {
        const state = JSON.parse(stored) as PanelState;
        
        // Merge with current panel state
        const panel = panels[panelId];
        if (panel) {
          Object.assign(panel, state);
        }
        
        return state;
      }
    } catch (error) {
      console.warn(`Failed to load panel state for ${panelId}:`, error);
    }
    
    return null;
  };

  const isValidForMode = (mode: WorkspaceMode): boolean => {
    const workspace = projectStore.workspace;
    if (!workspace) return false;
    
    return workspace.isValidForMode(mode);
  };

  // Initialize panel states from localStorage
  const initializePanelStates = (): void => {
    Object.keys(panels).forEach(panelId => {
      loadPanelState(panelId);
    });
  };

  // Watch for workspace state changes
  watch(
    () => projectStore.workspace,
    (newWorkspace) => {
      if (newWorkspace) {
        currentMode.value = newWorkspace.mode;
      }
    },
    { immediate: true }
  );

  // Watch for panel changes and update project store
  watch(
    panels,
    () => {
      Object.keys(panels).forEach(panelId => {
        projectStore.updatePanelConfiguration(panelId, panels[panelId]);
      });
    },
    { deep: true }
  );

  // Initialize
  initializePanelStates();

  return {
    // State
    get currentMode() { return currentMode.value; },
    get isProjectLoaded() { return isProjectLoaded.value; },
    panels,
    layoutConfig,
    
    // Actions
    switchMode,
    togglePanel,
    resizePanel,
    resetLayout,
    savePanelState,
    loadPanelState,
    isValidForMode
  };
}