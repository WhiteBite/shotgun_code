<template>
  <div :class="panelClasses" :style="panelStyle">
    <!-- Header -->
    <div class="panel-header">
      <h3 v-if="!isCollapsed" class="text-lg font-semibold text-white flex items-center gap-2">
        <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        Context Builder
      </h3>
      <button 
        class="collapse-btn" 
        @click="toggleCollapse" 
        :title="isCollapsed ? 'Expand panel' : 'Collapse panel'"
        aria-label="Toggle panel"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
    </div>

    <!-- Content -->
    <div class="panel-content">
      <div class="panel-scrollable custom-scrollbar">
        <div class="space-y-4">
          <!-- Header with counter -->
          <div class="flex items-center justify-between">
            <h4 class="text-sm font-semibold text-gray-300">Selected Files</h4>
            <div class="bg-purple-600 text-white text-xs px-2 py-1 rounded-full">
              {{ selectedFilesCount }}
            </div>
          </div>
           
          <!-- Text Field with Selected Files -->
          <div class="space-y-2">
            <label class="block text-xs font-medium text-gray-400">
              Files List (editable)
            </label>
            <textarea
              v-model="selectedFilesText"
              @input="updateSelectedFilesFromText"
              placeholder="Выберите файлы в дереве или введите пути файлов (по одному на строку)..."
              class="w-full h-32 px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent resize-none font-mono transition-all custom-scrollbar"
            ></textarea>
            <div class="flex items-center justify-between text-xs text-gray-500">
              <span>{{ selectedFilesCount }} файлов • По одному пути на строку</span>
              <span v-if="selectedFilesCount > 0" class="text-green-400">✓</span>
            </div>
          </div>
          
          <!-- Empty state -->
          <div v-if="selectedFiles.length === 0" class="text-center py-8 border-2 border-dashed border-gray-600 rounded-lg">
            <svg class="w-12 h-12 mx-auto text-gray-500 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <p class="text-sm text-gray-500 mb-1">Файлы не выбраны</p>
            <p class="text-xs text-gray-600">Выберите файлы в дереве для создания контекста</p>
          </div>

          <!-- Selected files list -->
          <div v-else class="space-y-2">
            <div class="text-xs font-medium text-gray-400 mb-2">
              Выбранные файлы:
            </div>
            <div class="space-y-1 max-h-64 overflow-y-auto custom-scrollbar pr-2">
              <div
                v-for="file in selectedFiles"
                :key="file"
                class="flex items-center gap-2 p-2 bg-gray-700/50 hover:bg-gray-700 rounded-lg transition-colors group"
              >
                <span class="text-gray-400 text-sm">{{ getFileIcon(file) }}</span>
                <div class="flex-1 min-w-0">
                  <div class="text-sm text-gray-300 truncate">{{ getFileName(file) }}</div>
                  <div class="text-xs text-gray-500 truncate">{{ file }}</div>
                </div>
                <button
                  @click="removeFile(file)"
                  class="opacity-0 group-hover:opacity-100 text-gray-500 hover:text-red-400 transition-all p-1 rounded"
                  title="Удалить файл"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="panel-footer">
      <button
        @click="buildContext"
        :disabled="selectedFilesCount === 0 || contextBuilderStore.isBuilding"
        class="btn btn-primary btn-md w-full"
      >
        <svg v-if="contextBuilderStore.isBuilding" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        {{ contextBuilderStore.isBuilding ? 'Создание...' : 'Собрать контекст (Ctrl+Enter)' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { getFileIcon } from '@/utils/fileIcons'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useProjectStore } from '@/stores/project.store'
import { useFileTreeStore } from '@/stores/file-tree.store'
import usePanelManager from '@/composables/usePanelManager'

// Panel management
const { panelClasses, panelStyle, isCollapsed, toggleCollapse } = usePanelManager('context', 320)

// Stores
const fileTreeStore = useFileTreeStore()
const contextBuilderStore = useContextBuilderStore()
const projectStore = useProjectStore()
// Computed properties
const selectedFiles = computed(() => fileTreeStore.selectedFiles)
const selectedFilesCount = computed(() => selectedFiles.value.length)

// Text field for selected files
const selectedFilesText = computed({
  get: () => selectedFiles.value.join('\n'),
  set: (value: string) => {
    const newFiles = value
      .split('\n')
      .map(v => v.trim().replace(/\\/g, '/'))
      .filter(Boolean)
    fileTreeStore.setSelectedFiles(newFiles)
  }
})

// Methods
function getFileName(filePath: string): string {
  return filePath.split('/').pop() || filePath.split('\\').pop() || filePath
}

function removeFile(filePath: string) {
  const remain = selectedFiles.value.filter(f => f !== filePath)
  fileTreeStore.setSelectedFiles(remain)
}

function updateSelectedFilesFromText() {
  const lines = selectedFilesText.value.split('\n').map(v => v.trim()).filter(Boolean)
  fileTreeStore.setSelectedFiles(lines)
}

function buildContext() {
  if (projectStore.currentProject?.path) {
    contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
  }
}
</script>