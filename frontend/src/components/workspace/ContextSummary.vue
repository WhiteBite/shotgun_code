<template>
  <div
      class="p-3 bg-gray-900/50 rounded-md border border-gray-700 flex-shrink-0"
  >
    <!-- Context Status Indicator -->
    <div class="mb-3 p-2 rounded-md border" :class="statusClasses">
      <div class="flex items-center gap-2">
        <div class="w-2 h-2 rounded-full" :class="statusDotClasses"></div>
        <span class="text-xs font-medium">{{
            contextBuilderStore.contextStatus.message
          }}</span>
      </div>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-2 gap-x-3 gap-y-1 text-xs mb-3">
      <div class="text-gray-400">Files:</div>
      <div class="text-white font-mono text-right">
        {{ contextBuilderStore.contextSummary.files }}
      </div>
      <div class="text-gray-400">Characters:</div>
      <div class="text-white font-mono text-right">
        {{ contextBuilderStore.contextSummary.characters.toLocaleString() }}
      </div>
      <div class="text-gray-400">Tokens (est.):</div>
      <div class="text-white font-mono text-right">
        ~{{ contextBuilderStore.contextSummary.tokens.toLocaleString() }}
      </div>
      <div class="text-gray-400">Cost (est.):</div>
      <div class="text-white font-mono text-right">
        ${{ contextBuilderStore.contextSummary.cost.toFixed(4) }}
      </div>
    </div>

    <!-- Actions -->
    <div class="space-y-2">
      <button
          @click="treeStateStore.clearSelection()"
          class="w-full px-3 py-1.5 bg-gray-700/50 hover:bg-gray-700/90 rounded-md text-sm"
      >
        Clear Selection
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useTreeStateStore } from "@/stores/tree-state.store";

const contextBuilderStore = useContextBuilderStore();
const treeStateStore = useTreeStateStore();

const statusClasses = computed(() => {
  const status = contextBuilderStore.contextStatus.status;
  switch (status) {
    case "current":
      return "border-green-500/30 bg-green-900/20";
    case "changed":
      return "border-yellow-500/30 bg-yellow-900/20";
    case "stale":
      return "border-orange-500/30 bg-orange-900/20";
    case "none":
    default:
      return "border-gray-500/30 bg-gray-900/20";
  }
});

const statusDotClasses = computed(() => {
  const status = contextBuilderStore.contextStatus.status;
  switch (status) {
    case "current":
      return "bg-green-500";
    case "changed":
      return "bg-yellow-500";
    case "stale":
      return "bg-orange-500";
    case "none":
    default:
      return "bg-gray-500";
  }
});
</script>