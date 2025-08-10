import { defineStore } from 'pinia';
import { ref } from 'vue';

// Import raw markdown content for prompts
import devTemplate from '../prompts/prompt_makeDiffGitFormat.md?raw';
import architectTemplate from '../prompts/prompt_makePlan.md?raw';
import findBugTemplate from '../prompts/prompt_analyzeBug.md?raw';
import projectManagerTemplate from '../prompts/prompt_projectManager.md?raw';

/**
 * Manages the available prompt templates and the currently selected one.
 */
export const usePromptTemplateStore = defineStore('promptTemplate', () => {
  const templates = {
    dev: { name: 'Dev', content: devTemplate },
    architect: { name: 'Architect', content: architectTemplate },
    findBug: { name: 'Find Bug', content: findBugTemplate },
    projectManager: { name: 'Project: Update Tasks', content: projectManagerTemplate },
  };

  const selectedTemplateKey = ref('dev');

  return {
    templates,
    selectedTemplateKey,
  };
});