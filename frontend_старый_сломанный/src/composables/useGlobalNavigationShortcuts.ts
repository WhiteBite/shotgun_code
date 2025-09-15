import { useUiStore } from "@/stores/ui.store";

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

export function attachGlobalNavigationShortcuts() {
  if (_attached) return;

  const uiStore = useUiStore();

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

export function detachGlobalNavigationShortcuts() {
  if (!_attached) return;
  window.removeEventListener("keydown", _keydown!);
  _keydown = null;
  _attached = false;
}