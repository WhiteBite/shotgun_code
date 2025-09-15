<template>
  <div
    class="flex items-center space-x-1 p-2 bg-gray-800/60 border-b border-gray-700"
  >
    <button
      v-for="panel in panels"
      :key="panel.id"
      :class="[
        'px-3 py-1 text-xs rounded transition-colors',
        panelToggleService.isPanelActive(panel.id)
          ? 'bg-blue-600 text-white'
          : 'bg-gray-700 text-gray-300 hover:bg-gray-600',
      ]"
      :title="panel.title"
      @click="panelToggleService.togglePanel(panel.id)"
    >
      {{ panel.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { APP_CONFIG } from '@/config/app-config';
import { usePanelToggleService } from '@/composables/usePanelToggleService';

interface Panel {
  id: string;
  label: string;
  title: string;
}

const panels: Panel[] = [
  { id: "files", label: "Файлы", title: "Панель файлов" },
  { id: "reports", label: "Отчеты", title: "Панель отчетов" },
];

const { panelToggleService } = usePanelToggleService();

// Expose active panels for parent components
defineExpose({
  activePanels: computed(() => panelToggleService.getActivePanels()),
});
</script>