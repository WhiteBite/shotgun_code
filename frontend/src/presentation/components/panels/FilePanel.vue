<template>
  <aside :class="panelClasses" :style="panelStyle">
    <!-- Header -->
    <div class="panel-header">
      <h3 v-if="!isCollapsed" class="text-sm font-semibold text-white mb-2">Файлы</h3>
      <button 
        class="collapse-btn" 
        @click="toggleCollapse" 
        :title="isCollapsed ? 'Expand panel' : 'Collapse panel'"
        aria-label="Toggle panel"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>

    <!-- Content -->
    <div class="panel-content">
      <div class="panel-scrollable smooth-scrollbar">
        <FileTree />
      </div>
    </div>

    <!-- Footer -->
    <div class="panel-footer">
      <div class="p-2 bg-gray-900/50 rounded border border-gray-700">
        <h4 class="text-xs font-medium text-gray-400 mb-2">Сводка контекста</h4>
        <div class="text-xs text-gray-300 space-y-1">
          <div class="flex items-center justify-between">
            <span>Выбрано</span>
            <span>
              {{ fileTreeStore.selectedFiles.length }}
              {{ fileTreeStore.selectedFiles.length === 1 ? 'файл' : 'файлов' }}
              из {{ fileTreeStore.totalFiles }}
            </span>
          </div>
          <div class="w-full h-2 bg-gray-800 rounded overflow-hidden">
            <div
              class="h-full bg-blue-600"
              :style="{ width: Math.min(100, Math.round(selectionPercent)) + '%' }"
            ></div>
          </div>
          <div class="flex items-center justify-between">
            <span>≈ токенов / символов</span>
            <span>{{ estimatedTokens.toLocaleString() }} / {{ selectedChars.toLocaleString() }}</span>
          </div>
        </div>
      </div>

      <!-- Ignore Rules -->
      <div class="mt-3 space-y-2 pt-2 border-t border-gray-700/50">
        <div class="flex items-center justify-between">
          <h3 class="font-semibold text-xs text-gray-400">
            Правила игнорирования
          </h3>
          <button
            class="p-1 text-gray-400 hover:text-white transition-colors"
            title="Настроить правила игнорирования"
            @click="openIgnoreDrawer"
          >
            <svg
              class="w-3 h-3"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
              />
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
              />
            </svg>
          </button>
        </div>
        <div class="space-y-2 text-sm text-gray-300">
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              v-model="fileTreeStore.useGitignore"
              type="checkbox"
              class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500"
              @change="updateIgnoreRules"
            />
            Использовать .gitignore
          </label>
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              v-model="fileTreeStore.useCustomIgnore"
              type="checkbox"
              class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500"
              @change="updateIgnoreRules"
            />
            Пользовательские правила
          </label>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useUiStore } from "@/stores/ui.store";
import FileTree from "@/presentation/components/workspace/FileTree.vue";
import { createTokenEstimator } from "@/infrastructure/context/token-estimator.service";
import usePanelManager from '@/composables/usePanelManager';

// Panel management
const { panelClasses, panelStyle, isCollapsed, toggleCollapse } = usePanelManager('files', 320)

const fileTreeStore = useFileTreeStore();
const uiStore = useUiStore();
const estimator = createTokenEstimator();

const selectionPercent = computed(() => {
  const total = fileTreeStore.totalFiles || 1;
  return (fileTreeStore.selectedFiles.length / total) * 100;
});

const estimatedTokens = computed(() => {
  // Грубая оценка: суммируем имена путей как прокси пока контент не загружен
  const concat = fileTreeStore.selectedFiles.join("\n");
  return estimator.estimate(concat);
});

const selectedChars = computed(() => {
  const concat = fileTreeStore.selectedFiles.join("\n");
  return concat.length;
});

async function updateIgnoreRules() {
  await fileTreeStore.refreshFiles();
}

function openIgnoreDrawer() {
  uiStore.openDrawer("ignore");
}
</script>