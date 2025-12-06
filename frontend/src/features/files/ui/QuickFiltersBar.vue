<template>
  <div class="flex items-center gap-2 px-3 py-2 border-b border-gray-700/30 overflow-x-auto overflow-y-visible">
    <span class="text-xs text-gray-500 whitespace-nowrap">Фильтры:</span>
    
    <div
      v-for="filter in enabledFilters"
      :key="filter.id"
      class="relative group"
      @mouseenter="handleMouseEnter(filter.id, $event)"
      @mouseleave="handleMouseLeave"
      @contextmenu.prevent="openFilterSettings(filter)"
    >
      <button
        @click="applyFilter(filter)"
        :aria-label="`Фильтр: ${filter.label}`"
        :class="['chip', activeFilter === filter.id ? 'chip-active' : 'chip-default']"
      >
        <span>{{ filter.label }}</span>
        <span 
          v-if="filter.extensions?.length || filter.patterns?.length"
          :class="['chip-count', activeFilter === filter.id ? 'chip-count-active' : 'chip-count-default']"
        >
          {{ getFilterCount(filter) }}
        </span>
      </button>
      
      <!-- Compact Tooltip with Teleport -->
      <Teleport to="body">
        <div
          v-if="hoveredFilter === filter.id"
          class="fixed z-[1070] px-3 py-2 bg-gray-800/95 backdrop-blur-sm border border-gray-700/50 rounded-xl shadow-2xl pointer-events-none max-w-[280px] animate-tooltip-in"
          :style="getTooltipStyle(filter.id)"
        >
          <div class="text-xs text-white font-semibold mb-1">{{ filter.label }}</div>
          <div class="text-xs text-gray-400 mb-2">{{ getShortDescription(filter) }}</div>
          
          <div v-if="filter.extensions" class="flex flex-wrap gap-1 mb-2">
            <span
              v-for="ext in filter.extensions.slice(0, 5)"
              :key="ext"
              class="px-1.5 py-0.5 bg-gray-700/50 rounded text-[10px] text-gray-300 font-mono"
            >
              {{ ext }}
            </span>
            <span v-if="filter.extensions.length > 5" class="text-[10px] text-gray-500">
              +{{ filter.extensions.length - 5 }}
            </span>
          </div>
          
          <div class="text-xs text-indigo-400">
            {{ getFilterCount(filter) }} files • Click to toggle
          </div>
        </div>
      </Teleport>
    </div>

    <button
      v-if="activeFilter"
      @click="clearFilter"
      class="icon-btn-sm icon-btn-danger !w-6 !h-6"
      title="Сбросить фильтр"
    >
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
    
    <button
      @click="showFilterSettingsModal = true"
      class="icon-btn-sm !w-6 !h-6"
      title="Настроить фильтры"
    >
      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
    </button>
  </div>
  
  <!-- Filter Settings Modal -->
  <Teleport to="body">
    <div v-if="showFilterSettingsModal" class="modal-overlay" @click.self="showFilterSettingsModal = false">
      <div class="modal-content !max-w-[500px] !p-0 overflow-hidden">
        <div class="section-header">
          <h3 class="section-title-text text-lg">Настройка фильтров</h3>
          <button @click="showFilterSettingsModal = false" class="icon-btn">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="p-4 overflow-y-auto max-h-[60vh] space-y-3">
          <div v-for="filter in allFilters" :key="filter.id" class="bg-gray-700/30 rounded-xl p-4 border border-gray-700/30">
            <div class="flex items-center justify-between mb-3">
              <label class="flex items-center gap-3">
                <div class="relative">
                  <input type="checkbox" v-model="filter.enabled" class="sr-only peer" @change="saveFilters" />
                  <div class="w-9 h-5 bg-gray-600 rounded-full peer-checked:bg-indigo-600 transition-colors"></div>
                  <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-4"></div>
                </div>
                <span class="text-white font-medium">{{ filter.label }}</span>
              </label>
              <span class="text-xs text-gray-500 bg-gray-800/50 px-2 py-1 rounded-lg">{{ getFilterCount(filter) }} файлов</span>
            </div>
            
            <div class="space-y-3 text-xs">
              <div>
                <label class="text-gray-400 block mb-1.5">Расширения (через запятую):</label>
                <input 
                  type="text" 
                  :value="filter.extensions?.join(', ')"
                  @change="updateFilterExtensions(filter, ($event.target as HTMLInputElement).value)"
                  class="w-full px-3 py-2 bg-gray-800/50 border border-gray-600/50 rounded-lg text-white text-xs focus:outline-none focus:border-indigo-500/50 transition-colors"
                  placeholder=".ts, .js, .vue"
                />
              </div>
              <div>
                <label class="text-gray-400 block mb-1.5">Паттерны (через запятую):</label>
                <input 
                  type="text" 
                  :value="filter.patterns?.join(', ')"
                  @change="updateFilterPatterns(filter, ($event.target as HTMLInputElement).value)"
                  class="w-full px-3 py-2 bg-gray-800/50 border border-gray-600/50 rounded-lg text-white text-xs focus:outline-none focus:border-indigo-500/50 transition-colors"
                  placeholder="**/test/**, **/*.spec.*"
                />
              </div>
            </div>
          </div>
        </div>
        
        <div class="p-4 border-t border-gray-700/50 flex justify-end gap-2">
          <button @click="resetFilters" class="btn btn-secondary btn-sm">
            Сбросить
          </button>
          <button @click="showFilterSettingsModal = false" class="btn btn-primary btn-sm">
            Готово
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useSettingsStore, type QuickFilterConfig } from '@/stores/settings.store'
import { computed, ref } from 'vue'
import { useFileStore } from '../model/file.store'

