import { ref, computed, watch } from 'vue';
import type { ContextChunk } from '@/domain/entities/ContextSummary';
import { useContextBuilderStore } from '@/stores/context-builder.store';

export interface PaginatedContextState {
  chunks: ContextChunk[];
  currentChunk: ContextChunk | null;
  isLoading: boolean;
  error: string | null;
  totalLines: number;
  currentLine: number;
  linesPerPage: number;
  hasMore: boolean;
}

/**
 * Composable for paginated context content loading
 * This prevents OOM issues by loading content in small chunks
 */
export function usePaginatedContext(contextId?: string, linesPerPage: number = 100) {
  const contextStore = useContextBuilderStore();
  
  // State
  const chunks = ref<ContextChunk[]>([]);
  const currentChunk = ref<ContextChunk | null>(null);
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const currentLine = ref(0);
  const totalLines = ref(0);
  
  // Computed
  const hasMore = computed(() => {
    return currentLine.value + linesPerPage < totalLines.value;
  });
  
  const currentPage = computed(() => {
    return Math.floor(currentLine.value / linesPerPage) + 1;
  });
  
  const totalPages = computed(() => {
    return Math.ceil(totalLines.value / linesPerPage);
  });
  
  const progress = computed(() => {
    if (totalLines.value === 0) return 0;
    return Math.min(100, Math.round(((currentLine.value + linesPerPage) / totalLines.value) * 100));
  });
  
  // Methods
  const loadChunk = async (startLine: number, lineCount: number = linesPerPage): Promise<ContextChunk | null> => {
    const activeContextId = contextId || contextStore.currentContextId;
    
    if (!activeContextId) {
      error.value = 'No context available';
      return null;
    }
    
    isLoading.value = true;
    error.value = null;
    
    try {
      const chunk = await contextStore.getContextContent(startLine, lineCount);
      
      // Cache the chunk
      const existingChunkIndex = chunks.value.findIndex(c => c.chunkId === chunk.chunkId);
      if (existingChunkIndex >= 0) {
        chunks.value[existingChunkIndex] = chunk;
      } else {
        chunks.value.push(chunk);
        // Limit cache size to prevent memory growth
        if (chunks.value.length > 10) {
          chunks.value.shift(); // Remove oldest chunk
        }
      }
      
      currentChunk.value = chunk;
      currentLine.value = startLine;
      
      return chunk;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load content';
      
      // Handle specific context not found errors
      if (errorMessage.includes('not found') || errorMessage.includes('streaming context not found')) {
        error.value = 'Context no longer available. Please rebuild the context.';
        // Clear any stale context references
        clearCache();
      } else {
        error.value = errorMessage;
      }
      
      return null;
    } finally {
      isLoading.value = false;
    }
  };
  
  const loadFirstChunk = async (): Promise<ContextChunk | null> => {
    return await loadChunk(0, linesPerPage);
  };
  
  const loadNextChunk = async (): Promise<ContextChunk | null> => {
    if (!hasMore.value) return null;
    return await loadChunk(currentLine.value + linesPerPage, linesPerPage);
  };
  
  const loadPreviousChunk = async (): Promise<ContextChunk | null> => {
    if (currentLine.value <= 0) return null;
    const prevStart = Math.max(0, currentLine.value - linesPerPage);
    return await loadChunk(prevStart, linesPerPage);
  };
  
  const goToLine = async (lineNumber: number): Promise<ContextChunk | null> => {
    const startLine = Math.max(0, Math.floor(lineNumber / linesPerPage) * linesPerPage);
    return await loadChunk(startLine, linesPerPage);
  };
  
  const goToPage = async (page: number): Promise<ContextChunk | null> => {
    const startLine = (page - 1) * linesPerPage;
    return await loadChunk(startLine, linesPerPage);
  };
  
  const searchInChunk = (query: string): { line: number; text: string }[] => {
    if (!currentChunk.value || !query) return [];
    
    const results: { line: number; text: string }[] = [];
    const lines = currentChunk.value.lines;
    
    lines.forEach((line, index) => {
      if (line.toLowerCase().includes(query.toLowerCase())) {
        results.push({
          line: currentChunk.value!.startLine + index,
          text: line
        });
      }
    });
    
    return results;
  };
  
  const clearCache = () => {
    chunks.value = [];
    currentChunk.value = null;
    error.value = null;
  };
  
  const getChunkFromCache = (startLine: number): ContextChunk | null => {
    return chunks.value.find(chunk => 
      chunk.startLine <= startLine && 
      chunk.endLine >= startLine
    ) || null;
  };
  
  // Estimate total lines from context summary
  watch(() => contextStore.contextSummaryState, (summary) => {
    if (summary) {
      // Rough estimation: assume average 80 chars per line
      totalLines.value = Math.ceil(summary.totalSize / 80);
    }
  }, { immediate: true });
  
  // Auto-load first chunk when context changes
  watch(() => contextStore.contextSummaryState?.id, async (newContextId) => {
    if (newContextId) {
      clearCache();
      await loadFirstChunk();
    }
  }, { immediate: true });
  
  return {
    // State
    chunks: computed(() => chunks.value),
    currentChunk: computed(() => currentChunk.value),
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),
    currentLine: computed(() => currentLine.value),
    totalLines: computed(() => totalLines.value),
    linesPerPage,
    
    // Computed
    hasMore,
    currentPage,
    totalPages,
    progress,
    
    // Methods
    loadChunk,
    loadFirstChunk,
    loadNextChunk,
    loadPreviousChunk,
    goToLine,
    goToPage,
    searchInChunk,
    clearCache,
    getChunkFromCache
  };
}