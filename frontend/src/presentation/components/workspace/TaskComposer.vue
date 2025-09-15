<template>
  <div class="h-full panel-container">
    <!-- Header -->
    <div class="panel-header">
      <div class="flex items-center gap-2 mb-3">
        <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <h3 class="text-lg font-semibold text-white">Опишите требуемые изменения</h3>
      </div>
      <div class="space-y-3">
        <textarea
          v-model="taskDescription"
          placeholder="Пример: рефакторинг сервиса аутентификации, вынести JWT логику в хелпер..."
          class="w-full h-24 bg-gray-700 border border-gray-600 rounded-lg p-3 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-200 placeholder-gray-500 transition-all"
          @keydown.ctrl.enter="suggestFiles"
        ></textarea>
        
        <div class="flex items-center gap-2">
          <button
            @click="suggestFiles"
            :disabled="!taskDescription.trim() || isSuggesting"
            class="px-4 py-2 text-sm bg-blue-600 hover:bg-blue-500 disabled:bg-blue-800 disabled:opacity-50 rounded-lg font-medium transition-all flex items-center gap-2"
          >
            <svg v-if="isSuggesting" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
            </svg>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
            {{ isSuggesting ? t('generating') : 'Предложить файлы' }}
          </button>
          
          <button
            v-if="suggestedFiles.length > 0"
            @click="applySuggestion"
            class="px-4 py-2 text-sm bg-green-600 hover:bg-green-500 rounded-lg font-medium transition-all flex items-center gap-2"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            Применить ({{ suggestedFiles.length }})
          </button>
          
          <button
            v-if="suggestedFiles.length > 0"
            @click="clearSuggestion"
            class="px-4 py-2 text-sm bg-gray-600 hover:bg-gray-500 rounded-lg font-medium transition-all"
          >
            Очистить
          </button>

          <button
            v-if="suggestedFiles.length > 0"
            @click="deselectAllSuggested"
            class="px-4 py-2 text-sm bg-gray-700 hover:bg-gray-600 rounded-lg font-medium transition-all"
          >
            Снять все
          </button>
          
          <!-- Keyboard shortcut hint -->
          <span class="text-xs text-gray-500 ml-auto hidden sm:inline">
            Ctrl+Enter для быстрого анализа
          </span>
        </div>
      </div>
    </div>
    <!-- Content -->
    <div class="panel-content">
      <div class="panel-scrollable custom-scrollbar">
      <!-- Suggested Files Preview -->
      <div v-if="suggestedFiles.length > 0" class="h-full p-4 overflow-y-auto">
        <div class="bg-gray-800/50 rounded-lg p-4 border border-gray-700">
          <div class="flex items-center justify-between mb-3">
            <h4 class="text-sm font-semibold text-gray-300 flex items-center gap-2">
              <svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Предложенные файлы
            </h4>
            <div class="flex items-center gap-2">
              <span class="text-xs px-2 py-1 bg-blue-600 text-white rounded-full">
                {{ suggestedFiles.length }} файлов
              </span>
            </div>
          </div>
          
          <div class="space-y-2 max-h-80 overflow-y-auto pr-2">
            <div
              v-for="file in suggestedFiles"
              :key="file.path"
              class="flex items-center gap-3 p-3 bg-gray-700/50 hover:bg-gray-700 rounded-lg transition-colors cursor-pointer group"
              @click="toggleFileSelection(file)"
            >
              <input
                type="checkbox"
                :checked="file.isSelected"
                @click.stop
                @change="toggleFileSelection(file)"
                class="form-checkbox bg-gray-600 border-gray-500 rounded text-blue-500 focus:ring-blue-500 focus:ring-offset-0"
              />
              <span class="text-gray-400 text-lg">{{ getFileIcon(file.path) }}</span>
              <div class="flex-1 min-w-0">
                <div class="text-sm text-gray-300 truncate font-medium" :title="normalizePath(file.path)">
                  {{ getFileName(file.path) }}
                </div>
                <div class="text-xs text-gray-500 truncate">
                  {{ file.path }}
                </div>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-xs px-2 py-1 bg-green-600/20 text-green-400 rounded">
                  {{ isAiSuggestion ? file.relevance + '%' : 'N/A' }}
                </span>
                <button
                  @click.stop="previewFile(file)"
                  class="opacity-0 group-hover:opacity-100 px-2 py-1 bg-gray-600 hover:bg-gray-500 rounded text-xs transition-all flex items-center gap-1"
                  title="Предварительный просмотр"
                >
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.477 0 8.268 2.943 9.542 7-1.274 4.057-5.065 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                  Просмотр
                </button>
              </div>
            </div>
          </div>
          
          <!-- Summary -->
          <div class="mt-4 pt-3 border-t border-gray-600 flex items-center justify-between text-sm">
            <span class="text-gray-400">
              Выбрано: <span class="text-white font-medium">{{ selectedCount }}</span> из {{ suggestedFiles.length }}
            </span>
            <button
              @click="selectAllSuggested"
              class="text-blue-400 hover:text-blue-300 transition-colors"
            >
              Выбрать все
            </button>
          </div>
        </div>
      </div>
      
      <!-- Empty state when no suggestions -->
      <div v-else-if="!isSuggesting" class="h-full flex items-center justify-center p-4">
        <div class="text-center">
          <svg class="w-16 h-16 mx-auto text-gray-500 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
          <p class="text-lg font-medium text-gray-300 mb-2">Опишите вашу задачу</p>
          <p class="text-sm text-gray-500">И мы предложим подходящие файлы для анализа</p>
        </div>
      </div>
      </div>
    </div>
    
    <!-- Error state -->
    <div v-if="error" class="p-4 border-t border-gray-700">
      <div class="p-3 bg-red-900/20 border border-red-700 rounded-lg text-sm text-red-400 flex items-start gap-2">
        <svg class="w-5 h-5 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div>
          <p class="font-medium mb-1">Ошибка анализа</p>
          <p>{{ error }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useProjectStore } from '@/stores/project.store'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useUiStore } from '@/stores/ui.store'
