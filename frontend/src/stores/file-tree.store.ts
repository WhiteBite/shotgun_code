import { defineStore } from 'pinia'
import { ref, computed, readonly } from 'vue'
import { useFileTree } from '@/composables/useFileTree'
import { useTreeStateStore } from './tree-state.store'

export const useFileTreeStore = defineStore('file-tree', () => {
  const {
    nodes,
    nodesMap,
    isLoading,
    error,
    searchQuery,
    visibleNodes,
    selectedFiles,
    totalFiles,
    fetchFileTree,
    toggleNodeSelection,
    clearSelection,
    selectAll,
    setSearchQuery,
    clearError
  } = useFileTree()
  
  const rootPath = ref<string>('')
  const useGitignore = ref(true)
  const useCustomIgnore = ref(true)

  const hasFiles = computed(() => nodes.value.length > 0)
  const hasSelectedFiles = computed(() => selectedFiles.value.length > 0)

  // Track expanded state
  const expandedPaths = ref<Set<string>>(new Set())

  function isNodeExpanded(path: string): boolean {
    return expandedPaths.value.has(path)
  }

  function toggleNodeExpanded(path: string) {
    if (expandedPaths.value.has(path)) {
      expandedPaths.value.delete(path)
    } else {
      expandedPaths.value.add(path)
    }
  }

  async function loadProject(projectPath: string) {
    rootPath.value = projectPath
    await fetchFileTree(projectPath, useGitignore.value, useCustomIgnore.value)
    
    // Auto-expand root folder by default
    if (nodes.value.length > 0) {
      expandedPaths.value.add(nodes.value[0].path)
      // Also expand first level directories
      nodes.value.forEach(node => {
        if (node.isDir && node.children && node.children.length > 0) {
          expandedPaths.value.add(node.path)
        }
      })
    }
    
    // Clear previous selections when loading new project
    clearSelection()
  }

  async function refreshFiles() {
    if (!rootPath.value) {
      return
    }
    
    await fetchFileTree(rootPath.value, useGitignore.value, useCustomIgnore.value)
  }

  async function updateIgnoreSettings(newUseGitignore: boolean, newUseCustomIgnore: boolean) {
    useGitignore.value = newUseGitignore
    useCustomIgnore.value = newUseCustomIgnore
    
    if (rootPath.value) {
      await refreshFiles()
    }
  }

  function clearProject() {
    rootPath.value = ''
    clearSelection()
    clearError()
  }

  function getAllFiles(): any[] {
    const allFiles: any[] = []
    
    const collectFiles = (nodes: readonly any[]) => {
      for (const node of nodes) {
        if (!node.isDir) {
          allFiles.push(node)
        }
        if (node.children) {
          collectFiles(node.children)
        }
      }
    }
    
    collectFiles(nodes.value)
    return allFiles
  }

  function getFileByPath(path: string): any | undefined {
    return nodesMap.value.get(path)
  }

  function getFileByRelPath(relPath: string): any | undefined {
    for (const node of nodesMap.value.values()) {
      if (node.relPath === relPath) {
        return node
      }
    }
    return undefined
  }

  function setSelectedFiles(filePaths: string[]) {
    // Очищаем текущий выбор
    clearSelection()
    
    // Добавляем новые файлы
    filePaths.forEach(relPath => {
      // Ищем файл по relPath
      const node = getFileByRelPath(relPath)
      if (node && !node.isDir && !node.isGitignored && !node.isCustomIgnored) {
        // Используем toggleNodeSelection для правильного обновления состояния
        toggleNodeSelection(node.path)
      }
    })
  }

  return {
    // State
    rootPath: readonly(rootPath),
    useGitignore,
    useCustomIgnore,
    
    // From useFileTree
    nodes: readonly(nodes),
    nodesMap: readonly(nodesMap),
    isLoading: readonly(isLoading),
    error: readonly(error),
    searchQuery,
    visibleNodes,
    selectedFiles,
    totalFiles,
    
    // Computed
    hasFiles,
    hasSelectedFiles,
    
    // Methods
    loadProject,
    refreshFiles,
    updateIgnoreSettings,
    clearProject,
    getAllFiles,
    getFileByPath,
    getFileByRelPath,
    toggleNodeSelection,
    clearSelection,
    selectAll,
    setSearchQuery,
    clearError,
    isNodeExpanded,
    toggleNodeExpanded,
    setSelectedFiles
  }
})