import { useUiStore } from "@/stores/ui.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useGenerationStore } from "@/stores/generation.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useProjectStore } from "@/stores/project.store";
import { useVisibleNodes } from "./useVisibleNodes";
import type { FileNode } from "@/types/dto";

let _attached = false;
let _keydown: ((e: KeyboardEvent) => void) | null = null;
let _keyup: ((e: KeyboardEvent) => void) | null = null;

function isEditableTarget(): boolean {
  const el = document.activeElement as HTMLElement | null;
  if (!el) return false;
  const tag = el.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || (el as any).isContentEditable) return true;
  return false;
}

export function attachShortcuts() {
  if (_attached) return;

  const uiStore = useUiStore();
  const fileTreeStore = useFileTreeStore();
  const contextBuilderStore = useContextBuilderStore();
  const generationStore = useGenerationStore();
  const treeStateStore = useTreeStateStore();
  const projectStore = useProjectStore();
  const { visibleNodes } = useVisibleNodes();

  const shortcuts: Record<string, () => void> = {
    "ctrl+k": () => {
      const el = document.querySelector('input[placeholder*="Фильтр"]') as HTMLInputElement;
      el?.focus();
    },
    "escape": () => {
      uiStore.closeContextMenu();
      uiStore.hideQuickLook();
    },
    "ctrl+d": () => {
      treeStateStore.clearSelection();
    },
    "ctrl+a": () => {
      visibleNodes.value.forEach((node: FileNode) => {
        if (!node.isDir && !node.isIgnored) {
          treeStateStore.selectedPaths.add(node.path);
        }
      });
    },
    "ctrl+enter": () => {
      if (generationStore.canGenerate) {
        generationStore.executeGeneration();
      } else if (contextBuilderStore.selectedFiles.length > 0) {
        contextBuilderStore.buildContext();
      }
    },
    "space": () => {
      if (document.activeElement && ["INPUT","TEXTAREA"].includes((document.activeElement as HTMLElement).tagName)) return;
      const activePath = treeStateStore.activeNodePath;
      const rootDir = projectStore.currentProject?.path || "";
      if (!activePath || !rootDir) return;
      const node = fileTreeStore.nodesMap.get(activePath);
      if (!node || node.isDir || node.isIgnored) return;
      const fakeEvent = new MouseEvent("click", { clientX: window.innerWidth/2, clientY: window.innerHeight/2 });
      uiStore.showQuickLook({
        rootDir,
        path: node.relPath,
        type: "fs",
        event: fakeEvent,
        isPinned: true,
      });
    },
    "ctrl+arrowright": () => {
      const active = treeStateStore.activeNodePath;
      if (!active) return;
      treeStateStore.toggleExpansionRecursive(active, fileTreeStore.nodesMap, true);
    },
    "ctrl+arrowleft": () => {
      const active = treeStateStore.activeNodePath;
      if (!active) return;
      treeStateStore.toggleExpansionRecursive(active, fileTreeStore.nodesMap, false);
    },
  };

  const handleKeydown = (event: KeyboardEvent) => {
    const normKey = event.code === "Space" ? "space" : event.key.toLowerCase();
    const parts = [
      event.ctrlKey && "ctrl",
      event.shiftKey && "shift",
      event.altKey && "alt",
      normKey,
    ].filter(Boolean).join("+");

    if (isEditableTarget() && normKey !== "escape") return;

    if (shortcuts[parts]) {
      event.preventDefault();
      shortcuts[parts]();
    }
  };

  _keydown = handleKeydown;
  _keyup = () => {};
  window.addEventListener("keydown", _keydown);
  window.addEventListener("keyup", _keyup);
  _attached = true;
}

export function detachShortcuts() {
  if (!_attached) return;
  window.removeEventListener("keydown", _keydown!);
  window.removeEventListener("keyup", _keyup!);
  _keydown = null;
  _keyup = null;
  _attached = false;
}
