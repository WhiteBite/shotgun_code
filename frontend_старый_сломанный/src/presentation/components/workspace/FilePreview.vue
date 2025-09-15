<template>
  <div
    class="file-preview-container p-4 border border-gray-700 bg-gray-800/50 rounded-lg flex flex-col min-h-0"
  >
    <div
      class="file-preview-header flex items-center justify-between text-xs text-gray-400 mb-2"
    >
      <div class="flex items-center gap-2">
        <h3 class="font-semibold">Context Preview</h3>
        <KeyboardShortcutsIcon />
      </div>
      <ContextActionsBar :has-content="hasContent">
        <SplitSettingsPopover
          v-model="splitSettings"
          :preview="splitPreview"
          @apply="refreshPreview"
          @refresh="refreshPreview"
        />
        <CopyMenuButton
          v-model:export-format="exportFormat"
          v-model:strip-comments="stripComments"
          :split-enabled="splitSettings.enableAutoSplit"
          :split-preview="splitPreview"
          @do-copy="copy"
          @open-export="openExportModal"
        />
      </ContextActionsBar>
    </div>

    <div
      v-if="showTabs"
      class="file-preview-tabs flex border-b border-gray-700"
    >
      <button
        v-for="tab in tabs"
        :key="tab.id"
        :class="tabClasses(tab.id)"
        class="px-3 py-1 text-xs font-medium transition-colors"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
        <span
          v-if="tab.count"
          class="ml-1 px-1 bg-gray-600 rounded-full text-[10px]"
        >
          {{ tab.count }}
        </span>
      </button>
    </div>

    <div
      class="file-preview-content relative flex-grow bg-gray-900 rounded-md border border-gray-700 min-h-0 overflow-hidden"
    >
      <div
        v-if="hasContent || hasSplitContent"
        class="absolute top-2 right-2 z-10 flex items-center gap-2"
      >
        <button
          class="p-2 bg-gray-800/50 hover:bg-gray-700 rounded transition-colors"
          title="Copy to clipboard"
          aria-label="Copy"
          @click="copy({ target: 'all', format: exportFormat, stripComments })"
        >
          <CopyIcon class="w-4 h-4" />
        </button>
        <button
          class="p-2 bg-gray-800/50 hover:bg-gray-700 rounded transition-colors"
          title="Export options"
          aria-label="Export options"
          @click="openExportModal()"
        >
          <ExportIcon class="w-4 h-4" />
        </button>
      </div>

      <!-- Context Tab -->
      <div v-show="activeTab === 'context'" class="h-full">
        <div
          v-if="!hasContent"
          class="flex items-center justify-center h-full text-gray-500"
        >
          <div class="text-center">
            <DocumentIcon class="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>Context will appear here after building...</p>
            <p class="text-xs mt-1">Select files and click "Build Context"</p>
          </div>
        </div>
        <ScrollableContent
          v-else
          :content="contextContent"
          :virtualize="true"
        />
      </div>

      <!-- Split Preview Tab -->
      <div v-show="activeTab === 'split'" class="h-full">
        <div
          v-if="!hasSplitContent"
          class="flex items-center justify-center h-full text-gray-500"
        >
          <div class="text-center">
            <CodeIcon class="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>Split preview will appear here...</p>
            <p class="text-xs mt-1">Enable auto-split and build context</p>
          </div>
        </div>
        <SplitPreviewTab
          v-else
          :split-preview="splitPreview"
          @copy-chunk="copyChunk"
        />
      </div>

      <!-- Generated Tab -->
      <div v-show="activeTab === 'generated'" class="h-full">
        <div
          v-if="!generationStore.hasResult"
          class="flex items-center justify-center h-full text-gray-500"
        >
          <div class="text-center">
            <CodeIcon class="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>Generated code will appear here...</p>
            <p class="text-xs mt-1">Describe your task and click "Generate"</p>
          </div>
        </div>
        <ScrollableContent
          v-else
          :content="generationStore.generatedDiff"
          language="diff"
          :highlight="true"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useGenerationStore } from "@/stores/generation.store";
import { useExportStore } from "@/stores/export.store";
import { useUiStore } from "@/stores/ui.store";

