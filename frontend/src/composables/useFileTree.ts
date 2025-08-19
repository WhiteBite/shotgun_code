import { ref, computed, readonly } from 'vue'
import type { FileNode } from '../types/api'
import type { DomainFileNode } from '../types/dto'
import { apiService } from '../services/api.service'
import { useErrorHandler } from './useErrorHandler'
import { useTreeStateStore } from '../stores/tree-state.store'

export function useFileTree() {
  const { handleError } = useErrorHandler()
  const treeStateStore = useTreeStateStore()
  
  const nodes = ref<FileNode[]>([])
  const nodesMap = ref<Map<string, FileNode>>(new Map())
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const searchQuery = ref('')

  const visibleNodes = computed(() => {
    if (!searchQuery.value.trim()) {
      return nodes.value
    }
    
    const query = searchQuery.value.toLowerCase()
    return filterNodes(nodes.value, query)
  })

  const selectedFiles = computed(() => {
    const selected: string[] = []
    nodesMap.value.forEach((node, path) => {
      if (treeStateStore.selectedPaths.has(path)) {
        selected.push(node.relPath)
      }
    })
    return selected
  })

  const totalFiles = computed(() => {
    let count = 0
    const countFiles = (nodes: FileNode[]) => {
      for (const node of nodes) {
        if (!node.isDir) {
          count++
        }
        if (node.children) {
          countFiles(node.children)
        }
      }
    }
    countFiles(nodes.value)
    return count
  })

  async function fetchFileTree(projectPath: string, useGitignore = true, useCustomIgnore = true) {
    isLoading.value = true
    error.value = null
    
    try {
      console.log('Fetching file tree for:', projectPath, { useGitignore, useCustomIgnore })
      const domainNodes = await apiService.listFiles(projectPath, useGitignore, useCustomIgnore)
      console.log('Received domain nodes:', domainNodes?.length || 0, 'nodes')
      console.log('First few nodes:', domainNodes?.slice(0, 3))
      
      if (!domainNodes || domainNodes.length === 0) {
        console.warn('No files received from API')
        nodes.value = []
        nodesMap.value.clear()
        return []
      }
      
      const fileNodes = mapDomainNodesToFileNodes(domainNodes)
      console.log('Mapped to file nodes:', fileNodes?.length || 0, 'nodes')
      console.log('First few mapped nodes:', fileNodes?.slice(0, 3))
      
      nodes.value = fileNodes
      buildNodesMap(fileNodes)
      console.log('Nodes map built, size:', nodesMap.value.size)
      return fileNodes
    } catch (err) {
      console.error('Error fetching file tree:', err)
      const errorMessage = err instanceof Error ? err.message : String(err)
      error.value = errorMessage
      handleError(err, 'Loading file tree')
      throw new Error(errorMessage)
    } finally {
      isLoading.value = false
    }
  }

  function mapDomainNodesToFileNodes(domainNodes: DomainFileNode[]): FileNode[] {
    return domainNodes.map((node: DomainFileNode) => {
      const fileNode: FileNode = {
        name: node.name,
        path: node.path,
        relPath: node.relPath,
        isDir: node.isDir,
        size: node.size,
        children: node.children ? mapDomainNodesToFileNodes(node.children) : undefined,
        isGitignored: node.isGitignored,
        isCustomIgnored: node.isCustomIgnored,
        isSelected: treeStateStore.selectedPaths.has(node.path) // Используем treeStateStore для определения выделения
      }
      return fileNode
    })
  }

  function buildNodesMap(fileNodes: FileNode[]) {
    const map = new Map<string, FileNode>()
    
    const addToMap = (nodes: FileNode[]) => {
      for (const node of nodes) {
        map.set(node.path, node)
        if (node.children) {
          addToMap(node.children)
        }
      }
    }
    
    addToMap(fileNodes)
    nodesMap.value = map
  }

  function filterNodes(nodes: FileNode[], query: string): FileNode[] {
    const filtered: FileNode[] = []
    
    for (const node of nodes) {
      const matchesQuery = node.name.toLowerCase().includes(query) ||
                          node.relPath.toLowerCase().includes(query)
      
      if (matchesQuery) {
        // Create a copy to avoid mutating original
        const filteredNode = { ...node }
        if (node.children) {
          filteredNode.children = filterNodes(node.children, query)
        }
        filtered.push(filteredNode)
      } else if (node.children) {
        // Check if any child matches
        const filteredChildren = filterNodes(node.children, query)
        if (filteredChildren.length > 0) {
          const filteredNode = { ...node, children: filteredChildren }
          filtered.push(filteredNode)
        }
      }
    }
    
    return filtered
  }

  function toggleNodeSelection(path: string) {
    const node = nodesMap.value.get(path)
    if (node) {
      // Используем treeStateStore для управления выделением
      treeStateStore.toggleNodeSelection(path, nodesMap.value as Map<string, any>)
      
      // Обновляем isSelected для всех узлов после изменения
      updateAllNodesSelectionState()
    }
  }

  function updateChildrenSelectionState(children: FileNode[]) {
    for (const child of children) {
      child.isSelected = treeStateStore.selectedPaths.has(child.path)
      if (child.children) {
        updateChildrenSelectionState(child.children)
      }
    }
  }

  function updateParentSelectionState(childPath: string) {
    const parentPath = getParentPath(childPath)
    if (parentPath) {
      const parent = nodesMap.value.get(parentPath)
      if (parent && parent.children) {
        // Обновляем isSelected на основе состояния в treeStateStore
        parent.isSelected = treeStateStore.selectedPaths.has(parentPath)
        updateParentSelectionState(parentPath)
      }
    }
  }

  function updateAllNodesSelectionState() {
    // Обновляем isSelected для всех узлов на основе treeStateStore
    nodesMap.value.forEach(node => {
      node.isSelected = treeStateStore.selectedPaths.has(node.path)
    })
  }

  function getParentPath(path: string): string | null {
    const lastSlashIndex = path.lastIndexOf('/')
    if (lastSlashIndex === -1) {
      return null
    }
    return path.substring(0, lastSlashIndex)
  }

  function clearSelection() {
    treeStateStore.clearSelection()
    updateAllNodesSelectionState()
  }

  function selectAll() {
    nodesMap.value.forEach(node => {
      if (!node.isDir && !node.isGitignored && !node.isCustomIgnored) {
        treeStateStore.selectedPaths.add(node.path)
      }
    })
    updateAllNodesSelectionState()
  }

  function setSearchQuery(query: string) {
    searchQuery.value = query
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    nodes: readonly(nodes),
    nodesMap: readonly(nodesMap),
    isLoading: readonly(isLoading),
    error: readonly(error),
    searchQuery,
    
    // Computed
    visibleNodes,
    selectedFiles,
    totalFiles,
    
    // Methods
    fetchFileTree,
    toggleNodeSelection,
    clearSelection,
    selectAll,
    setSearchQuery,
    clearError
  }
}
