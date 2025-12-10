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


    </div>

    <!-- Right: Actions -->
    <div class="action-bar-right">
      <!-- Reset Layout -->
      <button @click="handleResetLayout" class="lang-btn" :title="t('workspace.resetLayout')">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
      </button>

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
import { useProjectStore } from '@/stores/project.store'
import { useUIStore } from '@/stores/ui.store'

const contextStore = useContextStore()
const projectStore = useProjectStore()
const uiStore = useUIStore()
const { t, locale, setLocale } = useI18n()

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

const emit = defineEmits<{
  (e: 'open-export'): void
  (e: 'reset-layout'): void
}>()

async function handleResetLayout() {
  // Reset panel sizes
  emit('reset-layout')
  
  // Reset window size via backend
  try {
    // @ts-ignore - Wails bindings
    if (window.go?.main?.App?.ResetWindowState) {
      await window.go.main.App.ResetWindowState()
    }
  } catch (error) {
    console.warn('Failed to reset window state:', error)
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
</style>
