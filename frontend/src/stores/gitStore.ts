
import { defineStore } from 'pinia';
import { ref } from 'vue';
import { IsGitAvailable, GetUncommittedFiles, GetRichCommitHistory } from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notificationsStore';
import { useContextStore } from './contextStore';
import type { CommitWithFiles, FileStatus } from '@/types/dto';
import { useProjectStore } from './projectStore';
import { GitStatus } from '@/types/enums';

export const useGitStore = defineStore('git', () => {
  const notifications = useNotificationsStore();
  const projectStore = useProjectStore();
  const contextStore = useContextStore();

  const isAvailable = ref(false);
  const isLoading = ref(false);
  const commitHistory = ref<CommitWithFiles[]>([]);

  async function checkAvailabilityAndStatus() {
    if (!projectStore.currentProject) return;
    try {
      isAvailable.value = await IsGitAvailable();
      if (!isAvailable.value) return;

      isLoading.value = true;
      const statuses: FileStatus[] = await GetUncommittedFiles(projectStore.currentProject.path);

      const statusMap = new Map<string, GitStatus>();
      statuses.forEach(s => {
        const gitStatus = (s.status === '??' || s.status === 'A') ? GitStatus.Untracked : GitStatus.Modified;
        statusMap.set(s.path, gitStatus);
      });

      contextStore.nodesMap.forEach(node => {
        if (statusMap.has(node.relPath)) {
          node.gitStatus = statusMap.get(node.relPath)!;
        } else {
          node.gitStatus = GitStatus.Clean;
        }
      });

    } catch (err: any) {
      isAvailable.value = false;
      notifications.addLog(`Ошибка Git: ${err.message || err}`, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  async function fetchCommitHistory(branchName = '', limit = 50) {
    if (!projectStore.currentProject) return;
    isLoading.value = true;
    try {
      commitHistory.value = await GetRichCommitHistory(projectStore.currentProject.path, branchName, limit);
    } catch (err: any) {
      notifications.addLog(`Ошибка истории коммитов: ${err}`, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  function selectFilesFromCommits(_commitHashes: string[]) {
    notifications.addLog('Выбор файлов из коммитов еще не реализован.', 'info');
  }

  return {
    isAvailable, isLoading, commitHistory,
    checkAvailabilityAndStatus,
    fetchCommitHistory,
    selectFilesFromCommits,
  };
});