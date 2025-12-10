<template>
  <div class="sidebar-container">
    <!-- Tab Switcher with Sliding Indicator -->
    <div class="tabs-container">
      <!-- Sliding Indicator -->
      <div class="tabs-indicator-icon" :class="tabIndicatorClass"
        :style="{ transform: `translateX(${tabIndex * 100}%)` }"></div>

      <button v-for="tab in tabs" :key="tab.id" @click="currentTab = tab.id" :title="t(tab.label)"
        :class="['sidebar-tab-icon', currentTab === tab.id ? `sidebar-tab-icon-active ${tab.textClass}` : '']">
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
import { BarChart3, FileDown, FileText, MessageCircle, Settings } from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'
import AIChat from './sidebar/AIChat.vue'
import AISettings from './sidebar/AISettings.vue'
import ExportSettings from './sidebar/ExportSettings.vue'
import FileTypeStats from './sidebar/FileTypeStats.vue'
import PromptTemplates from './sidebar/PromptTemplates.vue'

const { t } = useI18n()

type TabId = 'stats' | 'export' | 'prompts' | 'chat' | 'settings'
const currentTab = ref<TabId>('chat')

const tabs = [
  { id: 'stats' as TabId, label: 'sidebar.stats', icon: BarChart3, textClass: 'text-indigo-300', indicatorClass: 'tabs-indicator-indigo' },
  { id: 'export' as TabId, label: 'sidebar.exportSettings', icon: FileDown, textClass: 'text-emerald-300', indicatorClass: 'tabs-indicator-emerald' },
  { id: 'prompts' as TabId, label: 'sidebar.prompts', icon: FileText, textClass: 'text-orange-300', indicatorClass: 'tabs-indicator-orange' },
  { id: 'chat' as TabId, label: 'sidebar.chat', icon: MessageCircle, textClass: 'text-purple-300', indicatorClass: 'tabs-indicator-purple' },
  { id: 'settings' as TabId, label: 'sidebar.settings', icon: Settings, textClass: 'text-indigo-300', indicatorClass: 'tabs-indicator-indigo' },
]

const tabIndex = computed(() => tabs.findIndex(t => t.id === currentTab.value))
const tabIndicatorClass = computed(() => tabs.find(t => t.id === currentTab.value)?.indicatorClass || 'tabs-indicator-indigo')

// Persist tab selection
watch(currentTab, (tab) => {
  try { localStorage.setItem('right-sidebar-tab', tab) } catch { /* ignore */ }
})

// Restore tab selection
try {
  const savedTab = localStorage.getItem('right-sidebar-tab') as TabId
  if (savedTab && tabs.some(t => t.id === savedTab)) {
    currentTab.value = savedTab
  }
} catch { /* ignore */ }
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

/* tabs-indicator-* and sidebar-tab-icon moved to design-tokens.css */

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
