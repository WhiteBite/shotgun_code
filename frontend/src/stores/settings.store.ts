import { defineStore } from 'pinia';
import { ref } from 'vue';
import { GetSettings, SaveSettings, RefreshAIModels } from '../../wailsjs/go/main/App';
import type { SettingsDTO } from '@/types/dto';
import { useUiStore } from './ui.store';
import { useErrorHandler } from '@/composables/useErrorHandler';

const emptySettings: SettingsDTO = {
  customIgnoreRules: '',
  customPromptRules: '',
  openAIAPIKey: '',
  geminiAPIKey: '',
  openRouterAPIKey: '',
  localAIAPIKey: '',
  localAIHost: '',
  localAIModelName: '',
  selectedProvider: 'openai',
  selectedModels: {},
  availableModels: {},
  useGitignore: true,
  useCustomIgnore: true,
};

export const useSettingsStore = defineStore('settings', () => {
  const uiStore = useUiStore();
  const { handleError } = useErrorHandler();
  const settings = ref<SettingsDTO>(JSON.parse(JSON.stringify(emptySettings)));
  const isLoading = ref(false);
  const isRefreshingModels = ref(false);

  async function fetchSettings() {
    isLoading.value = true;
    try {
      const newSettings = await GetSettings();
      settings.value = newSettings;
    } catch (err) {
      handleError(err, 'Fetch Settings');
    } finally {
      isLoading.value = false;
    }
  }

  async function saveSettings() {
    isLoading.value = true;
    try {
      await SaveSettings(settings.value);
      uiStore.addToast('Settings saved successfully', 'success');
    } catch (err) {
      handleError(err, 'Save Settings');
    } finally {
      isLoading.value = false;
    }
  }

  async function saveIgnoreSettings() {
    try {
      await SaveSettings(settings.value);
    } catch (err) {
      handleError(err, 'Save Ignore Settings');
    }
  }

  async function refreshModels(provider: string) {
    let apiKey = '';
    switch (provider) {
      case 'openai': apiKey = settings.value.openAIAPIKey; break;
      case 'gemini': apiKey = settings.value.geminiAPIKey; break;
      case 'openrouter': apiKey = settings.value.openRouterAPIKey; break;
      case 'localai': apiKey = settings.value.localAIAPIKey; break;
    }

    if (!apiKey && provider !== 'localai') {
      uiStore.addToast(`API key for ${provider} is not set.`, 'info');
      return;
    }

    isRefreshingModels.value = true;
    try {
      await RefreshAIModels(provider, apiKey);
      await fetchSettings();
      uiStore.addToast(`Model list for ${provider} has been updated.`, 'success');
    } catch (err)
    {
      handleError(err, `Refresh Models for ${provider}`);
    } finally {
      isRefreshingModels.value = false;
    }
  }

  fetchSettings();

  return {
    settings,
    isLoading,
    isRefreshingModels,
    fetchSettings,
    saveSettings,
    saveIgnoreSettings,
    refreshModels,
  };
});