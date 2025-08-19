<template>
  <aside class="w-80 bg-gray-800/60 p-3 border-r border-gray-700 flex flex-col flex-shrink-0">
    <div class="flex-shrink-0 mb-2">
      <h3 class="text-sm font-semibold text-white mb-2">Файлы</h3>
    </div>

    <div class="flex-grow bg-gray-900/50 rounded-md border border-gray-700 overflow-hidden min-h-0">
      <FileTree />
    </div>

    <!-- Context Summary -->
    <div class="flex-shrink-0 mt-2">
      <div class="p-2 bg-gray-900/50 rounded border border-gray-700">
        <h4 class="text-xs font-medium text-gray-400 mb-1">Сводка контекста</h4>
        <div class="text-xs text-gray-300">
          <div>Выбрано: {{ fileTreeStore.selectedFiles.length }}</div>
          <div>Всего: {{ fileTreeStore.totalFiles }}</div>
        </div>
      </div>
    </div>

    <!-- Ignore Rules -->
    <div class="flex-shrink-0 mt-2 space-y-2 pt-2 border-t border-gray-700/50">
      <div class="flex items-center justify-between">
        <h3 class="font-semibold text-xs text-gray-400">Правила игнорирования</h3>
        <button
          @click="openIgnoreDrawer"
          class="p-1 text-gray-400 hover:text-white transition-colors"
          title="Настроить правила игнорирования"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>
      </div>
      <div class="space-y-2 text-sm text-gray-300">
        <label class="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            v-model="fileTreeStore.useGitignore"
            @change="updateIgnoreRules"
            class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500"
          />
          Использовать .gitignore
        </label>
        <label class="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            v-model="fileTreeStore.useCustomIgnore"
            @change="updateIgnoreRules"
            class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500"
          />
          Пользовательские правила
        </label>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useUiStore } from '@/stores/ui.store'
import FileTree from '@/components/workspace/FileTree.vue'
import IgnoreDrawer from '@/components/drawers/IgnoreDrawer.vue'
import { useContextBuilderStore } from '@/stores/context-builder.store'

const fileTreeStore = useFileTreeStore()
const uiStore = useUiStore()
const contextBuilderStore = useContextBuilderStore()

async function updateIgnoreRules() {
  await fileTreeStore.refreshFiles()
}

function openIgnoreDrawer() {
  uiStore.openDrawer('ignore')
}
</script>