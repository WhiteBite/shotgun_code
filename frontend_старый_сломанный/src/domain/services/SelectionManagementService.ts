import type { FileNode } from "@/types/dto";
import { APP_CONFIG } from '@/config/app-config';

export type SelectionState = "on" | "off" | "partial";

export interface CascadeResult {
  success: boolean;
  affectedCount: number;
  error?: string;
  affectedPaths?: string[];
}

export interface FileCount {
  totalFiles: number;
  maxDepth: number;
}

/**
 * Domain service for managing tri-state selection logic
 * Encapsulates business rules for file/folder selection with cascade operations
 * Follows DDD principles by containing selection domain knowledge
 */
export class SelectionManagementService {
  private readonly maxCascadeFiles: number;
  private readonly maxRecursionDepth: number;

  constructor() {
    this.maxCascadeFiles = APP_CONFIG.ui.selection.MAX_CASCADE_FILES;
    this.maxRecursionDepth = APP_CONFIG.ui.selection.MAX_RECURSION_DEPTH;
  }

  /**
   * Computes the selection state for a node based on its children
   * Business rule: Directories are partial if some children are selected
   */
  computeSelectionState(
    node: FileNode, 
    selectedPaths: Set<string>, 
    nodesMap: Map<string, FileNode>
  ): SelectionState {
    if (!node.isDir) {
      return selectedPaths.has(node.path) ? "on" : "off";
    }

    const children = (node.children || [])
      .map((c) => nodesMap.get(c.path))
      .filter(Boolean) as FileNode[];

    if (children.length === 0) {
      return selectedPaths.has(node.path) ? "on" : "off";
    }

    let hasOn = false;
    let hasOff = false;

    for (const child of children) {
      if (child.isIgnored) continue;
      
      const childState = this.computeSelectionState(child, selectedPaths, nodesMap);
      
      if (childState === "partial") return "partial";
      if (childState === "on") hasOn = true;
      if (childState === "off") hasOff = true;
      
      if (hasOn && hasOff) return "partial";
    }

    if (hasOn && !hasOff) return "on";
    if (!hasOn && hasOff) return "off";
    return "off";
  }

  /**
   * Performs cascade selection with safety checks
   * Business rule: Prevent memory exhaustion by limiting cascade operations
   */
  performCascadeSelection(
    node: FileNode, 
    selectedPaths: Set<string>, 
    nodesMap: Map<string, FileNode>
  ): CascadeResult {
    // Safety check: Count files first to prevent OutOfMemory
    const fileCount = this.countFilesInSubtree(node, 0, nodesMap);
    
    if (fileCount.totalFiles > this.maxCascadeFiles) {
      return {
        success: false,
        affectedCount: 0,
        error: `Cannot select ${fileCount.totalFiles} files. Maximum allowed: ${this.maxCascadeFiles}.`
      };
    }
    
    if (fileCount.maxDepth > this.maxRecursionDepth) {
      return {
        success: false,
        affectedCount: 0,
        error: `Folder too deep (${fileCount.maxDepth} levels). Maximum allowed: ${this.maxRecursionDepth}.`
      };
    }
    
    // Determine if we're selecting or deselecting
    const isSelecting = !selectedPaths.has(node.path);
    const affectedPaths: string[] = [];
    
    // Use breadth-first traversal to prevent stack overflow
    const queue: FileNode[] = [node];
    
    while (queue.length > 0) {
      const currentNode = queue.shift()!;
      
      if (!currentNode.isIgnored) {
        if (isSelecting) {
          selectedPaths.add(currentNode.path);
        } else {
          selectedPaths.delete(currentNode.path);
        }
        affectedPaths.push(currentNode.path);
        
        // Add children to queue for processing
        if (currentNode.isDir && currentNode.children) {
          for (const childRef of currentNode.children) {
            const childNode = nodesMap.get(childRef.path);
            if (childNode) {
              queue.push(childNode);
            }
          }
        }
      }
    }
    
    return {
      success: true,
      affectedCount: affectedPaths.length,
      affectedPaths
    };
  }

