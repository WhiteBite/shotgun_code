import { useProjectStore } from "@/stores/project.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useVisibleNodes } from "./useVisibleNodes";
import type { FileNode } from "@/types/dto";

let _attached = false;
let _keydown: ((e: KeyboardEvent) => void) | null = null;

function isEditableTarget(): boolean {
  const el = document.activeElement as HTMLElement | null;
  if (!el) return false;
  const tag = el.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || el.isContentEditable)
    return true;
  return false;
}

export function attachContextOperationsShortcuts() {
  if (_attached) return;

  const projectStore = useProjectStore();
  const contextBuilderStore = useContextBuilderStore();
  const { visibleNodes } = useVisibleNodes();

  const shortcuts: Record<string, () => void> = {
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
    "ctrl+shift+g": () => {
      // Generate smart suggestions
      contextBuilderStore.generateSmartSuggestions();
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

    if (shortcuts[parts]) {
      event.preventDefault();
      shortcuts[parts]();
    }
  };

  _keydown = handleKeydown;
  window.addEventListener("keydown", _keydown);
  _attached = true;
}

export function detachContextOperationsShortcuts() {
  if (!_attached) return;
  window.removeEventListener("keydown", _keydown!);
  _keydown = null;
  _attached = false;
}