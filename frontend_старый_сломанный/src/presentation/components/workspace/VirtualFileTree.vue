<template>
  <div class="virtual-file-tree overflow-hidden h-full flex flex-col">
    <div v-if="fileTreeStore.isLoading" class="p-4 text-center text-gray-400">
      Loading...
    </div>
    <div v-else-if="fileTreeStore.error" class="p-4 text-red-400">
      Error: {{ fileTreeStore.error }}
    </div>
    <div v-else-if="!fileTreeStore.hasFiles" class="p-4 text-center text-gray-500">
      No files in project.
    </div>
    <div v-else class="flex-1 overflow-hidden flex flex-col">
      <!-- Search bar -->
      <div class="p-2 border-b border-gray-700">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search files..."
          class="w-full px-3 py-1 bg-gray-800 text-gray-200 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
          @input="handleSearch"
        />
      </div>
      
      <!-- Virtual scroll container -->
      <div 
        ref="containerRef"
        class="flex-1 overflow-auto relative"
        @scroll="handleScroll"
      >
        <!-- Total height placeholder -->
        <div :style="{ height: totalHeight + 'px' }"></div>
        
        <!-- Visible items container -->
        <div 
          class="absolute top-0 left-0 right-0"
          :style="{ transform: `translateY(${offsetY}px)` }"
        >
          <FileTreeItem
            v-for="item in visibleItems"
            :key="item.node.path"
            :node="item.node"
            :depth="item.depth"
            @select="handleNodeSelect"
            @expand="handleNodeExpand"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useTreeStateStore } from '@/stores/tree-state.store'
import FileTreeItem from './FileTreeItem.vue'
import { useVirtualScroll } from '@/composables/useVirtualScroll'
import type { FileNode } from '@/types/dto'
import { FileTreeFlatteningService } from '@/domain/services/FileTreeFlatteningService'

// Services
const flatteningService = new FileTreeFlatteningService()

// Stores
const fileTreeStore = useFileTreeStore()
const treeStateStore = useTreeStateStore()

// Store refs
const { nodes, searchQuery } = storeToRefs(fileTreeStore)
const { expandedPaths } = storeToRefs(treeStateStore)

// State
const flattenedNodes = ref<Array<{ node: FileNode; depth: number }>>([])

// Virtual scroll setup
const ITEM_HEIGHT = 24 // Match the height from FileTreeItem
const CONTAINER_HEIGHT = ref(400) // Will be updated on mount

const {
  containerRef,
  visibleStart,
  visibleEnd,
  totalHeight,
  offsetY,
  updateTotalItems,
  handleScroll
} = useVirtualScroll({
  itemHeight: ITEM_HEIGHT,
  containerHeight: CONTAINER_HEIGHT.value,
  buffer: 5,
  totalItems: computed(() => flattenedNodes.value.length)
})

// Computed properties
const visibleItems = computed(() => {
  return flattenedNodes.value.slice(visibleStart.value, visibleEnd.value + 1)
})

// Methods
const flattenTree = () => {
  if (!nodes.value || nodes.value.length === 0) {
    flattenedNodes.value = []
    updateTotalItems(0)
    return
  }

  // Flatten the tree based on expanded paths
  const flatList = flatteningService.flattenTree(nodes.value, expandedPaths.value)
  flattenedNodes.value = flatList
  updateTotalItems(flatList.length)
}

const handleNodeSelect = (node: FileNode, event: MouseEvent) => {
  if (!node.isDir) {
    treeStateStore.toggleNodeSelection(
      node.path,
      fileTreeStore.nodesMap as unknown as Map<string, FileNode>
    )
  }
}

const handleNodeExpand = (node: FileNode) => {
  if (node.isDir) {
    treeStateStore.toggleExpansion(node.path)
  }
}

const handleSearch = () => {
  // Search is handled by the store, just re-flatten
  flattenTree()
}

// Watchers
watch([nodes, expandedPaths, searchQuery], () => {
  flattenTree()
}, { deep: true, immediate: true })

// Lifecycle
onMounted(() => {
  // Set container height based on actual element
  if (containerRef.value) {
    CONTAINER_HEIGHT.value = containerRef.value.clientHeight
    // Update virtual scroll config
    updateTotalItems(flattenedNodes.value.length)
  }
  
  // Initial flatten
  flattenTree()
})

// Expose for parent components
defineExpose({
  containerRef
})
</script>

<style scoped>
.virtual-file-tree {
  height: 100%;
}

.virtual-file-tree :deep(.file-tree-item) {
  height: 24px; /* Match ITEM_HEIGHT */
}
</style>