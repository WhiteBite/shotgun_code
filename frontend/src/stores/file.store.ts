import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface FileNode {
  name: string
  path: string
  isDir: boolean
  isExpanded?: boolean
  isSelected?: boolean
  children?: FileNode[]
}

export const useFileStore = defineStore('file', () => {
  // State
  const nodes = ref<FileNode[]>([])
  const selectedPaths = ref<Set<string>>(new Set())

  // Computed
  const hasSelectedFiles = computed(() => selectedPaths.value.size > 0)
  const selectedCount = computed(() => selectedPaths.value.size)
  const selectedFilesList = computed(() => Array.from(selectedPaths.value))

  // Actions
  function setFileTree(tree: FileNode[]) {
    nodes.value = tree
  }

  function toggleSelect(path: string) {
    if (selectedPaths.value.has(path)) {
      selectedPaths.value.delete(path)
    } else {
      selectedPaths.value.add(path)
    }
  }

  function selectPath(path: string) {
    selectedPaths.value.add(path)
  }

  function deselectPath(path: string) {
    selectedPaths.value.delete(path)
  }

  function selectMultiple(paths: string[]) {
    paths.forEach(p => selectedPaths.value.add(p))
  }

  function clearSelection() {
    selectedPaths.value.clear()
  }

  function toggleExpand(path: string) {
    const node = findNode(nodes.value, path)
    if (node && node.isDir) {
      node.isExpanded = !node.isExpanded
    }
  }

  function expandPath(path: string) {
    const node = findNode(nodes.value, path)
    if (node && node.isDir) {
      node.isExpanded = true
    }
  }

  function collapsePath(path: string) {
    const node = findNode(nodes.value, path)
    if (node && node.isDir) {
      node.isExpanded = false
    }
  }

  function expandAll() {
    walkTree(nodes.value, node => {
      if (node.isDir) {
        node.isExpanded = true
      }
    })
  }

  function collapseAll() {
    walkTree(nodes.value, node => {
      if (node.isDir) {
        node.isExpanded = false
      }
    })
  }

  // Recursive selection
  function selectRecursive(path: string) {
    const node = findNode(nodes.value, path)
    if (node) {
      selectedPaths.value.add(path)
      if (node.children) {
        node.children.forEach(child => selectRecursive(child.path))
      }
    }
  }

  function deselectRecursive(path: string) {
    const node = findNode(nodes.value, path)
    if (node) {
      selectedPaths.value.delete(path)
      if (node.children) {
        node.children.forEach(child => deselectRecursive(child.path))
      }
    }
  }

  // Helper functions
  function findNode(tree: FileNode[], path: string): FileNode | null {
    for (const node of tree) {
      if (node.path === path) return node
      if (node.children) {
        const found = findNode(node.children, path)
        if (found) return found
      }
    }
    return null
  }

  function walkTree(tree: FileNode[], fn: (node: FileNode) => void) {
    for (const node of tree) {
      fn(node)
      if (node.children) {
        walkTree(node.children, fn)
      }
    }
  }

  return {
    // State
    nodes,
    selectedPaths,
    // Computed
    hasSelectedFiles,
    selectedCount,
    selectedFilesList,
    // Actions
    setFileTree,
    toggleSelect,
    selectPath,
    deselectPath,
    selectMultiple,
    clearSelection,
    toggleExpand,
    expandPath,
    collapsePath,
    expandAll,
    collapseAll,
    selectRecursive,
    deselectRecursive
  }
})
