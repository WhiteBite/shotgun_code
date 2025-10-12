<template>
  <div class="h-full flex flex-col bg-gray-800/50 backdrop-blur-sm p-4">
    <!-- Header with character counter -->
    <div class="flex items-center justify-between mb-3">
      <div class="flex items-center gap-2">
        <div class="w-1 h-5 bg-blue-500 rounded-full"></div>
        <h2 class="text-lg font-semibold text-white">Task Description</h2>
      </div>
      <span class="text-xs text-gray-400" :class="{ 'text-yellow-400': taskDescription.length > 4500, 'text-red-400': taskDescription.length >= 5000 }">
        {{ taskDescription.length }} / 5000
      </span>
    </div>
    
    <!-- Textarea with auto-resize -->
    <textarea
      ref="textareaRef"
      v-model="taskDescription"
      placeholder="Describe what you want to accomplish...&#10;&#10;Examples:&#10;• Refactor authentication service&#10;• Add dark mode support&#10;• Implement file upload feature"
      class="task-textarea flex-1 w-full px-4 py-3 bg-gray-900/80 border border-gray-700/50 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500/50 focus:bg-gray-900 resize-none transition-all duration-200"
      maxlength="5000"
      @input="autoResize"
    ></textarea>

    <!-- AI Suggestions -->
    <div v-if="suggestions.length > 0 && !taskDescription" class="mt-3 space-y-2">
      <p class="text-xs text-gray-400 font-medium">Quick suggestions:</p>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="suggestion in suggestions"
          :key="suggestion"
          @click="applySuggestion(suggestion)"
          class="px-3 py-1 text-xs bg-gray-700/60 hover:bg-gray-700 text-gray-200 rounded-full transition-colors"
        >
          {{ suggestion }}
        </button>
      </div>
    </div>

    <!-- Action buttons -->
    <div class="mt-4 flex gap-2">
      <button
        @click="analyzeTask"
        :disabled="!canAnalyze"
        class="flex-1 px-4 py-2 bg-blue-600/90 hover:bg-blue-600 disabled:bg-gray-700 disabled:cursor-not-allowed text-white rounded-lg transition-all duration-200 flex items-center justify-center gap-2 shadow-lg hover:shadow-blue-500/20"
      >
        <svg v-if="!isAnalyzing" class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
        </svg>
        <svg v-else class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span>{{ isAnalyzing ? 'Analyzing...' : 'Analyze Task' }}</span>
      </button>
      
      <button
        title="Build context (Ctrl+B)"
        @click="buildContext"
        :disabled="!taskDescription.trim()"
        class="px-4 py-2 bg-gray-700/80 hover:bg-gray-700 disabled:bg-gray-800 disabled:cursor-not-allowed text-white rounded-lg transition-all duration-200"
      >
        <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
      </button>
    </div>
    
    <!-- Progress indicator -->
    <div v-if="analysisProgress" class="mt-3 space-y-2">
      <div class="flex items-center justify-between text-xs">
        <span class="text-gray-400">{{ analysisProgress.stage }}</span>
        <span class="text-blue-400">{{ analysisProgress.percent }}%</span>
      </div>
      <div class="h-1 bg-gray-700 rounded-full overflow-hidden">
        <div 
          class="h-full bg-blue-500 transition-all duration-300"
          :style="{ width: `${analysisProgress.percent}%` }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useUIStore } from '@/stores/ui.store'
import { useFileStore } from '@/stores/file.store'
import { useContextStore } from '@/stores/context.store'

const textareaRef = ref<HTMLTextAreaElement>()
const taskDescription = ref('')
const isAnalyzing = ref(false)
const analysisProgress = ref<{ stage: string; percent: number } | null>(null)
const uiStore = useUIStore()

const suggestions = ref([
  'Add user authentication',
  'Implement data caching',
  'Refactor API calls',
  'Add unit tests'
])

const canAnalyze = computed(() => {
  return taskDescription.value.trim().length > 10 && !isAnalyzing.value
})

// Auto-resize textarea
function autoResize() {
  if (!textareaRef.value) return
  textareaRef.value.style.height = 'auto'
  textareaRef.value.style.height = textareaRef.value.scrollHeight + 'px'
}

function applySuggestion(suggestion: string) {
  taskDescription.value = suggestion
  autoResize()
}

// Save draft to localStorage with debounce
let saveTimeout: number | null = null
watch(taskDescription, (value) => {
  if (saveTimeout) clearTimeout(saveTimeout)
  saveTimeout = setTimeout(() => {
    try {
      localStorage.setItem('task-draft', value)
    } catch (err) {
      console.warn('Failed to save draft:', err)
    }
  }, 500) as unknown as number
})

// Restore draft on mount
onMounted(() => {
  try {
    const draft = localStorage.getItem('task-draft')
    if (draft) {
      taskDescription.value = draft
      autoResize()
    }
  } catch (err) {
    console.warn('Failed to load draft:', err)
  }
})

async function analyzeTask() {
  if (!canAnalyze.value) return
  
  isAnalyzing.value = true
  analysisProgress.value = { stage: 'Analyzing...', percent: 0 }
  
  try {
    // Simulate progress
    await new Promise(resolve => setTimeout(resolve, 300))
    analysisProgress.value = { stage: 'Processing AST...', percent: 30 }
    
    await new Promise(resolve => setTimeout(resolve, 400))
    analysisProgress.value = { stage: 'Finding dependencies...', percent: 60 }
    
    await new Promise(resolve => setTimeout(resolve, 300))
    analysisProgress.value = { stage: 'Generating suggestions...', percent: 90 }
    
    // TODO: Call backend API to analyze task
    await new Promise(resolve => setTimeout(resolve, 200))
    analysisProgress.value = { stage: 'Complete!', percent: 100 }
    
    uiStore.addToast('Task analyzed successfully', 'success')
  } catch (error) {
    uiStore.addToast('Failed to analyze task', 'error')
  } finally {
    setTimeout(() => {
      isAnalyzing.value = false
      analysisProgress.value = null
    }, 500)
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
