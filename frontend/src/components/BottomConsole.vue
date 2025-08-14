<template>
  <div
    :style="{ height: height + 'px' }"
    class="bg-gray-800 text-white p-3 text-xs overflow-y-auto select-text border-t border-gray-700"
    ref="consoleRootRef"
  >
    <div ref="consoleContentRef" class="space-y-1">
      <div
        v-for="(log, index) in notifications.logs"
        :key="index"
        class="whitespace-pre-wrap break-words rounded border-l-4 pl-2 py-0.5"
        :class="rowClass(log.type)"
      >
        <span class="text-gray-400 mr-2">[{{ log.timestamp }}]</span>
        <span v-if="log.type !== 'info'" class="font-semibold mr-2 uppercase">{{
          log.type
        }}</span>
        <span>{{ log.message }}</span>
      </div>
      <div v-if="notifications.logs.length === 0" class="text-gray-500">
        Console is empty.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from "vue";
import { useNotificationsStore } from "@/stores/notifications.store";
import type { LogEntry } from "@/types/dto";

defineProps<{ height: number }>();

const notifications = useNotificationsStore();
const consoleRootRef = ref<HTMLElement | null>(null);

function rowClass(type: LogEntry["type"]) {
  switch (type) {
    case "error":
      return "border-red-500/70 bg-red-900/10 text-red-300";
    case "warn":
      return "border-yellow-500/70 bg-yellow-900/10 text-yellow-300";
    case "success":
      return "border-green-500/70 bg-green-900/10 text-green-300";
    case "info":
    default:
      return "border-gray-600 bg-gray-900/40 text-gray-300";
  }
}

watch(
  () => notifications.logs,
  () => {
    nextTick(() => {
      if (consoleRootRef.value) {
        consoleRootRef.value.scrollTop = consoleRootRef.value.scrollHeight;
      }
    });
  },
  { deep: true },
);
</script>
