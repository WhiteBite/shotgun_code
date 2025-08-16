<template>
  <div
    class="p-4 border border-gray-700 bg-gray-800/50 rounded-lg flex flex-col flex-grow min-h-0"
  >
    <div class="flex justify-between items-center mb-3">
      <h2 class="text-lg font-semibold text-white">Контекст проекта</h2>
    </div>
    <div
      class="flex-grow bg-gray-900 rounded-md p-3 overflow-auto border border-gray-700 min-h-0"
    >
      <pre
        v-if="contextBuilderStore.shotgunContextText"
        class="text-xs font-mono text-gray-300 whitespace-pre-wrap"
      ><code>{{ contextBuilderStore.shotgunContextText }}</code></pre>
      <div v-else class="flex items-center justify-center h-full text-gray-500">
        Контекст появится здесь после сборки.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useFileTreeStore } from '@/stores/file-tree.store';
import { useContextBuilderStore } from '@/stores/context-builder.store';
const fileTreeStore = useFileTreeStore();
const contextBuilderStore = useContextBuilderStore();
</script>

