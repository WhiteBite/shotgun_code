<template>
  <div class="context-stats space-y-3">
    <label class="text-xs text-gray-400 mb-2 block">{{ t('context.optimization') }}</label>
    
    <SettingsToggle
      :model-value="excludeTests"
      :label="t('context.excludeTests')"
      @update:model-value="$emit('update:excludeTests', $event)"
    />
    <SettingsToggle
      :model-value="stripLicense"
      :label="t('context.stripLicense')"
      @update:model-value="$emit('update:stripLicense', $event)"
    />
    <SettingsToggle
      :model-value="compactDataFiles"
      :label="t('context.compactDataFiles')"
      @update:model-value="$emit('update:compactDataFiles', $event)"
    />
    <SettingsToggle
      :model-value="trimWhitespace"
      :label="t('context.trimWhitespace')"
      @update:model-value="$emit('update:trimWhitespace', $event)"
    />
    <SettingsToggle
      :model-value="collapseEmptyLines"
      :label="t('context.collapseEmptyLines')"
      @update:model-value="$emit('update:collapseEmptyLines', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';
import SettingsToggle from './SettingsToggle.vue';

const { t } = useI18n()

defineProps<{
  excludeTests: boolean
  stripLicense: boolean
  compactDataFiles: boolean
  trimWhitespace: boolean
  collapseEmptyLines: boolean
}>()

defineEmits<{
  'update:excludeTests': [value: boolean]
  'update:stripLicense': [value: boolean]
  'update:compactDataFiles': [value: boolean]
  'update:trimWhitespace': [value: boolean]
  'update:collapseEmptyLines': [value: boolean]
}>()
</script>
