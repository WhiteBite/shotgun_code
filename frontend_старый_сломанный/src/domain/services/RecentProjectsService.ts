import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';

/**
 * Recent Projects Service
 * Encapsulates all logic for managing recent projects
 */
export class RecentProjectsService {
  constructor(private projectRepository: ProjectRepository) {}

  /**
   * Loads recent projects
   * @returns Promise<Array<{id: string; name: string; path: string; lastOpened: string}>> - The recent projects
   */
  async load(): Promise<Array<{id: string; name: string; path: string; lastOpened: string}>> {
    try {
      return await this.projectRepository.getRecentProjects();
    } catch (err) {
      console.warn('Failed to load recent projects:', err);
      return [];
    }
  }

  /**
   * Adds a project to recent projects
   * @param project - The project to add
   */
  async add(project: {id: string; name: string; path: string}): Promise<void> {
    try {
      await this.projectRepository.addToRecentProjects(project);
    } catch (err) {
      console.warn('Failed to add project to recent list:', err);
    }
  }

  /**
   * Removes a project from recent projects
   * @param projectId - The ID of the project to remove
   */
  async remove(projectId: string): Promise<void> {
    try {
      await this.projectRepository.removeFromRecentProjects(projectId);
    } catch (err) {
      console.warn('Failed to remove project from recent list:', err);
    }
  }

  /**
   * Clears all recent projects
   */
  async clear(): Promise<void> {
    try {
      // This would require a new method in ProjectRepository
      // For now, we'll just load and remove each one
      const recentProjects = await this.load();
      for (const project of recentProjects) {
        await this.remove(project.id);
      }
    } catch (err) {
      console.warn('Failed to clear recent projects:', err);
    }
  }
}