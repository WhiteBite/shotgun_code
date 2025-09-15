<template>
  <div class="h-full flex flex-col">
    <!-- Results Header -->
    <div class="p-4 border-b border-gray-700">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-gray-100">Results</h2>
        <div class="flex items-center space-x-2">
          <button 
            class="text-xs px-2 py-1 bg-gray-700 rounded hover:bg-gray-600 transition-colors"
            @click="clearResults"
            title="Clear all results"
          >
            Clear
          </button>
          <button 
            class="text-xs px-2 py-1 bg-gray-700 rounded hover:bg-gray-600 transition-colors"
            @click="exportResults"
            :disabled="!hasResults"
            :class="{ 'opacity-50 cursor-not-allowed': !hasResults }"
            title="Export results to file"
          >
            Export
          </button>
        </div>
      </div>
    </div>
    
    <!-- Tabbed Interface -->
    <div class="border-b border-gray-700">
      <nav class="flex space-x-1 p-2">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="px-3 py-2 text-sm rounded-lg transition-colors"
          :class="[
            activeTab === tab.id 
              ? 'bg-gray-700 text-white' 
              : 'text-gray-400 hover:text-white hover:bg-gray-800'
          ]"
          @click="activeTab = tab.id"
        >
          {{ tab.name }}
          <span 
            v-if="tab.count > 0" 
            class="ml-2 px-1.5 py-0.5 text-xs bg-gray-600 rounded-full"
          >
            {{ tab.count }}
          </span>
        </button>
      </nav>
    </div>
    
    <!-- Tab Content -->
    <div class="flex-1 flex flex-col min-h-0">
      <!-- Output Preview Tab -->
      <div v-if="activeTab === 'output'" class="flex-1 flex flex-col">
        <div class="p-4 border-b border-gray-700">
          <div class="flex items-center justify-between text-sm">
            <div class="flex items-center space-x-4 text-gray-400">
              <span>Last generated: {{ lastGenerated || 'Never' }}</span>
              <span>•</span>
              <span>{{ outputLines }} lines</span>
            </div>
            <div class="flex items-center space-x-2">
              <button class="px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors">
                Apply Changes
              </button>
              <button class="px-3 py-1 bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors">
                Refine
              </button>
            </div>
          </div>
        </div>
        
        <div class="flex-1 overflow-auto p-4">
          <div v-if="generatedOutput" class="bg-gray-800 rounded-lg p-4">
            <pre class="text-sm text-gray-300 whitespace-pre-wrap font-mono">{{ generatedOutput }}</pre>
          </div>
          <div v-else class="flex items-center justify-center h-full text-gray-500">
            <div class="text-center">
              <p class="text-lg mb-2">✨</p>
              <p>No output generated yet</p>
              <p class="text-xs mt-1">Generate code to see results here</p>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Diff Viewer Tab -->
      <div v-else-if="activeTab === 'diff'" class="flex-1 flex flex-col">
        <div class="p-4 border-b border-gray-700">
          <div class="flex items-center space-x-4 text-sm text-gray-400">
            <span>{{ diffStats.additions }} additions</span>
            <span>{{ diffStats.deletions }} deletions</span>
            <span>{{ diffStats.files }} files changed</span>
          </div>
        </div>
        
        <div class="flex-1 overflow-auto p-4">
          <div class="text-center text-gray-500">
            <p class="text-lg mb-2">📄</p>
            <p>Diff Viewer</p>
            <p class="text-xs mt-1">Coming in Phase 5</p>
          </div>
        </div>
      </div>
      
      <!-- Console Tab -->
      <div v-else-if="activeTab === 'console'" class="flex-1 flex flex-col">
        <div class="p-4 border-b border-gray-700">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-4">
              <div class="flex items-center space-x-2">
                <div class="w-2 h-2 bg-green-400 rounded-full" />
                <span class="text-sm text-gray-400">Connected</span>
              </div>
            </div>
            <div class="flex items-center space-x-2">
              <button 
                class="text-xs px-2 py-1 bg-gray-700 rounded hover:bg-gray-600 transition-colors"
                @click="clearConsole"
                title="Clear console logs"
              >
                Clear
              </button>
              <button 
                class="text-xs px-2 py-1 bg-gray-700 rounded hover:bg-gray-600 transition-colors"
                @click="exportConsole"
                :disabled="consoleLogs.length === 0"
                :class="{ 'opacity-50 cursor-not-allowed': consoleLogs.length === 0 }"
                title="Export console logs"
              >
                Export
              </button>
            </div>
          </div>
        </div>
        
        <div class="flex-1 overflow-auto p-4 bg-black bg-opacity-30">
          <div class="font-mono text-sm space-y-1">
            <div v-for="log in consoleLogs" :key="log.id" class="flex">
              <span class="text-gray-500 mr-2">{{ log.timestamp }}</span>
              <span :class="getLogColor(log.level)">{{ log.message }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useExportStore } from '@/stores/export.store'
import { useUiStore } from '@/stores/ui.store'

const exportStore = useExportStore()
const uiStore = useUiStore()

const activeTab = ref('output')

const tabs = [
  { id: 'output', name: 'Output', count: 0 },
  { id: 'diff', name: 'Diff', count: 0 },
  { id: 'console', name: 'Console', count: 3 }
]

// Mock data
const generatedOutput = ref('')
const lastGenerated = ref('')

const outputLines = computed(() => {
  return generatedOutput.value ? generatedOutput.value.split('\n').length : 0
})

const hasResults = computed(() => {
  return generatedOutput.value.trim().length > 0
})

// Methods
const clearResults = () => {
  generatedOutput.value = ''
  lastGenerated.value = ''
  diffStats.value = { additions: 0, deletions: 0, files: 0 }
  uiStore.addToast('Results cleared', 'info')
}

const exportResults = () => {
  if (!hasResults.value) {
    uiStore.addToast('No results to export', 'warning')
    return
  }
  
  // Open export modal
  exportStore.open()
  uiStore.addToast('Export dialog opened', 'info')
}

const clearConsole = () => {
  consoleLogs.value = []
  uiStore.addToast('Console cleared', 'info')
}

const exportConsole = () => {
  if (consoleLogs.value.length === 0) {
    uiStore.addToast('No console logs to export', 'warning')
    return
  }
  
  // Create log text
  const logText = consoleLogs.value
    .map(log => `[${log.timestamp}] ${log.level.toUpperCase()}: ${log.message}`)
    .join('\n')
  
  // Copy to clipboard
  navigator.clipboard.writeText(logText).then(() => {
    uiStore.addToast('Console logs copied to clipboard', 'success')
  }).catch(() => {
    uiStore.addToast('Failed to copy console logs', 'error')
  })
}

const diffStats = ref({
  additions: 0,
  deletions: 0,
  files: 0
})

const consoleLogs = ref([
  {
    id: 1,
    timestamp: '14:32:15',
    level: 'info',
    message: 'Context generation started'
  },
  {
    id: 2,
    timestamp: '14:32:18',
    level: 'success',
    message: 'Context built successfully (1,247 tokens)'
  },
  {
    id: 3,
    timestamp: '14:32:20',
    level: 'info',
    message: 'Waiting for task input...'
  }
])

const getLogColor = (level: string) => {
  switch (level) {
    case 'error': return 'text-red-400'
    case 'warning': return 'text-yellow-400'
    case 'success': return 'text-green-400'
    case 'info': return 'text-blue-400'
    default: return 'text-gray-300'
  }
}
</script>