import { onMounted, onUnmounted } from 'vue';
import { useUiStore } from '@/stores/ui.store';
import { useContextStore } from '@/stores/context.store';
import { useGenerationStore } from '@/stores/generation.store';

export function useKeyboardShortcuts() {
  const uiStore = useUiStore();
  const contextStore = useContextStore();
  const generationStore = useGenerationStore();

  const shortcuts: Record<string, () => void> = {
    'ctrl+k': () => {
      const searchInput = document.querySelector('input[placeholder*="Filter files"]') as HTMLInputElement;
      searchInput?.focus();
    },
    'escape': () => {
      uiStore.closeDrawer();
      uiStore.hideQuickLook();
    },
    'ctrl+shift+c': () => {
      contextStore.clearSelection();
    },
    'ctrl+a': () => {
      contextStore.selectAllVisible();
    },
    'ctrl+enter': () => {
      if (generationStore.canGenerate) {
        generationStore.executeGeneration();
      } else if (contextStore.selectedFiles.length > 0) {
        contextStore.buildContext();
      }
    },
    ' ': () => {
      if(document.activeElement?.tagName === 'INPUT' || document.activeElement?.tagName === 'TEXTAREA') return;

      const activeNode = contextStore.activeNode;
      if (activeNode && !activeNode.isDir && !activeNode.isIgnored) {
        const fakeEvent = new MouseEvent('click', {
          clientX: window.innerWidth / 2 - 400,
          clientY: window.innerHeight / 2 - 300,
        });
        uiStore.showQuickLook({ path: activeNode.relPath, type: 'fs', event: fakeEvent });
      }
    }
  };

  const handleKeydown = (event: KeyboardEvent) => {
    // We use event.code for layout-independent keys like Space
    const key = [
      event.ctrlKey && 'ctrl',
      event.shiftKey && 'shift',
      event.altKey && 'alt',
      event.code === 'Space' ? ' ' : event.key.toLowerCase()
    ].filter(Boolean).join('+');

    if (shortcuts[key]) {
      event.preventDefault();
      shortcuts[key]();
    }
  };

  onMounted(() => {
    document.addEventListener('keydown', handleKeydown);
  });

  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeydown);
  });
}