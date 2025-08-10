import { defineStore } from 'pinia';
import { ref } from 'vue';
import { RequestShotgunContextGeneration } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { useStepsStore } from './stepsStore.js';
import { useFileTreeStore } from './fileTreeStore';
import { useProjectStore } from './projectStore';

export const useContextStore = defineStore('context', () => {
  const steps = useStepsStore();
  const fileTree = useFileTreeStore();
  const project = useProjectStore();

  const isGeneratingContext = ref(false);
  const generationProgress = ref({ current: 0, total: 0 });
  const shotgunPromptContext = ref('');
  const hasGeneratedContextOnce = ref(false);

  function triggerShotgunContextGeneration() {
    if (!project.projectRoot || fileTree.isFileTreeLoading) return;

    isGeneratingContext.value = true;
    generationProgress.value = { current: 0, total: 0 };
    shotgunPromptContext.value = '';

    const includedPaths = fileTree.collectIncludedPaths(fileTree.tree);
    RequestShotgunContextGeneration(project.projectRoot, includedPaths);
  }

  function reset() {
    isGeneratingContext.value = false;
    generationProgress.value = { current: 0, total: 0 };
    shotgunPromptContext.value = '';
    hasGeneratedContextOnce.value = false;
  }

  function setupWailsListeners() {
    EventsOn("shotgunContextGenerated", (output) => {
      shotgunPromptContext.value = output;
      isGeneratingContext.value = false;
      hasGeneratedContextOnce.value = true;
      steps.completeStep(1);
    });

    EventsOn("app:error", (errorMsg) => {
      if (isGeneratingContext.value) {
        shotgunPromptContext.value = "Error: " + errorMsg;
        isGeneratingContext.value = false;
      }
    });

    EventsOn("shotgunContextGenerationProgress", (progress) => {
      generationProgress.value = progress;
    });
  }

  setupWailsListeners();

  return {
    isGeneratingContext,
    generationProgress,
    shotgunPromptContext,
    hasGeneratedContextOnce,
    triggerShotgunContextGeneration,
    reset,
  };
});