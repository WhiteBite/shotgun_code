import type { 
  ProjectRepository, 
  FileTreeOptions, 
  FileChangeCallback 
} from '@/domain/repositories/ProjectRepository';
import type { DomainFileNode } from '@/types/dto';
import { 
  ListFiles, 
  ReadFileContent, 
  StartFileWatcher, 
  StopFileWatcher, 
  SelectDirectory,
  GetCurrentDirectory,
  GetFileStats
} from '../../../wailsjs/go/main/App';
import { defaultWailsApiAdapter, type WailsApiAdapter } from '../api/WailsApiAdapter';

/**
 * Project API Adapter - Infrastructure implementation of ProjectRepository
 * This handles project and file operations while conforming to Clean Architecture
 */
export class ProjectApiAdapter implements ProjectRepository {
  private readonly apiAdapter: WailsApiAdapter;
  private fileWatcherCallback: FileChangeCallback | null = null;

  constructor(apiAdapter: WailsApiAdapter = defaultWailsApiAdapter) {
    this.apiAdapter = apiAdapter;
  }

  async validatePath(path: string): Promise<boolean> {
    try {
      // Try to list files to verify path exists and is accessible
      await ListFiles(path, true, true);
      return true;
    } catch (error) {
      console.warn('Path validation failed:', error);
      return false;
    }
  }

  async loadFileTree(rootPath: string, options?: FileTreeOptions): Promise<DomainFileNode[]> {
    try {
      const useGitignore = options?.useGitignore ?? true;
      const useCustomIgnore = options?.useCustomIgnore ?? true;
      
      const files = await ListFiles(rootPath, useGitignore, useCustomIgnore);
      return files;
    } catch (error) {
      throw this.handleError(error, 'Failed to load file tree');
    }
  }

  async readFileContent(rootPath: string, relativePath: string): Promise<string> {
    try {
      return await ReadFileContent(rootPath, relativePath);
    } catch (error) {
      throw this.handleError(error, `Failed to read file content: ${relativePath}`);
    }
  }

  async watchFileChanges(rootPath: string, callback: FileChangeCallback): Promise<void> {
    try {
      this.fileWatcherCallback = callback;
      await StartFileWatcher(rootPath);
      
      // Note: In a real implementation, we would need to set up event listeners
      // for file change events from the backend. For now, this establishes
      // the watcher on the backend side.
    } catch (error) {
      throw this.handleError(error, 'Failed to start file watcher');
    }
  }

  async stopFileWatcher(): Promise<void> {
    try {
      await StopFileWatcher();
      this.fileWatcherCallback = null;
    } catch (error) {
      throw this.handleError(error, 'Failed to stop file watcher');
    }
  }

  async selectDirectory(): Promise<string> {
    try {
      return await SelectDirectory();
    } catch (error) {
      throw this.handleError(error, 'Failed to select directory');
    }
  }

  async getCurrentDirectory(): Promise<string> {
    try {
      const result = await this.apiAdapter.callApi<void, string>(GetCurrentDirectory);
      console.log("Current directory from backend:", result);
      return result;
    } catch (error) {
      console.error("Failed to get current directory:", error);
      // Fallback to a default directory in development mode
      if (process.env.NODE_ENV === 'development') {
        return '.'; // Current directory as fallback
      }
      throw error;
    }
  }

  async fileExists(rootPath: string, relativePath: string): Promise<boolean> {
    try {
      await this.readFileContent(rootPath, relativePath);
      return true;
    } catch (error) {
      return false;
    }
  }

