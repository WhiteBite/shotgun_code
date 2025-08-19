<template>
  <div class="file-tree h-full flex flex-col">
    <!-- Search and Controls -->
    <div class="flex-shrink-0 p-2 border-b border-gray-700">
      <div class="flex items-center gap-2">
        <input
          v-model="fileTreeStore.searchQuery"
          type="text"
          placeholder="Поиск файлов..."
          class="flex-1 px-3 py-1.5 bg-gray-900 border border-gray-600 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          @click="refreshFiles"
          :disabled="fileTreeStore.isLoading"
          class="p-1.5 rounded-md hover:bg-gray-700 disabled:opacity-50"
          title="Обновить файлы"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            :class="{ 'animate-spin': fileTreeStore.isLoading }"
          >
            <polyline points="23 4 23 10 17 10"></polyline>
            <polyline points="1 20 1 14 7 14"></polyline>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
          </svg>
        </button>
      </div>
    </div>

    <!-- Tree Content -->
    <div class="flex-1 overflow-auto">
      <div v-if="fileTreeStore.isLoading" class="p-4 text-center text-gray-400">
        <div class="flex items-center justify-center gap-2">
          <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          Загрузка файлов...
        </div>
      </div>
      
      <div v-else-if="fileTreeStore.error" class="p-4 text-center text-red-400">
        <p class="text-sm">{{ fileTreeStore.error }}</p>
        <button
          @click="refreshFiles"
          class="mt-2 px-3 py-1 bg-red-600 hover:bg-red-500 rounded text-xs"
        >
          Повторить
        </button>
      </div>
      
      <div v-else-if="!fileTreeStore.hasFiles" class="p-4 text-center text-gray-400">
        <p class="text-sm">Файлы не найдены</p>
        <button
          @click="refreshFiles"
          class="mt-2 px-3 py-1 bg-blue-600 hover:bg-blue-500 rounded text-xs"
        >
          Повторить
        </button>
      </div>
      
      <div v-else class="p-2">
        <FileTreeNode
          v-for="node in fileTreeStore.visibleNodes"
          :key="node.path"
          :node="node"
          :depth="0"
          @select="handleNodeSelect"
          @expand="handleNodeExpand"
        />
      </div>
    </div>

    <!-- Summary -->
    <div v-if="fileTreeStore.hasFiles" class="flex-shrink-0 p-2 border-t border-gray-700 bg-gray-800/50">
      <div class="flex items-center justify-between text-xs text-gray-400">
        <span>{{ fileTreeStore.selectedFiles.length }} выбрано</span>
        <span>{{ fileTreeStore.totalFiles }} всего файлов</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useProjectStore } from '@/stores/project.store'
import { useTreeStateStore } from '@/stores/tree-state.store'
import FileTreeItem from './FileTreeItem.vue'
import FileTreeNode from './FileTreeNode.vue'

const fileTreeStore = useFileTreeStore()
const projectStore = useProjectStore()
const treeStateStore = useTreeStateStore()

async function refreshFiles() {
  if (!projectStore.currentProject?.path) {
    return
  }
  
  try {
    await fileTreeStore.refreshFiles()
  } catch (error) {
    console.error('Failed to refresh files:', error)
  }
}

function handleNodeSelect(node: any) {
  // Используем treeStateStore для переключения выделения
  treeStateStore.toggleNodeSelection(node.path, fileTreeStore.nodesMap as Map<string, any>)
}

function handleNodeExpand(node: any) {
  // Для директорий - разворачиваем/сворачиваем
  if (node.isDir) {
    treeStateStore.toggleExpansion(node.path)
  }
}
</script>

<style scoped>
.file-tree {
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
}
</style>