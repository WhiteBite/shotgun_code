import type { DomainFileNode } from '@/types/dto';

/**
 * File tree options for loading project files
 */
export interface FileTreeOptions {
  useGitignore?: boolean;
  useCustomIgnore?: boolean;
  includeHidden?: boolean;
  maxDepth?: number;
}

/**
 * File change event for watching
 */
export interface FileChangeEvent {
  type: 'created' | 'modified' | 'deleted' | 'renamed';
  path: string;
  oldPath?: string; // For rename events
  timestamp: string;
}

export type FileChangeCallback = (event: FileChangeEvent) => void;

/**
 * Repository interface for project and file operations
 */
export interface ProjectRepository {
  /**
   * Validate if a path is a valid project directory
   * @param path Directory path to validate
   * @returns True if valid project path
   */
  validatePath(path: string): Promise<boolean>;

  /**
   * Load file tree for a project
   * @param rootPath Root project path
   * @param options File tree loading options
   * @returns File tree nodes
   */
  loadFileTree(rootPath: string, options?: FileTreeOptions): Promise<DomainFileNode[]>;

  /**
   * Read content of a specific file
   * @param rootPath Root project path
   * @param relativePath Relative path to file
   * @returns File content as string
   */
  readFileContent(rootPath: string, relativePath: string): Promise<string>;

  /**
   * Start watching for file changes
   * @param rootPath Root project path
   * @param callback Callback for file change events
   * @returns Promise that resolves when watcher is started
   */
  watchFileChanges(rootPath: string, callback: FileChangeCallback): Promise<void>;

  /**
   * Stop watching for file changes
   * @returns Promise that resolves when watcher is stopped
   */
  stopFileWatcher(): Promise<void>;

  /**
   * Select a directory using system dialog
   * @returns Selected directory path
   */
  selectDirectory(): Promise<string>;

  /**
   * Get current working directory
   * @returns Current working directory path
   */
  getCurrentDirectory(): Promise<string>;

  /**
   * Check if a file exists
   * @param rootPath Root project path
   * @param relativePath Relative path to file
   * @returns True if file exists
   */
  fileExists(rootPath: string, relativePath: string): Promise<boolean>;

  /**
   * Get file statistics
   * @param rootPath Root project path
   * @param relativePath Relative path to file
   * @returns File stats (size, modified date, etc.)
   */
  getFileStats(rootPath: string, relativePath: string): Promise<{
    size: number;
    modified: string;
    isDirectory: boolean;
    isFile: boolean;
  }>;
}