
<template>
  <div class="flex flex-col h-screen bg-gray-100">
    <HorizontalStepper />
    <div class="flex flex-1 overflow-hidden">
      <LeftSidebar />
      <CentralPanel />
    </div>
    <div
        @mousedown="startResize"
        class="w-full h-2 bg-gray-300 hover:bg-gray-400 cursor-row-resize select-none"
        title="Resize console height"
    ></div>
    <BottomConsole :height="consoleHeight" />
  </div>
</template>

<script setup>
import { ref } from 'vue';
import HorizontalStepper from './HorizontalStepper.vue';
import LeftSidebar from './LeftSidebar.vue';
import CentralPanel from './CentralPanel.vue';
import BottomConsole from './BottomConsole.vue';

// Логика ресайза консоли - это чисто UI-логика, ей место здесь.
const consoleHeight = ref(150);
const isResizing = ref(false);

function startResize(event) {
  isResizing.value = true;
  document.addEventListener('mousemove', doResize);
  document.addEventListener('mouseup', stopResize);
  event.preventDefault();
}

function doResize(event) {
  if (!isResizing.value) return;
  const newHeight = window.innerHeight - event.clientY;
  const minHeight = 50;
  const maxHeight = window.innerHeight * 0.8;
  consoleHeight.value = Math.max(minHeight, Math.min(newHeight, maxHeight));
}

function stopResize() {
  isResizing.value = false;
  document.removeEventListener('mousemove', doResize);
  document.removeEventListener('mouseup', stopResize);
}
</script>

<style scoped>
.flex-1 {
  min-height: 0;
}
</style>