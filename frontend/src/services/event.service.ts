import { EventsOn } from "../../wailsjs/runtime/runtime";
import { useNotificationsStore } from "@/stores/notifications.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useUiStore } from "@/stores/ui.store";
import { useTreeStateStore } from "@/stores/tree-state.store";

class EventService {
  private isInitialized = false;

  public initialize() {
    if (this.isInitialized) return;
    this.isInitialized = true;

    const notifications = useNotificationsStore();
    const fileTreeStore = useFileTreeStore();
    const contextBuilderStore = useContextBuilderStore();
    const uiStore = useUiStore();

    EventsOn("app:error", (errorMessage: string) => {
      notifications.addLog(errorMessage, "error");
    });

    EventsOn("shotgunContextGenerationStarted", (data: { fileCount: number; rootDir: string }) => {
      uiStore.setProgress({ isActive: true, message: `Starting context generation for ${data.fileCount} files...`, value: 0 });
    });

    EventsOn("shotgunContextGenerationProgress", (progress: { current: number; total: number }) => {
      const percent = progress.total > 0 ? (progress.current / progress.total) * 100 : 0;
      uiStore.setProgress({ isActive: true, message: "Building context...", value: percent });
    });

    EventsOn("shotgunContextGenerated", (context: string) => {
      console.log("shotgunContextGenerated event received, length:", context?.length);
      console.log("Context preview (first 200 chars):", context?.substring(0, 200));
      try {
        contextBuilderStore.setShotgunContext(context);
        uiStore.clearProgress();
        // Убираю дублирование - оставляю только одно уведомление
        notifications.addLog("Context generated successfully.", "success");
        console.log("Context generated, length:", context.length);
        console.log("ContextBuilderStore currentContext updated:", contextBuilderStore.currentContext?.content?.substring(0, 100));
      } catch (error) {
        console.error("Error in shotgunContextGenerated handler:", error);
        uiStore.clearProgress();
        notifications.addLog("Error processing context: " + error, "error");
      }
    });

    // НОВОЕ: обработка ошибок генерации контекста
    EventsOn("shotgunContextGenerationFailed", (errorMessage: string) => {
      uiStore.clearProgress();
      notifications.addLog("Context generation failed: " + errorMessage, "error");
      uiStore.addToast("Failed to build context. Check console for details.", "error");
    });

    // Обработка таймаута
    EventsOn("shotgunContextGenerationTimeout", () => {
      uiStore.clearProgress();
      notifications.addLog("Context generation timed out after 30 seconds", "error");
      uiStore.addToast("Context generation timed out. Try with fewer files.", "error");
    });

    let rescanTimer: number | null = null;
    EventsOn("projectFilesChanged", async () => {
      if (rescanTimer) clearTimeout(rescanTimer);
      rescanTimer = window.setTimeout(async () => {
        if (fileTreeStore.isLoading) return;

        const treeState = useTreeStateStore();
        const selectedPaths = Array.from(treeState.selectedPaths);
        const expandedPaths = Array.from(treeState.expandedPaths);

        await fileTreeStore.fetchFileTree();

        const newSelected = new Set<string>();
        selectedPaths.forEach((path) => {
          const node = fileTreeStore.nodesMap.get(path);
          if (node && !node.isIgnored) newSelected.add(path);
        });
        treeState.selectedPaths = newSelected;
        treeState.expandedPaths = new Set(expandedPaths);
      }, 1500);
    });

    console.log("EventService initialized.");
  }
}

export const eventService = new EventService();
