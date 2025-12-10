<template>
  <div class="filters-bar">
    <!-- Filter Groups -->
    <div class="filter-groups">
      <!-- Types Dropdown -->
      <div class="filter-dropdown" ref="typesDropdownRef">
        <button @click="toggleDropdown('types')"
          :class="['filter-dropdown-trigger', { 'filter-dropdown-trigger-active': hasActiveTypeFilters }]">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
          </svg>
          <span>{{ t('quickFilters.types') }}</span>
          <svg class="w-3 h-3 transition-transform" :class="{ 'rotate-180': openDropdown === 'types' }" fill="none"
            stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <Teleport to="body">
          <Transition name="dropdown">
            <div v-if="openDropdown === 'types'" class="filter-dropdown-menu" :style="dropdownStyle">
              <div class="filter-dropdown-header">
                <span>{{ t('quickFilters.fileTypes') }}</span>
                <button v-if="activeTypeFilters.length > 0" @click="clearTypeFilters" class="filter-clear-btn">
                  {{ t('quickFilters.clearAll') }}
                </button>
              </div>
              <div class="filter-dropdown-items">
                <button v-for="filter in typeFilters" :key="filter.id" @click="toggleFilter(filter, $event)" :class="['filter-dropdown-item', {
                  'filter-dropdown-item-active': isFilterActive(filter.id),
                  'filter-dropdown-item-excluded': isFilterExcluded(filter.id)
                }]">
                  <div class="filter-item-left">
                    <div :class="['filter-item-icon', `filter-item-icon-${filter.category}`]">
                      <component :is="getFilterIcon(filter.category)" class="w-3.5 h-3.5" />
                    </div>
                    <span :class="['filter-item-label', { 'filter-item-label-excluded': isFilterExcluded(filter.id) }]">
                      {{ filter.label }}
                    </span>
                  </div>
                  <div class="filter-item-right">
                    <span class="filter-item-count">{{ getFilterCount(filter) }}</span>
                    <div class="filter-item-bar">
                      <div class="filter-item-bar-fill" :style="{ width: getFilterPercentage(filter) + '%' }"></div>
                    </div>
                  </div>
                </button>
              </div>
              <div class="filter-dropdown-hint">
                <kbd>Ctrl</kbd> {{ t('quickFilters.multiSelect') }} · <kbd>Shift</kbd> {{ t('quickFilters.exclude') }}
              </div>
            </div>
          </Transition>
        </Teleport>
      </div>

      <!-- Languages Dropdown -->
      <div v-if="languageFilters.length > 0" class="filter-dropdown" ref="langsDropdownRef">
        <button @click="toggleDropdown('langs')"
          :class="['filter-dropdown-trigger', { 'filter-dropdown-trigger-active': hasActiveLanguageFilters }]">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
          </svg>
          <span>{{ t('quickFilters.languages') }}</span>
          <span class="filter-dropdown-badge">{{ languageFilters.length }}</span>
          <svg class="w-3 h-3 transition-transform" :class="{ 'rotate-180': openDropdown === 'langs' }" fill="none"
            stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <Teleport to="body">
          <Transition name="dropdown">
            <div v-if="openDropdown === 'langs'" class="filter-dropdown-menu" :style="dropdownStyle">
              <div class="filter-dropdown-header">
                <span>{{ t('quickFilters.projectLanguages') }}</span>
                <button v-if="activeLanguageFilters.length > 0" @click="clearLanguageFilters" class="filter-clear-btn">
                  {{ t('quickFilters.clearAll') }}
                </button>
              </div>
              <div class="filter-dropdown-items">
                <button v-for="filter in languageFilters" :key="filter.id" @click="toggleFilter(filter, $event)"
                  :class="['filter-dropdown-item', { 'filter-dropdown-item-active': isFilterActive(filter.id) }]">
                  <div class="filter-item-left">
                    <span class="filter-item-emoji">{{ filter.icon }}</span>
                    <span class="filter-item-label">{{ filter.language }}</span>
                    <span v-if="filter.primary" class="filter-item-primary">★</span>
                  </div>
                  <div class="filter-item-right">
                    <span class="filter-item-count">{{ filter.fileCount }}</span>
                    <div class="filter-item-bar">
                      <div class="filter-item-bar-fill filter-item-bar-lang"
                        :style="{ width: filter.percentage + '%' }"></div>
                    </div>
                  </div>
                </button>
              </div>
              <div class="filter-dropdown-footer">
                <span class="filter-dropdown-footer-text">{{ t('quickFilters.autoDetected') }} ✨</span>
              </div>
            </div>
          </Transition>
        </Teleport>
      </div>

      <!-- Smart Filters Dropdown (Framework-based) -->
      <div v-if="smartFilters.length > 0" class="filter-dropdown" ref="smartDropdownRef">
        <button @click="toggleDropdown('smart')"
          :class="['filter-dropdown-trigger filter-dropdown-trigger-smart', { 'filter-dropdown-trigger-active': hasActiveSmartFilters }]">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
          <span>{{ t('quickFilters.smart') }}</span>
          <span class="filter-dropdown-badge filter-dropdown-badge-smart">{{ smartFilters.length }}</span>
          <svg class="w-3 h-3 transition-transform" :class="{ 'rotate-180': openDropdown === 'smart' }" fill="none"
            stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <Teleport to="body">
          <Transition name="dropdown">
            <div v-if="openDropdown === 'smart'" class="filter-dropdown-menu filter-dropdown-menu-smart"
              :style="dropdownStyle">
              <div class="filter-dropdown-header filter-dropdown-header-smart">
                <span>{{ t('quickFilters.smartFilters') }}</span>
                <button v-if="activeSmartFilters.length > 0" @click="clearSmartFilters" class="filter-clear-btn">
                  {{ t('quickFilters.clearAll') }}
                </button>
              </div>
              <div class="filter-dropdown-items">
                <button v-for="filter in smartFilters" :key="filter.id" @click="toggleFilter(filter, $event)"
                  :class="['filter-dropdown-item', { 'filter-dropdown-item-active': isFilterActive(filter.id) }]">
                  <div class="filter-item-left">
                    <span class="filter-item-emoji">{{ filter.icon }}</span>
                    <div class="filter-item-info">
                      <span class="filter-item-label">{{ filter.label }}</span>
                      <span class="filter-item-framework">{{ filter.framework }}</span>
                    </div>
                  </div>
                  <div class="filter-item-right">
                    <span class="filter-item-count">{{ getFilterCount(filter) }}</span>
                    <div class="filter-item-bar">
                      <div class="filter-item-bar-fill filter-item-bar-smart"
                        :style="{ width: getFilterPercentage(filter) + '%' }"></div>
                    </div>
                  </div>
                </button>
              </div>
              <div class="filter-dropdown-footer filter-dropdown-footer-smart">
                <span class="filter-dropdown-footer-text">{{ t('quickFilters.basedOnFramework') }} 🧠</span>
              </div>
            </div>
          </Transition>
        </Teleport>
      </div>
    </div>

    <!-- Active & Excluded Filters Chips -->
    <div v-if="allActiveFilters.length > 0 || allExcludedFilters.length > 0" class="active-filters">
      <TransitionGroup name="chip">
        <!-- Active filters -->
        <button v-for="filter in allActiveFilters" :key="filter.id" @click="removeFilter(filter)"
          :class="['active-filter-chip', `active-filter-chip-${filter.category || 'lang'}`]">
          <span v-if="filter.icon" class="active-filter-icon">{{ filter.icon }}</span>
          <span>{{ filter.shortLabel || filter.label }}</span>
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        <!-- Excluded filters (with strikethrough style) -->
        <button v-for="filter in allExcludedFilters" :key="'ex-' + filter.id" @click="removeFilter(filter)"
          class="active-filter-chip active-filter-chip-excluded">
          <span class="excluded-icon">⊘</span>
          <span class="excluded-label">{{ filter.shortLabel || filter.label }}</span>
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </TransitionGroup>
    </div>

    <div class="flex-1"></div>

    <!-- Clear All & Settings -->
    <div class="filter-actions">
      <button v-if="allActiveFilters.length > 0" @click="clearAllFilters" class="filter-clear-all-btn"
        :title="t('quickFilters.clearAll')">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
        <span>{{ t('quickFilters.clear') }}</span>
      </button>

      <button @click="showSettingsModal = true" class="filter-settings-btn" :title="t('quickFilters.settings')">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
        </svg>
      </button>
    </div>
  </div>

  <!-- Settings Modal -->
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="showSettingsModal" class="modal-overlay" @click.self="showSettingsModal = false">
        <div class="modal-content filter-settings-modal">
          <div class="modal-header">
            <h3>{{ t('quickFilters.settingsTitle') }}</h3>
            <button @click="showSettingsModal = false" class="icon-btn">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="modal-body">
            <div class="settings-section">
              <h4>{{ t('quickFilters.customFilters') }}</h4>
              <div v-for="filter in editableFilters" :key="filter.id" class="settings-filter-card">
                <div class="settings-filter-header">
                  <label class="settings-toggle">
                    <input type="checkbox" v-model="filter.enabled" />
                    <span class="settings-toggle-track"></span>
                  </label>
                  <span class="settings-filter-name">{{ filter.label }}</span>
                  <span class="settings-filter-count">{{ getFilterCount(filter) }}</span>
                </div>
                <div class="settings-filter-inputs">
                  <input type="text" :value="filter.extensions?.join(', ')"
                    @change="updateFilterExtensions(filter, ($event.target as HTMLInputElement).value)"
                    class="input input-sm" :placeholder="t('quickFilters.extensionsPlaceholder')" />
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="resetFilters" class="btn btn-secondary btn-sm">{{ t('quickFilters.reset') }}</button>
            <button @click="showSettingsModal = false" class="btn btn-primary btn-sm">{{ t('quickFilters.done')
            }}</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>


