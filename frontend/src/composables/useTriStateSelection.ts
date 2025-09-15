import type { FileNode } from "@/types/dto";

export type Sel = "on" | "off" | "partial";

// CRITICAL: Protection limits for cascade selection
const MAX_CASCADE_FILES = 100;  // Maximum files in single cascade operation
const MAX_RECURSION_DEPTH = 5;  // Maximum folder nesting depth

interface CascadeResult {
  success: boolean;
  affectedCount: number;
  error?: string;
}

export function useTriStateSelection(
  nodesMap: Map<string, FileNode>,
  selectedPaths: Set<string>,
) {
  function computeSelection(node: FileNode): Sel {
    if (!node.isDir) return selectedPaths.has(node.path) ? "on" : "off";
    const children = (node.children || [])
      .map((c) => nodesMap.get(c.path))
      .filter(Boolean) as FileNode[];
    if (children.length === 0)
      return selectedPaths.has(node.path) ? "on" : "off";
    let hasOn = false,
      hasOff = false;
    for (const ch of children) {
      if (ch.isIgnored) continue;
      const st = computeSelection(ch);
      if (st === "partial") return "partial";
      if (st === "on") hasOn = true;
      if (st === "off") hasOff = true;
      if (hasOn && hasOff) return "partial";
    }
    if (hasOn && !hasOff) return "on";
    if (!hasOn && hasOff) return "off";
    return "off";
  }

  function toggleCascade(node: FileNode): CascadeResult {
    // CRITICAL: Count files first to prevent OutOfMemory
    const fileCount = countFilesInSubtree(node, 0);
    
    if (fileCount.totalFiles > MAX_CASCADE_FILES) {
      console.warn(`Cascade selection blocked: ${fileCount.totalFiles} files exceeds limit of ${MAX_CASCADE_FILES}`);
      return {
        success: false,
        affectedCount: 0,
        error: `Cannot select ${fileCount.totalFiles} files. Maximum allowed: ${MAX_CASCADE_FILES}.`
      };
    }
    
    if (fileCount.maxDepth > MAX_RECURSION_DEPTH) {
      console.warn(`Cascade selection blocked: depth ${fileCount.maxDepth} exceeds limit of ${MAX_RECURSION_DEPTH}`);
      return {
        success: false,
        affectedCount: 0,
        error: `Folder too deep (${fileCount.maxDepth} levels). Maximum allowed: ${MAX_RECURSION_DEPTH}.`
      };
    }
    
    // Proceed with safe cascade selection
    const select = !selectedPaths.has(node.path);
    const affectedPaths: string[] = [];
    
    // Use breadth-first traversal instead of recursive to prevent stack overflow
    const queue: FileNode[] = [node];
    
    while (queue.length > 0) {
      const cur = queue.shift()!;
      if (!cur.isIgnored) {
        if (select) {
          selectedPaths.add(cur.path);
        } else {
          selectedPaths.delete(cur.path);
        }
        affectedPaths.push(cur.path);
        
        if (cur.isDir && cur.children) {
          for (const c of cur.children) {
            const n = nodesMap.get(c.path);
            if (n) queue.push(n);
          }
        }
      }
    }
    
    return {
      success: true,
      affectedCount: affectedPaths.length,
    };
  }
  
  // Helper function to count files and check depth before cascade operation
  function countFilesInSubtree(node: FileNode, currentDepth: number): { totalFiles: number; maxDepth: number } {
    if (currentDepth > MAX_RECURSION_DEPTH) {
      return { totalFiles: 999999, maxDepth: currentDepth }; // Fail fast
    }
    
    let totalFiles = node.isDir ? 0 : 1;
    let maxDepth = currentDepth;
    
    if (node.isDir && node.children) {
      for (const childRef of node.children) {
        const child = nodesMap.get(childRef.path);
        if (child && !child.isIgnored) {
          const childCount = countFilesInSubtree(child, currentDepth + 1);
          totalFiles += childCount.totalFiles;
          maxDepth = Math.max(maxDepth, childCount.maxDepth);
          
          // Early exit if we already exceed limits
          if (totalFiles > MAX_CASCADE_FILES || maxDepth > MAX_RECURSION_DEPTH) {
            return { totalFiles, maxDepth };
          }
        }
      }
    }
    
    return { totalFiles, maxDepth };
  }

  return { computeSelection, toggleCascade, countFilesInSubtree };
}