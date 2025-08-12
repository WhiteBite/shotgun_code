import {
  ListFiles,
  RequestShotgunContextGeneration,
  GenerateCode,
  SuggestContextFiles,
  GetSettings,
  SaveSettings,
  SelectDirectory,
  IsGitAvailable,
  GetUncommittedFiles,
  GetRichCommitHistory,
  ReadFileContent
} from '../../wailsjs/go/main/App';
import type { DomainFileNode, SettingsDTO, FileStatus, CommitWithFiles } from '@/types/dto';

class ApiService {
  selectDirectory = (): Promise<string> => SelectDirectory();
  listFiles = (projectPath: string, useGitignore: boolean, useCustom: boolean): Promise<DomainFileNode[]> => ListFiles(projectPath, useGitignore, useCustom);
  readFileContent = (rootDir: string, relPath: string): Promise<string> => ReadFileContent(rootDir, relPath);
  buildContext = (projectPath: string, filePaths: string[]): Promise<void> =>
      RequestShotgunContextGeneration(projectPath, filePaths);

  generateCode = (systemPrompt: string, userPrompt: string): Promise<string> =>
      GenerateCode(systemPrompt, userPrompt);
  suggestFiles = (task: string, allFiles: DomainFileNode[]): Promise<string[]> =>
      SuggestContextFiles(task, allFiles);

  getSettings = (): Promise<SettingsDTO> => GetSettings();
  saveSettings = (dto: SettingsDTO): Promise<void> => SaveSettings(dto);

  isGitAvailable = (): Promise<boolean> => IsGitAvailable();
  getUncommittedFiles = (projectRoot: string): Promise<FileStatus[]> => GetUncommittedFiles(projectRoot);
  getRichCommitHistory = (root: string, branch: string, limit: number): Promise<CommitWithFiles[]> => GetRichCommitHistory(root, branch, limit);
}

export const apiService = new ApiService();