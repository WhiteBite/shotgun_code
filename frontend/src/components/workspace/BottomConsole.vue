<template>
  <div v-if='uiStore.isConsoleVisible' class='h-48 bg-gray-900 border-t border-gray-700 p-2 text-xs font-mono text-gray-400 overflow-y-auto'>
    <div v-for='log in notifications.logs' :key='log.id' :class='logColor(log.type)'>
      [{{ log.timestamp }}] {{ log.message }}
    </div>
  </div>
</template>
<script setup lang='ts'>
import { useUiStore } from '@/stores/ui.store';
import { useNotificationsStore } from '@/stores/notifications.store';
import type { LogType } from '@/types/dto';
const uiStore = useUiStore();
const notifications = useNotificationsStore();
const logColor = (type: LogType) => ({
  'text-red-400': type === 'error',
  'text-yellow-400': type === 'warn',
  'text-green-400': type === 'success',
});
</script>
