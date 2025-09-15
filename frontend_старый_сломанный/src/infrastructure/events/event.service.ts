import { EventsOn } from "../../../wailsjs/runtime/runtime";
import type { NotificationsStore } from "@/stores/notifications.store";
import type { FileTreeStore } from "@/stores/file-tree.store";
import type { ContextBuilderStore } from "@/stores/context-builder.store";
import type { UiStore } from "@/stores/ui.store";
import type { TreeStateStore } from "@/stores/tree-state.store";
import type { AutonomousStore } from "@/stores/autonomous.store";

interface EventServiceDependencies {
  notificationsStore: NotificationsStore;
  fileTreeStore: FileTreeStore;
  contextBuilderStore: ContextBuilderStore;
  uiStore: UiStore;
  treeStateStore: TreeStateStore;
  autonomousStore: AutonomousStore;
}

class EventService {
  private isInitialized = false;
  private deps: EventServiceDependencies | null = null;

  public setDependencies(deps: EventServiceDependencies) {
    this.deps = deps;
  }

  public async initialize() {
    if (this.isInitialized || !this.deps) return;
    
    // Check if we're in a browser environment (development) vs Wails environment
    const isWailsEnvironment = typeof window !== 'undefined' && 
                              typeof (window as any).runtime !== 'undefined' && 
                              typeof (window as any).runtime.EventsOnMultiple !== 'undefined';
    
    // If not in Wails environment, skip initialization
    if (!isWailsEnvironment) {
      console.warn('⚠️  Not in Wails environment, skipping event service initialization');
      return;
    }
    
    // Wait for Wails runtime to be fully available with retries
    let retries = 0;
    const maxRetries = 10;
    const retryDelay = 100; // ms
    
    while (retries < maxRetries) {
      if (typeof EventsOn !== 'undefined' && typeof (window as any).runtime !== 'undefined') {
        break;
      }
      
      retries++;
      await new Promise(resolve => setTimeout(resolve, retryDelay));
    }
    
    // Check if Wails runtime is available
    if (typeof EventsOn === 'undefined') {
      console.warn('⚠️  Wails runtime not available after retries, skipping event service initialization');
      return;
    }
    
    try {
      // Test if EventsOn is actually callable
      if (typeof EventsOn !== 'function') {
        throw new Error('EventsOn is not a function');
      }
      
      this.isInitialized = true;
      console.log('✅ Wails runtime available, initializing EventService...');

      const {
        notificationsStore: notifications,
        fileTreeStore,
        contextBuilderStore,
        uiStore,
        treeStateStore,
        autonomousStore
      } = this.deps;

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
        // Instead of storing full context text, we now build context through the repository
        // The event is kept for backward compatibility but the actual context building
        // is handled through the context repository with ContextSummary approach
        notifications.addLog("Context generated successfully.", "success");
        console.log("Context generated, length:", context.length);
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
    EventsOn("ark:task:status", (status: unknown) => {
      console.log("Autonomous task status update:", status);
      autonomousStore.handleTaskStatusUpdate(status);
    });

    EventsOn("ark:task:report_updated", (report: unknown) => {
      console.log("Autonomous task report updated:", report);
      autonomousStore.handleReportUpdate(report);
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

        const selectedPathsBefore = Array.from(treeStateStore.selectedPaths);
        const expandedPathsBefore = Array.from(treeStateStore.expandedPaths);
        const activeNodePathBefore = treeStateStore.activeNodePath;

        await fileTreeStore.refreshFiles();

        const newSelected = new Set<string>();
        selectedPathsBefore.forEach((path) => {
          const node = fileTreeStore.nodesMap.get(path);
          if (node && !node.isIgnored && !node.isDir) newSelected.add(path);
        });
        treeStateStore.selectedPaths.clear();
        newSelected.forEach((p) => treeStateStore.selectedPaths.add(p));

        // Restore expanded paths softly: only those that still exist
        treeStateStore.expandedPaths.clear();
        expandedPathsBefore.forEach((path) => {
          if (fileTreeStore.nodesMap.has(path)) treeStateStore.expandedPaths.add(path);
        });

        // Restore active node if still present
        if (activeNodePathBefore && fileTreeStore.nodesMap.has(activeNodePathBefore)) {
          treeStateStore.activeNodePath = activeNodePathBefore;
        } else {
          treeStateStore.activeNodePath = null;
        }
      }, 1500);
    });

    console.log('✅ EventService initialized successfully.');
    } catch (error) {
      console.error('❌ Failed to initialize EventService:', error);
      // Don't mark as initialized if it failed
      this.isInitialized = false;
    }
  }
}

export const eventService = new EventService();