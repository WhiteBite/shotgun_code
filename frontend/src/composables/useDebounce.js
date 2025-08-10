import { ref, watch } from 'vue';

/**
 * Composable для создания debounced-версии ref.
 * @param {import('vue').Ref} sourceRef - Исходный ref, за которым нужно следить.
 * @param {number} delay - Задержка в миллисекундах.
 * @returns {import('vue').Ref} - Новый ref, значение которого обновляется с задержкой.
 */
export function useDebouncedRef(sourceRef, delay) {
  const debouncedRef = ref(sourceRef.value);
  let timeoutId = null;

  watch(sourceRef, (newValue) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => {
      debouncedRef.value = newValue;
    }, delay);
  }, { deep: true });

  return debouncedRef;
}