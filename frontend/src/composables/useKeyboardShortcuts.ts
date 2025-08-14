import { ref } from "vue";
import { useUiStore } from "@/stores/ui.store";
import { useContextStore } from "@/stores/context.store";
import { useGenerationStore } from "@/stores/generation.store";
import { useKeyboardState } from "./useKeyboardState";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useProjectStore } from "@/stores/project.store";

const isCtrlPressed = ref(false);
let _attached = false;
let _keydown: ((e: KeyboardEvent) => void) | null = null;
let _keyup: ((e: KeyboardEvent) => void) | null = null;

export function attachShortcuts() {
  if (_attached) return;

  const uiStore = useUiStore();
  const contextStore = useContextStore();
  const generationStore = useGenerationStore();
  const treeStateStore = useTreeStateStore();
  const projectStore = useProjectStore();
  const { isCtrlPressed: ctrlState } = useKeyboardState();
  isCtrlPressed.value = ctrlState.value;

  const shortcuts: Record<string, () => void> = {
    "ctrl+k": () => {
      const searchInput =
        (document.querySelector(
          'input[placeholder*="Filter files"]',
        ) as HTMLInputElement) ||
        (document.querySelector(
          'input[placeholder*="Поиск по файлам"]',
        ) as HTMLInputElement);
      searchInput?.focus();
    },
    escape: () => {
      uiStore.closeDrawer();
      uiStore.hideQuickLook();
      uiStore.closeContextMenu();
    },
    "ctrl+shift+c": () => {
      treeStateStore.clearSelection();
    },
    "ctrl+a": () => {
      const visibleNodes = (contextStore as any).visibleNodes;
      if (visibleNodes) {
        visibleNodes.forEach((node: any) => {
          if (!node.isDir && !node.isIgnored) {
            treeStateStore.toggleSelection(node.path);
          }
        });
      }
    },
    "ctrl+enter": () => {
      if (generationStore.canGenerate) {
        generationStore.executeGeneration();
      } else if (contextStore.selectedFiles.length > 0) {
        contextStore.buildContext();
      }
    },
    " ": () => {
      if (
        document.activeElement?.tagName === "INPUT" ||
        document.activeElement?.tagName === "TEXTAREA"
      )
        return;

      const activePath = treeStateStore.activeNodePath;
      const activeNode = activePath
        ? contextStore.nodesMap.get(activePath)
        : null;
      const rootDir = projectStore.currentProject?.path || "";
      if (activeNode && !activeNode.isDir && !activeNode.isIgnored && rootDir) {
        const fakeEvent = new MouseEvent("click", {
          clientX: window.innerWidth / 2,
          clientY: window.innerHeight / 2,
        });
        uiStore.showQuickLook({
          rootDir,
          path: activeNode.relPath,
          type: "fs",
          event: fakeEvent,
          isPinned: true,
        });
      }
    },
  };

  const handleKeydown = (event: KeyboardEvent) => {
    if (event.key === "Control" || event.key === "Meta") {
      isCtrlPressed.value = true;
    }

    const key = [
      event.ctrlKey && "ctrl",
      event.shiftKey && "shift",
      event.altKey && "alt",
      event.code === "Space" ? " " : event.key.toLowerCase(),
    ]
      .filter(Boolean)
      .join("+");

    if (shortcuts[key]) {
      event.preventDefault();
      shortcuts[key]();
    }
  };

  const handleKeyup = (event: KeyboardEvent) => {
    if (event.key === "Control" || event.key === "Meta") {
      isCtrlPressed.value = false;
      // закрываем только hover‑режим
      const ui = useUiStore();
      if (!ui.quickLook.isPinned) ui.hideQuickLook();
    }
  };

  _keydown = handleKeydown;
  _keyup = handleKeyup;
  window.addEventListener("keydown", _keydown);
  window.addEventListener("keyup", _keyup);
  _attached = true;
}

export function detachShortcuts() {
  if (!_attached) return;
  if (_keydown) window.removeEventListener("keydown", _keydown);
  if (_keyup) window.removeEventListener("keyup", _keyup);
  _keydown = null;
  _keyup = null;
  _attached = false;
}

export function useKeyboardShortcutsState() {
  return { isCtrlPressed };
}
