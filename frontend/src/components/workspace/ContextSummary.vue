<template>
  <div class="p-3 bg-gray-900/50 rounded-md border border-gray-700 flex-shrink-0">
    <div class="grid grid-cols-2 gap-x-3 gap-y-1 text-xs mb-3">
      <div class="text-gray-400">Files:</div>
      <div class="text-white font-mono text-right">{{ contextStore.contextSummary.files }}</div>
      <div class="text-gray-400">Lines (est.):</div>
      <div class="text-white font-mono text-right">~{{ contextStore.contextSummary.lines.toLocaleString() }}</div>
      <div class="text-gray-400">Tokens (est.):</div>
      <div class="text-white font-mono text-right">~{{ contextStore.contextSummary.tokens.toLocaleString() }}</div>
    </div>
    <div class="space-y-2">
      <button @click="contextStore.clearSelection()" class="w-full py-1.5 bg-gray-700/50 hover:bg-gray-700 rounded-md font-semibold text-sm">
        Clear Selection
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useContextStore } from '@/stores/context.store';
const contextStore = useContextStore();
</script>