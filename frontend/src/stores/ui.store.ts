import { defineStore } from 'pinia';
import { ref } from 'vue';

type DrawerName = 'settings' | 'ignore' | 'prompts';
export type ToastType = 'info' | 'success' | 'error';
export interface Toast { id: number; message: string; type: ToastType; }
export interface ProgressState { isActive: boolean; message: string; value: number; }

export const useUiStore = defineStore('ui', () => {
  const activeDrawer = ref<DrawerName | null>(null);
  const isConsoleVisible = ref(false);

  const toasts = ref<Toast[]>([]);
  let toastIdCounter = 0;

  const progress = ref<ProgressState>({
    isActive: false,
    message: '',
    value: 0,
  });

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

  return {
    activeDrawer,
    isConsoleVisible,
    toasts,
    progress,
    openDrawer,
    closeDrawer,
    toggleConsole,
    addToast,
    removeToast,
    setProgress,
    clearProgress,
  };
});