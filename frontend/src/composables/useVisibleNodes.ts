import { computed } from "vue";
import { storeToRefs } from "pinia";
import { useContextStore } from "@/stores/context.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import type { FileNode } from "@/types/dto";

export function useVisibleNodes() {
  const contextStore = useContextStore();
  const treeStateStore = useTreeStateStore();

  const { nodesMap, debouncedQuery } = storeToRefs(contextStore);
  const { expandedPaths } = storeToRefs(treeStateStore);

  const visibleNodes = computed(() => {
    const result: FileNode[] = [];
    if (nodesMap.value.size === 0) return result;

    const roots = Array.from(nodesMap.value.values()).filter(
      (n) => n.depth === 0,
    );
    const query = debouncedQuery.value.toLowerCase().trim();
    const isFiltering = !!query;

    function buildTree(nodes: FileNode[]) {
      for (const node of nodes) {
        if (!isFiltering && node.isIgnored) continue;

        result.push(node);

        if (expandedPaths.value.has(node.path) && node.children) {
          const children = node.children
            .map((c) => nodesMap.value.get(c.path))
            .filter(Boolean) as FileNode[];
          buildTree(children);
        }
      }
    }

    function buildFlatList(nodes: FileNode[]) {
      for (const node of nodes) {
        if (node.name.toLowerCase().includes(query)) {
          result.push(node);
        }
        if (node.children) {
          const children = node.children
            .map((c) => nodesMap.value.get(c.path))
            .filter(Boolean) as FileNode[];
          buildFlatList(children);
        }
      }
    }

    if (isFiltering) {
      buildFlatList(roots);
      return result.sort((a, b) => a.path.localeCompare(b.path));
    } else {
      buildTree(roots);
      return result;
    }
  });

  return {
    visibleNodes,
  };
}
