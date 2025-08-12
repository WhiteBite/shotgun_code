import { useContextStore } from "@/stores/context.store";

export function useFileTree() {
  const store = useContextStore();

  const toggleExpansion = (path: string) => {
    store.toggleNodeExpansion(path);
  };

  const toggleSelection = (path: string) => {
    store.toggleNodeSelection(path);
  };

  const setActiveNode = (path: string) => {
    store.setActiveNode(path);
  }

  return {
    toggleExpansion,
    toggleSelection,
    setActiveNode,
  };
}