import { defineStore } from "pinia";
import { ref, computed } from "vue";
import {
  loadAndHighlight,
  type QuickLookType,
} from "@/infrastructure/quicklook/quicklook.service";

export interface QuickLookOptions {
  rootDir: string;
  path: string;
  type: QuickLookType;
  commitHash?: string;
  isPinned?: boolean;
  position?: { x: number; y: number };
  content?: string;
}

export const useUiStore = defineStore("ui", () => {
  // Context menu state
  const contextMenu = ref({
    isOpen: false,
    x: 0,
    y: 0,
    targetPath: "",
  });

  // QuickLook state
  const quickLook = ref({
    isActive: false,
    path: "",
    content: "",
    language: "text",
    truncated: false,
    error: null as string | null,
    isPinned: false,
    position: null as { x: number; y: number } | null,
  });

  // Drawer states
  const drawers = ref({
    ignore: false,
    prompts: false,
    settings: false,
  });

  // Toast notifications
  const toasts = ref<
    Array<{
      id: string;
      message: string;
      type: "success" | "error" | "info" | "warning";
      duration?: number;
    }>
  >([]);

  // Console visibility
  const isConsoleVisible = ref(false);

  // Console filters/state
  const consoleFilters = ref({
    levels: {
      debug: true,
      info: true,
      success: true,
      warning: true,
      error: true,
    },
    autoscroll: true,
    paused: false,
  });

  // Panel resize states
  const panelResizing = ref({
    isResizing: false,
    activePanel: null as 'context' | 'results' | null,
    startPosition: 0,
    startWidth: 0
  });
  
  // Layout animation states
  const animationStates = ref({
    transitioning: false,
    panelCollapsing: false,
    modeChanging: false
  });
  
  // Keyboard navigation state
  const keyboardNavigation = ref({
    focusedElement: null as string | null,
    tabIndex: 0,
    modalStack: [] as string[]
  });

  // Computed
  const isAnyDrawerOpen = computed(
    () =>
      drawers.value.ignore || drawers.value.prompts || drawers.value.settings,
  );

  // Methods
  function openContextMenu(x: number, y: number, targetPath: string) {
    contextMenu.value = { isOpen: true, x, y, targetPath };
  }

  function closeContextMenu() {
    contextMenu.value.isOpen = false;
  }

  async function showQuickLook(options: QuickLookOptions) {
    try {
      quickLook.value.isActive = true;
      quickLook.value.path = options.path;
      quickLook.value.isPinned = options.isPinned || false;
      quickLook.value.position = options.position || null;
      quickLook.value.error = null;

      // Если передан контент для типа "text", используем его
      if (options.type === "text" && options.content) {
        quickLook.value.content = options.content;
        quickLook.value.language = "text";
        quickLook.value.truncated = false;
        return;
      }

      const result = await loadAndHighlight(
        options.rootDir,
        options.path,
        options.type,
        options.commitHash,
      );

      quickLook.value.content = result.content;
      quickLook.value.language = result.language;
      quickLook.value.truncated = result.truncated;
    } catch (error) {
      quickLook.value.error =
        error instanceof Error ? error.message : String(error);
    }
  }

  function hideQuickLook() {
    quickLook.value.isActive = false;
    quickLook.value.error = null;
  }

  function togglePin() {
    // Просто переключаем закрепление; позицию не сбрасываем — она сохранится через setPosition
    quickLook.value.isPinned = !quickLook.value.isPinned;
  }

  function setPosition(position: { x: number; y: number }) {
    quickLook.value.position = position;
  }

  function openDrawer(name: keyof typeof drawers.value) {
    drawers.value[name] = true;
  }

  function closeDrawer(name: keyof typeof drawers.value) {
    drawers.value[name] = false;
  }

  function addToast(
    message: string,
    type: "success" | "error" | "info" | "warning" = "info",
    duration: number = 5000,
  ) {
    // Prevent toast spam by limiting the number of active toasts
    if (toasts.value.length > 5) {
      // Remove oldest non-error toast if we have too many
      const oldestNonErrorIndex = toasts.value.findIndex(t => t.type !== 'error');
      if (oldestNonErrorIndex >= 0) {
        toasts.value.splice(oldestNonErrorIndex, 1);
      } else if (type !== 'error') {
        // If all are errors and this is not an error, skip this toast
        return;
      }
    }
    
    // Check for duplicate messages (avoid repeating the same toast)
    const hasDuplicate = toasts.value.some(t => t.message === message && t.type === type);
    if (hasDuplicate) {
      return;
    }

    const id = Date.now().toString();
    toasts.value.push({ id, message, type, duration });

    if (duration > 0) {
      setTimeout(() => {
        removeToast(id);
      }, duration);
    }
  }

  function removeToast(id: string) {
    const index = toasts.value.findIndex((t) => t.id === id);
    if (index > -1) {
      toasts.value.splice(index, 1);
    }
  }

  function toggleConsole() {
    isConsoleVisible.value = !isConsoleVisible.value;
  }

  function setConsoleLevel(level: 'debug' | 'info' | 'success' | 'warning' | 'error', enabled: boolean) {
    consoleFilters.value.levels[level] = enabled;
  }

  function toggleAutoscroll() {
    consoleFilters.value.autoscroll = !consoleFilters.value.autoscroll;
  }

  function togglePauseConsole() {
    consoleFilters.value.paused = !consoleFilters.value.paused;
  }

  // Panel resize methods
  function startPanelResize(panel: 'context' | 'results', startX: number, startWidth: number) {
    panelResizing.value = {
      isResizing: true,
      activePanel: panel,
      startPosition: startX,
      startWidth
    };
  }
  
  function updatePanelResize(currentX: number) {
    if (!panelResizing.value.isResizing || !panelResizing.value.activePanel) return;
    
    const delta = currentX - panelResizing.value.startPosition;
    const newWidth = panelResizing.value.startWidth + delta;
    
    return {
      panel: panelResizing.value.activePanel,
      width: Math.max(200, Math.min(800, newWidth))
    };
  }
  
  function endPanelResize() {
    panelResizing.value = {
      isResizing: false,
      activePanel: null,
      startPosition: 0,
      startWidth: 0
    };
  }
  
  // Animation methods
  function setTransitioning(transitioning: boolean) {
    animationStates.value.transitioning = transitioning;
  }
  
  function setPanelCollapsing(collapsing: boolean) {
    animationStates.value.panelCollapsing = collapsing;
  }
  
  function setModeChanging(changing: boolean) {
    animationStates.value.modeChanging = changing;
  }
  
  // Keyboard navigation methods
  function setKeyboardFocus(elementId: string | null) {
    keyboardNavigation.value.focusedElement = elementId;
  }
  
  function pushModal(modalId: string) {
    keyboardNavigation.value.modalStack.push(modalId);
  }
  
  function popModal() {
    keyboardNavigation.value.modalStack.pop();
  }
  
  function getCurrentModal(): string | null {
    const stack = keyboardNavigation.value.modalStack;
    return stack.length > 0 ? stack[stack.length - 1] : null;
  }
  
  // Enhanced toast with actions
  function addActionToast(
    message: string,
    action: { label: string; callback: () => void },
    type: "success" | "error" | "info" | "warning" = "info",
    duration: number = 10000
  ) {
    const id = Date.now().toString();
    const toast = { id, message, type, duration, action };
    toasts.value.push(toast);
    
    if (duration > 0) {
      setTimeout(() => {
        removeToast(id);
      }, duration);
    }
    
    return id;
  }

  return {
    // State
    contextMenu: computed(() => contextMenu.value),
    quickLook: computed(() => quickLook.value),
    drawers: computed(() => drawers.value),
    toasts: computed(() => toasts.value),
    isConsoleVisible: computed(() => isConsoleVisible.value),
    consoleFilters: computed(() => consoleFilters.value),
    panelResizing: computed(() => panelResizing.value),
    animationStates: computed(() => animationStates.value),
    keyboardNavigation: computed(() => keyboardNavigation.value),
    isAnyDrawerOpen,

    // Methods
    openContextMenu,
    closeContextMenu,
    showQuickLook,
    hideQuickLook,
    togglePin,
    setPosition,
    openDrawer,
    closeDrawer,
    addToast,
    removeToast,
    toggleConsole,
    setConsoleLevel,
    toggleAutoscroll,
    togglePauseConsole,
    startPanelResize,
    updatePanelResize,
    endPanelResize,
    setTransitioning,
    setPanelCollapsing,
    setModeChanging,
    setKeyboardFocus,
    pushModal,
    popModal,
    getCurrentModal,
    addActionToast,
  };
});