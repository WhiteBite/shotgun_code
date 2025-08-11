import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { ListFiles } from '../../wailsjs/go/main/App';
import type { FileNode, DomainFileNode } from '@/types/dto';
import { TreeMode, GitStatus, ContextOrigin } from '@/types/enums';
import { toggleNodeSelection as toggleNodeSelectionService } from '@/services/treeManager';
import { useUiStore } from './uiStore';
import { useNotificationsStore } from './notificationsStore';

/**
 * Maps a DomainFileNode (from Go backend) to a frontend FileNode.
 * Adds frontend-specific state and ensures type safety.
 */
function mapDomainNodeToViewNode(node: DomainFileNode, depth: number, parentPath: string | null, expandedPaths: Set<string>): FileNode {
  return {
    name: node.name,
    path: node.path,
    relPath: node.relPath,
    isDir: node.isDir,
    children: node.children?.map(c => ({ path: c.path })), // Store only path for children to avoid recursion issues here

    depth,
    gitStatus: GitStatus.Clean, // Will be updated by GitStore
    contextOrigin: ContextOrigin.None, // Will be updated by selection logic
    isBinary: false, // Placeholder, can be detected by extension/content
    isIgnored: node.isCustomIgnored || node.isGitignored, // Use backend's ignore status

    expanded: expandedPaths.has(node.path) || depth < 1, // Restore from persistence or auto-expand root
    selected: 'off', // Default selection state
    parentPath,
  };
}

/**
 * Processes a flattened list of domain nodes in chunks to avoid blocking the main thread.
 * Returns a Map of FileNode instances, suitable for frontend use.
 */
function processNodesInChunks(nodes: DomainFileNode[], expandedPaths: Set<string>, chunkSize = 250): Promise<Map<string, FileNode>> {
  return new Promise((resolve) => {
    const map = new Map<string, FileNode>();
    const allNodes: {node: DomainFileNode, depth: number, parentPath: string | null}[] = [];

    function flatten(currentNodes: DomainFileNode[], depth: number, parentPath: string | null) {
      for (const node of currentNodes) {
        allNodes.push({ node, depth, parentPath });
        if (node.children) flatten(node.children, depth + 1, node.path);
      }
    }
    flatten(nodes, 0, null); // Start flattening from root

    let i = 0;
    function processChunk() {
      const end = Math.min(i + chunkSize, allNodes.length);
      for (; i < end; i++) {
        const { node, depth, parentPath } = allNodes[i];
        const fileNode = mapDomainNodeToViewNode(node, depth, parentPath, expandedPaths);
        map.set(node.path, fileNode);
      }

      if (i < allNodes.length) {
        setTimeout(processChunk, 0); // Yield to the event loop
      } else {
        resolve(map); // All nodes processed
      }
    }
    processChunk(); // Start the chunk processing
  });
}


