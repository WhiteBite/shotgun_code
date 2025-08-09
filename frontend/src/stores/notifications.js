
import { defineStore } from 'pinia';
import { ref } from 'vue';
import { EventsOn } from '../../wailsjs/runtime/runtime';

export const useNotificationsStore = defineStore('notifications', () => {
  /**
   * @type {import('vue').Ref<Array<{message: string, type: 'info'|'success'|'warn'|'error', timestamp: string}>>}
   */
  const logs = ref([]);

  function addLog(message, type = 'info') {
    const logEntry = {
      message,
      type,
      timestamp: new Date().toLocaleTimeString(),
    };
    logs.value.unshift(logEntry); // Добавляем в начало для отображения сверху
  }

  function setupWailsListeners() {
    EventsOn("app:error", (errorMessage) => {
      addLog(errorMessage, 'error');
    });
    // Можно добавить другие глобальные события
  }

  // Инициализируем слушатели один раз при создании стора
  setupWailsListeners();

  return {
    logs,
    addLog,
  };
});