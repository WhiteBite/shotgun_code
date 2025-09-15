<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useGenerationStore } from '@/stores/generation.store'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useUiStore } from '@/stores/ui.store'
import { copy } from '@/lib/clipboard'
import { MemoryErrorHandler } from '@/lib/memory/MemoryErrorHandler'
import RecoverableErrorBoundary from '@/presentation/components/shared/RecoverableErrorBoundary.vue'
import ContextViewer from '@/presentation/components/shared/ContextViewer.vue'
import StreamingContextViewer from '@/presentation/components/shared/StreamingContextViewer.vue'
import type { FileNode } from '@/types/dto'

// Icons
import { 
  CodeBracketIcon, 
  DocumentTextIcon, 
  LightBulbIcon, 
  ClockIcon,
  CpuChipIcon,
  DocumentIcon
} from '@heroicons/vue/24/outline'

// Stores
const generationStore = useGenerationStore()
const contextStore = useContextBuilderStore()
const fileTreeStore = useFileTreeStore()
const uiStore = useUiStore()

// Refs
const taskDescription = ref('')
const selectedModel = ref('gpt-4')
const isContextExpanded = ref(false)
const memoryWarning = ref(false)
const memoryStats = ref<{ used: number; total: number; percentage: number } | null>(null)
const error = ref<string | null>(null)
const errorBoundaryRef = ref<InstanceType<typeof RecoverableErrorBoundary> | null>(null)

// Computed
const canGenerate = computed(() => {
  return taskDescription.value.trim().length > 10 && contextStore.hasSelectedFiles
})

const truncatedContext = computed(() => {
  // Use the new context summary approach instead of full text
  if (!contextStore.contextSummaryState) return ''
  
  // If collapsed, show very limited preview
  if (!isContextExpanded.value) {
    return `Context with ${contextStore.contextSummaryState.fileCount} files, ${contextStore.contextSummaryState.tokenCount} tokens`
  }
  
  // If expanded, show summary information
  return `Context Summary:
- Files: ${contextStore.contextSummaryState.fileCount}
- Tokens: ${contextStore.contextSummaryState.tokenCount}
- Size: ${contextStore.contextSummaryState.totalSize} characters
- Status: ${contextStore.contextSummaryState.status}`
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
  // Show a user feedback for placeholder functionality
  uiStore.addToast('Code generation feature coming soon!', 'info')
  // This will be implemented with actual AI generation
}

const showTemplates = () => {
  console.log('Show task templates')
  // Show templates functionality
  uiStore.addToast('Task templates feature coming soon!', 'info')
}

const showExamples = () => {
  console.log('Show task examples')
  // Show examples functionality
  uiStore.addToast('Task examples feature coming soon!', 'info')
}

const showHistory = () => {
  console.log('Show task history')
  // Show history functionality
  uiStore.addToast('Task history feature coming soon!', 'info')
}

const collapseAll = () => {
  console.log('Collapse all context files')
  isContextExpanded.value = false
}

const expandAll = () => {
  console.log('Expand all context files')
  isContextExpanded.value = true
}

const copyContext = async () => {
  // Use the new context summary approach instead of full text
  if (!contextStore.contextSummaryState) {
    uiStore.addToast('No context to copy', 'warning')
    return
  }
  
  try {
    // Copy context summary information instead of full content
    const summaryText = `Context Summary:
- Files: ${contextStore.contextSummaryState.fileCount}
- Tokens: ${contextStore.contextSummaryState.tokenCount}
- Size: ${contextStore.contextSummaryState.totalSize} characters
- Status: ${contextStore.contextSummaryState.status}
- Created: ${contextStore.contextSummaryState.createdAt}
- Updated: ${contextStore.contextSummaryState.updatedAt}`
    
    await navigator.clipboard.writeText(summaryText)
    uiStore.addToast('Context summary copied to clipboard', 'success')
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
  if (contextStore.contextSummaryState) {
    uiStore.addToast('Context summary loaded successfully', 'info');
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
  const hasContextSummary = !!contextStore.contextSummaryState
  
  // Safety check: ensure selectedFiles exists and is an array
  const selectedFiles = contextStore.selectedFilesList.value || []
  
  // Safely calculate selected files size with error handling
  let selectedFilesSize = 0
  try {
    selectedFilesSize = selectedFiles.reduce((total, filePath) => {
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
    selectedFilesSize = selectedFiles.length * 1024 // Assume 1KB per file
  }
  
  // Very conservative threshold - even files as small as 100KB could cause issues
  return hasContextSummary || selectedFilesSize > 100 * 1024 || selectedFiles.length > 1
})
</script>