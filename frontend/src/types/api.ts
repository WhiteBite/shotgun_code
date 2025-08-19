// API Types for better type safety

export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  success: boolean;
}

export interface FileNode {
  name: string;
  path: string;
  relPath: string;
  isDir: boolean;
  size: number;
  children?: FileNode[];
  isGitignored: boolean;
  isCustomIgnored: boolean;
  isSelected?: boolean; // For UI state
}

export interface FileStatus {
  path: string;
  status: string;
}

export interface CommitWithFiles {
  hash: string;
  subject: string;
  author: string;
  date: string;
  files: string[];
  isMerge: boolean;
}

export interface SettingsDTO {
  customIgnoreRules: string;
  customPromptRules: string;
  openAIAPIKey: string;
  geminiAPIKey: string;
  openRouterAPIKey: string;
  localAIAPIKey: string;
  localAIHost: string;
  localAIModelName: string;
  selectedProvider: string;
  selectedModels: Record<string, string>;
  availableModels: Record<string, string[]>;
  useGitignore: boolean;
  useCustomIgnore: boolean;
}

export interface ExportSettings {
  mode: 'clipboard' | 'ai' | 'human';
  context: string;
  stripComments: boolean;
  includeManifest: boolean;
  exportFormat: 'plain' | 'manifest' | 'json';
  aiProfile: string;
  tokenLimit: number;
  fileSizeLimitKB: number;
  enableAutoSplit: boolean;
  maxTokensPerChunk: number;
  overlapTokens: number;
  splitStrategy: 'smart' | 'file' | 'token';
  theme: string;
  includeLineNumbers: boolean;
  includePageNumbers: boolean;
}

export interface ExportResult {
  mode: string;
  text?: string;
  fileName?: string;
  dataBase64?: string;
  filePath?: string;
  isLarge?: boolean;
  sizeBytes?: number;
}

export interface ContextAnalysisResult {
  task: string;
  taskType: string;
  priority: string;
  selectedFiles: FileNode[];
  dependencyFiles: FileNode[];
  context: string;
  analysisTime: number;
  recommendations: string[];
  estimatedTokens: number;
  confidence: number;
}

export interface Project {
  name: string;
  path: string;
  lastOpened?: string;
}

export interface Context {
  id: string;
  name: string;
  description: string;
  content: string;
  files: string[];
  createdAt: string;
  updatedAt: string;
  projectPath: string;
  tokenCount: number;
}

// Error types
export class ApiError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public code?: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export class ValidationError extends Error {
  constructor(
    message: string,
    public field?: string
  ) {
    super(message);
    this.name = 'ValidationError';
  }
}
