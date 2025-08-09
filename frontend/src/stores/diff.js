
import { defineStore } from 'pinia';
import { ref } from 'vue';
import { SplitShotgunDiff } from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notifications';
import { useStepsStore } from './steps';

export const useDiffStore = defineStore('diff', () => {
  // ИСПРАВЛЕНО: use...Store() вызовы перенесены внутрь
  const notifications = useNotificationsStore();
  const steps = useStepsStore();

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
      steps.completeStep(3);
      steps.navigateToStep(4);
    } catch (err) {
      notifications.addLog(`Ошибка разделения diff: ${err.message || err}`, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  return {
    gitDiffInput,
    splitLineLimit,
    splitDiffs,
    isLoading,
    processAndSplitDiff,
  };
});