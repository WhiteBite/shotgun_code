import { ref, watch } from "vue";
import { useProjectStore } from "@/stores/project.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { container } from "@/infrastructure/container";
import type { StorageRepository } from '@/domain/repositories/StorageRepository';

interface SessionData {
  selectedFiles: string[];
  expandedPaths: string[];
}

const KEY_PREFIX = "shotgun_session_";

export function setupProjectSession() {
  const projectStore = useProjectStore();
  const fileTreeStore = useFileTreeStore();
  const contextBuilderStore = useContextBuilderStore();
  const treeStateStore = useTreeStateStore();
  const restored = ref(false);
  
  // Get storageRepository from DI container
  const storageRepository = container.storageRepository;

  function sessionKey(): string | null {
    return projectStore.currentProject
      ? KEY_PREFIX + projectStore.currentProject.path
      : null;
  }

  function save() {
    const key = sessionKey();
    if (!key) return;

    const data: SessionData = {
      selectedFiles: contextBuilderStore.selectedFiles,
      expandedPaths: Array.from(treeStateStore.expandedPaths),
    };

    try {
      storageRepository.set(key, data);
    } catch {
      // ignore storage errors
    }
  }

  async function loadProject(projectPath: string) {
    try {
      await fileTreeStore.loadProject(projectPath);
    } catch (error) {
      console.error("Failed to load project:", error);
    }
  }

  function tryRestore() {
    if (restored.value) return;
    if (!projectStore.currentProject) return;
    if (fileTreeStore.isLoading) return;
    if (fileTreeStore.nodesMap.size === 0) return;

    const key = sessionKey();
    if (!key) {
      restored.value = true;
      return;
    }

    const data = storageRepository.get<SessionData>(key);
    if (!data) {
      restored.value = true;
      return;
    }

    try {
      // Восстанавливаем развернутые пути
      if (Array.isArray(data.expandedPaths)) {
        data.expandedPaths.forEach((path) => {
          treeStateStore.expandedPaths.add(path);
        });
      }

      // Восстанавливаем выделенные файлы
      if (Array.isArray(data.selectedFiles) && data.selectedFiles.length > 0) {
        // Используем treeStateStore для восстановления выделения
        data.selectedFiles.forEach((relPath) => {
          const node = fileTreeStore.getFileByRelPath(relPath);
          if (
            node &&
            !node.isDir &&
            !node.isGitignored &&
            !node.isCustomIgnored
          ) {
            // Используем treeStateStore для выделения
            treeStateStore.selectedPaths.add(node.path);
          }
        });

        // Обновляем contextBuilderStore
        contextBuilderStore.setSelectedFiles(data.selectedFiles);
      }
    } catch {
      // ignore parse errors
    }

    restored.value = true;
  }

  // Watch for project changes and load files
  watch(
    () => projectStore.currentProject?.path,
    async (newPath) => {
      if (newPath) {
        await loadProject(newPath);
      } else {
        fileTreeStore.clearProject();
        contextBuilderStore.clearSelectedFiles();
        treeStateStore.resetState();
      }
      restored.value = false;
    },
    { immediate: true },
  );

  // autosave selected files and expanded paths
  watch(
    () => [contextBuilderStore.selectedFiles, treeStateStore.expandedPaths],
    save,
    { deep: true },
  );

  // restore when both project loaded and tree ready
  watch(
    () => [
      projectStore.currentProject?.path,
      fileTreeStore.isLoading,
      fileTreeStore.nodesMap.size,
    ],
    () => tryRestore(),
    { immediate: true },
  );

  // reset restore flag on project change
  watch(
    () => projectStore.currentProject?.path,
    () => {
      restored.value = false;
    },
  );
}