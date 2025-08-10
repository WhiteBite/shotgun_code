<template>
  <div class="p-4 h-full flex flex-col">
    <div v-if="context.isGeneratingContext" class="flex-grow flex justify-center items-center">
      <!-- Context Generation Progress -->
      <div class="text-center">
        <div class="w-64 mx-auto">
          <p class="text-gray-600 mb-1 text-sm">Generating project context...</p>
          <div class="w-full bg-gray-200 rounded-full h-2.5">
            <div class="bg-blue-600 h-2.5 rounded-full" :style="{ width: progressBarWidth }"></div>
          </div>
          <p class="text-gray-500 mt-1 text-xs">
            {{ context.generationProgress.current }} / {{ context.generationProgress.total > 0 ? context.generationProgress.total : '...' }} bytes
          </p>
        </div>
      </div>
    </div>

    <div v-else-if="project.projectRoot" class="flex-grow flex flex-col">
      <div class="flex flex-col flex-grow">
        <!-- Task Configuration Area -->
        <div class="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label for="agent-select" class="block text-sm font-medium text-gray-700">Agent Mode</label>
            <select id="agent-select" v-model="promptTemplate.selectedTemplateKey" class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md">
              <option v-for="(template, key) in promptTemplate.templates" :key="key" :value="key">
                {{ template.name }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">Context Size</label>
            <div class="mt-1 p-2 border border-gray-300 rounded-md bg-gray-50 text-sm">
              ~{{ prompt.approximateTokens }} tokens
            </div>
          </div>
        </div>

        <!-- Task Input Area -->
        <div class="mb-4">
          <label for="task-input" class="block text-md font-medium text-gray-700 mb-2">Your Task:</label>
          <textarea
              id="task-input"
              v-model="prompt.userTask"
              rows="3"
              class="w-full p-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
              placeholder="e.g., 'Refactor the authentication service to use a JWT strategy' or 'Fix the bug where users cannot update their profile picture.'"
          ></textarea>
        </div>

        <!-- Context Display Area -->
        <div class="flex-grow flex flex-col">
          <h3 class="text-md font-medium text-gray-700 mb-2">Project Context:</h3>
          <div v-if="context.hasGeneratedContextOnce" class="flex-grow flex flex-col">
            <textarea
                :value="context.shotgunPromptContext"
                rows="10"
                readonly
                class="w-full p-2 border border-gray-300 rounded-md shadow-sm bg-gray-50 font-mono text-xs flex-grow"
            ></textarea>
          </div>
          <div v-else class="text-gray-500 p-3 border border-dashed rounded bg-gray-50 flex-grow flex flex-col justify-center items-center text-center">
            <p>Project context will appear here after generation.</p>
            <p class="text-xs mt-1">Adjust file selection in the left sidebar to refine the context.</p>
          </div>
        </div>
      </div>

      <!-- Action Button Area -->
      <div class="mt-4 flex flex-col items-center">
        <button
            @click="handleMainAction"
            :disabled="isMainButtonDisabled"
            class="px-6 py-3 w-full max-w-sm bg-blue-600 text-white font-semibold rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-wait flex items-center justify-center"
        >
          <svg v-if="aiExecution.isLoading" class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span>{{ mainButtonText }}</span>
        </button>
        <p v-if="aiExecution.error" class="text-red-500 text-xs mt-2 text-center">{{ aiExecution.error }}</p>
      </div>
    </div>

    <p v-else class="text-xs text-gray-500 mt-2 flex-grow flex justify-center items-center">
      Select a project folder to begin.
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useProjectStore } from '../../stores/projectStore';
import { useContextStore } from '../../stores/contextStore';
import { usePromptStore } from '../../stores/promptStore';
import { useAiExecutionStore } from '../../stores/aiExecutionStore';
import { usePromptTemplateStore } from '../../stores/promptTemplateStore';

const project = useProjectStore();
const context = useContextStore();
const prompt = usePromptStore();
const aiExecution = useAiExecutionStore();
const promptTemplate = usePromptTemplateStore();

const progressBarWidth = computed(() => {
  if (context.generationProgress?.total > 0) {
    const percentage = (context.generationProgress.current / context.generationProgress.total) * 100;
    return `${Math.min(100, Math.max(0, percentage))}%`;
  }
  return '0%';
});

const handleMainAction = () => {
  if (context.hasGeneratedContextOnce) {
    // Если контекст уже есть, генерируем решение
    aiExecution.executeAIGeneration();
  } else {
    // Если контекста еще нет, генерируем его
    context.triggerShotgunContextGeneration();
  }
};

const mainButtonText = computed(() => {
  if (aiExecution.isLoading) {
    return 'Generating Solution...';
  }
  if (!context.hasGeneratedContextOnce) {
    return 'Generate Project Context';
  }
  return 'Generate Solution';
});

const isMainButtonDisabled = computed(() => {
  if (aiExecution.isLoading || context.isGeneratingContext) {
    return true;
  }
  // Блокируем, только если задача пуста ПОСЛЕ того, как контекст был сгенерирован
  if (context.hasGeneratedContextOnce && !prompt.userTask.trim()) {
    return true;
  }
  return false;
});
</script>