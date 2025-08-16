import { defineStore } from "pinia";
import { ref } from "vue";
import type { FileNode } from "@/types/dto";

export const useTreeStateStore = defineStore("treeState", () => {
  const expandedPaths = ref(new Set<string>());
  const selectedPaths = ref(new Set<string>());
  const activeNodePath = ref<string | null>(null);

  function toggleExpansion(path: string) {
    if (expandedPaths.value.has(path)) {
      expandedPaths.value.delete(path);
    } else {
      expandedPaths.value.add(path);
    }
  }

  function expandRecursive(rootPath: string, nodesMap: Map<string, FileNode>) {
    const root = nodesMap.get(rootPath);
    if (!root || !root.isDir) return;
    const stack: FileNode[] = [root];
    while (stack.length) {
      const n = stack.pop()!;
      if (n.isDir) {
        expandedPaths.value.add(n.path);
        if (n.children) {
          n.children.forEach(c => {
            const child = nodesMap.get(c.path);
            if (child) stack.push(child);
          });
        }
      }
    }
  }

  function collapseRecursive(rootPath: string, nodesMap: Map<string, FileNode>) {
    const root = nodesMap.get(rootPath);
    if (!root || !root.isDir) return;
    const stack: FileNode[] = [root];
    while (stack.length) {
      const n = stack.pop()!;
      expandedPaths.value.delete(n.path);
      if (n.children) {
        n.children.forEach(c => {
          const child = nodesMap.get(c.path);
          if (child) stack.push(child);
        });
      }
    }
  }

  function toggleExpansionRecursive(path: string, nodesMap: Map<string, FileNode>, expand: boolean) {
    if (expand) expandRecursive(path, nodesMap);
    else collapseRecursive(path, nodesMap);
  }

  function toggleSelection(path: string) {
    if (selectedPaths.value.has(path)) {
      selectedPaths.value.delete(path);
    } else {
      selectedPaths.value.add(path);
    }
  }

  function toggleNodeSelection(path: string, nodesMap: Map<string, FileNode>) {
    const node = nodesMap.get(path);
    if (!node || node.isIgnored) return;
    const shouldSelect = !selectedPaths.value.has(path);
    const stack: FileNode[] = [node];
    while (stack.length > 0) {
      const current = stack.pop()!;
      if (!current.isIgnored) {
        if (shouldSelect) selectedPaths.value.add(current.path);
        else selectedPaths.value.delete(current.path);
        if (current.isDir && current.children) {
          current.children.forEach(c => {
            const childNode = nodesMap.get(c.path);
            if (childNode) stack.push(childNode);
          });
        }
      }
    }
  }

  function selectFilesByRelPaths(relPaths: string[], nodesMap: Map<string, FileNode>) {
    relPaths.forEach(relPath => {
      for (const node of nodesMap.values()) {
        if (node.relPath === relPath && !node.isDir && !node.isIgnored) {
          selectedPaths.value.add(node.path);
          break;
        }
      }
    });
  }

  function clearSelection() { selectedPaths.value.clear(); }
  function resetState() { expandedPaths.value.clear(); selectedPaths.value.clear(); activeNodePath.value = null; }

  return {
    expandedPaths, selectedPaths, activeNodePath,
    toggleExpansion, toggleSelection, toggleNodeSelection, selectFilesByRelPaths,
    expandRecursive, collapseRecursive, toggleExpansionRecursive,
    clearSelection, resetState
  };
});