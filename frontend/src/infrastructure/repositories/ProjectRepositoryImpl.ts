/**
 * Project Repository Implementation
 * Concrete implementation using local storage and Wails API
 */

import { Project } from '../../domain/entities';
import { ProjectPath } from '../../domain/value-objects';
import { ProjectRepository } from '../../domain/repositories';

export class ProjectRepositoryImpl implements ProjectRepository {
  private readonly STORAGE_KEY = 'shotgun_recent_projects';
  private readonly MAX_RECENT_PROJECTS = 10;

  async save(project: Project): Promise<void> {
    try {
      const recentProjects = await this.getRecentProjects();
      
      // Remove existing project if it exists
      const filtered = recentProjects.filter(p => !p.id.equals(project.id));
      
      // Add project to the beginning
      const updated = [project, ...filtered].slice(0, this.MAX_RECENT_PROJECTS);
      
      // Convert to serializable format
      const serializable = updated.map(p => ({
        id: p.id.value,
        name: p.name,
        path: p.path.value,
        lastOpened: p.lastOpened.toISOString()
      }));
      
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(serializable));
    } catch (error) {
      throw new Error(`Failed to save project: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  async findByPath(path: ProjectPath): Promise<Project | null> {
    try {
      const recentProjects = await this.getRecentProjects();
      return recentProjects.find(p => p.path.equals(path)) || null;
    } catch (error) {
      console.warn('Failed to find project by path:', error);
      return null;
    }
  }

  async getRecentProjects(limit?: number): Promise<Project[]> {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEY);
      if (!stored) return [];
      
      const data = JSON.parse(stored);
      const projects = data.map((item: any) => {
        const project = Project.create(item.path);
        if (item.lastOpened) {
          project.updateLastOpened();
        }
        return project;
      });
      
      return limit ? projects.slice(0, limit) : projects;
    } catch (error) {
      console.warn('Failed to load recent projects:', error);
      return [];
    }
  }

  async removeFromRecent(projectId: string): Promise<void> {
    try {
      const recentProjects = await this.getRecentProjects();
      const filtered = recentProjects.filter(p => p.id.value !== projectId);
      
      const serializable = filtered.map(p => ({
        id: p.id.value,
        name: p.name,
        path: p.path.value,
        lastOpened: p.lastOpened.toISOString()
      }));
      
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(serializable));
    } catch (error) {
      throw new Error(`Failed to remove project from recent: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  async exists(path: ProjectPath): Promise<boolean> {
    try {
      // This would typically check the file system
      // For now, we'll use a simple validation
      return path.value.length > 0 && path.value !== '/';
    } catch (error) {
      return false;
    }
  }
}