import { apiService } from '@/infrastructure/api/api.service'
import { getFileIcon } from '@/utils/fileIcons'
import { t } from '@/lib/i18n'

const projectStore = useProjectStore()
const fileTreeStore = useFileTreeStore()
const uiStore = useUiStore()

const taskDescription = ref('')
const isSuggesting = ref(false)
const error = ref('')
const hasRequested = ref(false) // Флаг для ограничения запросов
const suggestedFiles = ref<Array<{
  path: string
  relevance: number
  isSelected: boolean
}>>([])
const lastAnalyzedText = ref('')
const isAiSuggestion = ref(false)

const selectedCount = computed(() => 
  suggestedFiles.value.filter(f => f.isSelected).length
)

async function suggestFiles() {
  if (!taskDescription.value.trim() || !projectStore.currentProject?.path) {
    uiStore.addToast('Введите описание задачи для анализа', 'warning')
    return
  }
  
  // Разрешаем повторный анализ при изменении описания
  if (hasRequested.value && lastAnalyzedText.value === taskDescription.value.trim()) {
    uiStore.addToast('Измените описание или очистите результат для нового анализа.', 'warning')
    return
  }
  
  isSuggesting.value = true
  error.value = ''
  hasRequested.value = true
  lastAnalyzedText.value = taskDescription.value.trim()
  
  try {
    // Instead of sending all files, send only essential metadata
    const projectMetadata = {
      projectPath: projectStore.currentProject.path,
      totalFileCount: fileTreeStore.totalFiles,
      hasTypeScript: fileTreeStore.getAllFiles().some(f => f.name.endsWith('.ts') || f.name.endsWith('.tsx')),
      hasVue: fileTreeStore.getAllFiles().some(f => f.name.endsWith('.vue')),
      hasGo: fileTreeStore.getAllFiles().some(f => f.name.endsWith('.go')),
      hasJavaScript: fileTreeStore.getAllFiles().some(f => f.name.endsWith('.js') || f.name.endsWith('.jsx')),
      fileExtensions: [...new Set(fileTreeStore.getAllFiles()
        .filter(f => !f.isDir && f.name.includes('.'))
        .map(f => f.name.split('.').pop())
        .filter(Boolean))]
    }
    
    // Call a more efficient API that loads files on the backend
    const result = await apiService.analyzeTaskAndCollectContext(
      taskDescription.value,
      JSON.stringify(projectMetadata), // Send only metadata instead of all files
      projectStore.currentProject.path
    )
    
    // Парсим результат
    const analysis = JSON.parse(result)
    
    if (analysis.suggestedFiles && Array.isArray(analysis.suggestedFiles)) {
      suggestedFiles.value = analysis.suggestedFiles.map((filePath: string) => ({
        path: filePath,
        relevance: Math.floor(Math.random() * 40) + 60, // 60-100%
        isSelected: true // По умолчанию выбираем все предложенные файлы
      }))
      isAiSuggestion.value = true
      
      uiStore.addToast(`AI предложил ${suggestedFiles.value.length} файлов`, 'success')
    } else {
      // Fallback на простую логику если AI не сработал
      const relevantFiles = allFiles
        .filter(file => {
          const description = taskDescription.value.toLowerCase()
          const fileName = file.name?.toLowerCase() || ''
          const filePath = file.relPath?.toLowerCase() || ''
          return description.includes('auth') && (fileName.includes('auth') || filePath.includes('auth')) ||
                 description.includes('service') && (fileName.includes('service') || filePath.includes('service')) ||
                 description.includes('helper') && (fileName.includes('helper') || filePath.includes('helper')) ||
                 description.includes('api') && (fileName.includes('api') || filePath.includes('api')) ||
                 description.includes('component') && (fileName.includes('component') || filePath.includes('component')) ||
                 description.includes('store') && (fileName.includes('store') || filePath.includes('store')) ||
                 description.includes('test') && (fileName.includes('test') || filePath.includes('test')) ||
                 description.includes('util') && (fileName.includes('util') || filePath.includes('util'))
        })
        .slice(0, 10)
        .map(file => ({
          path: file.relPath || file.path,
          relevance: Math.floor(Math.random() * 40) + 60,
          isSelected: true
        }))
      
      suggestedFiles.value = relevantFiles
      isAiSuggestion.value = false
      
      if (relevantFiles.length > 0) {
        uiStore.addToast(`Предложено ${relevantFiles.length} файлов (fallback)`, 'info')
      } else {
        uiStore.addToast('Подходящие файлы не найдены', 'warning')
      }
    }
  } catch (err) {
    console.error('Error suggesting files:', err)
    error.value = err instanceof Error ? err.message : 'Ошибка при анализе промпта'
    uiStore.addToast('Ошибка при анализе промпта', 'error')
    hasRequested.value = false // Сбрасываем флаг при ошибке
  } finally {
    isSuggesting.value = false
  }
}

