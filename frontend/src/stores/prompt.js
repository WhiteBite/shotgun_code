
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { useNotificationsStore } from './notifications';
import { useProjectStore } from './project';
import { useSettingsStore } from './settings';
import { useStepsStore } from './steps';

import devTemplate from '../prompts/prompt_makeDiffGitFormat.md?raw';
import architectTemplate from '../prompts/prompt_makePlan.md?raw';
import findBugTemplate from '../prompts/prompt_analyzeBug.md?raw';
import projectManagerTemplate from '../prompts/prompt_projectManager.md?raw';

export const usePromptStore = defineStore('prompt', () => {
  // ИСПРАВЛЕНО: Все use...Store() вызовы перенесены внутрь
  const notifications = useNotificationsStore();
  const project = useProjectStore();
  const settings = useSettingsStore();
  const steps = useStepsStore();

  const userTask = ref('');
  const finalPrompt = ref('');
  const isLoading = ref(false);
  const selectedTemplateKey = ref('dev');

  let debounceTimer = null;

  const templates = {
    dev: { name: 'Dev', content: devTemplate },
    architect: { name: 'Architect', content: architectTemplate },
    findBug: { name: 'Find Bug', content: findBugTemplate },
    projectManager: { name: 'Project: Update Tasks', content: projectManagerTemplate },
  };

  const charCount = computed(() => (finalPrompt.value || '').length);
  const approximateTokens = computed(() => Math.round(charCount.value / 3).toLocaleString('en-US'));

  function generateFinalPrompt() {
    clearTimeout(debounceTimer);
    isLoading.value = true;
    debounceTimer = setTimeout(() => {
      const template = templates[selectedTemplateKey.value].content;
      let populated = template;
      populated = populated.replace('{TASK}', userTask.value || "No task provided.");
      populated = populated.replace('{RULES}', settings.customPromptRules);
      populated = populated.replace('{FILE_STRUCTURE}', project.shotgunPromptContext || "No context provided.");

      const yyyy = new Date().getFullYear();
      const mm = String(new Date().getMonth() + 1).padStart(2, '0');
      const dd = String(new Date().getDate()).padStart(2, '0');
      populated = populated.replaceAll('{CURRENT_DATE}', `${yyyy}-${mm}-${dd}`);

      finalPrompt.value = populated;
      isLoading.value = false;
      notifications.addLog('Промпт для LLM сгенерирован.', 'info');
      steps.completeStep(2);
    }, 750);
  }

  return {
    userTask,
    finalPrompt,
    isLoading,
    templates,
    selectedTemplateKey,
    charCount,
    approximateTokens,
    generateFinalPrompt,
  };
});