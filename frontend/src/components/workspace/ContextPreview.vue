<template>
  <div class="context-preview h-full flex flex-col bg-gray-900">
    <!-- Header -->
    <div class="flex-shrink-0 p-4 border-b border-gray-700 bg-gray-800">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-medium text-white">Контекст</h3>
        
        <div class="flex items-center gap-3">
          <!-- Build Context Button -->
          <button
            v-if="contextBuilderStore.selectedFilesCount > 0"
            @click="buildContext"
            :disabled="contextBuilderStore.isBuilding"
            class="px-3 py-1 text-xs bg-blue-600 hover:bg-blue-500 disabled:bg-blue-800 rounded font-medium"
          >
            {{ contextBuilderStore.isBuilding ? 'Построение...' : 'Построить контекст' }}
          </button>
          
          <!-- Simple Split Button -->
          <button
            v-if="contextBuilderStore.currentContext"
            @click="toggleSimpleSplit"
            class="px-3 py-1 bg-blue-600 hover:bg-blue-500 rounded text-xs transition-colors"
            :class="{ 'bg-green-600 hover:bg-green-500': isSimpleSplitEnabled }"
            title="Разделить контекст на части"
          >
            {{ isSimpleSplitEnabled ? 'Сплит ВКЛ' : 'Сплит ВЫКЛ' }}
          </button>
          
          <!-- Keyboard Shortcuts Icon -->
          <div class="relative group">
            <button
              @click="showShortcuts"
              class="p-1 text-gray-400 hover:text-white transition-colors"
              title="Горячие клавиши"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </button>
            <!-- Tooltip -->
            <div
              v-if="showShortcutsTooltip"
              class="absolute bottom-full right-0 mb-2 px-3 py-2 bg-gray-900 text-white text-xs rounded-lg shadow-lg whitespace-nowrap z-50 border border-gray-700"
              @click.stop
            >
              <div class="space-y-1">
                <div class="font-medium text-gray-300 mb-2">Горячие клавиши:</div>
                <div class="grid grid-cols-2 gap-x-4 gap-y-1">
                  <div><kbd class="bg-gray-700 px-1 rounded text-[10px]">Ctrl+K</kbd> <span class="text-[10px]">Поиск файлов</span></div>
                  <div><kbd class="bg-gray-700 px-1 rounded text-[10px]">Space</kbd> <span class="text-[10px]">QuickLook (закрепленный)</span></div>
                  <div><kbd class="bg-gray-700 px-1 rounded text-[10px]">Ctrl</kbd> <span class="text-[10px]">QuickLook (временный)</span></div>
                  <div><kbd class="bg-gray-700 px-1 rounded text-[10px]">Alt+→</kbd> <span class="text-[10px]">Развернуть рекурсивно</span></div>
                  <div><kbd class="bg-gray-700 px-1 rounded text-[10px]">Alt+←</kbd> <span class="text-[10px]">Свернуть рекурсивно</span></div>
                  <div><kbd class="bg-gray-700 px-1 rounded text-[10px]">Ctrl+Enter</kbd> <span class="text-[10px]">Построить/Генерировать</span></div>
                  <div><kbd class="bg-gray-700 px-1 rounded text-[10px]">Ctrl+A</kbd> <span class="text-[10px]">Выбрать все файлы</span></div>
                  <div><kbd class="bg-gray-700 px-1 rounded text-[10px]">Ctrl+D</kbd> <span class="text-[10px]">Снять выделение</span></div>
                  <div><kbd class="bg-gray-700 px-1 rounded text-[10px]">Esc</kbd> <span class="text-[10px]">Закрыть модалы</span></div>
                </div>
              </div>
              <div class="absolute top-full right-4 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900"></div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Status -->
      <div class="flex items-center justify-between text-xs mt-2">
        <div class="flex items-center gap-4">
          <span v-if="contextBuilderStore.selectedFilesCount > 0" class="text-gray-400">
            {{ contextBuilderStore.selectedFilesCount }} файлов выбрано
          </span>
          <span v-if="contextBuilderStore.currentContext" class="flex items-center gap-1 text-green-400">
            <div class="w-2 h-2 bg-green-400 rounded-full"></div>
            Контекст построен
          </span>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-hidden">
      <!-- Loading State -->
      <div v-if="contextBuilderStore.isBuilding" class="p-4 text-center text-gray-400">
        <div class="flex items-center justify-center gap-2">
          <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          Построение контекста...
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="contextBuilderStore.error" class="p-4 text-center text-red-400">
        <p class="text-sm">{{ contextBuilderStore.error }}</p>
        <button
          @click="clearError"
          class="mt-2 px-3 py-1 bg-red-600 hover:bg-red-500 rounded text-xs"
        >
          Очистить ошибку
        </button>
      </div>

      <!-- Empty State -->
      <div v-else-if="contextBuilderStore.selectedFilesCount === 0" class="p-4 text-center text-gray-400">
        <div class="flex flex-col items-center gap-2">
          <svg class="w-8 h-8 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p class="text-sm">Выберите файлы для построения контекста</p>
        </div>
      </div>

      <!-- Success State -->
      <div v-else class="h-full flex flex-col">
        <!-- Context Content -->
        <div class="flex-1 overflow-auto p-4">
          <div class="relative h-full">
            <!-- Action buttons -->
            <div class="absolute top-2 right-2 z-10 flex items-center gap-2">
              <button
                @click="saveContext"
                class="p-2 bg-gray-800/50 hover:bg-gray-700 rounded transition-colors"
                title="Сохранить контекст"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
                </svg>
              </button>
              <button
                @click="copyContext"
                class="p-2 bg-gray-800/50 hover:bg-gray-700 rounded transition-colors"
                title="Копировать в буфер"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              </button>
            </div>
            
            <!-- Context Content -->
            <div class="pr-20 h-full overflow-auto">
              <pre class="whitespace-pre-wrap text-sm text-gray-300 font-mono leading-relaxed p-4">{{ contextBuilderStore.currentContext?.content }}</pre>
            </div>
          </div>
        </div>
        
        <!-- Simple Split Chunks -->
        <div v-if="hasSplitContent" class="flex-shrink-0 border-t border-gray-700 p-4 bg-gray-800">
          <div class="flex items-center justify-between mb-3">
            <h4 class="text-sm font-medium text-white">Разделение для копирования в ИИ</h4>
            <div class="flex items-center gap-2">
              <span class="text-xs text-gray-400">{{ splitChunks.length || 0 }} частей</span>
              <button
                @click="copyAllChunks"
                class="px-2 py-1 bg-green-600 hover:bg-green-500 rounded text-xs"
                title="Копировать все части"
              >
                Копировать все
              </button>
            </div>
          </div>
          
          <!-- Simple Split chunks -->
          <div class="space-y-2 max-h-64 overflow-y-auto">
            <div
              v-for="(chunk, index) in splitChunks"
              :key="index"
              class="bg-gray-900 rounded border border-gray-700 p-3"
            >
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-3">
                  <span class="text-sm font-medium text-blue-400">Часть {{ index + 1 }}</span>
                  <span class="text-xs text-gray-500">{{ chunk.chars }} символов</span>
                </div>
                <button
                  @click="copyChunk(index)"
                  class="px-3 py-1 bg-blue-600 hover:bg-blue-500 rounded text-xs transition-colors"
                  title="Копировать часть"
                >
                  Копировать
                </button>
              </div>
              
              <!-- Chunk Content Preview -->
              <div class="text-xs text-gray-300 font-mono bg-gray-800 p-2 rounded border border-gray-600 max-h-24 overflow-y-auto">
                {{ chunk.text.substring(0, 300) }}{{ chunk.text.length > 300 ? '...' : '' }}
              </div>
            </div>
          </div>
          
          <!-- Simple Instructions -->
          <div class="mt-3 p-2 bg-blue-900/20 border border-blue-700 rounded text-xs text-blue-300">
            <div class="font-medium mb-1">💡 Как использовать:</div>
            <div>1. Нажмите "Копировать" на нужной части</div>
            <div>2. Вставьте в ИИ (Ctrl+V)</div>
            <div>3. Отправьте сообщение</div>
            <div>4. Повторите для следующей части</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useProjectStore } from '@/stores/project.store'
