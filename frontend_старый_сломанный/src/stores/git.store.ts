import { defineStore } from "pinia";
import { ref } from "vue";
import { useAdvancedErrorHandler } from "@/composables/useErrorHandler";
import { useProjectStore } from "./project.store";
import { useTreeStateStore } from "./tree-state.store";
import { useFileTreeStore } from "./file-tree.store";
import type { CommitWithFiles } from "@/types/dto";
// Import container and repository
import { container } from "@/infrastructure/container";
import type { GitRepository } from "@/domain/repositories/GitRepository";

export const useGitStore = defineStore("git", () => {
  const { handleStructuredError } = useAdvancedErrorHandler();
  const projectStore = useProjectStore();

  // Inject GitRepository
  const gitRepository: GitRepository = container.gitRepository;

  const isHistoryVisible = ref(false);
  const isLoading = ref(false);
  const commits = ref<CommitWithFiles[]>([]);

  async function fetchHistory(limit = 50) {
    if (!projectStore.currentProject) return;
    isLoading.value = true;
    commits.value = [];
    try {
      // Use GitRepository instead of apiService
      const history = await gitRepository.getRichCommitHistory(
        projectStore.currentProject.path,
        "",
        limit,
      );
      commits.value = history;
    } catch (err) {
      handleStructuredError(err, { operation: "Fetch Git History", component: "GitStore" });
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

  function applyCommitSelection(selectedHashes: string[]) {
    if (!selectedHashes || selectedHashes.length === 0) return;
    const treeStateStore = useTreeStateStore();
    const fileTreeStore = useFileTreeStore();
    const filesToSelect = new Set<string>();
    const selectedHashesSet = new Set(selectedHashes);

    for (const commit of commits.value) {
      if (selectedHashesSet.has(commit.hash)) {
        commit.files.forEach((fileRelPath) => filesToSelect.add(fileRelPath));
      }
    }

    treeStateStore.selectFilesByRelPaths(
      Array.from(filesToSelect),
      fileTreeStore.nodesRelMap,
    );
  }

  return {
    isHistoryVisible,
    isLoading,
    commits,
    showHistory,
    hideHistory,
    applyCommitSelection,
  };
});