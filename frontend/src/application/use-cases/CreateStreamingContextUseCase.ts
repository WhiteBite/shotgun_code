import type { ContextRepository } from '@/domain/repositories/ContextRepository';
import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';
import type { 
  StreamingContext, 
  ContextBuildOptions, 
  ContextValidationResult 
} from '@/domain/entities/ContextSummary';

/**
 * Create Streaming Context Use Case - For large context operations
 * This use case creates streaming contexts for large file sets to prevent OOM issues
 * CRITICAL OOM FIX: Uses streaming approach for memory-safe large context handling
 */
export class CreateStreamingContextUseCase {
  constructor(
    private contextRepository: ContextRepository,
    private projectRepository: ProjectRepository
  ) {}

  async execute(
    projectPath: string,
    selectedFiles: string[],
    options?: ContextBuildOptions
  ): Promise<StreamingContext> {
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

    // 4. Check if streaming is necessary/recommended
    if (validation.memoryImpact.riskLevel === 'low') {
      console.warn('Streaming context created for low-risk content. Regular context might be more efficient.');
    }

    // 5. Log warnings if any
    if (validation.warnings.length > 0) {
      console.warn('Streaming context warnings:', validation.warnings);
    }

    // 6. Create streaming context with memory-safe options
    const streamingOptions = {
      ...options,
      useStreaming: true,
      // Force memory-safe options for streaming
      stripComments: options?.stripComments ?? true, // Default to true to reduce size
      includeManifest: options?.includeManifest ?? true
    };

    try {
      const streamingContext = await this.contextRepository.createStreamingContext(
        projectPath,
        selectedFiles,
        streamingOptions
      );

      return streamingContext;
    } catch (error) {
      if (error instanceof Error) {
        throw new Error(`Failed to create streaming context: ${error.message}`);
      }
      throw new Error('Failed to create streaming context: Unknown error');
    }
  }

  /**
   * Check if streaming mode is recommended for the given parameters
   */
  async isStreamingRecommended(
    projectPath: string,
    selectedFiles: string[],
    options?: ContextBuildOptions
  ): Promise<{ recommended: boolean; reason: string; riskLevel: string }> {
    try {
      const validation = await this.contextRepository.validateContextBuild(
        projectPath,
        selectedFiles,
        options
      );

      const isRecommended = validation.memoryImpact.riskLevel === 'high' || 
                           validation.memoryImpact.riskLevel === 'critical';

      return {
        recommended: isRecommended,
        reason: isRecommended 
          ? `Memory risk level is ${validation.memoryImpact.riskLevel}` 
          : `Memory risk level is ${validation.memoryImpact.riskLevel}, streaming not necessary`,
        riskLevel: validation.memoryImpact.riskLevel
      };
    } catch (error) {
      return {
        recommended: true, // Default to streaming on validation error
        reason: 'Unable to assess memory impact, defaulting to streaming for safety',
        riskLevel: 'unknown'
      };
    }
  }
}