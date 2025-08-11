import { useContextStore } from "@/stores/contextStore";

let store: ReturnType<typeof useContextStore>;

export function useFileTree() {
  if (!store) {
    store = useContextStore();
  }

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