import { useUiStore } from '@/stores/ui.store'

const contextBuilderStore = useContextBuilderStore()
const projectStore = useProjectStore()
const uiStore = useUiStore()

// Simple split functionality
const isSimpleSplitEnabled = ref(false)
const splitChunks = ref<any[]>([])

const hasSplitContent = computed(() => 
  isSimpleSplitEnabled.value && 
  contextBuilderStore.currentContext?.content && 
  splitChunks.value.length > 1
)

const showShortcutsTooltip = ref(false)

function showShortcuts() {
  showShortcutsTooltip.value = !showShortcutsTooltip.value
}

// Закрытие тултипа при клике вне его
function handleClickOutside(event: Event) {
  const target = event.target as HTMLElement
  if (!target.closest('.shortcuts-tooltip') && !target.closest('.shortcuts-button')) {
    showShortcutsTooltip.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

async function buildContext() {
  if (!projectStore.currentProject?.path) {
    return
  }
  
  try {
    await contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
  } catch (error) {
    console.error('Failed to build context:', error)
  }
}

function clearError() {
  contextBuilderStore.clearError()
}

async function copyContext() {
  if (!contextBuilderStore.currentContext?.content) {
    return
  }
  await navigator.clipboard.writeText(contextBuilderStore.currentContext.content)
  uiStore.addToast('Контекст скопирован в буфер обмена', 'success')
}

async function saveContext() {
  if (!contextBuilderStore.currentContext?.content) {
    uiStore.addToast('Нет контекста для сохранения', 'error')
    return
  }
  
  try {
    // Создаем файл для скачивания
    const blob = new Blob([contextBuilderStore.currentContext.content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `context_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    
    uiStore.addToast('Контекст сохранен в файл', 'success')
  } catch (error) {
    uiStore.addToast('Ошибка при сохранении контекста', 'error')
  }
}

function toggleSimpleSplit() {
  isSimpleSplitEnabled.value = !isSimpleSplitEnabled.value
  if (isSimpleSplitEnabled.value && contextBuilderStore.currentContext?.content) {
    computeSimpleSplit()
  }
}

function computeSimpleSplit() {
  const text = contextBuilderStore.currentContext?.content || ""
  if (!text) {
    splitChunks.value = []
    return
  }
  
  // Простой split по символам
  const maxChars = 8000 // 8k символов на часть
  const overlapChars = 500 // 500 символов перекрытия
  
  if (text.length <= maxChars) {
    // Если текст короткий, не разделяем
    splitChunks.value = [{
      index: 0,
      text: text,
      chars: text.length
    }]
  } else {
    // Разделяем по символам
    const chunks = []
    let start = 0
    let index = 0
    
    while (start < text.length) {
      const end = Math.min(start + maxChars, text.length)
      const chunkText = text.substring(start, end)
      
      chunks.push({
        index: index++,
        text: chunkText,
        chars: chunkText.length
      })
      
      start = end - overlapChars
      if (start >= text.length) break
    }
    
    splitChunks.value = chunks
  }
}

function copyChunk(chunkIndex: number) {
  if (!splitChunks.value[chunkIndex]) {
    return
  }
  navigator.clipboard.writeText(splitChunks.value[chunkIndex].text)
  uiStore.addToast(`Часть ${chunkIndex + 1} скопирована в буфер обмена`, 'success')
}

function copyAllChunks() {
  if (!splitChunks.value.length) {
    return
  }
  const allText = splitChunks.value.map((chunk, index) => 
    `=== ЧАСТЬ ${index + 1} ===\n${chunk.text}\n\n`
  ).join('')
  navigator.clipboard.writeText(allText)
  uiStore.addToast('Все части скопированы в буфер обмена', 'success')
}

// Watch for context changes to update split
import { watch } from 'vue'
watch(() => contextBuilderStore.currentContext?.content, (newContent) => {
  if (isSimpleSplitEnabled.value && newContent) {
    computeSimpleSplit()
  }
})
</script>

<style scoped>
.context-preview {
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
}
</style>

