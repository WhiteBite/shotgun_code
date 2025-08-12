import {
  ListFiles,
  RequestShotgunContextGeneration,
  GenerateCode,
  SuggestContextFiles,
  GetSettings,
  SaveSettings,
  SelectDirectory,
  IsGitAvailable,
  GetUncommittedFiles
} from '../../wailsjs/go/main/App';
import type { DomainFileNode, SettingsDTO, FileStatus } from '@/types/dto';

/**
 * A service layer that centralizes all API calls to the Go backend.
 * This isolates Wails-specific calls and makes the rest of the app
 * easier to test and maintain.
 */
class ApiService {
  // Project and File System
  selectDirectory = (): Promise<string> => SelectDirectory();
  listFiles = (projectPath: string): Promise<DomainFileNode[]> => ListFiles(projectPath);
  buildContext = (projectPath: string, filePaths: string[]): Promise<void> =>
      RequestShotgunContextGeneration(projectPath, filePaths);

  // AI and Generation
  generateCode = (systemPrompt: string, userPrompt: string): Promise<string> =>
      GenerateCode(systemPrompt, userPrompt);
  suggestFiles = (task: string, allFiles: DomainFileNode[]): Promise<string[]> =>
      SuggestContextFiles(task, allFiles);

  // Settings
  getSettings = (): Promise<SettingsDTO> => GetSettings();
  saveSettings = (dto: SettingsDTO): Promise<void> => SaveSettings(dto);

  // Git
  isGitAvailable = (): Promise<boolean> => IsGitAvailable();
  getUncommittedFiles = (projectRoot: string): Promise<FileStatus[]> => GetUncommittedFiles(projectRoot);
}

// Export a singleton instance
export const apiService = new ApiService();