import ScrollableContent from "@/presentation/components/shared/ScrollableContent.vue";
import KeyboardShortcutsIcon from "@/presentation/components/workspace/KeyboardShortcutsIcon.vue";
import SplitPreviewTab from "@/presentation/components/workspace/SplitPreviewTab.vue";
import {
  CopyIcon,
  ExportIcon,
  DocumentIcon,
  CodeIcon,
} from "@/presentation/components/icons/index";

import ContextActionsBar from "@/presentation/components/workspace/ContextActions/ContextActionsBar.vue";
import CopyMenuButton from "@/presentation/components/workspace/ContextActions/CopyMenuButton.vue";
import SplitSettingsPopover from "@/presentation/components/workspace/ContextActions/SplitSettingsPopover.vue";

const contextBuilderStore = useContextBuilderStore();
const generationStore = useGenerationStore();
const exportStore = useExportStore();
const uiStore = useUiStore();

// State for paginated context content
const contextContent = ref<string>('');

const exportFormat = computed({
  get: () => exportStore.exportFormat,
  set: (v) => (exportStore.exportFormat = v),
});
const stripComments = computed({
  get: () => exportStore.stripComments,
  set: (v) => (exportStore.stripComments = v),
});
const splitSettings = computed({
  get: () => ({
    enableAutoSplit: exportStore.enableAutoSplit,
    maxTokensPerChunk: exportStore.maxTokensPerChunk,
    overlapTokens: exportStore.overlapTokens,
    splitStrategy: exportStore.splitStrategy,
  }),
  set: (v) => {
    exportStore.enableAutoSplit = v.enableAutoSplit;
    exportStore.maxTokensPerChunk = v.maxTokensPerChunk || 50000;
    exportStore.overlapTokens = v.overlapTokens || 1000;
    exportStore.splitStrategy = v.splitStrategy || "token";
  },
});
const splitPreview = computed(() => exportStore.splitPreview);
const refreshPreview = () => exportStore.computeSplitPreview();
const copy = (req: any) => exportStore.copy(req);
const openExportModal = () => exportStore.open();

const activeTab = ref("context");

// Updated to check for context summary instead of shotgunContextText
const hasContent = computed(() => !!contextBuilderStore.contextSummaryState?.id);
const hasSplitContent = computed(
  () =>
    splitSettings.value.enableAutoSplit &&
    splitPreview.value &&
    splitPreview.value.chunkCount > 1,
);
const showTabs = computed(
  () => hasContent.value || generationStore.hasResult || hasSplitContent.value,
);

const tabs = computed(() => {
  const tabList = [
    {
      id: "context",
      label: "Context",
      count: contextBuilderStore.selectedFilesList?.length || 0,
    },
  ];

  if (hasSplitContent.value) {
    tabList.push({
      id: "split",
      label: "Split Preview",
      count: splitPreview.value?.chunkCount || 0,
    });
  }

  if (generationStore.hasResult) {
    tabList.push({
      id: "generated",
      label: "Generated",
      count: 1,
    });
  }
  return tabList;
});

// Load context content when context is available
const loadContextContent = async () => {
  if (contextBuilderStore.contextSummaryState?.id) {
    try {
      // Load first chunk of context content
      const chunk = await contextBuilderStore.getContextContent(0, 1000);
      if (chunk) {
        contextContent.value = chunk.content;
      } else {
        contextContent.value = '';
      }
    } catch (error) {
      console.error('Failed to load context content:', error);
      contextContent.value = 'Failed to load context content';
    }
  } else {
    contextContent.value = '';
  }
};

// Watch for context changes and load content
watch(() => contextBuilderStore.contextSummaryState?.id, loadContextContent, { immediate: true });

function tabClasses(tabId: string) {
  return {
    "bg-gray-700 text-white": activeTab.value === tabId,
    "text-gray-400 hover:text-white hover:bg-gray-800":
      activeTab.value !== tabId,
  };
}

function copyChunk(chunkIndex: number) {
  copy({
    target: "chunk",
    chunkIndex,
    format: exportFormat.value,
    stripComments: stripComments.value,
  });
}

watch(hasSplitContent, (hasContent) => {
  if (hasContent && activeTab.value === "context") {
    activeTab.value = "split";
  }
});

watch(
  () => generationStore.hasResult,
  (hasResult) => {
    if (hasResult) activeTab.value = "generated";
  },
);
</script>