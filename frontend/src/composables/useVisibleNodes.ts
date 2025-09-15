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
    // Проверяем, что все необходимые значения существуют
    if (!rootNodes.value || !nodesMap.value || !expandedPaths.value) {
      return [];
    }

    if (rootNodes.value.length === 0) return [];

    const query = (searchQuery.value || "").toLowerCase().trim();
    if (!query) {
      // Без поиска возвращаем только корневые узлы
      return sortNodes(rootNodes.value);
    }

    // При поиске находим узлы, которые соответствуют запросу
    const matchedNodes = new Set<string>();
    for (const node of nodesMap.value.values()) {
      if (node.name.toLowerCase().includes(query)) {
        let current: FileNode | undefined = node;
        while (current) {
          matchedNodes.add(current.path);
          if (!current.parentPath) break;
          current = nodesMap.value.get(current.parentPath) as
            | FileNode
            | undefined;
        }
      }
    }

    // Возвращаем только корневые узлы, которые содержат совпадения
    return sortNodes(rootNodes.value).filter(node => 
      matchedNodes.has(node.path) || 
      (node.children && node.children.some(child => matchedNodes.has(child.path)))
    );
  });

  return { visibleNodes };
}