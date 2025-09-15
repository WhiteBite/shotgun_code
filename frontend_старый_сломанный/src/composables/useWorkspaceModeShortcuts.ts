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

export function attachWorkspaceModeShortcuts() {
  if (_attached) return;

  const workspaceStore = useWorkspaceStore();

  const shortcuts: Record<string, () => void> = {
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
    "f11": () => {
      // Toggle fullscreen mode
      if (document.fullscreenElement) {
        document.exitFullscreen();
      } else {
        document.documentElement.requestFullscreen();
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

export function detachWorkspaceModeShortcuts() {
  if (!_attached) return;
  window.removeEventListener("keydown", _keydown!);
  _keydown = null;
  _attached = false;
}