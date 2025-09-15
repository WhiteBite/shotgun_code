<template>
  <div
    class="fixed inset-y-0 right-0 w-full max-w-md bg-gradient-to-b from-gray-900 to-gray-800 border-l border-gray-700 shadow-2xl transform transition-transform duration-300 ease-in-out z-50 flex flex-col"
    :class="isOpen ? 'translate-x-0' : 'translate-x-full'"
  >
    <!-- Header -->
    <div
      class="flex items-center justify-between p-6 border-b border-gray-700 bg-gradient-to-r from-gray-800 to-gray-700"
    >
      <h2 class="text-xl font-bold text-white flex items-center gap-3">
        <svg
          class="w-6 h-6 text-blue-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
          />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
          />
        </svg>
        Settings
      </h2>
      <button
        class="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors duration-200"
        @click="close"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </div>

    <!-- Tab Navigation -->
    <div class="border-b border-gray-700 flex overflow-x-auto">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        class="px-4 py-3 text-sm font-medium whitespace-nowrap transition-colors"
        :class="activeTab === tab.id 
          ? 'text-blue-400 border-b-2 border-blue-400' 
          : 'text-gray-400 hover:text-gray-300'"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-6">
      <!-- AI Provider Settings -->
      <AIProviderTab
        v-show="activeTab === 'ai'"
        :settings="aiProviderSettings"
        @update:settings="updateAIProviderSettings"
        @provider-changed="onProviderChanged"
        @api-key-changed="onAPIKeyChanged"
      />

      <!-- Context Settings -->
      <ContextSettingsTab
        v-show="activeTab === 'context'"
        :settings="contextSettings"
        @update:settings="updateContextSettings"
      />

      <!-- Context Builder Settings -->
      <ContextBuilderTab
        v-show="activeTab === 'builder'"
        :settings="contextBuilderSettings"
        @update:settings="updateContextBuilderSettings"
      />

      <!-- Code Generation Settings -->
      <CodeGenerationTab
        v-show="activeTab === 'code'"
        :settings="codeGenerationSettings"
        @update:settings="updateCodeGenerationSettings"
      />

      <!-- Split Settings -->
      <SplitSettingsTab
        v-show="activeTab === 'split'"
        :settings="splitSettings"
        @update:settings="updateSplitSettings"
      />

      <!-- Safety Settings -->
      <SafetySettingsTab
        v-show="activeTab === 'safety'"
        :settings="safetySettings"
        @update:settings="updateSafetySettings"
      />

      <!-- Error Display -->
      <div
        v-if="error"
        class="p-4 bg-red-900/30 border border-red-700 rounded-lg"
      >
        <div class="flex items-center gap-2">
          <svg
            class="w-5 h-5 text-red-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <p class="text-red-300 text-sm">{{ error }}</p>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div
      class="p-6 border-t border-gray-700 bg-gradient-to-r from-gray-800 to-gray-700"
    >
      <div class="flex items-center justify-between">
        <button
          class="px-4 py-2 text-sm text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors duration-200"
          @click="resetToDefaults"
        >
          Reset to Defaults
        </button>
        <div class="flex items-center gap-3">
          <button
            class="px-4 py-2 text-sm text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors duration-200"
            @click="close"
          >
            Cancel
          </button>
          <Button
            :disabled="isSaving"
            variant="primary"
            size="md"
            @click="save"
          >
            <svg
              v-if="isSaving"
              class="animate-spin h-4 w-4 mr-2"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{ isSaving ? "Saving..." : "Save Settings" }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { useSettingsStore } from "@/stores/settings.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useNotificationsStore } from "@/stores/notifications.store";
import { Button } from "@/presentation/components/ui/button";

// Import decomposed tab components
import AIProviderTab from "@/presentation/components/settings/AIProviderTab.vue";
import ContextSettingsTab from "@/presentation/components/settings/ContextSettingsTab.vue";
import ContextBuilderTab from "@/presentation/components/settings/ContextBuilderTab.vue";
import CodeGenerationTab from "@/presentation/components/settings/CodeGenerationTab.vue";
import SplitSettingsTab from "@/presentation/components/settings/SplitSettingsTab.vue";
import SafetySettingsTab from "@/presentation/components/settings/SafetySettingsTab.vue";

defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  close: [];
}>();

const settingsStore = useSettingsStore();
const contextBuilderStore = useContextBuilderStore();
const notifications = useNotificationsStore();

const settings = ref({ ...settingsStore.settings });
const isSaving = ref(false);
const error = ref("");
const activeTab = ref("ai");

const tabs = [
  { id: "ai", label: "AI Provider" },
  { id: "context", label: "Context" },
  { id: "builder", label: "Builder" },
  { id: "code", label: "Code Gen" },
  { id: "split", label: "Split" },
  { id: "safety", label: "Safety" }
];

