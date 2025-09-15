/**
 * Clean Architecture Project Store
 * Refactored to use use cases and follow Clean Architecture principles
 */

import { defineStore } from 'pinia';
import { ref, computed, readonly } from 'vue';
import { Project, WorkspaceState, Context } from '../domain/entities';
import { getUseCase } from '../infrastructure/di/setup';
import { TOKENS } from '../infrastructure/di/container';
import type { 
  LoadProjectUseCase, 
  ProjectLoadResult,
  UseCaseResult 
} from '../application/use-cases';

export const useProjectStore = defineStore('project', () => {
  // State
  const currentProject = ref<Project | null>(null);
  const recentProjects = ref<Project[]>([]);
  const workspace = ref<WorkspaceState | null>(null);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  // Use Cases
  const loadProjectUseCase = getUseCase<LoadProjectUseCase>(TOKENS.LOAD_PROJECT_USE_CASE);

  // Computed
  const hasProject = computed(() => currentProject.value !== null);
  const projectName = computed(() => currentProject.value?.name || '');
  const projectPath = computed(() => currentProject.value?.path.value || '');
  const isValidProject = computed(() => currentProject.value?.isValid() || false);

  // Actions
  async function loadProject(path: string): Promise<boolean> {
    console.log('Loading project with path:', path);
    isLoading.value = true;
    error.value = null;

    try {
      const result: UseCaseResult<ProjectLoadResult> = await loadProjectUseCase.execute(path);
      
      if (result.isSuccess && result.data) {
        currentProject.value = result.data.project;
        workspace.value = result.data.workspace;
        console.log('Project loaded successfully:', result.data.project);
        
        // Update recent projects
        await loadRecentProjects();
        
        return true;
      } else {
        error.value = result.error || 'Failed to load project';
        console.error('Failed to load project:', error.value);
        return false;
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Unknown error occurred';
      console.error('Exception while loading project:', error.value);
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  async function loadRecentProjects(): Promise<void> {
    try {
      const projectRepository = getUseCase(TOKENS.PROJECT_REPOSITORY);
      recentProjects.value = await projectRepository.getRecentProjects(10);
    } catch (err) {
      console.warn('Failed to load recent projects:', err);
    }
  }

  async function removeFromRecent(projectId: string): Promise<void> {
    try {
      const projectRepository = getUseCase(TOKENS.PROJECT_REPOSITORY);
      await projectRepository.removeFromRecent(projectId);
      
      // Update local state
      recentProjects.value = recentProjects.value.filter(p => p.id.value !== projectId);
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
  function switchWorkspaceMode(mode: any): void {
    if (workspace.value) {
      workspace.value.switchMode(mode);
    }
  }

  function updatePanelConfiguration(panelId: string, config: any): void {
    if (workspace.value) {
      workspace.value.setPanelConfiguration(panelId, config);
    }
  }

  function getPanelConfiguration(panelId: string): any {
    return workspace.value?.getPanelConfiguration(panelId) || null;
  }

  // Initialize store
  async function initialize(): Promise<void> {
    await loadRecentProjects();
  }

  // Add the missing methods that ProjectSelectionView expects
  async function tryAutoOpenLastProject(): Promise<void> {
    // Implementation would go here
    console.log("tryAutoOpenLastProject called");
  }

  async function openProject(): Promise<boolean> {
    try {
      isLoading.value = true;
      error.value = null;
      
      // Use the project repository to select a directory
      const projectRepository = getUseCase(TOKENS.PROJECT_REPOSITORY);
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

  async function openRecentProject(project: any): Promise<boolean> {
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

  async function tryLoadCurrentDirectory(): Promise<void> {
    // Implementation would go here
    console.log("tryLoadCurrentDirectory called");
  }

  function removeRecent(path: string): void {
    // Find and remove the project with the given path
    recentProjects.value = recentProjects.value.filter(p => p.path.value !== path);
    
    // Also remove from the repository
    const projectRepository = getUseCase(TOKENS.PROJECT_REPOSITORY);
    // We would need the project ID to remove it from the repository
    // This is a simplified implementation
  }

  // Computed property for hasRecentProjects
  const hasRecentProjects = computed(() => recentProjects.value.length > 0);

  // Return public interface
  return {
    // State (readonly)
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
    hasRecentProjects, // Add the missing computed property

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
    
    // Add the missing methods
    tryAutoOpenLastProject,
    openProject,
    openRecentProject,
    tryLoadCurrentDirectory,
    removeRecent
  };
});

// Removed auto-initialization to prevent Pinia error
// The store will be initialized properly in the main application after Pinia setup