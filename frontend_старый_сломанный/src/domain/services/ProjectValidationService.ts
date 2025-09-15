import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';

/**
 * Project Validation Service
 * Encapsulates all project validation logic to maintain Single Responsibility Principle
 */
export class ProjectValidationService {
  constructor(private projectRepository: ProjectRepository) {}

  /**
   * Validates that a project path exists and is accessible
   * @param projectPath - The path to validate
   * @returns Promise<boolean> - True if path is valid and accessible
   */
  async validatePath(projectPath: string): Promise<boolean> {
    if (!projectPath?.trim()) {
      return false;
    }

    try {
      return await this.projectRepository.validatePath(projectPath);
    } catch (error) {
      console.warn('Path validation failed:', error);
      return false;
    }
  }

  /**
   * Checks if a file exists within a project
   * @param rootPath - The root path of the project
   * @param relativePath - The relative path of the file to check
   * @returns Promise<boolean> - True if file exists
   */
  async fileExists(rootPath: string, relativePath: string): Promise<boolean> {
    if (!rootPath?.trim() || !relativePath?.trim()) {
      return false;
    }

    try {
      return await this.projectRepository.fileExists(rootPath, relativePath);
    } catch (error) {
      console.warn('File existence check failed:', error);
      return false;
    }
  }

  /**
   * Validates multiple paths at once
   * @param paths - Array of paths to validate
   * @returns Promise<boolean[]> - Array of validation results
   */
  async validatePaths(paths: string[]): Promise<boolean[]> {
    return Promise.all(paths.map(path => this.validatePath(path)));
  }

  /**
   * Validates that all required paths for a project exist
   * @param projectPath - The main project path
   * @param requiredPaths - Additional paths that must exist
   * @returns Promise<boolean> - True if all paths are valid
   */
  async validateProjectStructure(projectPath: string, requiredPaths: string[] = []): Promise<boolean> {
    // Validate main project path
    const isMainPathValid = await this.validatePath(projectPath);
    if (!isMainPathValid) {
      return false;
    }

    // Validate required paths
    if (requiredPaths.length > 0) {
      const validationResults = await this.validatePaths(
        requiredPaths.map(path => `${projectPath}/${path}`)
      );
      return validationResults.every(result => result);
    }

    return true;
  }
}