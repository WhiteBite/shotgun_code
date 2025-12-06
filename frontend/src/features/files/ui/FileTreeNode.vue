<template>
  <div class="select-none">
    <div
      :class="[
        'flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors',
        isSelected ? 'bg-indigo-900/30 text-indigo-400' : 'text-gray-300 hover:bg-gray-800',
        { 'ignored-file': node.isIgnored }
      ]"
      :style="{ paddingLeft: `${depth * 16 + 8}px` }"
      @click="handleClick"
      @contextmenu.prevent="handleContextMenu"
      :title="node.isIgnored ? `${node.path} (ignored)` : node.path"
    >
      <!-- Expand/Collapse Icon (for directories) -->
      <div v-if="node.isDir" class="flex-shrink-0 w-4 h-4 flex items-center justify-center">
        <svg
          :class="['w-3 h-3 transition-transform', node.isExpanded ? 'rotate-90' : '']"
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
        </svg>
      </div>
      <div v-else class="w-4"></div>

      <!-- Checkbox -->
      <div
        class="flex-shrink-0 w-4 h-4 rounded border-2 flex items-center justify-center transition-colors"
        :class="checkboxClass"
        @click.stop="handleToggleSelect"
      >
        <svg v-if="isSelected && !node.isDir" class="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
        </svg>
        <svg v-else-if="node.isDir && selectionState === 'full'" class="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
        </svg>
        <div v-else-if="node.isDir && selectionState === 'partial'" class="w-2 h-0.5 bg-white"></div>
      </div>

      <!-- File/Folder Icon -->
      <div class="flex-shrink-0">
        <svg v-if="node.isDir" class="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
        </svg>
        <span v-else class="text-base">{{ getFileIcon(node.name) }}</span>
      </div>

      <!-- Ignored Icon with tooltip -->
      <div v-if="node.isIgnored" class="relative group">
        <svg class="w-4 h-4 text-gray-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
        </svg>
        <div class="absolute left-0 bottom-full mb-1 hidden group-hover:block z-50 px-2 py-1 bg-gray-800 text-xs text-white rounded whitespace-nowrap">
          Ignored by rules
        </div>
      </div>

      <!-- Name -->
      <span class="flex-1 truncate text-sm ignored-file" :class="{ 'opacity-50': node.isIgnored }">
        {{ displayName }}
      </span>

      <!-- File count badge (for directories) -->
      <span v-if="node.isDir && node.children" class="text-xs text-gray-500 flex-shrink-0" :title="`${fileCount} files in this folder`">
        {{ fileCount }}
      </span>
      
      <!-- File size (for files) -->
      <span v-else-if="node.size" class="text-xs text-gray-500 flex-shrink-0">
        {{ formatSize(node.size) }}
      </span>
    </div>

    <!-- Children (recursive) -->
    <div v-if="node.isDir && node.isExpanded && node.children">
      <FileTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
        :compact-mode="compactMode"
        @toggle-select="$emit('toggle-select', $event)"
        @toggle-expand="$emit('toggle-expand', $event)"
        @contextmenu="(childNode, event) => $emit('contextmenu', childNode, event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { getFileIcon } from '@/utils/fileIcons'
import { computed } from 'vue'
import { getCompactPath } from '../lib/file-utils'
import { useFileStore, type FileNode } from '../model/file.store'

interface Props {
  node: FileNode
  depth?: number
  compactMode?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  depth: 0,
  compactMode: false
})

const emit = defineEmits<{
  (e: 'toggle-select', path: string): void
  (e: 'toggle-expand', path: string): void
  (e: 'contextmenu', node: FileNode, event: MouseEvent): void
}>()

const fileStore = useFileStore()

const isSelected = computed(() => fileStore.selectedPaths.has(props.node.path))

const displayName = computed(() => {
  if (!props.compactMode || !props.node.isDir || !props.node.children) {
    return props.node.name
  }
  
  // Only use compact path if there's a single nested folder chain
  const compactPath = getCompactPath(props.node)
  return compactPath !== props.node.name ? compactPath : props.node.name
})

// Optimize: only calculate for expanded directories
const selectionState = computed(() => {
  if (!props.node.isDir) {
    return isSelected.value ? 'full' : 'none'
  }

  // CRITICAL OPTIMIZATION: For collapsed folders, use cached count instead of full traversal
  if (!props.node.isExpanded) {
    // Quick check: if no selected paths at all, return none
    if (fileStore.selectedCount === 0) {
      return 'none'
    }
    
    // Use the store method which is already optimized
    const allFiles = fileStore.getAllFilesInNode(props.node)
    if (allFiles.length === 0) {
      return 'none'
    }
    
    // Quick count using filter on already retrieved array
    const selectedCount = allFiles.filter(filePath => fileStore.selectedPaths.has(filePath)).length
    
    if (selectedCount === 0) {
      return 'none'
    } else if (selectedCount === allFiles.length) {
      return 'full'
    } else {
      return 'partial'
    }
  }

  // For expanded folders, do full calculation
  const allFiles = fileStore.getAllFilesInNode(props.node)
  if (allFiles.length === 0) {
    return isSelected.value ? 'full' : 'none'
  }

  // Оптимизация: считаем выбранные файлы за один проход по уже полученному массиву
  const selectedCount = allFiles.filter(filePath => fileStore.selectedPaths.has(filePath)).length
  
  if (selectedCount === 0) {
    return 'none'
  } else if (selectedCount === allFiles.length) {
    return 'full'
  } else {
    return 'partial'
  }
})

const checkboxClass = computed(() => {
  if (props.node.isDir) {
    switch (selectionState.value) {
      case 'full':
        return 'bg-indigo-600 border-indigo-600'
      case 'partial':
        return 'bg-indigo-600 border-indigo-600'
      case 'none':
        return 'border-gray-600 hover:border-gray-500'
    }
  } else {
    return isSelected.value ? 'bg-indigo-600 border-indigo-600' : 'border-gray-600 hover:border-gray-500'
  }
})

// File count for directories - only calculate when expanded to avoid performance issues
const fileCount = computed(() => {
  if (!props.node.isDir || !props.node.isExpanded) return 0
  return fileStore.getAllFilesInNode(props.node).length
})

function handleClick() {
  if (props.node.isDir) {
    emit('toggle-expand', props.node.path)
  }
}

function handleToggleSelect() {
  emit('toggle-select', props.node.path)
}

function handleContextMenu(event: MouseEvent) {
  emit('contextmenu', props.node, event)
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return Math.round(bytes / 1024) + ' KB'
  return Math.round(bytes / (1024 * 1024)) + ' MB'
}
</script>
