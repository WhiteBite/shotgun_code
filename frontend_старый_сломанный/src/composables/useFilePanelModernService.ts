import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useProjectStore } from '@/stores/project.store'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { FilePanelModernService } from '@/domain/services/FilePanelModernService'

export function useFilePanelModernService() {
  // Get stores
  const fileTreeStore = useFileTreeStore()
  const projectStore = useProjectStore()
  const contextBuilderStore = useContextBuilderStore()
  
  // Get reactive refs from stores
  const { 
    isLoading,
    error,
    selectedFiles,
    nodes: fileTree,
    totalFiles: totalFileCount
  } = storeToRefs(fileTreeStore)
  
  const { currentProject } = storeToRefs(projectStore)
  
  // Create service instance
  const filePanelModernService = new FilePanelModernService()
  
  // State
  const searchQuery = ref('')
  const showHidden = ref(false)
  const expandedFolders = ref<Set<string>>(new Set())
  
  // Computed properties that depend on store state
  const fileError = computed(() => {
    if (error?.value) {
      return {
        message: error.value,
        code: 'FILE_TREE_ERROR'
      }
    }
    return null
  })
  
  const hasFiles = computed(() => fileTree.value && Array.isArray(fileTree.value) && fileTree.value.length > 0)
  
  const filteredFileTree = computed(() => {
    if (!fileTree.value || !Array.isArray(fileTree.value)) return []
    return filePanelModernService.getFilteredFileTree(Array.from(fileTree.value), searchQuery.value || '')
  })
  
  const filteredFileCount = computed(() => {
    return filePanelModernService.getFilteredFileCount(
      searchQuery.value || '', 
      totalFileCount.value || 0, 
      filteredFileTree.value || [], 
      Array.from(fileTree.value || [])
    )
  })
  
  const totalSelectedSize = computed(() => {
    if (!selectedFiles.value) return 0
    return selectedFiles.value.reduce((total, file) => {
      // This would need to be implemented with actual file size data
      return total + 1024 // Placeholder
    }, 0)
  })
  
  // Methods
  function handleToggle(collapsed: boolean) {
    // This will be handled by the parent component
  }
  
  function handleResize(width: number) {
    // This will be handled by the parent component
  }
  
  function handleRetry() {
    refreshFileTree()
  }
  
  async function refreshFileTree() {
    // Check if project is loaded, if not try to initialize with current directory
    if (!projectStore.currentProject?.path) {
      console.warn('No project loaded, attempting to auto-load current directory...');
      try {
        await projectStore.tryAutoOpenLastProject();
      } catch (error) {
        console.error('Failed to auto-load project:', error);
        return;
      }
    }
    
    await fileTreeStore.refreshFiles();
  }
  
  function toggleShowHidden() {
    filePanelModernService.toggleShowHidden()
  }
  
  function handleSearch() {
    filePanelModernService.handleSearch()
  }
  
  function clearSearch() {
    filePanelModernService.clearSearch()
  }
  
  function handleFileSelect(filePath: string, isSelected: boolean) {
    filePanelModernService.handleFileSelect(
      filePath, 
      isSelected, 
      (path) => fileTreeStore.toggleNodeSelection(path),
      (path) => fileTreeStore.toggleNodeSelection(path)
    )
  }
  
  function handleFolderToggle(folderPath: string, isExpanded: boolean) {
    filePanelModernService.handleFolderToggle(folderPath, isExpanded)
  }
  
  function handleFileContextMenu(filePath: string, event: MouseEvent) {
    filePanelModernService.handleFileContextMenu(filePath, event)
  }
  
  function clearSelection() {
    filePanelModernService.clearSelection(() => fileTreeStore.clearSelection())
  }
  
  function buildContext() {
    filePanelModernService.buildContext(
      projectStore.currentProject?.path, 
      (path) => contextBuilderStore.buildContextFromSelection(path)
    )
  }
  
  function formatFileSize(bytes: number): string {
    return filePanelModernService.formatFileSize(bytes)
  }
  
  // Watch for project changes to refresh file tree
  watch(
    () => projectStore.currentProject?.path,
    (newPath) => {
      if (newPath) {
        refreshFileTree()
      }
    },
    { immediate: true }
  )
  
  return {
    filePanelModernService,
    fileTreeStore,
    projectStore,
    contextBuilderStore,
    // Store refs
    isLoading,
    error,
    selectedFiles,
    fileTree,
    totalFileCount,
    currentProject,
    // Local state
    searchQuery,
    showHidden,
    expandedFolders,
    // Computed properties
    fileError,
    hasFiles,
    filteredFileTree,
    filteredFileCount,
    totalSelectedSize,
    // Methods
    handleToggle,
    handleResize,
    handleRetry,
    refreshFileTree,
    toggleShowHidden,
    handleSearch,
    clearSearch,
    handleFileSelect,
    handleFolderToggle,
    handleFileContextMenu,
    clearSelection,
    buildContext,
    formatFileSize
  }
}