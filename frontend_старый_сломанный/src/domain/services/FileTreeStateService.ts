import { ref } from 'vue'

export interface TreeNode {
  path: string
  relPath: string
  name: string
  isDir: boolean
  children?: TreeNode[]
  // Add other properties as needed
}

export class FileTreeStateService {
  private nodesMap = ref<Map<string, TreeNode>>(new Map())
  private selectedPaths = ref<Set<string>>(new Set())
  private expandedPaths = ref<Set<string>>(new Set())

  constructor() {}

  // Node management
  getNodesMap() {
    return this.nodesMap.value
  }

  setNodesMap(nodesMap: Map<string, TreeNode>) {
    this.nodesMap.value = nodesMap
  }

  getNode(path: string): TreeNode | undefined {
    return this.nodesMap.value.get(path)
  }

  // Selection management
  getSelectedPaths() {
    return this.selectedPaths.value
  }

  isSelected(path: string): boolean {
    return this.selectedPaths.value.has(path)
  }

  toggleSelection(path: string) {
    if (this.selectedPaths.value.has(path)) {
      this.selectedPaths.value.delete(path)
    } else {
      this.selectedPaths.value.add(path)
    }
  }

  selectPath(path: string) {
    this.selectedPaths.value.add(path)
  }

  deselectPath(path: string) {
    this.selectedPaths.value.delete(path)
  }

  clearSelection() {
    this.selectedPaths.value.clear()
  }

  // Expansion management
  getExpandedPaths() {
    return this.expandedPaths.value
  }

  isExpanded(path: string): boolean {
    return this.expandedPaths.value.has(path)
  }

  toggleExpansion(path: string) {
    if (this.expandedPaths.value.has(path)) {
      this.expandedPaths.value.delete(path)
    } else {
      this.expandedPaths.value.add(path)
    }
  }

  expandPath(path: string) {
    this.expandedPaths.value.add(path)
  }

  collapsePath(path: string) {
    this.expandedPaths.value.delete(path)
  }

  // Recursive operations
  toggleExpansionRecursive(path: string, nodesMap: Map<string, TreeNode>, expand: boolean) {
    // Implementation would depend on the specific logic needed
    // This is a placeholder for the recursive expansion logic
  }
}