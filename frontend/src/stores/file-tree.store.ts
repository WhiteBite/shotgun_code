import { defineStore } from "pinia";
import { ref, computed, readonly } from "vue";
import { useDebounce } from "@vueuse/core";
import { useTreeStateStore } from "./tree-state.store";
import type { FileNode, DomainFileNode } from "@/types/dto";
import { apiService } from "@/infrastructure/api/api.service";
import { useAdvancedErrorHandler } from "@/composables/useErrorHandler";

export const useFileTreeStore = defineStore("file-tree", () => {
  const { handleStructuredError } = useAdvancedErrorHandler();
  const treeStateStore = useTreeStateStore();

  const nodes = ref<FileNode[]>([]);
  const nodesMap = ref<Map<string, FileNode>>(new Map());
  const nodesRelMap = ref<Map<string, FileNode>>(new Map());
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const searchQuery = ref("");
  const debouncedSearchQuery = useDebounce(searchQuery, 200);

  const visibleNodes = computed(() => {
    if (!debouncedSearchQuery.value.trim()) {
      return nodes.value;
    }
    const query = debouncedSearchQuery.value.toLowerCase();
    return filterNodes(nodes.value, query);
  });

  const selectedFiles = computed(() => {
    const selected: string[] = [];
    let count = 0;
    const maxSelectedFiles = 500; // Further reduced from 1000 to 500 to prevent memory issues
    
    nodesMap.value.forEach((node, path) => {
      if (count >= maxSelectedFiles) {
        if (count === maxSelectedFiles) {
          console.warn(`Maximum number of selected files (${maxSelectedFiles}) reached. Additional files will be ignored.`);
          count++; // Increment to prevent multiple warnings
        }
        return;
      }
      
      if (
        treeStateStore.selectedPaths.has(path) &&
        !node.isDir &&
        !node.isIgnored
      ) {
        selected.push(node.relPath);
        count++;
      }
    });
    
    return selected;
  });

  const totalFiles = computed(() => {
    let count = 0;
    const countFiles = (ns: FileNode[]) => {
      for (const n of ns) {
        if (!n.isDir) count++;
        if (n.children) countFiles(n.children);
      }
    };
    countFiles(nodes.value);
    return count;
  });

  const rootPath = ref<string>("");
  const useGitignore = ref(true);
  const useCustomIgnore = ref(true);

  const hasFiles = computed(() => nodes.value.length > 0);
  const hasSelectedFiles = computed(() => selectedFiles.value.length > 0);

  // Добавляем rootNodes как computed свойство
  const rootNodes = computed(() => {
    if (!nodes.value || nodes.value.length === 0) return [];
    
    // Находим корневые узлы (те, у которых нет parentPath или parentPath пустой)
    return nodes.value.filter(node => !node.parentPath || node.parentPath === "");
  });

  // Получаем treeStateStore для синхронизации состояния
  // tree state already initialized above

  // Track expanded state - теперь используем treeStateStore
  const expandedPaths = computed(() => treeStateStore.expandedPaths);

  function isNodeExpanded(path: string): boolean {
    return treeStateStore.expandedPaths.has(path);
  }

  function toggleNodeExpanded(path: string) {
    treeStateStore.toggleExpansion(path);
  }

  async function loadProject(projectPath: string) {
    console.log('Loading project:', projectPath);
    rootPath.value = projectPath;
    await fetchFileTree(projectPath, useGitignore.value, useCustomIgnore.value);

    console.log('File tree loaded, nodes:', nodes.value.length);
    console.log('Root nodes:', rootNodes.value.length);

    // Auto-expand root folder by default
    if (nodes.value.length > 0) {
      const rootNode = nodes.value[0];
      treeStateStore.expandedPaths.add(rootNode.path);
      
      // Also expand first-level folders but limit depth
      if (rootNode.children) {
        rootNode.children.forEach((node) => {
          if (node.isDir && treeStateStore.expandedPaths.size < 20) { // Limit expanded folders
            treeStateStore.expandedPaths.add(node.path);
          }
        });
      }
    }

    // Clear previous selections when loading new project
    clearSelection();
  }

  async function refreshFiles() {
    if (!rootPath.value) {
      return;
    }

    await fetchFileTree(
      rootPath.value,
      useGitignore.value,
      useCustomIgnore.value,
    );
  }

  async function updateIgnoreSettings(
    newUseGitignore: boolean,
    newUseCustomIgnore: boolean,
  ) {
    useGitignore.value = newUseGitignore;
    useCustomIgnore.value = newUseCustomIgnore;

    if (rootPath.value) {
      await refreshFiles();
    }
  }

  function clearProject() {
    rootPath.value = "";
    clearSelection();
    clearError();
    // Очищаем состояние развернутых папок
    treeStateStore.resetState();
  }

  function getAllFiles(): FileNode[] {
    const allFiles: FileNode[] = [];

    const collectFiles = (ns: readonly FileNode[]) => {
      for (const n of ns) {
        if (!n.isDir && allFiles.length < 1000) { // Limit total files to prevent memory issues
          allFiles.push(n);
        }
        if (n.children && allFiles.length < 1000) {
          collectFiles(n.children);
        }
      }
    };

    collectFiles(nodes.value);
    return allFiles;
  }

  function getFileByPath(path: string): FileNode | undefined {
    return nodesMap.value.get(path);
  }

  function getFileByRelPath(relPath: string): FileNode | undefined {
    return nodesRelMap.value.get(relPath);
  }

  function setSelectedFiles(filePaths: string[]) {
    // Очищаем текущий выбор
    clearSelection();

    // Добавляем новые файлы с ограничением
    let count = 0;
    const maxFiles = 200; // Limit selected files to prevent memory issues
    
    filePaths.forEach((relPath) => {
      if (count >= maxFiles) {
        if (count === maxFiles) {
          console.warn(`Maximum number of files to select (${maxFiles}) reached. Additional files will be ignored.`);
          count++;
        }
        return;
      }
      
      // Ищем файл по relPath
      const node = getFileByRelPath(relPath);
      if (node && !node.isDir && !node.isGitignored && !node.isCustomIgnored) {
        // Используем toggleNodeSelection для правильного обновления состояния
        toggleNodeSelection(node.path);
        count++;
      }
    });
  }

  // Internal helpers
  function clearError() {
    error.value = null;
  }

  function setSearchQuery(q: string) {
    searchQuery.value = q;
  }

  function toggleNodeSelection(path: string) {
    const node = nodesMap.value.get(path);
    if (!node) return;
    
    // Check if we're at the limit before adding
    if (!treeStateStore.selectedPaths.has(path) && selectedFiles.value.length >= 500) {
      console.warn("Maximum number of selected files (500) reached. Cannot select more files.");
      return;
    }
    
    treeStateStore.toggleNodeSelection(path, nodesMap.value as Map<string, FileNode>);
  }

  function clearSelection() {
    treeStateStore.clearSelection();
  }

  function selectAll() {
    let count = 0;
    const maxFiles = 200; // Limit total selected files
    
    nodesMap.value.forEach((node) => {
      if (count >= maxFiles) {
        if (count === maxFiles) {
          console.warn(`Maximum number of files to select (${maxFiles}) reached during select all operation.`);
          count++;
        }
        return;
      }
      
      if (!node.isDir && !node.isGitignored && !node.isCustomIgnored) {
        treeStateStore.selectedPaths.add(node.path);
        count++;
      }
    });
  }

  // fetch & map
  async function fetchFileTree(
    projectPath: string,
    useGit: boolean,
    useCustom: boolean,
  ) {
    console.log('Fetching file tree for:', projectPath);
    isLoading.value = true;
    error.value = null;
    try {
      const domainNodes = await apiService.listFiles(projectPath, useGit, useCustom);
      console.log('Received domain nodes:', domainNodes?.length || 0);
      if (!domainNodes || domainNodes.length === 0) {
        nodes.value = [];
        nodesMap.value.clear();
        nodesRelMap.value.clear();
        console.log('No nodes received, clearing file tree');
        return [] as FileNode[];
      }
      
      // Limit the number of nodes to prevent memory issues
      const maxNodes = 5000; // Limit total nodes
      let nodeCount = 0;
      
      const limitNodes = (nodes: DomainFileNode[]): DomainFileNode[] => {
        if (nodeCount >= maxNodes) return [];
        
        return nodes.map(node => {
          if (nodeCount >= maxNodes) return null as any;
          
          nodeCount++;
          if (node.children) {
            node.children = limitNodes(node.children);
          }
          return node;
        }).filter(Boolean) as DomainFileNode[];
      };
      
      const limitedDomainNodes = limitNodes(domainNodes);
      console.log('Limited domain nodes:', limitedDomainNodes.length);
      const fileNodes = mapDomainNodesToFileNodes(limitedDomainNodes);
      console.log('Mapped file nodes:', fileNodes.length);
      nodes.value = fileNodes;
      buildNodesMap(fileNodes);
      console.log('File tree updated, total nodes:', fileNodes.length);
      return fileNodes;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      error.value = errorMessage;
      handleStructuredError(err, { operation: "Loading file tree", component: "FileTreeStore" });
      console.error('Error fetching file tree:', errorMessage);
      throw new Error(errorMessage);
    } finally {
      isLoading.value = false;
    }
  }

  function mapDomainNodesToFileNodes(
    domainNodes: DomainFileNode[],
    parentPath: string | null = null,
  ): FileNode[] {
    return domainNodes.map((node: DomainFileNode) => {
      const fileNode: FileNode = {
        name: node.name,
        path: node.path,
        relPath: node.relPath,
        isDir: node.isDir,
        size: node.size,
        parentPath,
        children: node.children
          ? mapDomainNodesToFileNodes(node.children, node.path)
          : undefined,
        isGitignored: node.isGitignored,
        isCustomIgnored: node.isCustomIgnored,
        isIgnored: node.isIgnored || node.isGitignored || node.isCustomIgnored,
        // isSelected removed - managed by treeStateStore.selectedPaths
      };
      return fileNode;
    });
  }

  function buildNodesMap(fileNodes: FileNode[]) {
    const map = new Map<string, FileNode>();
    const relMap = new Map<string, FileNode>();
    const addToMap = (ns: FileNode[]) => {
      for (const n of ns) {
        map.set(n.path, n);
        relMap.set(n.relPath, n);
        if (n.children) addToMap(n.children);
      }
    };
    addToMap(fileNodes);
    nodesMap.value = map;
    nodesRelMap.value = relMap;
  }

  function filterNodes(ns: FileNode[], query: string): FileNode[] {
    const filtered: FileNode[] = [];
    for (const n of ns) {
      const matchesQuery =
        n.name.toLowerCase().includes(query) ||
        n.relPath.toLowerCase().includes(query);
      if (matchesQuery) {
        const filteredNode: FileNode = { ...n };
        if (n.children) filteredNode.children = filterNodes(n.children, query);
        filtered.push(filteredNode);
      } else if (n.children) {
        const filteredChildren = filterNodes(n.children, query);
        if (filteredChildren.length > 0) {
          const filteredNode: FileNode = { ...n, children: filteredChildren };
          filtered.push(filteredNode);
        }
      }
    }
    return filtered;
  }

  return {
    // State
    rootPath: readonly(rootPath),
    useGitignore,
    useCustomIgnore,

    // File tree
    nodes: readonly(nodes),
    nodesMap: readonly(nodesMap),
    nodesRelMap: readonly(nodesRelMap),
    isLoading: readonly(isLoading),
    error: readonly(error),
    searchQuery,
    visibleNodes,
    selectedFiles,
    totalFiles,

    // Computed
    hasFiles,
    hasSelectedFiles,
    rootNodes,

    // Methods
    loadProject,
    refreshFiles,
    updateIgnoreSettings,
    clearProject,
    getAllFiles,
    getFileByPath,
    getFileByRelPath,
    toggleNodeSelection,
    clearSelection,
    selectAll,
    setSearchQuery,
    clearError,
    isNodeExpanded,
    toggleNodeExpanded,
    setSelectedFiles,
  };
});