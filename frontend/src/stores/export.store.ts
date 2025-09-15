import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { createTokenEstimator } from "@/infrastructure/context/token-estimator.service";
import { apiService } from "@/infrastructure/api/api.service";
import { useUiStore } from "./ui.store";
import { useContextBuilderStore } from "./context-builder.store";
import { useSettingsStore } from "./settings.store";
import type { ExportMode } from "@/types/enums";
import { createSplitterService } from "@/infrastructure/context/splitter.service";
import type { ClipboardFormat, CopyRequest, SplitPreview, SplitSettings } from "@/types/splitter";

export const useExportStore = defineStore("export", () => {
  const uiStore = useUiStore();
  const contextBuilderStore = useContextBuilderStore();
  const settingsStore = useSettingsStore();

  const isOpen = ref(false);
  const isLoading = ref(false);

  const exportFormat = ref<"plain" | "manifest" | "json">("manifest");
  const stripComments = ref(false);
  const includeManifest = ref(true);

  const aiProfile = ref("Claude-3");

  const estimator = createTokenEstimator();
  const splitter = createSplitterService();
  const tokenLimit = ref(180000);
  const fileSizeLimitKB = ref(2048);

  const enableAutoSplit = ref(true);
  // Use settings from settings store for splitting
  const maxTokensPerChunk = computed(() => settingsStore.settings.maxTokensPerChunk || 8000);
  const overlapTokens = computed(() => settingsStore.settings.overlapTokens || 200);
  const splitStrategy = computed(() => settingsStore.settings.splitStrategy || 'semantic');

  const theme = ref("Dark");
  const includeLineNumbers = ref(true);
  const includePageNumbers = ref(true);

  // Split preview state (centralized)
  const splitPreview = ref<SplitPreview | null>(null);

  const aiProfileHint = computed(() => {
    switch (aiProfile.value) {
      case "Claude-3":
        return "Optimized for Claude models with tight formatting";
      case "GPT-4o":
        return "Balanced formatting for GPT models";
      case "Generic":
        return "Compatible with most AI models";
      default:
        return "";
    }
  });

  const splitStrategyHint = computed(() => {
    switch (splitStrategy.value) {
      case "token":
        return "Split by token count only, may break files";
      case "file":
        return "Split keeping whole files together";
      case "smart":
        return "Smart splitting considering file boundaries";
      default:
        return "";
    }
  });

  const shouldAutoSplit = computed(() => {
    if (!enableAutoSplit.value) return false;
    const text = contextBuilderStore.currentContext?.content || "";
    const estimatedTokens = estimator.estimate(text);
    return estimatedTokens > maxTokensPerChunk.value;
  });

  function getSplitSettings(): SplitSettings {
    return {
      enableAutoSplit: enableAutoSplit.value,
      maxTokensPerChunk: maxTokensPerChunk.value,
      overlapTokens: overlapTokens.value,
      splitStrategy: splitStrategy.value,
    };
  }

  function computeSplitPreview() {
    const text = contextBuilderStore.currentContext?.content || "";
    if (!text) {
      splitPreview.value = null;
      return;
    }
    try {
      const preview = splitter.split(
        { text },
        { ...getSplitSettings(), enableAutoSplit: true },
        { estimator: (t) => estimator.estimate(t) },
      );
      splitPreview.value = preview;
    } catch (e: any) {
      uiStore.addToast(`Split preview failed: ${e?.message || e}`, "error");
      splitPreview.value = null;
    }
  }

  function open() {
    isOpen.value = true;
  }
  function close() {
    isOpen.value = false;
  }

  async function copy(req: CopyRequest) {
    const text = contextBuilderStore.currentContext?.content || "";
    if (!text) {
      uiStore.addToast("No context to copy", "error");
      return;
    }

    let payload = text;
    if (
      enableAutoSplit.value &&
      splitPreview.value &&
      splitPreview.value.chunks.length > 1
    ) {
      if (req.target === "chunk" && typeof req.chunkIndex === "number") {
        const c = splitPreview.value.chunks[req.chunkIndex];
        payload = c ? c.text : text;
      } else {
        const parts = splitPreview.value.chunks.map(
          (c, i, arr) => `=== Part ${i + 1}/${arr.length} ===\n${c.text}`,
        );
        payload = parts.join("\n\n");
      }
    }

    try {
      const result = await apiService.exportContext({
        mode: "clipboard" as ExportMode,
        context: payload,
        exportFormat: req.format as ClipboardFormat,
        stripComments: req.stripComments,
        includeManifest: req.format === "manifest",
        aiProfile: aiProfile.value,
        tokenLimit: tokenLimit.value,
        fileSizeLimitKB: fileSizeLimitKB.value,
        enableAutoSplit: false,
        maxTokensPerChunk: maxTokensPerChunk.value,
        overlapTokens: overlapTokens.value,
        splitStrategy: splitStrategy.value,
        theme: theme.value,
        includeLineNumbers: includeLineNumbers.value,
        includePageNumbers: includePageNumbers.value,
      });
      if (result?.text) {
        await navigator.clipboard.writeText(result.text);
      } else {
        await navigator.clipboard.writeText(payload);
      }
      uiStore.addToast("Copied to clipboard", "success");
    } catch (e: any) {
      uiStore.addToast(`Copy failed: ${e?.message || e}`, "error");
    }
  }

  async function doExportClipboard() {
    if (!contextBuilderStore.currentContext?.content) {
      uiStore.addToast("Нет контекста для экспорта", "error");
      return;
    }
    isLoading.value = true;
    try {
      const settings = {
        mode: "clipboard" as ExportMode,
        context: contextBuilderStore.currentContext.content,
        exportFormat: exportFormat.value,
        stripComments: stripComments.value,
        includeManifest: includeManifest.value,
        aiProfile: aiProfile.value,
        tokenLimit: tokenLimit.value,
        fileSizeLimitKB: fileSizeLimitKB.value,
        enableAutoSplit: false,
        maxTokensPerChunk: maxTokensPerChunk.value,
        overlapTokens: overlapTokens.value,
        splitStrategy: splitStrategy.value,
        theme: theme.value,
        includeLineNumbers: includeLineNumbers.value,
        includePageNumbers: includePageNumbers.value,
      };
      const result = await apiService.exportContext(settings);
      if (result.text) {
        await navigator.clipboard.writeText(result.text);
        uiStore.addToast(`Контекст скопирован в буфер обмена`, "success");
      }
    } catch (error: any) {
      uiStore.addToast(`Ошибка экспорта: ${error?.message || error}`, "error");
    } finally {
      isLoading.value = false;
    }
  }

  async function doExportAI() {
    if (!contextBuilderStore.currentContext?.content) {
      uiStore.addToast("Нет контекста для экспорта", "error");
      return;
    }
    isLoading.value = true;
    try {
      const settings = {
        mode: "ai" as ExportMode,
        context: contextBuilderStore.currentContext.content,
        tokenLimit: tokenLimit.value,
        fileSizeLimitKB: fileSizeLimitKB.value,
        aiProfile: aiProfile.value,
        enableAutoSplit: enableAutoSplit.value,
        maxTokensPerChunk: maxTokensPerChunk.value,
        overlapTokens: overlapTokens.value,
        splitStrategy: splitStrategy.value,
        exportFormat: exportFormat.value,
        stripComments: stripComments.value,
        includeManifest: includeManifest.value,
        theme: theme.value,
        includeLineNumbers: includeLineNumbers.value,
        includePageNumbers: includePageNumbers.value,
      };
      const result = await apiService.exportContext(settings);

      if (result.isLarge && result.filePath) {
        // Большой файл - показываем информацию о сохранении
        const sizeInMB = ((result.sizeBytes || 0) / (1024 * 1024)).toFixed(1);
        uiStore.addToast(
          `Большой файл (${sizeInMB}MB) экспортирован во временную папку. Проверьте папку Downloads.`,
          "success",
        );

        // TODO: В будущем добавить автоматическое перемещение в Downloads
        // или показать диалог сохранения
        console.log("Large AI export saved to:", result.filePath);
      } else if (result.dataBase64 && result.fileName) {
        // Маленький файл - стандартная загрузка через base64
        const link = document.createElement("a");
        link.href = `data:application/octet-stream;base64,${result.dataBase64}`;
        link.download = result.fileName;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        uiStore.addToast(`AI экспорт загружен: ${result.fileName}`, "success");
      }
    } catch (error: any) {
      uiStore.addToast(
        `Ошибка AI экспорта: ${error?.message || error}`,
        "error",
      );
    } finally {
      isLoading.value = false;
    }
  }

  async function doExportHuman() {
    if (!contextBuilderStore.currentContext?.content) {
      uiStore.addToast("Нет контекста для экспорта", "error");
      return;
    }
    isLoading.value = true;
    try {
      const settings = {
        mode: "human" as ExportMode,
        context: contextBuilderStore.currentContext.content,
        theme: theme.value,
        includeLineNumbers: includeLineNumbers.value,
        includePageNumbers: includePageNumbers.value,
        exportFormat: exportFormat.value,
        stripComments: stripComments.value,
        includeManifest: includeManifest.value,
        aiProfile: aiProfile.value,
        tokenLimit: tokenLimit.value,
        fileSizeLimitKB: fileSizeLimitKB.value,
        enableAutoSplit: enableAutoSplit.value,
        maxTokensPerChunk: maxTokensPerChunk.value,
        overlapTokens: overlapTokens.value,
        splitStrategy: splitStrategy.value,
      };
      const result = await apiService.exportContext(settings);

      if (result.isLarge && result.filePath) {
        // Большой файл
        const sizeInMB = ((result.sizeBytes || 0) / (1024 * 1024)).toFixed(1);
        uiStore.addToast(
          `Большой файл (${sizeInMB}MB) экспортирован во временную папку. Проверьте папку Downloads.`,
          "success",
        );
        console.log("Large human export saved to:", result.filePath);
      } else if (result.dataBase64 && result.fileName) {
        // Маленький файл
        const link = document.createElement("a");
        link.href = `data:application/pdf;base64,${result.dataBase64}`;
        link.download = result.fileName;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        uiStore.addToast(`PDF экспорт загружен`, "success");
      }
    } catch (error: any) {
      uiStore.addToast(`Ошибка PDF экспорта`, "error");
    } finally {
      isLoading.value = false;
    }
  }

  return {
    isOpen,
    isLoading,
    exportFormat,
    stripComments,
    includeManifest,
    aiProfile,
    tokenLimit,
    fileSizeLimitKB,
    enableAutoSplit,
    maxTokensPerChunk,
    overlapTokens,
    splitStrategy,
    theme,
    includeLineNumbers,
    includePageNumbers,
    aiProfileHint,
    splitStrategyHint,
    shouldAutoSplit,
    splitPreview,
    computeSplitPreview,
    open,
    close,
    copy,
    doExportClipboard,
    doExportAI,
    doExportHuman,
  };
});