import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { useUiStore } from "./ui.store";
import type { LogEntry, LogType } from "@/types/dto";
import { APP_CONFIG } from '@/config/app-config';

let nextId = 1;

export const useNotificationsStore = defineStore("notifications", () => {
  const logs = ref<LogEntry[]>([]);
  const uiStore = useUiStore();

  function addLog(message: string, type: LogType = "info") {
    const newLog: LogEntry = {
      id: nextId++,
      message: message,
      type: type,
      timestamp: new Date().toLocaleTimeString(),
    };
    logs.value.unshift(newLog);
    if (logs.value.length > APP_CONFIG.performance.limits.MAX_CONSOLE_LOGS) {
      logs.value.pop();
    }
    if (type === "error" || type === "success" || type === "warning") {
      uiStore.addToast(message, type);
    }
  }

  const lastLine = computed(() => logs.value[0] ?? null);

  return { logs, addLog, lastLine };
});