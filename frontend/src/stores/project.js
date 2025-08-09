
import { defineStore } from 'pinia';
import { ref, reactive } from 'vue';
import {
  SelectDirectory,
  ListFiles,
  RequestShotgunContextGeneration,
  StartFileWatcher,
  StopFileWatcher
} from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { useNotificationsStore } from './notifications';
import { useSettingsStore } from './settings';
import { useStepsStore } from './steps';

export const useProjectStore = defineStore('project', () => {
  // ИСПРАВЛЕНО: Все use...Store() вызовы перенесены внутрь
  const notifications = useNotificationsStore();
  const settings = useSettingsStore();
  const steps = useStepsStore();

  const projectRoot = ref('');
  const fileTree = ref([]);
  const loadingError = ref('');
  const isFileTreeLoading = ref(false);
  const isGeneratingContext = ref(false);
  const generationProgress = ref({ current: 0, total: 0 });
  const shotgunPromptContext = ref('');
  const manuallyToggledNodes = reactive(new Map());

  let contextDebounceTimer = null;

  function isAnyParentVisuallyExcluded(node) {
    if (!node || !node.parent) return false;
    let current = node.parent;
    while (current) {
      if (current.excluded) return true;
      current = current.parent;
    }
    return false;
  }

  function toggleExcludeNode(nodeToToggle) {
    if (isAnyParentVisuallyExcluded(nodeToToggle) && nodeToToggle.excluded) {
      nodeToToggle.excluded = false;
    } else {
      nodeToToggle.excluded = !nodeToToggle.excluded;
    }
    manuallyToggledNodes.set(nodeToToggle.relPath, nodeToToggle.excluded);
    notifications.addLog(`Исключение для '${nodeToToggle.name}' установлено в: ${nodeToToggle.excluded}`, 'info');
    triggerShotgunContextGeneration();
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

  function mapDataToTreeRecursive(nodes, parent) {
    if (!nodes) return [];
    return nodes.map(node => {
      const isRootNode = parent === null;
      const reactiveNode = reactive({
        ...node,
        expanded: node.isDir ? isRootNode : false,
        parent: parent,
        children: []
      });
      if (node.children && node.children.length > 0) {
        reactiveNode.children = mapDataToTreeRecursive(node.children, reactiveNode);
      }
      return reactiveNode;
    });
  }

  async function loadFileTree(dirPath) {
    isFileTreeLoading.value = true;
    loadingError.value = '';
    notifications.addLog(`Загрузка дерева файлов для: ${dirPath}`, 'info');
    try {
      const treeData = await ListFiles(dirPath);
      fileTree.value = mapDataToTreeRecursive(treeData, null);
      updateAllNodesExcludedStateRecursive(fileTree.value, false);
      notifications.addLog(`Дерево файлов успешно загружено.`, 'success');
    } catch (err) {
      const errorMsg = `Ошибка загрузки дерева файлов: ${err.message || err}`;
      loadingError.value = errorMsg;
      fileTree.value = [];
    } finally {
      isFileTreeLoading.value = false;
    }
  }

  async function selectProjectFolder() {
    try {
      if (projectRoot.value) {
        await StopFileWatcher();
      }
      const selectedDir = await SelectDirectory();
      if (selectedDir) {
        projectRoot.value = selectedDir;
        manuallyToggledNodes.clear();
        fileTree.value = [];
        steps.resetSteps();
        await loadFileTree(selectedDir);
        await StartFileWatcher(selectedDir);
        triggerShotgunContextGeneration();
      }
    } catch (err) {
      const errorMsg = `Ошибка выбора директории: ${err.message || err}`;
      loadingError.value = errorMsg;
      notifications.addLog(errorMsg, 'error');
    }
  }

  function triggerShotgunContextGeneration() {
    clearTimeout(contextDebounceTimer);
    if (!projectRoot.value || isFileTreeLoading.value) {
      return;
    }
    isGeneratingContext.value = true;
    generationProgress.value = { current: 0, total: 0 };

    contextDebounceTimer = setTimeout(() => {
      const excludedPaths = [];
      // ... (логика сбора excludedPaths)
      RequestShotgunContextGeneration(projectRoot.value, excludedPaths);
    }, 750);
  }

  function setupWailsListeners() {
    EventsOn("shotgunContextGenerated", (output) => {
      shotgunPromptContext.value = output;
      isGeneratingContext.value = false;
      notifications.addLog(`Контекст проекта обновлен (${output.length} символов).`, 'success');
      steps.completeStep(1);
    });
    EventsOn("shotgunContextGenerationProgress", (progress) => {
      generationProgress.value = progress;
    });
    EventsOn("projectFilesChanged", async (changedRootDir) => {
      if (changedRootDir === projectRoot.value) {
        notifications.addLog('Обнаружены изменения в файлах проекта, обновляем дерево...', 'info');
        await loadFileTree(projectRoot.value);
        triggerShotgunContextGeneration();
      }
    });
  }

  setupWailsListeners();

  return {
    projectRoot,
    fileTree,
    loadingError,
    isFileTreeLoading,
    isGeneratingContext,
    generationProgress,
    shotgunPromptContext,
    selectProjectFolder,
    toggleExcludeNode,
    updateAllNodesExcludedState: () => updateAllNodesExcludedStateRecursive(fileTree.value, false),
    triggerShotgunContextGeneration,
  };
});