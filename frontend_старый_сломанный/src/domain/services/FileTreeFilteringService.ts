import type { FileNode } from '@/types/dto';

/**
 * File Tree Filtering Service
 * Encapsulates all file tree filtering and searching logic
 */
export class FileTreeFilteringService {
  /**
   * Filters a file tree based on a search query
   * @param nodes - The file nodes to filter
   * @param query - The search query
   * @returns FileNode[] - The filtered nodes
   */
  filterNodes(nodes: FileNode[], query: string): FileNode[] {
    if (!nodes || !query) return nodes || [];

    const lowerQuery = query.toLowerCase();
    return nodes.filter(node => {
      // Hide hidden files if query doesn't specifically target them
      if (node.name.startsWith('.') && !lowerQuery.includes('.')) return false;
      
      // Check if node name or path matches query
      if (node.name.toLowerCase().includes(lowerQuery) || 
          node.path.toLowerCase().includes(lowerQuery)) {
        return true;
      }
      
      // For folders, include if any children match
      if (node.children && node.isDir) {
        return this.hasMatchingChildren(node.children, lowerQuery);
      }
      
      return false;
    }).map(node => {
      // For folders, recursively filter children
      if (node.children && node.isDir) {
        return {
          ...node,
          children: this.filterNodes(node.children, query)
        };
      }
      return node;
    });
  }

  /**
   * Checks if any children match the query
   * @param children - The child nodes to check
   * @param query - The search query
   * @returns boolean - True if any children match
   */
  hasMatchingChildren(children: FileNode[], query: string): boolean {
    return children.some(child => 
      child.name.toLowerCase().includes(query) ||
      child.path.toLowerCase().includes(query) ||
      (child.children && this.hasMatchingChildren(child.children, query))
    );
  }

  /**
   * Counts the number of filtered files
   * @param nodes - The nodes to count
   * @returns number - The count of files
   */
  countFilteredFiles(nodes: FileNode[]): number {
    if (!nodes) return 0;
    
    return nodes.reduce((count, node) => {
      if (!node.isDir) {
        return count + 1;
      } else {
        return count + this.countFilteredFiles(node.children || []);
      }
    }, 0);
  }

  /**
   * Sorts nodes with directories first, then alphabetically
   * @param nodes - The nodes to sort
   * @returns FileNode[] - The sorted nodes
   */
  sortNodes(nodes: FileNode[]): FileNode[] {
    return [...(nodes || [])].sort((a, b) => {
      // Directories first
      if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
      // Then alphabetically
      return a.name.localeCompare(b.name);
    });
  }
}