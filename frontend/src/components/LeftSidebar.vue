
<template>
  <CustomRulesModal
      :is-visible="isCustomRulesModalVisible"
      :initial-rules="settings.customIgnoreRules"
      title="Edit Custom Ignore Rules"
      ruleType="ignore"
      @save="handleSaveCustomRules"
      @cancel="isCustomRulesModalVisible = false"
  />
  <aside class="w-80 bg-gray-50 p-4 border-r border-gray-200 overflow-y-auto flex flex-col flex-shrink-0">
    <div class="mb-6">
      <button
          @click="project.selectProjectFolder"
          class="w-full px-4 py-2 mb-2 bg-blue-600 text-white font-semibold rounded-md hover:bg-blue-700 focus:outline-none"
      >
        Select Project Folder
      </button>
      <div v-if="project.projectRoot" class="text-xs text-gray-600 mb-2 break-all">
        Selected: {{ project.projectRoot }}
      </div>

      <div v-if="project.projectRoot" class="mb-2">
        <label class="flex items-center text-sm text-gray-700">
          <input
              type="checkbox"
              :checked="settings.useGitignore"
              @change="settings.updateUseGitignore($event.target.checked)"
              class="form-checkbox h-4 w-4"
          />
          Use .gitignore rules
        </label>
        <label class="flex items-center text-sm text-gray-700 mt-1">
          <input
              type="checkbox"
              :checked="settings.useCustomIgnore"
              @change="settings.updateUseCustomIgnore($event.target.checked)"
              class="form-checkbox h-4 w-4"
          />
          Use custom rules
          <button @click="isCustomRulesModalVisible = true" title="Edit custom ignore rules" class="ml-2 p-0.5 hover:bg-gray-200 rounded text-xs">⚙️</button>
        </label>
      </div>

      <h2 class="text-lg font-semibold text-gray-700 mb-2">Project Files</h2>
      <div class="border border-gray-300 rounded min-h-[200px] bg-white text-sm overflow-auto max-h-[50vh]">
        <FileTree
            v-if="project.fileTree.length"
            :nodes="project.fileTree"
            @toggle-exclude="project.toggleExcludeNode"
        />
        <p v-else-if="project.projectRoot && !project.loadingError" class="p-2 text-xs text-gray-500">Loading tree...</p>
        <p v-else-if="!project.projectRoot" class="p-2 text-xs text-gray-500">Select a project folder.</p>
        <p v-if="project.loadingError" class="p-2 text-xs text-red-500">{{ project.loadingError }}</p>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref } from 'vue';
import FileTree from './FileTree.vue';
import CustomRulesModal from './CustomRulesModal.vue';
import { useProjectStore } from '../stores/project';
import { useSettingsStore } from '../stores/settings';

const project = useProjectStore();
const settings = useSettingsStore();

const isCustomRulesModalVisible = ref(false);

async function handleSaveCustomRules(newRules) {
  await settings.saveCustomIgnoreRules(newRules);
  isCustomRulesModalVisible.value = false;
  // Координатор позаботится об обновлении дерева
}
</script>