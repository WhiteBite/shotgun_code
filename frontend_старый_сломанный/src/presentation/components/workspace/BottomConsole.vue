﻿<template>
  <div v-if="uiStore.isConsoleVisible" class="bg-gray-900 border-t border-gray-700 text-xs font-mono text-gray-300 flex flex-col max-h-[33vh] min-h-[150px] overflow-hidden">
    <div class="flex items-center gap-2 p-2 border-b border-gray-700 flex-wrap">
      <label class="flex items-center gap-1"><input v-model="filters.levels.debug" type="checkbox"  /> debug</label>
      <label class="flex items-center gap-1"><input v-model="filters.levels.info" type="checkbox"  /> info</label>
      <label class="flex items-center gap-1"><input v-model="filters.levels.success" type="checkbox"  /> success</label>
      <label class="flex items-center gap-1"><input v-model="filters.levels.warning" type="checkbox"  /> warning</label>
      <label class="flex items-center gap-1"><input v-model="filters.levels.error" type="checkbox"  /> error</label>
      <span class="mx-2">|</span>
      <label class="flex items-center gap-1"><input v-model="filters.autoscroll" type="checkbox"  /> autoscroll</label>
      <button class="px-2 py-0.5 bg-gray-800 hover:bg-gray-700 rounded" @click="togglePause">{{ filters.paused ? 'Resume' : 'Pause' }}</button>
      <button class="px-2 py-0.5 bg-gray-800 hover:bg-gray-700 rounded" @click="clear">Clear</button>
    </div>
    <div ref="scroller" class="flex-1 overflow-y-auto p-2">
      <div v-for="log in filteredLogs" :key="log.id" :class="logColor(log.type)">
        [{{ log.timestamp }}] {{ log.message }}
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, ref, watch, nextTick } from "vue";
import { useUiStore } from "@/stores/ui.store";
import { useNotificationsStore } from "@/stores/notifications.store";
import type { LogType } from "@/types/dto";
const uiStore = useUiStore();
const notifications = useNotificationsStore();

const filters = computed(() => uiStore.consoleFilters);
const scroller = ref<HTMLElement | null>(null);

const filteredLogs = computed(() => {
  const lv = filters.value.levels;
  return notifications.logs.filter((l) => lv[l.type as keyof typeof lv]);
});

watch(
  () => [notifications.logs.length, filters.value.autoscroll, filters.value.paused],
  async () => {
    if (filters.value.autoscroll && !filters.value.paused) {
      await nextTick();
      scroller.value?.scrollTo({ top: scroller.value.scrollHeight });
    }
  },
);

function togglePause() {
  uiStore.togglePauseConsole();
}
function clear() {
  // Trust $reset for store cleanup
  notifications.$reset?.();
}

const logColor = (type: LogType) => ({
  "text-red-400": type === "error",
  "text-yellow-400": type === "warning",
  "text-green-400": type === "success",
});
</script>