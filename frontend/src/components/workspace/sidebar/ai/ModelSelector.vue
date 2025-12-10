<template>
  <div class="context-stats">
    <label class="text-xs text-gray-400 mb-2 block">{{ t('settings.selectModel') }}</label>
    <select
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
      class="input input-sm"
    >
      <option v-for="model in models" :key="model" :value="model">
        {{ model }}
      </option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';

const { t } = useI18n()

defineProps<{
  modelValue: string
  models: string[]
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
