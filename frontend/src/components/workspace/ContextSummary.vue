<template>
  <div class="space-y-2">
    <!-- Stats Grid -->
    <div class="grid grid-cols-2 gap-1 text-xs">
      <div class="flex justify-between">
        <div class="text-gray-400">Файлы:</div>
        <div class="text-white">{{ contextBuilderStore.contextSummary.files }}</div>
      </div>
      <div class="flex justify-between">
        <div class="text-gray-400">Символы:</div>
        <div class="text-white">{{ contextBuilderStore.contextSummary.characters.toLocaleString() }}</div>
      </div>
      <div class="flex justify-between">
        <div class="text-gray-400">Токены:</div>
        <div class="text-white">~{{ contextBuilderStore.contextSummary.tokens.toLocaleString() }}</div>
      </div>
      <div class="flex justify-between">
        <div class="text-gray-400">Стоимость:</div>
        <div class="text-white">${{ contextBuilderStore.contextSummary.cost.toFixed(4) }}</div>
      </div>
    </div>
    
    <!-- Context Preview -->
    <div v-if="contextBuilderStore.currentContext?.content" class="border-t border-gray-700 pt-2">
      <div class="flex items-center justify-between mb-1">
        <h5 class="text-xs font-medium text-gray-300">Превью</h5>
        <button
          @click="copyContext"
          class="p-0.5 text-gray-400 hover:text-white transition-colors"
          title="Копировать"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </button>
      </div>
      <div class="bg-gray-900 rounded p-1.5 text-xs text-gray-300 max-h-24 overflow-y-auto">
        <pre class="whitespace-pre-wrap text-[10px] leading-tight">{{ getContextPreview() }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useContextBuilderStore } from '@/stores/context-builder.store'

const contextBuilderStore = useContextBuilderStore()

function getContextPreview(): string {
  const content = contextBuilderStore.currentContext?.content
  if (!content) return ''
  
  // Show first 200 characters with ellipsis for better readability
  if (content.length <= 200) {
    return content
  }
  
  // Try to find a good break point
  const preview = content.substring(0, 200)
  const lastNewline = preview.lastIndexOf('\n')
  const lastSpace = preview.lastIndexOf(' ')
  
  // Prefer breaking at newline, then space, then just cut
  const breakPoint = lastNewline > 150 ? lastNewline : (lastSpace > 150 ? lastSpace : 200)
  
  return content.substring(0, breakPoint) + '...'
}

function copyContext() {
  if (contextBuilderStore.currentContext?.content) {
    navigator.clipboard.writeText(contextBuilderStore.currentContext.content)
  }
}
</script>