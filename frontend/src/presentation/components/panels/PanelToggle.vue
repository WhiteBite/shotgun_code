<template>
  <div
    class="flex items-center space-x-1 p-2 bg-gray-800/60 border-b border-gray-700"
  >
    <button
      v-for="panel in panels"
      :key="panel.id"
      :class="[
        'px-3 py-1 text-xs rounded transition-colors',
        activePanels.includes(panel.id)
          ? 'bg-blue-600 text-white'
          : 'bg-gray-700 text-gray-300 hover:bg-gray-600',
      ]"
      :title="panel.title"
      @click="togglePanel(panel.id)"
    >
      {{ panel.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

interface Panel {
  id: string;
  label: string;
  title: string;
}

const panels: Panel[] = [
  { id: "files", label: "Файлы", title: "Панель файлов" },
  { id: "reports", label: "Отчеты", title: "Панель отчетов" },
];

const activePanels = ref<string[]>(["files", "reports"]);

function togglePanel(panelId: string) {
  const index = activePanels.value.indexOf(panelId);
  if (index > -1) {
    activePanels.value.splice(index, 1);
  } else {
    activePanels.value.push(panelId);
  }
}

// Expose active panels for parent components
defineExpose({
  activePanels,
});
</script>