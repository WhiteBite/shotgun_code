import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { apiService } from "@/services/api.service";
import { useUiStore } from "./ui.store";
import { useContextStore } from "./context.store";

export type ExportMode = "clipboard" | "ai" | "human";
export type ClipboardFormat = "plain" | "manifest" | "json";

export const useExportStore = defineStore("export", () => {
  const ui = useUiStore();
  const ctx = useContextStore();

  const isOpen = ref(false);
  const isLoading = ref(false);

  // clipboard
  const exportFormat = ref<ClipboardFormat>("manifest");
  const stripComments = ref(true);
  const includeManifest = ref(true);

  // AI
  const aiProfile = ref<"Claude-3" | "GPT-4o" | "Generic">("Generic");
  const tokenLimit = ref<number>(180_000);
  const fileSizeLimitKB = ref<number>(20_000);

  // Human
  const theme = ref<"Dark" | "Light">("Dark");
  const includeLineNumbers = ref(true);
  const includePageNumbers = ref(true);

  function open() {
    isOpen.value = true;
  }
  function close() {
    isOpen.value = false;
  }

  function buildSettings(mode: ExportMode) {
    const base: any = { mode, context: ctx.shotgunContextText || "" };
    if (mode === "clipboard") {
      base.exportFormat = exportFormat.value;
      base.stripComments = stripComments.value;
      base.includeManifest = includeManifest.value;
    }
    if (mode === "ai") {
      base.aiProfile = aiProfile.value;
      base.tokenLimit = tokenLimit.value;
      base.fileSizeLimitKB = fileSizeLimitKB.value;
    }
    if (mode === "human") {
      base.theme = theme.value;
      base.includeLineNumbers = includeLineNumbers.value;
      base.includePageNumbers = includePageNumbers.value;
    }
    return base;
  }

  async function doExportClipboard() {
    if (!ctx.shotgunContextText) {
      ui.addToast("Context is empty. Build it first.", "info");
      return;
    }
    isLoading.value = true;
    try {
      const res = await apiService.exportContext(buildSettings("clipboard"));
      const text = res.text || "";
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
      } else {
        const ta = document.createElement("textarea");
        ta.value = text;
        ta.style.position = "fixed";
        ta.style.left = "-9999px";
        document.body.appendChild(ta);
        ta.select();
        document.execCommand("copy");
        document.body.removeChild(ta);
      }
      ui.addToast("Context copied to clipboard.", "success");
    } catch (e: any) {
      ui.addToast(`Export failed: ${e?.message || e}`, "error");
    } finally {
      isLoading.value = false;
    }
  }

  function downloadBase64(filename: string, base64: string) {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
    const blob = new Blob([bytes], {
      type: filename.toLowerCase().endsWith(".zip")
        ? "application/zip"
        : "application/pdf",
    });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    URL.revokeObjectURL(a.href);
    a.remove();
  }

  async function doExportAI() {
    if (!ctx.shotgunContextText) {
      ui.addToast("Context is empty. Build it first.", "info");
      return;
    }
    isLoading.value = true;
    try {
      const res = await apiService.exportContext(buildSettings("ai"));
      if (res.fileName && res.dataBase64) {
        downloadBase64(res.fileName, res.dataBase64);
        ui.addToast("AI PDF generated.", "success");
      } else ui.addToast("Nothing to export.", "info");
    } catch (e: any) {
      ui.addToast(`Export failed: ${e?.message || e}`, "error");
    } finally {
      isLoading.value = false;
    }
  }

  async function doExportHuman() {
    if (!ctx.shotgunContextText) {
      ui.addToast("Context is empty. Build it first.", "info");
      return;
    }
    isLoading.value = true;
    try {
      const res = await apiService.exportContext(buildSettings("human"));
      if (res.fileName && res.dataBase64) {
        downloadBase64(res.fileName, res.dataBase64);
        ui.addToast("Human PDF generated.", "success");
      } else ui.addToast("Nothing to export.", "info");
    } catch (e: any) {
      ui.addToast(`Export failed: ${e?.message || e}`, "error");
    } finally {
      isLoading.value = false;
    }
  }

  const aiProfileHint = computed(() => {
    switch (aiProfile.value) {
      case "Claude-3":
        return "Tight PDF, low overhead. Best for Claude family.";
      case "GPT-4o":
        return "Balanced PDF. Good for GPT family.";
      default:
        return "Generic compact PDF. Works with most models.";
    }
  });

  return {
    isOpen,
    isLoading,
    exportFormat,
    stripComments,
    includeManifest,
    aiProfile,
    tokenLimit,
    fileSizeLimitKB,
    aiProfileHint,
    theme,
    includeLineNumbers,
    includePageNumbers,
    open,
    close,
    doExportClipboard,
    doExportAI,
    doExportHuman,
  };
});
