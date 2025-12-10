<template>
  <div class="space-y-3">
    <SettingsToggle
      :model-value="enableAutoSplit"
      :label="t('export.autoSplit')"
      @update:model-value="$emit('update:enableAutoSplit', $event)"
    />

    <div v-if="enableAutoSplit" class="space-y-3">
      <!-- Settings -->
      <div class="space-y-2 pl-2 border-l-accent-indigo">
        <div>
          <label class="text-2xs text-gray-500 mb-1 block">{{ t('export.tokensPerChunk') }}</label>
          <input 
            type="number" 
            :value="maxTokensPerChunk"
            @input="$emit('update:maxTokensPerChunk', Number(($event.target as HTMLInputElement).value))" 
            min="500"
            max="32000" 
            class="input input-xs" 
          />
        </div>
        <div>
          <label class="text-2xs text-gray-500 mb-1 block">{{ t('export.strategy') }}</label>
          <select 
            :value="splitStrategy"
            @change="$emit('update:splitStrategy', ($event.target as HTMLSelectElement).value as SplitStrategy)"
            class="input input-xs"
          >
            <option value="smart">{{ t('export.strategySmart') }}</option>
            <option value="file">{{ t('export.strategyFile') }}</option>
            <option value="token">{{ t('export.strategyToken') }}</option>
          </select>
        </div>
      </div>

      <!-- Chunk Copy UI -->
      <ChunkCopyBar />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import ChunkCopyBar from './ChunkCopyBar.vue'
import SettingsToggle from './SettingsToggle.vue'

const { t } = useI18n()

export type SplitStrategy = 'smart' | 'file' | 'token'

defineProps<{
  enableAutoSplit: boolean
  maxTokensPerChunk: number
  splitStrategy: SplitStrategy
}>()

defineEmits<{
  'update:enableAutoSplit': [value: boolean]
  'update:maxTokensPerChunk': [value: number]
  'update:splitStrategy': [value: SplitStrategy]
}>()
</script>
