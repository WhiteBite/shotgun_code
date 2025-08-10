import { defineStore } from 'pinia';
import { ref } from 'vue';
import { SelectDirectory } from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notificationsStore';
import { useStepsStore } from './stepsStore';
import { useFileTreeStore } from './fileTreeStore';
import { useContextStore } from './contextStore';
import { useGitStore } from './gitStore';

export const useProjectStore = defineStore('projectState', () => {
  const notifications = useNotificationsStore();
  const steps = useStepsStore();
  const fileTree = useFileTreeStore();
  const context = useContextStore();
  const git = useGitStore();

  const projectRoot = ref('');

  async function selectProjectFolder() {
    try {
      if (projectRoot.value) await fileTree.stopWatcher();
      const selectedDir = await SelectDirectory();
      if (selectedDir) {
        projectRoot.value = selectedDir;

        steps.resetSteps();
        fileTree.reset();
        context.reset();
        git.reset();

        await fileTree.loadFileTree(selectedDir);
        await fileTree.startWatcher(selectedDir);
        await git.checkGitAvailability(selectedDir);

        notifications.addLog(`Выбрана новая директория проекта: ${selectedDir}`, 'success');
      }
    } catch (err) {
      notifications.addLog(`Ошибка выбора директории: ${err.message || err}`, 'error');
    }
  }

  return {
    projectRoot,
    selectProjectFolder,
  };
});