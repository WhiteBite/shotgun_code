/**
 * Clean Architecture Project Store
 * Refactored to use use cases and follow Clean Architecture principles
 */

import { defineStore } from 'pinia';
import { ref, computed, readonly } from 'vue';
import { Project } from '../domain/entities';
import { container } from '../infrastructure/container';
import { createStoreWithDependencies, type StoreDependencies } from './StoreDependencyContainer';
import type { 
  ProjectLoadResult,
  UseCaseResult 
} from '../application/use-cases';
import { RecentProjectsService } from '@/domain/services/RecentProjectsService';

export const useProjectStore = defineStore('project', () => {
  return createStoreWithDependencies('project', (dependencies: StoreDependencies) => {
    // State
    interface Workspace {
      switchMode: (mode: string) => void;
      setPanelConfiguration: (panelId: string, config: unknown) => void;
      getPanelConfiguration: (panelId: string) => unknown;
    }

    const currentProject = ref<Project | null>(null);
    const recentProjects = ref<Project[]>([]);
    const workspace = ref<Workspace | null>(null);
    const isLoading = ref(false);
    const error = ref<string | null>(null);

    // Services
    const recentProjectsService = new RecentProjectsService(dependencies.projectRepository);

    // Use Cases - using new container system
    const getLoadProjectUseCase = () => container.getLoadProjectUseCase();
    
    // Inject ProjectRepository through dependency injection
    const { projectRepository } = dependencies;

    // Computed
    const hasProject = computed(() => currentProject.value !== null);
    const projectName = computed(() => currentProject.value?.name || '');
    const projectPath = computed(() => currentProject.value?.path || '');
    const isValidProject = computed(() => currentProject.value !== null);

    // Actions
    async function loadProject(path: string): Promise<boolean> {
      console.log('🔄 ProjectStore.loadProject called with path:', path);
      isLoading.value = true;
      error.value = null;

      try {
        console.log('🔧 Getting LoadProjectUseCase from container...');
        const loadProjectUseCase = getLoadProjectUseCase();
        console.log('✅ LoadProjectUseCase obtained');
        
        console.log('🚀 Executing LoadProjectUseCase...');
        const result: UseCaseResult<ProjectLoadResult> = await loadProjectUseCase.execute(path);
        console.log('📊 LoadProjectUseCase result:', result);
        
        if (result.isSuccess && result.data) {
          console.log('✅ Project loading successful, updating store state...');
          currentProject.value = result.data.project;
          workspace.value = result.data.workspace;
          
          console.log('✅ Store state updated. Current project:', currentProject.value?.name);
          
          // Add to recent projects using the new service
          try {
            await recentProjectsService.add({
              id: currentProject.value.id || crypto.randomUUID(),
              name: currentProject.value.name,
              path: currentProject.value.path
            });
            console.log('✅ Project added to recent list');
          } catch (err) {
            console.warn('⚠️  Failed to add to recent projects:', err);
          }
          
          // Load file tree after project is loaded
          console.log('🔄 Loading file tree...');
          const { useFileTreeStore } = await import('./file-tree.store');
          const fileTreeStore = useFileTreeStore();
          await fileTreeStore.loadProject(currentProject.value.path);
          console.log('✅ File tree loaded');
          
          // Update recent projects
          console.log('🔄 Refreshing recent projects...');
          await loadRecentProjects();
          
          console.log('🎉 Project loading completed successfully!');
          return true;
        } else {
          const errorMsg = result.error || 'Failed to load project';
          console.error('❌ Project loading failed:', errorMsg);
          error.value = errorMsg;
          return false;
        }
      } catch (err) {
        const errorMsg = err instanceof Error ? err.message : 'Unknown error occurred';
        console.error('❌ Project loading error:', errorMsg, err);
        error.value = errorMsg;
        return false;
      } finally {
        isLoading.value = false;
        console.log('🏁 Project loading operation completed. Loading state:', isLoading.value);
      }
    }

    async function loadRecentProjects(): Promise<void> {
      try {
        console.log('📂 Loading recent projects...');
        // Use the new service to load recent projects
        const projects = await recentProjectsService.load();
        recentProjects.value = projects.map(project => 
          Project.create({
            id: project.id,
            name: project.name,
            path: project.path
          })
        );
        console.log(`✅ Loaded ${projects.length} recent projects`);
      } catch (err) {
        console.warn('Failed to load recent projects:', err);
        recentProjects.value = [];
      }
    }

    async function removeFromRecent(projectId: string): Promise<void> {
      try {
        // Use the new service to remove from recent projects
        await recentProjectsService.remove(projectId);
        
        // Update local state
        recentProjects.value = recentProjects.value.filter(p => p.id !== projectId);
      } catch (err) {
        console.warn('Failed to remove project from recent:', err);
      }
    }

    function clearProject(): void {
      currentProject.value = null;
      workspace.value = null;
      error.value = null;
    }

    function clearError(): void {
      error.value = null;
    }

    // Workspace Management
    function switchWorkspaceMode(mode: string): void {
      if (workspace.value) {
        workspace.value.switchMode(mode);
      }
    }

    function updatePanelConfiguration(panelId: string, config: unknown): void {
      if (workspace.value) {
        workspace.value.setPanelConfiguration(panelId, config);
      }
    }

    function getPanelConfiguration(panelId: string): unknown {
      return workspace.value?.getPanelConfiguration(panelId) || null;
    }

    // Initialize store
    async function initialize(): Promise<void> {
      await loadRecentProjects();
    }

    // Add the missing methods that ProjectSelectionView expects
    async function tryAutoOpenLastProject(): Promise<void> {
      console.log("tryAutoOpenLastProject called");
      
      // In development mode, automatically load the current directory
      if (process.env.NODE_ENV === 'development') {
        try {
          console.log("🔧 Development mode: Attempting to auto-load current directory...");
          const projectRepository = container.projectRepository;
          const currentDir = await projectRepository.getCurrentDirectory();
          console.log(`📂 Current directory: ${currentDir}`);
          
          if (currentDir) {
            console.log("🚀 Auto-loading current directory as project...");
            await loadProject(currentDir);
          }
        } catch (error) {
          console.warn("⚠️  Failed to auto-load current directory:", error);
        }
      }
    }

    async function openProject(): Promise<boolean> {
      try {
        isLoading.value = true;
        error.value = null;
        
        // Use the project repository to select a directory
        const projectRepository = container.projectRepository;
        const selectedPath = await projectRepository.selectDirectory();
        
        if (selectedPath) {
          const success = await loadProject(selectedPath);
          return success;
        }
        
        return false;
      } catch (err) {
        error.value = err instanceof Error ? err.message : 'Failed to open project';
        return false;
      } finally {
        isLoading.value = false;
      }
    }

    interface RecentProject {
      path: string;
      [key: string]: unknown;
    }

    async function openRecentProject(project: RecentProject): Promise<boolean> {
      try {
        isLoading.value = true;
        error.value = null;
        
        const success = await loadProject(project.path);
        return success;
      } catch (err) {
        error.value = err instanceof Error ? err.message : 'Failed to open recent project';
        return false;
      } finally {
        isLoading.value = false;
      }
    }
    
    /**
     * Development method to manually set a project path
     * This bypasses the directory selection dialog for testing
     */
    async function setProjectPath(path: string): Promise<boolean> {
      try {
        console.log(`🔧 Manually setting project path to: ${path}`);
        isLoading.value = true;
        error.value = null;
        
        const success = await loadProject(path);
        if (success) {
          console.log("✅ Project loaded successfully");
        } else {
          console.error("❌ Failed to load project");
        }
        return success;
      } catch (err) {
        const errorMsg = err instanceof Error ? err.message : 'Failed to set project path';
        console.error("❌ Error setting project path:", errorMsg);
        error.value = errorMsg;
        return false;
      } finally {
        isLoading.value = false;
      }
    }
    
    function removeRecent(path: string): void {
      // Implementation would go here
      console.log("removeRecent called with path:", path);
    }

    // Return store methods
    return {
      // State
      currentProject: readonly(currentProject),
      recentProjects: readonly(recentProjects),
      workspace: readonly(workspace),
      isLoading: readonly(isLoading),
      error: readonly(error),

      // Computed
      hasProject,
      projectName,
      projectPath,
      isValidProject,
      hasRecentProjects: computed(() => recentProjects.value.length > 0),

      // Actions
      loadProject,
      loadRecentProjects,
      removeFromRecent,
      clearProject,
      clearError,
      switchWorkspaceMode,
      updatePanelConfiguration,
      getPanelConfiguration,
      initialize,
      tryAutoOpenLastProject,
      openProject,
      openRecentProject,
      removeRecent,
      setProjectPath // Add this new method
    };
  }); // This closes the createStoreWithDependencies call
});
// Removed auto-initialization to prevent Pinia error
// The store will be initialized properly in the main application after Pinia setup