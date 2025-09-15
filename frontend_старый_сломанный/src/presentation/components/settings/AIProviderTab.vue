<template>
  <div class="ai-provider-tab">
    <!-- Provider Selection -->
    <div class="setting-group">
      <label class="setting-label">AI Provider</label>
      <select 
        v-model="localSettings.selectedProvider"
        class="setting-input"
        @change="onProviderChange"
      >
        <option value="">Select a provider...</option>
        <option value="openai">OpenAI</option>
        <option value="gemini">Google Gemini</option>
        <option value="openrouter">OpenRouter</option>
        <option value="localai">Local AI</option>
      </select>
    </div>

    <!-- API Key Input -->
    <div v-if="localSettings.selectedProvider" class="setting-group">
      <label class="setting-label">
        API Key 
        <span v-if="localSettings.selectedProvider !== 'localai'" class="text-red-400">*</span>
      </label>
      <div class="relative">
        <input
          v-model="apiKeyValue"
          type="password"
          class="setting-input pr-20"
          placeholder="Enter your API key..."
          @input="onApiKeyChange"
        />
        <div class="absolute right-3 top-1/2 transform -translate-y-1/2 flex items-center gap-2">
          <div class="flex items-center gap-1">
            <div :class="[connectionStatusClass, 'w-2 h-2 rounded-full']"></div>
            <span :class="[connectionStatusTextClass, 'text-xs']">
              {{ connectionStatusText }}
            </span>
          </div>
          <button
            @click="testConnection"
            class="text-xs text-blue-400 hover:text-blue-300 focus:outline-none"
            :disabled="!apiKeyValue || connectionStatus === 'testing'"
          >
            Test
          </button>
        </div>
      </div>
    </div>

    <!-- Model Selection -->
    <div v-if="localSettings.selectedProvider && localSettings.selectedProvider !== 'localai'" class="setting-group">
      <label class="setting-label">Model</label>
      <select 
        v-if="localSettings.selectedProvider === 'openai'"
        v-model="localSettings.openaiModel"
        class="setting-input"
        @change="onModelChange"
      >
        <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
        <option value="gpt-4">GPT-4</option>
        <option value="gpt-4-turbo">GPT-4 Turbo</option>
        <option value="gpt-4o">GPT-4o</option>
      </select>
      <select 
        v-else-if="localSettings.selectedProvider === 'gemini'"
        v-model="localSettings.geminiModel"
        class="setting-input"
        @change="onModelChange"
      >
        <option value="gemini-pro">Gemini Pro</option>
        <option value="gemini-pro-vision">Gemini Pro Vision</option>
        <option value="gemini-1.5-pro">Gemini 1.5 Pro</option>
      </select>
      <select 
        v-else-if="localSettings.selectedProvider === 'openrouter'"
        v-model="localSettings.openaiModel"
        class="setting-input"
        @change="onModelChange"
      >
        <option value="openai/gpt-3.5-turbo">GPT-3.5 Turbo</option>
        <option value="openai/gpt-4">GPT-4</option>
        <option value="anthropic/claude-3-haiku">Claude 3 Haiku</option>
        <option value="anthropic/claude-3-sonnet">Claude 3 Sonnet</option>
        <option value="anthropic/claude-3-opus">Claude 3 Opus</option>
      </select>
    </div>

    <!-- Local AI Configuration -->
    <div v-if="localSettings.selectedProvider === 'localai'" class="setting-group">
      <div class="text-sm text-gray-400 mb-4">
        Configure your local AI instance. Make sure your local server is running.
      </div>
      <label class="setting-label">Local Server URL</label>
      <input
        v-model="localSettings.localAIAPIKey"
        type="text"
        class="setting-input"
        placeholder="http://localhost:8080"
        @input="onApiKeyChange"
      />
    </div>

    <!-- Provider Information -->
    <div v-if="localSettings.selectedProvider" class="mt-6 p-4 bg-gray-800/50 rounded-lg">
      <h4 class="text-sm font-medium text-gray-300 mb-2">Provider Information</h4>
      <div v-if="localSettings.selectedProvider === 'openai'" class="text-sm text-gray-400">
        <p>OpenAI provides state-of-the-art language models. Get your API key from the OpenAI dashboard.</p>
        <a href="https://platform.openai.com/api-keys" target="_blank" class="text-blue-400 hover:text-blue-300">
          Get API Key →
        </a>
      </div>
      <div v-else-if="localSettings.selectedProvider === 'gemini'" class="text-sm text-gray-400">
        <p>Google's Gemini models offer excellent performance. Get your API key from Google AI Studio.</p>
        <a href="https://makersuite.google.com/app/apikey" target="_blank" class="text-blue-400 hover:text-blue-300">
          Get API Key →
        </a>
      </div>
      <div v-else-if="localSettings.selectedProvider === 'openrouter'" class="text-sm text-gray-400">
        <p>OpenRouter provides access to multiple AI models through a single API.</p>
        <a href="https://openrouter.ai/keys" target="_blank" class="text-blue-400 hover:text-blue-300">
          Get API Key →
        </a>
      </div>
      <div v-else-if="localSettings.selectedProvider === 'localai'" class="text-sm text-gray-400">
        <p>LocalAI allows you to run AI models locally for privacy and control.</p>
        <a href="https://localai.io/basics/getting_started/" target="_blank" class="text-blue-400 hover:text-blue-300">
          Setup Guide →
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { APP_CONFIG } from '@/config/app-config';

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

<style scoped>
.ai-provider-tab {
  space-y: 6;
}

.setting-group {
  margin-bottom: 1.5rem;
}

.setting-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: rgb(209, 213, 219);
  margin-bottom: 0.5rem;
}

.setting-input {
  width: 100%;
  padding: 0.75rem;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 0.5rem;
  color: rgb(248, 250, 252);
  font-size: 0.875rem;
  transition: all 0.2s;
}

.setting-input:focus {
  outline: none;
  border-color: rgb(59, 130, 246);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.setting-input::placeholder {
  color: rgb(100, 116, 139);
}

.relative {
  position: relative;
}

.absolute {
  position: absolute;
}

.right-3 {
  right: 0.75rem;
}

.top-1\/2 {
  top: 50%;
}

.transform {
  transform: var(--tw-transform);
}

.-translate-y-1\/2 {
  --tw-translate-y: -50%;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.pr-20 {
  padding-right: 5rem;
}
</style>