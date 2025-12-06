<template>
  <div class="space-y-4">
    <!-- Section Header -->
    <div class="flex items-center gap-2">
      <div class="section-icon section-icon-purple">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
        </svg>
      </div>
      <span class="text-sm font-semibold text-white">{{ t('settings.aiProvider') }}</span>
    </div>

    <!-- Provider Selection Dropdown -->
    <div class="context-stats">
      <label class="text-xs text-gray-400 mb-2 block">{{ t('settings.selectProvider') }}</label>
      <div class="relative">
        <select
          :value="settings.selectedProvider"
          @change="selectProvider(($event.target as HTMLSelectElement).value)"
          class="input text-sm !py-2.5 pr-10 appearance-none cursor-pointer"
        >
          <option v-for="provider in providers" :key="provider.id" :value="provider.id">
            {{ provider.icon }} {{ provider.name }}
          </option>
        </select>
        <div class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-gray-400">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>
      <p class="text-[10px] text-gray-500 mt-1.5">{{ currentProviderDescription }}</p>
    </div>

    <!-- API Key Input (not for CLI) -->
    <div v-if="settings.selectedProvider && settings.selectedProvider !== 'qwen-cli'" class="context-stats">
      <label class="text-xs text-gray-400 mb-2 block">
        {{ t('settings.apiKey') }} ({{ currentProviderName }})
      </label>
      <div class="relative">
        <input
          :type="showApiKey ? 'text' : 'password'"
          :value="currentApiKey"
          @input="updateApiKey(($event.target as HTMLInputElement).value)"
          class="input text-xs !py-2 pr-16"
          :placeholder="t('settings.apiKeyPlaceholder')"
        />
        <div class="absolute right-2 top-1/2 -translate-y-1/2 flex gap-1">
          <button
            @click="showApiKey = !showApiKey"
            class="icon-btn-sm"
            :title="showApiKey ? 'Hide' : 'Show'"
          >
            <svg v-if="showApiKey" class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
            </svg>
            <svg v-else class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
          </button>
          <button
            v-if="currentApiKey"
            @click="clearApiKey"
            class="icon-btn-sm icon-btn-danger"
            title="Clear"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
      <p class="text-[10px] text-gray-500 mt-1.5">{{ currentProviderHint }}</p>
    </div>

    <!-- Qwen CLI Info -->
    <div v-if="settings.selectedProvider === 'qwen-cli'" class="info-box-purple">
      <div class="flex items-start gap-2">
        <svg class="w-4 h-4 text-purple-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div>
          <p class="text-xs text-purple-300 font-medium mb-1">Qwen Code CLI</p>
          <p class="text-[10px] text-gray-400">Использует локально установленный qwen-coder-cli. API ключ не требуется.</p>
        </div>
      </div>
    </div>

    <!-- Model Selection -->
    <div v-if="settings.selectedProvider && availableModels.length > 0" class="context-stats">
      <label class="text-xs text-gray-400 mb-2 block">{{ t('settings.selectModel') }}</label>
      <select
        :value="selectedModel"
        @change="updateModel(($event.target as HTMLSelectElement).value)"
        class="input text-xs !py-2"
      >
        <option v-for="model in availableModels" :key="model" :value="model">
          {{ model }}
        </option>
      </select>
    </div>

    <!-- Host URL (for LocalAI, Qwen API) -->
    <div v-if="settings.selectedProvider === 'localai' || settings.selectedProvider === 'qwen'" class="context-stats">
      <label class="text-xs text-gray-400 mb-2 block">{{ t('settings.hostUrl') }}</label>
      <input
        type="text"
        :value="settings.selectedProvider === 'localai' ? settings.localAIHost : settings.qwenHost"
        @input="updateHost(($event.target as HTMLInputElement).value)"
        class="input text-xs !py-2"
        :placeholder="settings.selectedProvider === 'localai' ? 'http://localhost:8080' : 'https://dashscope.aliyuncs.com/compatible-mode/v1'"
      />
    </div>

    <!-- Save Button -->
    <button
      @click="saveSettings"
      :disabled="isSaving"
      class="btn-unified btn-unified-primary w-full"
    >
      <svg v-if="isSaving" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
      </svg>
      {{ isSaving ? t('settings.saving') : t('settings.save') }}
    </button>

    <!-- Status Message -->
    <div v-if="statusMessage" :class="['text-xs p-2.5 rounded-lg flex items-center gap-2', statusType === 'success' ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' : 'bg-red-500/20 text-red-300 border border-red-500/30']">
      <svg v-if="statusType === 'success'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
      </svg>
      <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
      {{ statusMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { apiService } from '@/services/api.service'
import { useAIStore } from '@/stores/ai.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, onMounted, reactive, ref } from 'vue'

const { t } = useI18n()
const uiStore = useUIStore()
const aiStore = useAIStore()

interface Settings {
  selectedProvider: string
  openAIAPIKey: string
  geminiAPIKey: string
  qwenAPIKey: string
  openRouterAPIKey: string
  localAIAPIKey: string
  localAIHost: string
  qwenHost: string
  selectedModels: Record<string, string>
  availableModels: Record<string, string[]>
}

const settings = reactive<Settings>({
  selectedProvider: '',
  openAIAPIKey: '',
  geminiAPIKey: '',
  qwenAPIKey: '',
  openRouterAPIKey: '',
  localAIAPIKey: '',
  localAIHost: 'http://localhost:8080',
  qwenHost: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
  selectedModels: {},
  availableModels: {}
})

const showApiKey = ref(false)
const isSaving = ref(false)
const statusMessage = ref('')
const statusType = ref<'success' | 'error'>('success')

const providers = [
  { id: 'openai', name: 'OpenAI', icon: '🤖', description: 'GPT-4o, GPT-4, GPT-3.5' },
  { id: 'gemini', name: 'Google Gemini', icon: '✨', description: 'Gemini Pro, Gemini Flash' },
  { id: 'qwen', name: 'Qwen (Alibaba)', icon: '🌐', description: 'Qwen-Max, Qwen-Plus (API)' },
  { id: 'qwen-cli', name: 'Qwen Code CLI', icon: '💻', description: 'Локальный CLI, без API ключа' },
  { id: 'openrouter', name: 'OpenRouter', icon: '🔀', description: 'Multiple providers' },
  { id: 'localai', name: 'LocalAI', icon: '🏠', description: 'Self-hosted models' },
]

const defaultModels: Record<string, string[]> = {
  openai: ['gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'gpt-4', 'gpt-3.5-turbo'],
  gemini: ['gemini-1.5-pro', 'gemini-1.5-flash', 'gemini-pro'],
  qwen: ['qwen-max', 'qwen-plus', 'qwen-turbo', 'qwen-long'],
  'qwen-cli': ['qwen-coder-plus-latest', 'qwen-coder-turbo-latest', 'qwen-turbo-latest'],
  openrouter: ['anthropic/claude-3.5-sonnet', 'openai/gpt-4o', 'google/gemini-pro'],
  localai: ['llama3', 'mistral', 'codellama']
}

const currentProviderName = computed(() => {
  const provider = providers.find(p => p.id === settings.selectedProvider)
  return provider?.name || ''
})

const currentProviderDescription = computed(() => {
  const provider = providers.find(p => p.id === settings.selectedProvider)
  return provider?.description || ''
})

const currentProviderHint = computed(() => {
  switch (settings.selectedProvider) {
    case 'openai': return 'Get key at platform.openai.com'
    case 'gemini': return 'Get key at aistudio.google.com'
    case 'qwen': return 'Get key at dashscope.console.aliyun.com'
    case 'openrouter': return 'Get key at openrouter.ai'
    case 'localai': return 'Optional for local models'
    default: return ''
  }
})

const currentApiKey = computed(() => {
  switch (settings.selectedProvider) {
    case 'openai': return settings.openAIAPIKey
    case 'gemini': return settings.geminiAPIKey
    case 'qwen': return settings.qwenAPIKey
    case 'openrouter': return settings.openRouterAPIKey
    case 'localai': return settings.localAIAPIKey
    default: return ''
  }
})

const availableModels = computed(() => {
  const provider = settings.selectedProvider
  return settings.availableModels[provider] || defaultModels[provider] || []
})

const selectedModel = computed(() => {
  return settings.selectedModels[settings.selectedProvider] || availableModels.value[0] || ''
})

function selectProvider(providerId: string) {
  settings.selectedProvider = providerId
}

function updateApiKey(value: string) {
  switch (settings.selectedProvider) {
    case 'openai': settings.openAIAPIKey = value; break
    case 'gemini': settings.geminiAPIKey = value; break
    case 'qwen': settings.qwenAPIKey = value; break
    case 'openrouter': settings.openRouterAPIKey = value; break
    case 'localai': settings.localAIAPIKey = value; break
  }
}

function clearApiKey() {
  updateApiKey('')
}

function updateModel(model: string) {
  settings.selectedModels[settings.selectedProvider] = model
}

function updateHost(value: string) {
  if (settings.selectedProvider === 'localai') {
    settings.localAIHost = value
  } else if (settings.selectedProvider === 'qwen') {
    settings.qwenHost = value
  }
}

async function loadSettings() {
  try {
    const dto = await apiService.getSettings()
    settings.selectedProvider = dto.selectedProvider || 'openai'
    settings.openAIAPIKey = dto.openAIAPIKey || ''
    settings.geminiAPIKey = dto.geminiAPIKey || ''
    settings.qwenAPIKey = dto.qwenAPIKey || ''
    settings.openRouterAPIKey = dto.openRouterAPIKey || ''
    settings.localAIAPIKey = dto.localAIAPIKey || ''
    settings.localAIHost = dto.localAIHost || 'http://localhost:8080'
    settings.qwenHost = dto.qwenHost || 'https://dashscope.aliyuncs.com/compatible-mode/v1'
    settings.selectedModels = dto.selectedModels || {}
    settings.availableModels = dto.availableModels || {}
  } catch (e) {
    console.error('Failed to load settings:', e)
  }
}

async function saveSettings() {
  isSaving.value = true
  statusMessage.value = ''
  
  try {
    const dto = await apiService.getSettings()
    
    // Update AI-related fields
    dto.selectedProvider = settings.selectedProvider
    dto.openAIAPIKey = settings.openAIAPIKey
    dto.geminiAPIKey = settings.geminiAPIKey
    dto.qwenAPIKey = settings.qwenAPIKey
    dto.openRouterAPIKey = settings.openRouterAPIKey
    dto.localAIAPIKey = settings.localAIAPIKey
    dto.localAIHost = settings.localAIHost
    dto.qwenHost = settings.qwenHost
    dto.selectedModels = settings.selectedModels
    
    await apiService.saveSettings(JSON.stringify(dto))
    
    // Update AI store with new model
    const selectedModel = settings.selectedModels[settings.selectedProvider] || ''
    aiStore.updateModel(settings.selectedProvider, selectedModel)
    
    statusMessage.value = t('settings.saved')
    statusType.value = 'success'
    uiStore.addToast(t('settings.saved'), 'success')
    
    setTimeout(() => { statusMessage.value = '' }, 3000)
  } catch (e) {
    const errorMsg = e instanceof Error ? e.message : 'Unknown error'
    statusMessage.value = errorMsg
    statusType.value = 'error'
    uiStore.addToast(t('settings.saveFailed'), 'error')
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>
