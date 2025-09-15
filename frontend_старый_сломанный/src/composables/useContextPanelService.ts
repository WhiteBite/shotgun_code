import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useProjectStore } from '@/stores/project.store'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useSettingsStore } from '@/stores/settings.store'
import usePanelManager from '@/composables/usePanelManager'
import { useContextChunking, type ContextChunk } from '@/infrastructure/context/context-chunking.service'
import { ContextPanelService } from '@/domain/services/ContextPanelService'

export function useContextPanelService() {
  // Get stores
  const contextBuilderStore = useContextBuilderStore()
  const projectStore = useProjectStore()
  const fileTreeStore = useFileTreeStore()
  const settingsStore = useSettingsStore()
  
  // Get reactive refs from stores
  const { 
    contextLoading,
    isBuilding,
    buildStatus,
    contextSummaryState,
    estimatedTokens,
    error,
    validationErrors
  } = storeToRefs(contextBuilderStore)
  
  const { selectedFiles } = storeToRefs(fileTreeStore)
  
  // Composables
  const { isCollapsed, toggleCollapse } = usePanelManager('context', 320)
  const { chunkContent, copyChunk, copyAll } = useContextChunking()
  
  // Create service instance
  const contextPanelService = new ContextPanelService()
  
  // Computed properties that depend on store state
  const selectedFilesCount = computed(() => selectedFiles.value.length)
  
  const contextData = computed(() => ({
    content: contextPanelService.paginatedContextContent,
    fileCount: contextSummaryState.value?.fileCount || 0,
    metadata: contextSummaryState.value?.metadata
  }))
  
  const hasContextContent = computed(() => !!contextSummaryState.value?.id)
  
  const contextError = computed(() => {
    if (error.value) {
      return {
        message: error.value,
        code: 'CONTEXT_BUILD_ERROR'
      }
    }
    return null
  })
  
  const contextChunks = computed(() => {
    if (!contextPanelService.enableSplit || !contextPanelService.paginatedContextContent) {
      return []
    }
    
    // Instead of modifying chunkedContext.value directly in computed, return the result
    if (contextPanelService.chunkedContext.length === 0) {
      // Auto-chunk the content when split is enabled
      const newChunks = chunkContent(contextPanelService.paginatedContextContent, {
        maxTokens: settingsStore.chunkSize || 1000,
        strategy: 'balanced',
        preserveCodeBlocks: true,
        preserveMarkdownStructure: true
      })
      return newChunks
    }
    
    return contextPanelService.chunkedContext
  })
  
  const activeChunk = computed(() => {
    return contextChunks.value[contextPanelService.activeChunkIndex] || null
  })
  
  const splitSettings = computed(() => ({
    enabled: contextPanelService.enableSplit,
    layout: settingsStore.splitLayout || 'vertical',
    chunkSize: settingsStore.chunkSize || 1000,
    chunks: contextChunks.value,
    activeChunkIndex: contextPanelService.activeChunkIndex
  }))
  
  // Text field for selected files
  const selectedFilesText = computed({
    get: () => selectedFiles.value.join('\n'),
    set: (value: string) => {
      const newFiles = value
        .split('\n')
        .map(v => v.trim().replace(/\\/g, '/'))
        .filter(Boolean)
      fileTreeStore.setSelectedFiles(newFiles)
    }
  })
  
  // Methods
  function handleToggle(collapsed: boolean) {
    toggleCollapse()
  }
  
  function handleResize(width: number) {
    // Handle panel resize if needed
    console.log('Panel resized to:', width)
  }
  
  function handleRetry() {
    if (projectStore.currentProject?.path) {
      contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
    }
  }
  
  function handleCopy() {
    if (contextPanelService.paginatedContextContent) {
      navigator.clipboard.writeText(contextPanelService.paginatedContextContent)
    }
  }
  
  function handleBuild() {
    if (projectStore.currentProject?.path) {
      contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
    }
  }
  
  function handleClear() {
    contextBuilderStore.resetContext()
    fileTreeStore.clearSelection()
  }
  
  function handleExport() {
    // Export context functionality
    console.log('Export context')
  }
  
  function handleImport() {
    // Import context functionality
    console.log('Import context')
  }
  
  function clearSelection() {
    fileTreeStore.clearSelection()
  }
  
  function buildContext() {
    if (projectStore.currentProject?.path) {
      contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
    }
  }
  
  return {
    contextPanelService,
    contextBuilderStore,
    projectStore,
    fileTreeStore,
    settingsStore,
    // Store refs
    contextLoading,
    isBuilding,
    buildStatus,
    contextSummaryState,
    estimatedTokens,
    error,
    validationErrors,
    selectedFiles,
    // Computed properties
    isCollapsed,
    selectedFilesCount,
    contextData,
    hasContextContent,
    contextError,
    contextChunks,
    activeChunk,
    splitSettings,
    selectedFilesText,
    // Methods
    handleToggle,
    handleResize,
    handleRetry,
    handleCopy,
    handleBuild,
    handleClear,
    handleExport,
    handleImport,
    clearSelection,
    buildContext,
    toggleCollapse
  }
}