// Settings for each tab component
const aiProviderSettings = ref({
  selectedProvider: settings.value.selectedProvider || "",
  openaiModel: settings.value.openaiModel || "gpt-4",
  geminiModel: settings.value.geminiModel || "gemini-pro",
  openAIAPIKey: settings.value.openAIAPIKey || "",
  geminiAPIKey: settings.value.geminiAPIKey || "",
  openRouterAPIKey: settings.value.openRouterAPIKey || "",
  localAIAPIKey: settings.value.localAIAPIKey || ""
});

const contextSettings = ref({
  maxContextSize: settings.value.maxContextSize || 10000,
  maxFilesInContext: settings.value.maxFilesInContext || 20,
  includeDependencies: settings.value.includeDependencies || false,
  includeTests: settings.value.includeTests || false
});

const contextBuilderSettings = ref({
  maxTokenLimit: contextBuilderStore.maxTokenLimit,
  maxFileLimit: contextBuilderStore.maxFileLimit,
  excludeBinaryFiles: contextBuilderStore.excludeBinaryFiles,
  autoBuildEnabled: contextBuilderStore.autoBuildEnabled,
  smartSuggestionsEnabled: contextBuilderStore.smartSuggestionsEnabled,
  allowedExtensions: [...contextBuilderStore.allowedExtensions]
});

const codeGenerationSettings = ref({
  temperature: settings.value.temperature || 0.7,
  maxTokens: settings.value.maxTokens || 2000,
  autoFormat: settings.value.autoFormat || true,
  includeComments: settings.value.includeComments || true
});

const splitSettings = ref({
  maxTokensPerChunk: (settings.value as any).maxTokensPerChunk || 8000,
  enableAutoSplitting: (settings.value as any).enableAutoSplitting || false,
  overlapSize: (settings.value as any).overlapSize || 100,
  splitStrategy: (settings.value as any).splitStrategy || 'semantic' as 'semantic' | 'balanced' | 'priority',
  preserveCodeBlocks: (settings.value as any).preserveCodeBlocks || true
});

const safetySettings = ref({
  enableGuardrails: (settings.value as any).enableGuardrails || false,
  maxMemoryWarning: (settings.value as any).maxMemoryWarning || 30,
  maxContextSizeMB: (settings.value as any).maxContextSizeMB || 1,
  validateFileTypes: (settings.value as any).validateFileTypes || true,
  excludeBinaryFiles: (settings.value as any).excludeBinaryFiles || true,
  scanForSecrets: (settings.value as any).scanForSecrets || false,
  enableContentFiltering: (settings.value as any).enableContentFiltering || false,
  requireUserConfirmation: (settings.value as any).requireUserConfirmation || true,
  safetyLevel: (settings.value as any).safetyLevel || 'balanced' as 'relaxed' | 'balanced' | 'strict'
});

// Update handlers for tab components
const updateAIProviderSettings = (newSettings: any) => {
  aiProviderSettings.value = { ...newSettings };
  // Update main settings object
  Object.assign(settings.value, newSettings);
};

const updateContextSettings = (newSettings: any) => {
  contextSettings.value = { ...newSettings };
  Object.assign(settings.value, newSettings);
};

const updateContextBuilderSettings = (newSettings: any) => {
  contextBuilderSettings.value = { ...newSettings };
};

const updateCodeGenerationSettings = (newSettings: any) => {
  codeGenerationSettings.value = { ...newSettings };
  Object.assign(settings.value, newSettings);
};

const updateSplitSettings = (newSettings: any) => {
  splitSettings.value = { ...newSettings };
  Object.assign(settings.value, newSettings);
};

const updateSafetySettings = (newSettings: any) => {
  safetySettings.value = { ...newSettings };
  Object.assign(settings.value, newSettings);
};

// Event handlers
const onProviderChanged = (provider: string) => {
  console.log('Provider changed to:', provider);
};

const onAPIKeyChanged = (provider: string, apiKey: string) => {
  console.log('API key changed for provider:', provider);
};

// Watch for context builder settings changes
watch(
  () => {
    const store = contextBuilderStore;
    return {
      maxTokenLimit: store.maxTokenLimit,
      maxFileLimit: store.maxFileLimit,
      excludeBinaryFiles: store.excludeBinaryFiles,
      autoBuildEnabled: store.autoBuildEnabled,
      smartSuggestionsEnabled: store.smartSuggestionsEnabled,
      allowedExtensions: store.allowedExtensions
    };
  },
  (newValues) => {
    contextBuilderSettings.value = {
      maxTokenLimit: newValues.maxTokenLimit,
      maxFileLimit: newValues.maxFileLimit,
      excludeBinaryFiles: newValues.excludeBinaryFiles,
      autoBuildEnabled: newValues.autoBuildEnabled,
      smartSuggestionsEnabled: newValues.smartSuggestionsEnabled,
      allowedExtensions: [...newValues.allowedExtensions]
    };
  },
  { deep: true }
);

