<template>
  <div class="h-14 flex items-center justify-between px-4 gap-3">
    <!-- Left: Project & Stats -->
    <div class="flex items-center gap-4 text-sm min-w-0 flex-shrink">
      <!-- Project Name -->
      <button 
        @click="changeProject" 
        class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-gray-800/50 border border-gray-700/50 hover:bg-gray-700/50 hover:border-gray-600/50 transition-all group"
        :title="t('hotkey.changeProject')"
      >
        <svg class="w-4 h-4 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        <span class="text-white text-xs font-medium truncate max-w-[140px]">{{ projectStore.projectName }}</span>
        <svg class="w-3 h-3 text-gray-500 group-hover:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      
      <!-- Stats (compact) with limit indicator and tooltip -->
      <div class="hidden md:flex items-center gap-2 text-xs">
        <div 
          class="token-limit-indicator relative"
          @mouseenter="showTooltip = true"
          @mouseleave="showTooltip = false"
        >
          <div 
            class="flex items-center gap-2 px-2 py-1 rounded border cursor-pointer transition-all"
            :class="[tokenLimitClass, pulseClass]"
          >
            <div class="w-16 h-1.5 bg-gray-700/50 rounded-full overflow-hidden">
              <div 
                class="h-full rounded-full transition-all duration-500 limits-bar-animate"
                :class="tokenBarClass"
                :style="{ width: tokenUsagePercent + '%' }"
              ></div>
            </div>
            <span>~{{ formatNumber(contextStore.tokenCount) }}</span>
            <span v-if="tokenUsagePercent >= 80" class="text-[9px] font-bold" :class="tokenBarClass.replace('bg-', 'text-')">
              {{ tokenUsagePercent }}%
            </span>
          </div>
          
          <!-- Tooltip -->
          <Transition name="tooltip-fade">
            <div v-if="showTooltip" class="token-tooltip">
              <div class="tooltip-header">
                <span class="tooltip-model">{{ aiStore.currentModel }}</span>
                <span class="tooltip-provider">{{ aiStore.currentProvider }}</span>
              </div>
              <div class="tooltip-stats">
                <div class="tooltip-stat">
                  <span class="tooltip-stat-label">{{ t('limits.contextUsed') }}</span>
                  <span class="tooltip-stat-value">{{ formatNumber(contextStore.tokenCount) }}</span>
                </div>
                <div class="tooltip-stat">
                  <span class="tooltip-stat-label">{{ t('limits.available') }}</span>
                  <span class="tooltip-stat-value text-emerald-400">{{ formatNumber(availableTokens) }}</span>
                </div>
                <div class="tooltip-stat">
                  <span class="tooltip-stat-label">{{ t('limits.maxContext') }}</span>
                  <span class="tooltip-stat-value">{{ formatNumber(modelContextLimit) }}</span>
                </div>
              </div>
              <div class="tooltip-bar">
                <div 
                  class="tooltip-bar-fill"
                  :class="tokenBarClass"
                  :style="{ width: tokenUsagePercent + '%' }"
                ></div>
              </div>
              <div class="tooltip-percent">{{ tokenUsagePercent }}% {{ t('limits.used') }}</div>
            </div>
          </Transition>
        </div>
        
        <span class="px-2 py-1 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
          ${{ contextStore.estimatedCost.toFixed(4) }}
        </span>
      </div>
    </div>

    <!-- Right: Action Buttons -->
    <div class="flex items-center gap-2 flex-shrink-0">
      <!-- Language Switcher -->
      <button
        @click="toggleLanguage"
        class="px-2 py-1.5 rounded-lg bg-gray-800/50 border border-gray-700/50 text-gray-400 hover:text-white hover:bg-gray-700/50 transition-colors text-xs font-medium"
        :title="locale === 'ru' ? 'Switch to English' : 'Переключить на русский'"
      >
        {{ locale.toUpperCase() }}
      </button>

      <!-- Build Context -->
      <button
        @click="$emit('build-context')"
        :disabled="fileStore.selectedPaths.size === 0 || contextStore.isBuilding"
        class="action-btn-hero !py-2 !px-4"
        :title="t('hotkey.buildContext')"
      >
        <svg v-if="!contextStore.isBuilding" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <svg v-else class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span class="text-xs font-medium">
          {{ contextStore.isBuilding ? t('context.building') : t('context.build') }}
          <span v-if="fileStore.selectedPaths.size > 0" class="opacity-75">({{ fileStore.selectedPaths.size }})</span>
        </span>
      </button>
      
      <!-- Copy -->
      <button
        @click="handleCopyContext"
        :disabled="!contextStore.hasContext"
        class="action-btn action-btn-success !py-2 !px-3"
        :title="t('hotkey.copy')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v2m-6 12h8a2 2 0 0 0 2-2v-8a2 2 0 0 0-2-2h-8a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2z" />
        </svg>
        <span class="hidden xl:inline text-xs">{{ t('action.copy') }}</span>
      </button>
      
      <!-- Export -->
      <button
        @click="$emit('open-export')"
        :disabled="!contextStore.hasContext"
        class="action-btn action-btn-accent !py-2 !px-3"
        :title="t('hotkey.export')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
        </svg>
        <span class="hidden xl:inline text-xs">{{ t('action.export') }}</span>
      </button>
      
      <!-- Generate Solution -->
      <button
        @click="$emit('generate-solution')"
        :disabled="!contextStore.hasContext"
        class="action-btn-hero !py-2 !px-4"
        :title="t('action.generateSolution')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        <span class="hidden xl:inline text-xs">{{ t('action.generateSolution') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useContextStore } from '@/features/context'
