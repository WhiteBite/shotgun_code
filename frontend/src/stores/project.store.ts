import { defineStore } from "pinia";
import { apiService } from "@/services/api.service";
import { useFileTreeStore } from "./file-tree.store";

interface Project {
  name: string;
  path: string;
}

export const useProjectStore = defineStore("project", {
  state: () => ({
    currentProject: null as Project | null,
    recentProjects: [] as Project[],
    isLoading: false,
  }),
  getters: {
    isProjectLoaded: (state): boolean => !!state.currentProject,
  },
  actions: {
    async openProject(): Promise<boolean> {
      const selected = await apiService.selectDirectory();
      if (!selected) return false;
      const name =
          selected.replace(/[\\/]+$/, "").split(/[\\/]/).pop() || "Untitled";
      return await this.setCurrentProject({ name, path: selected });
    },

    async setCurrentProject(project: Project): Promise<boolean> {
      this.isLoading = true;
      try {
        const fileTreeStore = useFileTreeStore();
        try {
          await apiService.stopFileWatcher();
        } catch (e) {
          console.warn("Could not stop previous watcher", e);
        }
        fileTreeStore.clearProjectData();

        fileTreeStore.rootPath = project.path;
        this.currentProject = project;

        try {
          await apiService.startFileWatcher(project.path);
        } catch (e) {
          console.error("Failed to start file watcher", e);
        }
        await fileTreeStore.fetchFileTree();

        this.recentProjects = this.recentProjects.filter(
            (p) => p.path !== project.path,
        );
        this.recentProjects.unshift(project);
        if (this.recentProjects.length > 10) this.recentProjects.pop();
        try {
          localStorage.setItem(
              "recentProjects",
              JSON.stringify(this.recentProjects),
          );
        } catch (e) {
          console.warn("Failed to save recent projects to localStorage", e);
        }
        return true;
      } finally {
        this.isLoading = false;
      }
    },

    removeRecent(path: string) {
      this.recentProjects = this.recentProjects.filter(p => p.path !== path);
      try {
        localStorage.setItem("recentProjects", JSON.stringify(this.recentProjects));
      } catch {}
    },

    loadRecentProjects() {
      const raw = localStorage.getItem("recentProjects");
      if (raw) {
        try {
          this.recentProjects = JSON.parse(raw) as Project[];
        } catch (e) {
          console.warn("Failed to parse recent projects from localStorage", e);
        }
      }
    },
  },
});