import { defineStore } from 'pinia';
import { ref } from 'vue';
import hljs from 'highlight.js';
import { apiService } from '@/services/api.service';
import { useProjectStore } from './project.store';
import { useErrorHandler } from '@/composables/useErrorHandler';

type DrawerName = 'settings' | 'ignore' | 'prompts';
export type ToastType = 'info' | 'success' | 'error';
export interface Toast { id: number; message: string; type: ToastType; }
export interface ProgressState { isActive: boolean; message: string; value: number; }
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
}

export const useUiStore = defineStore('ui', () => {
  const activeDrawer = ref<DrawerName | null>(null);
  const isConsoleVisible = ref(false);
  const toasts = ref<Toast[]>([]);
  let toastIdCounter = 0;
  const progress = ref<ProgressState>({ isActive: false, message: '', value: 0 });
  const quickLook = ref<QuickLookState>({
    visible: false,
    content: '',
    rawContent: '',
    language: 'plaintext',
    isLoading: false,
    error: null,
    x: 0,
    y: 0,
    isPinned: false,
    filePath: '',
  });

  const { handleError } = useErrorHandler();
  let hideTimeout: number | null = null;

  function openDrawer(drawerName: DrawerName) {
    activeDrawer.value = activeDrawer.value === drawerName ? null : drawerName;
  }
  function closeDrawer() { activeDrawer.value = null; }
  function toggleConsole() { isConsoleVisible.value = !isConsoleVisible.value; }
  function addToast(message: string, type: ToastType = 'info') {
    const id = toastIdCounter++;
    toasts.value.push({ id, message, type });
  }
  function removeToast(id: number) { toasts.value = toasts.value.filter(t => t.id !== id); }
  function setProgress(state: Partial<ProgressState> & { isActive: true }) {
    progress.value = { ...progress.value, ...state };
  }
  function clearProgress() {
    progress.value = { ...progress.value, isActive: false };
  }

  async function showQuickLook(payload: { path: string, type: 'fs' | 'git', commitHash?: string, event: MouseEvent }) {
    if (hideTimeout) clearTimeout(hideTimeout);
    if (quickLook.value.visible && quickLook.value.filePath === payload.path) return;

    quickLook.value.visible = true;
    quickLook.value.isLoading = true;
    quickLook.value.error = null;
    quickLook.value.filePath = payload.path;

    if (!quickLook.value.isPinned) {
      quickLook.value.x = payload.event.clientX + 15;
      quickLook.value.y = payload.event.clientY + 15;
    }

    const projectStore = useProjectStore();
    if (!projectStore.currentProject) return;

    try {
      let rawContent = '';
      if (payload.type === 'fs') {
        rawContent = await apiService.readFileContent(projectStore.currentProject.path, payload.path);
      } else if (payload.type === 'git' && payload.commitHash) {
        rawContent = await apiService.getFileContentAtCommit(projectStore.currentProject.path, payload.path, payload.commitHash);
      }

      quickLook.value.rawContent = rawContent;
      const lang = payload.path.split('.').pop() || 'plaintext';
      quickLook.value.language = lang;

      if (hljs.getLanguage(lang)) {
        quickLook.value.content = hljs.highlight(rawContent, { language: lang }).value;
      } else {
        quickLook.value.content = hljs.highlightAuto(rawContent).value;
      }
    } catch (err) {
      handleError(err, 'Quick Look');
      quickLook.value.error = `Could not load file: ${payload.path}`;
    } finally {
      quickLook.value.isLoading = false;
    }
  }

  function hideQuickLook() {
    if (!quickLook.value.isPinned) {
      quickLook.value.visible = false;
      quickLook.value.isPinned = false; // Also reset pin on hide
    }
  }

  function requestHideQuickLook() {
    if (hideTimeout) clearTimeout(hideTimeout);
    hideTimeout = window.setTimeout(() => {
      hideQuickLook();
    }, 100);
  }

  function keepQuickLookOpen() {
    if (hideTimeout) clearTimeout(hideTimeout);
  }

  function toggleQuickLookPin() {
    quickLook.value.isPinned = !quickLook.value.isPinned;
  }

  return {
    activeDrawer, isConsoleVisible, toasts, progress, quickLook,
    openDrawer, closeDrawer, toggleConsole, addToast, removeToast,
    setProgress, clearProgress, showQuickLook, hideQuickLook,
    requestHideQuickLook, keepQuickLookOpen, toggleQuickLookPin,
  };
});