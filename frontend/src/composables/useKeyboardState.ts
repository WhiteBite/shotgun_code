import { ref, onMounted, onUnmounted, readonly } from 'vue';

const _isCtrlPressed = ref(false);
const _isAltPressed = ref(false);

const keydownHandler = (e: KeyboardEvent) => {
  if (e.key === 'Control' || e.key === 'Meta') _isCtrlPressed.value = true;
  if (e.key === 'Alt') _isAltPressed.value = true;
};
const keyupHandler = (e: KeyboardEvent) => {
  if (e.key === 'Control' || e.key === 'Meta') _isCtrlPressed.value = false;
  if (e.key === 'Alt') _isAltPressed.value = false;
};
const handleBlur = () => {
  _isCtrlPressed.value = false;
  _isAltPressed.value = false;
};

// This function is called once in main.ts to attach the listeners.
export function initializeKeyboardState() {
  onMounted(() => {
    window.addEventListener('keydown', keydownHandler);
    window.addEventListener('keyup', keyupHandler);
    window.addEventListener('blur', handleBlur);
  });
  onUnmounted(() => {
    window.removeEventListener('keydown', keydownHandler);
    window.removeEventListener('keyup', keyupHandler);
    window.removeEventListener('blur', handleBlur);
  });
}

// Any component or composable can import this to get the reactive state.
export function useKeyboardState() {
  return {
    isCtrlPressed: readonly(_isCtrlPressed),
    isAltPressed: readonly(_isAltPressed),
  };
}