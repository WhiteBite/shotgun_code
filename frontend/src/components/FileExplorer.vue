<template>
  <div class="h-full flex flex-col bg-gray-900">
    <!-- Header -->
    <div class="flex items-center justify-between p-4 border-b border-gray-700">
      <div class="flex items-center gap-2">
        <svg class="h-5 w-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"></path>
        </svg>
        <h2 class="text-lg font-semibold text-white">Project Files</h2>
      </div>
      
      <div class="flex items-center gap-2 text-xs text-gray-400">
        <span class="bg-blue-900/30 text-blue-400 px-2 py-1 rounded">{{ fileStore.selectedCount }}</span>
        <span>selected</span>
      </div>
    </div>

    <!-- File Tree -->
    <div class="flex-1 overflow-auto p-2">
      <div v-if="fileStore.nodes.length === 0" class="flex items-center justify-center h-full text-gray-500">
        <p>No files loaded</p>
      </div>
      
      <FileTreeNode
        v-for="node in fileStore.nodes"
        :key="node.path"
        :node="node"
        @toggle-select="handleToggleSelect"
        @toggle-expand="handleToggleExpand"
      />
    </div>

    <!-- Bottom Panel -->
    <div class="border-t border-gray-700 p-4 space-y-2">
      <div class="flex items-center justify-between text-xs text-gray-400 mb-2">
        <span>Context Builder</span>
        <span>{{ fileStore.selectedCount }} files selected</span>
      </div>
      
      <button
        @click="handleBuildContext"
        :disabled="!fileStore.hasSelectedFiles || contextStore.isLoading"
        class="w-full px-4 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white rounded-lg transition-colors font-semibold flex items-center justify-center gap-2"
      >
        <svg v-if="contextStore.isLoading" class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span>Build Context</span>
        <span v-if="!contextStore.isLoading" class="text-xs opacity-60">Ctrl+Enter</span>
      </button>

      <button
        @click="fileStore.clearSelection"
        :disabled="!fileStore.hasSelectedFiles"
        class="w-full px-3 py-2 bg-gray-700 hover:bg-gray-600 disabled:bg-gray-800 disabled:cursor-not-allowed text-white text-sm rounded-lg transition-colors"
      >
        Clear Selection
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useFileStore } from '@/stores/file.store'
import { useContextStore } from '@/stores/context.store'
import { useUIStore } from '@/stores/ui.store'
import FileTreeNode from './FileTreeNode.vue'

const fileStore = useFileStore()
const contextStore = useContextStore()
const uiStore = useUIStore()

defineEmits<{
  (e: 'preview-file', filePath: string): void
}>()

function handleToggleSelect(path: string) {
  fileStore.toggleSelect(path)
}

function handleToggleExpand(path: string) {
  fileStore.toggleExpand(path)
}

async function handleBuildContext() {
  if (!fileStore.hasSelectedFiles) {
    uiStore.addToast('No files selected', 'warning')
    return
  }

  try {
    await contextStore.buildContext(fileStore.selectedFilesList)
    uiStore.addToast(`Context built with ${fileStore.selectedCount} files`, 'success')
  } catch (error) {
    uiStore.addToast('Failed to build context', 'error')
  }
}

// Keyboard shortcut: Ctrl+Enter
function handleKeydown(e: KeyboardEvent) {
  if (e.ctrlKey && e.key === 'Enter') {
    e.preventDefault()
    handleBuildContext()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>
