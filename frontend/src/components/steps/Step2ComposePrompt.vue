
<template>
  <!-- ИСПРАВЛЕНО: Восстановлена полная верстка (template) -->
  <div class="p-4 h-full flex flex-col">
    <p class="text-gray-700 mb-4 text-center text-sm">
      Write the task for the LLM in the central column and copy the final prompt
    </p>

    <CustomRulesModal
        :is-visible="isPromptRulesModalVisible"
        :initial-rules="settings.customPromptRules"
        title="Edit Custom Prompt Rules"
        ruleType="prompt"
        @save="handleSavePromptRules"
        @cancel="isPromptRulesModalVisible = false"
    />

    <div class="flex-grow flex flex-row space-x-4 overflow-hidden">
      <div class="w-1/2 flex flex-col space-y-3 overflow-y-auto p-2 border border-gray-200 rounded-md bg-gray-50">
        <div>
          <label for="user-task-ai" class="block text-sm font-medium text-gray-700 mb-1">Your task for AI:</label>
          <textarea
              id="user-task-ai"
              v-model="prompt.userTask"
              rows="15"
              class="w-full p-2 border border-gray-300 rounded-md shadow-sm"
              placeholder="Describe what the AI should do..."
          ></textarea>
        </div>

        <div>
          <label for="rules-content" class="block text-sm font-medium text-gray-700 mb-1 flex items-center">
            Custom rules:
            <button @click="isPromptRulesModalVisible = true" title="Edit custom prompt rules" class="ml-2 p-0.5 hover:bg-gray-200 rounded text-xs">⚙️</button>
          </label>
          <textarea
              id="rules-content"
              :value="settings.customPromptRules"
              @input="settings.updateCustomPromptRules($event.target.value)"
              rows="8"
              class="w-full p-2 border border-gray-300 rounded-md shadow-sm bg-gray-100"
              placeholder="Rules for AI..."
          ></textarea>
        </div>

        <div>
          <label for="file-list-context" class="block text-sm font-medium text-gray-700 mb-1">Files to include:</label>
          <textarea
              id="file-list-context"
              :value="project.shotgunPromptContext"
              rows="20"
              readonly
              class="w-full p-2 border border-gray-300 rounded-md shadow-sm bg-gray-100 font-mono text-xs"
              placeholder="File list from Step 1 will appear here..."
              style="min-height: 150px;"
          ></textarea>
        </div>
      </div>

      <div class="w-1/2 flex flex-col overflow-y-auto p-2 border border-gray-200 rounded-md bg-white">
        <div class="flex justify-between items-center mb-2">
          <div class="flex items-center space-x-2">
            <h3 class="text-md font-medium text-gray-700">Prompt:</h3>
            <select
                v-model="prompt.selectedTemplateKey"
                class="ml-2 p-1 border border-gray-300 rounded-md text-xs"
                :disabled="prompt.isLoading"
            >
              <option v-for="(template, key) in prompt.templates" :key="key" :value="key">
                {{ template.name }}
              </option>
            </select>
          </div>
          <div class="flex items-center space-x-3">
            <span class="text-xs font-medium">
              ~{{ prompt.approximateTokens }} tokens
            </span>
            <button
                @click="copy(prompt.finalPrompt)"
                :disabled="!prompt.finalPrompt || prompt.isLoading"
                class="px-3 py-1 bg-blue-500 text-white text-xs font-semibold rounded-md disabled:bg-gray-300"
            >
              {{ copyButtonText }}
            </button>
          </div>
        </div>

        <div v-if="prompt.isLoading" class="flex-grow flex justify-center items-center">
          <p class="text-gray-500">Updating prompt...</p>
        </div>

        <textarea
            v-else
            :value="prompt.finalPrompt"
            rows="20"
            class="w-full p-2 border border-gray-300 rounded-md shadow-sm font-mono text-xs flex-grow"
            placeholder="The final prompt will be generated here..."
            style="min-height: 300px;"
        ></textarea>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useProjectStore } from '../../stores/project';
import { useSettingsStore } from '../../stores/settings';
import { usePromptStore } from '../../stores/prompt';
import { useClipboard } from '../../composables/useClipboard';
import CustomRulesModal from '../CustomRulesModal.vue';

const project = useProjectStore();
const settings = useSettingsStore();
const prompt = usePromptStore();
const { copy, status } = useClipboard();

const isPromptRulesModalVisible = ref(false);

const copyButtonText = computed(() => {
  if (status.value === 'success') return 'Copied!';
  if (status.value === 'error') return 'Failed!';
  return 'Copy All';
});

async function handleSavePromptRules(newRules) {
  await settings.saveCustomPromptRules(newRules);
  isPromptRulesModalVisible.value = false;
}
</script>