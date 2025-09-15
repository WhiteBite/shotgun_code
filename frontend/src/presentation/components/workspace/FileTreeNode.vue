<template>
  <div>
    <FileTreeItem
      :node="node"
      :depth="depth"
      @select="handleNodeClick"
      @expand="handleToggleExpand"
    />
    <!-- Recursively render children if expanded -->
    <template
      v-if="node.children && isNodeExpanded"
    >
      <FileTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
        @node-click="handleNodeClick"
        @toggle-expand="handleToggleExpand"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { FileNode } from "@/types/dto";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import FileTreeItem from "./FileTreeItem.vue";

const props = defineProps<{
  node: FileNode;
  depth: number;
}>();

const emit = defineEmits<{
  'node-click': [node: FileNode, event: MouseEvent];
  'toggle-expand': [node: FileNode];
}>();

const fileTreeStore = useFileTreeStore();
const treeStateStore = useTreeStateStore();

// Используем computed для реактивности
const isNodeExpanded = computed(() => {
  return treeStateStore.expandedPaths.has(props.node.path);
});

// Обработчики событий от FileTreeItem
function handleNodeClick(node: FileNode, event: MouseEvent) {
  emit('node-click', node, event);
}

function handleToggleExpand(node: FileNode) {
  emit('toggle-expand', node);
}
</script>