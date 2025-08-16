import { defineStore } from "pinia";
import { ref, reactive } from "vue";
import type { ToastType } from "@/types/dto";
import { loadAndHighlight, type QuickLookType } from "@/services/quicklook.service";

export interface Toast { id: number; message: string; type: ToastType | "warn"; }
interface ContextMenu { isVisible: boolean; x: number; y: number; nodePath: string; }
interface ProgressState { isActive: boolean; message: string; value: number; }
interface QuickLookState {
  isActive: boolean; isPinned: boolean; path: string; rootDir: string;
  type: QuickLookType; commitHash?: string; event: MouseEvent | null;
  content: string; error: string | null; language: string; truncated: boolean;
  position: { x: number; y: number } | null;
}
export type DrawerType = "ignore" | "prompt" | "settings";

export const useUiStore = defineStore("ui", () => {
  const toasts = ref<Toast[]>([]);
  let toastId = 0;
  const isConsoleVisible = ref(false);
  const activeDrawer = ref<DrawerType | null>(null);
  const contextMenu = ref<ContextMenu | null>(null);
  const progress = reactive<ProgressState>({ isActive: false, message: "", value: 0 });
  const quickLook = reactive<QuickLookState>({
    isActive: false, isPinned: false, path: "", rootDir: "", type: "fs", event: null,
    content: "", error: null, language: "plaintext", truncated: false, position: null
  });

  function addToast(message: string, type: ToastType | "warn" = "info", duration = 4000) {
    const id = toastId++;
    toasts.value.unshift({ id, message, type });
    window.setTimeout(() => removeToast(id), duration);
  }
  function removeToast(id: number) { toasts.value = toasts.value.filter((t) => t.id !== id); }
  function toggleConsole() { isConsoleVisible.value = !isConsoleVisible.value; }
  function openDrawer(drawer: DrawerType) { activeDrawer.value = drawer; }
  function closeDrawer() { activeDrawer.value = null; }
  function openContextMenu(x: number, y: number, nodePath: string) { contextMenu.value = { isVisible: true, x, y, nodePath }; }
  function closeContextMenu() { if (contextMenu.value) contextMenu.value.isVisible = false; }
  function setProgress(state: Partial<ProgressState>) { Object.assign(progress, state); }
  function clearProgress() { progress.isActive = false; progress.message = ""; progress.value = 0; }

  async function showQuickLook(payload: {
    rootDir: string; path: string; type: QuickLookType; commitHash?: string;
    event?: MouseEvent; isPinned?: boolean;
  }) {
    const { rootDir, path, type, commitHash, event, isPinned } = payload;
    if (quickLook.isPinned && !isPinned) return;

    quickLook.isActive = true;
    quickLook.isPinned = !!isPinned;
    quickLook.path = path;
    quickLook.rootDir = rootDir;
    quickLook.type = type;
    quickLook.event = event || null;
    quickLook.error = null;
    quickLook.truncated = false;

    if (isPinned && !quickLook.position) {
      quickLook.position = { x: window.innerWidth / 2 - 300, y: window.innerHeight / 2 - 200 };
    }

    try {
      const res = await loadAndHighlight({ rootDir, path, type, commitHash });
      quickLook.content = res.html;
      quickLook.language = res.language;
      quickLook.truncated = res.truncated;
    } catch (err: any) {
      quickLook.error = err?.message || String(err);
      quickLook.content = "";
    }
  }

  function hideQuickLook() {
    if (!quickLook.isPinned) {
      quickLook.isActive = false;
      quickLook.position = null;
    }
  }

  function togglePin() {
    quickLook.isPinned = !quickLook.isPinned;
    if (quickLook.isPinned && !quickLook.position) {
      quickLook.position = { x: window.innerWidth / 2 - 300, y: window.innerHeight / 2 - 200 };
    }
    if (!quickLook.isPinned) {
      quickLook.position = null;
      quickLook.isActive = false;
    }
  }

  function setPosition(pos: { x: number; y: number }) {
    if (quickLook.isPinned) {
      quickLook.position = pos;
    }
  }


  return {
    toasts, addToast, removeToast, activeDrawer, openDrawer, closeDrawer, contextMenu, openContextMenu, closeContextMenu,
    progress, setProgress, clearProgress, quickLook, showQuickLook, hideQuickLook, togglePin, setPosition, isConsoleVisible, toggleConsole,
  };
});