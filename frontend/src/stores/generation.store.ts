import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { useNotificationsStore } from "./notifications.store";
import { useContextStore } from "./context.store";
import { apiService } from "@/services/api.service";
import { useUiStore } from "./ui.store";

export const useGenerationStore = defineStore("generation", () => {
  const userTask = ref("");
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const generatedDiff = ref("");

  const notifications = useNotificationsStore();
  const contextStore = useContextStore();
  const uiStore = useUiStore();

  const hasResult = computed(() => !!generatedDiff.value);
  const canGenerate = computed(
    () => !!contextStore.shotgunContextText && !!userTask.value.trim(),
  );

  async function executeGeneration() {
    if (!canGenerate.value) {
      notifications.addLog("Context is not built or task is empty.", "warn");
      return;
    }
    isLoading.value = true;
    error.value = null;
    generatedDiff.value = "";
    uiStore.setProgress({
      isActive: true,
      message: "Generating solution...",
      value: 0,
    });

    notifications.addLog("Sending request to AI...", "info");

    try {
      const systemPrompt =
        "You are an expert software developer. Your task is to implement the user's request by providing the necessary code changes in the form of a standard git diff. Do not include any explanations, comments, or apologies outside of the `git diff` block.";
      const requestPrompt = `TASK: ${userTask.value}\n\nPROJECT CONTEXT:\n${contextStore.shotgunContextText}`;

      const result = await apiService.generateCode(systemPrompt, requestPrompt);
      generatedDiff.value = result;
      notifications.addLog("AI response received successfully.", "success");
    } catch (err: any) {
      error.value = `Generation failed: ${err.message || err}`;
      notifications.addLog(error.value, "error");
    } finally {
      isLoading.value = false;
      uiStore.clearProgress();
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
