import type { ContextRepository } from '@/domain/repositories/ContextRepository';
import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';
import type { 
  ContextSummary, 
  ContextBuildOptions, 
  ContextValidationResult 
} from '@/domain/entities/ContextSummary';
import type { StreamingPolicy } from '@/domain/services/StreamingPolicy';
import type { MemoryManagementPolicy } from '@/domain/services/MemoryManagementPolicy';
import { defaultStreamingPolicy, defaultMemoryManagementPolicy } from '@/domain/services';

/**
 * Build Context Use Case - Memory-safe context building
 * This addresses the OOM issue by creating lightweight summaries instead of full content
 */
export class BuildContextUseCase {
  constructor(
    private contextRepository: ContextRepository,
    private projectRepository: ProjectRepository,
    private streamingPolicy: StreamingPolicy = defaultStreamingPolicy,
    private memoryPolicy: MemoryManagementPolicy = defaultMemoryManagementPolicy
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

    // 5. Assess memory impact and streaming decision using domain services
    const shouldUseStreaming = await this.shouldUseStreamingMode(validation, selectedFiles);

    if (shouldUseStreaming) {
      // Use streaming context for large builds
      // ContextRepository now returns ContextSummary directly
      return await this.contextRepository.createStreamingContext(
        projectPath,
        selectedFiles,
        { ...options, useStreaming: true }
      );
    } else {
      // Use regular build for smaller contexts
      return await this.contextRepository.buildContext(
        projectPath,
        selectedFiles,
        options
      );
    }
  }

  private async shouldUseStreamingMode(
    validation: ContextValidationResult, 
    selectedFiles: string[]
  ): Promise<boolean> {
    const { memoryImpact } = validation;
    
    // Use domain services for streaming decision
    const riskLevel = this.streamingPolicy.calculateRiskLevel({
      totalSize: memoryImpact.estimatedMemoryUsage,
      fileCount: selectedFiles.length,
      tokenCount: memoryImpact.estimatedTokens || 0
    });
    
    // Check memory limits using memory management policy
    const memoryCheck = await this.memoryPolicy.checkMemoryLimits({
      currentUsage: memoryImpact.estimatedMemoryUsage,
      requestedIncrease: memoryImpact.estimatedMemoryUsage,
      fileCount: selectedFiles.length
    });
    
    return this.streamingPolicy.shouldUseStreamingMode(riskLevel, memoryCheck);
  }

  // Removed convertStreamingToSummary method since ContextRepository returns ContextSummary directly
}