<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useProjectStore } from '@/stores/project.store'
import { useSettingsStore, type QuickFilterConfig } from '@/stores/settings.store'
import type { ProjectStructure } from '@/types/dto'
import { computed, h, onMounted, onUnmounted, ref, watch } from 'vue'
import { GetProjectStructure } from '../../../../wailsjs/go/main/App'
import { useFileStore } from '../model/file.store'

const { t } = useI18n()
const fileStore = useFileStore()
const settingsStore = useSettingsStore()
const projectStore = useProjectStore()

const openDropdown = ref<'types' | 'langs' | 'smart' | null>(null)
const dropdownStyle = ref<Record<string, string>>({})
const activeFilters = ref<Set<string>>(new Set())
const excludedFilters = ref<Set<string>>(new Set()) // For Shift+Click inversion
const showSettingsModal = ref(false)
const projectStructure = ref<ProjectStructure | null>(null)
const typesDropdownRef = ref<HTMLElement | null>(null)
const langsDropdownRef = ref<HTMLElement | null>(null)
const smartDropdownRef = ref<HTMLElement | null>(null)

const languageConfig: Record<string, { icon: string }> = {
  'Go': { icon: '🐹' }, 'TypeScript': { icon: '📘' }, 'JavaScript': { icon: '📒' },
  'Vue': { icon: '💚' }, 'Python': { icon: '🐍' }, 'Java': { icon: '☕' },
  'Kotlin': { icon: '🟣' }, 'Rust': { icon: '🦀' }, 'C#': { icon: '🟦' },
  'C++': { icon: '⚡' }, 'C': { icon: '🔧' }, 'Ruby': { icon: '💎' },
  'PHP': { icon: '🐘' }, 'Swift': { icon: '🍎' }, 'Dart': { icon: '🎯' },
}

