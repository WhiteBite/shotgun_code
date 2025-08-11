import { defineStore } from 'pinia';
import { ref } from 'vue';

type DrawerName = 'settings' | 'ignore' | 'prompts';
export type ToastType = 'info' | 'success' | 'error';
export interface Toast { id: number; message: string; type: ToastType; }

export const useUiStore = defineStore('ui', () => {
  const activeDrawer = ref<DrawerName | null>(null);
  const isCommitModalVisible = ref(false);
  const isConsoleVisible = ref(false);
  const toasts = ref<Toast[]>([]);
  let toastIdCounter = 0;

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
  function openCommitModal() { isCommitModalVisible.value = true; }
  function closeCommitModal() { isCommitModalVisible.value = false; }

  return {
    activeDrawer, isCommitModalVisible, isConsoleVisible, toasts,
    openDrawer, closeDrawer, toggleConsole, addToast, removeToast,
    openCommitModal, closeCommitModal,
  };
});