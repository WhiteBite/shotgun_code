import { defineStore } from "pinia";
import { ref } from "vue";
import type { SettingsDTO } from "@/types/dto";
import { useUiStore } from "./ui.store";
import { useAdvancedErrorHandler } from "@/composables/useErrorHandler";
import { APP_CONFIG } from "@/config/app-config";
import { createStoreWithDependencies, type StoreDependencies } from '@/stores/StoreDependencyContainer';
// Import repositories
import type { SettingsRepository } from "@/domain/repositories/SettingsRepository";
import type { StorageRepository } from "@/domain/repositories/StorageRepository";

const emptySettings: SettingsDTO = {
  customIgnoreRules: "",
  customPromptRules: "",
  openAIAPIKey: "",
  geminiAPIKey: "",
  openRouterAPIKey: "",
  localAIAPIKey: "",
  localAIHost: "",
  localAIModelName: "",
  selectedProvider: "openai",
  selectedModels: {},
  availableModels: {},
  useGitignore: true,
  useCustomIgnore: true,
  // Context splitting defaults using centralized configuration
  maxTokensPerChunk: APP_CONFIG.performance.streaming.CHUNK_SIZE,
  overlapTokens: Math.round(APP_CONFIG.performance.streaming.CHUNK_SIZE * 0.1),
  splitStrategy: 'semantic',
};

export const useSettingsStore = defineStore("settings", () => {
  return createStoreWithDependencies('settings', (dependencies: StoreDependencies) => {
    // Inject repositories through constructor parameters
    const { settingsRepository, storageRepository, securityValidationService, auditLoggingService } = dependencies;
    
    const uiStore = useUiStore();
    const { handleStructuredError } = useAdvancedErrorHandler();
    
    // Remove direct container access
    // const settingsRepository: SettingsRepository = container.settingsRepository;
    // const storageRepository: StorageRepository = localStorageService;

    const settings = ref<SettingsDTO>(JSON.parse(JSON.stringify(emptySettings)));
    const isLoading = ref(false);
    const isRefreshingModels = ref(false);
    
    // UI-specific settings (stored locally)
    const autoOpenLastProject = ref(true);

    async function fetchSettings() {
      isLoading.value = true;
      
      // Add retry logic for Wails initialization timing
      const maxRetries = 3;
      let retryCount = 0;
      
      while (retryCount < maxRetries) {
        try {
          // Use SettingsRepository instead of direct Wails function
          const newSettings = await settingsRepository.getSettings();
          settings.value = newSettings;
          break; // Success, exit retry loop
        } catch (err) {
          retryCount++;
          
          // Check if this is a Wails initialization error
          const errorMessage = err instanceof Error ? err.message : String(err);
          if (errorMessage.includes('Wails runtime not available') && retryCount < maxRetries) {
            console.warn(`Wails not ready, retrying... (${retryCount}/${maxRetries})`);
            // Wait before retrying
            await new Promise(resolve => setTimeout(resolve, 1000));
            continue;
          }
          
          // If we've exhausted retries or it's a different error, handle it
          handleStructuredError(err, { operation: "Fetch Settings", component: "SettingsStore" });
          break;
        }
      }
      
      isLoading.value = false;
    }

    async function saveSettings() {
      isLoading.value = true;
      try {
        // Use SettingsRepository instead of direct Wails function
        await settingsRepository.saveSettings(settings.value);
        uiStore.addToast("Settings saved successfully", "success");
      } catch (err) {
        handleStructuredError(err, { operation: "Save Settings", component: "SettingsStore" });
      } finally {
        isLoading.value = false;
      }
    }

    async function saveIgnoreSettings() {
      try {
        // Use SettingsRepository instead of direct Wails function
        await settingsRepository.saveSettings(settings.value);
      } catch (err) {
        handleStructuredError(err, { operation: "Save Ignore Settings", component: "SettingsStore" });
      }
    }

    async function refreshModels(provider: string) {
      let apiKey = "";
      switch (provider) {
        case "openai":
          apiKey = settings.value.openAIAPIKey;
          break;
        case "gemini":
          apiKey = settings.value.geminiAPIKey;
          break;
        case "openrouter":
          apiKey = settings.value.openRouterAPIKey;
          break;
        case "localai":
          apiKey = settings.value.localAIAPIKey;
          break;
      }

      if (!apiKey && provider !== "localai") {
        uiStore.addToast(`API key for ${provider} is not set.`, "info");
        return;
      }

      isRefreshingModels.value = true;
      try {
        // Use SettingsRepository instead of direct Wails function
        await settingsRepository.refreshAIModels(provider, apiKey);
        await fetchSettings();
        uiStore.addToast(
          `Model list for ${provider} has been updated.`,
          "success",
        );
      } catch (err) {
        handleStructuredError(err, { operation: `Refresh Models for ${provider}`, component: "SettingsStore" });
      } finally {
        isRefreshingModels.value = false;
      }
    }

    // Load UI settings from localStorage using domain service
    function loadUISettings() {
      try {
        // Use StorageRepository instead of localStorageService directly
        const uiSettings = storageRepository.get<{autoOpenLastProject?: boolean}>('ui-settings');
        if (uiSettings) {
          autoOpenLastProject.value = uiSettings.autoOpenLastProject ?? true;
        }
      } catch (err) {
        console.warn("Failed to load UI settings:", err);
      }
    }

    // Save UI settings to localStorage using domain service
    function saveUISettings() {
      try {
        const uiSettings = {
          autoOpenLastProject: autoOpenLastProject.value,
        };
        // Use StorageRepository instead of localStorageService directly
        storageRepository.set('ui-settings', uiSettings);
        // Note: auditLoggingService.logOperation method doesn't exist, removing this call
        console.log('UI settings saved:', { autoOpenLastProject: autoOpenLastProject.value });
      } catch (err) {
        console.warn("Failed to save UI settings:", err);
      }
    }

    // Toggle auto-open last project setting
    function toggleAutoOpenLastProject() {
      autoOpenLastProject.value = !autoOpenLastProject.value;
      saveUISettings();
    }

    // Initialize settings
    fetchSettings();
    loadUISettings();

    return {
      settings,
      isLoading,
      isRefreshingModels,
      autoOpenLastProject,
      fetchSettings,
      saveSettings,
      saveIgnoreSettings,
      refreshModels,
      toggleAutoOpenLastProject,
      loadUISettings,
      saveUISettings,
    };
  }); // Close dependency injection wrapper
});