const languageExtensions: Record<string, string[]> = {
  'Go': ['.go'], 'TypeScript': ['.ts', '.tsx'], 'JavaScript': ['.js', '.jsx'],
  'Vue': ['.vue'], 'Python': ['.py'], 'Java': ['.java'],
  'Kotlin': ['.kt', '.kts'], 'Rust': ['.rs'], 'C#': ['.cs'],
  'C++': ['.cpp', '.cc', '.cxx', '.hpp', '.h'], 'C': ['.c', '.h'],
  'Ruby': ['.rb'], 'PHP': ['.php'], 'Swift': ['.swift'], 'Dart': ['.dart'],
}

interface TypeFilter extends QuickFilterConfig {
  category: 'code' | 'test' | 'config' | 'docs' | 'styles'
  shortLabel: string
  icon?: string
}

interface LanguageFilter {
  id: string; label: string; language: string; icon: string
  extensions: string[]; fileCount: number; percentage: number
  primary: boolean; category: 'lang'; shortLabel: string
}

interface SmartFilter {
  id: string; label: string; shortLabel: string; icon: string
  extensions: string[]; patterns: string[]; framework: string
  category: 'smart'
}

// Framework-specific smart filters configuration
const frameworkFilters: Record<string, SmartFilter[]> = {
  'Vue.js': [
    { id: 'vue-components', label: 'Компоненты', shortLabel: 'Компоненты', icon: '🧩', extensions: ['.vue'], patterns: ['**/components/**'], framework: 'Vue.js', category: 'smart' },
    { id: 'vue-composables', label: 'Composables', shortLabel: 'Composables', icon: '🪝', extensions: ['.ts'], patterns: ['**/composables/**', '**/use*.ts'], framework: 'Vue.js', category: 'smart' },
    { id: 'vue-stores', label: 'Stores', shortLabel: 'Stores', icon: '🗄️', extensions: ['.ts'], patterns: ['**/stores/**', '**/*.store.ts'], framework: 'Vue.js', category: 'smart' },
    { id: 'vue-views', label: 'Views/Pages', shortLabel: 'Views', icon: '📄', extensions: ['.vue'], patterns: ['**/views/**', '**/pages/**'], framework: 'Vue.js', category: 'smart' },
  ],
  'React': [
    { id: 'react-components', label: 'Компоненты', shortLabel: 'Компоненты', icon: '🧩', extensions: ['.tsx', '.jsx'], patterns: ['**/components/**'], framework: 'React', category: 'smart' },
    { id: 'react-hooks', label: 'Хуки', shortLabel: 'Хуки', icon: '🪝', extensions: ['.ts', '.tsx'], patterns: ['**/hooks/**', '**/use*.ts', '**/use*.tsx'], framework: 'React', category: 'smart' },
    { id: 'react-pages', label: 'Pages', shortLabel: 'Pages', icon: '📄', extensions: ['.tsx', '.jsx'], patterns: ['**/pages/**', '**/app/**'], framework: 'React', category: 'smart' },
  ],
  'Next.js': [
    { id: 'next-pages', label: 'Pages/Routes', shortLabel: 'Pages', icon: '📄', extensions: ['.tsx', '.jsx'], patterns: ['**/app/**', '**/pages/**'], framework: 'Next.js', category: 'smart' },
    { id: 'next-components', label: 'Компоненты', shortLabel: 'Компоненты', icon: '🧩', extensions: ['.tsx', '.jsx'], patterns: ['**/components/**'], framework: 'Next.js', category: 'smart' },
    { id: 'next-api', label: 'API Routes', shortLabel: 'API', icon: '🔌', extensions: ['.ts', '.tsx'], patterns: ['**/api/**', '**/route.ts'], framework: 'Next.js', category: 'smart' },
  ],
  'Angular': [
    { id: 'angular-components', label: 'Компоненты', shortLabel: 'Компоненты', icon: '🧩', extensions: ['.ts'], patterns: ['**/*.component.ts'], framework: 'Angular', category: 'smart' },
    { id: 'angular-services', label: 'Сервисы', shortLabel: 'Сервисы', icon: '⚙️', extensions: ['.ts'], patterns: ['**/*.service.ts'], framework: 'Angular', category: 'smart' },
    { id: 'angular-modules', label: 'Модули', shortLabel: 'Модули', icon: '📦', extensions: ['.ts'], patterns: ['**/*.module.ts'], framework: 'Angular', category: 'smart' },
  ],
  'Gin': [
    { id: 'go-handlers', label: 'Handlers', shortLabel: 'Handlers', icon: '🎯', extensions: ['.go'], patterns: ['**/handlers/**', '**/*_handler.go'], framework: 'Gin', category: 'smart' },
    { id: 'go-services', label: 'Services', shortLabel: 'Services', icon: '⚙️', extensions: ['.go'], patterns: ['**/services/**', '**/*_service.go', '**/application/**'], framework: 'Gin', category: 'smart' },
    { id: 'go-domain', label: 'Domain', shortLabel: 'Domain', icon: '🏛️', extensions: ['.go'], patterns: ['**/domain/**', '**/entities/**', '**/models/**'], framework: 'Gin', category: 'smart' },
    { id: 'go-infra', label: 'Infrastructure', shortLabel: 'Infra', icon: '🔧', extensions: ['.go'], patterns: ['**/infrastructure/**', '**/repository/**', '**/adapters/**'], framework: 'Gin', category: 'smart' },
  ],
  'Echo': [
    { id: 'go-handlers', label: 'Handlers', shortLabel: 'Handlers', icon: '🎯', extensions: ['.go'], patterns: ['**/handlers/**', '**/*_handler.go'], framework: 'Echo', category: 'smart' },
    { id: 'go-services', label: 'Services', shortLabel: 'Services', icon: '⚙️', extensions: ['.go'], patterns: ['**/services/**', '**/*_service.go'], framework: 'Echo', category: 'smart' },
  ],
  'Wails': [
    { id: 'wails-backend', label: 'Backend (Go)', shortLabel: 'Backend', icon: '🐹', extensions: ['.go'], patterns: ['**/backend/**', '**/*.go'], framework: 'Wails', category: 'smart' },
    { id: 'wails-frontend', label: 'Frontend', shortLabel: 'Frontend', icon: '🎨', extensions: ['.vue', '.tsx', '.ts'], patterns: ['**/frontend/**'], framework: 'Wails', category: 'smart' },
  ],
  'Django': [
    { id: 'django-views', label: 'Views', shortLabel: 'Views', icon: '👁️', extensions: ['.py'], patterns: ['**/views.py', '**/views/**'], framework: 'Django', category: 'smart' },
    { id: 'django-models', label: 'Models', shortLabel: 'Models', icon: '🗃️', extensions: ['.py'], patterns: ['**/models.py', '**/models/**'], framework: 'Django', category: 'smart' },
    { id: 'django-urls', label: 'URLs', shortLabel: 'URLs', icon: '🔗', extensions: ['.py'], patterns: ['**/urls.py'], framework: 'Django', category: 'smart' },
  ],
  'FastAPI': [
    { id: 'fastapi-routes', label: 'Routes', shortLabel: 'Routes', icon: '🔌', extensions: ['.py'], patterns: ['**/routes/**', '**/routers/**', '**/api/**'], framework: 'FastAPI', category: 'smart' },
    { id: 'fastapi-models', label: 'Models', shortLabel: 'Models', icon: '🗃️', extensions: ['.py'], patterns: ['**/models/**', '**/schemas/**'], framework: 'FastAPI', category: 'smart' },
  ],
  'Spring Boot': [
    { id: 'spring-controllers', label: 'Controllers', shortLabel: 'Controllers', icon: '🎯', extensions: ['.java', '.kt'], patterns: ['**/*Controller.java', '**/*Controller.kt'], framework: 'Spring', category: 'smart' },
    { id: 'spring-services', label: 'Services', shortLabel: 'Services', icon: '⚙️', extensions: ['.java', '.kt'], patterns: ['**/*Service.java', '**/*Service.kt'], framework: 'Spring', category: 'smart' },
    { id: 'spring-repos', label: 'Repositories', shortLabel: 'Repos', icon: '🗄️', extensions: ['.java', '.kt'], patterns: ['**/*Repository.java', '**/*Repository.kt'], framework: 'Spring', category: 'smart' },
  ],
  'Flutter': [
    { id: 'flutter-screens', label: 'Screens', shortLabel: 'Screens', icon: '📱', extensions: ['.dart'], patterns: ['**/screens/**', '**/pages/**'], framework: 'Flutter', category: 'smart' },
    { id: 'flutter-widgets', label: 'Widgets', shortLabel: 'Widgets', icon: '🧩', extensions: ['.dart'], patterns: ['**/widgets/**', '**/components/**'], framework: 'Flutter', category: 'smart' },
    { id: 'flutter-bloc', label: 'BLoC/State', shortLabel: 'State', icon: '🔄', extensions: ['.dart'], patterns: ['**/bloc/**', '**/cubit/**', '**/providers/**'], framework: 'Flutter', category: 'smart' },
  ],
}

