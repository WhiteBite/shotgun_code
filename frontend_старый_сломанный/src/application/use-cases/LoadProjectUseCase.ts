/**
 * Project Loading Use Case
 * Handles project loading with proper result types expected by the project store
 */

import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';
import type { UseCaseResult, ProjectLoadResult } from '@/types/use-cases';
import { createSuccessResult, createErrorResult } from '@/types/use-cases';
import { ProjectLoadingService } from '@/domain/services/ProjectLoadingService';

export class LoadProjectUseCase {
  private loadingService: ProjectLoadingService;

  constructor(private projectRepository: ProjectRepository) {
    this.loadingService = new ProjectLoadingService(projectRepository);
  }

  async execute(projectPath: string): Promise<UseCaseResult<ProjectLoadResult>> {
    console.log('🔧 LoadProjectUseCase.execute called with path:', projectPath);
    
    try {
      // Load project using the new service
      console.log('🏗️ Loading project using ProjectLoadingService...');
      const project = await this.loadingService.loadProject(projectPath);
      console.log('✅ Project loaded successfully:', { id: project.id, name: project.name, path: project.path });
      
      // 5. Create a simple workspace state object
      console.log('🏗️ Creating workspace state...');
      const workspace = {
        mode: 'manual',
        panelConfigurations: {},
        switchMode: () => {},
        setPanelConfiguration: () => {},
        getPanelConfiguration: () => null,
        isValidForMode: () => true
      };
      console.log('✅ Workspace state created');

      const result = createSuccessResult({
        project,
        workspace
      });
      
      console.log('🎉 Project loading completed successfully!');
      return result;

    } catch (error) {
      const errorMsg = `Failed to load project: ${error instanceof Error ? error.message : 'Unknown error'}`;
      console.error('❌ LoadProjectUseCase.execute failed:', errorMsg, error);
      return createErrorResult(errorMsg);
    }
  }
}