<template>
  <div class="action-bar">
    <!-- Left: Project & Stats -->
    <div class="action-bar-left">
      <!-- Project Name -->
      <button @click="changeProject" class="project-btn" :title="t('hotkey.changeProject')">
        <div class="project-icon">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
        </div>
        <span class="project-name">{{ projectStore.projectName }}</span>
        <svg class="w-3 h-3 text-gray-500 group-hover:text-gray-300" fill="none" stroke="currentColor"
          viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      <!-- Token Stats with Tooltip -->
      <div class="hidden md:flex items-center gap-2">
        <div class="token-indicator" :class="tokenLimitClass" @mouseenter="showTooltip = true"
          @mouseleave="showTooltip = false">
          <div class="token-bar-bg">
            <div class="token-bar-fill" :class="tokenBarClass" :style="{ width: tokenUsagePercent + '%' }"></div>
          </div>
          <span class="token-value">~{{ formatNumber(contextStore.tokenCount) }}</span>
          <span v-if="tokenUsagePercent >= 80" class="token-percent">{{ tokenUsagePercent }}%</span>

          <!-- Tooltip -->
          <Transition name="tooltip-fade">
            <div v-if="showTooltip" class="token-tooltip">
              <div class="tooltip-header">
                <span class="tooltip-model">{{ aiStore.currentModel }}</span>
                <span class="tooltip-provider">{{ aiStore.currentProvider }}</span>
              </div>
              <div class="tooltip-stats">
                <div class="tooltip-stat">
                  <span>{{ t('limits.contextUsed') }}</span>
                  <span class="text-white">{{ formatNumber(contextStore.tokenCount) }}</span>
                </div>
                <div class="tooltip-stat">
                  <span>{{ t('limits.available') }}</span>
                  <span class="text-emerald-400">{{ formatNumber(availableTokens) }}</span>
                </div>
                <div class="tooltip-stat">
                  <span>{{ t('limits.maxContext') }}</span>
                  <span class="text-white">{{ formatNumber(modelContextLimit) }}</span>
                </div>
              </div>
              <div class="tooltip-bar">
                <div class="tooltip-bar-fill" :class="tokenBarClass" :style="{ width: tokenUsagePercent + '%' }"></div>
              </div>
              <div class="tooltip-percent">{{ tokenUsagePercent }}% {{ t('limits.used') }}</div>
            </div>
          </Transition>
        </div>

        <span class="cost-badge">${{ contextStore.estimatedCost.toFixed(4) }}</span>
      </div>
    </div>

    <!-- Right: Actions -->
    <div class="action-bar-right">
      <!-- Language -->
      <button @click="toggleLanguage" class="lang-btn"
        :title="locale === 'ru' ? 'Switch to English' : 'Переключить на русский'">
        {{ locale.toUpperCase() }}
      </button>

      <!-- Copy -->
      <button @click="handleCopyContext" :disabled="!contextStore.hasContext" class="action-btn action-btn-copy"
        :title="t('hotkey.copy')">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M8 16H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v2m-6 12h8a2 2 0 0 0 2-2v-8a2 2 0 0 0-2-2h-8a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2z" />
        </svg>
        <span class="hidden xl:inline">{{ t('action.copy') }}</span>
      </button>

      <!-- Export -->
      <button @click="$emit('open-export')" :disabled="!contextStore.hasContext" class="action-btn action-btn-export"
        :title="t('hotkey.export')">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
        </svg>
        <span class="hidden xl:inline">{{ t('action.export') }}</span>
      </button>
    </div>
  </div>
</template>


<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useContextStore } from '@/features/context'
import { MODEL_LIMITS, useAIStore } from '@/stores/ai.store'
import { useProjectStore } from '@/stores/project.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, onMounted, ref } from 'vue'

const contextStore = useContextStore()
const projectStore = useProjectStore()
const aiStore = useAIStore()
const uiStore = useUIStore()
const { t, locale, setLocale } = useI18n()

const showTooltip = ref(false)

defineEmits<{
  (e: 'open-export'): void
}>()

onMounted(() => {
  aiStore.loadProviderInfo()
})

const modelContextLimit = computed(() => MODEL_LIMITS[aiStore.currentModel] || 32000)
const tokenUsagePercent = computed(() => Math.min(Math.round((contextStore.tokenCount / modelContextLimit.value) * 100), 100))
const availableTokens = computed(() => Math.max(0, modelContextLimit.value - contextStore.tokenCount))

const tokenLimitClass = computed(() => {
  if (tokenUsagePercent.value >= 95) return 'token-critical'
  if (tokenUsagePercent.value >= 80) return 'token-warning'
  return 'token-normal'
})

const tokenBarClass = computed(() => {
  if (tokenUsagePercent.value >= 95) return 'bg-red-500'
  if (tokenUsagePercent.value >= 80) return 'bg-amber-500'
  return 'bg-purple-500'
})

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
  return num.toString()
}

function toggleLanguage() {
  const newLocale = locale.value === 'ru' ? 'en' : 'ru'
  setLocale(newLocale)
  uiStore.addToast(newLocale === 'ru' ? 'Язык изменён на русский' : 'Language changed to English', 'success')
}

function changeProject() {
  projectStore.clearProject()
  uiStore.addToast(t('action.changeProject'), 'info')
}

