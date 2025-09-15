<template>
  <div class="file-tree-enhanced">
    <div v-if="!nodes || nodes.length === 0" class="empty-message">
      <p>No files to display</p>
    </div>
    <div v-else class="tree-content">
      <!-- Use virtualized file tree instead of recursive rendering -->
      <VirtualFileTree ref="virtualFileTree" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import VirtualFileTree from '@/presentation/components/workspace/VirtualFileTree.vue'
import type { FileNode } from '@/types/dto'
import { FileTreeFilteringService } from '@/domain/services/FileTreeFilteringService'

// Services
const filteringService = new FileTreeFilteringService()

interface Props {
  nodes: FileNode[]
  searchQuery?: string
  showHidden?: boolean
  selectedFiles?: string[]
  expandedFolders?: Set<string>
  parentDepth?: number
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: '',
  showHidden: false,
  selectedFiles: () => [],
  expandedFolders: () => new Set(),
  parentDepth: 0
})

const emit = defineEmits<{
  'file-select': [filePath: string, isSelected: boolean]
  'folder-toggle': [folderPath: string, isExpanded: boolean]
  'file-context-menu': [filePath: string, event: MouseEvent]
}>()

const virtualFileTree = ref<InstanceType<typeof VirtualFileTree> | null>(null)

// We'll handle the filtering and other logic through the store instead of locally
// The VirtualFileTree component will use the store directly
</script>

<style scoped>
.file-tree-enhanced {
  width: 100%;
  height: 100%;
}

.tree-content {
  width: 100%;
  height: 100%;
}

.empty-message {
  padding: 20px;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 0.875rem;
}
</style>