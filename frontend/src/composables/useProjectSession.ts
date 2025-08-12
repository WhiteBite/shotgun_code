import { watch } from 'vue';
import { useProjectStore } from '@/stores/project.store';
import { useGenerationStore } from '@/stores/generation.store';
import { useContextStore } from '@/stores/context.store';
import { storeToRefs } from 'pinia';

interface SessionData {
  userTask: string;
  selectedFilePaths: string[];
}

const SESSION_STORAGE_KEY_PREFIX = 'shotgun_session_';

export function useProjectSession() {
  const projectStore = useProjectStore();
  const generationStore = useGenerationStore();
  const contextStore = useContextStore();

  const { currentProject } = storeToRefs(projectStore);
  const { userTask } = storeToRefs(generationStore);
  const { selectedFiles } = storeToRefs(contextStore);

  const getSessionKey = (): string | null => {
    return currentProject.value ? `${SESSION_STORAGE_KEY_PREFIX}${currentProject.value.path}` : null;
  };

  const saveSession = () => {
    const key = getSessionKey();
    if (!key) return;

    const data: SessionData = {
      userTask: userTask.value,
      selectedFilePaths: selectedFiles.value.map(f => f.relPath),
    };
    localStorage.setItem(key, JSON.stringify(data));
  };

  const loadSession = () => {
    const key = getSessionKey();
    if (!key) return;

    const savedData = localStorage.getItem(key);
    if (savedData) {
      try {
        const data: SessionData = JSON.parse(savedData);
        generationStore.userTask = data.userTask || '';
        if (data.selectedFilePaths && data.selectedFilePaths.length > 0) {
          // Wait for file tree to be loaded before selecting files
          const unwatch = watch(() => contextStore.isLoading, (isLoading) => {
            if (!isLoading) {
              contextStore.selectFilesByRelPaths(data.selectedFilePaths);
              unwatch(); // Stop watching after selection is applied
            }
          });
        }
      } catch (e) {
        console.error("Failed to parse session data:", e);
      }
    }
  };

  // Watch for changes and save them automatically
  watch([userTask, selectedFiles], saveSession, { deep: true });

  return {
    loadSession,
  };
}