  /**
   * Counts files in subtree with depth checking
   * Business rule: Fail fast if limits are exceeded
   */
  countFilesInSubtree(
    node: FileNode, 
    currentDepth: number, 
    nodesMap: Map<string, FileNode>
  ): FileCount {
    if (currentDepth > this.maxRecursionDepth) {
      return { totalFiles: 999999, maxDepth: currentDepth }; // Fail fast
    }
    
    let totalFiles = node.isDir ? 0 : 1;
    let maxDepth = currentDepth;
    
    if (node.isDir && node.children) {
      for (const childRef of node.children) {
        const child = nodesMap.get(childRef.path);
        if (child && !child.isIgnored) {
          const childCount = this.countFilesInSubtree(child, currentDepth + 1, nodesMap);
          totalFiles += childCount.totalFiles;
          maxDepth = Math.max(maxDepth, childCount.maxDepth);
          
          // Early exit if we already exceed limits
          if (totalFiles > this.maxCascadeFiles || maxDepth > this.maxRecursionDepth) {
            return { totalFiles, maxDepth };
          }
        }
      }
    }
    
    return { totalFiles, maxDepth };
  }

  /**
   * Validates selection constraints
   * Business rule: Ensure selection operations respect system limits
   */
  validateSelectionOperation(
    node: FileNode, 
    nodesMap: Map<string, FileNode>
  ): { valid: boolean; error?: string } {
    const fileCount = this.countFilesInSubtree(node, 0, nodesMap);
    
    if (fileCount.totalFiles > this.maxCascadeFiles) {
      return {
        valid: false,
        error: `Selection would affect ${fileCount.totalFiles} files, exceeding limit of ${this.maxCascadeFiles}`
      };
    }
    
    if (fileCount.maxDepth > this.maxRecursionDepth) {
      return {
        valid: false,
        error: `Directory structure too deep (${fileCount.maxDepth} levels), exceeding limit of ${this.maxRecursionDepth}`
      };
    }
    
    return { valid: true };
  }

  /**
   * Gets selection statistics for a subtree
   * Business rule: Provide insights for user decision-making
   */
  getSelectionStats(
    node: FileNode, 
    selectedPaths: Set<string>, 
    nodesMap: Map<string, FileNode>
  ): {
    totalFiles: number;
    selectedFiles: number;
    totalDirectories: number;
    selectedDirectories: number;
    maxDepth: number;
  } {
    const stats = {
      totalFiles: 0,
      selectedFiles: 0,
      totalDirectories: 0,
      selectedDirectories: 0,
      maxDepth: 0
    };

    const traverse = (currentNode: FileNode, depth: number) => {
      stats.maxDepth = Math.max(stats.maxDepth, depth);
      
      if (currentNode.isDir) {
        stats.totalDirectories++;
        if (selectedPaths.has(currentNode.path)) {
          stats.selectedDirectories++;
        }
        
        if (currentNode.children) {
          for (const childRef of currentNode.children) {
            const child = nodesMap.get(childRef.path);
            if (child && !child.isIgnored) {
              traverse(child, depth + 1);
            }
          }
        }
      } else {
        stats.totalFiles++;
        if (selectedPaths.has(currentNode.path)) {
          stats.selectedFiles++;
        }
      }
    };

    traverse(node, 0);
    return stats;
  }

  /**
   * Optimizes selection set by removing redundant selections
   * Business rule: If parent is selected, children don't need explicit selection
   */
  optimizeSelectionSet(
    selectedPaths: Set<string>, 
    nodesMap: Map<string, FileNode>
  ): Set<string> {
    const optimized = new Set<string>();
    
    for (const path of selectedPaths) {
      const node = nodesMap.get(path);
      if (!node) continue;
      
      // Check if any parent is already selected
      let hasSelectedParent = false;
      let parentPath = this.getParentPath(path);
      
      while (parentPath) {
        if (selectedPaths.has(parentPath)) {
          hasSelectedParent = true;
          break;
        }
        parentPath = this.getParentPath(parentPath);
      }
      
      // Only keep if no parent is selected
      if (!hasSelectedParent) {
        optimized.add(path);
      }
    }
    
    return optimized;
  }

  /**
   * Gets parent path for a given path
   */
  private getParentPath(path: string): string | null {
    const lastSlash = path.lastIndexOf('/');
    if (lastSlash <= 0) return null;
    return path.substring(0, lastSlash);
  }
}

// Default instance for dependency injection
export const defaultSelectionManagementService = new SelectionManagementService();