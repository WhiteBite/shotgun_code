import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { LogEntry } from '@/types/dto';
import { useUiStore } from './ui.store';

export const useNotificationsStore = defineStore('notifications', () => {
  const logs = ref<LogEntry[]>([]);
  const maxLogs = 150;

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

    if (type === 'error' || type === 'success') {
      uiStore.addToast(message, type);
    }
  }

  return {
    logs,
    addLog,
  };
});