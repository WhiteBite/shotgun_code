<template>
  <div v-if="ui.isSettingsModalVisible" class="fixed inset-0 bg-gray-600 bg-opacity-75 z-50 flex justify-center items-center" @click.self="handleCancel">
    <div class="bg-white rounded-lg shadow-xl w-full max-w-4xl max-h-[80vh] flex flex-col">
      <header class="p-4 border-b">
        <h2 class="text-xl font-semibold text-gray-900">Settings</h2>
      </header>
      <main class="flex-grow flex overflow-hidden">
        <!-- Tabs -->
        <div class="w-1/4 border-r p-2">
          <nav class="flex flex-col space-y-1">
            <button v-for="tab in tabs" :key="tab.id" @click="activeTab = tab.id"
                    :class="[
                      'px-3 py-2 text-left text-sm font-medium rounded-md w-full',
                      activeTab === tab.id ? 'bg-blue-100 text-blue-700' : 'text-gray-600 hover:bg-gray-100'
                    ]">
              {{ tab.name }}
            </button>
          </nav>
        </div>

        <!-- Tab Content -->
        <div class="w-3/4 p-4 overflow-y-auto">
          <!-- Ignore Rules Tab -->
          <div v-if="activeTab === 'ignore'">
            <h3 class="text-lg font-medium text-gray-800 mb-2">Ignore Rules</h3>
            <p class="text-sm text-gray-500 mb-4">These rules use .gitignore pattern syntax. They are applied when "Use custom rules" is checked in the sidebar.</p>
            <textarea
                v-model="editableSettings.customIgnoreRules"
                rows="15"
                class="w-full p-2 border border-gray-300 rounded-md shadow-sm font-mono text-sm bg-gray-50"
                placeholder="Enter custom ignore patterns, one per line (e.g., *.log, node_modules/)"
            ></textarea>
          </div>

          <!-- Prompt Rules Tab -->
          <div v-if="activeTab === 'prompt'">
            <h3 class="text-lg font-medium text-gray-800 mb-2">Prompt Rules</h3>
            <p class="text-sm text-gray-500 mb-4">These rules provide specific instructions or pre-defined text for the AI. They will be included in the final prompt.</p>
            <textarea
                v-model="editableSettings.customPromptRules"
                rows="15"
                class="w-full p-2 border border-gray-300 rounded-md shadow-sm font-mono text-sm bg-gray-50"
                placeholder="Enter rules for the AI, e.g., 'Always use TypeScript.'"
            ></textarea>
          </div>

          <!-- API Keys Tab -->
          <div v-if="activeTab === 'keys'">
            <h3 class="text-lg font-medium text-gray-800 mb-2">API & Model Settings</h3>
            <div class="space-y-4">
              <div>
                <label for="ai-provider" class="block text-sm font-medium text-gray-700">AI Provider</label>
                <select
                    id="ai-provider"
                    v-model="settings.selectedProvider"
                    class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    @change="settings.setSelectedProvider($event.target.value)"
                >
                  <option value="openai">OpenAI</option>
                  <option value="gemini">Gemini</option>
                  <option value="localai">Local AI (OpenAI compatible)</option>
                </select>
              </div>

              <!-- OpenAI Settings -->
              <div v-if="settings.selectedProvider === 'openai'" class="space-y-4">
                <div>
                  <label for="model-openai" class="block text-sm font-medium text-gray-700">Model</label>
                  <select id="model-openai" v-model="settings.selectedModel" @change="settings.setSelectedModel('openai', $event.target.value)" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm">
                    <option v-for="model in settings.availableModels" :key="model" :value="model">{{ model }}</option>
                  </select>
                </div>
                <div>
                  <label for="openai-key" class="block text-sm font-medium text-gray-700">OpenAI API Key</label>
                  <input type="password" id="openai-key" v-model="editableSettings.openAIKey" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md" placeholder="sk-..."/>
                </div>
              </div>

              <!-- Gemini Settings -->
              <div v-if="settings.selectedProvider === 'gemini'" class="space-y-4">
                <div>
                  <div class="flex items-center justify-between">
                    <label for="model-gemini" class="block text-sm font-medium text-gray-700">Model</label>
                    <button @click="handleRefreshModels" :disabled="settings.isRefreshingModels" class="text-xs px-2 py-1 bg-gray-200 rounded hover:bg-gray-300 disabled:opacity-50">
                      <span v-if="!settings.isRefreshingModels">Refresh</span>
                      <span v-else>Refreshing...</span>
                    </button>
                  </div>
                  <select id="model-gemini" v-model="settings.selectedModel" @change="settings.setSelectedModel('gemini', $event.target.value)" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm" :disabled="settings.isLoadingModels || settings.isRefreshingModels">
                    <option v-if="settings.isLoadingModels">Loading...</option>
                    <option v-for="model in settings.availableModels" :key="model" :value="model">{{ model }}</option>
                  </select>
                </div>
                <div>
                  <label for="gemini-key" class="block text-sm font-medium text-gray-700">Gemini API Key</label>
                  <input type="password" id="gemini-key" v-model="editableSettings.geminiKey" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md" :class="{ 'border-red-500': hasApiKeyError }" placeholder="AIzaSy..."/>
                </div>
              </div>

              <!-- Local AI Settings -->
              <div v-if="settings.selectedProvider === 'localai'" class="space-y-4 p-3 border border-indigo-200 bg-indigo-50 rounded-md">
                <p class="text-xs text-gray-600">For servers like LM Studio, Ollama (with OpenAI compatibility), etc.</p>
                <div>
                  <label for="localai-host" class="block text-sm font-medium text-gray-700">Host URL (incl. /v1)</label>
                  <input type="text" id="localai-host" v-model="editableSettings.localAIHost" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md" placeholder="http://localhost:1234/v1" />
                </div>
                <div>
                  <label for="localai-model" class="block text-sm font-medium text-gray-700">Model Name</label>
                  <input type="text" id="localai-model" v-model="editableSettings.localAIModelName" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md" placeholder="loaded-model-name" />
                </div>
                <div>
                  <label for="localai-key" class="block text-sm font-medium text-gray-700">API Key (optional)</label>
                  <input type="password" id="localai-key" v-model="editableSettings.localAIKey" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md" placeholder="Often not required" />
                </div>
              </div>

              <div v-if="settings.refreshError" class="p-3 bg-red-50 border border-red-200 rounded-md">
                <p class="text-sm font-medium text-red-800">Refresh Failed</p>
                <p class="text-xs text-red-700 mt-1">{{ settings.refreshError }}. Please check your API key and network connection.</p>
              </div>
            </div>
          </div>

          <!-- Prompts Viewer Tab -->
          <div v-if="activeTab === 'templates'">
            <TemplatesViewer />
          </div>
        </div>
      </main>
      <footer class="p-4 border-t flex justify-end space-x-4 bg-gray-50">
        <button @click="handleCancel" class="px-4 py-2 bg-gray-200 rounded-md hover:bg-gray-300">Cancel</button>
        <button @click="handleSave" class="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">Save & Close</button>
      </footer>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue';
