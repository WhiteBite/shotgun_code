import { useUiStore } from "@/stores/ui.store";
import { useKeyboardState } from "./useKeyboardState";
import type { FileNode } from "@/types/dto";

/**
 * Обновленный composable: теперь не тянет project store.
 * Корень проекта передается вызывающей стороной.
 */
export function useQuickLook() {
  const uiStore = useUiStore();
  const { isCtrlPressed } = useKeyboardState();

  function handleMouseEnter(
    event: MouseEvent,
    node: FileNode,
    rootDir: string,
  ) {
    if (isCtrlPressed.value && !node.isDir && !node.isIgnored) {
      uiStore.showQuickLook({
        rootDir,
        path: node.relPath,
        type: "fs",
        event,
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
        event,
        isPinned: true,
      });
    }
  }

  return {
    handleMouseEnter,
    handleMouseLeave,
    showPinnedQuickLook,
  };
}
