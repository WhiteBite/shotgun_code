import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { useNotificationsStore } from "@/stores/notifications.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useUiStore } from "@/stores/ui.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useAutonomousStore } from "@/stores/autonomous.store";

class EventService {
  private isInitialized = false;

  public initialize() {
    if (this.isInitialized) return;
    this.isInitialized = true;

    const notifications = useNotificationsStore();
    const fileTreeStore = useFileTreeStore();
    const contextBuilderStore = useContextBuilderStore();
    const uiStore = useUiStore();
    const autonomousStore = useAutonomousStore();

    EventsOn("app:error", (errorMessage: string) => {
      notifications.addLog(errorMessage, "error");
    });

    EventsOn(
      "shotgunContextGenerationStarted",
      (data: { fileCount: number; rootDir: string }) => {
        // Use notifications instead of progress for now
        notifications.addLog(
          `Starting context generation for ${data.fileCount} files...`,
          "info",
        );
      },
    );

    EventsOn(
      "shotgunContextGenerationProgress",
      (progress: { current: number; total: number }) => {
        const percent =
          progress.total > 0 ? (progress.current / progress.total) * 100 : 0;
        notifications.addLog(
          `Building context... ${Math.round(percent)}%`,
          "info",
        );
      },
    );

    EventsOn("shotgunContextGenerated", (context: string) => {
      console.log(
        "shotgunContextGenerated event received, length:",
        context?.length,
      );
      console.log(
        "Context preview (first 200 chars):",
        context?.substring(0, 200),
      );
      try {
        contextBuilderStore.setShotgunContext(context);
        // Убираю дублирование - оставляю только одно уведомление
        notifications.addLog("Context generated successfully.", "success");
        console.log("Context generated, length:", context.length);
        console.log(
          "ContextBuilderStore currentContext updated:",
          contextBuilderStore.currentContext?.content?.substring(0, 100),
        );
      } catch (error) {
        console.error("Error in shotgunContextGenerated handler:", error);
        notifications.addLog("Error processing context: " + error, "error");
      }
    });

    // НОВОЕ: обработка ошибок генерации контекста
    EventsOn("shotgunContextGenerationFailed", (errorMessage: string) => {
      notifications.addLog(
        "Context generation failed: " + errorMessage,
        "error",
      );
      uiStore.addToast(
        "Failed to build context. Check console for details.",
        "error",
      );
    });

    // Обработка таймаута
    EventsOn("shotgunContextGenerationTimeout", () => {
      notifications.addLog(
        "Context generation timed out after 30 seconds",
        "error",
      );
      uiStore.addToast(
        "Context generation timed out. Try with fewer files.",
        "error",
      );
    });

    // Autonomous Mode Events
    EventsOn("ark:task:status", (status: any) => {
      console.log("Autonomous task status update:", status);
      autonomousStore.handleTaskStatusUpdate(status as any);
    });

    EventsOn("ark:task:report_updated", (report: any) => {
      console.log("Autonomous task report updated:", report);
      autonomousStore.handleReportUpdate(report as any);
    });

    EventsOn("ark:task:completed", (result: unknown) => {
      console.log("Autonomous task completed:", result);
      autonomousStore.handleTaskCompleted(result);
    });

    EventsOn("ark:task:failed", (error: unknown) => {
      console.log("Autonomous task failed:", error);
      autonomousStore.handleTaskFailed(error);
    });

    EventsOn("ark:task:cancelled", (taskId: string) => {
      console.log("Autonomous task cancelled:", taskId);
      autonomousStore.handleTaskCancelled(taskId);
    });

    let rescanTimer: number | null = null;
    EventsOn("projectFilesChanged", async () => {
      if (rescanTimer) clearTimeout(rescanTimer);
      rescanTimer = window.setTimeout(async () => {
        if (fileTreeStore.isLoading) return;

        const treeState = useTreeStateStore();
        const selectedPathsBefore = Array.from(treeState.selectedPaths);
        const expandedPathsBefore = Array.from(treeState.expandedPaths);
        const activeNodePathBefore = treeState.activeNodePath;

        await fileTreeStore.refreshFiles();

        const newSelected = new Set<string>();
        selectedPathsBefore.forEach((path) => {
          const node = fileTreeStore.nodesMap.get(path);
          if (node && !node.isIgnored && !node.isDir) newSelected.add(path);
        });
        treeState.selectedPaths.clear();
        newSelected.forEach((p) => treeState.selectedPaths.add(p));

        // Restore expanded paths softly: only those that still exist
        treeState.expandedPaths.clear();
        expandedPathsBefore.forEach((path) => {
          if (fileTreeStore.nodesMap.has(path)) treeState.expandedPaths.add(path);
        });

        // Restore active node if still present
        if (activeNodePathBefore && fileTreeStore.nodesMap.has(activeNodePathBefore)) {
          treeState.activeNodePath = activeNodePathBefore;
        } else {
          treeState.activeNodePath = null;
        }
      }, 1500);
    });

    console.log("EventService initialized.");
  }
}

export const eventService = new EventService();