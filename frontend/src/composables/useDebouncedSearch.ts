import { ref, watch } from 'vue';

export function useDebouncedSearch(delay = 300) {
  const searchQuery = ref('');
  const debouncedQuery = ref('');
  let timeoutId: number | null = null;

  watch(searchQuery, (newValue) => {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    timeoutId = window.setTimeout(() => {
      debouncedQuery.value = newValue;
    }, delay);
  });

  return {
    searchQuery,
    debouncedQuery,
  };
}