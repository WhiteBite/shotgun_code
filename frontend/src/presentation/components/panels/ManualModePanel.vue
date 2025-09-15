<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useUiStore } from '@/stores/ui.store'
import { useContextActions } from '@/composables/useContextActions'
import { CodeBracketIcon } from '@heroicons/vue/24/outline'
import SplitSettingsPopover from '@/presentation/components/workspace/ContextActions/SplitSettingsPopover.vue'
import ContextViewer from '@/presentation/components/shared/ContextViewer.vue'
import StreamingContextViewer from '@/presentation/components/shared/StreamingContextViewer.vue'
import { MemoryErrorHandler } from '@/utils/memory-error-handler'
import RecoverableErrorBoundary from '@/presentation/components/shared/RecoverableErrorBoundary.vue'

const contextStore = useContextBuilderStore()
const fileTreeStore = useFileTreeStore()
const uiStore = useUiStore()
const { copy, splitSettings, splitPreview, refreshPreview } = useContextActions()

// Add error boundary reference
const errorBoundaryRef = ref<InstanceType<typeof RecoverableErrorBoundary> | null>(null)

const taskDescription = ref('')
const selectedModel = ref('gpt-4')
const isContextExpanded = ref(true)
const expandedFiles = ref(new Set<string>())
const error = ref<string | null>(null)

// Memory monitoring
const memoryWarning = ref(false)
const memoryStats = ref<{ used: number; total: number; percentage: number } | null>(null)

const canGenerate = computed(() => {
  return taskDescription.value.trim().length > 10 && contextStore.hasSelectedFiles
})

const truncatedContext = computed(() => {
  const context = contextStore.shotgunContextText
  if (!context) return ''
  
  // If collapsed, show very limited preview
  if (!isContextExpanded.value) {
    return context.length > 200 ? context.substring(0, 200) + '\n\n... (collapsed - click Expand All to see more)' : context
  }
  
  // If expanded, show more content but limit to prevent memory issues
  return context.length > 5000 ? context.substring(0, 5000) + '\n\n... (showing first 5,000 characters)' : context // Reduced from 10000 to 5000
})

// Enhanced tooltip content
const getTaskTooltipContent = computed(() => {
  const length = taskDescription.value.length
  return `
    <div class="space-y-2">
      <div class="font-medium">Task Description Tips</div>
      <div class="text-sm space-y-1">
        <div>• Be specific about what you want</div>
        <div>• Include technical details</div>
        <div>• Mention file names if relevant</div>
        <div>• Specify coding style preferences</div>
      </div>
      <div class="text-xs text-gray-400 border-t border-gray-600 pt-2">
        ${length < 20 ? 'Add more details for better results' : 
          length > 500 ? 'Consider breaking into smaller tasks' : 
          'Good length for clear instructions'}
      </div>
    </div>
  `
})

const getStatsTooltipContent = computed(() => {
  const chars = taskDescription.value.length
  const tokens = Math.ceil(chars / 4)
  const cost = (tokens * 0.00001).toFixed(5) // Rough GPT-4 pricing
  
  // Get memory stats if available
  const memStats = contextStore.getMemoryStats ? contextStore.getMemoryStats() : null
  const memoryInfo = memStats ? 
    `<div>Memory: ${memStats.used}MB / ${memStats.total}MB (${memStats.percentage}%)</div>` : 
    ''
  
  return `
    <div class="space-y-1">
      <div class="font-medium">Task Statistics</div>
      <div class="text-sm text-gray-300">
        <div>Characters: ${chars.toLocaleString()}</div>
        <div>Estimated tokens: ~${tokens.toLocaleString()}</div>
        <div>Estimated cost: ~$${cost}</div>
        ${memoryInfo}
      </div>
      <div class="text-xs text-gray-400 mt-2">
        Token count is approximate. Actual usage may vary.
      </div>
    </div>
  `
})

const getModelTooltipContent = computed(() => {
  const models = {
    'gpt-4': {
      name: 'GPT-4',
      description: 'Most capable model, best for complex tasks',
      cost: 'Higher cost',
      speed: 'Slower',
      icon: '🤖'
    },
    'claude-3': {
      name: 'Claude-3',
      description: 'Great for analysis and writing',
      cost: 'Medium cost',
      speed: 'Fast',
      icon: '💙'
    },
    'local': {
      name: 'Local Model',
      description: 'Privacy-focused, runs locally',
      cost: 'Free',
      speed: 'Variable',
      icon: '💻'
    }
  }
  
  const model = models[selectedModel.value as keyof typeof models] || models['gpt-4']
  return `
    <div class="space-y-2">
      <div class="font-medium flex items-center space-x-2">
        <span>${model.icon}</span>
        <span>${model.name}</span>
        </div>
      <div class="text-sm text-gray-300">${model.description}</div>
      <div class="text-xs text-gray-400 space-y-1">
        <div>Cost: ${model.cost}</div>
        <div>Speed: ${model.speed}</div>
      </div>
    </div>
  `
})

