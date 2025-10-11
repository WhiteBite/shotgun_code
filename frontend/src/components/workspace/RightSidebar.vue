<template>
  <div class="h-full flex flex-col border-l border-gray-700 bg-gray-900 p-4">
    <div class="text-gray-500 text-sm">
      <p class="mb-2 font-semibold">Инструменты</p>
      <p class="text-xs mb-4">Дополнительные функции и настройки</p>
      
      <!-- Quick Actions -->
      <div class="space-y-2">
        <button
          @click="$emit('open-export')"
          class="w-full px-3 py-2 bg-gray-800 hover:bg-gray-700 text-white text-xs rounded transition-colors text-left flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
          </svg>
          Экспорт контекста
        </button>
        
        <button
          class="w-full px-3 py-2 bg-gray-800 hover:bg-gray-700 text-white text-xs rounded transition-colors text-left flex items-center gap-2 opacity-50 cursor-not-allowed"
          disabled
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
          </svg>
          Настройки (скоро)
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineEmits<{
  (e: 'open-export'): void
}>()
</script>
