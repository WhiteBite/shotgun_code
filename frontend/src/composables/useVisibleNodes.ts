import { computed } from "vue";
import { storeToRefs } from "pinia";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import type { FileNode } from "@/types/dto";

function sortNodes(nodes: FileNode[]): FileNode[] {
  return [...nodes].sort((a, b) => {
    if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
    return a.name.localeCompare(b.name);
  });
}

export function useVisibleNodes() {
  const fileTreeStore = useFileTreeStore();
  const treeStateStore = useTreeStateStore();

  const { rootNodes, nodesMap, searchQuery } = storeToRefs(fileTreeStore);
  const { expandedPaths } = storeToRefs(treeStateStore);

  const visibleNodes = computed((): FileNode[] => {
    if (rootNodes.value.length === 0) return [];

    const query = searchQuery.value.toLowerCase().trim();
    if (!query) {
      const result: FileNode[] = [];
      const build = (nodes: FileNode[]) => {
        const childrenSorted = sortNodes(nodes);
        for (const node of childrenSorted) {
          result.push(node);
          if (node.isDir && expandedPaths.value.has(node.path) && node.children) {
            const children = node.children
              .map((c) => nodesMap.value.get(c.path))
              .filter(Boolean) as FileNode[];
            build(children);
          }
        }
      };
      build(sortNodes(rootNodes.value));
      return result;
    }

    const matchedNodes = new Set<string>();
    for (const node of nodesMap.value.values()) {
      if (node.name.toLowerCase().includes(query)) {
        let current: FileNode | undefined = node;
        while (current) {
          matchedNodes.add(current.path);
          if (!current.parentPath) break;
          current = nodesMap.value.get(current.parentPath) as FileNode | undefined;
        }
      }
    }

    const result: FileNode[] = [];
    const buildFiltered = (nodes: FileNode[]) => {
      const sorted = sortNodes(nodes);
      for (const node of sorted) {
        if (matchedNodes.has(node.path)) {
          result.push(node);
          if (node.isDir && node.children) {
            const children = node.children
              .map((c) => nodesMap.value.get(c.path))
              .filter(Boolean) as FileNode[];
            buildFiltered(children);
          }
        }
      }
    };

    buildFiltered(rootNodes.value);
    return result;
  });

  return { visibleNodes };
}
