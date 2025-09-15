import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';

/**
 * Read File Content Use Case
 * This use case handles reading file content using the project repository
 */
export class ReadFileContentUseCase {
  constructor(
    private projectRepository: ProjectRepository
  ) {}

  /**
   * Execute the read file content operation
   * @param rootPath Root project path
   * @param relativePath Relative path to file
   * @returns File content as string
   */
  async execute(rootPath: string, relativePath: string): Promise<string> {
    // Validate inputs
    if (!rootPath) {
      throw new Error('Root path is required');
    }

    if (!relativePath) {
      throw new Error('Relative path is required');
    }

    // Perform the read file content operation
    try {
      const content = await this.projectRepository.readFileContent(rootPath, relativePath);
      return content;
    } catch (error) {
      throw new Error(`Failed to read file content: ${error instanceof Error ? error.message : String(error)}`);
    }
  }
}