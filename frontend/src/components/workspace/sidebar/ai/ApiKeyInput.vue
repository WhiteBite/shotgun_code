<template>
  <div class="context-stats">
    <label class="text-xs text-gray-400 mb-2 block">
      {{ t('settings.apiKey') }} ({{ providerName }})
    </label>
    <div class="relative">
      <input
        :type="showKey ? 'text' : 'password'"
        :value="modelValue"
        @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
        class="input input-sm pr-16"
        :placeholder="t('settings.apiKeyPlaceholder')"
      />
      <div class="absolute right-2 top-1/2 -translate-y-1/2 flex gap-1">
        <button
          @click="$emit('toggle-visibility')"
          class="icon-btn-sm"
          :title="showKey ? t('settings.hideKey') : t('settings.showKey')"
        >
          <EyeOff v-if="showKey" class="w-3.5 h-3.5" />
          <Eye v-else class="w-3.5 h-3.5" />
        </button>
        <button
          v-if="modelValue"
          @click="$emit('clear')"
          class="icon-btn-sm icon-btn-danger"
          :title="t('settings.clearKey')"
        >
          <X class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
    <p v-if="hint" class="text-2xs text-gray-500 mt-1.5">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';
import { Eye, EyeOff, X } from 'lucide-vue-next';

const { t } = useI18n()

defineProps<{
  modelValue: string
  providerName: string
  showKey: boolean
  hint?: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
  'toggle-visibility': []
  'clear': []
}>()
</script>
