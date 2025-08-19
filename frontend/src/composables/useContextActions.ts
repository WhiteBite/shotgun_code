// Auto-generated: composable for context actions (copy/split)
import { shallowRef, watch, computed } from "vue";
import type { ClipboardFormat, CopyRequest, SplitSettings, SplitPreview } from "@/types/splitter";
import { createSplitterService } from "@/services/splitter.service";
import { useExportStore } from "@/stores/export.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useUiStore } from "@/stores/ui.store";
import { apiService } from "@/services/api.service";

const splitter = createSplitterService();

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
    if (!text) { splitPreview.value = null; return; }
    try {
      // Принудительно включаем сплит для вычисления превью
      const settings = { ...splitSettings.value, enableAutoSplit: true };
      
      // Простой split по символам
      const maxChars = 10000; // 10k символов на часть
      const overlapChars = 500; // 500 символов перекрытия
      
      if (text.length <= maxChars) {
        // Если текст короткий, не разделяем
        splitPreview.value = {
          totalTokens: text.length / 4, // Примерная оценка
          totalChars: text.length,
          chunkCount: 1,
          chunks: [{
            index: 0,
            start: 0,
            end: text.length,
            text: text,
            tokens: text.length / 4,
            chars: text.length
          }],
          warnings: []
        };
      } else {
        // Разделяем по символам
        const chunks = [];
        let start = 0;
        let index = 0;
        
        while (start < text.length) {
          const end = Math.min(start + maxChars, text.length);
          const chunkText = text.substring(start, end);
          
          chunks.push({
            index: index++,
            start: start,
            end: end,
            text: chunkText,
            tokens: chunkText.length / 4, // Примерная оценка
            chars: chunkText.length
          });
          
          start = end - overlapChars;
          if (start >= text.length) break;
        }
        
        splitPreview.value = {
          totalTokens: text.length / 4,
          totalChars: text.length,
          chunkCount: chunks.length,
          chunks: chunks,
          warnings: []
        };
      }
      
      console.log("Split preview computed:", splitPreview.value.chunks.length, "chunks");
    } catch (e: any) {
      console.error("Split preview failed:", e);
      ui.addToast(`Split preview failed: ${e?.message || e}`, "error");
      splitPreview.value = null;
    }
  }

  function refreshPreview() {
    if (debounceId) window.clearTimeout(debounceId);
    debounceId = window.setTimeout(() => { computePreviewNow(); debounceId = null; }, 200);
  }

  watch(() => [ctxStore.currentContext?.content, splitSettings.value], refreshPreview, { deep: true });

  async function copy(req: CopyRequest) {
    const text = ctxStore.currentContext?.content || "";
    if (!text) { ui.addToast("No context to copy", "error"); return; }

    let payload = text;
    if (splitSettings.value.enableAutoSplit && splitPreview.value && splitPreview.value.chunks.length > 1) {
      if (req.target === "chunk" && typeof req.chunkIndex === "number") {
        const c = splitPreview.value.chunks[req.chunkIndex];
        payload = c ? c.text : text;
      } else {
        const parts = splitPreview.value.chunks.map((c, i, arr) => `=== Part ${i+1}/${arr.length} ===\n${c.text}`);
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

  function openExportModal() { const s = useExportStore(); s.open(); }

  return {
    exportFormat, stripComments, splitSettings, splitPreview,
    refreshPreview, copy, openExportModal,
  };
}

export type UseContextActions = ReturnType<typeof useContextActions>;
