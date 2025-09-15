<template>
  <div class="space-y-6">
    <h3 class="text-lg font-semibold text-white flex items-center gap-2">
      <svg
        class="w-5 h-5 text-purple-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
        />
      </svg>
      AI Provider
    </h3>

    <div class="space-y-4">
      <label class="block">
        <span class="text-sm font-medium text-gray-300">Provider</span>
        <select
          v-model="localSettings.selectedProvider"
          class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          @change="onProviderChange"
        >
          <option value="">Select Provider</option>
          <option value="openai">OpenAI</option>
          <option value="gemini">Google Gemini</option>
          <option value="localai">LocalAI</option>
          <option value="llamacpp">LlamaCPP</option>
          <option value="openrouter">OpenRouter</option>
        </select>
      </label>

      <div
        v-if="localSettings.selectedProvider && localSettings.selectedProvider !== 'localai'"
        class="space-y-4"
      >
        <label class="block">
          <span class="text-sm font-medium text-gray-300">API Key</span>
          <input
            v-model="apiKeyValue"
            type="password"
            class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="Enter your API key"
            @input="onApiKeyChange"
          />
        </label>
      </div>

      <div v-if="localSettings.selectedProvider === 'openai'" class="space-y-4">
        <label class="block">
          <span class="text-sm font-medium text-gray-300">Model</span>
          <select
            v-model="localSettings.openaiModel"
            class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            @change="onModelChange"
          >
            <option value="gpt-4">GPT-4</option>
            <option value="gpt-4-turbo">GPT-4 Turbo</option>
            <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
          </select>
        </label>
      </div>

      <div v-if="localSettings.selectedProvider === 'gemini'" class="space-y-4">
        <label class="block">
          <span class="text-sm font-medium text-gray-300">Model</span>
          <select
            v-model="localSettings.geminiModel"
            class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            @change="onModelChange"
          >
            <option value="gemini-pro">Gemini Pro</option>
            <option value="gemini-pro-vision">Gemini Pro Vision</option>
          </select>
        </label>
      </div>

      <!-- Connection Status -->
      <div v-if="localSettings.selectedProvider && apiKeyValue" class="mt-4">
        <div class="flex items-center gap-2 text-sm">
          <div class="w-2 h-2 rounded-full" :class="connectionStatusClass"></div>
          <span :class="connectionStatusTextClass">{{ connectionStatusText }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';

interface AIProviderSettings {
  selectedProvider: string;
  openaiModel: string;
  geminiModel: string;
  openAIAPIKey: string;
  geminiAPIKey: string;
  openRouterAPIKey: string;
  localAIAPIKey: string;
}

const props = defineProps<{
  settings: AIProviderSettings;
}>();

const emit = defineEmits<{
  'update:settings': [settings: AIProviderSettings];
  'provider-changed': [provider: string];
  'api-key-changed': [provider: string, apiKey: string];
  'model-changed': [provider: string, model: string];
}>();

const localSettings = ref<AIProviderSettings>({ ...props.settings });

// Connection status (mock for now)
const connectionStatus = ref<'unknown' | 'connected' | 'disconnected' | 'testing'>('unknown');

const apiKeyValue = computed<string>({
  get() {
    const provider = localSettings.value.selectedProvider;
    switch (provider) {
      case "openai":
        return localSettings.value.openAIAPIKey || "";
      case "gemini":
        return localSettings.value.geminiAPIKey || "";
      case "openrouter":
        return localSettings.value.openRouterAPIKey || "";
      case "localai":
        return localSettings.value.localAIAPIKey || "";
      default:
        return "";
    }
  },
  set(value: string) {
    const provider = localSettings.value.selectedProvider;
    switch (provider) {
      case "openai":
        localSettings.value.openAIAPIKey = value;
        break;
      case "gemini":
        localSettings.value.geminiAPIKey = value;
        break;
      case "openrouter":
        localSettings.value.openRouterAPIKey = value;
        break;
      case "localai":
        localSettings.value.localAIAPIKey = value;
        break;
    }
    emit('update:settings', localSettings.value);
  },
});

const connectionStatusClass = computed(() => {
  switch (connectionStatus.value) {
    case 'connected':
      return 'bg-green-400';
    case 'disconnected':
      return 'bg-red-400';
    case 'testing':
      return 'bg-yellow-400 animate-pulse';
    default:
      return 'bg-gray-400';
  }
});

const connectionStatusTextClass = computed(() => {
  switch (connectionStatus.value) {
    case 'connected':
      return 'text-green-400';
    case 'disconnected':
      return 'text-red-400';
    case 'testing':
      return 'text-yellow-400';
    default:
      return 'text-gray-400';
  }
});

const connectionStatusText = computed(() => {
  switch (connectionStatus.value) {
    case 'connected':
      return 'Connected';
    case 'disconnected':
      return 'Connection failed';
    case 'testing':
      return 'Testing connection...';
    default:
      return 'Not tested';
  }
});

// Watch for external settings changes
watch(
  () => props.settings,
  (newSettings) => {
    localSettings.value = { ...newSettings };
  },
  { deep: true }
);

// Watch for local settings changes and emit updates
watch(
  localSettings,
  (newSettings) => {
    emit('update:settings', newSettings);
  },
  { deep: true }
);

const onProviderChange = () => {
  connectionStatus.value = 'unknown';
  emit('provider-changed', localSettings.value.selectedProvider);
  emit('update:settings', localSettings.value);
};

const onApiKeyChange = () => {
  connectionStatus.value = 'unknown';
  emit('api-key-changed', localSettings.value.selectedProvider, apiKeyValue.value);
};

const onModelChange = () => {
  const provider = localSettings.value.selectedProvider;
  const model = provider === 'openai' ? localSettings.value.openaiModel : localSettings.value.geminiModel;
  emit('model-changed', provider, model);
  emit('update:settings', localSettings.value);
};

// Method to test connection (could be called by parent)
const testConnection = async () => {
  if (!localSettings.value.selectedProvider || !apiKeyValue.value) {
    connectionStatus.value = 'disconnected';
    return false;
  }

  connectionStatus.value = 'testing';
  
  try {
    // Mock connection test - in real implementation, this would call an API
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Simulate success/failure based on API key length (mock logic)
    const isValidKey = apiKeyValue.value.length > 10;
    connectionStatus.value = isValidKey ? 'connected' : 'disconnected';
    
    return isValidKey;
  } catch (error) {
    connectionStatus.value = 'disconnected';
    return false;
  }
};

// Expose methods to parent component
defineExpose({
  testConnection
});
</script>