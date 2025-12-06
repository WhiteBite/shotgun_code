<template>
  <div class="sidebar-container">
    <!-- Tab Switcher with Sliding Indicator -->
    <div class="tabs-container">
      <!-- Sliding Indicator -->
      <div 
        class="tabs-indicator-icon"
        :class="tabIndicatorClass"
        :style="{ transform: `translateX(${tabIndex * 100}%)` }"
      ></div>
      
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="currentTab = tab.id"
        :title="t(tab.label)"
        :class="['sidebar-tab-icon', currentTab === tab.id ? `sidebar-tab-icon-active ${tab.textClass}` : '']"
      >
        <component :is="tab.icon" class="w-4 h-4" />
      </button>
    </div>

    <!-- Content -->
    <div class="sidebar-content scrollable-y">
      <!-- Stats Tab -->
      <div v-if="currentTab === 'stats'" class="tab-content">
        <FileTypeStats />
      </div>

      <!-- Export Settings Tab -->
      <div v-else-if="currentTab === 'export'" class="tab-content">
        <ExportSettings />
      </div>

      <!-- Prompts Tab -->
      <div v-else-if="currentTab === 'prompts'" class="tab-content">
        <PromptTemplates />
      </div>

      <!-- AI Chat Tab -->
      <div v-else-if="currentTab === 'chat'" class="tab-content-full">
        <AIChat />
      </div>

      <!-- Settings Tab -->
      <div v-else-if="currentTab === 'settings'" class="tab-content">
        <AISettings />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { h, ref, watch } from 'vue'
import AIChat from './sidebar/AIChat.vue'
import AISettings from './sidebar/AISettings.vue'
import ExportSettings from './sidebar/ExportSettings.vue'
import FileTypeStats from './sidebar/FileTypeStats.vue'
import PromptTemplates from './sidebar/PromptTemplates.vue'

const { t } = useI18n()

type TabId = 'stats' | 'export' | 'prompts' | 'chat' | 'settings'
const currentTab = ref<TabId>('stats')

// Icons
const ChartIcon = () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' })
])

const ExportIcon = () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' })
])

const TemplateIcon = () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' })
])

const ChatIcon = () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' })
])

const SettingsIcon = () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z' }),
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M15 12a3 3 0 11-6 0 3 3 0 016 0z' })
])

const tabs = [
  { id: 'stats' as TabId, label: 'sidebar.stats', icon: ChartIcon, textClass: 'text-indigo-300', indicatorClass: 'tabs-indicator-indigo' },
  { id: 'export' as TabId, label: 'sidebar.exportSettings', icon: ExportIcon, textClass: 'text-emerald-300', indicatorClass: 'tabs-indicator-emerald' },
  { id: 'prompts' as TabId, label: 'sidebar.prompts', icon: TemplateIcon, textClass: 'text-orange-300', indicatorClass: 'tabs-indicator-orange' },
  { id: 'chat' as TabId, label: 'sidebar.chat', icon: ChatIcon, textClass: 'text-purple-300', indicatorClass: 'tabs-indicator-purple' },
  { id: 'settings' as TabId, label: 'sidebar.settings', icon: SettingsIcon, textClass: 'text-indigo-300', indicatorClass: 'tabs-indicator-indigo' },
]

// Tab indicator computed
import { computed } from 'vue'

const tabIndex = computed(() => {
  return tabs.findIndex(t => t.id === currentTab.value)
})

const tabIndicatorClass = computed(() => {
  const tab = tabs.find(t => t.id === currentTab.value)
  return tab?.indicatorClass || 'tabs-indicator-indigo'
})

// Persist tab selection
watch(currentTab, (tab) => {
  try {
    localStorage.setItem('right-sidebar-tab', tab)
  } catch (err) {
    console.warn('Failed to save sidebar tab:', err)
  }
})

// Restore tab selection
try {
  const savedTab = localStorage.getItem('right-sidebar-tab') as TabId
  if (savedTab && tabs.some(t => t.id === savedTab)) {
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

/* Sliding Indicator for Icon Tabs */
.tabs-indicator-icon {
  @apply absolute top-2 bottom-2 rounded-lg;
  width: calc(20% - 8px);
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

.tabs-indicator-emerald {
  background: rgba(16, 185, 129, 0.25);
  border: 1px solid rgba(16, 185, 129, 0.4);
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.3);
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

.sidebar-tab-icon {
  @apply flex-1 flex items-center justify-center;
  @apply p-2.5 rounded-lg;
  @apply relative z-10;
  color: var(--text-muted);
  transition: color 150ms ease-out, transform 100ms ease-out;
}

.sidebar-tab-icon:hover:not(.sidebar-tab-icon-active) {
  color: var(--text-secondary);
}

.sidebar-tab-icon:active {
  transform: scale(0.93);
}

.sidebar-tab-icon-active {
  @apply font-semibold;
}

.sidebar-content {
  @apply flex-1 overflow-y-auto;
}

.tab-content {
  @apply p-3 space-y-4;
}

.tab-content-full {
  @apply h-full flex flex-col;
}
</style>
