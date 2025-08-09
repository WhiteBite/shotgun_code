
import { defineStore } from 'pinia';
import { ref } from 'vue';
import {
  GetCustomIgnoreRules,
  SetCustomIgnoreRules,
  GetCustomPromptRules,
  SetCustomPromptRules,
  SetUseGitignore as apiSetUseGitignore,
  SetUseCustomIgnore as apiSetUseCustomIgnore
} from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notifications';

export const useSettingsStore = defineStore('settings', () => {
  const notifications = useNotificationsStore();

  const useGitignore = ref(true);
  const useCustomIgnore = ref(true);
  const customIgnoreRules = ref('');
  const customPromptRules = ref('');

  async function initializeSettings() {
    try {
      customIgnoreRules.value = await GetCustomIgnoreRules();
      customPromptRules.value = await GetCustomPromptRules();
      notifications.addLog('Настройки успешно загружены.', 'info');
    } catch (err) {
      notifications.addLog(`Не удалось загрузить настройки: ${err}`, 'error');
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

  // ИСПРАВЛЕНО: Добавлен action для обновления правил промпта
  function updateCustomPromptRules(rules) {
    customPromptRules.value = rules;
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

  return {
    useGitignore,
    useCustomIgnore,
    customIgnoreRules,
    customPromptRules,
    initializeSettings,
    updateUseGitignore,
    updateUseCustomIgnore,
    updateCustomPromptRules, // Экспортируем новый action
    saveCustomIgnoreRules,
    saveCustomPromptRules,
  };
});