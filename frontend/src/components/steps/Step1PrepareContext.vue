
<template>
  <!-- ИСПРАВЛЕНО: Восстановлена полная верстка (template) -->
  <div class="p-4 h-full flex flex-col">
    <!-- Loading State -->
    <div v-if="project.isGeneratingContext" class="flex-grow flex justify-center items-center">
      <div class="text-center">
        <div class="w-64 mx-auto">
          <p class="text-gray-600 mb-1 text-sm">Generating project context...</p>
          <div class="w-full bg-gray-200 rounded-full h-2.5">
            <div class="bg-blue-600 h-2.5 rounded-full" :style="{ width: progressBarWidth }"></div>
          </div>
          <p class="text-gray-500 mt-1 text-xs">
            {{ project.generationProgress.current }} / {{ project.generationProgress.total > 0 ? project.generationProgress.total : 'calculating...' }} items
          </p>
        </div>
      </div>
    </div>

    <!-- Content Area -->
    <div v-else-if="project.projectRoot" class="mt-0 flex-grow flex flex-col">
      <div v-if="project.shotgunPromptContext && !project.shotgunPromptContext.startsWith('Error:')" class="flex-grow flex flex-col">
        <h3 class="text-md font-medium text-gray-700 mb-2">Generated Project Context:</h3>
        <textarea
            :value="project.shotgunPromptContext"
            rows="10"
            readonly
            class="w-full p-2 border border-gray-300 rounded-md shadow-sm bg-gray-50 font-mono text-xs flex-grow"
            placeholder="Context will appear here."
            style="min-height: 150px;"
        ></textarea>
        <button
            @click="copy(project.shotgunPromptContext)"
            class="mt-2 px-4 py-1 bg-gray-200 text-gray-700 font-semibold rounded-md hover:bg-gray-300 self-start"
        >
          {{ copyButtonText }}
        </button>
      </div>
      <div v-else-if="project.shotgunPromptContext && project.shotgunPromptContext.startsWith('Error:')" class="text-red-500 p-3 border border-red-300 rounded bg-red-50">
        <h4 class="font-semibold mb-1">Error Generating Context:</h4>
        <pre class="text-xs whitespace-pre-wrap">{{ project.shotgunPromptContext.substring(6).trim() }}</pre>
      </div>
      <p v-else class="text-xs text-gray-500 mt-2 flex-grow flex justify-center items-center">
        Project context will be generated automatically.
      </p>
    </div>

    <!-- Initial message -->
    <p v-else class="text-xs text-gray-500 mt-2 flex-grow flex justify-center items-center">
      Select a project folder to begin.
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useProjectStore } from '../../stores/project';
import { useClipboard } from '../../composables/useClipboard';

const project = useProjectStore();
const { copy, status } = useClipboard();

const progressBarWidth = computed(() => {
  if (project.generationProgress && project.generationProgress.total > 0) {
    const percentage = (project.generationProgress.current / project.generationProgress.total) * 100;
    return `${Math.min(100, Math.max(0, percentage))}%`;
  }
  return '0%';
});

const copyButtonText = computed(() => {
  if (status.value === 'success') return 'Copied!';
  if (status.value === 'error') return 'Failed!';
  return 'Copy All';
});
</script>