import { useTreeStateStore } from "@/stores/tree-state.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useProjectStore } from "@/stores/project.store";
import { useUiStore } from "@/stores/ui.store";
import { useVisibleNodes } from "./useVisibleNodes";
import type { FileNode } from "@/types/dto";

let _attached = false;
let _keydown: ((e: KeyboardEvent) => void) | null = null;
let _keyup: ((e: KeyboardEvent) => void) | null = null;

function isEditableTarget(): boolean {
  const el = document.activeElement as HTMLElement | null;
  if (!el) return false;
  const tag = el.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || el.isContentEditable)
    return true;
  return false;
}

export function attachFileNavigationShortcuts() {
  if (_attached) return;

  const treeStateStore = useTreeStateStore();
  const fileTreeStore = useFileTreeStore();
  const projectStore = useProjectStore();
  const uiStore = useUiStore();
  const { visibleNodes } = useVisibleNodes();

  const shortcuts: Record<string, () => void> = {
    // File Navigation and Selection
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
    space: () => {
      if (
        document.activeElement &&
        ["INPUT", "TEXTAREA"].includes(
          (document.activeElement as HTMLElement).tagName,
        )
      )
        return;
      const activePath = treeStateStore.activeNodePath;
      const rootDir = projectStore.currentProject?.path || "";
      if (!activePath || !rootDir) return;
      const node = fileTreeStore.nodesMap.get(activePath);
      if (!node || node.isDir || node.isIgnored) return;
      uiStore.showQuickLook({
        rootDir,
        path: node.relPath,
        type: "fs",
        isPinned: true,
        position: { x: window.innerWidth / 2, y: window.innerHeight / 2 },
      });
    },
    // Добавляем Ctrl для QuickLook
    ctrl: () => {
      // QuickLook по Ctrl будет обрабатываться в handleKeydown
    },
    // Добавляем Alt для рекурсивного разворачивания
    alt: () => {
      // Alt будет обрабатываться в handleKeydown
    },
    "ctrl+arrowright": () => {
      const active = treeStateStore.activeNodePath;
      if (!active) return;
      treeStateStore.toggleExpansionRecursive(
        active,
        fileTreeStore.nodesMap.value,
        true,
      );
    },
    "ctrl+arrowleft": () => {
      const active = treeStateStore.activeNodePath;
      if (!active) return;
      treeStateStore.toggleExpansionRecursive(
        active,
        fileTreeStore.nodesMap.value,
        false,
      );
    },
    // Alt + стрелки для рекурсивного разворачивания
    "alt+arrowright": () => {
      const active = treeStateStore.activeNodePath;
      if (!active) return;
      treeStateStore.toggleExpansionRecursive(
        active,
        fileTreeStore.nodesMap.value,
        true,
      );
    },
    "alt+arrowleft": () => {
      const active = treeStateStore.activeNodePath;
      if (!active) return;
      treeStateStore.toggleExpansionRecursive(
        active,
        fileTreeStore.nodesMap.value,
        false,
      );
    },
  };

  const handleKeydown = (event: KeyboardEvent) => {
    const normKey = event.code === "Space" ? "space" : event.key.toLowerCase();
    const parts = [
      event.ctrlKey && "ctrl",
      event.shiftKey && "shift",
      event.altKey && "alt",
      normKey,
    ]
      .filter(Boolean)
      .join("+");

    if (isEditableTarget() && normKey !== "escape") return;

    // Обработка Ctrl для QuickLook
    if (
      event.ctrlKey &&
      !event.altKey &&
      !event.shiftKey &&
      normKey !== "ctrl"
    ) {
      const activePath = treeStateStore.activeNodePath;
      const rootDir = projectStore.currentProject?.path || "";
      if (activePath && rootDir) {
        const node = fileTreeStore.nodesMap.get(activePath);
        if (node && !node.isDir && !node.isIgnored) {
          uiStore.showQuickLook({
            rootDir,
            path: node.relPath,
            type: "fs",
            isPinned: false,
            position: { x: window.innerWidth / 2, y: window.innerHeight / 2 },
          });
        }
      }
    }

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

export function detachFileNavigationShortcuts() {
  if (!_attached) return;
  window.removeEventListener("keydown", _keydown!);
  window.removeEventListener("keyup", _keyup!);
  _keydown = null;
  _keyup = null;
  _attached = false;
}