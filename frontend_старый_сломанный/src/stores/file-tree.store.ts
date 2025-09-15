import { defineStore } from "pinia";
import { ref, computed, readonly } from "vue";
import { useDebounce } from "@vueuse/core";
import { useTreeStateStore } from "./tree-state.store";
import type { FileNode, DomainFileNode } from "@/types/dto";
import { useAdvancedErrorHandler } from "@/composables/useErrorHandler";
import { APP_CONFIG } from "@/config/app-config";
// Import container and repository
import { container } from "@/infrastructure/container";
import type { ProjectRepository } from "@/domain/repositories/ProjectRepository";
import { FileTreeAnalysisService } from "@/domain/services/FileTreeAnalysisService";

const analysisService = new FileTreeAnalysisService();

export const useFileTreeStore = defineStore("file-tree", () => {
  const { handleStructuredError } = useAdvancedErrorHandler();
  const treeStateStore = useTreeStateStore();

  // Inject ProjectRepository
  const projectRepository: ProjectRepository = container.projectRepository;

  const nodes = ref<FileNode[]>([]);
  const nodesMap = ref<Map<string, FileNode>>(new Map());
  const nodesRelMap = ref<Map<string, FileNode>>(new Map());
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const searchQuery = ref("");
  const debouncedSearchQuery = useDebounce(searchQuery, APP_CONFIG.ui.keyboard.DEBOUNCE_DELAY);

  const visibleNodes = computed(() => {
    if (!debouncedSearchQuery.value.trim()) {
      return nodes.value;
    }
    const query = debouncedSearchQuery.value.toLowerCase();
    return filterNodes(nodes.value, query);
  });

  const selectedFiles = computed(() => {
    // CRITICAL: Ensure reactivity by accessing the Set size
    const selectedPathsSize = treeStateStore.selectedPaths.size;
    if (selectedPathsSize === 0) return [];
    
    // Use the new analysis service to count selected files
    return analysisService.collectAllFiles(
      Array.from(nodesMap.value.values()).filter(node => 
        treeStateStore.selectedPaths.has(node.path) && 
        !node.isDir && 
        !node.isIgnored
      ),
      APP_CONFIG.performance.limits.MAX_SELECTED_PATHS
    ).map(node => node.relPath);
  });

  const totalFiles = computed(() => {
    // Use the new analysis service to count total files
    return analysisService.countTotalFiles(nodes.value);
  });

  const rootPath = ref<string>("");
  const useGitignore = ref(true);
  const useCustomIgnore = ref(true);

  const hasFiles = computed(() => nodes.value.length > 0);
  const hasSelectedFiles = computed(() => selectedFiles.value.length > 0);

  // Добавляем rootNodes как computed свойство
  const rootNodes = computed(() => {
    if (!nodes.value || nodes.value.length === 0) return [];
    
    // Находим корневые узлы (те, у которых нет parentPath или parentPath пустая)
    return nodes.value.filter(node => !node.parentPath || node.parentPath === "");
  });

  // Получаем treeStateStore для синхронизации состояния
  // tree state already initialized above

  // Track expanded state - теперь используем treeStateStore
  function isNodeExpanded(path: string): boolean {
    return treeStateStore.expandedPaths.has(path);
  }

  function toggleNodeExpanded(path: string) {
    treeStateStore.toggleExpansion(path);
  }

  async function loadProject(projectPath: string) {
    try {
      rootPath.value = projectPath;
      await fetchFileTree(projectPath, useGitignore.value, useCustomIgnore.value);

      // Auto-expand root folder by default
      if (nodes.value.length > 0) {
        const rootNode = nodes.value[0];
        treeStateStore.expandedPaths.add(rootNode.path);
        
        // Also expand first-level folders but limit depth using centralized configuration
        if (rootNode.children) {
          rootNode.children.forEach((node) => {
            if (node.isDir && treeStateStore.expandedPaths.size < APP_CONFIG.performance.limits.MAX_EXPANDED_PATHS) {
              treeStateStore.expandedPaths.add(node.path);
            }
          });
        }
      }

      // Clear previous selections when loading new project
      clearSelection();
      
      console.log("✅ Project loaded successfully with", nodes.value.length, "nodes");
    } catch (err) {
      console.error("❌ Failed to load project:", err);
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    }
  }

  async function refreshFiles() {
    if (!rootPath.value) {
      console.warn("Cannot refresh files: no root path set");
      return;
    }

    console.log("Refreshing file tree for path:", rootPath.value);
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
    // Use the new analysis service to collect all files
    return analysisService.collectAllFiles(nodes.value, APP_CONFIG.performance.memory.MAX_FILE_LIMIT);
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

    // Добавляем новые файлы с ограничением из централизованной конфигурации
    let count = 0;
    const maxFiles = APP_CONFIG.performance.limits.MAX_SELECTED_PATHS;
    
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
      // Emit an event to show warning modal
      window.dispatchEvent(new CustomEvent('show-file-selection-warning', {
        detail: { fileCount: selectedFiles.value.length }
      }));
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
    isLoading.value = true;
    error.value = null;
    try {
      // Use ProjectRepository instead of apiService
      const domainNodes = await projectRepository.loadFileTree(projectPath, {
        useGitignore: useGit,
        useCustomIgnore: useCustom
      });
      if (!domainNodes || domainNodes.length === 0) {
        nodes.value = [];
        nodesMap.value.clear();
        nodesRelMap.value.clear();
        return [] as FileNode[];
      }
      
      // Limit the number of nodes to prevent memory issues
      const maxNodes = 5000; // Limit total nodes
      let nodeCount = 0;
      
      const limitNodes = (nodes: DomainFileNode[]): DomainFileNode[] => {
        if (nodeCount >= maxNodes) return [];
        
        return nodes.map(node => {
          if (nodeCount >= maxNodes) return null as unknown as DomainFileNode;
          
          nodeCount++;
          if (node.children) {
            node.children = limitNodes(node.children);
          }
          return node;
        }).filter(Boolean) as DomainFileNode[];
      };
      
      const limitedDomainNodes = limitNodes(domainNodes);
      const fileNodes = mapDomainNodesToFileNodes(limitedDomainNodes);
      nodes.value = fileNodes;
      buildNodesMap(fileNodes);
      return fileNodes;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      error.value = errorMessage;
      handleStructuredError(err, { operation: "Loading file tree", component: "FileTreeStore" });
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
        name: node.name || 'unknown',
        path: node.path || '',
        relPath: node.relPath || '',
        isDir: node.isDir ?? false,
        // CRITICAL: Ensure size is never undefined per memory guidance
        size: node.size ?? 0,
        parentPath,
        children: node.children
          ? mapDomainNodesToFileNodes(node.children, node.path)
          : undefined,
        isGitignored: node.isGitignored ?? false,
        isCustomIgnored: node.isCustomIgnored ?? false,
        isIgnored: (node.isIgnored ?? false) || (node.isGitignored ?? false) || (node.isCustomIgnored ?? false),
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

  function removeSelectedFile(filePath: string) {
    // Remove file from selection in treeStateStore
    treeStateStore.selectedPaths.delete(filePath);
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
    removeSelectedFile, // Add the new method
  };
});