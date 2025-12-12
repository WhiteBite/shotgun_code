<template>
  <div v-if="shouldShow" class="impact-preview">
    <div class="impact-header" @click="isExpanded = !isExpanded">
      <div class="flex items-center gap-2">
        <svg class="w-4 h-4 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <span class="text-sm font-medium text-gray-200">{{ t('context.impactAnalysis') }}</span>
      </div>
      <div class="flex items-center gap-2">
        <span v-if="result" class="chip-unified chip-unified-warning text-xs">
          {{ result.totalDependents }} {{ t('context.affected') }}
        </span>
        <svg class="w-4 h-4 text-gray-400 transition-transform" :class="{ 'rotate-180': isExpanded }" 
          fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </div>
    </div>

    <div v-if="isExpanded" class="impact-content">
      <div v-if="isLoading" class="flex items-center justify-center py-4">
        <svg class="animate-spin h-5 w-5 text-indigo-400" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <span class="ml-2 text-sm text-gray-400">{{ t('context.analyzingImpact') }}</span>
      </div>

      <template v-else-if="result">
        <!-- Risk Score -->
        <div class="risk-summary mb-3">
          <div class="flex items-center justify-between mb-1">
            <span class="text-xs text-gray-400">{{ t('context.riskScore') }}</span>
            <span class="text-xs font-medium" :class="getRiskColor(result.riskLevel)">
              {{ t(`files.risk.${result.riskLevel}`) }}
            </span>
          </div>
          <div class="risk-bar-bg">
            <div class="risk-bar" :class="getRiskBarClass(result.riskLevel)" 
              :style="{ width: `${Math.round(result.aggregateRisk * 100)}%` }" />
          </div>
        </div>

        <!-- Affected Files -->
        <div v-if="result.affectedFiles.length > 0" class="affected-section">
          <div class="section-header" @click="showAffected = !showAffected">
            <span class="text-xs text-gray-400">
              {{ t('context.affectedFiles') }} ({{ result.affectedFiles.length }})
            </span>
            <svg class="w-3 h-3 text-gray-500 transition-transform" :class="{ 'rotate-180': showAffected }" 
              fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </div>
          <div v-if="showAffected" class="affected-list">
            <div v-for="file in result.affectedFiles" :key="file.path" class="affected-item">
              <span class="file-icon">📄</span>
              <span class="file-path truncate">{{ file.path }}</span>
              <span class="file-type text-xs" :class="file.type === 'direct' ? 'text-amber-400' : 'text-gray-500'">
                {{ file.type }}
              </span>
            </div>
          </div>
        </div>

        <!-- Related Tests -->
        <div v-if="result.relatedTests.length > 0" class="tests-section mt-2">
          <div class="section-header" @click="showTests = !showTests">
            <span class="text-xs text-gray-400">
              🧪 {{ t('context.relatedTests') }} ({{ result.relatedTests.length }})
            </span>
            <svg class="w-3 h-3 text-gray-500 transition-transform" :class="{ 'rotate-180': showTests }" 
              fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </div>
          <div v-if="showTests" class="affected-list">
            <div v-for="test in result.relatedTests" :key="test" class="affected-item">
              <span class="file-icon">🧪</span>
              <span class="file-path truncate">{{ test }}</span>
            </div>
          </div>
        </div>

        <!-- No Impact -->
        <div v-if="result.totalDependents === 0" class="no-impact">
          <span class="text-xs text-gray-500">{{ t('context.noImpact') }}</span>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { apiService } from '@/services/api.service'
import { useProjectStore } from '@/stores/project.store'
import { computed, ref, watch } from 'vue'

interface AffectedFile {
  path: string
  type: 'direct' | 'transitive'
  dependents: number
}

interface ImpactResult {
  totalDependents: number
  aggregateRisk: number
  riskLevel: 'low' | 'medium' | 'high'
  affectedFiles: AffectedFile[]
  relatedTests: string[]
}

const props = defineProps<{
  selectedFiles: string[]
}>()

const { t } = useI18n()
const projectStore = useProjectStore()

const result = ref<ImpactResult | null>(null)
const isLoading = ref(false)
const isExpanded = ref(false)
const showAffected = ref(true)
const showTests = ref(true)

const projectPath = computed(() => projectStore.currentPath || '')
const hasFiles = computed(() => props.selectedFiles.length > 0)

// Show only when loading or has results with data
const shouldShow = computed(() => {
  if (!hasFiles.value) return false
  if (isLoading.value) return true
  if (result.value && result.value.totalDependents > 0) return true
  return false
})

// Debounced fetch
let debounceTimer: ReturnType<typeof setTimeout> | null = null

async function fetchImpact() {
  if (!projectPath.value || props.selectedFiles.length === 0) {
    result.value = null
    return
  }

  isLoading.value = true
  try {
    // @ts-ignore - method may not exist in wails bindings yet
    result.value = await apiService.getImpactPreview(projectPath.value, props.selectedFiles)
  } catch {
    // Silently fail - impact preview is optional feature
    result.value = null
  } finally {
    isLoading.value = false
  }
}

watch(() => props.selectedFiles, () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(fetchImpact, 300)
}, { immediate: true, deep: true })

function getRiskColor(level: string): string {
  switch (level) {
    case 'low': return 'text-green-400'
    case 'medium': return 'text-amber-400'
    case 'high': return 'text-red-400'
    default: return 'text-gray-400'
  }
}

function getRiskBarClass(level: string): string {
  switch (level) {
    case 'low': return 'bg-green-500'
    case 'medium': return 'bg-amber-500'
    case 'high': return 'bg-red-500'
    default: return 'bg-gray-500'
  }
}
</script>

<style scoped>
.impact-preview {
  @apply rounded-lg overflow-hidden mb-3;
  background: var(--bg-1);
  border: 1px solid var(--border-default);
}

.impact-header {
  @apply flex items-center justify-between px-3 py-2 cursor-pointer;
  background: var(--bg-2);
  transition: background 150ms ease-out;
}

.impact-header:hover {
  background: var(--bg-3);
}

.impact-content {
  @apply p-3;
}

.risk-summary {
  @apply pb-2;
  border-bottom: 1px solid var(--border-default);
}

.risk-bar-bg {
  @apply h-2 rounded-full overflow-hidden;
  background: var(--bg-3);
}

.risk-bar {
  @apply h-full rounded-full transition-all duration-300;
}

.section-header {
  @apply flex items-center justify-between py-1 cursor-pointer;
}

.affected-list {
  @apply space-y-1 mt-1 max-h-32 overflow-y-auto;
}

.affected-item {
  @apply flex items-center gap-2 px-2 py-1 rounded text-sm;
  background: var(--bg-2);
}

.file-icon {
  @apply flex-shrink-0 text-xs;
}

.file-path {
  @apply flex-1 min-w-0;
  color: var(--text-primary);
}

.no-impact {
  @apply text-center py-2;
}
</style>
