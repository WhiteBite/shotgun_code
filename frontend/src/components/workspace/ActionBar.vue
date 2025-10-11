<template>
  <div class="h-16 border-t border-gray-700 bg-gray-850 flex items-center justify-between px-4 flex-shrink-0">
    <!-- Left: Quick Stats -->
    <div class="flex items-center space-x-4 text-xs text-gray-400">
      <div>
        Модель: <span class="text-white font-semibold">gpt-4o</span>
      </div>
      <div>
        Токены: <span class="text-white font-semibold">~{{ formatNumber(contextStore.tokenCount) }}</span>
      </div>
      <div>
        Стоимость: <span class="text-green-400 font-semibold">${{ contextStore.estimatedCost.toFixed(4) }}</span>
      </div>
    </div>

    <!-- Right: Action Buttons -->
    <div class="flex items-center space-x-3">
      <button
        @click="$emit('build-context')"
        :disabled="fileStore.selectedPaths.size === 0"
        class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Собрать контекст
      </button>
      
      <button
        @click="$emit('open-export')"
        :disabled="!contextStore.hasContext"
        class="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white text-sm rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
        </svg>
        Экспорт
      </button>
      
      <button
        @click="$emit('generate-solution')"
        :disabled="!contextStore.hasContext"
        class="px-6 py-2 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-500 hover:to-blue-500 text-white text-sm font-semibold rounded transition-all shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Сгенерировать решение →
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useFileStore } from '@/stores/file.store'
import { useContextStore } from '@/stores/context.store'

const fileStore = useFileStore()
const contextStore = useContextStore()

defineEmits<{
  (e: 'build-context'): void
  (e: 'open-export'): void
  (e: 'generate-solution'): void
}>()

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
  return num.toString()
}
</script>
