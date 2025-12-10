<template>
  <div class="sidebar-container">
    <!-- Tab Switcher with Sliding Indicator -->
    <div class="tabs-container">
      <!-- Sliding Indicator -->
      <div 
        class="tabs-indicator"
        :class="tabIndicatorClass"
        :style="{ transform: `translateX(${tabIndex * 100}%)` }"
      ></div>
      
      <button
        @click="currentTab = 'files'"
        :class="['sidebar-tab', currentTab === 'files' ? 'sidebar-tab-active text-indigo-300' : '']"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        <span>{{ t('tabs.files') }}</span>
        <span v-if="fileStore.selectedPaths.size > 0" class="tab-badge">
          {{ fileStore.selectedPaths.size }}
        </span>
      </button>
      
      <button
        @click="currentTab = 'git'"
        :class="['sidebar-tab', currentTab === 'git' ? 'sidebar-tab-active text-orange-300' : '']"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
        </svg>
        <span>Git</span>
      </button>
      
      <button
        @click="currentTab = 'contexts'"
        :class="['sidebar-tab', currentTab === 'contexts' ? 'sidebar-tab-active text-purple-300' : '']"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <span>{{ t('tabs.contexts') }}</span>
      </button>
    </div>

    <!-- Content -->
    <div class="sidebar-content">
      <FileExplorer v-if="currentTab === 'files'" @preview-file="handlePreviewFile" @build-context="handleBuildContext" />
      <GitSourceSelector v-else-if="currentTab === 'git'" />
      <ContextList v-else />
    </div>
  </div>
</template>

<script setup lang="ts">
import GitSourceSelector from '@/components/GitSourceSelector.vue'
import { useI18n } from '@/composables/useI18n'
import { ContextList } from '@/features/context'
import { FileExplorer, useFileStore } from '@/features/files'
import { ref, watch } from 'vue'

const fileStore = useFileStore()
const { t } = useI18n()
const currentTab = ref<'files' | 'git' | 'contexts'>('files')

// Tab indicator computed
import { computed } from 'vue'

const tabIndex = computed(() => {
  const tabs = ['files', 'git', 'contexts']
  return tabs.indexOf(currentTab.value)
})

const tabIndicatorClass = computed(() => {
  const classes: Record<string, string> = {
    files: 'tabs-indicator-indigo',
    git: 'tabs-indicator-orange',
    contexts: 'tabs-indicator-purple'
  }
  return classes[currentTab.value]
})

const emit = defineEmits<{
  (e: 'preview-file', filePath: string): void
  (e: 'build-context'): void
}>()

function handlePreviewFile(filePath: string) {
  emit('preview-file', filePath)
}

function handleBuildContext() {
  emit('build-context')
}

// Persist tab selection
watch(currentTab, (tab) => {
  try {
    localStorage.setItem('left-sidebar-tab', tab)
  } catch (err) {
    console.warn('Failed to save sidebar tab:', err)
  }
})

// Restore tab selection
try {
  const savedTab = localStorage.getItem('left-sidebar-tab')
  if (savedTab === 'files' || savedTab === 'contexts' || savedTab === 'git') {
    currentTab.value = savedTab
  }
} catch (err) {
  console.warn('Failed to load sidebar tab:', err)
}
</script>

<style scoped>
.sidebar-container {
  @apply h-full flex flex-col;
  background: transparent;
}

.tabs-container {
  @apply relative flex gap-1 p-2;
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-1);
}

/* Sliding Indicator */
.tabs-indicator {
  @apply absolute top-2 bottom-2 rounded-xl;
  width: calc(33.333% - 8px);
  left: 8px;
  transition: transform 250ms cubic-bezier(0.4, 0, 0.2, 1),
              background 200ms ease-out,
              box-shadow 200ms ease-out;
  z-index: 0;
  pointer-events: none;
}

.tabs-indicator-indigo {
  background: rgba(99, 102, 241, 0.25);
  border: 1px solid rgba(99, 102, 241, 0.4);
  box-shadow: 0 0 12px rgba(99, 102, 241, 0.3);
}

.tabs-indicator-orange {
  background: rgba(249, 115, 22, 0.25);
  border: 1px solid rgba(249, 115, 22, 0.4);
  box-shadow: 0 0 12px rgba(249, 115, 22, 0.3);
}

.tabs-indicator-purple {
  background: rgba(168, 85, 247, 0.25);
  border: 1px solid rgba(168, 85, 247, 0.4);
  box-shadow: 0 0 12px rgba(168, 85, 247, 0.3);
}

.sidebar-tab {
  @apply flex-1 flex items-center justify-center gap-2;
  @apply px-3 py-2 text-sm font-medium rounded-xl;
  @apply relative z-10;
  color: var(--text-muted);
  transition: color 150ms ease-out, transform 100ms ease-out;
}

.sidebar-tab:hover:not(.sidebar-tab-active) {
  color: var(--text-secondary);
}

.sidebar-tab:active {
  transform: scale(0.97);
}

.sidebar-tab-active {
  @apply font-semibold;
}

.tab-badge {
  @apply px-1.5 py-0.5 text-xs font-bold rounded-full;
  background: rgba(255, 255, 255, 0.25);
  color: white;
}

.sidebar-content {
  @apply flex-1 overflow-hidden;
}
</style>
