import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import { useSettingsStore } from './settingsStore';
import { useContextStore } from './contextStore';
import { usePromptTemplateStore } from './promptTemplateStore';
import { useDebouncedRef } from '../composables/useDebounce';

export const usePromptStore = defineStore('prompt', () => {
  const settings = useSettingsStore();
  const context = useContextStore();
  const promptTemplate = usePromptTemplateStore();

  const userTask = ref('');
  const debouncedUserTask = useDebouncedRef(userTask, 300);

  const finalSystemPrompt = ref('');
  const finalUserPrompt = ref('');

  const charCount = computed(() => (finalUserPrompt.value?.length || 0) + (finalSystemPrompt.value?.length || 0));
  const approximateTokens = computed(() => Math.round(charCount.value / 3.5).toLocaleString('en-US'));

  function generateFinalPrompt() {
    const templateKey = promptTemplate.selectedTemplateKey;
    const templateContent = promptTemplate.templates[templateKey]?.content || '';

    // Build the system prompt from instructions and rules
    const systemInstructions = (templateContent.match(/INSTRUCTIONS:(.*?)RULES:/s)?.[1] || '').trim();
    const yyyy = new Date().getFullYear();
    const mm = String(new Date().getMonth() + 1).padStart(2, '0');
    const dd = String(new Date().getDate()).padStart(2, '0');
    const currentDate = `${yyyy}-${mm}-${dd}`;

    let system = `${systemInstructions}\n\n${settings.customPromptRules}`;
    finalSystemPrompt.value = system.replaceAll('{CURRENT_DATE}', currentDate).trim();

    // Build the user prompt from the task and context, using the template's structure
    // but excluding the parts already used in the system prompt.
    let user = templateContent;
    user = user.replace(/INSTRUCTIONS:.*?RULES:/s, ''); // Remove the instruction block
    user = user.replace('{RULES}', ''); // Remove the rules placeholder
    user = user.replace('{TASK}', debouncedUserTask.value || "No task provided.");
    user = user.replace('{FILE_STRUCTURE}', context.shotgunPromptContext || "No context provided.");
    finalUserPrompt.value = user.trim();
  }

  function getGeneratedPrompts() {
    return {
      systemPrompt: finalSystemPrompt.value,
      userPrompt: finalUserPrompt.value,
    };
  }

  watch(
      [
        debouncedUserTask,
        () => settings.customPromptRules,
        () => context.shotgunPromptContext,
        () => promptTemplate.selectedTemplateKey,
      ],
      () => {
        if (context.shotgunPromptContext || userTask.value) {
          generateFinalPrompt();
        }
      },
      { deep: true, immediate: true }
  );

  return {
    userTask,
    charCount,
    approximateTokens,
    generateFinalPrompt,
    getGeneratedPrompts,
  };
});