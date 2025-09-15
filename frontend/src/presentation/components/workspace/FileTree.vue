<template>
  <div class="file-tree overflow-auto h-full">
    <div v-if="fileTreeStore.isLoading" class="p-4 text-center text-gray-400">
      Loading...
    </div>
    <div v-else-if="fileTreeStore.error" class="p-4 text-red-400">
      Error: {{ fileTreeStore.error }}
    </div>
    <div v-else-if="!fileTreeStore.hasFiles" class="p-4 text-center text-gray-500">
      No files in project.
    </div>
    <div v-else v-auto-animate>
      <FileTreeNode
        v-for="node in topLevelNodes"
        :key="node.path"
        :node="node"
        :depth="0"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
// import { computed } from "vue";
import FileTreeNode from "./FileTreeNode.vue";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useVisibleNodes } from "@/composables/useVisibleNodes";
// import type { FileNode } from "@/types/dto";

const fileTreeStore = useFileTreeStore();
const treeStateStore = useTreeStateStore();

const { visibleNodes: topLevelNodes } = useVisibleNodes();
</script>

<style scoped>
.file-tree {
  /* Add any specific styling for the file tree container */
}
</style>