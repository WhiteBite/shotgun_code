/**
 * Lightweight context summary to replace full text storage in reactive state
 * This addresses the critical OOM issue caused by storing large context text in Vue reactivity
 */
export interface ContextSummary {
  id: string;
  projectPath: string;
  fileCount: number;
  totalSize: number;
  tokenCount: number;
  createdAt: string;
  updatedAt: string;
  status: ContextBuildStatus;
  metadata: ContextMetadata;
}

export interface ContextMetadata {
  buildDuration: number;
  lastModified: string;
  selectedFiles: string[];
  buildOptions?: ContextBuildOptions;
  warnings?: string[];
  errors?: string[];
}

export interface ContextBuildOptions {
  stripComments?: boolean;
  includeManifest?: boolean;
  maxTokens?: number;
  includeGitStatus?: boolean;
  includeCommitHistory?: boolean;
  useStreaming?: boolean;
  forceProcess?: boolean;
}

export type ContextBuildStatus = 
  | 'idle' 
  | 'building' 
  | 'ready' 
  | 'error' 
  | 'validating' 
  | 'streaming';

/**
 * Paginated context content for memory-safe display
 */
export interface ContextChunk {
  lines: string[];
  startLine: number;
  endLine: number;
  hasMore: boolean;
  chunkId: string;
  contextId: string;
}

/**
 * Context validation result
 */
export interface ContextValidationResult {
  isValid: boolean;
  errors: string[];
  warnings: string[];
  memoryImpact: MemoryImpact;
}

export interface MemoryImpact {
  estimatedSize: number;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  recommendations: string[];
}

/**
 * Context metrics for monitoring and optimization
 */
export interface ContextMetrics {
  tokenCount: number;
  estimatedCost: number;
  buildTime: number;
  fileCount: number;
  characterCount: number;
  averageFileSize: number;
  memoryUsage?: number;
  compressionRatio?: number;
}

/**
 * Context suggestion for intelligent file selection
 */
export interface ContextSuggestion {
  filePath: string;
  reason: string;
  confidence: number;
  type: 'dependency' | 'related' | 'recent' | 'pattern';
  metadata?: Record<string, unknown>;
}

/**
 * Pagination options for content loading
 */
export interface PaginationOptions {
  startLine: number;
  lineCount: number;
  includeLineNumbers?: boolean;
  enableSyntaxHighlighting?: boolean;
}

/**
 * Streaming context for real-time large context processing
 */
export interface StreamingContext {
  id: string;
  name: string;
  description: string;
  totalLines: number;
  totalCharacters: number;
  files: string[];
  createdAt: string;
  updatedAt: string;
  projectPath: string;
  tokenCount: number;
  status: ContextBuildStatus;
  progress?: StreamingProgress;
}

export interface StreamingProgress {
  filesProcessed: number;
  totalFiles: number;
  bytesProcessed: number;
  totalBytes: number;
  currentFile?: string;
  estimatedTimeRemaining?: number;
}