const fileStore = useFileStore()
const settingsStore = useSettingsStore()
const activeFilter = ref<string | null>(null)
const hoveredFilter = ref<string | null>(null)
const tooltipPosition = ref<Record<string, { x: number; y: number }>>({})
const showFilterSettingsModal = ref(false)

const allFilters = computed(() => settingsStore.settings.fileExplorer.quickFilters)
const enabledFilters = computed(() => allFilters.value.filter(f => f.enabled))

let showTimer: ReturnType<typeof setTimeout> | null = null

function handleMouseEnter(filterId: string, event: MouseEvent) {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  
  tooltipPosition.value[filterId] = {
    x: rect.left + rect.width / 2,
    y: rect.top - 8
  }
  
  if (showTimer) clearTimeout(showTimer)
  
  showTimer = setTimeout(() => {
    hoveredFilter.value = filterId
    showTimer = null
  }, 300)
}

function handleMouseLeave() {
  if (showTimer) {
    clearTimeout(showTimer)
    showTimer = null
  }
  hoveredFilter.value = null
}

function getTooltipStyle(filterId: string) {
  const pos = tooltipPosition.value[filterId]
  if (!pos) return {}
  
  let x = pos.x
  const tooltipWidth = 280
  
  if (x - tooltipWidth / 2 < 10) {
    x = tooltipWidth / 2 + 10
  } else if (x + tooltipWidth / 2 > window.innerWidth - 10) {
    x = window.innerWidth - tooltipWidth / 2 - 10
  }
  
  return {
    left: `${x}px`,
    bottom: `${window.innerHeight - pos.y}px`,
    transform: 'translateX(-50%)'
  }
}

function getShortDescription(filter: QuickFilterConfig): string {
  const parts: string[] = []
  if (filter.extensions?.length) parts.push(`${filter.extensions.length} расширений`)
  if (filter.patterns?.length) parts.push(`${filter.patterns.length} паттернов`)
  return parts.join(', ') || 'Не настроен'
}

function getFilterCount(filter: QuickFilterConfig): number {
  let count = 0
  const countFiles = (nodes: any[]) => {
    nodes.forEach(node => {
      if (!node.isDir) {
        if (filter.extensions?.some(ext => node.name.endsWith(ext))) {
          count++
        } else if (filter.patterns?.some(pattern => matchPattern(node.relPath || node.path, pattern))) {
          count++
        }
      }
      if (node.children) countFiles(node.children)
    })
  }
  countFiles(fileStore.nodes)
  return count
}

function matchPattern(path: string, pattern: string): boolean {
  const regexPattern = pattern.replace(/\*\*/g, '.*').replace(/\*/g, '[^/]*').replace(/\?/g, '.')
  try {
    return new RegExp(regexPattern, 'i').test(path)
  } catch {
    return false
  }
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const { t } = useI18n()

function applyFilter(filter: QuickFilterConfig) {
  if (activeFilter.value === filter.id) {
    clearFilter()
    return
  }
  activeFilter.value = filter.id
  if (filter.extensions?.length) {
    fileStore.setFilterExtensions(filter.extensions)
  }
}

function clearFilter() {
  activeFilter.value = null
  fileStore.setFilterExtensions([])
}

function openFilterSettings(_filter: QuickFilterConfig) {
  showFilterSettingsModal.value = true
}

function saveFilters() {}

function updateFilterExtensions(filter: QuickFilterConfig, value: string) {
  filter.extensions = value.split(',').map(s => s.trim()).filter(s => s)
}

function updateFilterPatterns(filter: QuickFilterConfig, value: string) {
  filter.patterns = value.split(',').map(s => s.trim()).filter(s => s)
}

function resetFilters() {
  settingsStore.settings.fileExplorer.quickFilters = [
    { id: 'source', label: 'Исходный код', extensions: ['.ts', '.js', '.tsx', '.jsx', '.vue', '.go', '.py', '.java', '.cpp', '.c', '.rs'], patterns: [], enabled: true },
    { id: 'tests', label: 'Тесты', extensions: [], patterns: ['**/*.test.*', '**/*.spec.*', '**/test/**', '**/tests/**', '**/Test/**'], enabled: true },
    { id: 'config', label: 'Конфигурация', extensions: ['.json', '.yaml', '.yml', '.toml', '.ini', '.env'], patterns: [], enabled: true },
    { id: 'docs', label: 'Документация', extensions: ['.md', '.txt', '.rst', '.adoc'], patterns: [], enabled: true },
    { id: 'styles', label: 'Стили', extensions: ['.css', '.scss', '.sass', '.less'], patterns: [], enabled: true }
  ]
}
</script>
