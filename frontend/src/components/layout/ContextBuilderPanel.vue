<template>
  <aside class="w-80 bg-gray-800/60 p-3 border-r border-gray-700 flex flex-col flex-shrink-0">
    <!-- Header with Mode Toggle and Search -->
    <div class="flex-shrink-0 mb-2">
      <div class="flex items-center justify-between mb-2">
        <h3 class="font-semibold text-sm text-gray-300">Файлы проекта</h3>
        <div class="flex items-center gap-1 p-0.5 bg-gray-900/50 rounded-md">
          <button @click="contextStore.treeMode = TreeMode.Navigation" :class="['px-2 py-0.5 text-xs rounded', contextStore.treeMode === TreeMode.Navigation ? 'bg-blue-600 text-white' : 'text-gray-400 hover:bg-gray-700']">Навигация</button>
          <button @click="contextStore.treeMode = TreeMode.Selection" :class="['px-2 py-0.5 text-xs rounded', contextStore.treeMode === TreeMode.Selection ? 'bg-blue-600 text-white' : 'text-gray-400 hover:bg-gray-700']">Выбор</button>
        </div>
      </div>
      <input
          v-model="contextStore.searchQuery"
          type="text"
          placeholder="Поиск по файлам..."
          class="w-full px-3 py-1.5 bg-gray-900 border border-gray-600 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
    </div>

    <!-- File Tree -->
    <div class="flex-grow bg-gray-900/50 rounded-md border border-gray-700 overflow-hidden min-h-0">
      <div v-if="contextStore.isLoading" class="p-4 text-center text-gray-400">
        <svg class="animate-spin h-6 w-6 mx-auto mb-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path></svg>
        Сканирование проекта...
      </div>
      <div v-else-if="contextStore.error" class="p-4 text-center text-red-400">{{ contextStore.error }}</div>
      <FileTree v-else />
    </div>

    <!-- Other sections are restored and connected -->
    <div class="flex-shrink-0 mt-2 space-y-3 pt-2 border-t border-gray-700/50">
      <ContextSummary />
      <div v-if="gitStore.isAvailable">
        <h3 class="font-semibold text-sm mb-2 text-gray-300">Git</h3>
        <div class="space-y-2 text-sm">
          <button @click="gitStore.fetchUncommitted()" :disabled="gitStore.isLoading" class="w-full text-left p-2 bg-gray-900/50 hover:bg-gray-700/80 rounded-md">Незакоммиченные</button>
          <button @click="uiStore.openCommitModal()" :disabled="gitStore.isLoading" class="w-full text-left p-2 bg-gray-900/50 hover:bg-gray-700/80 rounded-md">Коммиты</button>
        </div>
      </div>
      <div>
        <h3 class="font-semibold text-sm mb-2 text-gray-300">Правила</h3>
        <div class="space-y-2 text-sm text-gray-300">
          <label class="flex items-center gap-2 cursor-pointer"><input type="checkbox" class="form-checkbox bg-gray-700 border-gray-500 rounded"> Использовать .gitignore</label>
          <label class="flex items-center gap-2 cursor-pointer"><input type="checkbox" class="form-checkbox bg-gray-700 border-gray-500 rounded"> Кастомные правила</label>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useContextStore } from '@/stores/contextStore';
import { useGitStore } from '@/stores/gitStore';
import { useUiStore } from '@/stores/uiStore';
import { TreeMode } from '@/types/enums';
import FileTree from '../workspace/FileTree.vue';
import ContextSummary from '../workspace/ContextSummary.vue';

const contextStore = useContextStore();
const gitStore = useGitStore();
const uiStore = useUiStore();
</script>