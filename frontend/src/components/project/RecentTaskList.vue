<template>
  <div class="bg-gray-800/50 p-6 rounded-lg border border-gray-700">
    <h2 class="text-xl font-semibold mb-4 text-white">Последние задачи</h2>
    <div v-if="projectStore.recentTasks.length > 0" class="space-y-3">
      <div v-for="task in projectStore.recentTasks" :key="task.id"
           @click="projectStore.loadTask(task.id)"
           class="p-3 bg-gray-700/50 hover:bg-gray-700/80 rounded-md cursor-pointer border border-transparent hover:border-blue-500 transition-all">
        <div class="flex justify-between items-start">
          <p class="font-semibold text-white">{{ task.name }}</p>
          <span class="text-xs text-gray-500">{{ task.date }}</span>
        </div>
        <p class="text-xs text-gray-400 mt-1">{{ task.contextSummary }}</p>
      </div>
    </div>
    <div v-else class="p-4 text-center text-gray-500">
      Нет недавних задач.
    </div>
  </div>
</template>

<script setup lang="ts">
import { useProjectStore } from '../../stores/projectStore';
const projectStore = useProjectStore();
</script>