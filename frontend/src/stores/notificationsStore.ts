import { defineStore } from 'pinia';
import { ref } from 'vue';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import type { LogEntry } from '@/types/dto';
import { useUiStore } from './uiStore';

export const useNotificationsStore = defineStore('notifications', () => {
  const logs = ref<LogEntry[]>([]);
  const maxLogs = 100;

  function addLog(message: string, type: LogEntry['type'] = 'info') {
    const uiStore = useUiStore();
    const logEntry: LogEntry = {
      message,
      type,
      timestamp: new Date().toLocaleTimeString(),
    };
    logs.value.unshift(logEntry);
    if (logs.value.length > maxLogs) {
      logs.value.pop();
    }
    // Also show a toast for errors
    if (type === 'error') {
      uiStore.addToast(message, 'error');
    }
  }

  function setupWailsListeners() {
    EventsOn("app:error", (errorMessage: string) => {
      addLog(errorMessage, 'error');
    });
  }

  setupWailsListeners();

  return {
    logs,
    addLog,
  };
});