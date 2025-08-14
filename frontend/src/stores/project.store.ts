import { defineStore } from "pinia";
import { useNotificationsStore } from "./notifications.store";
import { useErrorHandler } from "@/composables/useErrorHandler";
import type { Project } from "@/types/dto";
import { apiService } from "@/services/api.service";
import { useContextStore } from "./context.store";

export const useProjectStore = defineStore("project", {
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
      const contextStore = useContextStore();
      this.isLoading = true;

      try {
        notifications.addLog(`Opening project: ${project.name}...`, "info");

        // Останавливаем прошлый watcher, чтобы не было гонок
        try {
          await apiService.stopFileWatcher();
        } catch (e) {
          /* ignore */
        }

        // Обновляем список недавних
        const existingIndex = this.recentProjects.findIndex(
          (p) => p.path === project.path,
        );
        if (existingIndex > -1) this.recentProjects.splice(existingIndex, 1);
        this.recentProjects.unshift(project);

        // Сбрасываем состояние прошлого проекта
        contextStore.clearProjectData();

        // Устанавливаем текущий
        this.currentProject = project;

        // Стартуем watcher для нового пути
        try {
          await apiService.startFileWatcher(project.path);
        } catch (e) {
          handleError(e, "Start File Watcher");
        }
      } catch (err) {
        handleError(err, "Project Selection");
      } finally {
        this.isLoading = false;
      }
    },

    async selectProjectFromDialog() {
      const { handleError } = useErrorHandler();
      try {
        const selectedDir = await apiService.selectDirectory();
        if (selectedDir) {
          const projectName =
            selectedDir.split(/[\\/]/).pop() || "Unnamed Project";
          const newProject: Project = {
            id: Date.now().toString(),
            name: projectName,
            path: selectedDir,
          };
          await this.selectProject(newProject);
        }
      } catch (err) {
        handleError(err, "Directory Dialog");
      }
    },

    async clearProject() {
      const { handleError } = useErrorHandler();
      const notifications = useNotificationsStore();
      const contextStore = useContextStore();
      try {
        await apiService.stopFileWatcher();
      } catch (e) {
        handleError(e, "Stop File Watcher");
      }
      contextStore.clearProjectData();
      this.currentProject = null;
      notifications.addLog("Project closed.", "info");
    },
  },
  persist: {
    paths: ["recentProjects"],
  },
});

export default useProjectStore;
