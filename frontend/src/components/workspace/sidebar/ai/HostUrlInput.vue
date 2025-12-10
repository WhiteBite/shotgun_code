<template>
  <div class="context-stats">
    <label class="text-xs text-gray-400 mb-2 block">{{ t('settings.hostUrl') }}</label>
    <input
      type="text"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      class="input input-sm"
      :placeholder="placeholder"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';

const { t } = useI18n()

defineProps<{
  modelValue: string
  placeholder?: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
