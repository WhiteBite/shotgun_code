import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useContextStore } from './context.store';
import type { FileNode } from '@/types/dto';

export const useTreeStateStore = defineStore('treeState', () => {
  const expandedPaths = ref(new Set<string>());
  const selectedPaths = ref(new Set<string>());
  const activeNodePath = ref<string | null>(null);

  function toggleExpansion(path: string, recursive = false) {
    const contextStore = useContextStore();
    const node = contextStore.nodesMap.get(path);
    if (!node?.isDir) return;

    const newState = !expandedPaths.value.has(path);

    const stack: FileNode[] = [node];
    while(stack.length > 0) {
      const current = stack.pop()!;
      if (newState) {
        expandedPaths.value.add(current.path);
      } else {
        expandedPaths.value.delete(current.path);
      }

      if (recursive && current.children) {
        current.children.forEach(childRef => {
          const childNode = contextStore.nodesMap.get(childRef.path);
          if (childNode?.isDir) {
            stack.push(childNode);
          }
        });
      }
    }
  }

  function toggleSelection(path: string) {
    const contextStore = useContextStore();
    const node = contextStore.nodesMap.get(path);
    if (!node || node.isIgnored) return;

    const newSelectionState = !selectedPaths.value.has(path);

    const stack: FileNode[] = [node];
    while(stack.length > 0) {
      const current = stack.pop()!;
      if (!current.isIgnored) {
        if (newSelectionState) {
          selectedPaths.value.add(current.path);
        } else {
          selectedPaths.value.delete(current.path);
        }

        if (current.children) {
          current.children.forEach(childRef => {
            const childNode = contextStore.nodesMap.get(childRef.path);
            if (childNode) stack.push(childNode);
          });
        }
      }
    }
    // Parent state will be recalculated by a computed property in the new architecture
  }

  function clearSelection() {
    selectedPaths.value.clear();
  }

  function addSelectedPaths(paths: string[]) {
    paths.forEach(p => selectedPaths.value.add(p));
  }

  return {
    expandedPaths,
    selectedPaths,
    activeNodePath,
    toggleExpansion,
    toggleSelection,
    clearSelection,
    addSelectedPaths
  };
});