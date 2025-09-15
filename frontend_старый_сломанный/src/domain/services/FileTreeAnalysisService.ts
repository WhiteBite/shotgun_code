import type { FileNode } from '@/types/dto';

/**
 * File Tree Analysis Service
 * Encapsulates all file tree analysis and metrics calculation logic
 */
export class FileTreeAnalysisService {
  /**
   * Counts the total number of files in a file tree
   * @param nodes - The nodes to count
   * @returns number - The total file count
   */
  countTotalFiles(nodes: FileNode[]): number {
    if (!nodes) return 0;
    
    let count = 0;
    const countFiles = (ns: FileNode[]) => {
      for (const n of ns) {
        if (!n.isDir) count++;
        if (n.children) countFiles(n.children);
      }
    };
    countFiles(nodes);
    return count;
  }

  /**
   * Counts the number of selected files
   * @param nodesMap - Map of all nodes by path
   * @param selectedPaths - Set of selected paths
   * @returns number - The count of selected files
   */
  countSelectedFiles(nodesMap: Map<string, FileNode>, selectedPaths: Set<string>): number {
    if (!nodesMap || !selectedPaths) return 0;
    
    let count = 0;
    nodesMap.forEach((node, path) => {
      if (selectedPaths.has(path) && !node.isDir && !node.isIgnored) {
        count++;
      }
    });
    
    return count;
  }

  /**
   * Collects all files from a file tree
   * @param nodes - The nodes to collect from
   * @param maxFiles - Maximum number of files to collect
   * @returns FileNode[] - The collected files
   */
  collectAllFiles(nodes: FileNode[], maxFiles: number = 10000): FileNode[] {
    const allFiles: FileNode[] = [];
    
    const collectFiles = (ns: readonly FileNode[]) => {
      for (const n of ns) {
        if (allFiles.length >= maxFiles) {
          return;
        }
        
        if (!n.isDir) {
          allFiles.push(n);
        }
        
        if (n.children && allFiles.length < maxFiles) {
          collectFiles(n.children);
        }
      }
    };
    
    collectFiles(nodes);
    return allFiles;
  }

  /**
   * Gets file statistics
   * @param nodes - The nodes to analyze
   * @returns object - Statistics about the file tree
   */
  getFileTreeStats(nodes: FileNode[]): { 
    totalFiles: number; 
    totalDirectories: number; 
    totalSize: number;
    fileTypes: Record<string, number>;
  } {
    let totalFiles = 0;
    let totalDirectories = 0;
    let totalSize = 0;
    const fileTypes: Record<string, number> = {};

    const analyzeNodes = (ns: FileNode[]) => {
      for (const n of ns) {
        if (n.isDir) {
          totalDirectories++;
          if (n.children) analyzeNodes(n.children);
        } else {
          totalFiles++;
          totalSize += n.size || 0;
          
          // Extract file extension
          const ext = n.name.split('.').pop()?.toLowerCase() || 'unknown';
          fileTypes[ext] = (fileTypes[ext] || 0) + 1;
        }
      }
    };

    analyzeNodes(nodes);
    
    return {
      totalFiles,
      totalDirectories,
      totalSize,
      fileTypes
    };
  }
}