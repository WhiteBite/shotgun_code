import { useContextStore } from "@/stores/context.store";
import { useTreeStateStore } from "@/stores/tree-state.store";

export function useTreeActions() {
  const contextStore = useContextStore();
  const treeStateStore = useTreeStateStore();

  function toggleNodeExpansion(path: string, isAltPressed: boolean) {
    treeStateStore.toggleExpansion(path, isAltPressed);
  }

  function toggleNodeSelection(path: string) {
    treeStateStore.toggleSelection(path);
    // Here you would recalculate parent selections if needed,
    // but for now, the computed property handles the visual state.
  }

  function setActiveNode(path: string) {
    treeStateStore.activeNodePath = path;
  }

  return {
    toggleNodeExpansion,
    toggleNodeSelection,
    setActiveNode,
  };
}
