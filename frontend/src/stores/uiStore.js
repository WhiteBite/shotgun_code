import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useUiStore = defineStore('ui', () => {
  const isCommitModalVisible = ref(false);
  const isSettingsModalVisible = ref(false);

  function openCommitModal() {
    isCommitModalVisible.value = true;
  }

  function closeCommitModal() {
    isCommitModalVisible.value = false;
  }

  function openSettingsModal() {
    isSettingsModalVisible.value = true;
  }

  function closeSettingsModal() {
    isSettingsModalVisible.value = false;
  }

  return {
    isCommitModalVisible,
    openCommitModal,
    closeCommitModal,
    isSettingsModalVisible,
    openSettingsModal,
    closeSettingsModal,
  };
});