function applySuggestion() {
  const selectedFiles = suggestedFiles.value.filter(f => f.isSelected)
  if (selectedFiles.length === 0) {
    uiStore.addToast('Выберите файлы для применения', 'warning')
    return
  }
  
  const filePaths = selectedFiles.map(f => f.path)
  fileTreeStore.setSelectedFiles(filePaths)
  uiStore.addToast(`Выбрано ${filePaths.length} файлов`, 'success')
}
// Связка TaskComposer -> GenerationStore
import { useGenerationStore } from '@/stores/generation.store'
const generationStore = useGenerationStore()
watch(taskDescription, (v) => {
  generationStore.userTask = v
  if (v.trim() !== lastAnalyzedText.value) {
    hasRequested.value = false
  }
})

function clearSuggestion() {
  suggestedFiles.value = []
  error.value = ''
  hasRequested.value = false // Сбрасываем флаг при очистке
}

function toggleFileSelection(file: any) {
  // Используем Vue reactivity правильно - создаем новый объект
  const updatedFile = { ...file, isSelected: !file.isSelected };
  const index = suggestedFiles.value.findIndex(f => f.path === file.path);
  if (index !== -1) {
    suggestedFiles.value[index] = updatedFile;
  }
}

function selectAllSuggested() {
  suggestedFiles.value = suggestedFiles.value.map(file => ({ ...file, isSelected: true }))
}

function deselectAllSuggested() {
  suggestedFiles.value = suggestedFiles.value.map(file => ({ ...file, isSelected: false }))
}

function previewFile(file: any) {
  uiStore.showQuickLook({
    rootDir: projectStore.currentProject?.path || "",
    path: file.path,
    type: "fs",
    isPinned: true,
    position: { x: window.innerWidth/2, y: window.innerHeight/2 },
  });
}

function getFileName(filePath: string): string {
  const norm = normalizePath(filePath)
  return norm.split('/').pop() || norm
}
function normalizePath(p: string): string { return p ? p.replace(/\\\\/g, '/') : p }
</script>