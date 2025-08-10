import { defineStore } from 'pinia';
import { ref } from 'vue';
import { IsGitAvailable, GetUncommittedFiles, GetRichCommitHistory } from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notificationsStore.js';
import { useFileTreeStore } from './fileTreeStore';
import { useContextStore } from './contextStore';

export const useGitStore = defineStore('git', () => {
  const notifications = useNotificationsStore();
  const fileTree = useFileTreeStore();
  const context = useContextStore();

  const isAvailable = ref(false);
  const isLoading = ref(false);
  const richCommitHistory = ref([]);

  async function checkGitAvailability() {
    isAvailable.value = await IsGitAvailable();
  }

  async function selectUncommittedFiles(projectRoot) {
    if (!projectRoot) return;
    isLoading.value = true;
    try {
      const files = await GetUncommittedFiles(projectRoot);
      fileTree.applySelectionSet(new Set(files));
      notifications.addLog(`Выбрано ${files.length} незакоммиченных файлов.`, 'success');
      context.triggerShotgunContextGeneration();
    } catch (err) {
      notifications.addLog(`Ошибка получения незакоммиченных файлов: ${err}`, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  async function fetchRichCommitHistory(projectRoot, branchName) {
    if (!projectRoot) return;
    isLoading.value = true;
    richCommitHistory.value = [];
    try {
      const commits = await GetRichCommitHistory(projectRoot, branchName, 50);
      richCommitHistory.value = commits;
    } catch (err) {
      notifications.addLog(`Ошибка получения истории коммитов: ${err}`, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  async function selectFilesFromCommits(commitHashes) {
    const fileSet = new Set();
    const selectedCommits = richCommitHistory.value.filter(c => commitHashes.includes(c.hash));
    selectedCommits.forEach(c => {
      c.files.forEach(f => fileSet.add(f));
    });

    fileTree.applySelectionSet(fileSet);
    notifications.addLog(`Выбрано ${fileSet.size} уникальных файлов из ${commitHashes.length} коммитов.`, 'success');
    context.triggerShotgunContextGeneration();
  }

  function reset() {
    isAvailable.value = false;
    isLoading.value = false;
    richCommitHistory.value = [];
  }

  return {
    isAvailable,
    isLoading,
    richCommitHistory,
    checkGitAvailability,
    selectUncommittedFiles,
    fetchRichCommitHistory,
    selectFilesFromCommits,
    reset,
  };
});