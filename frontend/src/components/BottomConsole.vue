<template>
  <div
      :style="{ height: height + 'px' }"
      class="bg-gray-800 text-white p-3 text-xs overflow-y-auto flex flex-col-reverse select-text"
      ref="consoleRootRef"
  >
    <div ref="consoleContentRef">
      <div v-for="(log, index) in notifications.logs" :key="index"
           :class="['whitespace-pre-wrap', 'break-words', getLogColor(log.type)]">
        <span class="font-medium mr-1">[{{ log.timestamp }}]</span>
        <span v-if="log.type !== 'info'" class="font-semibold mr-1">[{{ log.type.toUpperCase() }}]</span>
        <span>{{ log.message }}</span>
      </div>
      <div v-if="notifications.logs.length === 0" class="text-gray-500">
        Console is empty.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';
import { useNotificationsStore } from '@/stores/notifications.store';
import type { LogEntry } from '@/types/dto';

defineProps<{
  height: number
}>();

const notifications = useNotificationsStore();
const consoleRootRef = ref<HTMLElement | null>(null);

function getLogColor(type: LogEntry['type']) {
  switch (type) {
    case 'error': return 'text-red-400';
    case 'warn': return 'text-yellow-400';
    case 'success': return 'text-green-400';
    case 'info':
    default:
      return 'text-gray-300';
  }
}

watch(() => notifications.logs, () => {
  nextTick(() => {
    if (consoleRootRef.value) {
      // scroll to top because of flex-col-reverse
      consoleRootRef.value.scrollTop = 0;
    }
  });
}, { deep: true });
</script>