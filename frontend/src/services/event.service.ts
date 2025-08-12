import { EventsOn } from '../../wailsjs/runtime/runtime';
import { useNotificationsStore } from '@/stores/notifications.store';
import { useContextStore } from '@/stores/context.store';
import { useUiStore } from '@/stores/ui.store';

/**
 * A centralized service to handle all events emitted from the Go backend.
 */
class EventService {
  private isInitialized = false;

  public initialize() {
    if (this.isInitialized) {
      return;
    }
    this.isInitialized = true;

    const notifications = useNotificationsStore();
    const contextStore = useContextStore();
    const uiStore = useUiStore();

    EventsOn('app:error', (errorMessage: string) => {
      notifications.addLog(errorMessage, 'error');
    });

    EventsOn('shotgunContextGenerationProgress', (progress: { current: number; total: number }) => {
      const percent = progress.total > 0 ? (progress.current / progress.total) * 100 : 0;
      uiStore.setProgress({
        isActive: true,
        message: 'Building context...',
        value: percent,
      });
    });

    EventsOn('shotgunContextGenerated', (context: string) => {
      contextStore.setShotgunContext(context);
      uiStore.clearProgress();
      notifications.addLog('Context generated successfully.', 'success');
    });

    console.log('EventService initialized.');
  }
}

export const eventService = new EventService();