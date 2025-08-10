import { defineStore } from 'pinia';
import { ref } from 'vue';
import {
  GetCustomIgnoreRules, SetCustomIgnoreRules,
  GetCustomPromptRules, SetCustomPromptRules,
  GetOpenAIKey, SetOpenAIKey,
  GetGeminiKey, SetGeminiKey,
  GetLocalAIKey, SetLocalAIKey,
  GetLocalAIHost, SetLocalAIHost,
  GetLocalAIModelName, SetLocalAIModelName,
  GetSelectedAIProvider, SetSelectedAIProvider,
  GetModels, GetSelectedModel, SetSelectedModel,
  SetUseGitignore as apiSetUseGitignore,
  SetUseCustomIgnore as apiSetUseCustomIgnore,
  RefreshAIModels
} from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notificationsStore';

export const useSettingsStore = defineStore('settings', () => {
  const notifications = useNotificationsStore();

  const useGitignore = ref(true);
  const useCustomIgnore = ref(true);
  const customIgnoreRules = ref('');
  const customPromptRules = ref('');

  const openAIKey = ref('');
  const geminiKey = ref('');
  const localAIKey = ref('');
  const localAIHost = ref('');
  const localAIModelName = ref('');

  const selectedProvider = ref('openai');
  const availableModels = ref([]);
  const loadingError = ref('');
  const selectedModel = ref('');
  const isLoadingModels = ref(false);
  const isRefreshingModels = ref(false);
  const refreshError = ref('');

  async function initializeSettings() {
    try {
      customIgnoreRules.value = await GetCustomIgnoreRules();
      customPromptRules.value = await GetCustomPromptRules();
      openAIKey.value = await GetOpenAIKey();
      geminiKey.value = await GetGeminiKey();
      localAIKey.value = await GetLocalAIKey();
      localAIHost.value = await GetLocalAIHost();
      localAIModelName.value = await GetLocalAIModelName();
      selectedProvider.value = await GetSelectedAIProvider() || 'openai';
      await fetchAvailableModels(selectedProvider.value);
      notifications.addLog('Настройки успешно загружены.', 'info');
    } catch (err) {
      notifications.addLog(`Не удалось загрузить настройки: ${err}`, 'error');
    }
  }

  async function fetchAvailableModels(provider, forceSelectDefault = false) {
    isLoadingModels.value = true;
    loadingError.value = '';
    try {
      availableModels.value = await GetModels(provider);
      const currentSelected = await GetSelectedModel(provider);
      if (forceSelectDefault || !availableModels.value.includes(currentSelected)) {
        if (availableModels.value.length > 0) {
          await setSelectedModel(provider, availableModels.value[0]);
          selectedModel.value = availableModels.value[0];
        } else {
          selectedModel.value = '';
        }
      } else {
        selectedModel.value = currentSelected;
      }
    } catch (err) {
      loadingError.value = err.message || err;
      notifications.addLog(`Не удалось получить список моделей для ${provider}: ${err}`, 'error');
      availableModels.value = [];
    } finally {
      isLoadingModels.value = false;
    }
  }

  async function refreshModels(apiKey) {
    isRefreshingModels.value = true;
    refreshError.value = '';
    try {
      await RefreshAIModels(selectedProvider.value, apiKey);
      notifications.addLog(`Список моделей для ${selectedProvider.value} обновлен.`, 'success');
      await fetchAvailableModels(selectedProvider.value, true);
    } catch (err) {
      const errorMessage = err.message || err;
      refreshError.value = errorMessage;
      notifications.addLog(`Ошибка обновления списка моделей: ${errorMessage}`, 'error');
    } finally {
      isRefreshingModels.value = false;
    }
  }

  async function updateUseGitignore(value) {
    useGitignore.value = value;
    await apiSetUseGitignore(value);
    notifications.addLog(`Использование .gitignore ${value ? 'включено' : 'выключено'}.`, 'info');
  }

  async function updateUseCustomIgnore(value) {
    useCustomIgnore.value = value;
    await apiSetUseCustomIgnore(value);
    notifications.addLog(`Использование кастомных правил ${value ? 'включено' : 'выключено'}.`, 'info');
  }

  async function saveCustomIgnoreRules(rules) {
    await SetCustomIgnoreRules(rules);
    customIgnoreRules.value = rules;
    notifications.addLog('Кастомные правила игнорирования сохранены.', 'success');
  }

  async function saveCustomPromptRules(rules) {
    await SetCustomPromptRules(rules);
    customPromptRules.value = rules;
    notifications.addLog('Кастомные правила для промптов сохранены.', 'success');
  }

  async function setOpenAIKey(key) { await SetOpenAIKey(key); openAIKey.value = key; }
  async function setGeminiKey(key) { await SetGeminiKey(key); geminiKey.value = key; }
  async function setLocalAIKey(key) { await SetLocalAIKey(key); localAIKey.value = key; }
  async function setLocalAIHost(host) { await SetLocalAIHost(host); localAIHost.value = host; }
  async function setLocalAIModelName(name) { await SetLocalAIModelName(name); localAIModelName.value = name; }

  async function setSelectedProvider(provider) {
    await SetSelectedAIProvider(provider);
    selectedProvider.value = provider;
    refreshError.value = '';
    await fetchAvailableModels(provider);
    notifications.addLog(`Выбран AI провайдер: ${provider}.`, 'info');
  }

  async function setSelectedModel(provider, model) {
    await SetSelectedModel(provider, model);
    selectedModel.value = model;
    notifications.addLog(`Выбрана модель: ${model}.`, 'info');
  }

  return {
    useGitignore, useCustomIgnore, customIgnoreRules, customPromptRules,
    openAIKey, geminiKey, localAIKey, localAIHost, localAIModelName,
    selectedProvider, availableModels, loadingError,
    isLoadingModels, isRefreshingModels, selectedModel, refreshError,
    initializeSettings, updateUseGitignore, updateUseCustomIgnore,
    saveCustomIgnoreRules, saveCustomPromptRules,
    setOpenAIKey, setGeminiKey, setLocalAIKey, setLocalAIHost, setLocalAIModelName,
    setSelectedProvider, setSelectedModel, refreshModels,
  };
});