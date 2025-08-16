import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { useFileTreeStore } from "./file-tree.store";
import { useTreeStateStore } from "./tree-state.store";
import { useNotificationsStore } from "./notifications.store";
import { apiService } from "@/services/api.service";
import { GitStatus } from "@/types/enums";

export const useContextBuilderStore = defineStore("contextBuilder", () => {
  const fileTreeStore = useFileTreeStore();
  const treeStateStore = useTreeStateStore();
  const notifications = useNotificationsStore();

  const shotgunContextText = ref("");

  const lastContextGeneration = ref<{
    selectedPaths: string[];
    timestamp: number;
    contentHash: string;
  } | null>(null);

  const selectedFiles = computed(() => {
    const selected = [];
    const sortedSelectedPaths = Array.from(treeStateStore.selectedPaths).sort();
    for (const path of sortedSelectedPaths) {
      const node = fileTreeStore.nodesMap.get(path);
      if (node && !node.isDir && !node.isIgnored) {
        selected.push(node);
      }
    }
    return selected;
  });

  const contextSummary = computed(() => {
    const fileCount = selectedFiles.value.length;
    // live characters: either current generated context OR sum of sizes of selected files
    const liveChars =
        shotgunContextText.value.length > 0
            ? shotgunContextText.value.length
            : selectedFiles.value.reduce((acc, f) => acc + (f.size || 0), 0);

    const estimatedTokens = Math.ceil(liveChars / 4);
    // Pricing based on a model like GPT-4o input (placeholder)
    const estimatedCost = (estimatedTokens / 1_000_000) * 5.0;

    return {
      files: fileCount,
      tokens: estimatedTokens,
      cost: estimatedCost,
      characters: liveChars,
    };
  });

  const contextStatus = computed(() => {
    if (!fileTreeStore.rootPath) {
      return { status: "none", message: "No project open" };
    }
    if (!lastContextGeneration.value) {
      return { status: "none", message: "No context generated yet" };
    }

    const currentPaths = Array.from(treeStateStore.selectedPaths).sort();
    const lastPaths = [...lastContextGeneration.value.selectedPaths].sort();

    if (
        currentPaths.length !== lastPaths.length ||
        !currentPaths.every((path, i) => path === lastPaths[i])
    ) {
      return {
        status: "changed",
        message: "Selection changed since last generation",
      };
    }

    const hasModifiedSelectedFiles = selectedFiles.value.some(
        (file) => file.gitStatus !== GitStatus.Unmodified,
    );
    if (hasModifiedSelectedFiles) {
      return {
        status: "changed",
        message: "Some selected files have uncommitted changes",
      };
    }

    const timeSinceGeneration = Date.now() - lastContextGeneration.value.timestamp;
    if (timeSinceGeneration > 300000) {
      // 5 minutes
      return {
        status: "stale",
        message: "Context may be outdated (last generated > 5 min ago)",
      };
    }

    return { status: "current", message: "Context is current" };
  });

  function setShotgunContext(context: string) {
    shotgunContextText.value = context;
    lastContextGeneration.value = {
      selectedPaths: Array.from(treeStateStore.selectedPaths).sort(),
      timestamp: Date.now(),
      contentHash: btoa(context).slice(0, 16),
    };
  }

  function buildContext() {
    if (selectedFiles.value.length === 0) {
      notifications.addLog("No files selected for context generation", "warn");
      return;
    }
    if (!fileTreeStore.rootPath) {
      notifications.addLog("Cannot build context without a root path.", "error");
      return;
    }

    const paths = selectedFiles.value.map((f) => f.relPath).sort();
    apiService.requestShotgunContextGeneration(fileTreeStore.rootPath, paths);
    notifications.addLog(`Building context from ${paths.length} files...`, "info");
  }

  function clearProjectData() {
    shotgunContextText.value = "";
    lastContextGeneration.value = null;
    notifications.addLog("Context builder data cleared", "info");
  }

  return {
    shotgunContextText,
    selectedFiles,
    contextSummary,
    contextStatus,
    setShotgunContext,
    buildContext,
    clearProjectData,
  };
});