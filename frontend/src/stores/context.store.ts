import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { FileNode, DomainFileNode } from '@/types/dto';
import { GitStatus, ContextOrigin } from '@/types/enums';
import { useUiStore } from './ui.store';
import { useNotificationsStore } from './notifications.store';
import { useProjectStore } from './project.store';
import { useSettingsStore } from './settings.store';
import { apiService } from '@/services/api.service';
import { useDebouncedSearch } from '@/composables/useDebouncedSearch';
import { useTreeStateStore } from './tree-state.store';

function mapDomainNodeToViewNode(node: DomainFileNode, depth: number, parentPath: string | null): FileNode {
  return {
    name: node.name,
    path: node.path,
    relPath: node.relPath,
    isDir: node.isDir,
    children: node.children?.map(c => ({ path: c.path })) ?? undefined,
    depth,
    gitStatus: GitStatus.Clean,
    contextOrigin: ContextOrigin.None,
    isBinary: false,
    isIgnored: node.isCustomIgnored || node.isGitignored,
    isGitignored: node.isGitignored,
    isCustomIgnored: node.isCustomIgnored,
    parentPath,
  };
}

function processNodes(nodes: DomainFileNode[]): Map<string, FileNode> {
  const map = new Map<string, FileNode>();
  const stack: { node: DomainFileNode; depth: number; parentPath: string | null }[] = nodes.map(n => ({ node: n, depth: 0, parentPath: null }));

  while (stack.length > 0) {
    const { node, depth, parentPath } = stack.pop()!;
    const viewNode = mapDomainNodeToViewNode(node, depth, parentPath);
    map.set(viewNode.path, viewNode);

    if (node.children) {
      for (const child of node.children) {
        stack.push({ node: child, depth: depth + 1, parentPath: node.path });
      }
    }
  }
  return map;
}

export const useContextStore = defineStore('context', () => {
  const nodesMap = ref(new Map<string, FileNode>());
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const shotgunContextText = ref('');

  const notifications = useNotificationsStore();
  const projectStore = useProjectStore();
  const settingsStore = useSettingsStore();
  const treeStateStore = useTreeStateStore();
  const { searchQuery, debouncedQuery } = useDebouncedSearch(250);

  const selectedFiles = computed(() => {
    const selected = [];
    for (const path of treeStateStore.selectedPaths) {
      const node = nodesMap.value.get(path);
      if (node && !node.isDir) {
        selected.push(node);
      }
    }
    return selected;
  });

  const contextSummary = computed(() => {
    const files = selectedFiles.value;
    const estimatedLines = files.length * 150;
    const estimatedTokens = Math.round(estimatedLines * 1.25);
    const cost = (estimatedTokens / 1000000) * 15;
    return {
      files: files.length,
      lines: Math.max(0, estimatedLines),
      tokens: Math.max(0, estimatedTokens),
      cost: cost,
    };
  });

  async function fetchFileTree() {
    if (!projectStore.currentProject) return;

    isLoading.value = true;
    error.value = null;
    nodesMap.value.clear();

    try {
      const { useGitignore, useCustomIgnore } = settingsStore.settings;
      const treeData = await apiService.listFiles(projectStore.currentProject.path, useGitignore, useCustomIgnore);
      nodesMap.value = processNodes(treeData);
      notifications.addLog('File tree loaded.', 'success');
      updateGitStatuses();
    } catch (err: any) {
      error.value = `Failed to load file tree: ${err.message || err}`;
      notifications.addLog(error.value, 'error');
    } finally {
      isLoading.value = false;
    }
  }

  async function updateGitStatuses() {
    if (!projectStore.currentProject) return;
    try {
      const isGit = await apiService.isGitAvailable();
      if (!isGit) return;

      const statuses = await apiService.getUncommittedFiles(projectStore.currentProject.path);
      const statusMap = new Map<string, GitStatus>();
      statuses.forEach(s => {
        const gitStatus = (s.status === 'U' || s.status === 'A') ? GitStatus.Untracked : GitStatus.Modified;
        statusMap.set(s.path, gitStatus);
      });

      nodesMap.value.forEach(node => {
        if (!node.isDir) {
          node.gitStatus = statusMap.get(node.relPath) || GitStatus.Clean;
        }
      });

    } catch (err: any) {
      notifications.addLog(`Git status check failed: ${err.message || err}`, 'error');
    }
  }

  function clearProjectData() {
    nodesMap.value.clear();
    shotgunContextText.value = '';
    treeStateStore.$reset();
  }

  async function buildContext() {
    if (!projectStore.currentProject) return;
    const paths = selectedFiles.value.map(n => n.relPath);
    if (paths.length === 0) {
      useUiStore().addToast('No files selected for context.', 'info');
      return;
    }

    try {
      await apiService.buildContext(projectStore.currentProject.path, paths);
    } catch (err: any) {
      notifications.addLog(`Context build failed: ${err.message || err}`, 'error');
    }
  }

  function setShotgunContext(context: string) {
    shotgunContextText.value = context;
  }

  function selectFilesByRelPaths(relPaths: string[], origin: ContextOrigin = ContextOrigin.Git) {
    const relPathMap = new Map<string, string>();
    nodesMap.value.forEach(node => {
      if (!node.isDir) relPathMap.set(node.relPath, node.path);
    });

    const pathsToSelect: string[] = [];
    relPaths.forEach(relPath => {
      const path = relPathMap.get(relPath);
      const node = path ? nodesMap.value.get(path) : undefined;
      if (node && !node.isIgnored) {
        pathsToSelect.push(node.path);
      }
    });

    if (pathsToSelect.length > 0) {
      treeStateStore.addSelectedPaths(pathsToSelect);
      useUiStore().addToast(`${pathsToSelect.length} files added to context from ${origin}.`, 'success');
    } else {
      useUiStore().addToast(`No new files selected. They might be ignored or already selected.`, 'info');
    }
  }

  return {
    nodesMap, isLoading, error, shotgunContextText,
    searchQuery, debouncedQuery,
    selectedFiles, contextSummary,
    fetchFileTree,
    clearProjectData,
    buildContext, setShotgunContext, selectFilesByRelPaths,
  };
});