const typeFilters = computed<TypeFilter[]>(() => {
  const categoryMap: Record<string, TypeFilter['category']> = {
    'source': 'code', 'tests': 'test', 'config': 'config', 'docs': 'docs', 'styles': 'styles'
  }
  const labelMap: Record<string, string> = {
    'source': 'Код', 'tests': 'Тесты', 'config': 'Конфиг', 'docs': 'Доки', 'styles': 'Стили'
  }
  return settingsStore.settings.fileExplorer.quickFilters.map(f => ({
    ...f, category: categoryMap[f.id] || 'code', shortLabel: labelMap[f.id] || f.label
  }))
})

const languageFilters = computed<LanguageFilter[]>(() => {
  if (!projectStructure.value?.languages?.length) return []
  return projectStructure.value.languages
    .filter(lang => lang.percentage > 3 || lang.fileCount > 2)
    .map(lang => ({
      id: `lang-${lang.name.toLowerCase().replace(/[^a-z0-9]/g, '')}`,
      label: lang.name, language: lang.name,
      icon: languageConfig[lang.name]?.icon || '📄',
      extensions: languageExtensions[lang.name] || [],
      fileCount: lang.fileCount, percentage: lang.percentage,
      primary: lang.primary, category: 'lang' as const, shortLabel: lang.name
    }))
    .filter(f => f.extensions.length > 0)
})

