import { defineStore } from 'pinia';
import { ref } from 'vue';
import { GenerateCode } from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notificationsStore';
import { usePromptStore } from './promptStore';
import { useContextStore } from './contextStore';
import { useDiffStore } from './diffStore';
import { useStepsStore } from './stepsStore';

/**
 * Manages the lifecycle of AI code generation requests.
 */
export const useAiExecutionStore = defineStore('aiExecution', () => {
  const notifications = useNotificationsStore();
  const prompt = usePromptStore();
  const context = useContextStore();
  const diff = useDiffStore();
  const steps = useStepsStore();

  const isLoading = ref(false);
  const error = ref('');

  async function executeAIGeneration() {
    if (!context.shotgunPromptContext) {
      notifications.addLog('Невозможно сгенерировать решение без контекста проекта.', 'warn');
      return;
    }

    // Ensure the final prompt is up-to-date before sending
    prompt.generateFinalPrompt();
    const { systemPrompt, userPrompt } = prompt.getGeneratedPrompts();

    if (!userPrompt || !systemPrompt) {
      notifications.addLog('Промпт не был сгенерирован, выполнение отменено.', 'error');
      return;
    }

    isLoading.value = true;
    error.value = '';
    try {
      notifications.addLog('Отправка запроса к AI...', 'info');
      const result = await GenerateCode(systemPrompt, userPrompt);
      notifications.addLog('Ответ от AI успешно получен.', 'success');

      diff.setDiffInputAndProcess(result);
      steps.navigateToStep(2);

    } catch (err) {
      const errorMessage = `Ошибка при генерации AI: ${err.message || err}`;
      error.value = errorMessage;
      notifications.addLog(errorMessage, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  return {
    isLoading,
    error,
    executeAIGeneration,
  };
});