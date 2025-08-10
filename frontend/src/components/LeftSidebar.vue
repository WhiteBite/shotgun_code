<template>
  <aside class="w-80 bg-gray-50 p-4 border-r border-gray-200 overflow-y-auto flex flex-col flex-shrink-0">
    <div class="mb-4">
      <button
          @click="project.selectProjectFolder"
          class="w-full px-4 py-2 mb-2 bg-blue-600 text-white font-semibold rounded-md hover:bg-blue-700 focus:outline-none"
      >
        Select Project Folder
      </button>
      <div v-if="project.projectRoot" class="text-xs text-gray-600 mb-2 break-all">
        Selected: {{ project.projectRoot }}
      </div>
    </div>

    <div v-if="project.projectRoot" class="mb-4">
      <div class="flex items-center justify-between">
        <h3 class="text-md font-semibold text-gray-700">Ignore Rules</h3>
        <button @click="ui.openSettingsModal" title="Edit all settings" class="p-1 hover:bg-gray-200 rounded text-lg">⚙️</button>
      </div>
      <label class="flex items-center text-sm text-gray-700 mt-2">
        <input type="checkbox" :checked="settings.useGitignore" @change="settings.updateUseGitignore($event.target.checked)" class="form-checkbox h-4 w-4" />
        Use .gitignore rules
      </label>
      <label class="flex items-center text-sm text-gray-700 mt-1">
        <input type="checkbox" :checked="settings.useCustomIgnore" @change="settings.updateUseCustomIgnore($event.target.checked)" class="form-checkbox h-4 w-4" />
        Use custom rules
      </label>
    </div>

    <div v-if="project.projectRoot" class="mb-4">
      <h3 class="text-md font-semibold text-gray-700 mb-2">Context Selection</h3>
      <div class="space-y-2">
        <button @click="contextAnalysis.suggestAndApplyContext" :disabled="contextAnalysis.isSuggesting || git.isLoading" class="w-full px-4 py-2 text-sm bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-wait">
          Auto-select Files
        </button>
        <button v-if="git.isAvailable" @click="git.selectUncommittedFiles(project.projectRoot)" :disabled="git.isLoading || contextAnalysis.isSuggesting" class="w-full px-4 py-2 text-sm bg-gray-200 text-gray-800 rounded-md hover:bg-gray-300 disabled:opacity-50 disabled:cursor-wait">
          Select Uncommitted
        </button>
        <button v-if="git.isAvailable" @click="showCommitHistory" :disabled="git.isLoading || contextAnalysis.isSuggesting" class="w-full px-4 py-2 text-sm bg-gray-200 text-gray-800 rounded-md hover:bg-gray-300 disabled:opacity-50 disabled:cursor-wait">
          Select by Commit
        </button>
      </div>
    </div>

    <div class="flex-grow flex flex-col min-h-0">
      <h2 class="text-lg font-semibold text-gray-700 mb-2">Project Files</h2>
      <div class="border border-gray-300 rounded bg-white text-sm overflow-auto flex-grow">
        <FileTree v-if="fileTree.tree.length" :nodes="fileTree.tree" @toggle-exclude="fileTree.toggleExcludeNode" />
        <p v-else-if="fileTree.isFileTreeLoading" class="p-2 text-xs text-gray-500">Loading tree...</p>
        <p v-else-if="project.projectRoot && !fileTree.loadingError" class="p-2 text-xs text-gray-500">Tree is empty.</p>
        <p v-else-if="!project.projectRoot" class="p-2 text-xs text-gray-500">Select a project folder.</p>
        <p v-if="fileTree.loadingError" class="p-2 text-xs text-red-500">{{ fileTree.loadingError }}</p>
      </div>
    </div>
  </aside>
</template>

<script setup>
import FileTree from './FileTree.vue';
import { useProjectStore } from '../stores/projectStore';
import { useSettingsStore } from '../stores/settingsStore.js';
import { useFileTreeStore } from '../stores/fileTreeStore';
import { useGitStore } from '../stores/gitStore';
import { useUiStore } from '../stores/uiStore';
import { useContextAnalysisStore } from '../stores/contextAnalysisStore.js';

const project = useProjectStore();
const settings = useSettingsStore();
const fileTree = useFileTreeStore();
const git = useGitStore();
const ui = useUiStore();
const contextAnalysis = useContextAnalysisStore();

async function showCommitHistory() {
  await git.fetchRichCommitHistory(project.projectRoot, '');
  ui.openCommitModal();
}
</script>