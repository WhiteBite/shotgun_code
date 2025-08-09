
import { ref, computed } from 'vue'; // ИСПРАВЛЕНО: Добавлен импорт 'computed'
import { useNotificationsStore } from '../stores/notifications';

/**
 * @typedef {'idle' | 'success' | 'error'} ClipboardStatus
 */

/**
 * Composable для работы с буфером обмена.
 * @returns {{copy: (text: string) => Promise<void>, status: import('vue').Ref<ClipboardStatus>}}
 */
export function useClipboard() {
  const notifications = useNotificationsStore();
  const status = ref('idle');

  const copy = async (textToCopy) => {
    if (!textToCopy) {
      notifications.addLog('Текст для копирования пуст.', 'warn');
      status.value = 'error';
      setTimeout(() => status.value = 'idle', 2000);
      return;
    }
    try {
      await navigator.clipboard.writeText(textToCopy);
      status.value = 'success';
      setTimeout(() => status.value = 'idle', 2000);
    } catch (err) {
      console.error('Ошибка копирования в буфер обмена:', err);
      notifications.addLog(`Ошибка копирования: ${err.message}`, 'error');
      status.value = 'error';
      setTimeout(() => status.value = 'idle', 2000);
    }
  };

  return {
    copy,
    status,
  };
}