import { useSettingsStore } from "@/stores/settings.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useUiStore } from "@/stores/ui.store";

// Debounced single rescan after multiple edits (perf)
let rescanTimer: number | null = null;
function scheduleRescan(fn: () => Promise<void>) {
  if (rescanTimer) window.clearTimeout(rescanTimer);
  rescanTimer = window.setTimeout(async () => {
    try {
      await fn();
    } finally {
      rescanTimer = null;
    }
  }, 600);
}

function normalizeRule(relPath: string, isDir: boolean): string {
  let r = (relPath || "").replace(/\\/g, "/").trim();
  if (!r) return r;
  if (isDir && !r.endsWith("/")) r = r + "/";
  return r;
}

function splitRules(raw: string): string[] {
  const lines = (raw || "").replace(/\r\n/g, "\n").split("\n");
  return lines.map((l) => l.trim()).filter(Boolean);
}

export function useIgnoreRules() {
  const settingsStore = useSettingsStore();
  const fileTreeStore = useFileTreeStore();
  const uiStore = useUiStore();

  function hasRule(rule: string): boolean {
    const lines = splitRules(settingsStore.settings.customIgnoreRules);
    return lines.includes(rule);
  }

  async function addIgnore(relPath: string, isDir: boolean) {
    const rule = normalizeRule(relPath, isDir);
    if (!rule) return;
    const lines = splitRules(settingsStore.settings.customIgnoreRules);
    if (lines.includes(rule)) {
      uiStore.addToast("Уже в правилах игнорирования", "info");
      return;
    }
    lines.push(rule);
    settingsStore.settings.customIgnoreRules = lines.join("\n") + "\n";
    await settingsStore.saveIgnoreSettings();
    scheduleRescan(async () => {
      await fileTreeStore.refreshFiles();
      uiStore.addToast("Правило добавлено в игнор", "success");
    });
  }

  async function removeIgnore(relPath: string, isDir: boolean) {
    const rule = normalizeRule(relPath, isDir);
    if (!rule) return;
    const lines = splitRules(settingsStore.settings.customIgnoreRules);
    const next = lines.filter((l) => l !== rule);
    if (next.length === lines.length) {
      uiStore.addToast("Правило не найдено в игноре", "info");
      return;
    }
    settingsStore.settings.customIgnoreRules =
      next.join("\n") + (next.length ? "\n" : "");
    await settingsStore.saveIgnoreSettings();
    scheduleRescan(async () => {
      await fileTreeStore.refreshFiles();
      uiStore.addToast("Правило убрано из игнора", "success");
    });
  }

  async function toggleIgnore(relPath: string, isDir: boolean) {
    const rule = normalizeRule(relPath, isDir);
    if (!rule) return;
    if (hasRule(rule)) {
      await removeIgnore(relPath, isDir);
    } else {
      await addIgnore(relPath, isDir);
    }
  }

  return {
    normalizeRule,
    hasRule,
    addIgnore,
    removeIgnore,
    toggleIgnore,
  };
}
