import { watch } from 'vue';
import { useSettingsStore } from './settingsStore';
import { useProjectStore } from './projectStore';
import { usePromptStore } from './promptStore';
import { useFileTreeStore } from './fileTreeStore';
import { useContextStore } from './contextStore';
import { usePromptTemplateStore } from './promptTemplateStore';

let contextDebounceTimer = null;

export function useAppCoordinator() {
  const settings = useSettingsStore();
  const project = useProjectStore();
  const prompt = usePromptStore();
  const fileTree = useFileTreeStore();
  const context = useContextStore();
  const promptTemplate = usePromptTemplateStore();

  const handleContextRelevantChange = () => {
    if (context.hasGeneratedContextOnce) {
      clearTimeout(contextDebounceTimer);
      contextDebounceTimer = setTimeout(() => {
        if (project.projectRoot && !fileTree.isFileTreeLoading) {
          context.triggerShotgunContextGeneration();
        }
      }, 800);
    }
  };

  watch([() => settings.useGitignore, () => settings.useCustomIgnore], async () => {
    if (project.projectRoot) {
      await fileTree.loadFileTree(project.projectRoot);
    }
  }, { deep: true });

  watch(
      () => fileTree.tree,
      () => { handleContextRelevantChange() },
      { deep: true }
  );

  watch(
      [
        () => prompt.userTask,
        () => settings.customPromptRules,
        () => context.shotgunPromptContext,
        () => promptTemplate.selectedTemplateKey,
      ],
      () => {
        if (context.shotgunPromptContext || prompt.userTask) {
          prompt.generateFinalPrompt();
        }
      },
      { deep: true, immediate: true }
  );

  settings.initializeSettings();
}