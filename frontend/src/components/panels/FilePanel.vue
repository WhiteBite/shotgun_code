<!-- frontend/src/components/panels/FilePanel.vue -->
<template>
  <aside
    class="w-80 bg-gray-800/60 p-3 border-r border-gray-700 flex flex-col flex-shrink-0"
  >
    <div class="flex-shrink-0 mb-2 flex items-center gap-2">
      <input
        v-model="contextStore.searchQuery"
        type="text"
        placeholder="Filter files..."
        class="w-full px-3 py-1.5 bg-gray-900 border border-gray-600 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      <button
        @click="rescanFiles"
        class="p-2 rounded-md hover:bg-gray-700"
        title="Rescan Project Files"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="text-gray-400"
        >
          <polyline points="23 4 23 10 17 10"></polyline>
          <polyline points="1 20 1 14 7 14"></polyline>
          <path
            d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"
          ></path>
        </svg>
      </button>
    </div>

    <div
      class="flex-grow bg-gray-900/50 rounded-md border border-gray-700 overflow-hidden min-h-0"
    >
      <div v-if="contextStore.isLoading" class="p-4 text-center text-gray-400">
        Loading file tree...
      </div>
      <FileTree v-else :nodes="visibleNodes" />
    </div>

    <div class="flex-shrink-0 mt-2 space-y-3 pt-2 border-t border-gray-700/50">
      <div>
        <h3 class="font-semibold text-xs mb-2 text-gray-400">Git</h3>
        <button
          @click="gitStore.showHistory"
          class="w-full text-left p-2 text-sm bg-gray-900/50 hover:bg-gray-700/80 rounded-md"
        >
          Commit History
        </button>
      </div>
      <div>
        <h3 class="font-semibold text-xs mb-2 text-gray-400">Ignore Rules</h3>
        <div class="space-y-2 text-sm text-gray-300">
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              v-model="settingsStore.settings.useGitignore"
              @change="updateIgnoreRules"
              class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500 focus:ring-blue-500/50"
            />
            Use .gitignore
          </label>
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              v-model="settingsStore.settings.useCustomIgnore"
              @change="updateIgnoreRules"
              class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500 focus:ring-blue-500/50"
            />
            Custom Rules
            <button
              @click="uiStore.openDrawer('ignore')"
              class="text-xs text-blue-400 hover:underline"
            >
              (Edit)
            </button>
          </label>
        </div>
      </div>
      <ContextSummary />
    </div>
    <CommitHistoryModal />
  </aside>
</template>

<script setup lang="ts">
import { useContextStore } from "@/stores/context.store";
import { useSettingsStore } from "@/stores/settings.store";
import { useUiStore } from "@/stores/ui.store";
import { useGitStore } from "@/stores/git.store";
import { useVisibleNodes } from "@/composables/useVisibleNodes";
import FileTree from "@/components/workspace/FileTree.vue";
import ContextSummary from "@/components/workspace/ContextSummary.vue";
import CommitHistoryModal from "@/components/modals/CommitHistoryModal.vue";
import { useTreeStateStore } from "@/stores/tree-state.store";

const contextStore = useContextStore();
const settingsStore = useSettingsStore();
const uiStore = useUiStore();
const gitStore = useGitStore();
const treeStateStore = useTreeStateStore();
const { visibleNodes } = useVisibleNodes();

async function updateIgnoreRules() {
  await settingsStore.saveIgnoreSettings();
}

function rescanFiles() {
  contextStore.fetchFileTree();
}
</script>
