import type { ContextRepository } from '@/domain/repositories/ContextRepository';
import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';
import type { 
  ContextSummary, 
  ContextBuildOptions, 
  ContextValidationResult 
} from '@/domain/entities/ContextSummary';

/**
 * Build Context Use Case - Memory-safe context building
 * This addresses the OOM issue by creating lightweight summaries instead of full content
 */
export class BuildContextUseCase {
  constructor(
    private contextRepository: ContextRepository,
    private projectRepository: ProjectRepository
  ) {}

  async execute(
    projectPath: string,
    selectedFiles: string[],
    options?: ContextBuildOptions
  ): Promise<ContextSummary> {
    // 1. Validate inputs
    if (!projectPath) {
      throw new Error('Project path is required');
    }

    if (!selectedFiles || selectedFiles.length === 0) {
      throw new Error('At least one file must be selected');
    }

    // 2. Validate project path
    const isValidPath = await this.projectRepository.validatePath(projectPath);
    if (!isValidPath) {
      throw new Error('Invalid or inaccessible project path');
    }

    // 3. Validate context build request
    const validation = await this.contextRepository.validateContextBuild(
      projectPath,
      selectedFiles,
      options
    );

    if (!validation.isValid) {
      throw new Error(`Context validation failed: ${validation.errors.join(', ')}`);
    }

    // 4. Log warnings if any
    if (validation.warnings.length > 0) {
      console.warn('Context build warnings:', validation.warnings);
    }

    // 5. Assess memory impact and decide on approach
    const shouldUseStreaming = this.shouldUseStreamingMode(validation);

    if (shouldUseStreaming) {
      // Use streaming context for large builds
      const streamingContext = await this.contextRepository.createStreamingContext(
        projectPath,
        selectedFiles,
        { ...options, useStreaming: true }
      );

      // Convert streaming context to summary
      return this.convertStreamingToSummary(streamingContext);
    } else {
      // Use regular build for smaller contexts
      return await this.contextRepository.buildContext(
        projectPath,
        selectedFiles,
        options
      );
    }
  }

  private shouldUseStreamingMode(validation: ContextValidationResult): boolean {
    const { memoryImpact } = validation;
    
    // Use streaming for high or critical memory impact
    return memoryImpact.riskLevel === 'high' || memoryImpact.riskLevel === 'critical';
  }

  private convertStreamingToSummary(streamingContext: any): ContextSummary {
    return {
      id: streamingContext.id,
      projectPath: streamingContext.projectPath,
      fileCount: streamingContext.files.length,
      totalSize: streamingContext.totalCharacters,
      tokenCount: streamingContext.tokenCount,
      createdAt: streamingContext.createdAt,
      updatedAt: streamingContext.updatedAt,
      status: 'streaming',
      metadata: {
        buildDuration: 0,
        lastModified: streamingContext.updatedAt,
        selectedFiles: streamingContext.files,
        buildOptions: { useStreaming: true },
        warnings: [],
        errors: []
      }
    };
  }
}