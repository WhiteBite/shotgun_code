import { ref } from 'vue';
import { ClipboardSetText } from '../../wailsjs/runtime/runtime';

/**
 * @typedef {'idle' | 'success' | 'error'} ClipboardStatus
 */

/**
 * Composable for clipboard operations. Returns the operation status.
 * @returns {{copy: (text: string) => Promise<void>, status: import('vue').Ref<ClipboardStatus>}}
 */
export function useClipboard() {
  const status = ref('idle');

  const copy = async (textToCopy) => {
    status.value = 'idle';
    if (!textToCopy) {
      console.warn('Text to copy is empty.');
      status.value = 'error';
      setTimeout(() => status.value = 'idle', 2000);
      return;
    }

    try {
      if (window.wails?.runtime?.ClipboardSetText) {
        await ClipboardSetText(textToCopy);
      } else {
        await navigator.clipboard.writeText(textToCopy);
      }
      status.value = 'success';
      setTimeout(() => status.value = 'idle', 2000);
    } catch (err) {
      console.error('Failed to copy to clipboard:', err);
      status.value = 'error';
      setTimeout(() => status.value = 'idle', 2000);
    }
  };

  return {
    copy,
    status,
  };
}