const getGenerateTooltipContent = computed(() => {
  if (!canGenerate.value) {
    if (taskDescription.value.trim().length <= 10) {
      return `
        <div class="space-y-1">
          <div class="font-medium text-red-400">Cannot Generate</div>
          <div class="text-sm text-gray-300">Task description too short</div>
          <div class="text-xs text-gray-400">Add more details about what you want</div>
        </div>
      `
    }
    if (!contextStore.hasSelectedFiles) {
      return `
        <div class="space-y-1">
          <div class="font-medium text-red-400">Cannot Generate</div>
          <div class="text-sm text-gray-300">No files selected</div>
          <div class="text-xs text-gray-400">Select files to provide context</div>
        </div>
      `
    }
  }
  
  return `
    <div class="space-y-1">
      <div class="font-medium text-green-400">Ready to Generate</div>
      <div class="text-sm text-gray-300">Send task to AI model</div>
      <div class="text-xs text-gray-400">
        <kbd class="bg-gray-700 px-1 rounded">Ctrl+Enter</kbd> to execute
      </div>
    </div>
  `
})

const getContextStatsTooltip = computed(() => {
  const files = contextStore.selectedFilesCount
  const tokens = contextStore.contextMetrics.tokenCount
  const chars = contextStore.contextMetrics.characterCount
  
  // Get memory stats
  const memStats = contextStore.getMemoryStats ? contextStore.getMemoryStats() : null
  const memoryInfo = memStats ? 
    `<div class="border-t border-gray-600 pt-1 mt-1">
       <div>Memory: ${memStats.used}MB / ${memStats.total}MB</div>
       <div>Usage: ${memStats.percentage}%</div>
       ${memStats.percentage > 80 ? '<div class="text-red-400">Warning: High memory usage!</div>' : ''}
     </div>` : 
    ''
  
  return `
    <div class="space-y-2">
      <div class="font-medium">Context Statistics</div>
      <div class="text-sm space-y-1">
        <div>Files: ${files}</div>
        <div>Characters: ${chars?.toLocaleString() || 'N/A'}</div>
        <div>Tokens: ${tokens.toLocaleString()}</div>
      </div>
      ${memoryInfo}
      <div class="text-xs text-gray-400 border-t border-gray-600 pt-1">
        This content will be sent to the AI as context
      </div>
    </div>
  `
})

// Methods
const generateCode = () => {
  console.log('Generate code with task:', taskDescription.value)
  // This will be implemented with actual AI generation
}

const showTemplates = () => {
  console.log('Show task templates')
}

const showExamples = () => {
  console.log('Show task examples')
}

const showHistory = () => {
  console.log('Show task history')
}

const collapseAll = () => {
  console.log('Collapse all context files')
  isContextExpanded.value = false
  expandedFiles.value.clear()
}

const expandAll = () => {
  console.log('Expand all context files')
  isContextExpanded.value = true
  // Add all files to expanded set
  contextStore.selectedFilesList.forEach(file => {
    expandedFiles.value.add(file)
  })
}

const copyContext = async () => {
  if (!contextStore.shotgunContextText) {
    uiStore.addToast('No context to copy', 'warning')
    return
  }
  
  try {
    await copy({ 
      target: 'all', 
      format: 'plain', 
      stripComments: false 
    })
    uiStore.addToast('Context copied to clipboard', 'success')
  } catch (error) {
    console.error('Failed to copy context:', error)
    uiStore.addToast('Failed to copy context', 'error')
  }
}

// New optimized context handlers
const handleContextCopy = () => {
  uiStore.addToast('Context copied to clipboard', 'success')
}

const handleMemoryWarning = (usage: number) => {
  memoryWarning.value = true
  uiStore.addToast(
    `High memory usage detected (${usage.toFixed(1)} MB). This might affect application performance.`,
    'warning'
  )
}

const handleContextError = (err: Error) => {
  // Use the memory error handler to handle the error
  error.value = err.message;
  MemoryErrorHandler.handle(err, {
    showNotification: true,
    context: {
      component: 'ManualModePanel',
      filesCount: contextStore.selectedFilesCount,
      contextSize: contextStore.contextMetrics.characterCount
    }
  })
}

// Function to clear selected files when there's an error
const clearSelectedFiles = () => {
  contextStore.clearSelectedFiles();
  error.value = null;
}

