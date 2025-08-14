<template>
  <div
    class="h-14 bg-gray-800 border-t border-gray-700 flex-shrink-0 flex items-center justify-between px-4 text-sm text-gray-300"
  >
    <div class="flex items-center gap-4">
      <button
        @click="uiStore.toggleConsole()"
        class="p-2 rounded-md hover:bg-gray-700"
        title="Показать/скрыть консоль"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <polyline points="4 17 10 11 4 5"></polyline>
          <line x1="12" y1="19" x2="20" y2="19"></line>
        </svg>
      </button>
    </div>
    <div class="flex items-center gap-3">
      <button
        v-if="!contextStore.shotgunContextText"
        @click="contextStore.buildContext()"
        :disabled="contextStore.isLoading"
        class="px-5 py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-bold text-white flex items-center gap-2 disabled:opacity-50"
      >
        Собрать контекст
      </button>
      <button
        v-else
        @click="generationStore.executeGeneration()"
        :disabled="generationStore.isLoading"
        class="px-5 py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-bold text-white flex items-center gap-2 disabled:opacity-50"
      >
        <svg
          v-if="generationStore.isLoading"
          class="animate-spin -ml-1 mr-2 h-5 w-5"
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
        {{
          generationStore.isLoading ? "Генерация..." : "Сгенерировать решение"
        }}
        <svg
          v-if="!generationStore.isLoading"
          xmlns="http://www.w3.org/2000/svg"
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <line x1="5" y1="12" x2="19" y2="12"></line>
          <polyline points="12 5 19 12 12 19"></polyline>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useContextStore } from "@/stores/context.store";
import { useGenerationStore } from "@/stores/generation.store";
import { useUiStore } from "@/stores/ui.store";

const contextStore = useContextStore();
const generationStore = useGenerationStore();
const uiStore = useUiStore();
</script>
