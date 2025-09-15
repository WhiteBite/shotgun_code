import { ref, readonly } from "vue";

const _isCtrlPressed = ref(false);
const _isAltPressed = ref(false);

const keydownHandler = (e: KeyboardEvent) => {
  if (e.key === "Control" || e.key === "Meta") _isCtrlPressed.value = true;
  if (e.key === "Alt") _isAltPressed.value = true;
};
const keyupHandler = (e: KeyboardEvent) => {
  if (e.key === "Control" || e.key === "Meta") _isCtrlPressed.value = false;
  if (e.key === "Alt") _isAltPressed.value = false;
};
const handleBlur = () => {
  _isCtrlPressed.value = false;
  _isAltPressed.value = false;
};

// Explicit attach/detach to be called from a component (e.g., App.vue)
let _isAttached = false;
export function attachKeyboardState() {
  if (_isAttached) return;
  window.addEventListener("keydown", keydownHandler);
  window.addEventListener("keyup", keyupHandler);
  window.addEventListener("blur", handleBlur);
  _isAttached = true;
}
export function detachKeyboardState() {
  if (!_isAttached) return;
  window.removeEventListener("keydown", keydownHandler);
  window.removeEventListener("keyup", keyupHandler);
  window.removeEventListener("blur", handleBlur);
  _isAttached = false;
}

// Any component or composable can import this to get the reactive state.
export function useKeyboardState() {
  return {
    isCtrlPressed: readonly(_isCtrlPressed),
    isAltPressed: readonly(_isAltPressed),
  };
}