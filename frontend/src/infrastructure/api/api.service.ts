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
  // Autonomous task methods
  StartAutonomousTask,
  CancelAutonomousTask,
  GetAutonomousTaskStatus,
  // Context management methods
  BuildContext,
  GetContext,
  GetProjectContexts,
  DeleteContext,
  // Streaming context methods
  GetContextLines,
  CreateStreamingContext,
  GetStreamingContext,
  CloseStreamingContext,
  // Report methods
  GetReport,
  ListReports,
  // SLA policy method
  SetSLAPolicy,
} from "../../../wailsjs/go/main/App";

import type {
  DomainFileNode,
  FileStatus,
  CommitWithFiles,
  SettingsDTO,
  AutonomousTaskRequest,
  AutonomousTaskResponse,
  AutonomousTaskStatus,
  GenericReport,
  SLAPolicy,
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

  async testBackend(allFiles: DomainFileNode[], rootDir: string): Promise<string> {
    try {
      return await TestBackend(JSON.stringify(allFiles), rootDir);
    } catch (error) {
      throw this.handleError(error, 'Failed to test backend');
    }
  }

  async suggestContextFiles(
      task: string,
      allFiles: DomainFileNode[],
  ): Promise<string[]> {
    try {
      return await SuggestContextFiles(task, allFiles as unknown as any[]);
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

  // Stub: current working directory (not provided by backend yet)
  async getCurrentDirectory(): Promise<string> {
    try {
      // TODO: wire to backend if needed. Return empty string to skip auto-load.
      return "";
    } catch (error) {
      throw this.handleError(error, 'Failed to get current directory');
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

  // New streaming context methods
  async getContextLines(contextId: string, startIndex: number, endIndex: number): Promise<{ lines: string[] }> {
    try {
      const result = await GetContextLines(contextId, startIndex, endIndex);
      return JSON.parse(result);
    } catch (error) {
      throw this.handleError(error, 'Failed to get context lines');
    }
  }

  async startContextStreaming(contextId: string): Promise<void> {
    try {
      // This would be implemented in the backend to start streaming
      // For now, we'll just return
      return;
    } catch (error) {
      throw this.handleError(error, 'Failed to start context streaming');
    }
  }

  async getContextSummary(contextId: string): Promise<{ summary: string }> {
    try {
      const result = await GetStreamingContext(contextId);
      const stream = JSON.parse(result);
      return { summary: stream.description || 'Context summary not available' };
    } catch (error) {
      throw this.handleError(error, 'Failed to get context summary');
    }
  }

  async createStreamingContext(
    projectPath: string,
    includedPaths: string[],
    options?: {
      stripComments?: boolean
      includeManifest?: boolean
      maxTokens?: number
    }
  ): Promise<any> {
    try {
      const optionsJson = JSON.stringify(options || {});
      const result = await CreateStreamingContext(projectPath, includedPaths, optionsJson);
      return JSON.parse(result);
    } catch (error) {
      throw this.handleError(error, 'Failed to create streaming context');
    }
  }

  async closeStreamingContext(contextId: string): Promise<void> {
    try {
      await CloseStreamingContext(contextId);
    } catch (error) {
      throw this.handleError(error, 'Failed to close streaming context');
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
      const optionsJson = JSON.stringify(options || {});
      const contextJson = await BuildContext(projectPath, includedPaths, optionsJson);
      return JSON.parse(contextJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to build context');
    }
  }

  async getContext(id: string): Promise<Context> {
    try {
      const contextJson = await GetContext(id);
      return JSON.parse(contextJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to get context');
    }
  }

  async getProjectContexts(projectPath: string): Promise<Context[]> {
    try {
      const contextsJson = await GetProjectContexts(projectPath);
      return JSON.parse(contextsJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to get project contexts');
    }
  }

  async deleteContext(id: string): Promise<void> {
    try {
      await DeleteContext(id);
    } catch (error) {
      throw this.handleError(error, 'Failed to delete context');
    }
  }

  // New methods for ARK Code autonomous mode
  async startAutonomousTask(
    request: AutonomousTaskRequest,
  ): Promise<AutonomousTaskResponse> {
    try {
      const requestJson = JSON.stringify(request);
      const responseJson = await StartAutonomousTask(requestJson);
      return JSON.parse(responseJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to start autonomous task');
    }
  }

  async cancelAutonomousTask(taskId: string): Promise<void> {
    try {
      await CancelAutonomousTask(taskId);
    } catch (error) {
      throw this.handleError(error, 'Failed to cancel autonomous task');
    }
  }

  async getAutonomousTaskStatus(
    taskId: string,
  ): Promise<AutonomousTaskStatus> {
    try {
      const statusJson = await GetAutonomousTaskStatus(taskId);
      return JSON.parse(statusJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to get autonomous task status');
    }
  }

  async getReport(reportId: string): Promise<GenericReport> {
    try {
      const reportJson = await GetReport(reportId);
      return JSON.parse(reportJson);
    } catch (error) {
      throw this.handleError(error, "Failed to get report");
    }
  }

  async listReports(reportType?: string): Promise<GenericReport[]> {
    try {
      const reportsJson = await ListReports(reportType || "");
      return JSON.parse(reportsJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to list reports');
    }
  }

  async setSLAPolicy(policy: SLAPolicy): Promise<void> {
    try {
      const policyJson = JSON.stringify(policy);
      await SetSLAPolicy(policyJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to set SLA policy');
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
    
    // Check if this is a domain error from backend
    if (message.startsWith('domain_error:')) {
      try {
        const domainErrorJson = message.substring('domain_error:'.length);
        const domainError = JSON.parse(domainErrorJson);
        
        // Create structured frontend error
        const structuredError = new ApiError(`${context}: ${domainError.message}`);
        (structuredError as any).code = domainError.code;
        (structuredError as any).recoverable = domainError.recoverable;
        (structuredError as any).context = domainError.context;
        (structuredError as any).cause = domainError.cause;
        
        return structuredError;
      } catch (parseErr) {
        console.error('Failed to parse domain error:', parseErr);
      }
    }
    
    return new ApiError(`${context}: ${message}`);
  }
}

export const apiService = new ApiService();