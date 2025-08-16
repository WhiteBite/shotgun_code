import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { FileDiff, Hunk } from '@/types/dto';
export const useReviewStore = defineStore('review', () => {
  const gitDiff = ref('');
  const parsedDiff = computed<FileDiff[]>(() => {
    if (!gitDiff.value) return [];
    // Parsing logic here...
    return [];
  });
  function setGeneratedDiff(diff: string) { gitDiff.value = diff; }
  return { gitDiff, parsedDiff, setGeneratedDiff };
});
