<template>
  <div class="filters-bar">
    <span class="filters-label">Фильтры:</span>

    <div v-for="filter in enabledFilters" :key="filter.id" class="relative group"
      @mouseenter="handleMouseEnter(filter.id, $event)" @mouseleave="handleMouseLeave"
      @contextmenu.prevent="openFilterSettings(filter)">
      <button @click="applyFilter(filter)" :aria-label="`Фильтр: ${filter.label}`"
        :class="['chip', activeFilter === filter.id ? 'chip-active' : 'chip-default']">
        <span>{{ filter.label }}</span>
        <span v-if="filter.extensions?.length || filter.patterns?.length"
          :class="['chip-count', activeFilter === filter.id ? 'chip-count-active' : 'chip-count-default']">
          {{ getFilterCount(filter) }}
        </span>
      </button>

      <!-- Compact Tooltip with Teleport -->
      <Teleport to="body">
        <div v-if="hoveredFilter === filter.id" class="filter-tooltip animate-tooltip-in"
          :style="getTooltipStyle(filter.id)">
          <div class="tooltip-title">{{ filter.label }}</div>
          <div class="tooltip-desc">{{ getShortDescription(filter) }}</div>

          <div v-if="filter.extensions" class="tooltip-extensions">
            <span v-for="ext in filter.extensions.slice(0, 5)" :key="ext" class="tooltip-ext">
              {{ ext }}
            </span>
            <span v-if="filter.extensions.length > 5" class="tooltip-more">
              +{{ filter.extensions.length - 5 }}
            </span>
          </div>

          <div class="tooltip-hint">
            {{ getFilterCount(filter) }} files • Click to toggle
          </div>
        </div>
      </Teleport>
    </div>

    <button v-if="activeFilter" @click="clearFilter" class="icon-btn-sm icon-btn-danger !w-6 !h-6"
      title="Сбросить фильтр">
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>

    <button @click="showFilterSettingsModal = true" class="icon-btn-sm !w-6 !h-6" title="Настроить фильтры">
      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
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

        <div class="modal-body">
          <div v-for="filter in allFilters" :key="filter.id" class="filter-card">
            <div class="filter-card-header">
              <label class="toggle-label">
                <div class="toggle-container">
                  <input type="checkbox" v-model="filter.enabled" class="sr-only peer" @change="saveFilters" />
                  <div class="toggle-track"></div>
                  <div class="toggle-thumb"></div>
                </div>
                <span class="toggle-text-lg">{{ filter.label }}</span>
              </label>
              <span class="filter-count-badge">{{ getFilterCount(filter) }} файлов</span>
            </div>

            <div class="filter-inputs">
              <div>
                <label class="input-label">Расширения (через запятую):</label>
                <input type="text" :value="filter.extensions?.join(', ')"
                  @change="updateFilterExtensions(filter, ($event.target as HTMLInputElement).value)"
                  class="input input-sm" placeholder=".ts, .js, .vue" />
              </div>
              <div>
                <label class="input-label">Паттерны (через запятую):</label>
                <input type="text" :value="filter.patterns?.join(', ')"
                  @change="updateFilterPatterns(filter, ($event.target as HTMLInputElement).value)"
                  class="input input-sm" placeholder="**/test/**, **/*.spec.*" />
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
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

interface FileNode {
  name: string
  path: string
  relPath?: string
  isDir: boolean
  children?: FileNode[]
}

function getFilterCount(filter: QuickFilterConfig): number {
  let count = 0
  const countFiles = (nodes: FileNode[]) => {
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
  countFiles(fileStore.nodes as FileNode[])
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

// Локализация доступна через useI18n() при необходимости

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

function saveFilters() { }

function updateFilterExtensions(filter: QuickFilterConfig, value: string) {
  filter.extensions = value.split(',').map(s => s.trim()).filter(s => s)
}

function updateFilterPatterns(filter: QuickFilterConfig, value: string) {
  filter.patterns = value.split(',').map(s => s.trim()).filter(s => s)
}

function resetFilters() {
  // Используем resetToDefaults из store для сброса фильтров
  settingsStore.resetToDefaults()
}
</script>


<style scoped>
.filters-bar {
  @apply flex items-center gap-2 px-3 py-2 overflow-x-auto overflow-y-visible;
  border-bottom: 1px solid var(--border-default);
}

.filters-label {
  @apply text-xs whitespace-nowrap font-medium;
  color: var(--text-secondary);
}

/* Tooltip styles */
.filter-tooltip {
  @apply fixed z-[1070] px-3 py-2 rounded-xl pointer-events-none max-w-[280px];
  background: var(--bg-1);
  border: 1px solid var(--border-default);
  box-shadow: var(--shadow-xl);
  backdrop-filter: blur(8px);
}

.tooltip-title {
  @apply text-xs font-semibold mb-1;
  color: var(--text-primary);
}

.tooltip-desc {
  @apply text-xs mb-2;
  color: var(--text-secondary);
}

.tooltip-extensions {
  @apply flex flex-wrap gap-1 mb-2;
}

.tooltip-ext {
  @apply px-1.5 py-0.5 rounded text-[10px] font-mono;
  background: var(--bg-2);
  color: var(--text-secondary);
}

.tooltip-more {
  @apply text-[10px];
  color: var(--text-muted);
}

.tooltip-hint {
  @apply text-xs;
  color: var(--accent-indigo);
}

/* Modal styles */
.modal-body {
  @apply p-4 overflow-y-auto max-h-[60vh] space-y-3;
}

.filter-card {
  @apply rounded-xl p-4;
  background: var(--bg-2);
  border: 1px solid var(--border-default);
}

.filter-card-header {
  @apply flex items-center justify-between mb-3;
}

.toggle-label {
  @apply flex items-center gap-3 cursor-pointer;
}

.toggle-container {
  @apply relative w-9 h-5;
}

.toggle-track {
  @apply absolute inset-0 rounded-full;
  background: var(--bg-3);
  transition: all 200ms ease-out;
}

.peer:checked~.toggle-track {
  background: var(--accent-indigo);
}

.toggle-thumb {
  @apply absolute left-0.5 top-0.5 w-4 h-4 rounded-full bg-white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  transition: transform 200ms ease-out;
}

.peer:checked~.toggle-thumb {
  transform: translateX(16px);
}

.toggle-text-lg {
  @apply font-medium;
  color: var(--text-primary);
}

.filter-count-badge {
  @apply text-xs px-2 py-1 rounded-lg;
  background: var(--bg-3);
  color: var(--text-secondary);
}

.filter-inputs {
  @apply space-y-3 text-xs;
}

.input-label {
  @apply block mb-1.5;
  color: var(--text-secondary);
}

.input-sm {
  @apply text-xs !py-1.5;
}

.modal-footer {
  @apply p-4 flex justify-end gap-2;
  border-top: 1px solid var(--border-default);
}
</style>
