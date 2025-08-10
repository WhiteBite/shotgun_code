import { defineStore } from 'pinia';
import { ref } from 'vue';
import { SplitShotgunDiff } from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notificationsStore';

export const useDiffStore = defineStore('diff', () => {
  const notifications = useNotificationsStore();

  const gitDiffInput = ref('');
  const splitLineLimit = ref(500);
  const splitDiffs = ref([]);
  const isLoading = ref(false);

  async function processAndSplitDiff() {
    if (!gitDiffInput.value.trim()) {
      notifications.addLog('Нет diff для обработки.', 'warn');
      return;
    }
    isLoading.value = true;
    splitDiffs.value = [];
    try {
      const result = await SplitShotgunDiff(gitDiffInput.value, splitLineLimit.value);
      splitDiffs.value = result;
      notifications.addLog(`Diff успешно разделен на ${result.length} частей.`, 'success');
    } catch (err) {
      notifications.addLog(`Ошибка разделения diff: ${err.message || err}`, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  async function setDiffInputAndProcess(diffText) {
    gitDiffInput.value = diffText;
    await processAndSplitDiff();
  }

  return {
    gitDiffInput,
    splitLineLimit,
    splitDiffs,
    isLoading,
    processAndSplitDiff,
    setDiffInputAndProcess,
  };
});