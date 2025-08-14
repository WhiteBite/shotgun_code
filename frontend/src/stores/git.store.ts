import { defineStore } from "pinia";
import { ref } from "vue";
import { useErrorHandler } from "@/composables/useErrorHandler";
import { useProjectStore } from "./project.store";
import { apiService } from "@/services/api.service";
import type { CommitWithFiles } from "@/types/dto";

export const useGitStore = defineStore("git", () => {
  const { handleError } = useErrorHandler();
  const projectStore = useProjectStore();

  const isHistoryVisible = ref(false);
  const isLoading = ref(false);
  const commits = ref<CommitWithFiles[]>([]);

  async function fetchHistory(limit = 50) {
    if (!projectStore.currentProject) return;
    isLoading.value = true;
    commits.value = [];
    try {
      const history = await apiService.getRichCommitHistory(
        projectStore.currentProject.path,
        "",
        limit,
      );
      commits.value = history;
    } catch (err) {
      handleError(err, "Fetch Git History");
    } finally {
      isLoading.value = false;
    }
  }

  function showHistory() {
    isHistoryVisible.value = true;
    fetchHistory();
  }

  function hideHistory() {
    isHistoryVisible.value = false;
  }

  return {
    isHistoryVisible,
    isLoading,
    commits,
    showHistory,
    hideHistory,
  };
});
