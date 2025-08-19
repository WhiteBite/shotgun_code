import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { loadAndHighlight, type QuickLookType } from "@/services/quicklook.service";

export interface QuickLookOptions {
  rootDir: string; path: string; type: QuickLookType; commitHash?: string; event: MouseEvent | null;
  isPinned?: boolean; position?: { x: number; y: number }; content?: string;
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
  const toasts = ref<Array<{
    id: string;
    message: string;
    type: "success" | "error" | "info" | "warning";
    duration?: number;
  }>>([]);

  // Console visibility
  const isConsoleVisible = ref(false);

  // Computed
  const isAnyDrawerOpen = computed(() => 
    drawers.value.ignore || drawers.value.prompts || drawers.value.settings
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
        options.commitHash
      );

      quickLook.value.content = result.content;
      quickLook.value.language = result.language;
      quickLook.value.truncated = result.truncated;
    } catch (error) {
      quickLook.value.error = error instanceof Error ? error.message : String(error);
    }
  }

  function hideQuickLook() {
    quickLook.value.isActive = false;
    quickLook.value.error = null;
  }

  function togglePin() {
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
    duration: number = 5000
  ) {
    const id = Date.now().toString();
    toasts.value.push({ id, message, type, duration });
    
    if (duration > 0) {
      setTimeout(() => {
        removeToast(id);
      }, duration);
    }
  }

  function removeToast(id: string) {
    const index = toasts.value.findIndex(t => t.id === id);
    if (index > -1) {
      toasts.value.splice(index, 1);
    }
  }

  function toggleConsole() {
    isConsoleVisible.value = !isConsoleVisible.value;
  }

  return {
    // State
    contextMenu: computed(() => contextMenu.value),
    quickLook: computed(() => quickLook.value),
    drawers: computed(() => drawers.value),
    toasts: computed(() => toasts.value),
    isConsoleVisible: computed(() => isConsoleVisible.value),
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
  };
});