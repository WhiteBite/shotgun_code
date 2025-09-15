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
  SelectDirectory 
} from '../../../wailsjs/go/main/App';

/**
 * Project API Adapter - Infrastructure implementation of ProjectRepository
 * This handles project and file operations while conforming to Clean Architecture
 */
export class ProjectApiAdapter implements ProjectRepository {
  private fileWatcherCallback: FileChangeCallback | null = null;

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
      // TODO: Implement getCurrentDirectory in backend if needed
      // For now, return empty string to maintain compatibility
      return "";
    } catch (error) {
      throw this.handleError(error, 'Failed to get current directory');
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

  async getFileStats(rootPath: string, relativePath: string): Promise<{
    size: number;
    modified: string;
    isDirectory: boolean;
    isFile: boolean;
  }> {
    try {
      // Load file tree to get file stats
      const files = await this.loadFileTree(rootPath);
      const file = this.findFileInTree(files, relativePath);
      
      if (!file) {
        throw new Error('File not found');
      }
      
      return {
        size: file.size || 0,
        modified: file.modTime || new Date().toISOString(),
        isDirectory: file.isDir || false,
        isFile: !file.isDir
      };
    } catch (error) {
      throw this.handleError(error, `Failed to get file stats: ${relativePath}`);
    }
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