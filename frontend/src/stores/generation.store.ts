import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { useNotificationsStore } from "./notifications.store";
import { useContextBuilderStore } from "./context-builder.store";
import { apiService } from "@/services/api.service";
import { useUiStore } from "./ui.store";
import { useFileTreeStore } from "./file-tree.store";
import { useTreeStateStore } from "./tree-state.store";

export const useGenerationStore = defineStore("generation", () => {
  const userTask = ref("");
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const generatedDiff = ref("");
  const notifications = useNotificationsStore();
  const contextBuilderStore = useContextBuilderStore();
  const uiStore = useUiStore();

  const hasResult = computed(() => !!generatedDiff.value);
  const canGenerate = computed(() => !!contextBuilderStore.shotgunContextText && !!userTask.value.trim());

  async function executeGeneration() {
    if (!canGenerate.value) return;
    isLoading.value = true;
    error.value = null;
    generatedDiff.value = "";
    uiStore.setProgress({ isActive: true, message: "Generating solution...", value: 0 });
    try {
      const systemPrompt = "You are an expert software developer. Your task is to implement the user's request by providing the necessary code changes in the form of a standard git diff. Do not include any explanations, comments, or apologies outside of the `git diff` block.";
      const requestPrompt = `TASK: ${userTask.value}\n\nPROJECT CONTEXT:\n${contextBuilderStore.shotgunContextText}`;

      generatedDiff.value = await apiService.generateCode(systemPrompt, requestPrompt);
      notifications.addLog("AI response received successfully.", "success");
    } catch (err: any) {
      error.value = `Generation failed: ${err.message || err}`;
      notifications.addLog(error.value, "error");
    } finally {
      isLoading.value = false;
      uiStore.clearProgress();
    }
  }

  async function suggestContext() {
    const task = userTask.value.trim();
    if (!task) {
      uiStore.addToast("Опишите задачу, чтобы получить предложение.", "info");
      return;
    }
    const fileTreeStore = useFileTreeStore();
    const treeStateStore = useTreeStateStore();
    isLoading.value = true;
    try {
      const allNodes = Array.from(fileTreeStore.nodesMap.values());
      const suggestedFiles = await apiService.suggestContextFiles(task, allNodes as any);
      treeStateStore.selectFilesByRelPaths(suggestedFiles, fileTreeStore.nodesMap);
      notifications.addLog(`Предложено ${suggestedFiles.length} файлов. Добавлены в контекст.`, "success");
    } catch (err: any) {
      const msg = `Ошибка при предложении файлов: ${err?.message || err}`;
      notifications.addLog(msg, "error");
      uiStore.addToast(msg, "error");
    } finally {
      isLoading.value = false;
    }
  }

  function clearResult() {
    generatedDiff.value = "";
    error.value = null;
  }

  return { userTask, isLoading, error, generatedDiff, hasResult, canGenerate, executeGeneration, suggestContext, clearResult };
});