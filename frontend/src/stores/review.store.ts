import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { FileDiff, Hunk } from '@/types/dto';

export const useReviewStore = defineStore('review', () => {
  const gitDiff = ref('');
  const selectedFiles = ref(new Set<string>());

  const parsedDiff = computed<FileDiff[]>(() => {
    if (!gitDiff.value) return [];

    const fileDiffs = gitDiff.value.split(/^diff --git /m).filter(s => s.trim() !== '');
    const result: FileDiff[] = [];

    for (const fileDiff of fileDiffs) {
      const lines = fileDiff.split('\n');
      const fileHeaderMatch = lines[0].match(/a\/(.+) b\/(.+)/);
      if (!fileHeaderMatch) continue;

      const filePath = fileHeaderMatch[2].trim();
      const hunks: Hunk[] = [];
      let currentHunk: Hunk | null = null;
      let added = 0;
      let removed = 0;

      for (let i = 1; i < lines.length; i++) {
        const line = lines[i];
        if (line.startsWith('@@')) {
          if (currentHunk) hunks.push(currentHunk);
          currentHunk = { header: line, lines: [] };
        } else if (currentHunk) {
          if (line.startsWith('+') && !line.startsWith('+++')) added++;
          if (line.startsWith('-') && !line.startsWith('---')) removed++;
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
    selectedFiles.value = new Set(parsedDiff.value.map(f => f.filePath));
  }

  return {
    gitDiff,
    selectedFiles,
    parsedDiff,
    setGeneratedDiff,
  };
});