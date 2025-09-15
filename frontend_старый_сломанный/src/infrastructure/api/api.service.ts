/**
 * API Service
 * 
 * Centralized service for all backend API calls using Wails bridge.
 * This service provides a clean abstraction over the Wails-generated API
 * and handles error management, logging, and type safety.
 */

import type { DomainFileNode } from '@/types/dto';
import type { ExportSettings, ExportResult } from '@/types/export';
import { defaultWailsApiAdapter } from '@/infrastructure/api/WailsApiAdapter';

// Type definitions for API methods
export interface ApiService {
  listFiles: (rootPath: string, useGitignore: boolean, useCustomIgnore: boolean) => Promise<DomainFileNode[]>;
  readFileContent: (rootPath: string, relativePath: string) => Promise<string>;
  startFileWatcher: (rootPath: string) => Promise<void>;
  stopFileWatcher: () => Promise<void>;
  exportContext: (settings: ExportSettings) => Promise<ExportResult>;
  requestShotgunContextGeneration: (rootDir: string, includedPaths: string[]) => Promise<void>;
  getFileContentAtCommit: (rootDir: string, path: string, commitHash: string) => Promise<string>;
}

// Create the API service instance using WailsApiAdapter
class WailsApiService implements ApiService {
  async listFiles(rootPath: string, useGitignore: boolean, useCustomIgnore: boolean): Promise<DomainFileNode[]> {
    try {
      return await defaultWailsApiAdapter.callApi(
        window.go.main.App.ListFiles,
        [rootPath, useGitignore, useCustomIgnore]
      );
    } catch (error) {
      console.error('Failed to list files:', error);
      throw error;
    }
  }

  async readFileContent(rootPath: string, relativePath: string): Promise<string> {
    try {
      return await defaultWailsApiAdapter.callApi(
        window.go.main.App.ReadFileContent,
        [rootPath, relativePath]
      );
    } catch (error) {
      console.error('Failed to read file content:', error);
      throw error;
    }
  }

  async startFileWatcher(rootPath: string): Promise<void> {
    try {
      await defaultWailsApiAdapter.callApi(
        window.go.main.App.StartFileWatcher,
        [rootPath]
      );
    } catch (error) {
      console.error('Failed to start file watcher:', error);
      throw error;
    }
  }

  async stopFileWatcher(): Promise<void> {
    try {
      await defaultWailsApiAdapter.callApi(
        window.go.main.App.StopFileWatcher,
        []
      );
    } catch (error) {
      console.error('Failed to stop file watcher:', error);
      throw error;
    }
  }

  async exportContext(settings: ExportSettings): Promise<ExportResult> {
    try {
      return await defaultWailsApiAdapter.callApi(
        window.go.main.App.ExportContext,
        [settings]
      );
    } catch (error) {
      console.error('Failed to export context:', error);
      throw error;
    }
  }

  async requestShotgunContextGeneration(rootDir: string, includedPaths: string[]): Promise<void> {
    try {
      await defaultWailsApiAdapter.callApi(
        window.go.main.App.RequestShotgunContextGeneration,
        [rootDir, includedPaths]
      );
    } catch (error) {
      console.error('Failed to request Shotgun context generation:', error);
      throw error;
    }
  }

  async getFileContentAtCommit(rootDir: string, path: string, commitHash: string): Promise<string> {
    try {
      return await defaultWailsApiAdapter.callApi(
        window.go.main.App.GetFileContentAtCommit,
        [rootDir, path, commitHash]
      );
    } catch (error) {
      console.error('Failed to get file content at commit:', error);
      throw error;
    }
  }
}

// Export singleton instance
export const apiService: ApiService = new WailsApiService();

export default apiService;