<template>
  <div class="space-y-4">
    <p class="text-xs font-semibold text-gray-400">{{ t('sidebar.exportSettings') }}</p>

    <!-- Format Selection -->
    <div class="context-stats">
      <label class="text-xs text-gray-400 mb-2 block">{{ t('export.format') }}</label>
      <div class="grid grid-cols-2 gap-2">
        <button
          v-for="format in formats"
          :key="format.id"
          @click="updateFormat(format.id)"
          :class="[
            'btn-unified text-xs',
            exportSettings.exportFormat === format.id
              ? 'btn-unified-primary'
              : 'btn-unified-secondary'
          ]"
        >
          {{ format.label }}
        </button>
      </div>
    </div>

    <!-- Options -->
    <div class="context-stats space-y-3">
      <label class="toggle-label">
        <div class="toggle-container">
          <input
            type="checkbox"
            class="sr-only peer"
            :checked="exportSettings.includeManifest"
            @change="updateSetting('includeManifest', !exportSettings.includeManifest)"
          />
          <div class="toggle-track"></div>
          <div class="toggle-thumb"></div>
        </div>
        <span class="toggle-text">{{ t('export.includeMetadata') }}</span>
      </label>

      <label class="toggle-label">
        <div class="toggle-container">
          <input
            type="checkbox"
            class="sr-only peer"
            :checked="exportSettings.includeLineNumbers"
            @change="updateSetting('includeLineNumbers', !exportSettings.includeLineNumbers)"
          />
          <div class="toggle-track"></div>
          <div class="toggle-thumb"></div>
        </div>
        <span class="toggle-text">{{ t('export.includeLineNumbers') }}</span>
      </label>

      <label class="toggle-label">
        <div class="toggle-container">
          <input
            type="checkbox"
            class="sr-only peer"
            :checked="exportSettings.stripComments"
            @change="updateSetting('stripComments', !exportSettings.stripComments)"
          />
          <div class="toggle-track"></div>
          <div class="toggle-thumb"></div>
        </div>
        <span class="toggle-text">{{ t('context.stripComments') }}</span>
      </label>
    </div>

    <!-- AI Chunking Options -->
    <div class="context-stats space-y-3">
      <label class="toggle-label">
        <div class="toggle-container">
          <input
            type="checkbox"
            class="sr-only peer"
            :checked="exportSettings.enableAutoSplit"
            @change="updateSetting('enableAutoSplit', !exportSettings.enableAutoSplit)"
          />
          <div class="toggle-track"></div>
          <div class="toggle-thumb"></div>
        </div>
        <span class="toggle-text">Авто-разбиение на чанки</span>
      </label>

      <div v-if="exportSettings.enableAutoSplit" class="space-y-2 pl-2 border-l-2 border-indigo-500/30">
        <div>
          <label class="text-[10px] text-gray-500 mb-1 block">Токенов на чанк</label>
          <input
            type="number"
            :value="exportSettings.maxTokensPerChunk"
            @input="updateSetting('maxTokensPerChunk', Number(($event.target as HTMLInputElement).value))"
            min="500"
            max="32000"
            class="input text-xs !py-1.5"
          />
        </div>
        <div>
          <label class="text-[10px] text-gray-500 mb-1 block">Стратегия</label>
          <select
            :value="exportSettings.splitStrategy"
            @change="updateSetting('splitStrategy', ($event.target as HTMLSelectElement).value as 'smart' | 'file' | 'token')"
            class="input text-xs !py-1.5"
          >
            <option value="smart">Smart</option>
            <option value="file">По файлам</option>
            <option value="token">По токенам</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Token Limit -->
    <div class="context-stats">
      <div class="flex items-center justify-between mb-2">
        <label class="text-xs text-gray-400">Лимит токенов</label>
        <span class="text-xs text-indigo-400">{{ formatNumber(exportSettings.tokenLimit) }}</span>
      </div>
      <input
        type="range"
        :value="exportSettings.tokenLimit"
        @input="updateSetting('tokenLimit', Number(($event.target as HTMLInputElement).value))"
        min="1000"
        max="128000"
        step="1000"
        class="w-full h-1.5 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-indigo-500"
      />
      <div class="flex justify-between text-[10px] text-gray-600 mt-1">
        <span>1K</span>
        <span>128K</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useExport, type ExportSettings } from '@/composables/useExport'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()
const { settings: exportSettings } = useExport()

type FormatType = 'plain' | 'manifest' | 'json'

const formats: { id: FormatType; label: string }[] = [
  { id: 'manifest', label: 'Markdown' },
  { id: 'plain', label: 'Plain Text' },
  { id: 'json', label: 'JSON' },
]

function updateFormat(format: FormatType) {
  exportSettings.exportFormat = format
  saveToLocalStorage()
}

function updateSetting<K extends keyof ExportSettings>(key: K, value: ExportSettings[K]) {
  (exportSettings as unknown)[key] = value
  saveToLocalStorage()
}

function saveToLocalStorage() {
  try {
    localStorage.setItem('export-settings', JSON.stringify({
      exportFormat: exportSettings.exportFormat,
      stripComments: exportSettings.stripComments,
      includeManifest: exportSettings.includeManifest,
      includeLineNumbers: exportSettings.includeLineNumbers,
      enableAutoSplit: exportSettings.enableAutoSplit,
      maxTokensPerChunk: exportSettings.maxTokensPerChunk,
      splitStrategy: exportSettings.splitStrategy,
      tokenLimit: exportSettings.tokenLimit
    }))
  } catch (e) {
    console.warn('Failed to save export settings:', e)
  }
}

// Load settings on init
function loadFromLocalStorage() {
  try {
    const saved = localStorage.getItem('export-settings')
    if (saved) {
      const parsed = JSON.parse(saved)
      Object.assign(exportSettings, parsed)
    }
  } catch (e) {
    console.warn('Failed to load export settings:', e)
  }
}

loadFromLocalStorage()

function formatNumber(num: number): string {
  if (num >= 1000) return `${(num / 1000).toFixed(0)}K`
  return num.toString()
}
</script>

<style scoped>
.toggle-label {
  @apply flex items-center gap-3 cursor-pointer;
  color: var(--text-secondary);
}

.toggle-label:hover .toggle-text {
  color: var(--text-primary);
}

.toggle-container {
  @apply relative w-9 h-5;
}

.toggle-track {
  @apply absolute inset-0 rounded-full;
  background: var(--bg-3);
  transition: all 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.toggle-container:hover .toggle-track {
  background: rgba(148, 163, 184, 0.35);
}

.peer:checked ~ .toggle-track {
  background: var(--accent-indigo);
  box-shadow: 0 0 12px rgba(99, 102, 241, 0.4);
}

.toggle-thumb {
  @apply absolute left-0.5 top-0.5 w-4 h-4 rounded-full bg-white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  transition: transform 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.peer:checked ~ .toggle-thumb {
  transform: translateX(16px);
}

.toggle-text {
  @apply text-xs;
  transition: color 150ms ease-out;
}
</style>
