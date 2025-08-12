<template>
  <div class="p-4 border border-gray-700 bg-gray-800/50 rounded-lg flex flex-col flex-grow min-h-0">
    <div class="flex justify-between items-center mb-3">
      <h2 class="text-lg font-semibold text-white">
        {{ tab === 'context' ? 'Project Context' : 'Generated Diff' }}
      </h2>
      <div class="flex items-center gap-1 p-0.5 bg-gray-900/50 rounded-md text-sm">
        <button @click="tab = 'context'" :class="['px-3 py-1 rounded-md', tab === 'context' ? 'bg-blue-600' : 'hover:bg-gray-700']">Context</button>
        <button @click="tab = 'result'" :disabled="!generationStore.hasResult" :class="['px-3 py-1 rounded-md', tab === 'result' ? 'bg-blue-600' : 'hover:bg-gray-700 disabled:text-gray-600 disabled:cursor-not-allowed']">Result</button>
      </div>
    </div>
    <div class="flex-grow bg-gray-900 rounded-md p-3 overflow-auto border border-gray-700 min-h-0">
      <!-- Context View -->
      <pre v-if="tab === 'context'" class="text-xs font-mono text-gray-300 whitespace-pre-wrap"><code>{{ contextStore.shotgunContextText || 'Context will appear here after being built.' }}</code></pre>

      <!-- Result View -->
      <pre v-if="tab === 'result'" class="text-sm font-mono whitespace-pre-wrap"><code v-html="highlightedDiff"></code></pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useContextStore } from '@/stores/context.store';
import { useGenerationStore } from '@/stores/generation.store';

const tab = ref<'context' | 'result'>('context');
const contextStore = useContextStore();
const generationStore = useGenerationStore();

const highlightedDiff = computed(() => {
  if (!generationStore.generatedDiff) return '<span class="text-gray-500">No diff generated yet.</span>';
  return generationStore.generatedDiff
  .split('\n')
  .map(line => {
    const sanitizedLine = line.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
    if (sanitizedLine.startsWith('+')) return `<span class="text-green-400">${sanitizedLine}</span>`;
    if (sanitizedLine.startsWith('-')) return `<span class="text-red-400">${sanitizedLine}</span>`;
    if (sanitizedLine.startsWith('@@')) return `<span class="text-cyan-400">${sanitizedLine}</span>`;
    return `<span class="text-gray-400">${sanitizedLine}</span>`;
  })
  .join('\n');
});

watch(() => generationStore.hasResult, (hasResult) => {
  if (hasResult) {
    tab.value = 'result';
  }
});
</script>