import { useUiStore } from '../stores/uiStore';
import { useSettingsStore } from '../stores/settingsStore';
import TemplatesViewer from './settings/TemplatesViewer.vue';

const ui = useUiStore();
const settings = useSettingsStore();

const activeTab = ref('keys');
const tabs = [
  { id: 'keys', name: 'API & Model' },
  { id: 'ignore', name: 'Ignore Rules' },
  { id: 'prompt', name: 'Prompt Rules' },
  { id: 'templates', name: 'View Prompts' },
];

const editableSettings = ref({
  customIgnoreRules: '',
  customPromptRules: '',
  openAIKey: '',
  geminiKey: '',
  localAIHost: '',
  localAIModelName: '',
  localAIKey: '',
});

watch(() => ui.isSettingsModalVisible, (isVisible) => {
  if (isVisible) {
    editableSettings.value.customIgnoreRules = settings.customIgnoreRules;
    editableSettings.value.customPromptRules = settings.customPromptRules;
    editableSettings.value.openAIKey = settings.openAIKey;
    editableSettings.value.geminiKey = settings.geminiKey;
    editableSettings.value.localAIHost = settings.localAIHost;
    editableSettings.value.localAIModelName = settings.localAIModelName;
    editableSettings.value.localAIKey = settings.localAIKey;
    settings.refreshError = '';
    activeTab.value = 'keys';
  }
});

const hasApiKeyError = computed(() => {
  const err = settings.refreshError.toLowerCase();
  return err.includes('api ключ') || err.includes('unauthenticated');
});

function handleRefreshModels() {
  const apiKey = settings.selectedProvider === 'gemini'
      ? editableSettings.value.geminiKey
      : ''; // Только для Gemini
  settings.refreshModels(apiKey);
}

async function handleSave() {
  await settings.saveCustomIgnoreRules(editableSettings.value.customIgnoreRules);
  await settings.saveCustomPromptRules(editableSettings.value.customPromptRules);
  await settings.setOpenAIKey(editableSettings.value.openAIKey);
  await settings.setGeminiKey(editableSettings.value.geminiKey);
  await settings.setLocalAIHost(editableSettings.value.localAIHost);
  await settings.setLocalAIModelName(editableSettings.value.localAIModelName);
  await settings.setLocalAIKey(editableSettings.value.localAIKey);
  ui.closeSettingsModal();
}

function handleCancel() {
  ui.closeSettingsModal();
}
</script>