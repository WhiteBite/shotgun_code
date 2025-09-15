/**
 * File System Repository Implementation
 * Concrete implementation using Wails API for file system operations
 */

import { FileNode } from '../../domain/entities';
import { FileSystemRepository, FileSystemChange, FileMetadata } from '../../domain/repositories';
import { apiService } from '../api/api.service';
import type { DomainFileNode } from '../../types/dto';

export class FileSystemRepositoryImpl implements FileSystemRepository {
  private watcherActive = false;
  private watcherCallback: ((changes: FileSystemChange[]) => void) | null = null;

  async validatePath(path: string): Promise<boolean> {
    try {
      if (!path || path.trim().length === 0) return false;
      
      // Use Wails API to check if path exists
      const files = await apiService.listFiles(path, false, false);
      return Array.isArray(files);
    } catch (error) {
      console.warn('Path validation failed:', error);
      return false;
    }
  }

  async loadFileTree(
    rootPath: string, 
    useGitignore = true, 
    useCustomIgnore = true
  ): Promise<FileNode[]> {
    try {
      const domainNodes = await apiService.listFiles(rootPath, useGitignore, useCustomIgnore);
      return this.transformDomainNodesToFileNodes(domainNodes);
    } catch (error) {
      throw new Error(`Failed to load file tree: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  async readFileContent(rootPath: string, relativePath: string): Promise<string> {
    try {
      return await apiService.readFileContent(rootPath, relativePath);
    } catch (error) {
      throw new Error(`Failed to read file content: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  async startWatching(
    rootPath: string, 
    callback: (changes: FileSystemChange[]) => void
  ): Promise<void> {
    try {
      if (this.watcherActive) {
        await this.stopWatching();
      }

      this.watcherCallback = callback;
      this.watcherActive = true;

      // Start file watcher via Wails API
      await apiService.startFileWatcher(rootPath);

      // Note: In a real implementation, we'd set up event listeners
      // for file system changes from the backend
    } catch (error) {
      throw new Error(`Failed to start file watcher: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  async stopWatching(): Promise<void> {
    try {
      if (!this.watcherActive) return;

      await apiService.stopFileWatcher();
      this.watcherActive = false;
      this.watcherCallback = null;
    } catch (error) {
      throw new Error(`Failed to stop file watcher: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  async getFileMetadata(path: string): Promise<FileMetadata> {
    try {
      // This is a simplified implementation
      // In practice, you'd make a specific API call to get file metadata
      const stats = await this.getFileStats(path);
      
      return {
        path,
        size: stats.size,
        modified: stats.modified,
        created: stats.created,
        isDirectory: stats.isDirectory,
        permissions: {
          readable: true,
          writable: true
        }
      };
    } catch (error) {
      throw new Error(`Failed to get file metadata: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private transformDomainNodesToFileNodes(domainNodes: DomainFileNode[]): FileNode[] {
    const nodeMap = new Map<string, FileNode>();
    const rootNodes: FileNode[] = [];

    // First pass: create all nodes
    for (const domainNode of domainNodes) {
      const fileNode = domainNode.isDir
        ? FileNode.createDirectory(
            domainNode.path,
            new Date(domainNode.modified || Date.now()),
            domainNode.isIgnored || domainNode.isGitignored || domainNode.isCustomIgnored
          )
        : FileNode.createFile(
            domainNode.path,
            domainNode.size || 0,
            new Date(domainNode.modified || Date.now()),
            domainNode.isIgnored || domainNode.isGitignored || domainNode.isCustomIgnored
          );

      nodeMap.set(domainNode.path, fileNode);

      // Determine if this is a root node
      if (!domainNode.parentPath || domainNode.parentPath === '') {
        rootNodes.push(fileNode);
      }
    }

    // Second pass: establish parent-child relationships
    for (const domainNode of domainNodes) {
      if (domainNode.parentPath) {
        const parent = nodeMap.get(domainNode.parentPath);
        const child = nodeMap.get(domainNode.path);
        
        if (parent && child) {
          parent.addChild(child);
        }
      }
    }

    return rootNodes;
  }

  private async getFileStats(path: string): Promise<{
    size: number;
    modified: Date;
    created: Date;
    isDirectory: boolean;
  }> {
    // In a real implementation, this would make an API call to get actual file stats
    // For now, we'll return mock data
    return {
      size: Math.floor(Math.random() * 10000),
      modified: new Date(),
      created: new Date(Date.now() - Math.floor(Math.random() * 10000000)),
      isDirectory: path.endsWith('/') || path.endsWith('\\')
    };
  }

  // Handle file system change events (would be called from backend)
  private handleFileSystemChange(changes: any[]) {
    if (!this.watcherCallback) return;

    const transformedChanges: FileSystemChange[] = changes.map(change => ({
      type: change.type,
      path: change.path,
      isDirectory: change.isDirectory
    }));

    this.watcherCallback(transformedChanges);
  }
}