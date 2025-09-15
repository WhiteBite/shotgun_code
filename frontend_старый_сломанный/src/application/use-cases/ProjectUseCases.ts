import type { ProjectRepository, FileTreeOptions } from '@/domain/repositories/ProjectRepository';
import type { DomainFileNode } from '@/types/dto';
import type { ProjectValidationService } from '@/domain/services/ProjectValidationService';

/**
 * Load Project Files Use Case
 */
export class LoadProjectFilesUseCase {
  constructor(
    private projectRepository: ProjectRepository,
    private validationService: ProjectValidationService
  ) {}

  async execute(
    projectPath: string,
    options?: FileTreeOptions
  ): Promise<DomainFileNode[]> {
    // 1. Validate project path
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }

    // 2. Validate path accessibility using the injected service
    const isValid = await this.validationService.validatePath(projectPath);
    if (!isValid) {
      throw new Error('Project path is not accessible or does not exist');
    }

    // 3. Load file tree with options
    try {
      return await this.projectRepository.loadFileTree(projectPath, options);
    } catch (error) {
      throw new Error(`Failed to load project files: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

/**
 * Select Project Directory Use Case
 */
export class SelectProjectDirectoryUseCase {
  constructor(
    private projectRepository: ProjectRepository,
    private validationService: ProjectValidationService
  ) {}

  async execute(): Promise<string> {
    try {
      const selectedPath = await this.projectRepository.selectDirectory();
      
      if (!selectedPath) {
        throw new Error('No directory selected');
      }

      // Validate the selected directory using the injected service
      const isValid = await this.validationService.validatePath(selectedPath);
      if (!isValid) {
        throw new Error('Selected directory is not accessible');
      }

      return selectedPath;
    } catch (error) {
      throw new Error(`Failed to select directory: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

/**
 * Read File Content Use Case
 */
export class ReadFileContentUseCase {
  constructor(
    private projectRepository: ProjectRepository,
    private validationService: ProjectValidationService
  ) {}

  async execute(rootPath: string, relativePath: string): Promise<string> {
    // 1. Validate inputs
    if (!rootPath?.trim()) {
      throw new Error('Root path is required');
    }

    if (!relativePath?.trim()) {
      throw new Error('Relative path is required');
    }

    // 2. Validate root path using the injected service
    const isValidRoot = await this.validationService.validatePath(rootPath);
    if (!isValidRoot) {
      throw new Error('Root path is not accessible');
    }

    // 3. Check if file exists using the injected service
    const fileExists = await this.validationService.fileExists(rootPath, relativePath);
    if (!fileExists) {
      throw new Error(`File does not exist: ${relativePath}`);
    }

    // 4. Read file content
    try {
      return await this.projectRepository.readFileContent(rootPath, relativePath);
    } catch (error) {
      throw new Error(`Failed to read file: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

/**
 * Start File Watching Use Case
 */
export class StartFileWatchingUseCase {
  constructor(
    private projectRepository: ProjectRepository,
    private validationService: ProjectValidationService
  ) {}

  async execute(
    rootPath: string,
    onFileChange: (event: { type: string; path: string; timestamp: string }) => void
  ): Promise<void> {
    // 1. Validate root path using the injected service
    if (!rootPath?.trim()) {
      throw new Error('Root path is required');
    }

    const isValid = await this.validationService.validatePath(rootPath);
    if (!isValid) {
      throw new Error('Root path is not accessible');
    }

    // 2. Start file watching
    try {
      await this.projectRepository.watchFileChanges(rootPath, (event) => {
        onFileChange({
          type: event.type,
          path: event.path,
          timestamp: event.timestamp
        });
      });
    } catch (error) {
      throw new Error(`Failed to start file watching: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

/**
 * Stop File Watching Use Case
 */
export class StopFileWatchingUseCase {
  constructor(private projectRepository: ProjectRepository) {}

  async execute(): Promise<void> {
    try {
      await this.projectRepository.stopFileWatcher();
    } catch (error) {
      throw new Error(`Failed to stop file watching: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

/**
 * Get File Statistics Use Case
 */
export class GetFileStatsUseCase {
  constructor(private projectRepository: ProjectRepository) {}

  async execute(rootPath: string, relativePath: string): Promise<{
    size: number;
    modified: string;
    isDirectory: boolean;
    isFile: boolean;
  }> {
    // 1. Validate inputs
    if (!rootPath?.trim()) {
      throw new Error('Root path is required');
    }

    if (!relativePath?.trim()) {
      throw new Error('Relative path is required');
    }

    // 2. Get file stats
    try {
      return await this.projectRepository.getFileStats(rootPath, relativePath);
    } catch (error) {
      throw new Error(`Failed to get file stats: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}