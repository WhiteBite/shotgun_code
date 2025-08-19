import {
  ListFiles,
  ReadFileContent,
  RequestShotgunContextGeneration,
  GetUncommittedFiles,
  GetRichCommitHistory,
  GetFileContentAtCommit,
  GetGitignoreContent,
  IsGitAvailable,
  GenerateCode,
  SuggestContextFiles,
  GetSettings,
  SaveSettings,
  RefreshAIModels,
  StartFileWatcher,
  StopFileWatcher,
  SelectDirectory,
  ExportContext,
  GenerateIntelligentCode,
  GenerateCodeWithOptions,
  GetProviderInfo,
  ListAvailableModels,
  AnalyzeTaskAndCollectContext,
  TestBackend,
} from "../../wailsjs/go/main/App";

import type {
  DomainFileNode,
  FileStatus,
  CommitWithFiles,
  SettingsDTO,
} from "@/types/dto";

import type {
  Context,
  ExportSettings,
  ExportResult,
} from "@/types/api";

import { ApiError, ValidationError } from "@/types/api";

class ApiService {
  async listFiles(
      dirPath: string,
      useGitignore: boolean,
      useCustomIgnore: boolean,
  ): Promise<DomainFileNode[]> {
    try {
      return await ListFiles(dirPath, useGitignore, useCustomIgnore);
    } catch (error) {
      throw this.handleError(error, 'Failed to list files');
    }
  }

  async readFileContent(rootDir: string, relPath: string): Promise<string> {
    try {
      return await ReadFileContent(rootDir, relPath);
    } catch (error) {
      throw this.handleError(error, 'Failed to read file content');
    }
  }

  async requestShotgunContextGeneration(
      rootDir: string,
      includedPaths: string[],
  ): Promise<void> {
    try {
      await RequestShotgunContextGeneration(rootDir, includedPaths);
    } catch (error) {
      throw this.handleError(error, 'Failed to generate context');
    }
  }

  async getUncommittedFiles(projectRoot: string): Promise<FileStatus[]> {
    try {
      return await GetUncommittedFiles(projectRoot);
    } catch (error) {
      throw this.handleError(error, 'Failed to get uncommitted files');
    }
  }

  async getRichCommitHistory(
      projectRoot: string,
      branchName: string,
      limit: number,
  ): Promise<CommitWithFiles[]> {
    try {
      return await GetRichCommitHistory(projectRoot, branchName, limit);
    } catch (error) {
      throw this.handleError(error, 'Failed to get commit history');
    }
  }

  async getFileContentAtCommit(
      projectRoot: string,
      filePath: string,
      commitHash: string,
  ): Promise<string> {
    try {
      return await GetFileContentAtCommit(projectRoot, filePath, commitHash);
    } catch (error) {
      throw this.handleError(error, 'Failed to get file content at commit');
    }
  }

  async getGitignoreContent(projectRoot: string): Promise<string> {
    try {
      return await GetGitignoreContent(projectRoot);
    } catch (error) {
      throw this.handleError(error, 'Failed to get gitignore content');
    }
  }

  async isGitAvailable(): Promise<boolean> {
    try {
      return await IsGitAvailable();
    } catch (error) {
      throw this.handleError(error, 'Failed to check git availability');
    }
  }

  async generateCode(
      systemPrompt: string,
      userPrompt: string,
  ): Promise<string> {
    try {
      return await GenerateCode(systemPrompt, userPrompt);
    } catch (error) {
      throw this.handleError(error, 'Failed to generate code');
    }
  }

