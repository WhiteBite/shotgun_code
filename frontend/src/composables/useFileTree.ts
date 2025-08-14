import { useTreeStateStore } from "@/stores/tree-state.store";

export function useFileTree() {
  const tree = useTreeStateStore();

  const toggleExpansion = (path: string, recursive = false) => {
    tree.toggleExpansion(path, recursive);
  };

  const toggleSelection = (path: string) => {
    tree.toggleSelection(path);
  };

  const setActiveNode = (path: string) => {
    tree.activeNodePath = path;
  };

  return {
    toggleExpansion,
    toggleSelection,
    setActiveNode,
  };
}
