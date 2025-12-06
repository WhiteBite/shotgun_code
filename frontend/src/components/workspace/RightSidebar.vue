<template>
  <div class="h-full flex flex-col bg-transparent">
    <!-- Tab Switcher - Icons only with tooltips -->
    <div class="flex gap-1 p-2 border-b border-gray-700/30">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="currentTab = tab.id"
        :title="t(tab.label)"
        :class="[
          'flex-1 p-2 rounded-lg transition-all flex items-center justify-center',
          currentTab === tab.id 
            ? `${tab.activeClass}` 
            : 'text-gray-400 hover:text-white hover:bg-gray-800/50'
        ]"
      >
        <component :is="tab.icon" class="w-4 h-4" />
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto scrollable-y">
      <!-- Stats Tab -->
      <div v-if="currentTab === 'stats'" class="p-3 space-y-4">
        <FileTypeStats />
      </div>

      <!-- Export Settings Tab -->
      <div v-else-if="currentTab === 'export'" class="p-3 space-y-4">
        <ExportSettings />
      </div>

      <!-- Prompts Tab -->
      <div v-else-if="currentTab === 'prompts'" class="p-3 space-y-3">
        <PromptTemplates />
      </div>

      <!-- AI Chat Tab -->
      <div v-else-if="currentTab === 'chat'" class="h-full flex flex-col">
        <AIChat />
      </div>

      <!-- Settings Tab -->
      <div v-else-if="currentTab === 'settings'" class="p-3 space-y-4">
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

// Icons as render functions
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
  { id: 'stats' as TabId, label: 'sidebar.stats', icon: ChartIcon, activeClass: 'tab-btn-active-indigo' },
  { id: 'export' as TabId, label: 'sidebar.exportSettings', icon: ExportIcon, activeClass: 'tab-btn-active-emerald' },
  { id: 'prompts' as TabId, label: 'sidebar.prompts', icon: TemplateIcon, activeClass: 'tab-btn-active-orange' },
  { id: 'chat' as TabId, label: 'sidebar.chat', icon: ChatIcon, activeClass: 'tab-btn-active-purple' },
  { id: 'settings' as TabId, label: 'sidebar.settings', icon: SettingsIcon, activeClass: 'tab-btn-active-indigo' },
]

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