// Smart filters based on detected frameworks
const smartFilters = computed<SmartFilter[]>(() => {
  if (!projectStructure.value?.frameworks?.length) return []
  const filters: SmartFilter[] = []
  for (const fw of projectStructure.value.frameworks) {
    const fwFilters = frameworkFilters[fw.name]
    if (fwFilters) filters.push(...fwFilters)
  }
  return filters
})

const editableFilters = computed(() => settingsStore.settings.fileExplorer.quickFilters)
const activeTypeFilters = computed(() => typeFilters.value.filter(f => activeFilters.value.has(f.id)))
const activeLanguageFilters = computed(() => languageFilters.value.filter(f => activeFilters.value.has(f.id)))
const activeSmartFilters = computed(() => smartFilters.value.filter(f => activeFilters.value.has(f.id)))
const allActiveFilters = computed(() => [...activeTypeFilters.value, ...activeLanguageFilters.value, ...activeSmartFilters.value])
// Excluded filters
const excludedTypeFilters = computed(() => typeFilters.value.filter(f => excludedFilters.value.has(f.id)))
const excludedLanguageFilters = computed(() => languageFilters.value.filter(f => excludedFilters.value.has(f.id)))
const excludedSmartFilters = computed(() => smartFilters.value.filter(f => excludedFilters.value.has(f.id)))
const allExcludedFilters = computed(() => [...excludedTypeFilters.value, ...excludedLanguageFilters.value, ...excludedSmartFilters.value])
const hasActiveTypeFilters = computed(() => activeTypeFilters.value.length > 0 || excludedTypeFilters.value.length > 0)
const hasActiveLanguageFilters = computed(() => activeLanguageFilters.value.length > 0 || excludedLanguageFilters.value.length > 0)
const hasActiveSmartFilters = computed(() => activeSmartFilters.value.length > 0 || excludedSmartFilters.value.length > 0)

interface FileNode { name: string; isDir: boolean; children?: FileNode[] }

const totalFiles = computed(() => {
  let count = 0
  const countFiles = (nodes: FileNode[]) => {
    nodes.forEach(node => { if (!node.isDir) count++; if (node.children) countFiles(node.children) })
  }
  countFiles(fileStore.nodes as FileNode[])
  return count
})

function toggleDropdown(type: 'types' | 'langs' | 'smart') {
  if (openDropdown.value === type) { openDropdown.value = null; return }
  const refMap = { types: typesDropdownRef.value, langs: langsDropdownRef.value, smart: smartDropdownRef.value }
  const ref = refMap[type]
  if (ref) {
    const rect = ref.getBoundingClientRect()
    dropdownStyle.value = { position: 'fixed', top: `${rect.bottom + 4}px`, left: `${rect.left}px`, zIndex: '1100' }
  }
  openDropdown.value = type
}

function closeDropdowns(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.filter-dropdown') && !target.closest('.filter-dropdown-menu')) openDropdown.value = null
}

