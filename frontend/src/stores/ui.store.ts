import { defineStore } from "pinia";
import { ref } from "vue";
import hljs from "highlight.js";
import { useErrorHandler } from "@/composables/useErrorHandler";
import { useNotificationsStore } from "./notifications.store";

type DrawerName = "settings" | "ignore" | "prompts";
export type ToastType = "info" | "success" | "error";
export interface Toast {
  id: number;
  message: string;
  type: ToastType;
}
export interface ProgressState {
  isActive: boolean;
  message: string;
  value: number;
}

export interface QuickLookState {
  visible: boolean;
  content: string;
  rawContent: string;
  language: string;
  isLoading: boolean;
  error: string | null;
  x: number;
  y: number;
  isPinned: boolean;
  filePath: string;
  truncated: boolean;
}

export interface ShowQuickLookPayload {
  rootDir: string;
  path: string;
  type: "fs" | "git";
  commitHash?: string;
  event: MouseEvent;
  isPinned: boolean;
}

export interface ContextMenuState {
  visible: boolean;
  x: number;
  y: number;
  targetPath: string | null;
}

export const useUiStore = defineStore("ui", () => {
  const activeDrawer = ref<DrawerName | null>(null);
  const isConsoleVisible = ref(false);
  const toasts = ref<Toast[]>([]);
  let toastIdCounter = 0;

  const progress = ref<ProgressState>({
    isActive: false,
    message: "",
    value: 0,
  });

  const quickLook = ref<QuickLookState>({
    visible: false,
    content: "",
    rawContent: "",
    language: "plaintext",
    isLoading: false,
    error: null,
    x: 0,
    y: 0,
    isPinned: false,
    filePath: "",
    truncated: false,
  });

  const contextMenu = ref<ContextMenuState>({
    visible: false,
    x: 0,
    y: 0,
    targetPath: null,
  });

  const notifications = useNotificationsStore();
  const { handleError } = useErrorHandler();

  function openDrawer(d: DrawerName) {
    activeDrawer.value = activeDrawer.value === d ? null : d;
  }
  function closeDrawer() {
    activeDrawer.value = null;
  }
  function toggleConsole() {
    isConsoleVisible.value = !isConsoleVisible.value;
  }

  function addToast(message: string, type: ToastType = "info") {
    const id = toastIdCounter++;
    toasts.value.push({ id, message, type });
  }
  function removeToast(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }

  function setProgress(state: Partial<ProgressState> & { isActive: true }) {
    progress.value = { ...progress.value, ...state };
  }
  function clearProgress() {
    progress.value = { ...progress.value, isActive: false };
  }

  async function showQuickLook(payload: ShowQuickLookPayload) {
    // hover не перебивает pinned
    if (quickLook.value.isPinned && !payload.isPinned) return;

    // лог для диагностики
    notifications.addLog(
      `QuickLook: open ${payload.path} (pinned=${payload.isPinned})`,
      "info",
    );

    quickLook.value.visible = true;
    quickLook.value.isLoading = true;
    quickLook.value.error = null;
    quickLook.value.filePath = payload.path;
    quickLook.value.isPinned = payload.isPinned;
    quickLook.value.truncated = false;

    if (!payload.isPinned) {
      quickLook.value.x = payload.event.clientX + 20;
      quickLook.value.y = payload.event.clientY + 20;
    } else {
      // центр экрана
      quickLook.value.x = window.innerWidth / 2 - 300;
      quickLook.value.y = window.innerHeight * 0.15;
    }

    if (!payload.rootDir) {
      quickLook.value.isLoading = false;
      quickLook.value.error = "Project root is not set.";
      notifications.addLog("QuickLook error: root is empty", "error");
      return;
    }

    try {
      let rawContent = "";
      if (payload.type === "fs") {
        // вызов сделан снаружи (apiService) — оставим как есть
        const { apiService } = await import("@/services/api.service");
        rawContent = await apiService.readFileContent(
          payload.rootDir,
          payload.path,
        );
      } else if (payload.type === "git" && payload.commitHash) {
        const { apiService } = await import("@/services/api.service");
        rawContent = await apiService.getFileContentAtCommit(
          payload.rootDir,
          payload.path,
          payload.commitHash,
        );
      }

      quickLook.value.rawContent = rawContent;
      const lang = payload.path.split(".").pop() || "plaintext";
      quickLook.value.language = lang;

      if (hljs.getLanguage(lang)) {
        quickLook.value.content = hljs.highlight(rawContent, {
          language: lang,
        }).value;
      } else {
        quickLook.value.content = hljs.highlightAuto(rawContent).value;
      }
      notifications.addLog(`QuickLook: loaded ${payload.path}`, "info");
    } catch (err) {
      handleError(err, "Quick Look");
      quickLook.value.error = `Could not load file: ${payload.path}`;
    } finally {
      quickLook.value.isLoading = false;
    }
  }

  function hideQuickLook() {
    quickLook.value.visible = false;
    quickLook.value.isPinned = false;
  }

  function togglePinQuickLook() {
    quickLook.value.isPinned = !quickLook.value.isPinned;
  }

  function openContextMenu(x: number, y: number, path: string) {
    contextMenu.value = { visible: true, x, y, targetPath: path };
  }
  function closeContextMenu() {
    contextMenu.value.visible = false;
    contextMenu.value.targetPath = null;
  }

  return {
    activeDrawer,
    isConsoleVisible,
    toasts,
    progress,
    quickLook,
    contextMenu,
    openDrawer,
    closeDrawer,
    toggleConsole,
    addToast,
    removeToast,
    setProgress,
    clearProgress,
    showQuickLook,
    hideQuickLook,
    togglePinQuickLook,
    openContextMenu,
    closeContextMenu,
  };
});