export const useContextStore = defineStore('context', {
  state: () => ({
    nodesMap: new Map<string, FileNode>(),
    treeMode: TreeMode.Navigation,
    isLoading: false,
    error: null as string | null,
    activeNodePath: null as string | null,
    searchQuery: '', // For future search implementation
    expandedPaths: new Set<string>(), // Persisted state for expanded folders
    shotgunContextText: '', // Generated context text
  }),
  getters: {
    visibleNodes: (state) => {
      if (state.nodesMap.size === 0) return [];

      const query = state.searchQuery.toLowerCase().trim();
      const result: FileNode[] = [];

      if (query) {
        const matchingPaths = new Set<string>();
        for (const node of state.nodesMap.values()) {
          if (node.name.toLowerCase().includes(query)) {
            matchingPaths.add(node.path);
            let parentPath = node.parentPath;
            while (parentPath) {
              matchingPaths.add(parentPath);
              const parent = state.nodesMap.get(parentPath);
              parentPath = parent ? parent.parentPath : null;
            }
          }
        }
        // Return a flat, sorted list of all matching nodes and their parents
        return Array.from(state.nodesMap.values())
        .filter(node => matchingPaths.has(node.path))
        .sort((a,b) => a.path.localeCompare(b.path));
      }

      // Default tree view (no search)
      const roots = Array.from(state.nodesMap.values()).filter(n => n.depth === 0);

      function traverse(node: FileNode) {
        result.push(node);
        if (node.expanded && node.children) {
          const children = node.children.map(c => state.nodesMap.get(c.path)!).filter(Boolean);
          children
          .sort((a,b) => (a.isDir === b.isDir) ? a.name.localeCompare(b.name) : (a.isDir ? -1 : 1))
          .forEach(traverse);
        }
      }
      roots.sort((a,b) => (a.isDir === b.isDir) ? a.name.localeCompare(b.name) : (a.isDir ? -1 : 1))
      .forEach(traverse);
      return result;
    },

    // Selected files for context generation (computed dynamically)
    selectedFiles: (state) => {
      return Array.from(state.nodesMap.values()).filter(node => !node.isDir && node.selected === 'on');
    },

    // Summary metrics for context (computed dynamically)
    contextSummary: (state) => {
      const files = state.selectedFiles;
      const estimatedLines = files.length * 150; // Average lines per file
      const estimatedTokens = Math.round(estimatedLines * 1.25); // Average tokens per line
      return {
        files: files.length,
        lines: Math.max(0, estimatedLines),
        tokens: Math.max(0, estimatedTokens),
        cost: (estimatedTokens / 1000) * 0.005, // Rough estimate for gpt-4o input cost
      };
    },
  },
  actions: {
    // Fetches the file tree from the backend and processes it in chunks.
    async fetchFileTree(projectPath: string) {
      const notifications = useNotificationsStore();
      this.isLoading = true;
      this.error = null;
      this.nodesMap.clear(); // Clear existing map before loading new project data

      try {
        const treeData: DomainFileNode[] = await ListFiles(projectPath); // Use DomainFileNode type here
        this.nodesMap = await processNodesInChunks(treeData, this.expandedPaths);
        notifications.addLog('Дерево файлов успешно загружено.', 'success');
      } catch (err: any) {
        this.error = `Ошибка загрузки дерева файлов: ${err.message || err}`;
        notifications.addLog(this.error, 'error');
        this.nodesMap = new Map(); // Ensure map is empty on error
      } finally {
        this.isLoading = false;
      }
    },

    // Clears all project-specific data in the store.
    clearProjectData() {
      this.nodesMap.clear();
      this.activeNodePath = null;
      this.error = null;
      this.shotgunContextText = '';
      // Optionally clear expandedPaths if persistence should not carry over between projects
      // this.expandedPaths.clear();
    },

    // Updates Git statuses for nodes based on backend data.
    updateGitStatuses(statuses: { path: string; status: string }[]) {
      // Reset all file nodes to clean first
      this.nodesMap.forEach(node => { if (!node.isDir) node.gitStatus = GitStatus.Clean; });

      const statusMap: Record<string, GitStatus> = {
        'M': GitStatus.Modified, 'A': GitStatus.Untracked, '??': GitStatus.Untracked,
        'D': GitStatus.Modified, 'R': GitStatus.Modified, 'C': GitStatus.Conflict,
      };

      const relPathToNodeMap = new Map<string, FileNode>();
      this.nodesMap.forEach(node => { if (!node.isDir) relPathToNodeMap.set(node.relPath, node); });

      statuses.forEach(s => {
        const node = relPathToNodeMap.get(s.path);
        if (node) {
          node.gitStatus = statusMap[s.status] || GitStatus.Modified; // Default to modified if unknown
        }
      });
    },

    // Toggles expansion state of a directory node.
    toggleNodeExpansion(path: string) {
      const node = this.nodesMap.get(path);
      if (node?.isDir) {
        node.expanded = !node.expanded;
        if (node.expanded) {
          this.expandedPaths.add(path);
        } else {
          this.expandedPaths.delete(path);
        }
      }
    },

    // Sets the currently active/focused node (for UI highlighting).
    setActiveNode(path: string | null) {
      this.activeNodePath = path;
    },

    // Toggles selection state of a node (and its children/parents for consistency).
    toggleNodeSelection(path: string) {
      toggleNodeSelectionService(this.nodesMap, path); // Delegate to service for complex logic
    },

    // Generates context string based on selected files.
    async buildContext() {
      const projectStore = useProjectStore(); // Get instance here if needed
      const uiStore = useUiStore();
      const notifications = useNotificationsStore();

      if (!projectStore.currentProject) {
        notifications.addLog('Проект не выбран.', 'error');
        return;
      }
      const selectedNodes = this.selectedFiles; // Use getter
      const paths = selectedNodes.map(n => n.relPath);

      if (paths.length === 0) {
        uiStore.addToast('Не выбрано ни одного файла для контекста', 'info');
        return;
      }
      this.isLoading = true;
      notifications.addLog(`Сборка контекста из ${paths.length} файлов...`);
      try {
        const ctx: string = await RequestShotgunContextGeneration(projectStore.currentProject.path, paths);
        this.shotgunContextText = ctx;
        notifications.addLog('Контекст успешно собран.', 'success');
      } catch (err: any) {
        notifications.addLog(`Ошибка сборки контекста: ${err.message || err}`, 'error');
      } finally {
        this.isLoading = false;
      }
    },

    // Clears all selected files.
    clearSelection() {
      this.nodesMap.forEach(node => node.selected = 'off');
      // Need to re-evaluate parent selection states after clear
      Array.from(this.nodesMap.values()).filter(n => n.depth === 0)
      .forEach(root => toggleNodeSelectionService(this.nodesMap, root.path)); // Re-evaluate all from roots
      useUiStore().addToast('Выбор файлов очищен', 'info');
    },
  },
  // Persist expandedPaths and treeMode
  persist: {
    key: 'shotgun-context-store',
    paths: ['expandedPaths', 'treeMode'],
    serializer: {
      serialize: (state) => JSON.stringify({ ...state, expandedPaths: Array.from(state.expandedPaths || []) }),
      deserialize: (str) => {
        const state = JSON.parse(str);
        if (state.expandedPaths) state.expandedPaths = new Set(state.expandedPaths);
        return state;
      }
    }
  },
});