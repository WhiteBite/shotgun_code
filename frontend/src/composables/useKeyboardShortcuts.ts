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
    }
  };

  const handleKeydown = (event: KeyboardEvent) => {
    const key = [
      event.ctrlKey && 'ctrl',
      event.shiftKey && 'shift',
      event.altKey && 'alt',
      event.key.toLowerCase()
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