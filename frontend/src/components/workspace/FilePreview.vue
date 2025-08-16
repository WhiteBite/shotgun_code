<template>
  <div class="file-preview-container p-4 border border-gray-700 bg-gray-800/50 rounded-lg flex flex-col min-h-0">
    <div class="file-preview-header flex items-center justify-between text-xs text-gray-400 mb-2">
      <div class="flex items-center gap-2">
        <h3 class="font-semibold">Context Preview</h3>
        <KeyboardShortcutsIcon />
      </div>
      <ContextActionsBar :hasContent="hasContent">
        <SplitSettingsPopover
            v-model="splitSettings"
            :preview="splitPreview"
            @apply="refreshPreview"
            @refresh="refreshPreview"
        />
        <CopyMenuButton
            v-model:exportFormat="exportFormat"
            v-model:stripComments="stripComments"
            :splitEnabled="splitSettings.enableAutoSplit"
            :splitPreview="splitPreview"
            @do-copy="copy"
            @open-export="openExportModal"
        />
      </ContextActionsBar>
    </div>

    <div class="file-preview-tabs flex border-b border-gray-700" v-if="showTabs">
      <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="tabClasses(tab.id)"
          class="px-3 py-1 text-xs font-medium transition-colors"
      >
        {{ tab.label }}
        <span v-if="tab.count" class="ml-1 px-1 bg-gray-600 rounded-full text-[10px]">
          {{ tab.count }}
        </span>
      </button>
    </div>

    <div class="file-preview-content relative flex-grow bg-gray-900 rounded-md border border-gray-700 min-h-0 overflow-hidden">
      <div v-if="hasContent || hasSplitContent" class="absolute top-2 right-2 z-10 flex items-center gap-2">
        <button
            @click="copy({ target: 'all', format: exportFormat, stripComments })"
            class="p-2 bg-gray-800/50 hover:bg-gray-700 rounded transition-colors"
            title="Copy to clipboard"
            aria-label="Copy"
        >
          <CopyIcon class="w-4 h-4" />
        </button>
        <button
            @click="openExportModal()"
            class="p-2 bg-gray-800/50 hover:bg-gray-700 rounded transition-colors"
            title="Export options"
            aria-label="Export options"
        >
          <ExportIcon class="w-4 h-4" />
        </button>
      </div>

      <!-- Context Tab -->
      <div v-show="activeTab === 'context'" class="h-full">
        <div v-if="!hasContent" class="flex items-center justify-center h-full text-gray-500">
          <div class="text-center">
            <DocumentIcon class="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>Context will appear here after building...</p>
            <p class="text-xs mt-1">Select files and click "Build Context"</p>
          </div>
        </div>
        <ScrollableContent v-else :content="contextBuilderStore.shotgunContextText" :virtualize="true" />
      </div>

      <!-- Split Preview Tab -->
      <div v-show="activeTab === 'split'" class="h-full">
        <div v-if="!hasSplitContent" class="flex items-center justify-center h-full text-gray-500">
          <div class="text-center">
            <CodeIcon class="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>Split preview will appear here...</p>
            <p class="text-xs mt-1">Enable auto-split and build context</p>
          </div>
        </div>
        <SplitPreviewTab
            v-else
            :splitPreview="splitPreview"
            @copy-chunk="copyChunk"
        />
      </div>

      <!-- Generated Tab -->
      <div v-show="activeTab === 'generated'" class="h-full">
        <div v-if="!generationStore.hasResult" class="flex items-center justify-center h-full text-gray-500">
          <div class="text-center">
            <CodeIcon class="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>Generated code will appear here...</p>
            <p class="text-xs mt-1">Describe your task and click "Generate"</p>
          </div>
        </div>
        <ScrollableContent v-else :content="generationStore.generatedDiff" language="diff" :highlight="true" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useGenerationStore } from '@/stores/generation.store'
import { useExportStore } from '@/stores/export.store'
import { useUiStore } from '@/stores/ui.store'

import ScrollableContent from '@/components/shared/ScrollableContent.vue'
import KeyboardShortcutsIcon from '@/components/workspace/KeyboardShortcutsIcon.vue'
import SplitPreviewTab from '@/components/workspace/SplitPreviewTab.vue'
import { CopyIcon, ExportIcon, DocumentIcon, CodeIcon } from '@/components/icons/index'

import ContextActionsBar from '@/components/workspace/ContextActions/ContextActionsBar.vue'
import CopyMenuButton from '@/components/workspace/ContextActions/CopyMenuButton.vue'
import SplitSettingsPopover from '@/components/workspace/ContextActions/SplitSettingsPopover.vue'
import { useContextActions } from '@/composables/useContextActions'

const contextBuilderStore = useContextBuilderStore()
const generationStore = useGenerationStore()
const exportStore = useExportStore()
const uiStore = useUiStore()

const { exportFormat, stripComments, splitSettings, splitPreview, refreshPreview, copy, openExportModal } = useContextActions()

const activeTab = ref('context')

const hasContent = computed(() => !!contextBuilderStore.shotgunContextText)
const hasSplitContent = computed(() =>
    splitSettings.value.enableAutoSplit &&
    splitPreview.value &&
    splitPreview.value.chunkCount > 1
)
const showTabs = computed(() => hasContent.value || generationStore.hasResult || hasSplitContent.value)

const tabs = computed(() => {
  const tabList = [
    { id: 'context', label: 'Context', count: contextBuilderStore.selectedFiles.length }
  ];

  if (hasSplitContent.value) {
    tabList.push({
      id: 'split',
      label: 'Split Preview',
      count: splitPreview.value?.chunkCount || 0
    });
  }

  if (generationStore.hasResult) {
    tabList.push({
      id: 'generated',
      label: 'Generated',
      count: 1
    });
  }
  return tabList;
})

function tabClasses(tabId: string) {
  return {
    'bg-gray-700 text-white': activeTab.value === tabId,
    'text-gray-400 hover:text-white hover:bg-gray-800': activeTab.value !== tabId,
  }
}

function copyChunk(chunkIndex: number) {
  copy({
    target: 'chunk',
    chunkIndex,
    format: exportFormat.value,
    stripComments: stripComments.value
  });
}

watch(hasSplitContent, (hasContent) => {
  if (hasContent && activeTab.value === 'context') {
    activeTab.value = 'split';
  }
});

watch(() => generationStore.hasResult, (hasResult) => {
  if (hasResult) activeTab.value = 'generated';
});
</script>