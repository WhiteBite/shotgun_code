import { defineStore } from 'pinia';
import { ref } from 'vue';
import { GetSettings, SaveSettings, RefreshAIModels } from '../../wailsjs/go/main/App';
import type { SettingsDTO } from '@/types/dto';
import { useUiStore } from './uiStore';

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
  const settings = ref<SettingsDTO>(JSON.parse(JSON.stringify(emptySettings)));
  const isLoading = ref(false);
  const isRefreshingModels = ref(false);

  async function fetchSettings() {
    isLoading.value = true;
    try {
      const newSettings = await GetSettings();
      settings.value = newSettings;
    } catch (err: any) {
      uiStore.addToast(`Ошибка загрузки настроек: ${err.message || err}`, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  async function saveSettings() {
    isLoading.value = true;
    try {
      await SaveSettings(settings.value);
      uiStore.addToast('Настройки успешно сохранены', 'success');
    } catch (err: any) {
      uiStore.addToast(`Ошибка сохранения настроек: ${err.message || err}`, 'error');
    } finally {
      isLoading.value = false;
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

    if (!apiKey && provider !== 'localai') { // localai might not need a key
      uiStore.addToast(`API ключ для ${provider} не указан.`, 'info');
      return;
    }

    isRefreshingModels.value = true;
    try {
      await RefreshAIModels(provider, apiKey);
      await fetchSettings(); // Refetch settings to get the new model list
      uiStore.addToast(`Список моделей для ${provider} обновлен.`, 'success');
    } catch (err: any)
    {
      uiStore.addToast(`Ошибка обновления моделей: ${err.message || err}`, 'error');
    } finally {
      isRefreshingModels.value = false;
    }
  }

  // Initial fetch
  fetchSettings();

  return {
    settings,
    isLoading,
    isRefreshingModels,
    fetchSettings,
    saveSettings,
    refreshModels,
  };
});