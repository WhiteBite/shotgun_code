<template>
  <div class="h-16 bg-gray-800/95 backdrop-blur-sm border-t border-gray-700 flex-shrink-0 flex items-center justify-between px-6 text-sm text-gray-300 shadow-lg">
    <div class="flex items-center gap-6 min-w-0">
      <button 
        @click="uiStore.toggleConsole()" 
        class="btn btn-ghost btn-icon" 
        title="Toggle Console" 
        aria-label="Toggle Console"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="4 17 10 11 4 5"></polyline>
          <line x1="12" y1="19" x2="20" y2="19"></line>
        </svg>
      </button>
      
      <!-- Mode Toggle -->
      <div class="flex items-center bg-gray-700/80 rounded-lg p-1 shadow-lg border border-gray-600/50">
        <button
          @click="switchToManual"
          :class="[
            'btn btn-sm transition-all duration-200',
            uiStore.workspaceMode === 'manual' 
              ? 'btn-primary transform scale-105' 
              : 'btn-ghost hover:bg-gray-600/80'
          ]"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
          Manual
        </button>
        <button
          @click="switchToAutonomous"
          :class="[
            'btn btn-sm transition-all duration-200',
            uiStore.workspaceMode === 'autonomous' 
              ? 'btn-secondary transform scale-105' 
              : 'btn-ghost hover:bg-gray-600/80'
          ]"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          Autonomous
        </button>
      </div>
      
      <div v-if="contextBuilderStore.isBuilding" class="flex items-center gap-2 text-sm text-blue-400 font-medium">
        <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        Построение контекста...
      </div>
    </div>
    
    <div class="flex items-center gap-4">
      <!-- Build Context Button -->
      <button
        @click="buildContext"
        :disabled="!contextBuilderStore.canBuildContext || contextBuilderStore.isBuilding"
        class="btn btn-secondary btn-md"
        :title="buildDisabledReason"
      >
        <svg v-if="contextBuilderStore.isBuilding" class="animate-spin w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        {{ contextBuilderStore.isBuilding ? t('generating') : t('context.build') }}
      </button>
       
      <!-- Generate Button -->
      <button
        @click="generationStore.executeGeneration()"
        :disabled="!generationStore.canGenerate || contextBuilderStore.isBuilding"
        class="btn btn-primary btn-lg"
        :title="genDisabledReason"
      >
        <svg v-if="generationStore.isLoading" class="animate-spin w-4 h-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a 8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        {{ generationStore.isLoading ? t('generating') : t('generate') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUiStore } from "@/stores/ui.store";
import { useGenerationStore } from "@/stores/generation.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useProjectStore } from "@/stores/project.store";
import { t } from "@/lib/i18n";

const uiStore = useUiStore();
const generationStore = useGenerationStore();
const contextBuilderStore = useContextBuilderStore();
const projectStore = useProjectStore();

const buildContext = () => {
  if (projectStore.currentProject) {
    contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path);
  }
};

const switchToManual = () => {
  uiStore.setWorkspaceMode('manual');
};

const switchToAutonomous = () => {
  uiStore.setWorkspaceMode('autonomous');
};

const buildDisabledReason = computed(() => {
  if (contextBuilderStore.isBuilding) return 'Идёт сборка контекста';
  if (!contextBuilderStore.canBuildContext) return 'Нужен выбор файлов в дереве';
  return 'Собрать контекст';
});

const genDisabledReason = computed(() => {
  if (generationStore.isLoading) return 'Идёт генерация';
  if (!generationStore.canGenerate) return 'Нужны задача и готовый контекст';
  return 'Сгенерировать';
});
</script>