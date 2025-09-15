import { defineStore } from "pinia";
import { ref, computed, readonly, watch } from "vue";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useProjectStore } from "@/stores/project.store";
import type { Context } from "@/types/api";
import { container } from "@/infrastructure/container";
import type {
  ContextSummary,
  ContextMetrics,
  ContextBuildStatus,
  ContextSuggestion,
  StreamingContext,
  ContextValidationResult,
  MemoryImpact,
  ContextBuildOptions,
  ContextChunk,
  PaginationOptions
} from "@/domain/entities/ContextSummary";



export const useContextBuilderStore = defineStore("context-builder", () => {
  const fileTreeStore = useFileTreeStore();
  const projectStore = useProjectStore();

  // CRITICAL OOM FIX: Replace shotgunContextText with lightweight ContextSummary
  // This addresses the primary cause of memory issues by not storing full text in reactive state
  const contextSummaryState = ref<ContextSummary | null>(null);
  const streamingContext = ref<StreamingContext | null>(null);
  const contextLoading = ref(false);
  const contextError = ref<string | null>(null);

  // Legacy Context support for backward compatibility
  const currentContext = ref<Context | null>(null);

  const selectedFilesList = ref<string[]>([]);
  const isBuilding = ref(false);
  const error = ref<string | null>(null);
  // REMOVED: shotgunContextText - this was causing OOM issues
  const lastContextGeneration = ref<Date | null>(null);
  const taskDescription = ref<string>("");
  // Enhanced metrics and suggestions
  const contextMetrics = ref<ContextMetrics>({
    tokenCount: 0,
    estimatedCost: 0,
    buildTime: 0,
    fileCount: 0,
    characterCount: 0,
    averageFileSize: 0
  });
  
  const suggestions = ref<ContextSuggestion[]>([]);
  const buildStatus = ref<ContextBuildStatus>('idle');
  const validationErrors = ref<string[]>([]);
  
  // Legacy support - keeping original for compatibility
  const contextSummary = ref({
    files: 0,
    characters: 0,
    tokens: 0,
    cost: 0,
  });
  const contextStatus = ref({
    status: "none" as "none" | "current" | "changed" | "stale",
    message: "No context built",
  });
  
  // Auto-build settings
  const autoBuildEnabled = ref(true);
  const smartSuggestionsEnabled = ref(true);
  const maxTokenLimit = ref(0); // 0 = unlimited
  // EMERGENCY CRITICAL LIMITS - 4x reduction to prevent OutOfMemory
  const maxFileLimit = ref(10);                          // CRITICAL: Was 50 -> now 10
  const excludeBinaryFiles = ref(true);                  // CRITICAL: Exclude binary files by default
  const maxIndividualFileSize = ref(32 * 1024);          // CRITICAL: Was 128KB -> now 32KB
  const maxTotalContextSize = ref(256 * 1024);           // CRITICAL: Was 1MB -> now 256KB
  const useChunkedProcessing = ref(true);                // CRITICAL: Use chunked processing
  const chunkSize = ref(16 * 1024);                      // CRITICAL: Was 50KB -> now 16KB
  const maxMemoryUsageWarning = ref(10);                 // CRITICAL: Was 30MB -> now 10MB
  const allowedExtensions = ref([
    '.ts', '.js', '.vue', '.tsx', '.jsx',
    '.py', '.java', '.go', '.rs', '.cpp', '.c',
    '.cs', '.php', '.rb', '.swift', '.kt',
    '.md', '.txt', '.json', '.yaml', '.yml',
    '.xml', '.html', '.css', '.scss', '.less',
    '.sql', '.sh', '.bat', '.ps1', '.dockerfile'
  ]);
  const buildStartTime = ref<Date | null>(null);

  // CRITICAL OOM FIX: Add pagination state instead of storing full content
  const currentContextId = ref<string | null>(null);
  const currentContextChunks = ref<ContextChunk[]>([]);
  const currentContextChunkIndex = ref(0);
  const contextPageSize = ref(50); // CRITICAL: Load only 50 lines at a time (was 100)

  const hasSelectedFiles = computed(() => selectedFilesList.value.length > 0);
  const canBuildContext = computed(() => 
    selectedFilesList.value.length > 0 && 
    buildStatus.value !== 'building' && 
    buildStatus.value !== 'validating'
  );
  
  const isContextValid = computed(() => {
    return contextMetrics.value.tokenCount > 0 && 
           contextMetrics.tokenCount <= maxTokenLimit.value &&
           validationErrors.value.length === 0;
  });
  
  const shouldShowSuggestions = computed(() => {
    return smartSuggestionsEnabled.value && 
           suggestions.value.length > 0 &&
           selectedFilesList.value.length < 5; // Further reduced from 10
  });
  
  const contextHealth = computed(() => {
    const metrics = contextMetrics.value;
    let score = 100;
    
    // Penalize for too many tokens
    if (metrics.tokenCount > maxTokenLimit.value * 0.7) { // Further reduced from 0.8
      score -= 25; // Increased penalty
    }
    
    // Penalize for too few files (might lack context)
    if (metrics.fileCount < 2) {
      score -= 15;
    }
    
    // Penalize for validation errors
    score -= validationErrors.value.length * 15; // Increased penalty
    
    return Math.max(0, Math.min(100, score));
  });
  const selectedFilesCount = computed(() => {
    // Count only files, not directories
    return selectedFilesList.value.filter((filePath) => {
      const node = fileTreeStore.getFileByRelPath(filePath);
      return node && !node.isDir && !node.isGitignored && !node.isCustomIgnored;
    }).length;
  });
  const totalFilesCount = computed(() => fileTreeStore.totalFiles);

  // Add streaming context support
  const isStreaming = ref<boolean>(false);
  const streamPosition = ref<number>(0);
  const streamChunkSize = ref<number>(512); // CRITICAL: Reduced from 1KB to 512 bytes for streaming

  // Helper functions - moved before watchers to fix hoisting issues
  const generateSmartSuggestions = async () => {
    if (!smartSuggestionsEnabled.value || selectedFilesList.value.length === 0) {
      suggestions.value = [];
      return;
    }

    const newSuggestions: ContextSuggestion[] = [];
    
    // Analyze selected files for dependency suggestions
    for (const filePath of selectedFilesList.value) {
      const node = fileTreeStore.getFileByRelPath(filePath);
      if (node && !node.isDir) {
        // Suggest related files based on same directory
        const directory = filePath.substring(0, filePath.lastIndexOf('/'));
        
        // Get all files and filter for the same directory
        const allFiles = fileTreeStore.getAllFiles();
        const relatedFiles = allFiles
          .filter(f => {
            const fileDir = f.relPath.substring(0, f.relPath.lastIndexOf('/'));
            return fileDir === directory && 
                   !selectedFilesList.value.includes(f.relPath) && 
                   !f.relPath.endsWith('.test.ts') && 
                   !f.relPath.endsWith('.spec.ts');
          })
          .map(f => f.relPath)
          .slice(0, 2); // Further reduced from 3
        
        relatedFiles.forEach((relatedFile: string) => {
          newSuggestions.push({
            filePath: relatedFile,
            reason: `Related to ${filePath}`,
            confidence: 0.7,
            type: 'related'
          });
        });
      }
    }
    
    // Remove duplicates and sort by confidence
    const uniqueSuggestions = newSuggestions
      .filter((s, i, arr) => arr.findIndex(t => t.filePath === s.filePath) === i)
      .sort((a, b) => b.confidence - a.confidence)
      .slice(0, 3); // Further reduced from 5
    
    suggestions.value = uniqueSuggestions;
  };

  // Sync selected files from file tree store
  watch(
    () => fileTreeStore.selectedFiles,
    (newSelectedFiles) => {
      selectedFilesList.value = newSelectedFiles;
      
      // Auto-generate suggestions when selection changes
      if (smartSuggestionsEnabled.value) {
        generateSmartSuggestions();
      }
      
      // Auto-build context if enabled and reasonable number of files
      if (autoBuildEnabled.value && 
          selectedFilesList.value.length > 0 && 
          selectedFilesList.value.length <= 10) { // Further reduced limit
        if (projectStore.currentProject?.path) {
          buildContextFromSelection(projectStore.currentProject.path);
        }
      }
    },
    { immediate: true }
  );

  // Watch for project changes
  watch(
    () => projectStore.currentProject,
    (newProject) => {
      if (newProject) {
        // Reset context when project changes
        resetContext();
      }
    }
  );

  // Computed properties for context data access
  const estimatedTokens = computed(() => {
    if (contextSummaryState.value) {
      return contextSummaryState.value.tokenCount;
    }
    if (contextMetrics.value.tokenCount > 0) {
      return contextMetrics.value.tokenCount;
    }
    return 0;
  });

  // CRITICAL OOM FIX: Replace direct content access with paginated access
  const getContextContent = async (startLine: number = 0, lineCount: number = contextPageSize.value): Promise<ContextChunk | null> => {
    if (!currentContextId.value) {
      throw new Error("No context available");
    }

    try {
      // Use the GetContextContentUseCase for memory-safe content retrieval
      const getContextContentUseCase = container.getGetContextContentUseCase();
      const pagination: PaginationOptions = {
        startLine,
        lineCount: Math.min(lineCount, 1000) // Limit to prevent memory issues
      };
      
      const chunk = await getContextContentUseCase.execute(currentContextId.value, pagination);
      return chunk;
    } catch (error) {
      console.error("Error getting context content:", error);
      throw new Error(`Failed to get context content: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  // CRITICAL OOM FIX: Load context chunks on demand instead of storing full content
  const loadContextChunk = async (chunkIndex: number): Promise<ContextChunk | null> => {
    if (!currentContextId.value) {
      return null;
    }

    try {
      const startLine = chunkIndex * contextPageSize.value;
      const chunk = await getContextContent(startLine, contextPageSize.value);
      
      if (chunk) {
        // Update the chunk in our local cache
        currentContextChunks.value[chunkIndex] = chunk;
        return chunk;
      }
      
      return null;
    } catch (error) {
      console.error("Error loading context chunk:", error);
      return null;
    }
  };

  // CRITICAL OOM FIX: Get current context chunk without storing full content
  const getCurrentContextChunk = async (): Promise<ContextChunk | null> => {
    return loadContextChunk(currentContextChunkIndex.value);
  };

  const buildContextFromSelection = async (projectPath: string) => {
    if (selectedFilesList.value.length === 0) {
      error.value = "No files selected";
      return;
    }

    // Validate selection before building
    const validation = await validateSelection();
    if (!validation.isValid) {
      error.value = validation.errors[0] || "Invalid selection";
      validationErrors.value = validation.errors;
      return;
    }

    isBuilding.value = true;
    error.value = null;
    buildStatus.value = 'building';
    buildStartTime.value = new Date();

    try {
      // Use BuildContextUseCase for memory-safe context building
      const buildContextUseCase = container.getBuildContextUseCase();
      
      const options: ContextBuildOptions = {
        stripComments: true,
        includeManifest: true,
        maxTokens: maxTokenLimit.value > 0 ? maxTokenLimit.value : undefined,
        includeGitStatus: false,
        includeCommitHistory: false,
        useStreaming: true, // Always use streaming for memory safety
        forceProcess: false
      };

      const contextSummary = await buildContextUseCase.execute(
        projectPath,
        selectedFilesList.value,
        options
      );

      // CRITICAL OOM FIX: Store only the context ID, not the full content
      contextSummaryState.value = contextSummary;
      currentContextId.value = contextSummary.id;
      
      // Clear any existing chunks
      currentContextChunks.value = [];
      currentContextChunkIndex.value = 0;

      // Update metrics
      contextMetrics.value = {
        tokenCount: contextSummary.tokenCount,
        estimatedCost: contextSummary.tokenCount * 0.0001,
        buildTime: contextSummary.updatedAt ? 
          new Date(contextSummary.updatedAt).getTime() - new Date(contextSummary.createdAt).getTime() : 0,
        fileCount: contextSummary.fileCount,
        characterCount: contextSummary.totalSize,
        averageFileSize: contextSummary.fileCount > 0 ? contextSummary.totalSize / contextSummary.fileCount : 0
      };

      // Update legacy summary for compatibility
      contextSummary.value = {
        files: contextSummary.fileCount,
        characters: contextSummary.totalSize,
        tokens: contextSummary.tokenCount,
        cost: contextSummary.tokenCount * 0.0001,
      };

      // Update status
      contextStatus.value = {
        status: "current",
        message: `Context built with ${contextSummary.fileCount} files (${(contextSummary.totalSize / (1024 * 1024)).toFixed(2)} MB)`,
      };

      console.log("Context summary created successfully:", contextSummary);
    } catch (error) {
      console.error("Error setting context summary:", error);
      buildStatus.value = 'error';
      contextError.value = `Failed to create context summary: ${error instanceof Error ? error.message : 'Unknown error'}`;
    } finally {
      isBuilding.value = false;
      buildStatus.value = contextError.value ? 'error' : 'ready';
    }
  }

  // Additional helper functions
  
  // Memory management function to clear large objects - updated for ContextSummary
  const clearLargeObjects = () => {
    try {
      // Clear context summary state
      contextSummaryState.value = null;
      
      // Clear streaming context
      streamingContext.value = null;
      
      // Clear legacy context content but keep the structure
      if (currentContext.value) {
        currentContext.value.content = "";
      }
      
      // CRITICAL OOM FIX: Clear pagination state
      currentContextId.value = null;
      currentContextChunks.value = [];
      currentContextChunkIndex.value = 0;
      
      // Force garbage collection if available
      if (typeof window !== 'undefined' && window.gc) {
        try {
          window.gc();
        } catch (e) {
          console.warn('Failed to trigger garbage collection', e);
        }
      }
      
      console.log("Large objects cleared to free memory");
    } catch (error) {
      console.error("Error clearing large objects:", error);
    }
  };
  
  // New function to get memory usage statistics
  const getMemoryStats = () => {
    if ('performance' in window && 'memory' in (performance as any)) {
      const memory = (performance as any).memory;
      return {
        used: Math.round(memory.usedJSHeapSize / (1024 * 1024)),
        total: Math.round(memory.jsHeapSizeLimit / (1024 * 1024)),
        percentage: Math.round((memory.usedJSHeapSize / memory.jsHeapSizeLimit) * 100)
      };
    }
    return null;
  };
  
  const validateSelection = async (): Promise<{ isValid: boolean; errors: string[]; warnings: string[] }> => {
    const errors: string[] = [];
    const warnings: string[] = [];
    
    if (selectedFilesList.value.length === 0) {
      errors.push('No files selected');
      return { isValid: false, errors, warnings };
    }
    
    // CRITICAL: Even more strict limits to prevent memory issues
    const maxSelectedFiles = 10; // CRITICAL: Reduced from 50 to 10 files
    if (selectedFilesList.value.length > maxSelectedFiles) {
      errors.push(`Too many files selected (${selectedFilesList.value.length}). Maximum allowed is ${maxSelectedFiles} to prevent memory issues.`);
    }
    
    // Only apply file count limits if they are enabled (not unlimited)
    if (maxFileLimit.value > 0 && selectedFilesList.value.length > maxFileLimit.value) {
      errors.push(`Too many files selected (max ${maxFileLimit.value})`);
    }
    
    // Memory management checks
    let totalSelectedSize = 0;
    let oversizedFiles: string[] = [];
    
    // Check individual file sizes and total size
    for (const filePath of selectedFilesList.value) {
      const node = fileTreeStore.getFileByRelPath(filePath);
      if (node && !node.isDir) {
        // Check if file is too large
        if (node.size > maxIndividualFileSize.value) {
          const fileSizeMB = (node.size / (1024 * 1024)).toFixed(2);
          oversizedFiles.push(`${filePath} (${fileSizeMB} MB)`);
        }
        
        totalSelectedSize += node.size;
      }
    }
    
    // Add warnings for large files
    if (oversizedFiles.length > 0) {
      if (oversizedFiles.length > 1) { // Further reduced
        warnings.push(`${oversizedFiles.length} files exceed the recommended size limit (${maxIndividualFileSize.value / (1024 * 1024)} MB)`);
      } else {
        warnings.push(`These files exceed the recommended size limit: ${oversizedFiles.join(', ')}`);
      }
    }
    
    // Check if total context size is too large
    const maxSizeForMemory = maxTotalContextSize.value * 0.2; // CRITICAL: Further reduced threshold to 20%
    if (totalSelectedSize > maxSizeForMemory) {
      const totalSizeMB = (totalSelectedSize / (1024 * 1024)).toFixed(2);
      const limitMB = (maxSizeForMemory / (1024 * 1024)).toFixed(2);
      errors.push(`Total context size (${totalSizeMB} MB) exceeds the memory-safe limit (${limitMB} MB). Select fewer or smaller files to prevent crashes.`);
    }
    
    // Estimate token count for validation - only if limit is enabled
    if (maxTokenLimit.value > 0) {
      const estimatedTokens = Math.ceil(totalSelectedSize / 4); // Rough estimate
      if (estimatedTokens > maxTokenLimit.value * 0.5) { // Further reduced threshold to 50%
        warnings.push(`Estimated token count (${estimatedTokens.toLocaleString()}) is approaching limit (${maxTokenLimit.value.toLocaleString()})`);
      }
    }
    
    // Check available system memory
    if ('performance' in window && 'memory' in (performance as any)) {
      const memory = (performance as any).memory;
      const usedMemoryMB = memory.usedJSHeapSize / (1024 * 1024);
      const totalMemoryMB = memory.jsHeapSizeLimit / (1024 * 1024);
      
      if (usedMemoryMB > maxMemoryUsageWarning.value * 0.5) { // Further reduced threshold to 50%
        warnings.push(`High memory usage detected (${usedMemoryMB.toFixed(0)} MB). Consider closing other tabs or applications.`);
      }
      
      // Estimate if we'll run out of memory - even more conservative estimate
      const estimatedMemoryNeeded = totalSelectedSize * 1.5; // Further conservative estimate (1.5x file size)
      const availableMemory = memory.jsHeapSizeLimit - memory.usedJSHeapSize;
      
      if (estimatedMemoryNeeded > availableMemory * 0.2) { // Further reduced to 20% of available memory as threshold
        const neededMB = (estimatedMemoryNeeded / (1024 * 1024)).toFixed(0);
        const availableMB = (availableMemory / (1024 * 1024)).toFixed(0);
        errors.push(`Insufficient memory: need ~${neededMB} MB but only ${availableMB} MB available. Select fewer files.`);
      }
    }
    
    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  };
  
  const updateContextMetrics = (updates: Partial<ContextMetrics>) => {
    contextMetrics.value = { ...contextMetrics.value, ...updates };
  };
  
  const applySuggestion = (suggestion: ContextSuggestion) => {
    if (!selectedFilesList.value.includes(suggestion.filePath)) {
      addSelectedFile(suggestion.filePath);
      // Remove applied suggestion
      suggestions.value = suggestions.value.filter(s => s.filePath !== suggestion.filePath);
    }
  };
  
  const dismissSuggestion = (suggestion: ContextSuggestion) => {
    suggestions.value = suggestions.value.filter(s => s.filePath !== suggestion.filePath);
  };
  
  const toggleAutoBuild = () => {
    autoBuildEnabled.value = !autoBuildEnabled.value;
    localStorage.setItem('context-auto-build', autoBuildEnabled.value.toString());
  };
  
  const toggleSmartSuggestions = () => {
    smartSuggestionsEnabled.value = !smartSuggestionsEnabled.value;
    localStorage.setItem('context-smart-suggestions', smartSuggestionsEnabled.value.toString());
    if (!smartSuggestionsEnabled.value) {
      suggestions.value = [];
    }
  };
  
  const setMaxTokenLimit = (limit: number) => {
    maxTokenLimit.value = Math.max(0, Math.min(50000, limit)); // Further reduced max to 50k
    localStorage.setItem('context-max-tokens', maxTokenLimit.value.toString());
  };
  
  const setMaxFileLimit = (limit: number) => {
    maxFileLimit.value = Math.max(0, Math.min(200, limit)); // Further reduced max to 200
    localStorage.setItem('context-max-files', maxFileLimit.value.toString());
  };
  
  const setSelectedFiles = (files: string[]) => {
    selectedFilesList.value = files;
  };
  
  const addSelectedFile = (filePath: string) => {
    if (!selectedFilesList.value.includes(filePath)) {
      selectedFilesList.value.push(filePath);
    }
  };
  
  const removeSelectedFile = (filePath: string) => {
    selectedFilesList.value = selectedFilesList.value.filter(f => f !== filePath);
  };
  
  const clearSelectedFiles = () => {
    selectedFilesList.value = [];
  };
  
  const resetContext = () => {
    contextSummaryState.value = null;
    currentContext.value = null;
    contextLoading.value = false;
    contextError.value = null;
    isBuilding.value = false;
    error.value = null;
    buildStatus.value = 'idle';
    contextMetrics.value = {
      tokenCount: 0,
      estimatedCost: 0,
      buildTime: 0,
      fileCount: 0,
      characterCount: 0,
      averageFileSize: 0
    };
    contextSummary.value = {
      files: 0,
      characters: 0,
      tokens: 0,
      cost: 0,
    };
    contextStatus.value = {
      status: "none",
      message: "No context built",
    };
    
    // CRITICAL OOM FIX: Clear pagination state
    currentContextId.value = null;
    currentContextChunks.value = [];
    currentContextChunkIndex.value = 0;
    
    suggestions.value = [];
    validationErrors.value = [];
  };
  
  const setCurrentContext = (context: Context) => {
    currentContext.value = context;
    if (context.content) {
      lastContextGeneration.value = new Date();
    }
  };
  
  // CRITICAL OOM FIX: Remove setShotgunContext function that was storing full content
  // const setShotgunContext = (content: string) => {
  //   // This function was causing OOM issues by storing full content in reactive state
  //   // Replaced with pagination-based approach
  // };
  
  // CRITICAL OOM FIX: Add function to set context ID for pagination
  const setContextId = (contextId: string) => {
    currentContextId.value = contextId;
    // Clear any existing chunks when setting new context
    currentContextChunks.value = [];
    currentContextChunkIndex.value = 0;
  };
  
  // CRITICAL OOM FIX: Add function to navigate between chunks
  const setCurrentChunkIndex = (index: number) => {
    currentContextChunkIndex.value = index;
  };
  
  const incrementChunkIndex = () => {
    currentContextChunkIndex.value++;
  };
  
  const decrementChunkIndex = () => {
    if (currentContextChunkIndex.value > 0) {
      currentContextChunkIndex.value--;
    }
  };

  return {
    // State
    contextSummaryState: readonly(contextSummaryState),
    streamingContext: readonly(streamingContext),
    contextLoading: readonly(contextLoading),
    contextError: readonly(contextError),
    currentContext: readonly(currentContext),
    selectedFilesList: readonly(selectedFilesList),
    isBuilding: readonly(isBuilding),
    error: readonly(error),
    lastContextGeneration: readonly(lastContextGeneration),
    taskDescription: readonly(taskDescription),
    contextMetrics: readonly(contextMetrics),
    suggestions: readonly(suggestions),
    buildStatus: readonly(buildStatus),
    validationErrors: readonly(validationErrors),
    contextSummary: readonly(contextSummary),
    contextStatus: readonly(contextStatus),
    autoBuildEnabled: readonly(autoBuildEnabled),
    smartSuggestionsEnabled: readonly(smartSuggestionsEnabled),
    maxTokenLimit: readonly(maxTokenLimit),
    maxFileLimit: readonly(maxFileLimit),
    excludeBinaryFiles: readonly(excludeBinaryFiles),
    maxIndividualFileSize: readonly(maxIndividualFileSize),
    maxTotalContextSize: readonly(maxTotalContextSize),
    useChunkedProcessing: readonly(useChunkedProcessing),
    chunkSize: readonly(chunkSize),
    maxMemoryUsageWarning: readonly(maxMemoryUsageWarning),
    allowedExtensions: readonly(allowedExtensions),
    buildStartTime: readonly(buildStartTime),
    
    // CRITICAL OOM FIX: Add pagination state
    currentContextId: readonly(currentContextId),
    currentContextChunks: readonly(currentContextChunks),
    currentContextChunkIndex: readonly(currentContextChunkIndex),
    contextPageSize: readonly(contextPageSize),
    
    // Computed
    hasSelectedFiles,
    canBuildContext,
    isContextValid,
    shouldShowSuggestions,
    contextHealth,
    selectedFilesCount,
    totalFilesCount,
    estimatedTokens,
    
    // Add streaming context support
    isStreaming: readonly(isStreaming),
    streamPosition: readonly(streamPosition),
    streamChunkSize: readonly(streamChunkSize),
    
    // Methods
    buildContextFromSelection,
    validateSelection,
    updateContextMetrics,
    applySuggestion,
    dismissSuggestion,
    toggleAutoBuild,
    toggleSmartSuggestions,
    setMaxTokenLimit,
    setMaxFileLimit,
    setSelectedFiles,
    addSelectedFile,
    removeSelectedFile,
    clearSelectedFiles,
    resetContext,
    setCurrentContext,
    clearLargeObjects,
    getMemoryStats,
    
    // CRITICAL OOM FIX: Add pagination methods
    getContextContent,
    loadContextChunk,
    getCurrentContextChunk,
    setContextId,
    setCurrentChunkIndex,
    incrementChunkIndex,
    decrementChunkIndex
  };
});