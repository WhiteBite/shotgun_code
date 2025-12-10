<template>
  <div class="context-stats space-y-3">
    <SettingsToggle
      :model-value="includeManifest"
      :label="t('export.includeMetadata')"
      @update:model-value="$emit('update:includeManifest', $event)"
    />
    <SettingsToggle
      :model-value="includeLineNumbers"
      :label="t('export.includeLineNumbers')"
      @update:model-value="$emit('update:includeLineNumbers', $event)"
    />
    <SettingsToggle
      :model-value="stripComments"
      :label="t('context.stripComments')"
      @update:model-value="$emit('update:stripComments', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';
import SettingsToggle from './SettingsToggle.vue';

const { t } = useI18n()

defineProps<{
  includeManifest: boolean
  includeLineNumbers: boolean
  stripComments: boolean
}>()

defineEmits<{
  'update:includeManifest': [value: boolean]
  'update:includeLineNumbers': [value: boolean]
  'update:stripComments': [value: boolean]
}>()
</script>
