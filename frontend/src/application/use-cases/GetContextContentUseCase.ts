import type { ContextRepository } from '@/domain/repositories/ContextRepository';
import type { 
  ContextChunk, 
  PaginationOptions 
} from '@/domain/entities/ContextSummary';

/**
 * Get Context Content Use Case - Memory-safe content retrieval
 * This use case retrieves paginated context content without loading full content into memory
 * CRITICAL OOM FIX: Prevents memory issues by using pagination
 */
export class GetContextContentUseCase {
  constructor(private contextRepository: ContextRepository) {}

  async execute(
    contextId: string,
    pagination: PaginationOptions
  ): Promise<ContextChunk> {
    // 1. Validate inputs
    if (!contextId) {
      throw new Error('Context ID is required');
    }

    if (!pagination) {
      throw new Error('Pagination options are required');
    }

    if (pagination.startLine < 0) {
      throw new Error('Start line cannot be negative');
    }

    if (pagination.lineCount <= 0) {
      throw new Error('Line count must be positive');
    }

    // 2. Limit the chunk size to prevent memory issues
    const maxChunkSize = 1000; // Maximum 1000 lines per chunk
    if (pagination.lineCount > maxChunkSize) {
      throw new Error(`Line count cannot exceed ${maxChunkSize} lines per chunk`);
    }

    // 3. Get paginated content from repository
    try {
      const chunk = await this.contextRepository.getContextContent(contextId, pagination);
      
      // 4. Validate returned chunk
      if (!chunk) {
        throw new Error('No content returned for the specified range');
      }

      return chunk;
    } catch (error) {
      if (error instanceof Error) {
        throw new Error(`Failed to get context content: ${error.message}`);
      }
      throw new Error('Failed to get context content: Unknown error');
    }
  }
}