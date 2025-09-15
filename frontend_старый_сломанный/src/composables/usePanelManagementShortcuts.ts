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

export function attachPanelManagementShortcuts() {
  if (_attached) return;

  const workspaceStore = useWorkspaceStore();

  const shortcuts: Record<string, () => void> = {
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

export function detachPanelManagementShortcuts() {
  if (!_attached) return;
  window.removeEventListener("keydown", _keydown!);
  _keydown = null;
  _attached = false;
}