function toggleFilter(filter: TypeFilter | LanguageFilter | SmartFilter, event: MouseEvent) {
  const isMulti = event.ctrlKey || event.metaKey
  const isExclude = event.shiftKey

  if (isExclude) {
    // Shift+Click: exclude this filter (invert)
    if (excludedFilters.value.has(filter.id)) {
      excludedFilters.value.delete(filter.id)
    } else {
      activeFilters.value.delete(filter.id) // Remove from active if was there
      excludedFilters.value.add(filter.id)
    }
  } else if (isMulti) {
    // Ctrl+Click: multi-select
    excludedFilters.value.delete(filter.id) // Remove from excluded if was there
    if (activeFilters.value.has(filter.id)) {
      activeFilters.value.delete(filter.id)
    } else {
      activeFilters.value.add(filter.id)
    }
  } else {
    // Normal click: single select
    excludedFilters.value.clear()
    if (activeFilters.value.has(filter.id) && activeFilters.value.size === 1) {
      activeFilters.value.clear()
    } else {
      activeFilters.value.clear()
      activeFilters.value.add(filter.id)
    }
  }
  applyFilters()
  if (!isMulti && !isExclude) openDropdown.value = null
}

function removeFilter(filter: TypeFilter | LanguageFilter | SmartFilter) {
  activeFilters.value.delete(filter.id)
  excludedFilters.value.delete(filter.id)
  applyFilters()
}
function isFilterActive(id: string) { return activeFilters.value.has(id) }
function isFilterExcluded(id: string) { return excludedFilters.value.has(id) }
function clearTypeFilters() { typeFilters.value.forEach(f => { activeFilters.value.delete(f.id); excludedFilters.value.delete(f.id) }); applyFilters() }
function clearLanguageFilters() { languageFilters.value.forEach(f => { activeFilters.value.delete(f.id); excludedFilters.value.delete(f.id) }); applyFilters() }
function clearSmartFilters() { smartFilters.value.forEach(f => { activeFilters.value.delete(f.id); excludedFilters.value.delete(f.id) }); applyFilters() }
function clearAllFilters() { activeFilters.value.clear(); excludedFilters.value.clear(); applyFilters() }

function applyFilters() {
  const includeExts: string[] = []
  const excludeExts: string[] = []

  // Collect included extensions
  activeTypeFilters.value.forEach(f => f.extensions && includeExts.push(...f.extensions))
  activeLanguageFilters.value.forEach(f => f.extensions && includeExts.push(...f.extensions))
  activeSmartFilters.value.forEach(f => f.extensions && includeExts.push(...f.extensions))

  // Collect excluded extensions
  excludedTypeFilters.value.forEach(f => f.extensions && excludeExts.push(...f.extensions))
  excludedLanguageFilters.value.forEach(f => f.extensions && excludeExts.push(...f.extensions))
  excludedSmartFilters.value.forEach(f => f.extensions && excludeExts.push(...f.extensions))

  fileStore.setFilterExtensions([...new Set(includeExts)], [...new Set(excludeExts)])
}

function getFilterCount(filter: QuickFilterConfig | LanguageFilter | SmartFilter): number {
  let count = 0
  const countFiles = (nodes: FileNode[]) => {
    nodes.forEach(node => {
      if (!node.isDir && filter.extensions?.some(e => node.name.endsWith(e))) count++
      if (node.children) countFiles(node.children)
    })
  }
  countFiles(fileStore.nodes as FileNode[])
  return count
}

function getFilterPercentage(filter: QuickFilterConfig | LanguageFilter | SmartFilter) {
  return totalFiles.value === 0 ? 0 : Math.min(100, (getFilterCount(filter) / totalFiles.value) * 100)
}

function getFilterIcon(category: string) {
  const paths: Record<string, string> = {
    code: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4',
    test: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    config: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z',
    docs: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
    styles: 'M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01'
  }
  return () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
    h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: paths[category] || paths.code })
  ])
}

function updateFilterExtensions(filter: QuickFilterConfig, value: string) {
  filter.extensions = value.split(',').map(s => s.trim()).filter(s => s)
}

function resetFilters() { settingsStore.resetToDefaults() }

async function loadProjectStructure() {
  if (!projectStore.currentPath) return
  try { projectStructure.value = await GetProjectStructure(projectStore.currentPath) as unknown as ProjectStructure }
  catch (e) { console.warn('Failed to load project structure:', e) }
}

// Persistence - save/load filters per project
const STORAGE_KEY = 'quick-filters-state'

function getProjectKey(): string {
  return projectStore.currentPath?.replace(/[\\/:]/g, '_') || 'default'
}

function saveFiltersState() {
  try {
    const allStates = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}')
    allStates[getProjectKey()] = Array.from(activeFilters.value)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(allStates))
  } catch (e) { console.warn('Failed to save filters state:', e) }
}

function loadFiltersState() {
  try {
    const allStates = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}')
    const saved = allStates[getProjectKey()]
    if (saved && Array.isArray(saved)) {
      activeFilters.value = new Set(saved)
      applyFilters()
    }
  } catch (e) { console.warn('Failed to load filters state:', e) }
}

// Watch for filter changes and save
watch(activeFilters, () => { saveFiltersState() }, { deep: true })

onMounted(() => {
  document.addEventListener('click', closeDropdowns)
  if (projectStore.currentPath) {
    loadProjectStructure()
    loadFiltersState()
  }
})

onUnmounted(() => { document.removeEventListener('click', closeDropdowns) })

watch(() => projectStore.currentPath, (p) => {
  if (p) {
    loadProjectStructure()
    activeFilters.value.clear()
    // Load saved filters for this project after structure is loaded
    setTimeout(() => loadFiltersState(), 100)
  }
})
</script>


