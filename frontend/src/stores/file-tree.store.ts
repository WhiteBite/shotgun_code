import { defineStore } from "pinia";
import { ref } from "vue";
import type { FileNode, DomainFileNode } from "@/types/dto";
import { GitStatus, ContextOrigin } from "@/types/enums";
import { useNotificationsStore } from "./notifications.store";
import { useSettingsStore } from "./settings.store";
import { useTreeStateStore } from "./tree-state.store";
import { apiService } from "@/services/api.service";

function mapDomainNodeToViewNode(
    node: DomainFileNode,
    depth: number,
    parentPath: string | null,
): FileNode {
  return {
    name: node.name,
    path: node.path,
    relPath: node.relPath,
    isDir: node.isDir,
    children: node.children?.map((c) => ({ path: c.path })),
    depth,
    gitStatus: GitStatus.Unmodified,
    contextOrigin: ContextOrigin.None,
    isBinary: false,
    isIgnored: node.isCustomIgnored || node.isGitignored,
    isGitignored: node.isGitignored,
    isCustomIgnored: node.isCustomIgnored,
    parentPath,
    size: node.size || 0,
  };
}

export const useFileTreeStore = defineStore("fileTree", () => {
  const notifications = useNotificationsStore();
  const settingsStore = useSettingsStore();
  const treeStateStore = useTreeStateStore();

  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const rootPath = ref<string>("");
  const searchQuery = ref("");

  const nodesMap = ref(new Map<string, FileNode>());
  const rootNodes = ref<FileNode[]>([]);

  async function fetchFileTree() {
    if (!rootPath.value) {
      notifications.addLog("No project root path set. Aborting fetch.", "warn");
      return;
    }
    if (isLoading.value) {
      notifications.addLog("fetchFileTree already in progress.", "debug");
      return;
    }

    isLoading.value = true;
    error.value = null;
    notifications.addLog(`Starting file tree fetch for: ${rootPath.value}`, "info");

    try {
      const { useGitignore, useCustomIgnore } = settingsStore.settings;
      const treeData = await apiService.listFiles(rootPath.value, useGitignore, useCustomIgnore);

      // Process nodes into a map and a root list
      const newNodesMap = new Map<string, FileNode>();
      const newRootNodes: FileNode[] = [];

      const stack: {
        node: DomainFileNode;
        depth: number;
        parentPath: string | null;
      }[] = treeData.map((n) => ({ node: n, depth: 0, parentPath: null }));

      while (stack.length > 0) {
        const { node, depth, parentPath } = stack.pop()!;
        const viewNode = mapDomainNodeToViewNode(node, depth, parentPath);
        newNodesMap.set(viewNode.path, viewNode);

        if (depth === 0) {
          newRootNodes.push(viewNode);
        }

        if (node.children) {
          for (const child of [...node.children].reverse()) {
            stack.push({ node: child, depth: depth + 1, parentPath: node.path });
          }
        }
      }

      nodesMap.value = newNodesMap;
      rootNodes.value = newRootNodes.sort((a, b) => {
        if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
        return a.name.localeCompare(b.name);
      });

      // Auto-expand the root node(s)
      rootNodes.value.forEach((r) => treeStateStore.expandedPaths.add(r.path));

      await updateGitStatuses();
      notifications.addLog(`Fetched ${nodesMap.value.size} total nodes.`, "info");
    } catch (err: any) {
      const message = err.message || String(err);
      error.value = `Failed to load file tree: ${message}`;
      notifications.addLog(error.value, "error");
    } finally {
      isLoading.value = false;
    }
  }

  async function updateGitStatuses() {
    if (!rootPath.value) return;
    try {
      const isGit = await apiService.isGitAvailable();
      if (!isGit) return;

      const statuses = await apiService.getUncommittedFiles(rootPath.value);
      const statusMap = new Map<string, GitStatus>();
      statuses.forEach((s) => {
        let mappedStatus: GitStatus;
        switch (s.status) {
          case "M": mappedStatus = GitStatus.Modified; break;
          case "A": mappedStatus = GitStatus.Added; break;
          case "D": mappedStatus = GitStatus.Deleted; break;
          case "R": mappedStatus = GitStatus.Renamed; break;
          case "C": mappedStatus = GitStatus.Copied; break;
          case "U": case "??": mappedStatus = GitStatus.Untracked; break;
          case "UM": mappedStatus = GitStatus.UnmergedConflict; break;
          default: mappedStatus = GitStatus.Unmodified; break;
        }
        statusMap.set(s.path, mappedStatus);
      });

      nodesMap.value.forEach((node) => {
        if (!node.isDir) {
          node.gitStatus = statusMap.get(node.relPath) || GitStatus.Unmodified;
        }
      });
    } catch (err: any) {
      notifications.addLog(`Git status check failed: ${err.message}`, "error");
    }
  }

  function clearProjectData() {
    isLoading.value = false;
    error.value = null;
    rootPath.value = "";
    searchQuery.value = "";
    nodesMap.value.clear();
    rootNodes.value = [];
    treeStateStore.resetState();
    notifications.addLog("File tree data cleared", "info");
  }

  return {
    isLoading,
    error,
    rootPath,
    searchQuery,
    nodesMap,
    rootNodes,
    fetchFileTree,
    clearProjectData,
  };
});