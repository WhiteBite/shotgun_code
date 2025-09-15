/**
 * Project Loading Use Case
 * Handles project loading with proper result types expected by the project store
 */

import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';
import type { UseCaseResult, ProjectLoadResult } from '@/types/use-cases';
import { createSuccessResult, createErrorResult } from '@/types/use-cases';
import { Project } from '@/domain/entities';

export class LoadProjectUseCase {
  constructor(private projectRepository: ProjectRepository) {}

  async execute(projectPath: string): Promise<UseCaseResult<ProjectLoadResult>> {
    try {
      // 1. Validate project path
      if (!projectPath?.trim()) {
        throw new Error('Project path is required');
      }

      // 2. Validate path accessibility
      const isValid = await this.projectRepository.validatePath(projectPath);
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
      const project = Project.create({
        name: projectName,
        path: projectPath
      });

      // 5. Create a simple workspace state object
      const workspace = {
        mode: 'manual',
        panelConfigurations: {},
        switchMode: () => {},
        setPanelConfiguration: () => {},
        getPanelConfiguration: () => null,
        isValidForMode: () => true
      };

      const result = createSuccessResult({
        project,
        workspace
      });

      return result;

    } catch (error) {
      const errorMsg = `Failed to load project: ${error instanceof Error ? error.message : 'Unknown error'}`;
      return createErrorResult(errorMsg);
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