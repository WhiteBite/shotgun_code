<template>
  <div
    class="p-4 border border-gray-700 bg-gray-800/50 rounded-lg flex flex-col flex-grow min-h-0"
  >
    <!-- Заголовок + формат + копирование + табы -->
    <div class="flex flex-wrap items-center justify-between gap-3 mb-3">
      <div class="flex items-center gap-3 min-w-0">
        <h2 class="text-lg font-semibold text-white truncate">
          {{ tabTitle }}
        </h2>

        <!-- Источники (как было) -->
        <div class="flex items-center gap-1">
          <span
            v-for="origin in contextOrigins"
            :key="origin"
            class="px-2 py-0.5 text-xs rounded-full"
            :class="originColors[origin].bg"
            :title="`Contains files selected from: ${origin}`"
          >
            {{ origin }}
          </span>
        </div>
      </div>

      <!-- Формат и копирование для Clipboard -->
      <div class="flex items-center gap-2">
        <div
          class="hidden md:flex items-center gap-1 bg-gray-800/80 border border-gray-700 rounded-md overflow-hidden"
        >
          <button
            :class="seg('plain')"
            @click="exportStore.exportFormat = 'plain'"
            title="Simple headers + content. Most compatible."
          >
            Plain
          </button>
          <button
            :class="seg('manifest')"
            @click="exportStore.exportFormat = 'manifest'"
            title="Adds file tree manifest, then file contents. Recommended."
          >
            Manifest + Text
          </button>
          <button
            :class="seg('json')"
            @click="exportStore.exportFormat = 'json'"
            title="Machine-readable JSON array of files"
          >
            JSON
          </button>
        </div>

        <label
          class="hidden lg:inline-flex items-center gap-1 text-xs text-gray-400"
        >
          <input
            type="checkbox"
            class="form-checkbox"
            v-model="exportStore.stripComments"
          />
          Strip
        </label>
        <label
          class="hidden lg:inline-flex items-center gap-1 text-xs text-gray-400"
          :class="{ 'opacity-50': exportStore.exportFormat !== 'manifest' }"
          :title="
            exportStore.exportFormat !== 'manifest'
              ? 'Only for Manifest format'
              : ''
          "
        >
          <input
            type="checkbox"
            class="form-checkbox"
            v-model="exportStore.includeManifest"
            :disabled="exportStore.exportFormat !== 'manifest'"
          />
          Manifest
        </label>

        <button
          class="px-3 py-1.5 bg-gray-700 hover:bg-gray-600 rounded-md text-sm"
          :disabled="!contextStore.shotgunContextText || exportStore.isLoading"
          @click="exportStore.doExportClipboard()"
        >
          Copy
        </button>

        <!-- Табы -->
        <div
          class="flex items-center gap-1 p-0.5 bg-gray-900/50 rounded-md text-sm"
        >
          <button
            @click="tab = 'context'"
            :class="[
              'px-3 py-1 rounded-md',
              tab === 'context' ? 'bg-blue-600' : 'hover:bg-gray-700',
            ]"
          >
            Context
          </button>
          <button
            @click="tab = 'result'"
            :disabled="!generationStore.hasResult"
            :class="[
              'px-3 py-1 rounded-md',
              tab === 'result'
                ? 'bg-blue-600'
                : 'hover:bg-gray-700 disabled:text-gray-600 disabled:cursor-not-allowed',
            ]"
          >
            Result
          </button>
        </div>
      </div>
    </div>

    <div
      class="flex-grow bg-gray-900 rounded-md p-3 overflow-auto border border-gray-700 min-h-0"
      ref="contentArea"
    >
      <pre
        v-show="tab === 'context'"
        class="text-xs font-mono text-gray-300 whitespace-pre-wrap"
      ><code>{{ contextStore.shotgunContextText || 'Context will appear here after being built.' }}</code></pre>
      <pre
        v-show="tab === 'result'"
        class="text-sm font-mono whitespace-pre-wrap"
      ><code v-html="highlightedDiff"></code></pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useContextStore } from "@/stores/context.store";
import { useGenerationStore } from "@/stores/generation.store";
import { ContextOrigin } from "@/types/enums";
import hljs from "highlight.js";
import { useExportStore } from "@/stores/export.store";

const tab = ref<"context" | "result">("context");
const contextStore = useContextStore();
const generationStore = useGenerationStore();
const exportStore = useExportStore();

const originColors: Record<ContextOrigin, { bg: string }> = {
  [ContextOrigin.None]: { bg: "" },
  [ContextOrigin.Manual]: { bg: "bg-blue-500/30 text-blue-300" },
  [ContextOrigin.Git]: { bg: "bg-green-500/30 text-green-300" },
  [ContextOrigin.AI]: { bg: "bg-purple-500/30 text-purple-300" },
};

const tabTitle = computed(() => {
  switch (tab.value) {
    case "result":
      return "Generated Diff";
    case "context":
    default:
      return "Project Context";
  }
});

const contextOrigins = computed(() => {
  const origins = new Set<ContextOrigin>();
  contextStore.selectedFiles.forEach((file) => {
    if (file.contextOrigin !== ContextOrigin.None) {
      origins.add(file.contextOrigin);
    }
  });
  return Array.from(origins);
});

const highlightedDiff = computed(() => {
  if (!generationStore.generatedDiff)
    return '<span class="text-gray-500">No diff generated yet.</span>';
  return hljs.highlight(generationStore.generatedDiff, { language: "diff" })
    .value;
});

watch(
  () => generationStore.hasResult,
  (hasResult) => {
    if (hasResult) {
      tab.value = "result";
    } else {
      tab.value = "context";
    }
  },
);

function seg(v: string) {
  return [
    "px-2 py-1 text-xs",
    exportStore.exportFormat === v
      ? "bg-blue-600 text-white"
      : "text-gray-300 hover:bg-gray-700",
  ];
}
</script>
