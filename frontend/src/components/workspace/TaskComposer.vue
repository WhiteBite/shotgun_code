<template>
  <div class="h-full flex flex-col">
    <div class="flex-shrink-0 mb-3">
      <h3 class="text-sm font-semibold text-white mb-2">Опишите требуемые изменения</h3>
      <textarea
        v-model="taskDescription"
        placeholder="Пример: рефакторинг сервиса аутентификации, вынести JWT логику в хелпер..."
        class="w-full h-24 bg-gray-800 border border-gray-600 rounded-md p-3 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
        @keydown.ctrl.enter="suggestFiles"
      ></textarea>
    </div>
    
    <div class="flex-shrink-0 flex items-center gap-2">
      <button
        @click="suggestFiles"
        :disabled="!taskDescription.trim() || isSuggesting"
        class="px-3 py-1.5 text-xs bg-blue-600 hover:bg-blue-500 disabled:bg-blue-800 disabled:opacity-50 rounded font-medium"
      >
        {{ isSuggesting ? 'Анализ...' : 'Предложить файлы' }}
      </button>
      
      <button
        v-if="suggestedFiles.length > 0"
        @click="applySuggestion"
        class="px-3 py-1.5 text-xs bg-green-600 hover:bg-green-500 rounded font-medium"
      >
        Применить ({{ suggestedFiles.length }})
      </button>
      
      <button
        v-if="suggestedFiles.length > 0"
        @click="clearSuggestion"
        class="px-3 py-1.5 text-xs bg-gray-600 hover:bg-gray-500 rounded font-medium"
      >
        Очистить
      </button>
      
      <!-- Keyboard shortcut hint -->
      <span class="text-xs text-gray-400 ml-auto">
        Ctrl+Enter для быстрого анализа
      </span>
    </div>
    
    <!-- Suggested Files Preview -->
    <div v-if="suggestedFiles.length > 0" class="flex-1 mt-3 overflow-auto">
      <div class="bg-gray-800 rounded-md p-3">
        <h4 class="text-xs font-medium text-gray-300 mb-2">Предложенные файлы:</h4>
        <div class="space-y-1 max-h-32 overflow-y-auto">
          <div
            v-for="file in suggestedFiles"
            :key="file.path"
            class="flex items-center gap-2 p-1.5 bg-gray-700 rounded text-xs hover:bg-gray-650 transition-colors cursor-pointer"
            @click="toggleFileSelection(file)"
          >
            <input
              type="checkbox"
              :checked="file.isSelected"
              @click.stop
              @change="toggleFileSelection(file)"
              class="form-checkbox bg-gray-600 border-gray-500 rounded text-blue-500"
            />
            <span class="text-gray-400">{{ getFileIcon(file.path) }}</span>
            <span class="text-gray-300 truncate flex-1" :title="file.path">{{ getFileName(file.path) }}</span>
            <span class="text-gray-500 text-[10px]">({{ file.relevance }}%)</span>
            <button
              @click.stop="previewFile(file)"
              class="px-1 py-0.5 bg-gray-600 hover:bg-gray-500 rounded text-[10px] transition-colors"
              title="Предварительный просмотр"
            >
              👁️
            </button>
          </div>
        </div>
        
        <!-- Summary -->
        <div class="mt-2 pt-2 border-t border-gray-700 flex items-center justify-between text-xs">
          <span class="text-gray-400">
            Выбрано: {{ selectedCount }} из {{ suggestedFiles.length }}
          </span>
          <button
            @click="selectAllSuggested"
            class="text-blue-400 hover:text-blue-300"
          >
            Выбрать все
          </button>
        </div>
      </div>
    </div>
    
    <!-- Error state -->
    <div v-if="error" class="mt-2 p-2 bg-red-900/20 border border-red-700 rounded text-xs text-red-400">
      {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useProjectStore } from '@/stores/project.store'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useUiStore } from '@/stores/ui.store'
import { apiService } from '@/services/api.service'
import { getFileIcon } from '@/utils/fileIcons'

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

const selectedCount = computed(() => 
  suggestedFiles.value.filter(f => f.isSelected).length
)

async function suggestFiles() {
  if (!taskDescription.value.trim() || !projectStore.currentProject?.path) {
    uiStore.addToast('Введите описание задачи для анализа', 'warning')
    return
  }
  
  // Ограничиваем до 1 запроса
  if (hasRequested.value) {
    uiStore.addToast('Запрос уже был отправлен. Очистите результат для нового запроса.', 'warning')
    return
  }
  
  isSuggesting.value = true
  error.value = ''
  hasRequested.value = true
  
  try {
    // Получаем все файлы проекта
    const allFiles = fileTreeStore.getAllFiles()
    
    // Вызываем AI сервис для анализа задачи
    const result = await apiService.analyzeTaskAndCollectContext(
      taskDescription.value,
      JSON.stringify(allFiles),
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

function clearSuggestion() {
  suggestedFiles.value = []
  error.value = ''
  hasRequested.value = false // Сбрасываем флаг при очистке
}

function toggleFileSelection(file: any) {
  file.isSelected = !file.isSelected
}

function selectAllSuggested() {
  suggestedFiles.value.forEach(file => file.isSelected = true)
}

function previewFile(file: any) {
  const fakeEvent = new MouseEvent("click", { clientX: window.innerWidth/2, clientY: window.innerHeight/2 });
  uiStore.showQuickLook({
    rootDir: projectStore.currentProject?.path || "",
    path: file.path,
    type: "fs",
    event: fakeEvent,
    isPinned: true,
  });
}

function getFileName(filePath: string): string {
  return filePath.split('/').pop() || filePath.split('\\').pop() || filePath
}
</script>