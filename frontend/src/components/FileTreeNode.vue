<template>
  <div class="select-none">
    <div
      :class="[
        'flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors',
        isSelected ? 'bg-blue-900/30 text-blue-400' : 'text-gray-300 hover:bg-gray-800'
      ]"
      :style="{ paddingLeft: `${depth * 16 + 8}px` }"
      @click="handleClick"
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
        :class="isSelected ? 'bg-blue-600 border-blue-600' : 'border-gray-600 hover:border-gray-500'"
        @click.stop="handleToggleSelect"
      >
        <svg v-if="isSelected" class="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
        </svg>
      </div>

      <!-- File/Folder Icon -->
      <div class="flex-shrink-0">
        <svg v-if="node.isDir" class="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
        </svg>
        <svg v-else class="w-5 h-5 text-gray-500" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z" clip-rule="evenodd" />
        </svg>
      </div>

      <!-- Name -->
      <span class="flex-1 truncate text-sm">{{ node.name }}</span>

      <!-- File count badge (for directories) -->
      <span v-if="node.isDir && node.children" class="text-xs text-gray-500 flex-shrink-0">
        {{ node.children.length }}
      </span>
    </div>

    <!-- Children (recursive) -->
    <div v-if="node.isDir && node.isExpanded && node.children">
      <FileTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
        @toggle-select="$emit('toggle-select', $event)"
        @toggle-expand="$emit('toggle-expand', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useFileStore, type FileNode } from '@/stores/file.store'

interface Props {
  node: FileNode
  depth?: number
}

const props = withDefaults(defineProps<Props>(), {
  depth: 0
})

const emit = defineEmits<{
  (e: 'toggle-select', path: string): void
  (e: 'toggle-expand', path: string): void
}>()

const fileStore = useFileStore()

const isSelected = computed(() => fileStore.selectedPaths.has(props.node.path))

function handleClick() {
  if (props.node.isDir) {
    emit('toggle-expand', props.node.path)
  }
}

function handleToggleSelect() {
  emit('toggle-select', props.node.path)
}
</script>
