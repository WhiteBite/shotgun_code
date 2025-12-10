<template>
  <div class="space-y-3">
    <!-- Format Selection -->
    <CollapsibleSection 
      :label="t('export.section.format')" 
      :icon="FileText" 
      icon-color="indigo"
    >
      <FormatSelector
        :model-value="settings.outputFormat"
        @update:model-value="update('outputFormat', $event)"
      />
    </CollapsibleSection>

    <!-- Output Options -->
    <CollapsibleSection 
      :label="t('export.section.output')" 
      :icon="Settings2" 
      icon-color="purple"
    >
      <OutputOptions
        :include-manifest="settings.includeManifest"
        :include-line-numbers="settings.includeLineNumbers"
        :strip-comments="settings.stripComments"
        @update:include-manifest="update('includeManifest', $event)"
        @update:include-line-numbers="update('includeLineNumbers', $event)"
        @update:strip-comments="update('stripComments', $event)"
      />
    </CollapsibleSection>

    <!-- Content Optimization -->
    <CollapsibleSection 
      :label="t('export.section.optimization')" 
      :icon="Sparkles" 
      icon-color="emerald"
    >
      <OptimizationOptions
        :exclude-tests="settings.excludeTests"
        :strip-license="settings.stripLicense"
        :compact-data-files="settings.compactDataFiles"
        :trim-whitespace="settings.trimWhitespace"
        :collapse-empty-lines="settings.collapseEmptyLines"
        @update:exclude-tests="update('excludeTests', $event)"
        @update:strip-license="update('stripLicense', $event)"
        @update:compact-data-files="update('compactDataFiles', $event)"
        @update:trim-whitespace="update('trimWhitespace', $event)"
        @update:collapse-empty-lines="update('collapseEmptyLines', $event)"
      />
    </CollapsibleSection>

    <!-- AI Chunking Options -->
    <CollapsibleSection 
      :label="t('export.section.chunking')" 
      :icon="Layers" 
      icon-color="orange"
    >
      <ChunkingOptions
        :enable-auto-split="settings.enableAutoSplit"
        :max-tokens-per-chunk="settings.maxTokensPerChunk"
        :split-strategy="settings.splitStrategy"
        @update:enable-auto-split="update('enableAutoSplit', $event)"
        @update:max-tokens-per-chunk="update('maxTokensPerChunk', $event)"
        @update:split-strategy="update('splitStrategy', $event)"
      />
    </CollapsibleSection>

    <!-- Token Limit -->
    <CollapsibleSection 
      :label="t('export.section.tokenLimit')" 
      :icon="Gauge" 
      icon-color="pink"
    >
      <TokenLimitSlider
        :model-value="settings.maxTokens"
        @update:model-value="update('maxTokens', $event)"
      />
    </CollapsibleSection>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useSettingsStore, type ContextSettings } from '@/stores/settings.store'
import { FileText, Gauge, Layers, Settings2, Sparkles } from 'lucide-vue-next'
import { computed } from 'vue'
import {
    ChunkingOptions,
    CollapsibleSection,
    FormatSelector,
    OptimizationOptions,
    OutputOptions,
    TokenLimitSlider
} from './export'

const { t } = useI18n()
const settingsStore = useSettingsStore()

const settings = computed(() => settingsStore.settings.context)

function update<K extends keyof ContextSettings>(key: K, value: ContextSettings[K]) {
  settingsStore.updateContextSettings({ [key]: value })
}
</script>
