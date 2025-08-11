import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { SelectDirectory } from '../../wailsjs/go/main/App';
import { useNotificationsStore } from './notificationsStore';
import type { Project, Task } from '@/types/dto';
import { useContextStore } from './contextStore';

export const useProjectStore = defineStore('project', () => {
  const notifications = useNotificationsStore();
  const router = useRouter();

  const currentProject = ref<Project | null>(null);
  const recentProjects = ref<Project[]>([]);
  const recentTasks = ref<Task[]>([]);
  const isLoading = ref(false);

  const isProjectLoaded = computed(() => !!currentProject.value);

  async function selectProject(project: Project) {
    const contextStore = useContextStore();
    isLoading.value = true;
    notifications.addLog(`Открытие проекта: ${project.name}...`, 'info');

    currentProject.value = project;

    const existingIndex = recentProjects.value.findIndex(p => p.path === project.path);
    if (existingIndex > -1) {
      recentProjects.value.splice(existingIndex, 1);
    }
    recentProjects.value.unshift(project);

    // Clear old project data before navigating
    contextStore.nodesMap = new Map();

    await router.push('/workspace');

    isLoading.value = false;
  }

  async function selectProjectFromDialog() {
    try {
      const selectedDir = await SelectDirectory();
      if (selectedDir) {
        const projectName = selectedDir.split(/[\\/]/).pop() || 'Unnamed Project';
        const newProject: Project = {
          id: Date.now().toString(),
          name: projectName,
          path: selectedDir,
          gitStatus: 'unknown',
        };
        await selectProject(newProject);
      }
    } catch (err: any) {
      notifications.addLog(`Ошибка выбора директории: ${err.message || err}`, 'error');
    }
  }

  function selectMostRecentProject() {
    if(recentProjects.value.length > 0) {
      selectProject(recentProjects.value[0]);
    }
  }

  function loadTask(taskId: string) {
    const task = recentTasks.value.find(t => t.id === taskId);
    if (task) {
      notifications.addLog(`Загрузка задачи "${task.name}"...`, 'info');
      router.push('/workspace');
    }
  }

  return {
    currentProject, recentProjects, recentTasks, isLoading, isProjectLoaded,
    selectProject, selectProjectFromDialog, selectMostRecentProject, loadTask,
  };
}, { persist: { paths: ['currentProject', 'recentProjects'] } });