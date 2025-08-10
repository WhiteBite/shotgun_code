import { defineStore } from 'pinia';
import { ref } from 'vue';
import { SuggestContextFiles } from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notificationsStore';
import { useFileTreeStore } from './fileTreeStore';
import { usePromptStore } from './promptStore';

/**
 * Manages the logic for automatically suggesting context files.
 */
export const useContextAnalysisStore = defineStore('contextAnalysis', () => {
  const notifications = useNotificationsStore();
  const fileTree = useFileTreeStore();
  const prompt = usePromptStore();

  const isSuggesting = ref(false);

  async function suggestAndApplyContext() {
    if (!prompt.userTask.trim()) {
      notifications.addLog('Невозможно предложить файлы без описания задачи.', 'warn');
      return;
    }
    if (!fileTree.tree || fileTree.tree.length === 0) {
      notifications.addLog('Дерево файлов не загружено.', 'error');
      return;
    }

    isSuggesting.value = true;
    try {
      notifications.addLog('Анализ задачи для авто-выбора файлов...', 'info');
      const suggestedPaths = await SuggestContextFiles(prompt.userTask, fileTree.tree);

      if (suggestedPaths && suggestedPaths.length > 0) {
        fileTree.applySelectionSet(new Set(suggestedPaths));
        notifications.addLog(`Автоматически выбрано ${suggestedPaths.length} файлов.`, 'success');
        // УБРАЛИ АВТОМАТИЧЕСКИЙ ЗАПУСК ГЕНЕРАЦИИ КОНТЕКСТА
      } else {
        notifications.addLog('Не удалось найти релевантные файлы по задаче. Попробуйте выбрать вручную.', 'info');
      }
    } catch (err) {
      notifications.addLog(`Ошибка при авто-выборе файлов: ${err.message || err}`, 'error');
    } finally {
      isSuggesting.value = false;
    }
  }

  return {
    isSuggesting,
    suggestAndApplyContext,
  };
});