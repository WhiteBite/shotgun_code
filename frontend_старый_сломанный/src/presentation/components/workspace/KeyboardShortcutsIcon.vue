<!-- KeyboardShortcutsIcon.vue -->
<template>
  <div class="keyboard-shortcuts-help relative">
    <button
      class="flex items-center text-gray-400 hover:text-white p-1 rounded"
      title="Keyboard shortcuts"
      @click="showHelp = !showHelp"
    >
      ⌨️
    </button>

    <div
      v-if="showHelp"
      class="absolute top-full left-0 mt-1 bg-gray-800 border border-gray-600 rounded-lg shadow-lg p-4 w-80 z-50"
    >
      <h3 class="font-semibold text-white mb-3">Keyboard Shortcuts</h3>
      <div class="space-y-2 text-xs">
        <div
          v-for="shortcut in shortcuts"
          :key="shortcut.key"
          class="flex justify-between"
        >
          <span class="text-gray-300">{{ shortcut.description }}</span>
          <kbd class="bg-gray-700 px-2 py-1 rounded text-gray-300">{{
            shortcut.key
          }}</kbd>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

const showHelp = ref(false);

const shortcuts = [
  { key: "Ctrl+K", description: "Focus search" },
  { key: "Space", description: "Quick look" },
  { key: "Ctrl+D", description: "Deselect all" },
  { key: "Ctrl+A", description: "Select all visible" },
  { key: "Ctrl+Enter", description: "Generate/Build" },
  { key: "Esc", description: "Close modal" },
  { key: "Ctrl+→", description: "Expand recursively" },
  { key: "Ctrl+←", description: "Collapse recursively" },
];
</script>
