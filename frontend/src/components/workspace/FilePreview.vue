<template>
  <div class="p-4 border border-gray-700 bg-gray-800/50 rounded-lg flex flex-col flex-grow min-h-0">
    <div class="flex justify-between items-center mb-3">
      <h2 class="text-lg font-semibold text-white">
        {{ tab === 'context' ? 'Project Context' : (tab === 'result' ? 'Generated Diff' : 'File Preview') }}
      </h2>
      <div class="flex items-center gap-1 p-0.5 bg-gray-900/50 rounded-md text-sm">
        <button @click="tab = 'preview'" :class="['px-3 py-1 rounded-md', tab === 'preview' ? 'bg-blue-600' : 'hover:bg-gray-700']">Preview</button>
        <button @click="tab = 'context'" :class="['px-3 py-1 rounded-md', tab === 'context' ? 'bg-blue-600' : 'hover:bg-gray-700']">Context</button>
        <button @click="tab = 'result'" :disabled="!generationStore.hasResult" :class="['px-3 py-1 rounded-md', tab === 'result' ? 'bg-blue-600' : 'hover:bg-gray-700 disabled:text-gray-600 disabled:cursor-not-allowed']">Result</button>
      </div>
    </div>
    <div class="flex-grow bg-gray-900 rounded-md p-3 overflow-auto border border-gray-700 min-h-0">
      <div v-if="tab === 'preview'">
        <pre v-if="isLoading" class="text-gray-500">Loading file...</pre>
        <pre v-else-if="error" class="text-red-400">{{ error }}</pre>
        <pre v-else-if="fileContent"><code class="hljs" v-html="highlightedContent"></code></pre>
        <pre v-else class="text-gray-500">Select a file to preview its content.</pre>
      </div>
      <pre v-if="tab === 'context'" class="text-xs font-mono text-gray-300 whitespace-pre-wrap"><code>{{ contextStore.shotgunContextText || 'Context will appear here after being built.' }}</code></pre>
      <pre v-if="tab === 'result'" class="text-sm font-mono whitespace-pre-wrap"><code v-html="highlightedDiff"></code></pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useContextStore } from '@/stores/context.store';
import { useGenerationStore } from '@/stores/generation.store';
import { useProjectStore } from '@/stores/project.store';
import { apiService } from '@/services/api.service';
import hljs from 'highlight.js';

const tab = ref<'preview' | 'context' | 'result'>('preview');
const contextStore = useContextStore();
const generationStore = useGenerationStore();
const projectStore = useProjectStore();

const fileContent = ref('');
const isLoading = ref(false);
const error = ref<string | null>(null);

const highlightedContent = computed(() => {
  if (!fileContent.value) return '';
  const lang = contextStore.activeNode?.name.split('.').pop() || 'plaintext';
  if (hljs.getLanguage(lang)) {
    return hljs.highlight(fileContent.value, { language: lang }).value;
  }
  return hljs.highlightAuto(fileContent.value).value;
});

const highlightedDiff = computed(() => {
  if (!generationStore.generatedDiff) return '<span class="text-gray-500">No diff generated yet.</span>';
  return hljs.highlight(generationStore.generatedDiff, { language: 'diff' }).value;
});

watch(() => contextStore.activeNodePath, async (newPath) => {
  if (!newPath || !projectStore.currentProject) {
    fileContent.value = '';
    return;
  }
  const node = contextStore.nodesMap.get(newPath);
  if (!node || node.isDir) {
    fileContent.value = '';
    return;
  }

  isLoading.value = true;
  error.value = null;
  fileContent.value = '';
  tab.value = 'preview';
  try {
    fileContent.value = await apiService.readFileContent(projectStore.currentProject.path, node.relPath);
  } catch(e: any) {
    error.value = e.message || 'Failed to read file';
  } finally {
    isLoading.value = false;
  }
});

watch(() => generationStore.hasResult, (hasResult) => {
  if (hasResult) {
    tab.value = 'result';
  }
});
</script>

<style>
/* Basic styles for highlight.js default theme on dark background */
.hljs {
  color: #abb2bf;
  background: #282c34;
}
.hljs-comment, .hljs-quote {
  color: #5c6370;
  font-style: italic;
}
.hljs-doctag, .hljs-keyword, .hljs-formula {
  color: #c678dd;
}
.hljs-section, .hljs-name, .hljs-selector-tag, .hljs-deletion, .hljs-subst {
  color: #e06c75;
}
.hljs-literal {
  color: #56b6c2;
}
.hljs-string, .hljs-regexp, .hljs-addition, .hljs-attribute, .hljs-meta-string {
  color: #98c379;
}
.hljs-built_in, .hljs-class .hljs-title {
  color: #e6c07b;
}
.hljs-attr, .hljs-variable, .hljs-template-variable, .hljs-type, .hljs-selector-class, .hljs-selector-attr, .hljs-selector-pseudo, .hljs-number {
  color: #d19a66;
}
.hljs-symbol, .hljs-bullet, .hljs-link, .hljs-meta, .hljs-selector-id, .hljs-title {
  color: #61afef;
}
.hljs-emphasis {
  font-style: italic;
}
.hljs-strong {
  font-weight: bold;
}
</style>