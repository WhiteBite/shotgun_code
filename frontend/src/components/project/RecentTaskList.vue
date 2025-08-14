<template>
  <aside
    class="w-80 bg-gray-800/60 p-3 border-r border-gray-700 flex flex-col flex-shrink-0"
  >
    <div class="flex-shrink-0 mb-2">
      <input
        v-model="contextStore.searchQuery"
        type="text"
        placeholder="Поиск по файлам..."
        class="w-full px-3 py-1.5 bg-gray-900 border border-gray-600 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
    </div>

    <div
      class="flex-grow bg-gray-900/50 rounded-md border border-gray-700 overflow-hidden min-h-0"
    >
      <div v-if="contextStore.isLoading" class="p-4 text-center text-gray-400">
        <svg
          class="animate-spin h-6 w-6 mx-auto mb-2"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
          ></path>
        </svg>
        Сканирование проекта...
      </div>
      <div v-else-if="contextStore.error" class="p-4 text-center text-red-400">
        {{ contextStore.error }}
      </div>
      <FileTree v-else :nodes="visibleNodes" />
    </div>

    <div class="flex-shrink-0 mt-2">
      <ContextSummary />
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useContextStore } from "@/stores/context.store";
import FileTree from "../workspace/FileTree.vue";
import ContextSummary from "../workspace/ContextSummary.vue";
import { useVisibleNodes } from "@/composables/useVisibleNodes";

const contextStore = useContextStore();
const { visibleNodes } = useVisibleNodes();
</script>
