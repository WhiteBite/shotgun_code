import { useGenerationStore } from "@/stores/generation.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useProjectStore } from "@/stores/project.store";
import { useWorkspaceStore } from "@/stores/workspace.store";

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

export function attachGenerationExecutionShortcuts() {
  if (_attached) return;

  const generationStore = useGenerationStore();
  const contextBuilderStore = useContextBuilderStore();
  const projectStore = useProjectStore();
  const workspaceStore = useWorkspaceStore();

  const shortcuts: Record<string, () => void> = {
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

export function detachGenerationExecutionShortcuts() {
  if (!_attached) return;
  window.removeEventListener("keydown", _keydown!);
  _keydown = null;
  _attached = false;
}