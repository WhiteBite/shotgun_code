<template>
  <div class="h-full flex flex-col bg-gray-800/50 backdrop-blur-sm p-4">
    <div class="flex items-center gap-2 mb-3">
      <div class="w-1 h-5 bg-blue-500 rounded-full"></div>
      <h2 class="text-lg font-semibold text-white">Task Description</h2>
    </div>
    
    <textarea
      v-model="taskDescription"
      placeholder="Describe what you want to accomplish..."
      class="task-textarea flex-1 w-full px-4 py-3 bg-gray-900/80 border border-gray-700/50 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-gray-600 focus:bg-gray-900 resize-none transition-all duration-200"
    ></textarea>

    <div class="mt-4 flex gap-2">
      <button
        @click="analyzeTask"
        :disabled="!taskDescription.trim() || isAnalyzing"
        class="px-4 py-2 bg-blue-600/90 hover:bg-blue-600 disabled:bg-gray-700 disabled:cursor-not-allowed text-white rounded-lg transition-all duration-200 flex items-center gap-2 shadow-lg hover:shadow-blue-500/20"
      >
        <svg v-if="isAnalyzing" class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span>Analyze Task</span>
      </button>
      
      <button
        @click="buildContext"
        :disabled="!taskDescription.trim()"
        class="px-4 py-2 bg-gray-700/80 hover:bg-gray-700 disabled:bg-gray-800 disabled:cursor-not-allowed text-white rounded-lg transition-all duration-200"
      >
        Build Context
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useUIStore } from '@/stores/ui.store'

const taskDescription = ref('')
const isAnalyzing = ref(false)
const uiStore = useUIStore()

async function analyzeTask() {
  if (!taskDescription.value.trim()) return
  
  isAnalyzing.value = true
  try {
    // TODO: Call backend API to analyze task
    await new Promise(resolve => setTimeout(resolve, 1000))
    uiStore.addToast('Task analyzed successfully', 'success')
  } catch (error) {
    uiStore.addToast('Failed to analyze task', 'error')
  } finally {
    isAnalyzing.value = false
  }
}

async function buildContext() {
  if (!taskDescription.value.trim()) return
  
  try {
    // TODO: Call backend API to build context
    uiStore.addToast('Building context...', 'info')
  } catch (error) {
    uiStore.addToast('Failed to build context', 'error')
  }
}
</script>

<style scoped>
.task-textarea {
  font-family: inherit;
  line-height: 1.6;
}

.task-textarea::placeholder {
  font-style: italic;
}

/* Плавная анимация при фокусе */
.task-textarea:focus {
  box-shadow: 0 0 0 3px rgba(107, 114, 128, 0.1);
}

/* Улучшенный scrollbar для textarea */
.task-textarea::-webkit-scrollbar {
  width: 6px;
}

.task-textarea::-webkit-scrollbar-track {
  background: transparent;
}

.task-textarea::-webkit-scrollbar-thumb {
  background: rgba(107, 114, 128, 0.3);
  border-radius: 3px;
}

.task-textarea::-webkit-scrollbar-thumb:hover {
  background: rgba(107, 114, 128, 0.5);
}
</style>
