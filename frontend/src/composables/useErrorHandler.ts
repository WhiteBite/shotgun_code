import { useUiStore } from "@/stores/ui.store";
import { useNotificationsStore } from "@/stores/notifications.store";

export function useErrorHandler() {
  const uiStore = useUiStore();
  const notifications = useNotificationsStore();

  const handleError = (error: unknown, context: string = "Unknown") => {
    const message = error instanceof Error ? error.message : String(error);

    console.error(`[${context}]`, error);
    notifications.addLog(`[${context}] ${message}`, "error");

    // Specific error handling
    if (message.includes("network") || message.includes("fetch")) {
      uiStore.addToast("Network problems. Check connection.", "error");
    } else if (message.includes("permission") || message.includes("access")) {
      uiStore.addToast("Insufficient access rights.", "error");
    } else if (
      message.includes("not found") ||
      message.includes("does not exist")
    ) {
      uiStore.addToast("File or folder not found.", "error");
    } else {
      uiStore.addToast(`Error: ${message}`, "error");
    }
  };

  const handleAsyncError = async <T>(
    operation: () => Promise<T>,
    context: string,
    fallback?: T,
  ): Promise<T | undefined> => {
    try {
      return await operation();
    } catch (error) {
      handleError(error, context);
      return fallback;
    }
  };

  return {
    handleError,
    handleAsyncError,
  };
}