<style scoped>
/* Base styles moved to design-tokens.css: .filters-bar, .filter-groups, .active-filters, .filter-actions */

.filter-dropdown {
  @apply relative;
}

.filter-dropdown-trigger {
  @apply flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-full;
  background: var(--bg-2);
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  transition: all 150ms ease-out;
}

.filter-dropdown-trigger:hover {
  background: var(--bg-3);
  border-color: var(--border-strong);
  color: var(--text-primary);
}

.filter-dropdown-trigger-active {
  background: var(--accent-indigo-bg);
  border-color: var(--accent-indigo-border);
  color: white;
}

.filter-dropdown-badge {
  @apply px-1.5 py-0.5 text-[10px] font-bold rounded-full min-w-[18px] text-center;
  background: var(--accent-indigo);
  color: white;
}

.filter-dropdown-menu {
  @apply rounded-xl overflow-hidden;
  background: var(--bg-1);
  border: 1px solid var(--border-default);
  box-shadow: var(--shadow-xl);
  min-width: 260px;
  max-width: 320px;
}

.filter-dropdown-header {
  @apply flex items-center justify-between px-3 py-2 text-xs font-semibold;
  background: var(--bg-2);
  border-bottom: 1px solid var(--border-default);
  color: var(--text-primary);
}

.filter-clear-btn {
  @apply text-[10px] px-2 py-0.5 rounded;
  color: var(--accent-indigo);
  transition: all 150ms ease-out;
}

.filter-clear-btn:hover {
  background: var(--accent-indigo-bg);
}

.filter-dropdown-items {
  @apply py-1 max-h-[280px] overflow-y-auto;
}

.filter-dropdown-item {
  @apply w-full flex items-center justify-between px-3 py-2 text-sm;
  color: var(--text-secondary);
  transition: all 150ms ease-out;
}

.filter-dropdown-item:hover {
  background: var(--bg-2);
  color: var(--text-primary);
}

.filter-dropdown-item-active {
  background: var(--accent-indigo-bg);
  color: white;
}

.filter-dropdown-item-active:hover {
  background: rgba(99, 102, 241, 0.35);
}

.filter-dropdown-item-excluded {
  background: var(--color-danger-soft);
  color: var(--color-danger);
}

.filter-dropdown-item-excluded:hover {
  background: rgba(248, 113, 113, 0.25);
}

.filter-item-label-excluded {
  text-decoration: line-through;
  opacity: 0.7;
}

.filter-item-left {
  @apply flex items-center gap-2;
}

.filter-item-icon {
  @apply w-6 h-6 rounded-md flex items-center justify-center;
}

.filter-item-icon-code {
  background: var(--accent-indigo-bg);
  color: #a5b4fc;
}

.filter-item-icon-test {
  background: var(--accent-emerald-bg);
  color: #6ee7b7;
}

.filter-item-icon-config {
  background: var(--accent-orange-bg);
  color: #fdba74;
}

.filter-item-icon-docs {
  background: var(--accent-purple-bg);
  color: #d8b4fe;
}

.filter-item-icon-styles {
  background: rgba(236, 72, 153, 0.25);
  color: #f9a8d4;
}

.filter-item-emoji {
  @apply text-base;
}

.filter-item-label {
  @apply font-medium;
}

.filter-item-primary {
  @apply text-[10px] text-amber-400;
}

.filter-item-right {
  @apply flex items-center gap-2;
}

.filter-item-count {
  @apply text-xs tabular-nums;
  color: var(--text-muted);
}

.filter-item-bar {
  @apply w-12 h-1.5 rounded-full overflow-hidden;
  background: var(--bg-3);
}

.filter-item-bar-fill {
  @apply h-full rounded-full;
  background: var(--accent-indigo);
  transition: width 300ms ease-out;
}

.filter-item-bar-lang {
  background: linear-gradient(90deg, var(--accent-indigo), var(--accent-purple));
}

.filter-item-bar-smart {
  background: linear-gradient(90deg, var(--accent-emerald), var(--accent-indigo));
}

/* Smart filter specific styles */
.filter-dropdown-trigger-smart {
  background: linear-gradient(135deg, var(--bg-2), rgba(16, 185, 129, 0.1));
}

.filter-dropdown-badge-smart {
  background: var(--accent-emerald);
}

.filter-dropdown-menu-smart {
  border-color: var(--accent-emerald-border);
}

.filter-dropdown-header-smart {
  background: linear-gradient(135deg, var(--bg-2), rgba(16, 185, 129, 0.1));
}

.filter-dropdown-footer-smart {
  background: linear-gradient(135deg, var(--bg-2), rgba(16, 185, 129, 0.1));
}

.filter-item-info {
  @apply flex flex-col;
}

.filter-item-framework {
  @apply text-[10px];
  color: var(--text-muted);
}

.active-filter-chip-smart {
  background: linear-gradient(135deg, var(--accent-emerald-bg), var(--accent-indigo-bg));
  border: 1px solid var(--accent-emerald-border);
  color: white;
}

