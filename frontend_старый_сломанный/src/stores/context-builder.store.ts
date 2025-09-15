import { defineStore } from "pinia";
import { ref, computed, readonly, watch } from "vue";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useProjectStore } from "@/stores/project.store";
import type { Context } from "@/types/api";
import { container } from "@/infrastructure/container";
import { createStoreWithDependencies, type StoreDependencies } from '@/stores/StoreDependencyContainer';
import type {
  ContextSummary,
  ContextMetrics,
  ContextBuildStatus,
  ContextSuggestion,
  StreamingContext,
  ContextBuildOptions,
  ContextChunk,
  PaginationOptions
} from "@/domain/entities/ContextSummary";
// Import repository types
import type { ContextRepository } from '@/domain/repositories/ContextRepository';
import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';
import type { StorageRepository } from '@/domain/repositories/StorageRepository';
import type { SettingsRepository } from '@/domain/repositories/SettingsRepository';
import { APP_CONFIG } from '@/config/app-config';

export const useContextBuilderStore = defineStore("context-builder", () => {
  return createStoreWithDependencies('context-builder', (dependencies: StoreDependencies) => {
    const fileTreeStore = useFileTreeStore();
    const projectStore = useProjectStore();
    
    // Domain services from dependency injection
    const { 
      localStorageService, 
      tokenEstimationService,
      memoryManagementPolicy,
      streamingPolicy,
      performanceMonitoringService,
      observabilityService
    } = dependencies;

    // Inject repositories directly
    const contextRepository: ContextRepository = container.contextRepository;
    const projectRepository: ProjectRepository = container.projectRepository;
    const settingsRepository: SettingsRepository = container.settingsRepository;
    const storageRepository: StorageRepository = localStorageService; // Already StorageRepository

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
  const maxTokenLimit = ref(APP_CONFIG.performance.limits.MAX_TOKENS_PER_REQUEST); // Use config value (10000)
  // EMERGENCY CRITICAL LIMITS - Use values from centralized configuration
  const maxFileLimit = ref(APP_CONFIG.fileTree.limits.MAX_SELECTED_FILES); // Use config value (500)
  const excludeBinaryFiles = ref(true);
  const maxIndividualFileSize = ref(APP_CONFIG.performance.memory.MAX_INDIVIDUAL_FILE_SIZE); // Use config value (32KB)
  const maxTotalContextSize = ref(APP_CONFIG.context.estimation.RISK_THRESHOLDS.MEDIUM); // Use config value (50MB)
  const useChunkedProcessing = ref(true);
  const chunkSize = ref(16 * 1024); // 16KB
  const maxMemoryUsageWarning = ref(APP_CONFIG.performance.memory.WARNING_THRESHOLD_MB); // Use config value (20MB)
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
  const contextPageSize = ref(APP_CONFIG.context.pagination.DEFAULT_PAGE_SIZE); // Use config value (50)

  const hasSelectedFiles = computed(() => selectedFilesList.value.length > 0);
  const canBuildContext = computed(() => 
    selectedFilesList.value.length > 0 && 
    buildStatus.value !== 'building' && 
    buildStatus.value !== 'validating'
  );
  
  const isContextValid = computed(() => {
    return contextMetrics.value.tokenCount > 0 && 
           contextMetrics.value.tokenCount <= maxTokenLimit.value &&
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

  // Sync selected files from file tree store with enhanced debugging
  watch(
    () => fileTreeStore.selectedFiles,
    (newSelectedFiles, oldSelectedFiles) => {
      console.log(`🔄 File selection changed:`, {
        old: oldSelectedFiles?.length || 0,
        new: newSelectedFiles.length,
        files: newSelectedFiles
      });
      
      selectedFilesList.value = newSelectedFiles;
      
      // Auto-generate suggestions when selection changes
      if (smartSuggestionsEnabled.value) {
        generateSmartSuggestions();
      }
      
      // Auto-build context if enabled and reasonable number of files
      if (autoBuildEnabled.value && 
          selectedFilesList.value.length > 0 && 
          selectedFilesList.value.length <= 3) { // Use reduced limit
        if (projectStore.currentProject?.path) {
          console.log(`🚀 Auto-building context for ${selectedFilesList.value.length} files:`, selectedFilesList.value);
          buildContextFromSelection(projectStore.currentProject.path);
        } else {
          console.warn('❌ Cannot auto-build context: no current project');
        }
      } else if (selectedFilesList.value.length > 3) {
        console.log(`⏭️  Auto-build skipped: ${selectedFilesList.value.length} files exceeds limit of 3`);
      } else if (selectedFilesList.value.length === 0) {
        console.log(`🔍 No files selected, auto-build skipped`);
      }
    },
    { immediate: true, deep: true }
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
      // First, verify that the context still exists
      await contextRepository.getContextSummary(currentContextId.value);
      
      // Use the context repository directly for memory-safe content retrieval
      const pagination: PaginationOptions = {
        startLine,
        lineCount: Math.min(lineCount, 1000) // Limit to prevent memory issues
      };
      
      const chunk = await contextRepository.getContextContent(currentContextId.value, pagination);
      return chunk;
    } catch (error) {
      console.error("Error getting context content:", error);
      
      // If context is not found, clear the current context state
      if (error instanceof Error && error.message.includes('not found')) {
        console.warn('Context not found, clearing current context state');
        currentContextId.value = null;
        contextSummaryState.value = null;
        currentContextChunks.value = [];
        currentContextChunkIndex.value = 0;
      }
      
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

  // New method to validate if current context exists and is accessible
  const validateCurrentContext = async (): Promise<boolean> => {
    if (!currentContextId.value) {
      return false;
    }
    
    try {
      await contextRepository.getContextSummary(currentContextId.value);
      return true;
    } catch (error) {
      console.warn(`Context ${currentContextId.value} is no longer valid:`, error);
      // Clear invalid context state
      currentContextId.value = null;
      contextSummaryState.value = null;
      currentContextChunks.value = [];
      currentContextChunkIndex.value = 0;
      return false;
    }
  };

  // Enhanced memory management with proactive cleanup
  const performMemoryCleanup = () => {
    try {
      // Clear old chunks if we have too many - more aggressive
      if (currentContextChunks.value.length > 3) {
        const chunksToKeep = 2; // Keep only the most recent 2 chunks
        currentContextChunks.value = currentContextChunks.value.slice(-chunksToKeep);
        console.log(`Cleaned up context chunks, kept ${chunksToKeep} most recent`);
      }
      
      // Clear large objects if memory usage is high - lower threshold
      const memoryStats = getMemoryStats();
      if (memoryStats && memoryStats.used > 15) { // 15MB threshold (was 18MB)
        clearLargeObjects();
        console.log(`Memory cleanup triggered at ${memoryStats.used}MB usage`);
      }
      
      // Force garbage collection more aggressively
      if (typeof window !== 'undefined' && window.gc) {
        try {
          window.gc();
          console.log('Forced garbage collection');
        } catch (e) {
          // Ignore garbage collection errors
        }
      }
    } catch (error) {
      console.error('Error during memory cleanup:', error);
    }
  };

  // Improved context building with better cleanup, timeout and streaming support
  const buildContextFromSelection = async (projectPath: string) => {
    console.log(`🚀 Starting context build for project: ${projectPath}`);
    console.log(`📁 Selected files: ${selectedFilesList.value.length}`, selectedFilesList.value);
    
    // Perform memory cleanup before starting new context build
    performMemoryCleanup();
    
    if (selectedFilesList.value.length === 0) {
      error.value = "No files selected. Please select at least one file to build context.";
      console.error('❌ Context build failed: No files selected');
      return;
    }

    // Validate selection before building
    const validation = await validateSelection();
    if (!validation.isValid) {
      error.value = validation.errors[0] || "Invalid file selection. Please check your selected files.";
      validationErrors.value = validation.errors;
      console.error('❌ Context build failed: Validation errors', validation.errors);
      
      // Also show warnings to user
      if (validation.warnings.length > 0) {
        console.warn('⚠️  Context build warnings:', validation.warnings);
      }
      
      return;
    }

    isBuilding.value = true;
    error.value = null;
    buildStatus.value = 'building';
    buildStartTime.value = new Date();
    
    console.log('⚙️  Context building started');

    // Set timeout to prevent infinite spinner - CRITICAL FIX
    console.log('⏰ Setting 30-second timeout for context building...');
    const timeout = setTimeout(() => {
      if (isBuilding.value) {
        console.error('⏰ Context build timeout after 30 seconds');
        isBuilding.value = false;
        buildStatus.value = 'error';
        error.value = 'Context building timed out after 30 seconds. Please try with fewer or smaller files.';
      }
    }, 30000); // 30 second timeout

    try {
      // Clear any existing context before building new one
      if (currentContextId.value) {
        console.log(`🗑️  Cleaning up old context: ${currentContextId.value}`);
        try {
          await contextRepository.deleteContext(currentContextId.value);
          console.log('✅ Old context deleted successfully');
        } catch (deleteError) {
          console.warn('⚠️  Failed to delete old context:', deleteError);
        }
      }
      
      // Reset context state
      currentContextId.value = null;
      contextSummaryState.value = null;
      currentContextChunks.value = [];
      currentContextChunkIndex.value = 0;
      
      // Use context repository directly for memory-safe context building
      const options: ContextBuildOptions = {
        stripComments: true,
        includeManifest: true,
        maxTokens: maxTokenLimit.value > 0 ? maxTokenLimit.value : undefined,
        includeGitStatus: false,
        includeCommitHistory: false,
        useStreaming: true, // Always use streaming for memory safety
        forceProcess: false
      };
      
      console.log('🔧 Building context with options:', options);
      console.log('📁 Files to include:', selectedFilesList.value);

      // Try streaming context first for better memory management
      let contextResult;
      try {
        console.log('🌊 Attempting streaming context creation...');
        const streamingContext = await contextRepository.createStreamingContext(
          projectPath,
          selectedFilesList.value,
          options
        );
        
        console.log('✅ Streaming context created successfully:', streamingContext);
        
        // Convert streaming context to summary format
        contextResult = {
          id: streamingContext.id,
          projectPath: streamingContext.projectPath || projectPath,
          fileCount: streamingContext.files?.length || selectedFilesList.value.length,
          totalSize: streamingContext.totalCharacters || 0,
          tokenCount: streamingContext.tokenCount || 0,
          createdAt: streamingContext.createdAt,
          updatedAt: streamingContext.updatedAt,
          status: 'streaming' as const,
          metadata: {
            buildDuration: 0,
            lastModified: streamingContext.updatedAt,
            selectedFiles: streamingContext.files || selectedFilesList.value,
            buildOptions: options,
            warnings: [],
            errors: []
          }
        };
        
        console.log('✨ Converted streaming context to summary:', contextResult);
      } catch (streamingError) {
        console.warn('⚠️  Streaming context creation failed, falling back to regular context:', streamingError);
        
        // CRITICAL: Enhanced backend error detection for different types of backend failures
        if (streamingError instanceof Error) {
          const errorMsg = streamingError.message.toLowerCase();
          
          // Check for various backend "nil map" and service errors
          if (errorMsg.includes('nil map') || 
              errorMsg.includes('assignment to entry in nil map') ||
              errorMsg.includes('cannot read properties of undefined') ||
              errorMsg.includes('invalid json passed to callback') ||
              errorMsg.includes('unexpected end of json input')) {
            console.error('🚨 Backend service failure detected:', streamingError.message);
            clearTimeout(timeout);
            
            // Try one more time with minimal context to see if backend can recover
            try {
              console.log('🔄 Attempting minimal context creation as recovery...');
              const minimalOptions: ContextBuildOptions = {
                stripComments: true,
                includeManifest: false,
                includeGitStatus: false,
                includeCommitHistory: false,
                useStreaming: false, // Disable streaming for recovery
                forceProcess: false
              };
              
              // Try with only one file to minimize backend load
              const singleFile = selectedFilesList.value.slice(0, 1);
              console.log('🎯 Trying minimal context with single file:', singleFile);
              
              const minimalResult = await contextRepository.buildContext(
                projectPath,
                singleFile,
                minimalOptions
              );
              
              console.log('✅ Minimal context recovery successful:', minimalResult);
              return minimalResult; // Return the minimal result instead of throwing
              
            } catch (recoveryError) {
              console.error('❌ Minimal context recovery also failed:', recoveryError);
              throw new Error('Backend service error: The Go backend has encountered a critical error (nil map). Please restart the application and try again with fewer or smaller files.');
            }
          }
        }
        
        // Fallback to regular context building for other errors
        try {
          console.log('🔄 Attempting regular context building as fallback...');
          contextResult = await contextRepository.buildContext(
            projectPath,
            selectedFilesList.value,
            options
          );
          
          console.log('✅ Regular context built successfully:', contextResult);
        } catch (fallbackError) {
          console.error('❌ Both streaming and regular context building failed:', fallbackError);
          clearTimeout(timeout);
          throw fallbackError;
        }
      }

      // Clear timeout since we succeeded
      clearTimeout(timeout);

      // CRITICAL OOM FIX: Store only the context ID and summary, not the full content
      contextSummaryState.value = contextResult;
      currentContextId.value = contextResult.id;
      
      console.log(`📝 Context ID stored: ${currentContextId.value}`);
      
      // Clear any existing chunks
      currentContextChunks.value = [];
      currentContextChunkIndex.value = 0;

      // Update metrics
      contextMetrics.value = {
        tokenCount: contextResult.tokenCount,
        estimatedCost: contextResult.tokenCount * 0.0001,
        buildTime: contextResult.updatedAt ? 
          new Date(contextResult.updatedAt).getTime() - new Date(contextResult.createdAt).getTime() : 0,
        fileCount: contextResult.fileCount,
        characterCount: contextResult.totalSize,
        averageFileSize: contextResult.fileCount > 0 ? contextResult.totalSize / contextResult.fileCount : 0
      };

      // Update legacy summary for compatibility
      contextSummary.value = {
        files: contextResult.fileCount,
        characters: contextResult.totalSize,
        tokens: contextResult.tokenCount,
        cost: contextResult.tokenCount * 0.0001,
      };

      // Update status
      contextStatus.value = {
        status: "current",
        message: `Context built with ${contextResult.fileCount} files (${(contextResult.totalSize / (1024 * 1024)).toFixed(2)} MB)`,
      };

      buildStatus.value = 'ready';
      lastContextGeneration.value = new Date();
      
      console.log("🎉 Context summary created successfully:", contextSummary.value);
      console.log("🔍 Context metrics:", contextMetrics.value);
      
    } catch (buildError) {
      console.error("❌ Error building context:", buildError);
      
      // Clear timeout
      clearTimeout(timeout);
      
      buildStatus.value = 'error';
      
      // Provide more helpful error messages
      let errorMessage = 'Unknown error occurred during context building';
      if (buildError instanceof Error) {
        errorMessage = buildError.message;
        
        // Special handling for backend errors
        if (errorMessage.includes('nil map') || errorMessage.includes('assignment to entry in nil map')) {
          errorMessage = 'Backend service error detected. Please restart the application and try again with fewer files.';
        } else if (errorMessage.includes('timeout')) {
          errorMessage = 'Context building timed out. Please try selecting fewer or smaller files.';
        } else if (errorMessage.includes('memory')) {
          errorMessage = 'Not enough memory to build context. Please select fewer files or restart the application.';
        } else if (errorMessage.includes('network')) {
          errorMessage = 'Network error occurred. Please check your connection and try again.';
        }
      }
      
      error.value = `Failed to build context: ${errorMessage}`;
      
      // Clear context state on error
      currentContextId.value = null;
      contextSummaryState.value = null;
      currentContextChunks.value = [];
      currentContextChunkIndex.value = 0;
      
    } finally {
      isBuilding.value = false;
      
      console.log(`🏁 Context building completed. Status: ${buildStatus.value}`);
      
      // Perform cleanup after context building
      performMemoryCleanup();
    }
  };

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
      const memory: { usedJSHeapSize: number; jsHeapSizeLimit: number } = (performance as any).memory;
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
    
    // Use the application configuration limit instead of hardcoded value
    const maxSelectedFiles = APP_CONFIG.fileTree.limits.MAX_SELECTED_FILES; // 500 from config
    if (selectedFilesList.value.length > maxSelectedFiles) {
      errors.push(`Too many files selected (${selectedFilesList.value.length}). Maximum allowed is ${maxSelectedFiles}.`);
    }
    
    // Only apply file count limits if they are enabled (not unlimited)
    if (maxFileLimit.value > 0 && selectedFilesList.value.length > maxFileLimit.value) {
      errors.push(`Too many files selected (max ${maxFileLimit.value})`);
    }
    
    // Memory management checks
    let totalSelectedSize = 0;
    const oversizedFiles: string[] = [];
    
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
      if (oversizedFiles.length > 3) { // Show specific files only if few
        warnings.push(`${oversizedFiles.length} files exceed the recommended size limit (${maxIndividualFileSize.value / (1024 * 1024)} MB)`);
      } else {
        warnings.push(`These files exceed the recommended size limit: ${oversizedFiles.join(', ')}`);
      }
    }
    
    // Check if total context size is too large (use more reasonable threshold)
    const maxSizeForMemory = maxTotalContextSize.value * 0.8; // Increased from 20% to 80%
    if (totalSelectedSize > maxSizeForMemory) {
      const totalSizeMB = (totalSelectedSize / (1024 * 1024)).toFixed(2);
      const limitMB = (maxSizeForMemory / (1024 * 1024)).toFixed(2);
      errors.push(`Total context size (${totalSizeMB} MB) exceeds the memory-safe limit (${limitMB} MB). Select fewer or smaller files.`);
    }
    
    // Estimate token count for validation - only if limit is enabled
    if (maxTokenLimit.value > 0) {
      const estimatedTokens = Math.ceil(totalSelectedSize / 4); // Rough estimate
      if (estimatedTokens > maxTokenLimit.value * 0.8) { // Increased from 50% to 80%
        warnings.push(`Estimated token count (${estimatedTokens.toLocaleString()}) is approaching limit (${maxTokenLimit.value.toLocaleString()})`);
      }
    }
    
    // Check available system memory
    if ('performance' in window && 'memory' in (performance as any)) {
      const memory: { usedJSHeapSize: number; jsHeapSizeLimit: number } = (performance as any).memory;
      const usedMemoryMB = memory.usedJSHeapSize / (1024 * 1024);
      
      if (usedMemoryMB > maxMemoryUsageWarning.value * 0.8) { // Increased from 50% to 80%
        warnings.push(`High memory usage detected (${usedMemoryMB.toFixed(0)} MB). Consider closing other tabs or applications.`);
      }
      
      // Estimate if we'll run out of memory (more reasonable estimate)
      const estimatedMemoryNeeded = totalSelectedSize * 2; // More realistic estimate (2x file size)
      const availableMemory = memory.jsHeapSizeLimit - memory.usedJSHeapSize;
      
      if (estimatedMemoryNeeded > availableMemory * 0.7) { // Increased from 20% to 70% of available memory as threshold
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
    storageRepository.set('context-auto-build', autoBuildEnabled.value);
  };
  
  const toggleSmartSuggestions = () => {
    smartSuggestionsEnabled.value = !smartSuggestionsEnabled.value;
    storageRepository.set('context-smart-suggestions', smartSuggestionsEnabled.value);
    if (!smartSuggestionsEnabled.value) {
      suggestions.value = [];
    }
  };
  
  const setMaxTokenLimit = (limit: number) => {
    maxTokenLimit.value = Math.max(0, Math.min(50000, limit)); // Further reduced max to 50k
    storageRepository.set('context-max-tokens', maxTokenLimit.value);
  };
  
  const setMaxFileLimit = (limit: number) => {
    maxFileLimit.value = Math.max(0, Math.min(200, limit)); // Further reduced max to 200
    storageRepository.set('context-max-files', maxFileLimit.value);
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
    decrementChunkIndex,
    
    // Enhanced context management
    validateCurrentContext,
    performMemoryCleanup
  };
  }); // Close dependency injection wrapper
});