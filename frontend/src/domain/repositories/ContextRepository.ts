import type { 
  ContextSummary, 
  ContextChunk, 
  ContextBuildOptions, 
  ContextValidationResult, 
  PaginationOptions, 
  StreamingContext, 
  ContextMetrics 
} from '../entities/ContextSummary';

/**
 * Repository interface for context operations
 * This defines the contract without depending on infrastructure details
 */
export interface ContextRepository {
  /**
   * Build a new context and return lightweight summary
   * @param projectPath Root project path
   * @param includedPaths Files to include in context
   * @param options Build options
   * @returns Context summary without full content
   */
  buildContext(
    projectPath: string, 
    includedPaths: string[], 
    options?: ContextBuildOptions
  ): Promise<ContextSummary>;

  /**
   * Get context summary by ID
   * @param id Context identifier
   * @returns Lightweight context summary
   */
  getContextSummary(id: string): Promise<ContextSummary>;

  /**
   * Get paginated context content for memory-safe display
   * @param id Context identifier
   * @param pagination Pagination options
   * @returns Content chunk
   */
  getContextContent(id: string, pagination: PaginationOptions): Promise<ContextChunk>;

  /**
   * Delete context and free resources
   * @param id Context identifier
   */
  deleteContext(id: string): Promise<void>;

  /**
   * List all contexts for a project
   * @param projectPath Project path
   * @returns Array of context summaries
   */
  getProjectContexts(projectPath: string): Promise<ContextSummary[]>;

  /**
   * Validate context build request
   * @param projectPath Project path
   * @param includedPaths Files to include
   * @param options Build options
   * @returns Validation result with memory impact assessment
   */
  validateContextBuild(
    projectPath: string, 
    includedPaths: string[], 
    options?: ContextBuildOptions
  ): Promise<ContextValidationResult>;

  /**
   * Create streaming context for large content
   * @param projectPath Project path
   * @param includedPaths Files to include
   * @param options Build options
   * @returns Streaming context handle
   */
  createStreamingContext(
    projectPath: string, 
    includedPaths: string[], 
    options?: ContextBuildOptions
  ): Promise<StreamingContext>;

  /**
   * Get streaming context content by lines
   * @param contextId Streaming context ID
   * @param startIndex Start line index
   * @param endIndex End line index
   * @returns Content lines
   */
  getStreamingContextLines(
    contextId: string, 
    startIndex: number, 
    endIndex: number
  ): Promise<{ lines: string[] }>;

  /**
   * Close streaming context and free resources
   * @param contextId Streaming context ID
   */
  closeStreamingContext(contextId: string): Promise<void>;

  /**
   * Get context metrics and statistics
   * @param id Context identifier
   * @returns Context metrics
   */
  getContextMetrics(id: string): Promise<ContextMetrics>;
}