  async generateIntelligentCode(
      task: string,
      context: string,
      optionsJson: string,
  ): Promise<string> {
    try {
      return await GenerateIntelligentCode(task, context, optionsJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to generate intelligent code');
    }
  }

  async generateCodeWithOptions(
      systemPrompt: string,
      userPrompt: string,
      optionsJson: string,
  ): Promise<string> {
    try {
      return await GenerateCodeWithOptions(systemPrompt, userPrompt, optionsJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to generate code with options');
    }
  }

  async getProviderInfo(): Promise<string> {
    try {
      return await GetProviderInfo();
    } catch (error) {
      throw this.handleError(error, 'Failed to get provider info');
    }
  }

  async listAvailableModels(): Promise<string[]> {
    try {
      return await ListAvailableModels();
    } catch (error) {
      throw this.handleError(error, 'Failed to list available models');
    }
  }

  async analyzeTaskAndCollectContext(task: string, allFilesJson: string, rootDir: string): Promise<string> {
    try {
      return await AnalyzeTaskAndCollectContext(task, allFilesJson, rootDir);
    } catch (error) {
      throw this.handleError(error, 'Failed to analyze task and collect context');
    }
  }

  async testBackend(allFilesJson: string, rootDir: string): Promise<string> {
    try {
      return await TestBackend(allFilesJson, rootDir);
    } catch (error) {
      throw this.handleError(error, 'Failed to test backend');
    }
  }

  async suggestContextFiles(
      task: string,
      allFiles: any[], // Используем any для совместимости
  ): Promise<string[]> {
    try {
      return await SuggestContextFiles(task, allFiles);
    } catch (error) {
      throw this.handleError(error, 'Failed to suggest context files');
    }
  }

  async getSettings(): Promise<SettingsDTO> {
    try {
      return await GetSettings();
    } catch (error) {
      throw this.handleError(error, 'Failed to get settings');
    }
  }

  async saveSettings(dto: SettingsDTO): Promise<void> {
    try {
      return await SaveSettings(JSON.stringify(dto));
    } catch (error) {
      throw this.handleError(error, 'Failed to save settings');
    }
  }

  async refreshAIModels(provider: string, apiKey: string): Promise<void> {
    try {
      return await RefreshAIModels(provider, apiKey);
    } catch (error) {
      throw this.handleError(error, 'Failed to refresh AI models');
    }
  }

  async startFileWatcher(rootDirPath: string): Promise<void> {
    try {
      return await StartFileWatcher(rootDirPath);
    } catch (error) {
      throw this.handleError(error, 'Failed to start file watcher');
    }
  }

  async stopFileWatcher(): Promise<void> {
    try {
      return await StopFileWatcher();
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

  async exportContext(settings: ExportSettings): Promise<ExportResult> {
    try {
      const result = await ExportContext(JSON.stringify(settings));

      // Обработка больших файлов с FilePath
      if (result.filePath && result.isLarge) {
        console.log('Large file exported to:', result.filePath);
      }

      return result;
    } catch (error) {
      throw this.handleError(error, 'Failed to export context');
    }
  }

  async cleanupTempFiles(filePath: string): Promise<void> {
    try {
      // TODO: Add CleanupTempFiles method to backend
      console.log('Cleanup temp file:', filePath);
    } catch (error) {
      throw this.handleError(error, 'Failed to cleanup temp files');
    }
  }

  // New methods for context management
  async buildContext(
    projectPath: string, 
    includedPaths: string[], 
    options?: {
      stripComments?: boolean
      includeManifest?: boolean
      maxTokens?: number
    }
  ): Promise<Context> {
    try {
      // TODO: Add BuildContext method to backend
      // For now, return a mock context
      return {
        id: Date.now().toString(),
        name: `Context ${new Date().toLocaleString()}`,
        description: `Context with ${includedPaths.length} files`,
        content: includedPaths.join('\n'),
        files: includedPaths,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        projectPath,
        tokenCount: includedPaths.length * 100
      }
    } catch (error) {
      throw this.handleError(error, 'Failed to build context');
    }
  }

  async getContext(id: string): Promise<Context> {
    try {
      // TODO: Add GetContext method to backend
      throw new Error('GetContext not implemented');
    } catch (error) {
      throw this.handleError(error, 'Failed to get context');
    }
  }

  async getProjectContexts(projectPath: string): Promise<Context[]> {
    try {
      // TODO: Add GetProjectContexts method to backend
      return [];
    } catch (error) {
      throw this.handleError(error, 'Failed to get project contexts');
    }
  }

  async deleteContext(id: string): Promise<void> {
    try {
      // TODO: Add DeleteContext method to backend
      console.log('Delete context:', id);
    } catch (error) {
      throw this.handleError(error, 'Failed to delete context');
    }
  }

  async verifyProjectPath(path: string): Promise<boolean> {
    try {
      // Try to list files to verify path exists and is accessible
      await this.listFiles(path, true, true);
      return true;
    } catch (error) {
      return false;
    }
  }

  // Private methods
  private handleError(error: unknown, context: string): Error {
    if (error instanceof ApiError) {
      return error;
    }
    
    if (error instanceof ValidationError) {
      return error;
    }
    
    const message = error instanceof Error ? error.message : String(error);
    return new ApiError(`${context}: ${message}`);
  }
}

export const apiService = new ApiService();
