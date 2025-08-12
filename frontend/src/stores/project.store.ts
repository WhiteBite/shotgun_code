import { defineStore } from 'pinia';
// useRouter is no longer needed here.
import { useNotificationsStore } from './notifications.store';
import { useErrorHandler } from '@/composables/useErrorHandler';
import type { Project } from '@/types/dto';
import { apiService } from '@/services/api.service';

export const useProjectStore = defineStore('project', {
  state: () => ({
    currentProject: null as Project | null,
    recentProjects: [] as Project[],
    isLoading: false,
  }),
  getters: {
    isProjectLoaded: (state) => !!state.currentProject,
  },
  actions: {
    async selectProject(project: Project) {
      const notifications = useNotificationsStore();
      const { handleError } = useErrorHandler();
      this.isLoading = true;

      try {
        notifications.addLog(`Opening project: ${project.name}...`, 'info');

        const existingIndex = this.recentProjects.findIndex(p => p.path === project.path);
        if (existingIndex > -1) this.recentProjects.splice(existingIndex, 1);
        this.recentProjects.unshift(project);

        // The store's only job is to set the state.
        this.currentProject = project;
        // Navigation will be handled by a watcher in App.vue.

      } catch (err) {
        handleError(err, 'Project Selection');
      } finally {
        this.isLoading = false;
      }
    },

    async selectProjectFromDialog() {
      const { handleError } = useErrorHandler();
      try {
        const selectedDir = await apiService.selectDirectory();
        if (selectedDir) {
          const projectName = selectedDir.split(/[\\/]/).pop() || 'Unnamed Project';
          const newProject: Project = {
            id: Date.now().toString(),
            name: projectName,
            path: selectedDir,
          };
          await this.selectProject(newProject);
        }
      } catch (err) {
        handleError(err, 'Directory Dialog');
      }
    },
  },
  persist: {
    paths: ['recentProjects'],
  },
});