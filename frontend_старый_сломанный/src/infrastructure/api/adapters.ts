// Infrastructure layer notification adapter implementing NotificationPort interface
import type { NotificationPort } from '../../application/ports';

export class ToastNotificationAdapter implements NotificationPort {
  showSuccess(message: string): void {
    // Integration with existing toast system
    this.showToast(message, 'success');
  }

  showError(message: string): void {
    this.showToast(message, 'error');
  }

  showWarning(message: string): void {
    this.showToast(message, 'warning');
  }

  showInfo(message: string): void {
    this.showToast(message, 'info');
  }

  private showToast(message: string, type: 'success' | 'error' | 'warning' | 'info'): void {
    // This would integrate with the existing toast notification system
    // For now, we'll use a simple event dispatch or direct store access
    const event = new CustomEvent('show-toast', {
      detail: { message, type }
    });
    window.dispatchEvent(event);
  }
}

export class WailsDialogAdapter {
  async showOpenDialog(options: unknown): Promise<string | null> {
    try {
      // Integration with Wails dialog system
      const result = await window.go.main.App.ShowOpenDialog(options);
      return result || null;
    } catch (error) {
      console.error('Failed to show open dialog:', error);
      return null;
    }
  }

  async showSaveDialog(options: unknown): Promise<string | null> {
    try {
      const result = await window.go.main.App.ShowSaveDialog(options);
      return result || null;
    } catch (error) {
      console.error('Failed to show save dialog:', error);
      return null;
    }
  }

  async showMessageBox(options: unknown): Promise<unknown> {
    try {
      const result = await window.go.main.App.ShowMessageBox(options);
      return result;
    } catch (error) {
      console.error('Failed to show message box:', error);
      return null;
    }
  }
}

// Extend Wails type definitions
declare global {
  interface Window {
    go: {
      main: {
        App: {
          ShowOpenDialog(options: unknown): Promise<string>;
          ShowSaveDialog(options: unknown): Promise<string>;
          ShowMessageBox(options: unknown): Promise<unknown>;
          ApiGet(endpoint: string, params: Record<string, unknown>): Promise<unknown>;
          ApiPost(endpoint: string, data: unknown): Promise<unknown>;
          ApiPut(endpoint: string, data: unknown): Promise<unknown>;
          ApiDelete(endpoint: string): Promise<unknown>;
        };
      };
    };
  }
}