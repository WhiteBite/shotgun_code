import { useUiStore } from "@/stores/ui.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useGenerationStore } from "@/stores/generation.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useProjectStore } from "@/stores/project.store";
import { useWorkspaceStore } from "@/stores/workspace.store";
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

export function attachShortcuts() {
  if (_attached) return;

  const uiStore = useUiStore();
  const fileTreeStore = useFileTreeStore();
  const contextBuilderStore = useContextBuilderStore();
  const generationStore = useGenerationStore();
  const treeStateStore = useTreeStateStore();
  const projectStore = useProjectStore();
  const workspaceStore = useWorkspaceStore();
  const { visibleNodes } = useVisibleNodes();

  const shortcuts: Record<string, () => void> = {
    // Global Navigation
    "ctrl+k": () => {
      const el = document.querySelector(
        'input[placeholder*="Search"], input[placeholder*="Фильтр"]',
      ) as HTMLInputElement;
      el?.focus();
    },
    escape: () => {
      uiStore.closeContextMenu();
      uiStore.hideQuickLook();
      // Close any open modals or drawers
      uiStore.closeDrawer('settings');
      uiStore.closeDrawer('prompts');
      uiStore.closeDrawer('ignore');
    },
    
    // Mode Switching
    "ctrl+1": () => {
      workspaceStore.setMode('manual');
    },
    "ctrl+2": () => {
      workspaceStore.setMode('autonomous');
    },
    "ctrl+shift+m": () => {
      workspaceStore.toggleMode();
    },
    
    // Panel Management
    "ctrl+shift+1": () => {
      workspaceStore.togglePanelVisibility('contextArea');
    },
    "ctrl+shift+2": () => {
      workspaceStore.togglePanelVisibility('resultsArea');
    },
    "ctrl+shift+3": () => {
      workspaceStore.togglePanelVisibility('console');
    },
    "f11": () => {
      // Toggle fullscreen mode
      if (document.fullscreenElement) {
        document.exitFullscreen();
      } else {
        document.documentElement.requestFullscreen();
      }
    },
    
    // Context Operations
    "ctrl+b": () => {
      if (contextBuilderStore.canBuildContext) {
        const projectPath = projectStore.currentProject?.path || "";
        if (projectPath) {
          contextBuilderStore.buildContextFromSelection(projectPath);
        }
      }
    },
    "ctrl+shift+a": () => {
      // Select all files in project
      visibleNodes.value.forEach((node: FileNode) => {
        if (!node.isDir && !node.isIgnored) {
          contextBuilderStore.addSelectedFile(node.path);
        }
      });
    },
    "ctrl+shift+d": () => {
      // Clear context selection
      contextBuilderStore.clearSelectedFiles();
    },
    // Generation and Execution
    "ctrl+g": () => {
      if (workspaceStore.isManualMode) {
        if (generationStore.canGenerate) {
          generationStore.executeGeneration();
        }
      } else if (workspaceStore.isAutonomousMode) {
        // Start autonomous execution
        console.log('Start autonomous execution');
      }
    },
    "ctrl+enter": () => {
      // Context-aware execution
      if (generationStore.canGenerate) {
        generationStore.executeGeneration();
      } else if (contextBuilderStore.hasSelectedFiles) {
        const projectPath = projectStore.currentProject?.path || "";
        if (projectPath) {
          contextBuilderStore.buildContextFromSelection(projectPath);
        }
      }
    },
    "ctrl+shift+g": () => {
      // Generate smart suggestions
      contextBuilderStore.generateSmartSuggestions();
    },
    
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
        fileTreeStore.nodesMap as any,
        true,
      );
    },
    "ctrl+arrowleft": () => {
      const active = treeStateStore.activeNodePath;
      if (!active) return;
      treeStateStore.toggleExpansionRecursive(
        active,
        fileTreeStore.nodesMap as any,
        false,
      );
    },
    // Alt + стрелки для рекурсивного разворачивания
    "alt+arrowright": () => {
      const active = treeStateStore.activeNodePath;
      if (!active) return;
      treeStateStore.toggleExpansionRecursive(
        active,
        fileTreeStore.nodesMap as any,
        true,
      );
    },
    "alt+arrowleft": () => {
      const active = treeStateStore.activeNodePath;
      if (!active) return;
      treeStateStore.toggleExpansionRecursive(
        active,
        fileTreeStore.nodesMap as any,
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

export function detachShortcuts() {
  if (!_attached) return;
  window.removeEventListener("keydown", _keydown!);
  window.removeEventListener("keyup", _keyup!);
  _keydown = null;
  _keyup = null;
  _attached = false;
}