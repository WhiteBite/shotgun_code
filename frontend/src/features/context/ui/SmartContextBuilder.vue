<template>
  <div v-if="shouldShow" class="smart-context-builder">
    <div class="smart-header" @click="isExpanded = !isExpanded">
      <div class="flex items-center gap-2">
        <svg class="w-4 h-4 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
            d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
        </svg>
        <span class="text-sm font-medium text-gray-200">{{ t('context.smartSuggestions') }}</span>
        <span v-if="suggestions.length > 0" class="chip-unified chip-unified-accent text-xs">{{ suggestions.length }}</span>
      </div>
      <svg class="w-4 h-4 text-gray-400 transition-transform" :class="{ 'rotate-180': isExpanded }" 
        fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </div>

    <div v-if="isExpanded" class="smart-content">
      <div v-if="isLoading" class="flex items-center justify-center py-4">
        <svg class="animate-spin h-5 w-5 text-indigo-400" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <span class="ml-2 text-sm text-gray-400">{{ t('context.analyzingFiles') }}</span>
      </div>

      <div v-else class="space-y-1">
        <div v-for="suggestion in suggestions" :key="suggestion.path"
          class="suggestion-item" :class="{ 'suggestion-selected': selectedPaths.has(suggestion.path) }"
          @click="toggleSelection(suggestion.path)">
          <div class="flex items-center gap-2 flex-1 min-w-0">
            <span class="suggestion-icon" :title="getSourceTitle(suggestion.source)">
              {{ getSourceIcon(suggestion.source) }}
            </span>
            <span class="suggestion-path truncate">{{ suggestion.path }}</span>
          </div>
          <span class="suggestion-reason text-xs text-gray-500 truncate">{{ suggestion.reason }}</span>
          <input type="checkbox" :checked="selectedPaths.has(suggestion.path)" 
            class="suggestion-checkbox" @click.stop />
        </div>
      </div>

      <div v-if="suggestions.length > 0" class="smart-actions">
        <button @click="addSelected" class="btn btn-sm btn-secondary" :disabled="selectedPaths.size === 0">
          <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('context.addSelected') }} ({{ selectedPaths.size }})
        </button>
        <button @click="addAll" class="btn btn-sm btn-primary">
          <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('context.addAll') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';
import { apiService, type SmartSuggestion } from '@/services/api.service';
import { useProjectStore } from '@/stores/project.store';
import { computed, ref, watch } from 'vue';

const props = defineProps<{
  selectedFiles: string[]
}>()

const emit = defineEmits<{
  (e: 'add-files', files: string[]): void
}>()

const { t } = useI18n()
const projectStore = useProjectStore()

const suggestions = ref<SmartSuggestion[]>([])
const selectedPaths = ref<Set<string>>(new Set())
const isLoading = ref(false)
const isExpanded = ref(true)

const projectPath = computed(() => projectStore.currentPath || '')

// Show component only when loading or has suggestions
const shouldShow = computed(() => isLoading.value || suggestions.value.length > 0)

// Timeout for loading state
let loadingTimeout: ReturnType<typeof setTimeout> | null = null
const LOADING_TIMEOUT_MS = 30000

// Fetch suggestions when selected files change
watch(() => props.selectedFiles, async (files) => {
  // Clear previous timeout
  if (loadingTimeout) {
    clearTimeout(loadingTimeout)
    loadingTimeout = null
  }

  if (!projectPath.value || files.length === 0) {
    suggestions.value = []
    isLoading.value = false
    return
  }

  isLoading.value = true
  
  // Set timeout to prevent infinite loading
  loadingTimeout = setTimeout(() => {
    if (isLoading.value) {
      isLoading.value = false
    }
  }, LOADING_TIMEOUT_MS)

  try {
    const result = await apiService.getSmartSuggestions(projectPath.value, files)
    suggestions.value = result.suggestions
    // Pre-select all by default
    selectedPaths.value = new Set(result.suggestions.map(s => s.path))
  } catch {
    suggestions.value = []
  } finally {
    isLoading.value = false
    if (loadingTimeout) {
      clearTimeout(loadingTimeout)
      loadingTimeout = null
    }
  }
}, { immediate: true, deep: true })

function getSourceIcon(source: string): string {
  switch (source) {
    case 'git': return '🔀'
    case 'arch': return '🏗️'
    case 'semantic': return '🔍'
    default: return '📄'
  }
}

function getSourceTitle(source: string): string {
  switch (source) {
    case 'git': return t('context.sourceGit')
    case 'arch': return t('context.sourceArch')
    case 'semantic': return t('context.sourceSemantic')
    default: return ''
  }
}

function toggleSelection(path: string) {
  if (selectedPaths.value.has(path)) {
    selectedPaths.value.delete(path)
  } else {
    selectedPaths.value.add(path)
  }
  selectedPaths.value = new Set(selectedPaths.value) // trigger reactivity
}

function addSelected() {
  if (selectedPaths.value.size > 0) {
    emit('add-files', Array.from(selectedPaths.value))
    // Remove added files from suggestions
    suggestions.value = suggestions.value.filter(s => !selectedPaths.value.has(s.path))
    selectedPaths.value.clear()
  }
}

function addAll() {
  const allPaths = suggestions.value.map(s => s.path)
  emit('add-files', allPaths)
  suggestions.value = []
  selectedPaths.value.clear()
}
</script>

<style scoped>
.smart-context-builder {
  @apply rounded-lg overflow-hidden mb-3;
  background: var(--bg-1);
  border: 1px solid var(--border-default);
}

.smart-header {
  @apply flex items-center justify-between px-3 py-2 cursor-pointer;
  background: var(--bg-2);
  transition: background 150ms ease-out;
}

.smart-header:hover {
  background: var(--bg-3);
}

.smart-content {
  @apply p-2;
}

.suggestion-item {
  @apply flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer;
  transition: background 150ms ease-out;
}

.suggestion-item:hover {
  background: var(--bg-2);
}

.suggestion-selected {
  background: rgba(99, 102, 241, 0.1);
  border-left: 2px solid var(--color-primary);
}

.suggestion-icon {
  @apply flex-shrink-0 text-sm;
}

.suggestion-path {
  @apply text-sm;
  color: var(--text-primary);
}

.suggestion-reason {
  @apply flex-shrink-0 max-w-32;
}

.suggestion-checkbox {
  @apply flex-shrink-0 w-4 h-4 rounded;
  accent-color: var(--color-primary);
}

.smart-actions {
  @apply flex items-center justify-end gap-2 pt-2 mt-2;
  border-top: 1px solid var(--border-default);
}
</style>
