<template>
  <div class="right-sidebar-container">
    <!-- Tab Switcher with Sliding Indicator -->
    <div class="right-sidebar-tabs">
      <!-- Sliding Indicator -->
      <div class="tabs-indicator-icon" :class="tabIndicatorClass"
        :style="{ transform: `translateX(calc(${tabIndex} * (2rem + 0.25rem)))` }"></div>

      <button 
        v-for="tab in tabs" 
        :key="tab.id" 
        @click="currentTab = tab.id" 
        :title="t(tab.label)"
        :aria-label="t(tab.label)"
        :aria-pressed="currentTab === tab.id"
        :class="['sidebar-tab-icon', currentTab === tab.id ? `sidebar-tab-icon-active ${tab.textClass}` : '']"
      >
        <component :is="tab.icon" class="w-4 h-4" aria-hidden="true" />
      </button>
    </div>

    <!-- Content -->
    <div class="right-sidebar-content scrollable-y">
      <!-- Stats Tab -->
      <div v-if="currentTab === 'stats'" class="right-sidebar-tab-content">
        <FileTypeStats />
      </div>

      <!-- Export Settings Tab -->
      <div v-else-if="currentTab === 'export'" class="right-sidebar-tab-content">
        <ExportSettings />
      </div>

      <!-- Prompts Tab -->
      <div v-else-if="currentTab === 'prompts'" class="right-sidebar-tab-content">
        <PromptTemplates />
      </div>

      <!-- AI Chat Tab -->
      <div v-else-if="currentTab === 'chat'" class="right-sidebar-tab-content-full" data-tour="ai-chat">
        <AIChat />
      </div>

      <!-- Settings Tab -->
      <div v-else-if="currentTab === 'settings'" class="right-sidebar-tab-content">
        <AISettings />
      </div>

      <!-- Memory Tab -->
      <div v-else-if="currentTab === 'memory'" class="right-sidebar-tab-content-full">
        <ContextMemoryPanel />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { BarChart3, BookMarked, FileDown, FileText, MessageCircle, Settings } from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'
import AIChat from './sidebar/AIChat.vue'
import AISettings from './sidebar/AISettings.vue'
import ContextMemoryPanel from './sidebar/ContextMemoryPanel.vue'
import ExportSettings from './sidebar/ExportSettings.vue'
import FileTypeStats from './sidebar/FileTypeStats.vue'
import PromptTemplates from './sidebar/PromptTemplates.vue'

const { t } = useI18n()

type TabId = 'stats' | 'export' | 'prompts' | 'chat' | 'settings' | 'memory'
const currentTab = ref<TabId>('chat')

const tabs = [
  { id: 'stats' as TabId, label: 'sidebar.stats', icon: BarChart3, textClass: 'text-indigo-300', indicatorClass: 'tabs-indicator-indigo' },
  { id: 'export' as TabId, label: 'sidebar.exportSettings', icon: FileDown, textClass: 'text-emerald-300', indicatorClass: 'tabs-indicator-emerald' },
  { id: 'prompts' as TabId, label: 'sidebar.prompts', icon: FileText, textClass: 'text-orange-300', indicatorClass: 'tabs-indicator-orange' },
  { id: 'chat' as TabId, label: 'sidebar.chat', icon: MessageCircle, textClass: 'text-purple-300', indicatorClass: 'tabs-indicator-purple' },
  { id: 'memory' as TabId, label: 'sidebar.memory', icon: BookMarked, textClass: 'text-cyan-300', indicatorClass: 'tabs-indicator-cyan' },
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

<style>
.right-sidebar-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: transparent;
}

.right-sidebar-tabs {
  position: relative;
  display: flex;
  gap: 0.25rem;
  padding: 0.5rem;
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-1);
}

.right-sidebar-content {
  flex: 1;
  overflow-y: auto;
}

.right-sidebar-tab-content {
  padding: 0.75rem;
}

.right-sidebar-tab-content > * + * {
  margin-top: 1rem;
}

.right-sidebar-tab-content-full {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>
