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
    expanded: depth < 1,
    selected: 'off',
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
  const activeNodePath = ref<string | null>(null);
  const shotgunContextText = ref('');

  const notifications = useNotificationsStore();
  const projectStore = useProjectStore();
  const settingsStore = useSettingsStore();
  const { searchQuery, debouncedQuery } = useDebouncedSearch(250);

  const activeNode = computed(() => activeNodePath.value ? nodesMap.value.get(activeNodePath.value) : null);

  const visibleNodes = computed(() => {
    const result: FileNode[] = [];
    if (nodesMap.value.size === 0) return result;

    const roots = Array.from(nodesMap.value.values()).filter(n => n.depth === 0);
    const query = debouncedQuery.value.toLowerCase().trim();
    const isFiltering = !!query;

    function traverse(nodes: FileNode[]) {
      for (const node of nodes) {
        const children = node.children?.map(c => nodesMap.value.get(c.path)).filter(Boolean) as FileNode[] || [];

        if (isFiltering) {
          if (node.name.toLowerCase().includes(query)) {
            result.push(node);
          }
          traverse(children);
        } else {
          result.push(node);
          if (node.expanded) {
            traverse(children);
          }
        }
      }
    }

    traverse(roots);

    if(isFiltering) {
      return result.sort((a,b) => a.path.localeCompare(b.path));
    }
    return result;
  });

  const selectedFiles = computed(() => Array.from(nodesMap.value.values()).filter(n => !n.isDir && n.selected === 'on'));

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

  async function fetchFileTree(preserveState = false) {
    if (!projectStore.currentProject) return;

    const oldState = {
      expandedPaths: new Set<string>(),
      selectedPaths: new Set<string>(),
    };

    if (preserveState) {
      nodesMap.value.forEach(node => {
        if (node.expanded) oldState.expandedPaths.add(node.path);
        if (node.selected === 'on') oldState.selectedPaths.add(node.path);
      });
    }

    isLoading.value = true;
    error.value = null;
    if (!preserveState) {
      nodesMap.value.clear();
    }

    try {
      const { useGitignore, useCustomIgnore } = settingsStore.settings;
      const treeData = await apiService.listFiles(projectStore.currentProject.path, useGitignore, useCustomIgnore);
      const newNodesMap = processNodes(treeData);

      if (preserveState) {
        newNodesMap.forEach(newNode => {
          if (oldState.expandedPaths.has(newNode.path)) {
            newNode.expanded = true;
          }
          if (!newNode.isIgnored && oldState.selectedPaths.has(newNode.path)) {
            newNode.selected = 'on';
            newNode.contextOrigin = ContextOrigin.Manual;
          }
        });

        newNodesMap.forEach(node => {
          if (node.selected === 'on' && node.parentPath) {
            _updateParentSelection(node.parentPath, newNodesMap);
          }
        });
      }

      nodesMap.value = newNodesMap;

      if (!preserveState) {
        notifications.addLog('File tree loaded.', 'success');
        updateGitStatuses();
      } else {
        notifications.addLog('Ignore rules updated.', 'info');
      }

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
    activeNodePath.value = null;
    error.value = null;
    shotgunContextText.value = '';
  }

  function toggleNodeExpansion(path: string, recursive = false) {
    const node = nodesMap.value.get(path);
    if (!node?.isDir) return;

    const newState = !node.expanded;
    const stack: FileNode[] = [node];

    while(stack.length > 0) {
      const current = stack.pop()!;
      current.expanded = newState;
      if (recursive && current.children) {
        current.children.forEach(childRef => {
          const childNode = nodesMap.value.get(childRef.path);
          if (childNode?.isDir) {
            stack.push(childNode);
          }
        });
      }
    }
  }

  function _updateParentSelection(path: string | null, map: Map<string, FileNode>) {
    if (!path) return;
    const parent = map.get(path);
    if (!parent || !parent.children) return;

    const children = parent.children.map(c => map.get(c.path)!).filter(Boolean);
    if (children.length === 0) return;

    const selectedCount = children.filter(c => c.selected === 'on').length;
    const partialCount = children.filter(c => c.selected === 'partial').length;

    let newStatus: 'on' | 'off' | 'partial' = 'off';
    if (selectedCount === children.length) newStatus = 'on';
    else if (selectedCount > 0 || partialCount > 0) newStatus = 'partial';

    if(parent.selected !== newStatus) {
      parent.selected = newStatus;
      _updateParentSelection(parent.parentPath, map);
    }
  }

  function toggleNodeSelection(path: string) {
    const node = nodesMap.value.get(path);
    if (!node || node.isIgnored) return;

    const newSelection = node.selected === 'on' ? 'off' : 'on';

    const stack: FileNode[] = [node];
    while(stack.length > 0) {
      const current = stack.pop()!;
      if (!current.isIgnored) {
        current.selected = newSelection;
        if (newSelection === 'on') {
          current.contextOrigin = ContextOrigin.Manual;
        } else {
          current.contextOrigin = ContextOrigin.None;
        }
        if(current.children) {
          current.children.forEach(c => {
            const childNode = nodesMap.value.get(c.path);
            if (childNode) stack.push(childNode);
          });
        }
      }
    }
    _updateParentSelection(node.parentPath, nodesMap.value);
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

  function clearSelection() {
    nodesMap.value.forEach(node => {
      if (node.selected !== 'off') {
        node.selected = 'off';
        node.contextOrigin = ContextOrigin.None;
      }
    });
    useUiStore().addToast('Selection cleared.', 'info');
  }

  function selectAllVisible() {
    visibleNodes.value.forEach(node => {
      if (!node.isDir && node.selected !== 'on' && !node.isIgnored) {
        toggleNodeSelection(node.path);
      }
    });
    useUiStore().addToast('All visible files selected.', 'success');
  }

  function selectFilesByRelPaths(relPaths: string[], origin: ContextOrigin = ContextOrigin.Git) {
    const relPathMap = new Map<string, FileNode>();
    nodesMap.value.forEach(node => {
      if (!node.isDir) relPathMap.set(node.relPath, node);
    });

    let selectedCount = 0;
    relPaths.forEach(relPath => {
      const node = relPathMap.get(relPath);
      if (node && node.selected !== 'on' && !node.isIgnored) {
        node.selected = 'on';
        node.contextOrigin = origin;
        _updateParentSelection(node.parentPath, nodesMap.value);
        selectedCount++;
      }
    });
    if (selectedCount > 0) {
      useUiStore().addToast(`${selectedCount} files added to context from ${origin}.`, 'success');
    } else {
      useUiStore().addToast(`No new files selected. They might be ignored or already selected.`, 'info');
    }
  }

  return {
    nodesMap, isLoading, error, activeNodePath, searchQuery, shotgunContextText,
    activeNode, visibleNodes, selectedFiles, contextSummary,
    fetchFileTree, clearProjectData, toggleNodeExpansion, toggleNodeSelection,
    setActiveNode: (path: string | null) => { activeNodePath.value = path; },
    buildContext, setShotgunContext, clearSelection, selectAllVisible, selectFilesByRelPaths,
  };
});