  async getRecentProjects(): Promise<Array<{id: string; name: string; path: string; lastOpened: string}>> {
    try {
      // Use localStorage to store and retrieve recent projects
      const storedProjects = localStorage.getItem('shotgun-recent-projects');
      if (storedProjects) {
        const projects = JSON.parse(storedProjects);
        // Ensure projects have all required fields and sort by lastOpened
        const validProjects = projects
          .filter((p: any) => p.id && p.name && p.path && p.lastOpened)
          .sort((a: any, b: any) => new Date(b.lastOpened).getTime() - new Date(a.lastOpened).getTime())
          .slice(0, 10); // Limit to 10 most recent
          
        console.log('📂 Retrieved recent projects from localStorage:', validProjects.length, 'projects');
        return validProjects;
      }
      
      // If no projects exist, add some test data for demonstration
      console.log('📂 No recent projects found, creating demo data...');
      const demoProjects = [
        {
          id: 'demo-1',
          name: 'Example Project 1',
          path: 'C:\\Projects\\ExampleProject1',
          lastOpened: new Date(Date.now() - 86400000).toISOString() // Yesterday
        },
        {
          id: 'demo-2',
          name: 'Example Project 2', 
          path: 'C:\\Projects\\ExampleProject2',
          lastOpened: new Date(Date.now() - 172800000).toISOString() // 2 days ago
        }
      ];
      
      // Save demo projects
      localStorage.setItem('shotgun-recent-projects', JSON.stringify(demoProjects));
      console.log('📂 Created demo recent projects:', demoProjects.length, 'projects');
      return demoProjects;
    } catch (error) {
      console.warn('Failed to load recent projects from localStorage:', error);
      return [];
    }
  }

  async addToRecentProjects(project: {id: string; name: string; path: string}): Promise<void> {
    try {
      const existing = await this.getRecentProjects();
      
      // Remove existing entry for this path to avoid duplicates
      const filtered = existing.filter(p => p.path !== project.path);
      
      // Add new entry at the beginning
      const updated = [{
        ...project,
        lastOpened: new Date().toISOString()
      }, ...filtered].slice(0, 10); // Keep only 10 most recent
      
      localStorage.setItem('shotgun-recent-projects', JSON.stringify(updated));
    } catch (error) {
      console.warn('Failed to add project to recent list:', error);
    }
  }

  async removeFromRecentProjects(projectId: string): Promise<void> {
    try {
      const existing = await this.getRecentProjects();
      const filtered = existing.filter(p => p.id !== projectId);
      localStorage.setItem('shotgun-recent-projects', JSON.stringify(filtered));
    } catch (error) {
      console.warn('Failed to remove project from recent list:', error);
    }
  }

  async getFileStats(rootPath: string, relativePath: string): Promise<{
    size: number;
    modified: string;
    isDirectory: boolean;
    isFile: boolean;
  }> {
    const fullPath = `${rootPath}/${relativePath}`.replace(/\/+/g, '/');
    const stats = await this.apiAdapter.callApi<string, {
      size: number;
      modTime: number;
      isDir: boolean;
      name: string;
      mode: string;
    }>(GetFileStats, fullPath);
    
    return {
      size: stats.size,
      modified: new Date(stats.modTime * 1000).toISOString(),
      isDirectory: stats.isDir,
      isFile: !stats.isDir
    };
  }

  // Private helper methods
  private findFileInTree(files: DomainFileNode[], targetPath: string): DomainFileNode | null {
    for (const file of files) {
      if (file.relPath === targetPath) {
        return file;
      }
      if (file.children) {
        const found = this.findFileInTree(file.children, targetPath);
        if (found) return found;
      }
    }
    return null;
  }

  private handleError(error: unknown, context: string): Error {
    const message = error instanceof Error ? error.message : String(error);
    
    // Check if this is a domain error from backend
    if (message.startsWith('domain_error:')) {
      try {
        const domainErrorJson = message.substring('domain_error:'.length);
        const domainError = JSON.parse(domainErrorJson);
        
        // Create structured frontend error
        const structuredError = new Error(`${context}: ${domainError.message}`);
        (structuredError as any).code = domainError.code;
        (structuredError as any).recoverable = domainError.recoverable;
        (structuredError as any).context = domainError.context;
        (structuredError as any).cause = domainError.cause;
        
        return structuredError;
      } catch (parseErr) {
        console.error('Failed to parse domain error:', parseErr);
      }
    }
    
    return new Error(`${context}: ${message}`);
  }
}