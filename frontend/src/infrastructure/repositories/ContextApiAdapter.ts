import type { 
  ContextRepository 
} from '@/domain/repositories/ContextRepository';
import type { 
  ContextSummary, 
  ContextChunk, 
  ContextBuildOptions, 
  ContextValidationResult, 
  PaginationOptions, 
  StreamingContext, 
  ContextMetrics 
} from '@/domain/entities/ContextSummary';
import type { Context } from '@/types/api';
import { 
  BuildContext,
  GetContext,
  GetProjectContexts,
  DeleteContext,
  CreateStreamingContext,
  GetContextLines,
  CloseStreamingContext,
  RequestShotgunContextGeneration
} from '../../../wailsjs/go/main/App';

/**
 * CRITICAL OOM FIX: Context API Adapter with memory-safe implementation
 * This handles context operations using ContextSummary approach to prevent memory issues
 */
export class ContextApiAdapter implements ContextRepository {
  async buildContext(
    projectPath: string, 
    includedPaths: string[], 
    options?: ContextBuildOptions
  ): Promise<ContextSummary> {
    // Input validation
    if (!projectPath || projectPath.trim() === '') {
      throw new Error('Project path is required');
    }
    
    if (!includedPaths || includedPaths.length === 0) {
      throw new Error('At least one file must be included');
    }
    
    try {
      const optionsJson = JSON.stringify(options || {});
      // CRITICAL OOM FIX: BuildContext now returns ContextSummary instead of full Context
      const contextSummaryJson = await BuildContext(projectPath, includedPaths, optionsJson);
      const contextSummary: ContextSummary = JSON.parse(contextSummaryJson);
      
      return contextSummary;
    } catch (error) {
      if (error instanceof Error && error.message === 'Project path is required') {
        throw error;
      }
      if (error instanceof Error && error.message === 'At least one file must be included') {
        throw error;
      }
      if (error instanceof SyntaxError) {
        throw new Error('Failed to parse context summary response');
      }
      throw this.handleError(error, 'Failed to build context');
    }
  }

  async getContextSummary(id: string): Promise<ContextSummary> {
    try {
      const contextJson = await GetContext(id);
      const context: Context = JSON.parse(contextJson);
      
      // Convert to lightweight summary
      return this.convertToContextSummary(context);
    } catch (error) {
      throw this.handleError(error, 'Failed to get context summary');
    }
  }

  async getContextContent(id: string, pagination: PaginationOptions): Promise<ContextChunk> {
    // Input validation
    if (!id || id.trim() === '') {
      throw new Error('Context ID is required');
    }
    
    try {
      // CRITICAL OOM FIX: Use GetContextLines for paginated access
      const linesJson = await GetContextLines(id, pagination.startLine, pagination.lineCount);
      const lines: string[] = JSON.parse(linesJson);
      
      // Create a context chunk from the lines
      const content = lines.join('\n');
      const tokens = Math.ceil(content.length / 4); // Rough estimation
      
      const chunk: ContextChunk = {
        id: `${id}-chunk-${pagination.startLine}-${pagination.lineCount}`,
        content,
        tokens,
        startLine: pagination.startLine,
        endLine: pagination.startLine + lines.length - 1,
        startChar: 0,
        endChar: content.length,
        significance: 'medium',
        metadata: {}
      };
      
      return chunk;
    } catch (error) {
      if (error instanceof Error && error.message === 'Context ID is required') {
        throw error;
      }
      if (error instanceof SyntaxError) {
        throw new Error('Failed to parse context content response');
      }
      throw this.handleError(error, 'Failed to get context content');
    }
  }

  async deleteContext(id: string): Promise<void> {
    try {
      await DeleteContext(id);
    } catch (error) {
      throw this.handleError(error, 'Failed to delete context');
    }
  }

  async getProjectContexts(projectPath: string): Promise<ContextSummary[]> {
    try {
      const contextsJson = await GetProjectContexts(projectPath);
      const contexts: Context[] = JSON.parse(contextsJson);
      
      // Convert all contexts to summaries
      return contexts.map(context => this.convertToContextSummary(context));
    } catch (error) {
      throw this.handleError(error, 'Failed to get project contexts');
    }
  }

