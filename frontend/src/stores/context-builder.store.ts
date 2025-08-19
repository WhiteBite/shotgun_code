import { defineStore } from 'pinia'
import { ref, computed, readonly, watch } from 'vue'
import type { FileNode } from '@/types/api'
import { useContext } from '@/composables/useContext'
import { useFileTreeStore } from '@/stores/file-tree.store'

export const useContextBuilderStore = defineStore('context-builder', () => {
  const { buildContext, currentContext, isLoading: contextLoading, error: contextError } = useContext()
  const fileTreeStore = useFileTreeStore()
  
  const selectedFilesList = ref<string[]>([])
  const isBuilding = ref(false)
  const error = ref<string | null>(null)
  const shotgunContextText = ref<string>('')
  const lastContextGeneration = ref<Date | null>(null)
  const contextSummary = ref({
    files: 0,
    characters: 0,
    tokens: 0,
    cost: 0
  })
  const contextStatus = ref({
    status: 'none' as 'none' | 'current' | 'changed' | 'stale',
    message: 'No context built'
  })

  const hasSelectedFiles = computed(() => selectedFilesList.value.length > 0)
  const selectedFilesCount = computed(() => {
    // Count only files, not directories
    return selectedFilesList.value.filter(filePath => {
      const node = fileTreeStore.getFileByRelPath(filePath)
      return node && !node.isDir && !node.isGitignored && !node.isCustomIgnored
    }).length
  })
  const totalFilesCount = computed(() => fileTreeStore.totalFiles)

  // Sync selected files from file tree store
  watch(() => fileTreeStore.selectedFiles, (newSelectedFiles) => {
    selectedFilesList.value = newSelectedFiles
  }, { immediate: true })

  async function buildContextFromSelection(projectPath: string, options?: { 
    includeGitStatus?: boolean
    includeCommitHistory?: boolean
    maxTokens?: number
  }) {
    isBuilding.value = true
    error.value = null
    
    try {
      // Just call buildContext to start generation
      // The real context will be set via setShotgunContext when backend responds
      await buildContext(projectPath, selectedFilesList.value, options)
      
      // Don't update currentContext here - it will be updated by setShotgunContext
      // when the backend sends the shotgunContextGenerated event
      
      return currentContext.value
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to build context'
      error.value = errorMessage
      throw new Error(errorMessage)
    } finally {
      isBuilding.value = false
    }
  }

  function setSelectedFiles(files: string[]) {
    selectedFilesList.value = files
  }

  function addSelectedFile(file: string) {
    if (!selectedFilesList.value.includes(file)) {
      selectedFilesList.value.push(file)
    }
  }

  function removeSelectedFile(file: string) {
    selectedFilesList.value = selectedFilesList.value.filter(f => f !== file)
  }

  function clearSelectedFiles() {
    selectedFilesList.value = []
  }

  function clearError() {
    error.value = null
  }

  function setShotgunContext(context: string) {
    console.log("setShotgunContext called with context length:", context?.length);
    console.log("Context preview:", context?.substring(0, 200));
    
    shotgunContextText.value = context
    lastContextGeneration.value = new Date()
    
    // Update currentContext with real content from backend
    if (currentContext.value) {
      currentContext.value.content = context
      currentContext.value.name = `Контекст ${new Date().toLocaleString('ru-RU')}`
      currentContext.value.description = `Контекст с ${selectedFilesList.value.length} файлами`
      currentContext.value.updatedAt = new Date().toISOString()
      currentContext.value.tokenCount = Math.ceil(context.length / 4)
    }
    
    console.log("currentContext updated:", currentContext.value?.content?.substring(0, 100));
    
    // Update summary
    contextSummary.value = {
      files: selectedFilesList.value.length,
      characters: context.length,
      tokens: Math.ceil(context.length / 4), // Rough estimate
      cost: Math.ceil(context.length / 4) * 0.0001 // Rough estimate
    }
    
    // Update status
    contextStatus.value = {
      status: 'current',
      message: `Контекст построен с ${selectedFilesList.value.length} файлами`
    }
  }

  return {
    // State
    selectedFiles: readonly(selectedFilesList),
    isBuilding: readonly(isBuilding),
    error: readonly(error),
    currentContext: readonly(currentContext),
    contextLoading: readonly(contextLoading),
    contextError: readonly(contextError),
    shotgunContextText: readonly(shotgunContextText),
    lastContextGeneration: readonly(lastContextGeneration),
    contextSummary: readonly(contextSummary),
    contextStatus: readonly(contextStatus),
    
    // Computed
    hasSelectedFiles,
    selectedFilesCount,
    totalFilesCount,
    
    // Methods
    buildContextFromSelection,
    setSelectedFiles,
    addSelectedFile,
    removeSelectedFile,
    clearSelectedFiles,
    clearError,
    setShotgunContext
  }
})