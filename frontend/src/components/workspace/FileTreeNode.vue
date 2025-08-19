<template>
  <div>
    <FileTreeItem
      :node="node"
      :depth="depth"
      @select="$emit('select', $event)"
      @expand="$emit('expand', $event)"
    />
    <!-- Recursively render children if expanded -->
    <template v-if="node.children && treeStateStore.expandedPaths.has(node.path)">
      <FileTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
        @select="$emit('select', $event)"
        @expand="$emit('expand', $event)"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import type { FileNode } from '@/types/api'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useTreeStateStore } from '@/stores/tree-state.store'
import FileTreeItem from './FileTreeItem.vue'

const props = defineProps<{
  node: FileNode
  depth: number
}>()

const emit = defineEmits<{
  select: [node: FileNode]
  expand: [node: FileNode]
}>()

const fileTreeStore = useFileTreeStore()
const treeStateStore = useTreeStateStore()
</script>
