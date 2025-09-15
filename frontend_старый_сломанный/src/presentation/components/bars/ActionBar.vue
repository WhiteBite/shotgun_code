<template>
  <div class="h-16 bg-gray-800/95 backdrop-blur-sm border-t border-gray-700 flex-shrink-0 flex items-center justify-between px-6 text-sm text-gray-300 shadow-lg">
    <div class="flex items-center gap-6 min-w-0">
      <button 
        class="btn btn-ghost btn-icon" 
        title="Toggle Console" 
        aria-label="Toggle Console" 
        @click="uiStore.toggleConsole()"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="4 17 10 11 4 5"></polyline>
          <line x1="12" y1="19" x2="20" y2="19"></line>
        </svg>
      </button>
      
      <!-- Mode Toggle -->
      <div class="flex items-center bg-gray-700/80 rounded-lg p-1 shadow-lg border border-gray-600/50">
        <button
          :class="[
            'btn btn-sm transition-all duration-200',
            workspaceStore.currentMode === 'manual' 
              ? 'btn-primary transform scale-105' 
              : 'btn-ghost hover:bg-gray-600/80'
          ]"
          @click="switchToManual"
        >
          <FileTextIcon class="w-4 h-4" />
          Manual
        </button>
        <button
          :class="[
            'btn btn-sm transition-all duration-200',
            workspaceStore.currentMode === 'autonomous' 
              ? 'btn-secondary transform scale-105' 
              : 'btn-ghost hover:bg-gray-600/80'
          ]"
          @click="switchToAutonomous"
        >
          <CpuIcon class="w-4 h-4" />
          Autonomous
        </button>
      </div>
      
      <div v-if="contextBuilderStore.isBuilding" class="flex items-center gap-2 text-sm text-blue-400 font-medium">
        <LoaderIcon class="animate-spin w-4 h-4" />
        Построение контекста...
      </div>
    </div>
    
    <div class="flex items-center gap-4">
      <!-- Build Context Button -->
      <button
        :disabled="!contextBuilderStore.canBuildContext || contextBuilderStore.isBuilding"
        class="btn btn-secondary btn-md"
        :title="buildDisabledReason"
        @click="buildContext"
      >
        <LoaderIcon v-if="contextBuilderStore.isBuilding" class="animate-spin w-4 h-4" />
        <ZapIcon v-else class="w-4 h-4" />
        {{ contextBuilderStore.isBuilding ? 'Generating' : 'Build Context' }}
      </button>
       
      <!-- Generate Button -->
      <button
        :disabled="!generationStore.canGenerate || contextBuilderStore.isBuilding"
        class="btn btn-primary btn-lg"
        :title="genDisabledReason"
        @click="generationStore.executeGeneration()"
      >
        <LoaderIcon v-if="generationStore.isLoading" class="animate-spin w-4 h-4" />
        <PlayIcon v-else class="w-4 h-4" />
        {{ generationStore.isLoading ? 'Generating' : 'Generate' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { 
  PlayIcon, 
  ZapIcon, 
  CpuIcon,
  FileTextIcon,
  LoaderIcon
} from 'lucide-vue-next';
import { useUiStore } from '@/stores/ui.store';
import { useWorkspaceStore } from '@/stores/workspace.store';
import { useContextBuilderStore } from '@/stores/context-builder.store';
import { useGenerationStore } from '@/stores/generation.store';
import { useProjectStore } from '@/stores/project.store';
import { UIFormattingService } from '@/domain/services/UIFormattingService';
import { APP_CONFIG } from '@/config/app-config';

const uiStore = useUiStore();
const workspaceStore = useWorkspaceStore();
const contextBuilderStore = useContextBuilderStore();
const generationStore = useGenerationStore();
const projectStore = useProjectStore();
const uiFormattingService = new UIFormattingService();

const buildContext = () => {
  if (projectStore.currentProject) {
    contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path);
  }
};

const switchToManual = () => {
  workspaceStore.setMode('manual');
};

const switchToAutonomous = () => {
  workspaceStore.setMode('autonomous');
};

const isManualMode = computed(() => workspaceStore.currentMode === 'manual');
const isAutonomousMode = computed(() => workspaceStore.currentMode === 'autonomous');

const buttonStates = computed(() => ({
  manual: {
    active: isManualMode.value,
    icon: FileTextIcon,
    label: 'Manual',
    title: 'Ручной режим',
    action: switchToManual
  },
  autonomous: {
    active: isAutonomousMode.value,
    icon: CpuIcon,
    label: 'Autonomous',
    title: 'Автономный режим',
    action: switchToAutonomous
  }
}));

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