  async validateContextBuild(
    projectPath: string, 
    includedPaths: string[], 
    options?: ContextBuildOptions
  ): Promise<ContextValidationResult> {
    try {
      // Enhanced validation logic with memory impact assessment
      const errors: string[] = [];
      const warnings: string[] = [];
      
      if (includedPaths.length === 0) {
        errors.push('No files selected');
      }
      
      // More stringent limits to prevent OOM
      if (includedPaths.length > 50) {
        errors.push('Too many files selected (maximum 50 allowed)');
      } else if (includedPaths.length > 30) {
        warnings.push('Large number of files selected may impact performance');
      }
      
      // Estimate memory impact with conservative calculations
      const estimatedSize = includedPaths.length * 15360; // More conservative: 15KB per file
      let riskLevel: 'low' | 'medium' | 'high' | 'critical' = 'low';
      const recommendations: string[] = [];
      
      if (estimatedSize > 2 * 1024 * 1024) { // > 2MB (reduced from 5MB)
        riskLevel = 'critical';
        recommendations.push('Reduce the number of files or use streaming mode');
        warnings.push('Context size will likely cause memory issues');
      } else if (estimatedSize > 1 * 1024 * 1024) { // > 1MB
        riskLevel = 'high';
        recommendations.push('Consider using streaming mode for better performance');
        warnings.push('Context size may cause performance issues');
      } else if (estimatedSize > 512 * 1024) { // > 512KB
        riskLevel = 'medium';
        recommendations.push('Monitor memory usage during context build');
      }
      
      return {
        isValid: errors.length === 0,
        errors,
        warnings,
        memoryImpact: {
          estimatedSize,
          riskLevel,
          recommendations
        }
      };
    } catch (error) {
      throw this.handleError(error, 'Failed to validate context build');
    }
  }

  async createStreamingContext(
    projectPath: string, 
    includedPaths: string[], 
    options?: ContextBuildOptions
  ): Promise<StreamingContext> {
    try {
      const optionsJson = JSON.stringify(options || {});
      const result = await CreateStreamingContext(projectPath, includedPaths, optionsJson);
      return JSON.parse(result);
    } catch (error) {
      throw this.handleError(error, 'Failed to create streaming context');
    }
  }

  async getStreamingContextLines(
    contextId: string, 
    startIndex: number, 
    endIndex: number
  ): Promise<{ lines: string[] }> {
    try {
      const result = await GetContextLines(contextId, startIndex, endIndex);
      return JSON.parse(result);
    } catch (error) {
      throw this.handleError(error, 'Failed to get streaming context lines');
    }
  }

  async closeStreamingContext(contextId: string): Promise<void> {
    try {
      await CloseStreamingContext(contextId);
    } catch (error) {
      throw this.handleError(error, 'Failed to close streaming context');
    }
  }

  async getContextMetrics(id: string): Promise<ContextMetrics> {
    try {
      const summary = await this.getContextSummary(id);
      
      return {
        tokenCount: summary.tokenCount,
        estimatedCost: summary.tokenCount * 0.0001, // $0.0001 per token estimate
        buildTime: summary.metadata.buildDuration,
        fileCount: summary.fileCount,
        characterCount: summary.totalSize,
        averageFileSize: summary.fileCount > 0 ? summary.totalSize / summary.fileCount : 0
      };
    } catch (error) {
      throw this.handleError(error, 'Failed to get context metrics');
    }
  }

  // Legacy method for backward compatibility - DEPRECATED
  async requestShotgunContextGeneration(
    projectPath: string,
    includedPaths: string[]
  ): Promise<void> {
    try {
      await RequestShotgunContextGeneration(projectPath, includedPaths);
    } catch (error) {
      throw this.handleError(error, 'Failed to request shotgun context generation');
    }
  }

  // Private helper methods
  private convertToContextSummary(
    context: Context, 
    selectedFiles?: string[], 
    options?: ContextBuildOptions
  ): ContextSummary {
    return {
      id: context.id,
      projectPath: context.projectPath || '',
      fileCount: context.files?.length || selectedFiles?.length || 0,
      totalSize: context.content?.length || 0,
      tokenCount: context.tokenCount || Math.ceil((context.content?.length || 0) / 4),
      createdAt: context.createdAt,
      updatedAt: context.updatedAt,
      status: 'ready',
      metadata: {
        buildDuration: 0, // Would need to be tracked during build
        lastModified: context.updatedAt,
        selectedFiles: selectedFiles || context.files || [],
        buildOptions: options,
        warnings: [],
        errors: []
      }
    };
  }

  private handleError(error: unknown, context: string): Error {
    const message = error instanceof Error ? error.message : String(error);
    
    // Check if this is a domain error from backend
    if (message.startsWith('domain_error:')) {
      try {
        const domainErrorJson = message.substring('domain_error:'.length);
        const domainError = JSON.parse(domainErrorJson);
        
        const structuredError = new Error(`${context}: ${domainError.message}`);
        (structuredError as any).code = domainError.code;
        (structuredError as any).recoverable = domainError.recoverable;
        (structuredError as any).context = domainError.context;
        (structuredError as any).cause = domainError.cause;
        
        return structuredError;
      } catch (parseErr) {
        console.error('Failed to parse domain error:', parseErr);
      }
    }
    
    return new Error(`${context}: ${message}`);
  }
}