import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { FileDiff, Hunk } from "@/types/dto";

interface GeneratedPatch {
  id: number;
  content: string;
}

type ViewMode = "split" | "unified";
type SplitStrategy = "file" | "hunk" | "lines";

export const useReviewStore = defineStore("review", () => {
  const gitDiff = ref("");
  const selectedFiles = ref(new Set<string>());

  // UI settings
  const viewMode = ref<ViewMode>("split");
  const showLineNumbers = ref(true);
  const ignoreWhitespace = ref(false);

  // Split/export settings
  const splitStrategy = ref<SplitStrategy>("file");
  const splitLineLimit = ref(400);
  const generatedPatches = ref<GeneratedPatch[]>([]);

  const parsedDiff = computed<FileDiff[]>(() => {
    if (!gitDiff.value) return [];

    const chunks = gitDiff.value
      .split(/^diff --git /m)
      .filter((s) => s.trim() !== "");
    const result: FileDiff[] = [];

    for (const chunk of chunks) {
      const lines = chunk.split("\n");
      const fileHeaderMatch = lines[0].match(/a\/(.+) b\/(.+)/);
      if (!fileHeaderMatch) continue;

      const filePath = fileHeaderMatch[2].trim();
      const hunks: Hunk[] = [];
      let currentHunk: Hunk | null = null;
      let added = 0;
      let removed = 0;

      for (let i = 1; i < lines.length; i++) {
        const line = lines[i];
        if (line.startsWith("@@")) {
          if (currentHunk) hunks.push(currentHunk);
          currentHunk = { header: line, lines: [] };
        } else if (currentHunk) {
          if (line.startsWith("+") && !line.startsWith("+++")) added++;
          if (line.startsWith("-") && !line.startsWith("---")) removed++;
          currentHunk.lines.push(line);
        }
      }
      if (currentHunk) hunks.push(currentHunk);

      result.push({ filePath, hunks, stats: { added, removed } });
    }
    return result;
  });

  function setGeneratedDiff(diff: string) {
    gitDiff.value = diff;
    selectedFiles.value = new Set(parsedDiff.value.map((f) => f.filePath));
    generatedPatches.value = [];
  }

  function toggleFileSelection(filePath: string) {
    const s = selectedFiles.value;
    if (s.has(filePath)) s.delete(filePath);
    else s.add(filePath);
  }

  function generatePatches() {
    const diff = gitDiff.value || "";
    const patches: GeneratedPatch[] = [];
    let counter = 1;

    if (!diff.trim()) {
      generatedPatches.value = [];
      return;
    }

    if (splitStrategy.value === "file") {
      const parts = diff
        .split(/(?=^diff --git )/m)
        .filter((p) => p.trim() !== "");
      for (const part of parts) {
        patches.push({ id: counter++, content: part.trim() });
      }
    } else if (splitStrategy.value === "lines") {
      const lines = diff.split("\n");
      const limit = Math.max(50, splitLineLimit.value || 400);
      for (let i = 0; i < lines.length; i += limit) {
        const chunk = lines
          .slice(i, i + limit)
          .join("\n")
          .trim();
        if (chunk) patches.push({ id: counter++, content: chunk });
      }
    } else if (splitStrategy.value === "hunk") {
      const files = diff
        .split(/(?=^diff --git )/m)
        .filter((p) => p.trim() !== "");
      const hunkHeader = /^@@ .* @@/;
      for (const file of files) {
        const fileHeaderEnd = file.indexOf("\n@@ ");
        const header =
          fileHeaderEnd !== -1 ? file.substring(0, fileHeaderEnd + 1) : "";
        const body =
          fileHeaderEnd !== -1 ? file.substring(fileHeaderEnd + 1) : file;
        const parts = body
          .split(/(?=^@@ .* @@)/m)
          .filter((p) => p.trim() !== "");

        let current = header;
        let currentLines = current.split("\n").length;

        const limit = Math.max(50, splitLineLimit.value || 400);
        for (const h of parts) {
          const hLines = h.split("\n").length;
          if (currentLines + hLines > limit) {
            patches.push({ id: counter++, content: current.trim() });
            current = header + h;
            currentLines = current.split("\n").length;
          } else {
            current += h;
            currentLines += hLines;
          }
        }
        if (current.trim()) {
          patches.push({ id: counter++, content: current.trim() });
        }
      }
    }

    generatedPatches.value = patches;
  }

  return {
    // state
    gitDiff,
    selectedFiles,

    // ui flags
    viewMode,
    showLineNumbers,
    ignoreWhitespace,

    // splitting
    splitStrategy,
    splitLineLimit,
    generatedPatches,

    // computed
    parsedDiff,

    // actions
    setGeneratedDiff,
    toggleFileSelection,
    generatePatches,
  };
});
