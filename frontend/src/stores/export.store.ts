import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { createTokenEstimator } from "@/services/token-estimator.service";
import { apiService } from "@/services/api.service";
import { useUiStore } from "./ui.store";
import { useContextBuilderStore } from "./context-builder.store";
import type { ExportMode } from "@/types/enums";

export const useExportStore = defineStore("export", () => {
  const uiStore = useUiStore();
  const contextBuilderStore = useContextBuilderStore();

  const isOpen = ref(false);
  const isLoading = ref(false);

  const exportFormat = ref<"plain" | "manifest" | "json">("manifest");
  const stripComments = ref(false);
  const includeManifest = ref(true);

  const aiProfile = ref("Claude-3");

  const estimator = createTokenEstimator();
  const tokenLimit = ref(180000);
  const fileSizeLimitKB = ref(2048);

  const enableAutoSplit = ref(true);
  const maxTokensPerChunk = ref(150000);
  const overlapTokens = ref(5000);
  const splitStrategy = ref<"token" | "file" | "smart">("smart");

  const theme = ref("Dark");
  const includeLineNumbers = ref(true);
  const includePageNumbers = ref(true);

  const aiProfileHint = computed(() => {
    switch (aiProfile.value) {
      case "Claude-3": return "Optimized for Claude models with tight formatting";
      case "GPT-4o": return "Balanced formatting for GPT models";
      case "Generic": return "Compatible with most AI models";
      default: return "";
    }
  });

  const splitStrategyHint = computed(() => {
    switch (splitStrategy.value) {
      case "token": return "Split by token count only, may break files";
      case "file": return "Split keeping whole files together";
      case "smart": return "Smart splitting considering file boundaries";
      default: return "";
    }
  });

  const shouldAutoSplit = computed(() => {
    if (!enableAutoSplit.value) return false;
    const text = contextBuilderStore.shotgunContextText || "";
    const estimatedTokens = estimator.estimate(text);
    return estimatedTokens > maxTokensPerChunk.value;
  });

  function open() { isOpen.value = true; }
  function close() { isOpen.value = false; }

  async function doExportClipboard() {
    if (!contextBuilderStore.shotgunContextText) {
      uiStore.addToast("No context to export", "error");
      return;
    }
    isLoading.value = true;
    try {
      const settings = {
        mode: "clipboard" as ExportMode,
        context: contextBuilderStore.shotgunContextText,
        exportFormat: exportFormat.value,
        stripComments: stripComments.value,
        includeManifest: includeManifest.value,
      };
      const result = await apiService.exportContext(settings);
      if (result.text) {
        await navigator.clipboard.writeText(result.text);
        uiStore.addToast(`Context copied to clipboard`, "success");
      }
    } catch (error: any) {
      uiStore.addToast(`Export failed: ${error?.message || error}`, "error");
    } finally {
      isLoading.value = false;
    }
  }

  async function doExportAI() {
    if (!contextBuilderStore.shotgunContextText) {
      uiStore.addToast("No context to export", "error");
      return;
    }
    isLoading.value = true;
    try {
      const settings = {
        mode: "ai" as ExportMode,
        context: contextBuilderStore.shotgunContextText,
        tokenLimit: tokenLimit.value,
        fileSizeLimitKB: fileSizeLimitKB.value,
        aiProfile: aiProfile.value,
        enableAutoSplit: enableAutoSplit.value,
        maxTokensPerChunk: maxTokensPerChunk.value,
        overlapTokens: overlapTokens.value,
        splitStrategy: splitStrategy.value,
      };
      const result = await apiService.exportContext(settings);

      if (result.isLarge && result.filePath) {
        // Большой файл - показываем информацию о сохранении
        const sizeInMB = (result.sizeBytes / (1024 * 1024)).toFixed(1);
        uiStore.addToast(`Large file (${sizeInMB}MB) exported to temp location. Check Downloads folder.`, "success");

        // TODO: В будущем добавить автоматическое перемещение в Downloads
        // или показать диалог сохранения
        console.log('Large AI export saved to:', result.filePath);

      } else if (result.dataBase64 && result.fileName) {
        // Маленький файл - стандартная загрузка через base64
        const link = document.createElement("a");
        link.href = `data:application/octet-stream;base64,${result.dataBase64}`;
        link.download = result.fileName;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        uiStore.addToast(`AI export downloaded: ${result.fileName}`, "success");
      }
    } catch (error: any) {
      uiStore.addToast(`AI export failed: ${error?.message || error}`, "error");
    } finally {
      isLoading.value = false;
    }
  }

  async function doExportHuman() {
    if (!contextBuilderStore.shotgunContextText) {
      uiStore.addToast("No context to export", "error");
      return;
    }
    isLoading.value = true;
    try {
      const settings = {
        mode: "human" as ExportMode,
        context: contextBuilderStore.shotgunContextText,
        theme: theme.value,
        includeLineNumbers: includeLineNumbers.value,
        includePageNumbers: includePageNumbers.value,
      };
      const result = await apiService.exportContext(settings);

      if (result.isLarge && result.filePath) {
        // Большой файл
        const sizeInMB = (result.sizeBytes / (1024 * 1024)).toFixed(1);
        uiStore.addToast(`Large file (${sizeInMB}MB) exported to temp location. Check Downloads folder.`, "success");
        console.log('Large human export saved to:', result.filePath);

      } else if (result.dataBase64 && result.fileName) {
        // Маленький файл
        const link = document.createElement("a");
        link.href = `data:application/pdf;base64,${result.dataBase64}`;
        link.download = result.fileName;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        uiStore.addToast(`Human export downloaded`, "success");
      }
    } catch (error: any) {
      uiStore.addToast(`Human export failed`, "error");
    } finally {
      isLoading.value = false;
    }
  }

  return {
    isOpen, isLoading, exportFormat, stripComments, includeManifest, aiProfile,
    tokenLimit, fileSizeLimitKB, enableAutoSplit, maxTokensPerChunk,
    overlapTokens, splitStrategy, theme, includeLineNumbers, includePageNumbers,
    aiProfileHint, splitStrategyHint, shouldAutoSplit, open, close,
    doExportClipboard, doExportAI, doExportHuman,
  };
});
