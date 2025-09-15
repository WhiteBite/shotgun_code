<template>
  <div class="context-preview flex flex-col h-full panel-container">
    <!-- Header -->
    <div class="panel-header backdrop-blur-sm flex-shrink-0">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          <h3 class="text-lg font-semibold text-white">Контекст</h3>
        </div>
        
        <!-- Action buttons moved to header -->
        <div class="flex items-center gap-2">
          <button
            v-if="contextBuilderStore.currentContext?.content"
            @click="copyContext"
            class="p-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors flex items-center gap-1 text-sm"
            title="Копировать в буфер"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            <span class="hidden sm:inline">Copy</span>
          </button>
          <button
            v-if="contextBuilderStore.currentContext?.content"
            @click="saveContext"
            class="p-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors flex items-center gap-1 text-sm"
            title="Сохранить контекст"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5m-1 4l-3 3m0 0l-3-3m3 3V4" />
            </svg>
            <span class="hidden sm:inline">Save</span>
          </button>
        </div>
      </div>
      
      <!-- Status with better styling -->
      <div class="flex items-center justify-between text-xs mt-3 pt-3 border-t border-gray-600">
        <div class="flex items-center gap-4">
          <span v-if="contextBuilderStore.selectedFilesCount > 0" class="text-gray-400">
            {{ contextBuilderStore.selectedFilesCount }} файлов выбрано
          </span>
          <span v-if="contextBuilderStore.currentContext" class="flex items-center gap-2 text-green-400">
            <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
            Контекст построен
          </span>
        </div>
        <div v-if="contextBuilderStore.currentContext" class="text-gray-500">
          {{ Math.ceil((contextBuilderStore.currentContext.content?.length || 0) / 4) }} токенов
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 min-h-0 overflow-hidden bg-gray-900">
      <!-- Loading State -->
      <div v-if="contextBuilderStore.isBuilding" class="p-8 text-center text-gray-400 h-full flex items-center justify-center">
        <div class="flex flex-col items-center gap-4">
          <svg class="animate-spin h-8 w-8 text-purple-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          <div class="text-center">
            <p class="text-lg font-medium">Построение контекста...</p>
            <p class="text-sm text-gray-500 mt-1">Анализируем выбранные файлы</p>
          </div>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="contextBuilderStore.error" class="p-8 text-center text-red-400 h-full flex items-center justify-center">
        <div class="bg-red-900/20 border border-red-700 rounded-lg p-6 max-w-md">
          <svg class="w-8 h-8 mx-auto mb-3 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p class="text-sm mb-3">{{ contextBuilderStore.error }}</p>
          <button
            @click="clearError"
            class="px-4 py-2 bg-red-600 hover:bg-red-500 rounded-lg text-sm font-medium"
          >
            Очистить ошибку
          </button>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else-if="contextBuilderStore.selectedFilesCount === 0" class="p-8 text-center text-gray-400 h-full flex items-center justify-center">
        <div class="flex flex-col items-center gap-4 max-w-md">
          <svg class="w-16 h-16 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <div class="text-center">
            <p class="text-lg font-medium mb-2">Выберите файлы для анализа</p>
            <p class="text-sm text-gray-500">Отметьте файлы в дереве слева и нажмите "Собрать контекст"</p>
          </div>
        </div>
      </div>

      <!-- Success State -->
      <div v-else-if="contextBuilderStore.currentContext?.content" class="h-full flex flex-col">
        <ScrollableContent
          :content="contextBuilderStore.currentContext.content"
          :language="'markdown'"
          :highlight="true"
          class="flex-1 min-h-0 context-content"
        />
      </div>

      <!-- Waiting for context -->
      <div v-else class="p-8 text-center text-gray-400 h-full flex items-center justify-center">
        <div class="flex flex-col items-center gap-4">
          <svg class="w-12 h-12 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <div class="text-center">
            <p class="text-lg font-medium">Готов к работе</p>
            <p class="text-sm text-gray-500 mt-1">Нажмите Ctrl+Enter для построения контекста</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useUiStore } from '@/stores/ui.store'
import ScrollableContent from '@/presentation/components/shared/ScrollableContent.vue'
import { t } from '@/lib/i18n'

const contextBuilderStore = useContextBuilderStore()
const uiStore = useUiStore()

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
</script>

<style scoped>
.context-preview {
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
}
</style>