// Watch for settings changes and update local copy
watch(
  () => settingsStore.settings,
  (newSettings) => {
    settings.value = { ...newSettings };
    // Update tab-specific settings
    Object.assign(aiProviderSettings.value, {
      selectedProvider: newSettings.selectedProvider || "",
      openaiModel: newSettings.openaiModel || "gpt-4",
      geminiModel: newSettings.geminiModel || "gemini-pro",
      openAIAPIKey: newSettings.openAIAPIKey || "",
      geminiAPIKey: newSettings.geminiAPIKey || "",
      openRouterAPIKey: newSettings.openRouterAPIKey || "",
      localAIAPIKey: newSettings.localAIAPIKey || ""
    });
  },
  { deep: true },
);

async function save() {
  isSaving.value = true;
  error.value = "";

  try {
    // Save general settings
    settingsStore.settings = { ...settings.value } as any;
    await settingsStore.saveSettings();
    
    // Save context builder settings
    contextBuilderStore.setMaxTokenLimit(contextBuilderSettings.value.maxTokenLimit);
    contextBuilderStore.setMaxFileLimit(contextBuilderSettings.value.maxFileLimit);
    if (contextBuilderSettings.value.excludeBinaryFiles !== contextBuilderStore.excludeBinaryFiles) {
      contextBuilderStore.toggleExcludeBinaryFiles();
    }
    if (contextBuilderSettings.value.autoBuildEnabled !== contextBuilderStore.autoBuildEnabled) {
      contextBuilderStore.toggleAutoBuild();
    }
    if (contextBuilderSettings.value.smartSuggestionsEnabled !== contextBuilderStore.smartSuggestionsEnabled) {
      contextBuilderStore.toggleSmartSuggestions();
    }
    contextBuilderStore.updateAllowedExtensions(contextBuilderSettings.value.allowedExtensions);
    
    notifications.addLog("Settings saved successfully", "success");
    emit("close");
  } catch (err) {
    console.error("Error saving settings:", err);
    error.value =
      err instanceof Error ? err.message : "Failed to save settings";
    notifications.addLog("Failed to save settings", "error");
  } finally {
    isSaving.value = false;
  }
}

function resetToDefaults() {
  // Reset AI provider settings
  aiProviderSettings.value = {
    selectedProvider: "",
    openaiModel: "gpt-4",
    geminiModel: "gemini-pro",
    openAIAPIKey: "",
    geminiAPIKey: "",
    openRouterAPIKey: "",
    localAIAPIKey: ""
  };
  
  // Reset context settings
  contextSettings.value = {
    maxContextSize: 10000,
    maxFilesInContext: 20,
    includeDependencies: false,
    includeTests: false
  };
  
  // Reset context builder settings
  contextBuilderSettings.value = {
    maxTokenLimit: 0,
    maxFileLimit: 0,
    excludeBinaryFiles: true,
    autoBuildEnabled: true,
    smartSuggestionsEnabled: true,
    allowedExtensions: [
      '.ts', '.js', '.vue', '.tsx', '.jsx',
      '.py', '.java', '.go', '.rs', '.cpp', '.c',
      '.cs', '.php', '.rb', '.swift', '.kt',
      '.md', '.txt', '.json', '.yaml', '.yml',
      '.xml', '.html', '.css', '.scss', '.less',
      '.sql', '.sh', '.bat', '.ps1', '.dockerfile'
    ]
  };
  
  // Reset code generation settings
  codeGenerationSettings.value = {
    temperature: 0.7,
    maxTokens: 2000,
    autoFormat: true,
    includeComments: true
  };
  
  // Reset split settings
  splitSettings.value = {
    maxTokensPerChunk: 8000,
    enableAutoSplitting: false,
    overlapSize: 100,
    splitStrategy: 'semantic',
    preserveCodeBlocks: true
  };
  
  // Reset safety settings
  safetySettings.value = {
    enableGuardrails: false,
    maxMemoryWarning: 30,
    maxContextSizeMB: 1,
    validateFileTypes: true,
    excludeBinaryFiles: true,
    scanForSecrets: false,
    enableContentFiltering: false,
    requireUserConfirmation: true,
    safetyLevel: 'balanced'
  };
  
  // Update main settings object
  Object.assign(settings.value, aiProviderSettings.value, contextSettings.value, 
    codeGenerationSettings.value, splitSettings.value, safetySettings.value);
  
  notifications.addLog("Settings reset to defaults", "info");
}

function sanitizeNumber(key: string, min: number, max: number, fallback: number) {
  const val = Number((settings.value as any)[key]);
  if (Number.isNaN(val)) {
    (settings.value as any)[key] = fallback;
    return;
  }
  (settings.value as any)[key] = Math.min(max, Math.max(min, val));
}

function close() {
  emit("close");
}
</script>

<style scoped>
.slider::-webkit-slider-thumb {
  appearance: none;
  height: 16px;
  width: 16px;
  border-radius: 50%;
  background: #3b82f6;
  cursor: pointer;
  border: 2px solid #1e40af;
}

.slider::-moz-range-thumb {
  height: 16px;
  width: 16px;
  border-radius: 50%;
  background: #3b82f6;
  cursor: pointer;
  border: 2px solid #1e40af;
}
</style>