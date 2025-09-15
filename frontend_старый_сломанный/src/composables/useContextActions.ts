// Auto-generated: composable for context actions (copy/split)
import { shallowRef, watch, computed } from "vue";
import type {
  ClipboardFormat,
  CopyRequest,
  SplitSettings,
  SplitPreview,
} from "@/types/splitter";
import { useExportStore } from "@/stores/export.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useUiStore } from "@/stores/ui.store";
import { container } from "@/infrastructure/container";

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
      const exportContextUseCase = container.getExportContextUseCase();
      const preview = exportContextUseCase.createSplitPreview(text, splitSettings.value);
      splitPreview.value = preview;
    } catch (e: unknown) {
      console.error("Split preview failed:", e);
      ui.addToast(`Split preview failed: ${e instanceof Error ? e.message : String(e)}`, "error");
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

    try {
      const exportContextUseCase = container.getExportContextUseCase();
      const result = await exportContextUseCase.exportToClipboard(
        text,
        req.format,
        req.stripComments,
        splitSettings.value,
        splitPreview.value,
        req.target,
        req.chunkIndex
      );
      
      if (result?.text) {
        await navigator.clipboard.writeText(result.text);
      } else {
        // Fallback: prepare payload manually if use case doesn't return text
        const exportContextUseCase = container.getExportContextUseCase();
        const payload = exportContextUseCase.prepareContextPayload(
          text,
          splitSettings.value,
          splitPreview.value,
          req.target || 'all',
          req.chunkIndex
        );
        await navigator.clipboard.writeText(payload);
      }
      
      ui.addToast("Copied to clipboard", "success");
    } catch (e: unknown) {
      ui.addToast(`Copy failed: ${e instanceof Error ? e.message : String(e)}`, "error");
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
