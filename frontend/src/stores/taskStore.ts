import { defineStore } from "pinia";
import { ref } from "vue";
import { SuggestContextFiles } from "../../wailsjs/go/main/App";
import { useContextStore } from "./context.store";
import { useUiStore } from "./ui.store";
import { ContextOrigin } from "@/types/enums";

export const useTaskStore = defineStore("task", () => {
  const userTask = ref("");
  const isSuggesting = ref(false);
  const contextStore = useContextStore();
  const uiStore = useUiStore();

  async function suggestContext() {
    if (!userTask.value.trim()) {
      uiStore.addToast("Опишите задачу, чтобы получить предложение.", "info");
      return;
    }
    isSuggesting.value = true;
    try {
      const allNodes = Array.from(contextStore.nodesMap.values()).map(
        ({ children: _children, ...rest }) => rest,
      );
      const suggestedFiles = await SuggestContextFiles(
        userTask.value,
        allNodes as any,
      );
      contextStore.selectFilesByRelPaths(suggestedFiles, ContextOrigin.AI);
      uiStore.addToast(
        `Предложено ${suggestedFiles.length} файлов. Добавлены в контекст.`,
        "success",
      );
    } catch (err: any) {
      uiStore.addToast(`Ошибка при предложении файлов: ${err}`, "error");
    } finally {
      isSuggesting.value = false;
    }
  }

  return {
    userTask,
    isSuggesting,
    suggestContext,
  };
});
