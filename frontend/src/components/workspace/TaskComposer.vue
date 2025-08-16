<template>
  <div class="p-4 border border-gray-700 bg-gray-800/50 rounded-lg flex flex-col h-full">
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-lg font-semibold text-white">
        Describe Required Changes
      </h2>
    </div>
    <textarea
        v-model="generationStore.userTask"
        placeholder="Example: refactor authentication service, extract JWT logic to a helper..."
        class="w-full flex-grow p-3 bg-gray-900 border border-gray-600 rounded-md text-base text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-y"
    ></textarea>
    <div class="flex-shrink-0 pt-3 mt-3 border-t border-gray-700 flex items-center">
      <button
          @click="generationStore.suggestContext()"
          :disabled="!generationStore.userTask.trim() || generationStore.isLoading"
          class="px-3 py-1.5 bg-gray-700 hover:bg-gray-600 text-sm rounded-md disabled:opacity-50"
      >
        Suggest Files
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useGenerationStore } from "@/stores/generation.store";
const generationStore = useGenerationStore();
</script>