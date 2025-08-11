import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { GenerateCode } from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notificationsStore';
import { useTaskStore } from './taskStore';
import { useContextStore } from './contextStore';
import { useReviewStore } from './reviewStore';

export const useGenerationStore = defineStore('generation', () => {
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  const notifications = useNotificationsStore();
  const taskStore = useTaskStore();
  const contextStore = useContextStore();
  const reviewStore = useReviewStore();
  const router = useRouter();

  async function executeGeneration() {
    if (!contextStore.shotgunContextText) {
      notifications.addLog('Контекст не собран. Генерация невозможна.', 'warn');
      return;
    }
    isLoading.value = true;
    error.value = null;
    notifications.addLog('Отправка запроса к AI...', 'info');

    try {
      // System prompt can be enhanced using settingsStore in the future
      const systemPrompt = "You are an expert software developer that provides changes in git diff format.";
      const userPrompt = `TASK: ${taskStore.userTask}\n\nCONTEXT:\n${contextStore.shotgunContextText}`;

      const result = await GenerateCode(systemPrompt, userPrompt);

      notifications.addLog('Ответ от AI успешно получен.', 'success');
      reviewStore.setGeneratedDiff(result);
      router.push('/review');
    } catch (err: any) {
      error.value = `Ошибка генерации: ${err.message || err}`;
      notifications.addLog(error.value, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  return {
    isLoading,
    error,
    executeGeneration,
  };
});