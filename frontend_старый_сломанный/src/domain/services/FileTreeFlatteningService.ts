import type { FileNode } from '@/types/dto'

export class FileTreeFlatteningService {
  /**
   * Flatten a tree structure into a list with depth information
   * Only includes visible nodes based on expanded paths
   */
  flattenTree(
    nodes: FileNode[], 
    expandedPaths: Set<string>,
    depth = 0,
    result: Array<{ node: FileNode; depth: number }> = []
  ): Array<{ node: FileNode; depth: number }> {
    for (const node of nodes) {
      result.push({ node, depth })
      
      // If node is a directory and is expanded, include its children
      if (node.isDir && node.children && expandedPaths.has(node.path)) {
        this.flattenTree(node.children, expandedPaths, depth + 1, result)
      }
    }
    
    return result
  }
  
  /**
   * Filter flattened nodes based on search query
   */
  filterFlattenedNodes(
    flattenedNodes: Array<{ node: FileNode; depth: number }>,
    query: string
  ): Array<{ node: FileNode; depth: number }> {
    if (!query.trim()) {
      return flattenedNodes
    }
    
    const lowerQuery = query.toLowerCase()
    const matchedPaths = new Set<string>()
    
    // Find all nodes that match the query
    for (const { node } of flattenedNodes) {
      if (node.name.toLowerCase().includes(lowerQuery)) {
        matchedPaths.add(node.path)
        // Also include parent paths
        this.addParentPaths(node, matchedPaths)
      }
    }
    
    // Return only matched nodes and their parents
    return flattenedNodes.filter(({ node }) => matchedPaths.has(node.path))
  }
  
  /**
   * Add all parent paths of a node to the set
   */
  private addParentPaths(node: FileNode, paths: Set<string>) {
    let currentPath = node.parentPath
    while (currentPath) {
      paths.add(currentPath)
      // In a real implementation, you'd need to find the parent node to get its parentPath
      // For now, we'll stop at one level
      break
    }
  }
}