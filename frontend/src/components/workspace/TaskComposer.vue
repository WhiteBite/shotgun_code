<template>
  <div class="p-4 border border-gray-700 bg-gray-800/50 rounded-lg">
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-lg font-semibold text-white">Опишите требуемые изменения</h2>
      <button @click="taskStore.suggestContext()" :disabled="taskStore.isSuggesting" class="text-sm text-blue-400 hover:underline disabled:opacity-50">
        {{ taskStore.isSuggesting ? 'Анализ...' : 'Предложить файлы' }}
      </button>
    </div>
    <textarea
        v-model="taskStore.userTask"
        placeholder="Пример: рефакторинг сервиса аутентификации, вынести JWT в хелпер..."
        class="w-full h-32 p-3 bg-gray-900 border border-gray-600 rounded-md text-base text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-y"
    ></textarea>
  </div>
</template>

<script setup lang="ts">
import { useTaskStore } from '@/stores/taskStore';
const taskStore = useTaskStore();
</script>