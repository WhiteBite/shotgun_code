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
} from "../../wailsjs/go/main/App";

import type {
  DomainFileNode,
  FileStatus,
  CommitWithFiles,
  SettingsDTO,
  FileNode,
} from "@/types/dto";

class ApiService {
  async listFiles(
      dirPath: string,
      useGitignore: boolean,
      useCustomIgnore: boolean,
  ): Promise<DomainFileNode[]> {
    return await ListFiles(dirPath, useGitignore, useCustomIgnore);
  }

  async readFileContent(rootDir: string, relPath: string): Promise<string> {
    return await ReadFileContent(rootDir, relPath);
  }

  async requestShotgunContextGeneration(
      rootDir: string,
      includedPaths: string[],
  ): Promise<void> {
    return await RequestShotgunContextGeneration(rootDir, includedPaths);
  }

  async getUncommittedFiles(projectRoot: string): Promise<FileStatus[]> {
    return await GetUncommittedFiles(projectRoot);
  }

  async getRichCommitHistory(
      projectRoot: string,
      branchName: string,
      limit: number,
  ): Promise<CommitWithFiles[]> {
    return await GetRichCommitHistory(projectRoot, branchName, limit);
  }

  async getFileContentAtCommit(
      projectRoot: string,
      filePath: string,
      commitHash: string,
  ): Promise<string> {
    return await GetFileContentAtCommit(projectRoot, filePath, commitHash);
  }

  async getGitignoreContent(projectRoot: string): Promise<string> {
    return await GetGitignoreContent(projectRoot);
  }

  async isGitAvailable(): Promise<boolean> {
    return await IsGitAvailable();
  }

  async generateCode(
      systemPrompt: string,
      userPrompt: string,
  ): Promise<string> {
    return await GenerateCode(systemPrompt, userPrompt);
  }

  async suggestContextFiles(
      task: string,
      allFiles: FileNode[],
  ): Promise<string[]> {
    return await SuggestContextFiles(task, allFiles);
  }

  async getSettings(): Promise<SettingsDTO> {
    return await GetSettings();
  }

  async saveSettings(dto: SettingsDTO): Promise<void> {
    return await SaveSettings(JSON.stringify(dto));
  }

  async refreshAIModels(provider: string, apiKey: string): Promise<void> {
    return await RefreshAIModels(provider, apiKey);
  }

  async startFileWatcher(rootDirPath: string): Promise<void> {
    return await StartFileWatcher(rootDirPath);
  }

  async stopFileWatcher(): Promise<void> {
    return await StopFileWatcher();
  }

  async selectDirectory(): Promise<string> {
    return await SelectDirectory();
  }

  async exportContext(settings: any): Promise<any> {
    const result = await ExportContext(JSON.stringify(settings));

    // Обработка больших файлов с FilePath
    if (result.filePath && result.isLarge) {
      // Для больших файлов показываем диалог сохранения
      try {
        // В Wails можно использовать runtime.SaveFileDialog, но пока просто логируем
        console.log('Large file exported to:', result.filePath);
        // TODO: Добавить диалог сохранения или автоматическое перемещение в Downloads
      } catch (error) {
        console.error('Failed to handle large file:', error);
      }
    }

    return result;
  }

  // Новый метод для очистки временных файлов
  async cleanupTempFiles(filePath: string): Promise<void> {
    // Вызываем новый метод из backend
    // return await CleanupTempFiles(filePath);
    // Пока просто логируем, метод будет добавлен в следующих итерациях
    console.log('Cleanup temp file:', filePath);
  }
}

export const apiService = new ApiService();