// Error handling and recovery
const handleErrorBoundaryError = (error: Error) => {
  console.error('Error boundary caught error:', error);
  uiStore.addToast('An error occurred while rendering the context. Error has been contained.', 'error');
}

const handleErrorBoundaryRetry = () => {
  // Retry loading the context, but with smaller limits
  if (contextStore.shotgunContextText && contextStore.shotgunContextText.length > 250000) { // Reduced from 500000 to 250000
    // If context is very large, notify user we're trimming it
    uiStore.addToast('Context is very large. Trimming for better performance...', 'info');
  }
  
  // Reset any error states
  uiStore.addToast('Retrying context rendering...', 'info');
}

const handleErrorBoundaryRecovery = () => {
  // Free memory by clearing large objects
  if (contextStore.clearLargeObjects) {
    contextStore.clearLargeObjects();
  }
  
  // Restart memory monitoring with delay
  setTimeout(() => {
    contextStore.startMemoryMonitoring();
    uiStore.addToast('Memory monitoring restarted', 'success');
  }, 1000);
  
  uiStore.addToast('Recovery attempted. Memory has been cleared.', 'success');
}

const handleErrorBoundaryReset = () => {
  // Reset to a clean state
  if (contextStore.clearLargeObjects) {
    contextStore.clearLargeObjects();
  }
  uiStore.addToast('Context view reset', 'info');
}

// Memory monitoring
const checkMemoryUsage = () => {
  if (contextStore.getMemoryStats) {
    const stats = contextStore.getMemoryStats();
    if (stats) {
      memoryStats.value = stats;
      if (stats.percentage > 70) { // Reduced from 80
        memoryWarning.value = true;
        if (stats.percentage > 85) { // Reduced from 90
          uiStore.addToast(
            `Critical memory usage: ${stats.percentage}%. Consider closing other applications or reducing context size.`,
            'error'
          );
        } else {
          uiStore.addToast(
            `High memory usage: ${stats.percentage}%. Consider reducing context size.`,
            'warning'
          );
        }
      }
    }
  }
}

// Install global error handler on component mount
onMounted(() => {
  MemoryErrorHandler.installGlobalHandler()
  
  // Monitor for memory issues
  window.addEventListener('error', (event) => {
    if (event.error && MemoryErrorHandler.categorizeError(event.error) === 'memory') {
      handleContextError(event.error)
    }
  })
  
  // Start memory monitoring
  const memoryInterval = setInterval(checkMemoryUsage, 5000);
  
  // Clean up on unmount
  onUnmounted(() => {
    clearInterval(memoryInterval);
  });
})

onUnmounted(() => {
  // Clean up memory monitoring when component is destroyed
  if (contextStore.stopMemoryMonitoring) {
    contextStore.stopMemoryMonitoring()
  }
})

const toggleContextExpansion = () => {
  isContextExpanded.value = !isContextExpanded.value
}

const buildContext = () => {
  // This would trigger context building
  if (contextStore.hasSelectedFiles) {
    uiStore.addToast('Building context from selected files...', 'info')
  } else {
    uiStore.addToast('No files selected to build context', 'warning')
  }
}

// Determine if we should use streaming context viewer
const shouldUseStreaming = computed(() => {
  // Use streaming for even small files to avoid memory issues
  const contextSize = contextStore.shotgunContextText?.length || 0
  
  // Safely calculate selected files size with error handling
  let selectedFilesSize = 0
  try {
    selectedFilesSize = contextStore.selectedFilesList.reduce((total, filePath) => {
      // Check if fileTreeStore is available and has the method
      if (fileTreeStore && typeof fileTreeStore.getFileByRelPath === 'function') {
        const node = fileTreeStore.getFileByRelPath(filePath)
        return total + (node?.size || 0)
      }
      return total
    }, 0)
  } catch (err) {
    console.warn('Error calculating selected files size:', err)
    // Fallback to a conservative approach
    selectedFilesSize = contextStore.selectedFilesList.length * 1024 // Assume 1KB per file
  }
  
  // Very conservative threshold - even files as small as 100KB could cause issues
  return contextSize > 100 * 1024 || selectedFilesSize > 100 * 1024 || contextStore.selectedFilesList.length > 1
})
</script>

