import { computed } from "vue";
import { storeToRefs } from "pinia";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import type { FileNode } from "@/types/dto";
import { FileTreeFilteringService } from "@/domain/services/FileTreeFilteringService";
import { APP_CONFIG } from '@/config/app-config';

const filteringService = new FileTreeFilteringService();

function sortNodes(nodes: FileNode[]): FileNode[] {
  return filteringService.sortNodes(nodes);
}

export function useVisibleNodes() {
  const fileTreeStore = useFileTreeStore();
  const treeStateStore = useTreeStateStore();

  const { nodes, nodesMap, searchQuery } = storeToRefs(fileTreeStore);
  const { expandedPaths } = storeToRefs(treeStateStore);

  const visibleNodes = computed((): FileNode[] => {
    // CRITICAL: Enhanced null safety checks per memory guidance
    if (!nodes.value || !nodesMap.value || !expandedPaths.value) {
      return [];
    }

    if (nodes.value.length === 0) return [];

    // Get root nodes directly from nodes array using centralized config
    const rootNodePath = APP_CONFIG.fileTree.ROOT_NODE_PATH;
    const rootNodes = nodes.value.filter(node => 
      !node.parentPath || 
      node.parentPath === rootNodePath ||
      node.path === rootNodePath
    );
    
    if (rootNodes.length === 0) {
      // Fallback: if no parentPath logic works, return first level nodes
      // Use APP_CONFIG to determine what constitutes a root node
      const fallbackRootNodes = nodes.value.filter(node => 
        !node.parentPath || 
        node.parentPath === APP_CONFIG.fileTree.ROOT_NODE_PATH ||
        node.parentPath === rootNodePath
      );
      return sortNodes(fallbackRootNodes.length > 0 ? fallbackRootNodes : nodes.value);
    }

    const query = (searchQuery.value || "").toLowerCase().trim();
    if (!query) {
      // Without search, return root nodes
      return sortNodes(rootNodes);
    }

    // With search, find matching nodes
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

    // Return only root nodes that contain matches
    return sortNodes(rootNodes).filter(node => 
      matchedNodes.has(node.path) || 
      (node.children && node.children.some(child => matchedNodes.has(child.path)))
    );
  });

  return { visibleNodes };
}