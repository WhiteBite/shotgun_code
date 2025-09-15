import { defineStore } from "pinia";
import { ref } from "vue";
import type { FileNode } from "@/types/dto";
import { useTriStateSelection } from "@/composables/useTriStateSelection";
import { createStoreWithDependencies, type StoreDependencies } from '@/stores/StoreDependencyContainer';
import { APP_CONFIG } from '@/config/app-config';

// Use centralized configuration instead of hardcoded constants
const MAX_SELECTED_PATHS = APP_CONFIG.performance.limits.MAX_SELECTED_PATHS;
const MAX_EXPANDED_PATHS = APP_CONFIG.performance.limits.MAX_EXPANDED_PATHS;
const CLEANUP_INTERVAL = APP_CONFIG.performance.limits.CLEANUP_INTERVAL;

export const useTreeStateStore = defineStore("treeState", () => {
  return createStoreWithDependencies('tree-state', (dependencies: StoreDependencies) => {
    const { 
      localStorageService, 
      selectionManagementService,
      performanceMonitoringService,
      memoryManagementPolicy,
      observabilityService 
    } = dependencies;
    
    const expandedPaths = ref(new Set<string>());
    const selectedPaths = ref(new Set<string>());
    const activeNodePath = ref<string | null>(null);
    
    // CRITICAL: Add automatic cleanup interval
    let cleanupInterval: number | null = null;
    
    // Start cleanup interval when store is created
    if (typeof window !== 'undefined' && window.setInterval) {
      try {
        cleanupInterval = window.setInterval(performAutomaticCleanup, CLEANUP_INTERVAL);
      } catch (error) {
        console.warn('Failed to set cleanup interval:', error);
        cleanupInterval = null;
      }
    }
    
    // CRITICAL: Automatic cleanup function
    function performAutomaticCleanup() {
      let cleaned = false;
      
      // Clean up selected paths if too many
      if (selectedPaths.value.size > MAX_SELECTED_PATHS) {
        const pathArray = Array.from(selectedPaths.value);
        const toKeep = pathArray.slice(-MAX_SELECTED_PATHS); // Keep the most recent
        selectedPaths.value = new Set(toKeep);
        console.warn(`Cleaned up selected paths: ${pathArray.length} -> ${toKeep.length}`);
        cleaned = true;
      }
      
      // Clean up expanded paths if too many
      if (expandedPaths.value.size > MAX_EXPANDED_PATHS) {
        const pathArray = Array.from(expandedPaths.value);
        const toKeep = pathArray.slice(-MAX_EXPANDED_PATHS); // Keep the most recent
        expandedPaths.value = new Set(toKeep);
        console.warn(`Cleaned up expanded paths: ${pathArray.length} -> ${toKeep.length}`);
        cleaned = true;
      }
      
      // Force garbage collection if we cleaned something
      if (cleaned && window.gc) {
        try {
          window.gc();
          console.log('Garbage collection triggered after Set cleanup');
        } catch (_e) {
          console.warn('Failed to trigger garbage collection after cleanup');
        }
      }
    }

    function toggleExpansion(path: string) {
      if (expandedPaths.value.has(path)) {
        expandedPaths.value.delete(path);
      } else {
        // CRITICAL: Check limits before adding
        if (expandedPaths.value.size >= MAX_EXPANDED_PATHS) {
          console.warn('Expanded paths limit reached, cleaning up oldest entries');
          performAutomaticCleanup();
        }
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
            n.children.forEach((c) => {
              const child = nodesMap.get(c.path);
              if (child) stack.push(child);
            });
          }
        }
      }
    }

    function collapseRecursive(
      rootPath: string,
      nodesMap: Map<string, FileNode>,
    ) {
      const root = nodesMap.get(rootPath);
      if (!root || !root.isDir) return;
      const stack: FileNode[] = [root];
      while (stack.length) {
        const n = stack.pop()!;
        expandedPaths.value.delete(n.path);
        if (n.children) {
          n.children.forEach((c) => {
            const child = nodesMap.get(c.path);
            if (child) stack.push(child);
          });
        }
      }
    }

    function toggleExpansionRecursive(
      path: string,
      nodesMap: Map<string, FileNode>,
      expand: boolean,
    ) {
      if (expand) expandRecursive(path, nodesMap);
      else collapseRecursive(path, nodesMap);
    }

    function toggleSelection(path: string) {
      // CRITICAL: Create new Set for proper reactivity
      const newSelectedPaths = new Set(selectedPaths.value);
      
      if (newSelectedPaths.has(path)) {
        newSelectedPaths.delete(path);
      } else {
        // CRITICAL: Check limits before adding
        if (newSelectedPaths.size >= MAX_SELECTED_PATHS) {
          console.warn('Selected paths limit reached, cleaning up oldest entries');
          performAutomaticCleanup();
        }
        newSelectedPaths.add(path);
      }
      
      // Trigger reactivity by replacing the entire Set
      selectedPaths.value = newSelectedPaths;
    }

    function getParentPath(
      path: string,
      nodesMap: Map<string, FileNode>,
    ): string | null {
      for (const node of nodesMap.values()) {
        if (node.children?.some((child) => child.path === path)) {
          return node.path;
        }
      }
      return null;
    }

    function updateParentSelection(path: string, nodesMap: Map<string, FileNode>) {
      let parentPath = getParentPath(path, nodesMap);
      const newSelectedPaths = new Set(selectedPaths.value);
      
      while (parentPath) {
        const parentNode = nodesMap.get(parentPath);
        if (!parentNode || !parentNode.children) break;

        const allChildrenSelected = parentNode.children
          .filter((child) => !nodesMap.get(child.path)?.isIgnored)
          .every((child) => newSelectedPaths.has(child.path));

        if (allChildrenSelected) {
          newSelectedPaths.add(parentPath);
        } else {
          newSelectedPaths.delete(parentPath);
        }
        parentPath = getParentPath(parentPath, nodesMap);
      }
      
      // Trigger reactivity by replacing the entire Set
      selectedPaths.value = newSelectedPaths;
    }

    function toggleNodeSelection(path: string, nodesMap: Map<string, FileNode>) {
      const node = nodesMap.get(path);
      if (!node || node.isIgnored) return;
      
      // CRITICAL: Create a copy of selectedPaths for proper reactivity
      const currentSelectedPaths = new Set(selectedPaths.value);
      const { toggleCascade } = useTriStateSelection(nodesMap, currentSelectedPaths);
      const result = toggleCascade(node);
      
      // Handle cascade selection errors
      if (!result.success && result.error) {
        console.warn('Cascade selection failed:', result.error);
        
        // Try to show user notification if available
        try {
          // Dynamic import to avoid circular dependencies
          import('@/stores/ui.store').then(({ useUiStore }) => {
            const uiStore = useUiStore();
            uiStore.addToast(result.error || 'Selection failed', 'warning', 5000);
          }).catch(() => {
            // Fallback to alert if ui store unavailable
            alert(result.error || 'Selection failed');
          });
        } catch (_e) {
          // Fallback to alert if toast system unavailable
          alert(result.error || 'Selection failed');
        }
        
        return; // Don't proceed with parent selection update
      }
      
      // CRITICAL: Trigger reactivity by replacing the entire Set
      if (result.success) {
        selectedPaths.value = currentSelectedPaths;
        updateParentSelection(path, nodesMap);
        console.log(`Cascade selection successful: ${result.affectedCount} files affected`);
      }
    }

    function selectFilesByRelPaths(
      relPaths: string[],
      nodesRelMap: Map<string, FileNode>,
    ) {
      relPaths.forEach((relPath) => {
        const node = nodesRelMap.get(relPath);
        if (node && !node.isDir && !node.isIgnored) {
          selectedPaths.value.add(node.path);
        }
      });
    }

    function clearSelection() {
      selectedPaths.value.clear();
    }
    
    function resetState() {
      expandedPaths.value.clear();
      selectedPaths.value.clear();
      activeNodePath.value = null;
      
      // CRITICAL: Clear cleanup interval and force GC
      if (cleanupInterval) {
        clearInterval(cleanupInterval);
        cleanupInterval = null;
      }
      
      if (window.gc) {
        try {
          window.gc();
          console.log('Garbage collection triggered after state reset');
        } catch (_e) {
          console.warn('Failed to trigger garbage collection after reset');
        }
      }
    }

    return {
      expandedPaths,
      selectedPaths,
      activeNodePath,
      toggleExpansion,
      toggleSelection,
      toggleNodeSelection,
      selectFilesByRelPaths,
      expandRecursive,
      collapseRecursive,
      toggleExpansionRecursive,
      clearSelection,
      resetState,
      performAutomaticCleanup, // CRITICAL: Expose cleanup function
    };
  }); // Close dependency injection wrapper
});