<template>
  <div class="manual-mode-panel h-full flex flex-col bg-gray-850">
    <!-- Error Boundary -->
    <RecoverableErrorBoundary
      ref="errorBoundaryRef"
      @error="handleErrorBoundaryError"
      @retry="handleErrorBoundaryRetry"
      @recover="handleErrorBoundaryRecovery"
      @reset="handleErrorBoundaryReset"
      :max-retries="3"
    >
      <template #default>
        <div class="flex-1 flex flex-col overflow-hidden">
          <!-- Task Input Section -->
          <div class="p-4 border-b border-gray-700 bg-gray-800">
            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-300 mb-2">
                Task Description
              </label>
              <textarea
                v-model="taskDescription"
                placeholder="Describe what you want to accomplish..."
                class="w-full p-3 bg-gray-750 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                :class="{ 'ring-2 ring-red-500': taskDescription.length > 0 && taskDescription.length < 10 }"
                rows="4"
              ></textarea>
              <div class="flex justify-between items-center mt-2">
                <div class="text-xs text-gray-400">
                  {{ taskDescription.length }} characters
                </div>
                <div class="flex space-x-2">
                  <button
                    @click="showTemplates"
                    class="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
                    title="Task Templates"
                  >
                    Templates
                  </button>
                  <button
                    @click="showExamples"
                    class="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
                    title="Task Examples"
                  >
                    Examples
                  </button>
                  <button
                    @click="showHistory"
                    class="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
                    title="Task History"
                  >
                    History
                  </button>
                </div>
              </div>
            </div>

            <div class="flex flex-wrap gap-3 items-center">
              <div class="flex-1 min-w-[200px]">
                <label class="block text-sm font-medium text-gray-300 mb-2">
                  AI Model
                </label>
                <select
                  v-model="selectedModel"
                  class="w-full p-2 bg-gray-750 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="gpt-4">GPT-4</option>
                  <option value="claude-3">Claude-3</option>
                  <option value="local">Local Model</option>
                </select>
              </div>

              <div class="flex items-end space-x-2">
                <button
                  @click="generateCode"
                  :disabled="!canGenerate"
                  class="px-4 py-2 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-500 hover:to-blue-500 disabled:from-gray-600 disabled:to-gray-600 disabled:cursor-not-allowed text-white font-medium rounded-lg transition-all duration-200 flex items-center gap-2 shadow-lg hover:shadow-xl transform hover:scale-105 disabled:transform-none"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  Generate Code
                  <kbd class="ml-auto px-1.5 py-0.5 text-xs bg-white/20 rounded">Ctrl+Enter</kbd>
                </button>
              </div>
            </div>
          </div>

          <!-- Context Section -->
          <div class="flex-1 flex flex-col overflow-hidden">
            <div class="p-4 border-b border-gray-700 bg-gray-800">
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-sm font-bold text-white flex items-center gap-2">
                  <CodeBracketIcon class="w-4 h-4 text-blue-400" />
                  Context Preview
                </h3>
                <div class="flex items-center space-x-2">
                  <button
                    @click="collapseAll"
                    class="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
                  >
                    Collapse All
                  </button>
                  <button
                    @click="expandAll"
                    class="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
                  >
                    Expand All
                  </button>
                  <button
                    @click="copyContext"
                    class="px-2 py-1 text-xs bg-blue-600 hover:bg-blue-500 text-white rounded transition-colors"
                  >
                    Copy Context
                  </button>
                </div>
              </div>
              <div class="text-xs text-gray-400 mb-2">
                Files selected for context: {{ contextStore.selectedFilesCount }}
              </div>
              <div class="w-full bg-gray-700 rounded-full h-1.5">
                <div
                  class="bg-gradient-to-r from-blue-500 to-purple-500 h-1.5 rounded-full transition-all duration-300"
                  :style="{ width: Math.min((contextStore.selectedFilesCount / 10) * 100, 100) + '%' }"
                ></div>
              </div>
            </div>

            <!-- Context Viewer -->
            <div class="flex-1 overflow-hidden bg-gray-850">
              <!-- Error message when file selection causes problems -->
              <div v-if="error" class="p-4 bg-red-900/30 text-red-300 rounded m-2">
                <div class="font-medium">Error loading context</div>
                <div class="text-sm mt-1">{{ error }}</div>
                <button 
                  @click="error = null; clearSelectedFiles()" 
                  class="mt-2 px-2 py-1 bg-red-700 hover:bg-red-600 text-white text-xs rounded"
                >
                  Clear Selection
                </button>
              </div>
            
              <!-- Use streaming context viewer for large contexts -->
              <StreamingContextViewer
                v-else-if="shouldUseStreaming && contextStore.streamingContext?.id"
                :content-id="contextStore.streamingContext.id"
                :container-height="300"
                @copy="handleContextCopy"
                @error="handleContextError"
              />
              <!-- Use regular context viewer for smaller contexts -->
              <ContextViewer
                v-else
                :content="contextStore.shotgunContextText"
                :is-expanded="isContextExpanded"
                :container-height="300"
                @copy="handleContextCopy"
                @memory-warning="handleMemoryWarning"
                @toggle-expansion="toggleContextExpansion"
              />
            </div>
          </div>
        </div>
      </template>
    </RecoverableErrorBoundary>
  </div>
</template>