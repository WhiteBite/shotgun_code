import { EventsOn } from "../../wailsjs/runtime/runtime";
import { useNotificationsStore } from "@/stores/notifications.store";
import { useContextStore } from "@/stores/context.store";
import { useUiStore } from "@/stores/ui.store";
import { useTreeStateStore } from "@/stores/tree-state.store";

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

    EventsOn("app:error", (errorMessage: string) => {
      notifications.addLog(errorMessage, "error");
    });

    EventsOn(
      "shotgunContextGenerationProgress",
      (progress: { current: number; total: number }) => {
        const percent =
          progress.total > 0 ? (progress.current / progress.total) * 100 : 0;
        uiStore.setProgress({
          isActive: true,
          message: "Building context...",
          value: percent,
        });
      },
    );

    EventsOn("shotgunContextGenerated", (context: string) => {
      contextStore.setShotgunContext(context);
      uiStore.clearProgress();
      notifications.addLog("Context generated successfully.", "success");
    });

    // Debounce + rate limit for noisy folders
    let rescanTimer: number | null = null;
    let lastRescanAt = 0;
    const QUIET_DELAY = 1500; // ждать тишины
    const MIN_INTERVAL = 4000; // минимум между ресканами

    EventsOn("projectFilesChanged", async (_root: string) => {
      // если слишком часто сыпется — просто сдвигаем таймер тишины
      if (rescanTimer) {
        clearTimeout(rescanTimer);
      }
      rescanTimer = window.setTimeout(async () => {
        rescanTimer = null;
        const now = Date.now();
        if (now - lastRescanAt < MIN_INTERVAL) {
          return; // слишком часто — пропускаем
        }
        // если уже идёт загрузка — не начинаем ещё один рескан
        if (contextStore.isLoading) {
          return;
        }

        try {
          const treeState = useTreeStateStore();
          const selectedPaths = Array.from(treeState.selectedPaths);
          const expandedPaths = Array.from(treeState.expandedPaths);

          await contextStore.fetchFileTree();

          const newSelected = new Set<string>();
          selectedPaths.forEach((path) => {
            const node = contextStore.nodesMap.get(path);
            if (node && !node.isIgnored) newSelected.add(path);
          });
          treeState.selectedPaths = newSelected;
          treeState.expandedPaths = new Set(expandedPaths);

          lastRescanAt = now;
        } catch (e: any) {
          notifications.addLog(
            `Auto-rescan failed: ${e?.message || e}`,
            "error",
          );
        }
      }, QUIET_DELAY);
    });

    console.log("EventService initialized.");
  }
}

export const eventService = new EventService();
