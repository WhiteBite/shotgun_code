import { defineStore } from 'pinia';
import { ref } from 'vue';
import { EventsOn } from '../../wailsjs/runtime/runtime';

export const useNotificationsStore = defineStore('notifications', () => {
  const logs = ref([]);

  function addLog(message, type = 'info') {
    const logEntry = {
      message,
      type,
      timestamp: new Date().toLocaleTimeString(),
    };
    logs.value.unshift(logEntry);
  }

  function setupWailsListeners() {
    EventsOn("app:error", (errorMessage) => {
      addLog(errorMessage, 'error');
    });
  }

  setupWailsListeners();

  return {
    logs,
    addLog,
  };
});