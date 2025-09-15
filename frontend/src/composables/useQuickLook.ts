import { useUiStore } from "@/stores/ui.store";
import type { FileNode } from "@/types/dto";

export function useQuickLook() {
  const uiStore = useUiStore();

  function handleMouseEnter(
    event: MouseEvent,
    node: FileNode,
    rootDir: string,
  ) {
    if ((event.ctrlKey || event.metaKey) && !node.isDir && !node.isIgnored) {
      uiStore.showQuickLook({
        rootDir,
        path: node.relPath,
        type: "fs",
        position: { x: event.clientX, y: event.clientY },
        isPinned: false,
      });
    }
  }

  function handleMouseLeave() {
    uiStore.hideQuickLook();
  }

  function showPinnedQuickLook(
    event: MouseEvent,
    node: FileNode,
    rootDir: string,
  ) {
    if (!node.isDir && !node.isIgnored) {
      uiStore.showQuickLook({
        rootDir,
        path: node.relPath,
        type: "fs",
        position: { x: event.clientX, y: event.clientY },
        isPinned: true,
      });
    }
  }

  return { handleMouseEnter, handleMouseLeave, showPinnedQuickLook };
}