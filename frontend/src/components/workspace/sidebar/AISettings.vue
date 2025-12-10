<template>
  <div class="space-y-4">
    <!-- Section Header -->
    <div class="flex items-center gap-2">
      <div class="section-icon section-icon-purple">
        <Lightbulb class="w-4 h-4" />
      </div>
      <span class="text-sm font-semibold text-white">{{ t('settings.aiProvider') }}</span>
    </div>

    <!-- Provider Selection -->
    <ProviderSelector
      :model-value="settings.selectedProvider"
      :providers="providers"
      :description="currentProvider?.description"
      @update:model-value="selectProvider"
    />

    <!-- API Key Input (not for CLI) -->
    <ApiKeyInput
      v-if="needsApiKey"
      :model-value="currentApiKey"
      :provider-name="currentProvider?.name || ''"
      :show-key="showApiKey"
      :hint="currentProviderHint"
      @update:model-value="updateApiKey"
      @toggle-visibility="toggleShowApiKey"
      @clear="clearApiKey"
    />

    <!-- Qwen CLI Info -->
    <QwenCliInfo v-if="settings.selectedProvider === 'qwen-cli'" />

    <!-- Model Selection -->
    <ModelSelector
      v-if="settings.selectedProvider && availableModels.length > 0"
      :model-value="selectedModel"
      :models="availableModels"
      @update:model-value="updateModel"
    />

    <!-- Host URL (for LocalAI, Qwen API) -->
    <HostUrlInput
      v-if="needsHostUrl"
      :model-value="currentHostUrl"
      :placeholder="hostPlaceholder"
      @update:model-value="updateHost"
    />

    <!-- Save Button -->
    <button
      @click="saveSettings"
      :disabled="isSaving"
      class="btn-unified btn-unified-primary w-full"
    >
      <Loader2 v-if="isSaving" class="animate-spin w-4 h-4" />
      {{ isSaving ? t('settings.saving') : t('settings.save') }}
    </button>

    <!-- Status Message -->
    <StatusMessage :message="statusMessage" :type="statusType" />
  </div>
</template>

<script setup lang="ts">
import { useAISettings } from '@/composables/useAISettings'
import { useI18n } from '@/composables/useI18n'
import { Lightbulb, Loader2 } from 'lucide-vue-next'
import { onMounted } from 'vue'
import {
    ApiKeyInput,
    HostUrlInput,
    ModelSelector,
    ProviderSelector,
    QwenCliInfo,
    StatusMessage
} from './ai'

const { t } = useI18n()

const {
  settings,
  showApiKey,
  isSaving,
  statusMessage,
  statusType,
  providers,
  currentProvider,
  currentProviderHint,
  currentApiKey,
  availableModels,
  selectedModel,
  needsApiKey,
  needsHostUrl,
  currentHostUrl,
  hostPlaceholder,
  selectProvider,
  updateApiKey,
  clearApiKey,
  updateModel,
  updateHost,
  toggleShowApiKey,
  loadSettings,
  saveSettings
} = useAISettings()

onMounted(() => {
  loadSettings()
})
</script>
