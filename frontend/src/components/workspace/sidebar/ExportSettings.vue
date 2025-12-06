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
            'px-3 py-2 text-xs rounded-lg border transition-all',
            exportSettings.exportFormat === format.id
              ? 'bg-indigo-500/20 border-indigo-500/50 text-indigo-300'
              : 'bg-gray-800/50 border-gray-700/50 text-gray-400 hover:border-gray-600/50'
          ]"
        >
          {{ format.label }}
        </button>
      </div>
    </div>

    <!-- Options -->
    <div class="context-stats space-y-3">
      <label class="flex items-center gap-3 text-sm text-gray-300 cursor-pointer group">
        <div class="relative">
          <input
            type="checkbox"
            class="sr-only peer"
            :checked="exportSettings.includeManifest"
            @change="updateSetting('includeManifest', !exportSettings.includeManifest)"
          />
          <div class="w-9 h-5 bg-gray-700 rounded-full peer-checked:bg-indigo-600 transition-colors"></div>
          <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-4"></div>
        </div>
        <span class="text-xs group-hover:text-white transition-colors">{{ t('export.includeMetadata') }}</span>
      </label>

      <label class="flex items-center gap-3 text-sm text-gray-300 cursor-pointer group">
        <div class="relative">
          <input
            type="checkbox"
            class="sr-only peer"
            :checked="exportSettings.includeLineNumbers"
            @change="updateSetting('includeLineNumbers', !exportSettings.includeLineNumbers)"
          />
          <div class="w-9 h-5 bg-gray-700 rounded-full peer-checked:bg-indigo-600 transition-colors"></div>
          <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-4"></div>
        </div>
        <span class="text-xs group-hover:text-white transition-colors">{{ t('export.includeLineNumbers') }}</span>
      </label>

      <label class="flex items-center gap-3 text-sm text-gray-300 cursor-pointer group">
        <div class="relative">
          <input
            type="checkbox"
            class="sr-only peer"
            :checked="exportSettings.stripComments"
            @change="updateSetting('stripComments', !exportSettings.stripComments)"
          />
          <div class="w-9 h-5 bg-gray-700 rounded-full peer-checked:bg-indigo-600 transition-colors"></div>
          <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-4"></div>
        </div>
        <span class="text-xs group-hover:text-white transition-colors">{{ t('context.stripComments') }}</span>
      </label>
    </div>

    <!-- AI Chunking Options -->
    <div class="context-stats space-y-3">
      <label class="flex items-center gap-3 text-sm text-gray-300 cursor-pointer group">
        <div class="relative">
          <input
            type="checkbox"
            class="sr-only peer"
            :checked="exportSettings.enableAutoSplit"
            @change="updateSetting('enableAutoSplit', !exportSettings.enableAutoSplit)"
          />
          <div class="w-9 h-5 bg-gray-700 rounded-full peer-checked:bg-indigo-600 transition-colors"></div>
          <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-4"></div>
        </div>
        <span class="text-xs group-hover:text-white transition-colors">Авто-разбиение на чанки</span>
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
  (exportSettings as any)[key] = value
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
