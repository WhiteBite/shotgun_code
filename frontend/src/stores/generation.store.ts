import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { useContextBuilderStore } from "./context-builder.store";
import { apiService } from "@/infrastructure/api/api.service";
import { useUiStore } from "./ui.store";

export const useGenerationStore = defineStore("generation", () => {
  const userTask = ref("");
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const generatedDiff = ref("");
  const contextBuilderStore = useContextBuilderStore();
  const uiStore = useUiStore();

  const hasResult = computed(() => !!generatedDiff.value);
  const canGenerate = computed(
    () =>
      !!contextBuilderStore.currentContext?.content && !!userTask.value.trim(),
  );

  async function executeGeneration() {
    if (!canGenerate.value) return;
    isLoading.value = true;
    error.value = null;
    generatedDiff.value = "";

    try {
      const systemPrompt =
        "You are an expert software developer. Your task is to implement the user's request by providing the necessary code changes in the form of a standard git diff. Do not include any explanations, comments, or apologies outside of the `git diff` block.";
      const requestPrompt = `TASK: ${userTask.value}\n\nPROJECT CONTEXT:\n${contextBuilderStore.currentContext?.content}`;

      generatedDiff.value = await apiService.generateCode(
        systemPrompt,
        requestPrompt,
      );
      uiStore.addToast("AI ответ получен успешно.", "success");
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : "An unknown error occurred";
      error.value = errorMessage;
      uiStore.addToast(errorMessage, "error");
    } finally {
      isLoading.value = false;
    }
  }

  function clearResult() {
    generatedDiff.value = "";
    error.value = null;
  }

  return {
    userTask,
    isLoading,
    error,
    generatedDiff,
    hasResult,
    canGenerate,
    executeGeneration,
    clearResult,
  };
});