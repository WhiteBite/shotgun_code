<template>
  <RecycleScroller
    v-if="flattenedNodes.length > 0"
    class="h-full"
    :items="flattenedNodes"
    :item-size="28"
    key-field="id"
    v-slot="{ item }"
  >
    <VirtualTreeRow
      :item="item"
      :compact-mode="compactMode"
      :is-selected="isNodeSelected(item.node)"
      :checkbox-state="getCheckboxState(item.node)"
      :file-count="getFileCount(item.node)"
      @toggle-select="$emit('toggle-select', $event)"
      @toggle-expand="$emit('toggle-expand', $event)"
      @contextmenu="(node, event) => $emit('contextmenu', node, event)"
      @quicklook="$emit('quicklook', $event)"
    />
  </RecycleScroller>
  <div v-else class="empty-state h-full">
    <p class="empty-state-text">{{ t('files.noFiles') }}</p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useVirtualTree } from '@/composables/useVirtualTree'
import { computed, toRef } from 'vue'
import { RecycleScroller } from 'vue-virtual-scroller'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'
import { useFileStore, type FileNode } from '../model/file.store'
import VirtualTreeRow from './VirtualTreeRow.vue'

const { t } = useI18n()
const fileStore = useFileStore()

interface Props {
  nodes: FileNode[]
  compactMode?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  compactMode: false,
})

defineEmits<{
  (e: 'toggle-select', path: string): void
  (e: 'toggle-expand', path: string): void
  (e: 'contextmenu', node: FileNode, event: MouseEvent): void
  (e: 'quicklook', path: string): void
}>()

const nodesRef = toRef(props, 'nodes')
const { flattenedVisibleNodes } = useVirtualTree({ nodes: nodesRef })

const flattenedNodes = computed(() => flattenedVisibleNodes.value)

// Selection helpers - computed at parent level for better performance
function isNodeSelected(node: FileNode): boolean {
  if (!node.isDir) {
    return fileStore.selectedPaths.has(node.path)
  }
  // For directories, check if any files inside are selected
  const allFiles = fileStore.getAllFilesInNode(node)
  return allFiles.some(filePath => fileStore.selectedPaths.has(filePath))
}

function getCheckboxState(node: FileNode): 'none' | 'partial' | 'full' {
  if (!node.isDir) {
    return fileStore.selectedPaths.has(node.path) ? 'full' : 'none'
  }
  if (fileStore.selectedCount === 0) return 'none'
  
  const allFiles = fileStore.getAllFilesInNode(node)
  if (allFiles.length === 0) return 'none'
  
  let selectedCount = 0
  for (const filePath of allFiles) {
    if (fileStore.selectedPaths.has(filePath)) selectedCount++
  }
  
  if (selectedCount === 0) return 'none'
  if (selectedCount === allFiles.length) return 'full'
  return 'partial'
}

function getFileCount(node: FileNode): number {
  if (!node.isDir) return 0
  return fileStore.getAllFilesInNode(node).length
}
</script>
