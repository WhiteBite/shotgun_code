<template>
  <div class="h-full flex flex-col bg-transparent">
    <!-- Tab Switcher -->
    <div class="flex gap-1 p-2 border-b border-gray-700/30">
      <button
        @click="currentTab = 'files'"
        :class="['tab-btn', currentTab === 'files' ? 'tab-btn-active tab-btn-active-indigo' : 'tab-btn-inactive']"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        {{ t('tabs.files') }}
        <span v-if="fileStore.selectedPaths.size > 0" class="badge badge-primary">
          {{ fileStore.selectedPaths.size }}
        </span>
      </button>
      <button
        @click="currentTab = 'git'"
        :class="['tab-btn', currentTab === 'git' ? 'tab-btn-active tab-btn-active-orange' : 'tab-btn-inactive']"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
        </svg>
        Git
      </button>
      <button
        @click="currentTab = 'contexts'"
        :class="['tab-btn', currentTab === 'contexts' ? 'tab-btn-active tab-btn-active-purple' : 'tab-btn-inactive']"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        {{ t('tabs.contexts') }}
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-hidden">
      <FileExplorer v-if="currentTab === 'files'" @preview-file="handlePreviewFile" />
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

const emit = defineEmits<{
  (e: 'preview-file', filePath: string): void
}>()

function handlePreviewFile(filePath: string) {
  emit('preview-file', filePath)
}

// Save tab to localStorage
watch(currentTab, (tab) => {
  try {
    localStorage.setItem('left-sidebar-tab', tab)
  } catch (err) {
    console.warn('Failed to save sidebar tab:', err)
  }
})

// Restore tab from localStorage
try {
  const savedTab = localStorage.getItem('left-sidebar-tab')
  if (savedTab === 'files' || savedTab === 'contexts' || savedTab === 'git') {
    currentTab.value = savedTab
  }
} catch (err) {
  console.warn('Failed to load sidebar tab:', err)
}
</script>
