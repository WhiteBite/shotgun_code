import { defineStore } from 'pinia';
import { ref, reactive } from 'vue';
import { ListFiles, StartFileWatcher, StopFileWatcher } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { useNotificationsStore } from './notificationsStore.js';
import { useSettingsStore } from './settingsStore.js';
import { useProjectStore } from './projectStore.js';

export const useFileTreeStore = defineStore('fileTree', () => {
  const notifications = useNotificationsStore();
  const settings = useSettingsStore();

  const tree = ref([]);
  const loadingError = ref('');
  const isFileTreeLoading = ref(false);
  const manuallyToggledNodes = reactive(new Map());

  function mapDataToTreeRecursive(nodes, parent) {
    if (!nodes) return [];
    return nodes.map(node => {
      const isRootNode = parent === null;
      const reactiveNode = reactive({ ...node, expanded: node.isDir ? isRootNode : false, parent: parent, children: [] });
      if (node.children && node.children.length > 0) {
        reactiveNode.children = mapDataToTreeRecursive(node.children, reactiveNode);
      }
      return reactiveNode;
    });
  }

  function updateAllNodesExcludedStateRecursive(nodes, parentIsVisuallyExcluded) {
    if (!nodes || nodes.length === 0) return;
    nodes.forEach(node => {
      const manualToggle = manuallyToggledNodes.get(node.relPath);
      let isExcludedByRule = (settings.useGitignore && node.isGitignored) || (settings.useCustomIgnore && node.isCustomIgnored);
      if (manualToggle !== undefined) {
        node.excluded = manualToggle;
      } else {
        node.excluded = isExcludedByRule || parentIsVisuallyExcluded;
      }
      if (node.children && node.children.length > 0) {
        updateAllNodesExcludedStateRecursive(node.children, node.excluded);
      }
    });
  }

  async function loadFileTree(dirPath) {
    isFileTreeLoading.value = true;
    loadingError.value = '';
    try {
      const treeData = await ListFiles(dirPath);
      tree.value = mapDataToTreeRecursive(treeData, null);
      updateAllNodesExcludedStateRecursive(tree.value, false);
    } catch (err) {
      loadingError.value = `Ошибка загрузки дерева файлов: ${err.message || err}`;
      notifications.addLog(loadingError.value, 'error');
    } finally {
      isFileTreeLoading.value = false;
    }
  }

  function setChildrenExcludedState(node, isExcluded) {
    if (!node.isDir || !node.children) return;
    node.children.forEach(child => {
      child.excluded = isExcluded;
      manuallyToggledNodes.set(child.relPath, isExcluded);
      if (child.isDir) {
        setChildrenExcludedState(child, isExcluded);
      }
    });
  }

  function toggleExcludeNode(nodeToToggle) {
    if (nodeToToggle.isDir) {
      const areAllChildrenExcluded = (node) => {
        if (!node.children || node.children.length === 0) return node.excluded;
        return node.children.every(child => areAllChildrenExcluded(child));
      };
      const newExcludedState = !areAllChildrenExcluded(nodeToToggle);
      nodeToToggle.excluded = newExcludedState;
      manuallyToggledNodes.set(nodeToToggle.relPath, newExcludedState);
      setChildrenExcludedState(nodeToToggle, newExcludedState);
    } else {
      nodeToToggle.excluded = !nodeToToggle.excluded;
      manuallyToggledNodes.set(nodeToToggle.relPath, nodeToToggle.excluded);
    }
    notifications.addLog(`Исключение для '${nodeToToggle.name}' обновлено.`, 'debug');
  }

  function applySelectionSet(selectionSet) {
    manuallyToggledNodes.clear();
    const inclusiveSet = new Set();

    if (selectionSet.size > 0) {
      inclusiveSet.add('.');
    }

    selectionSet.forEach(path => {
      const normalizedPath = path.replace(/\\/g, '/');
      inclusiveSet.add(normalizedPath);

      let parentPath = normalizedPath;
      let lastSlashIndex;
      while ((lastSlashIndex = parentPath.lastIndexOf('/')) !== -1) {
        parentPath = parentPath.substring(0, lastSlashIndex);
        inclusiveSet.add(parentPath);
      }
    });

    function traverseAndApply(nodes) {
      if (!nodes) return;
      nodes.forEach(node => {
        const normalizedPath = node.relPath.replace(/\\/g, '/');
        const shouldBeIncluded = inclusiveSet.has(normalizedPath);
        node.excluded = !shouldBeIncluded;
        manuallyToggledNodes.set(node.relPath, node.excluded);

        if (node.isDir && shouldBeIncluded) {
          node.expanded = true;
        }

        if (node.children) {
          traverseAndApply(node.children);
        }
      });
    }

    traverseAndApply(tree.value);
  }

  function collectIncludedPaths(nodes) {
    const included = [];
    function recurse(nodesToScan) {
      if (!nodesToScan) return;
      nodesToScan.forEach(node => {
        if (!node.excluded && !node.isDir) {
          included.push(node.relPath);
        }
        if (node.children && node.children.length > 0) {
          recurse(node.children);
        }
      });
    }
    recurse(nodes);
    return included;
  }

  function hasVisuallyIncludedDescendant(node) {
    if (!node.isDir || !node.children || node.children.length === 0) return false;
    for (const child of node.children) {
      if (!child.excluded) return true;
      if (hasVisuallyIncludedDescendant(child)) return true;
    }
    return false;
  }

  function collectTrulyExcludedPaths(nodes) {
    const excludedPathsArray = [];
    function recurse(nodesToScan) {
      if (!nodesToScan) return;
      nodesToScan.forEach(node => {
        if (node.excluded && !hasVisuallyIncludedDescendant(node)) {
          excludedPathsArray.push(node.relPath);
        } else if (node.children && node.children.length > 0) {
          recurse(node.children);
        }
      });
    }
    recurse(nodes);
    return excludedPathsArray;
  }

  async function startWatcher(path) {
    await StartFileWatcher(path);
  }

  async function stopWatcher() {
    await StopFileWatcher();
  }

  function reset() {
    tree.value = [];
    loadingError.value = '';
    isFileTreeLoading.value = false;
    manuallyToggledNodes.clear();
  }

  function setupWailsListeners() {
    EventsOn("projectFilesChanged", async (changedRootDir) => {
      const projectRoot = useProjectStore().projectRoot;
      if (changedRootDir === projectRoot) {
        await loadFileTree(projectRoot);
      }
    });
  }

  setupWailsListeners();

  return {
    tree, loadingError, isFileTreeLoading, manuallyToggledNodes,
    loadFileTree, toggleExcludeNode,
    collectTrulyExcludedPaths, startWatcher, stopWatcher, reset, applySelectionSet,
    collectIncludedPaths,
  };
});