.filter-dropdown-hint {
  @apply px-3 py-2 text-[10px] text-center;
  background: var(--bg-2);
  border-top: 1px solid var(--border-default);
  color: var(--text-muted);
}

.filter-dropdown-hint kbd {
  @apply px-1.5 py-0.5 rounded text-[9px] font-mono;
  background: var(--bg-3);
  border: 1px solid var(--border-default);
}

.filter-dropdown-footer {
  @apply px-3 py-2 text-center;
  background: var(--bg-2);
  border-top: 1px solid var(--border-default);
}

.filter-dropdown-footer-text {
  @apply text-[10px];
  color: var(--accent-emerald);
}

.active-filter-chip {
  @apply flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-full;
  transition: all 150ms ease-out;
}

.active-filter-chip:hover {
  transform: scale(1.05);
}

.active-filter-chip:hover svg {
  opacity: 1;
}

.active-filter-chip svg {
  opacity: 0.6;
  transition: opacity 150ms ease-out;
}

.active-filter-chip-code {
  background: var(--accent-indigo-bg);
  border: 1px solid var(--accent-indigo-border);
  color: white;
}

.active-filter-chip-test {
  background: var(--accent-emerald-bg);
  border: 1px solid var(--accent-emerald-border);
  color: white;
}

.active-filter-chip-config {
  background: var(--accent-orange-bg);
  border: 1px solid var(--accent-orange-border);
  color: white;
}

.active-filter-chip-docs {
  background: var(--accent-purple-bg);
  border: 1px solid var(--accent-purple-border);
  color: white;
}

.active-filter-chip-styles {
  background: rgba(236, 72, 153, 0.25);
  border: 1px solid rgba(236, 72, 153, 0.5);
  color: white;
}

.active-filter-chip-lang {
  background: linear-gradient(135deg, var(--accent-indigo-bg), var(--accent-purple-bg));
  border: 1px solid var(--accent-purple-border);
  color: white;
}

.active-filter-chip-excluded {
  background: var(--color-danger-soft);
  border: 1px solid var(--color-danger-border);
  color: var(--color-danger);
}

.excluded-icon {
  @apply text-xs;
}

.excluded-label {
  text-decoration: line-through;
  opacity: 0.8;
}

.active-filter-icon {
  @apply text-sm leading-none;
}

.filter-clear-all-btn {
  @apply flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-lg;
  color: var(--color-danger);
  background: var(--color-danger-soft);
  border: 1px solid var(--color-danger-border);
  transition: all 150ms ease-out;
}

.filter-clear-all-btn:hover {
  background: rgba(248, 113, 113, 0.25);
}

.filter-settings-btn {
  @apply p-1.5 rounded-lg;
  color: var(--text-muted);
  transition: all 150ms ease-out;
}

.filter-settings-btn:hover {
  color: var(--text-primary);
  background: var(--bg-2);
}

.filter-settings-modal {
  @apply max-w-md;
}

.modal-header {
  @apply flex items-center justify-between px-4 py-3;
  border-bottom: 1px solid var(--border-default);
}

.modal-header h3 {
  @apply text-lg font-semibold;
  color: var(--text-primary);
}

.modal-body {
  @apply p-4 max-h-[60vh] overflow-y-auto;
}

.settings-section h4 {
  @apply text-xs font-semibold mb-3;
  color: var(--text-muted);
}

.settings-filter-card {
  @apply p-3 rounded-xl mb-2;
  background: var(--bg-2);
  border: 1px solid var(--border-default);
}

.settings-filter-header {
  @apply flex items-center gap-3 mb-2;
}

.settings-toggle {
  @apply relative inline-flex cursor-pointer;
}

.settings-toggle input {
  @apply sr-only;
}

.settings-toggle-track {
  @apply w-8 h-4 rounded-full;
  background: var(--bg-3);
  transition: all 200ms ease-out;
}

.settings-toggle input:checked+.settings-toggle-track {
  background: var(--accent-indigo);
}

.settings-toggle-track::after {
  content: '';
  @apply absolute left-0.5 top-0.5 w-3 h-3 rounded-full bg-white;
  transition: transform 200ms ease-out;
}

.settings-toggle input:checked+.settings-toggle-track::after {
  transform: translateX(16px);
}

.settings-filter-name {
  @apply flex-1 text-sm font-medium;
  color: var(--text-primary);
}

.settings-filter-count {
  @apply text-xs;
  color: var(--text-muted);
}

.settings-filter-inputs {
  @apply mt-2;
}

.modal-footer {
  @apply flex justify-end gap-2 px-4 py-3;
  border-top: 1px solid var(--border-default);
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 150ms ease-out;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.chip-enter-active,
.chip-leave-active {
  transition: all 200ms ease-out;
}

.chip-enter-from {
  opacity: 0;
  transform: scale(0.8);
}

.chip-leave-to {
  opacity: 0;
  transform: scale(0.8) translateX(-10px);
}

.chip-move {
  transition: transform 200ms ease-out;
}

.modal-enter-active,
.modal-leave-active {
  transition: all 200ms ease-out;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-content,
.modal-leave-to .modal-content {
  transform: scale(0.95);
}
</style>
