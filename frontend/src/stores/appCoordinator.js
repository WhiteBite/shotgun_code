
import { watch } from 'vue';
import { useSettingsStore } from './settings';
import { useProjectStore } from './project';
import { usePromptStore } from './prompt';

/**
 * Composable для координации взаимодействия между сторами.
 * Запускается один раз при инициализации приложения.
 */
export function useAppCoordinator() {
  const settings = useSettingsStore();
  const project = useProjectStore();
  const prompt = usePromptStore();

  // Когда меняются настройки игнорирования, пересчитываем состояние дерева
  watch([() => settings.useGitignore, () => settings.useCustomIgnore], () => {
    if (project.projectRoot) {
      project.updateAllNodesExcludedState();
      project.triggerShotgunContextGeneration();
    }
  });

  // Когда меняются данные для промпта, генерируем финальный промпт
  watch(
      [
        () => prompt.userTask,
        () => settings.customPromptRules,
        () => project.shotgunPromptContext,
        () => prompt.selectedTemplateKey,
      ],
      () => {
        // Запускаем генерацию, только если мы на 2-м шаге или дальше,
        // и если есть хотя бы задача или контекст
        if (project.shotgunPromptContext || prompt.userTask) {
          prompt.generateFinalPrompt();
        }
      },
      { deep: true }
  );

  // Вызываем инициализацию настроек при старте
  settings.initializeSettings();
}