async function handleCopyContext() {
  if (!contextStore.contextId) return
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
.action-bar {
  @apply h-14 flex items-center justify-between px-4 gap-3;
}

.action-bar-left {
  @apply flex items-center gap-4 min-w-0 flex-shrink;
}

.action-bar-right {
  @apply flex items-center gap-2 flex-shrink-0;
}

/* Project Button */
.project-btn {
  @apply flex items-center gap-2 px-3 py-1.5 rounded-xl;
  background: var(--bg-1);
  border: 1px solid var(--border-default);
  transition: all 200ms ease-out;
}

.project-btn:hover {
  background: var(--bg-2);
  border-color: var(--border-strong);
}

.project-icon {
  @apply w-6 h-6 rounded-lg flex items-center justify-center;
  background: var(--accent-purple-bg);
  color: #d8b4fe;
}

.project-name {
  @apply text-xs font-medium truncate max-w-[140px];
  color: var(--text-primary);
}

/* Token Indicator */
.token-indicator {
  @apply relative flex items-center gap-2 px-2.5 py-1.5 rounded-lg;
  @apply cursor-pointer;
  transition: all 200ms ease-out;
}

.token-normal {
  background: var(--accent-purple-bg);
  color: #d8b4fe;
  border: 1px solid var(--accent-purple-border);
}

.token-warning {
  background: var(--color-warning-soft);
  color: var(--color-warning);
  border: 1px solid var(--color-warning-border);
  animation: pulse-warning 2s ease-in-out infinite;
}

.token-critical {
  background: var(--color-danger-soft);
  color: var(--color-danger);
  border: 1px solid var(--color-danger-border);
  animation: pulse-critical 1s ease-in-out infinite;
}

@keyframes pulse-warning {

  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.4);
  }

  50% {
    box-shadow: 0 0 0 4px rgba(245, 158, 11, 0);
  }
}

@keyframes pulse-critical {

  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.5);
  }

  50% {
    box-shadow: 0 0 0 6px rgba(239, 68, 68, 0);
  }
}

.token-bar-bg {
  @apply w-16 h-1.5 rounded-full overflow-hidden;
  background: var(--bg-3);
}

.token-bar-fill {
  @apply h-full rounded-full transition-all duration-500;
}

.token-value {
  @apply text-xs font-medium;
}

.token-percent {
  @apply text-[9px] font-bold;
}

/* Cost Badge */
.cost-badge {
  @apply px-2.5 py-1.5 rounded-lg text-xs font-medium;
  background: var(--color-success-soft);
  color: var(--color-success);
  border: 1px solid var(--color-success-border);
}

/* Language Button */
.lang-btn {
  @apply px-2.5 py-1.5 rounded-lg text-xs font-medium;
  background: var(--bg-1);
  border: 1px solid var(--border-default);
  color: var(--text-muted);
  transition: all 200ms ease-out;
}

.lang-btn:hover {
  color: var(--text-primary);
  background: var(--bg-2);
}

/* Action Buttons */
.action-btn {
  @apply flex items-center gap-2 px-3 py-2 rounded-xl text-sm font-medium;
  @apply transition-all duration-200;
  @apply disabled:opacity-50 disabled:cursor-not-allowed;
}

.action-btn-copy {
  background: var(--color-success-soft);
  color: var(--color-success);
  border: 1px solid var(--color-success-border);
}

.action-btn-copy:hover:not(:disabled) {
  background: rgba(74, 222, 128, 0.25);
  border-color: rgba(74, 222, 128, 0.55);
  transform: translateY(-1px);
}

.action-btn-export {
  background: var(--color-warning-soft);
  color: var(--color-warning);
  border: 1px solid var(--color-warning-border);
}

.action-btn-export:hover:not(:disabled) {
  background: rgba(251, 191, 36, 0.25);
  border-color: rgba(251, 191, 36, 0.55);
  transform: translateY(-1px);
}

/* Tooltip */
.token-tooltip {
  @apply absolute bottom-full left-0 mb-2 z-50;
  @apply rounded-xl p-3 min-w-[200px];
  background: var(--bg-1);
  border: 1px solid var(--border-default);
  box-shadow: var(--shadow-xl);
  backdrop-filter: blur(12px);
}

.token-tooltip::after {
  content: '';
  @apply absolute -bottom-2 left-4;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-top: 8px solid var(--bg-1);
}

.tooltip-header {
  @apply flex items-center justify-between mb-2 pb-2;
  border-bottom: 1px solid var(--border-default);
}

.tooltip-model {
  @apply text-xs font-medium;
  color: var(--text-primary);
}

.tooltip-provider {
  @apply text-[10px] px-1.5 py-0.5 rounded;
  background: var(--accent-purple-bg);
  color: #d8b4fe;
}

.tooltip-stats {
  @apply space-y-1.5 mb-3;
}

.tooltip-stat {
  @apply flex items-center justify-between text-xs;
  color: var(--text-muted);
}

.tooltip-bar {
  @apply h-2 rounded-full overflow-hidden mb-2;
  background: var(--bg-3);
}

.tooltip-bar-fill {
  @apply h-full rounded-full transition-all duration-500;
}

.tooltip-percent {
  @apply text-center text-[10px];
  color: var(--text-subtle);
}

/* Tooltip Transition */
.tooltip-fade-enter-active {
  transition: all 0.2s ease-out;
}

.tooltip-fade-leave-active {
  transition: all 0.15s ease-in;
}

.tooltip-fade-enter-from,
.tooltip-fade-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
</style>
