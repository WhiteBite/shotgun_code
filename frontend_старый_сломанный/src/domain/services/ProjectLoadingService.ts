import type { ProjectRepository, FileTreeOptions } from '@/domain/repositories/ProjectRepository';
import { Project } from '@/domain/entities';
import type { DomainFileNode } from '@/types/dto';
import { ProjectValidationService } from './ProjectValidationService';

/**
 * Project Loading Service
 * Encapsulates all project loading logic including validation and file tree loading
 */
export class ProjectLoadingService {
  private validationService: ProjectValidationService;

  constructor(private projectRepository: ProjectRepository) {
    this.validationService = new ProjectValidationService(projectRepository);
  }

  /**
   * Loads a project with full validation
   * @param projectPath - The path to the project
   * @returns Promise<Project> - The loaded project entity
   */
  async loadProject(projectPath: string): Promise<Project> {
    // 1. Validate project path
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }

    // 2. Validate path accessibility
    const isValid = await this.validationService.validatePath(projectPath);
    if (!isValid) {
      throw new Error('Project path is not accessible or does not exist');
    }

    // 3. Load file tree to verify project structure
    try {
      await this.projectRepository.loadFileTree(projectPath, {
        useGitignore: true,
        useCustomIgnore: true,
        maxDepth: 1 // Just check top level for validation
      });
    } catch (error) {
      throw new Error(`Failed to load project structure: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }

    // 4. Create project entity
    const projectName = this.getBasename(projectPath);
    return Project.create({
      name: projectName,
      path: projectPath,
      type: 'general'
    });
  }

  /**
   * Loads the file tree for a project
   * @param projectPath - The path to the project
   * @param options - File tree loading options
   * @returns Promise<DomainFileNode[]> - The loaded file tree
   */
  async loadFileTree(projectPath: string, options?: FileTreeOptions): Promise<DomainFileNode[]> {
    // 1. Validate project path
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }

    // 2. Validate path accessibility
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

  /**
   * Selects a project directory using the system dialog
   * @returns Promise<string> - The selected directory path
   */
  async selectProjectDirectory(): Promise<string> {
    try {
      const selectedPath = await this.projectRepository.selectDirectory();
      
      if (!selectedPath) {
        throw new Error('No directory selected');
      }

      // Validate the selected directory
      const isValid = await this.validationService.validatePath(selectedPath);
      if (!isValid) {
        throw new Error('Selected directory is not accessible');
      }

      return selectedPath;
    } catch (error) {
      throw new Error(`Failed to select directory: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Simple basename function for browser environment
   * @param path - The path to extract basename from
   * @returns string - The basename of the path
   */
  private getBasename(path: string): string {
    return path.split(/[/\\]/).pop() || path;
  }
}