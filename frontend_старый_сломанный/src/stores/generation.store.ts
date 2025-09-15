import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { useContextBuilderStore } from "./context-builder.store";
import { useUiStore } from "./ui.store";
// Import container and repository
import { container } from "@/infrastructure/container";
import type { AIRepository } from "@/domain/repositories/AIRepository";

export const useGenerationStore = defineStore("generation", () => {
  const userTask = ref("");
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const generatedDiff = ref("");
  const contextBuilderStore = useContextBuilderStore();
  const uiStore = useUiStore();

  // Inject AIRepository
  const aiRepository: AIRepository = container.aiRepository;

  const hasResult = computed(() => !!generatedDiff.value);
  const canGenerate = computed(() => {
    // Контекст теперь хранится постранично, используем наличие contextId или метрик
    const hasContext = !!contextBuilderStore.currentContextId || (contextBuilderStore.contextMetrics?.tokenCount || 0) > 0;
    return hasContext && !!userTask.value.trim();
  });

  async function executeGeneration() {
    if (!canGenerate.value) return;
    isLoading.value = true;
    error.value = null;
    generatedDiff.value = "";

    try {
      const systemPrompt =
        "You are an expert software developer. Your task is to implement the user's request by providing the necessary code changes in the form of a standard git diff. Do not include any explanations, comments, or apologies outside of the `git diff` block.";
      const requestPrompt = `TASK: ${userTask.value}\n\nPROJECT CONTEXT SUMMARY:\nTokens: ${contextBuilderStore.contextMetrics.tokenCount}\nFiles: ${contextBuilderStore.contextMetrics.fileCount}`;

      // Use AIRepository instead of apiService
      generatedDiff.value = await aiRepository.generateCode(
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
