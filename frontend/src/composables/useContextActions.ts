// Auto-generated: composable for context actions (copy/split)
import { shallowRef, watch, computed } from "vue";
import type {
  ClipboardFormat,
  CopyRequest,
  SplitSettings,
  SplitPreview,
} from "@/types/splitter";
import { createSplitterService } from "@/infrastructure/context/splitter.service";
import { createTokenEstimator } from "@/infrastructure/context/token-estimator.service";
import { useExportStore } from "@/stores/export.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useUiStore } from "@/stores/ui.store";
import { apiService } from "@/infrastructure/api/api.service";

const splitter = createSplitterService();
const estimator = createTokenEstimator();

export function useContextActions() {
  const exportStore = useExportStore();
  const ctxStore = useContextBuilderStore();
  const ui = useUiStore();

  const exportFormat = computed<ClipboardFormat>({
    get: () => exportStore.exportFormat,
    set: (v) => (exportStore.exportFormat = v),
  });
  const stripComments = computed<boolean>({
    get: () => exportStore.stripComments,
    set: (v) => (exportStore.stripComments = v),
  });
  const splitSettings = computed<SplitSettings>({
    get: () => ({
      enableAutoSplit: exportStore.enableAutoSplit,
      maxTokensPerChunk: exportStore.maxTokensPerChunk,
      overlapTokens: exportStore.overlapTokens,
      splitStrategy: exportStore.splitStrategy,
    }),
    set: (v) => {
      exportStore.enableAutoSplit = v.enableAutoSplit;
      exportStore.maxTokensPerChunk = v.maxTokensPerChunk || 50000; // Уменьшаем дефолтное значение
      exportStore.overlapTokens = v.overlapTokens || 1000; // Уменьшаем дефолтное значение
      exportStore.splitStrategy = v.splitStrategy || "token"; // Простая стратегия по умолчанию
    },
  });

  const splitPreview = shallowRef<SplitPreview | null>(null);
  let debounceId: number | null = null;

  async function computePreviewNow() {
    const text = ctxStore.currentContext?.content || "";
    if (!text) {
      splitPreview.value = null;
      return;
    }
    try {
      const settings = { ...splitSettings.value, enableAutoSplit: true };
      const preview = splitter.split(
        { text },
        settings,
        { estimator: (t) => estimator.estimate(t) },
      );
      splitPreview.value = preview;
    } catch (e: any) {
      console.error("Split preview failed:", e);
      ui.addToast(`Split preview failed: ${e?.message || e}`, "error");
      splitPreview.value = null;
    }
  }

  function refreshPreview() {
    if (debounceId) window.clearTimeout(debounceId);
    debounceId = window.setTimeout(() => {
      computePreviewNow();
      debounceId = null;
    }, 200);
  }

  watch(
    () => [ctxStore.currentContext?.content, splitSettings.value],
    refreshPreview,
    { deep: true },
  );

  async function copy(req: CopyRequest) {
    const text = ctxStore.currentContext?.content || "";
    if (!text) {
      ui.addToast("No context to copy", "error");
      return;
    }

    let payload = text;
    if (
      splitSettings.value.enableAutoSplit &&
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
        mode: "clipboard",
        context: payload,
        exportFormat: req.format,
        stripComments: req.stripComments,
        includeManifest: req.format === "manifest",
        aiProfile: "Generic",
        tokenLimit: 180000,
        fileSizeLimitKB: 2048,
        enableAutoSplit: false,
        maxTokensPerChunk: 50000,
        overlapTokens: 1000,
        splitStrategy: "token",
        theme: "Dark",
        includeLineNumbers: true,
        includePageNumbers: true,
      });
      if (result?.text) {
        await navigator.clipboard.writeText(result.text);
      } else {
        await navigator.clipboard.writeText(payload);
      }
      ui.addToast("Copied to clipboard", "success");
    } catch (e: any) {
      ui.addToast(`Copy failed: ${e?.message || e}`, "error");
    }
  }

  function openExportModal() {
    const s = useExportStore();
    s.open();
  }

  return {
    exportFormat,
    stripComments,
    splitSettings,
    splitPreview,
    refreshPreview,
    copy,
    openExportModal,
  };
}

export type UseContextActions = ReturnType<typeof useContextActions>;