import { useFileStore } from '@/features/files'
import { MODEL_LIMITS, useAIStore } from '@/stores/ai.store'
import { useProjectStore } from '@/stores/project.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, onMounted, ref } from 'vue'

const fileStore = useFileStore()
const contextStore = useContextStore()
const projectStore = useProjectStore()
const aiStore = useAIStore()
const uiStore = useUIStore()
const { t, locale, setLocale } = useI18n()

const showTooltip = ref(false)

defineEmits<{
  (e: 'build-context'): void
  (e: 'open-export'): void
  (e: 'generate-solution'): void
}>()

// Load AI provider info on mount
onMounted(() => {
  aiStore.loadProviderInfo()
})

const modelContextLimit = computed(() => {
  return MODEL_LIMITS[aiStore.currentModel] || 32000
})

const tokenUsagePercent = computed(() => {
  const percent = Math.round((contextStore.tokenCount / modelContextLimit.value) * 100)
  return Math.min(percent, 100)
})

const availableTokens = computed(() => {
  return Math.max(0, modelContextLimit.value - contextStore.tokenCount)
})

const tokenLimitClass = computed(() => {
  if (tokenUsagePercent.value >= 95) return 'bg-red-500/10 text-red-400 border-red-500/20'
  if (tokenUsagePercent.value >= 80) return 'bg-amber-500/10 text-amber-400 border-amber-500/20'
  return 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20'
})

const tokenBarClass = computed(() => {
  if (tokenUsagePercent.value >= 95) return 'bg-red-500'
  if (tokenUsagePercent.value >= 80) return 'bg-amber-500'
  return 'bg-indigo-500'
})

const pulseClass = computed(() => {
  if (tokenUsagePercent.value >= 95) return 'limits-critical-pulse'
  if (tokenUsagePercent.value >= 80) return 'limits-warning-pulse'
  return ''
})

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
  return num.toString()
}

function toggleLanguage() {
  const newLocale = locale.value === 'ru' ? 'en' : 'ru'
  setLocale(newLocale)
  uiStore.addToast(
    newLocale === 'ru' ? 'Язык изменён на русский' : 'Language changed to English',
    'success'
  )
}

function changeProject() {
  projectStore.clearProject()
  uiStore.addToast(t('action.changeProject'), 'info')
}

async function handleCopyContext() {
  if (!contextStore.contextId) {
    return
  }
  
  try {
    const content = await contextStore.getFullContextContent()
    await navigator.clipboard.writeText(content)
    uiStore.addToast(t('toast.contextCopied'), 'success')
  } catch (error) {
    console.error('Failed to copy context:', error)
    uiStore.addToast(t('toast.copyError'), 'error')
  }
}
</script>

<style scoped>
.token-limit-indicator {
  position: relative;
}

.token-tooltip {
  @apply absolute top-full left-1/2 -translate-x-1/2 mt-2 z-50;
  @apply bg-gray-800 border border-gray-700 rounded-xl p-3 shadow-2xl;
  @apply min-w-[200px];
  backdrop-filter: blur(12px);
}

.token-tooltip::before {
  content: '';
  @apply absolute -top-2 left-1/2 -translate-x-1/2;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-bottom: 8px solid rgb(55 65 81);
}

.tooltip-header {
  @apply flex items-center justify-between mb-2 pb-2 border-b border-gray-700;
}

.tooltip-model {
  @apply text-xs font-medium text-white;
}

.tooltip-provider {
  @apply text-[10px] px-1.5 py-0.5 bg-indigo-500/20 text-indigo-300 rounded;
}

.tooltip-stats {
  @apply space-y-1.5 mb-3;
}

.tooltip-stat {
  @apply flex items-center justify-between text-xs;
}

.tooltip-stat-label {
  @apply text-gray-500;
}

.tooltip-stat-value {
  @apply text-white font-medium;
}

.tooltip-bar {
  @apply h-2 bg-gray-700 rounded-full overflow-hidden mb-2;
}

.tooltip-bar-fill {
  @apply h-full rounded-full transition-all duration-500;
}

.tooltip-percent {
  @apply text-center text-[10px] text-gray-400;
}
</style>
