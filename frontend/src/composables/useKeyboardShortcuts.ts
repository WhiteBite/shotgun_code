import { onMounted, onUnmounted, ref } from 'vue';
import { useUiStore } from '@/stores/ui.store';
import { useContextStore } from '@/stores/context.store';
import { useGenerationStore } from '@/stores/generation.store';

const isCtrlPressed = ref(false);

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
      const visibleNodes = (contextStore as any).visibleNodes; // Workaround for direct access
      if (visibleNodes) {
        visibleNodes.forEach((node: any) => {
          if (!node.isDir && !node.isIgnored) {
            contextStore.toggleNodeSelection(node.path);
          }
        });
      }
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
          clientX: window.innerWidth / 2,
          clientY: window.innerHeight / 2,
        });
        uiStore.showQuickLook({
          path: activeNode.relPath,
          type: 'fs',
          event: fakeEvent,
          isPinned: true
        });
      }
    }
  };

  const handleKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Control' || event.key === 'Meta') {
      isCtrlPressed.value = true;
    }

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

  const handleKeyup = (event: KeyboardEvent) => {
    if (event.key === 'Control' || event.key === 'Meta') {
      isCtrlPressed.value = false;
      uiStore.hideQuickLook();
    }
  };

  onMounted(() => {
    window.addEventListener('keydown', handleKeydown);
    window.addEventListener('keyup', handleKeyup);
  });

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown);
    window.removeEventListener('keyup', handleKeyup);
  });

  return {
    isCtrlPressed,
  };
}