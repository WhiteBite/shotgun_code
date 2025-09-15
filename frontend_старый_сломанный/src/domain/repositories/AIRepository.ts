import type { DomainFileNode } from '@/types/dto';

/**
 * AI generation options
 */
export interface AIGenerationOptions {
  provider?: string;
  model?: string;
  temperature?: number;
  maxTokens?: number;
  streaming?: boolean;
  context?: string;
}

/**
 * AI provider information
 */
export interface AIProviderInfo {
  name: string;
  status: 'available' | 'unavailable' | 'error';
  models: string[];
  capabilities: string[];
  endpoint?: string;
}

/**
 * Repository interface for AI operations
 */
export interface AIRepository {
  /**
   * Generate code using AI
   * @param systemPrompt System prompt
   * @param userPrompt User prompt
   * @param options Generation options
   * @returns Generated code
   */
  generateCode(
    systemPrompt: string,
    userPrompt: string,
    options?: AIGenerationOptions
  ): Promise<string>;

  /**
   * Generate intelligent code with context
   * @param task Task description
   * @param context Context string
   * @param options Generation options (as JSON string)
   * @returns Generated code
   */
  generateIntelligentCode(
    task: string,
    context: string,
    options?: string
  ): Promise<string>;

  /**
   * Generate code with advanced options
   * @param systemPrompt System prompt
   * @param userPrompt User prompt
   * @param options Options as JSON string
   * @returns Generated code
   */
  generateCodeWithOptions(
    systemPrompt: string,
    userPrompt: string,
    options?: string
  ): Promise<string>;

  /**
   * Get available AI provider information
   * @returns Provider information as JSON string
   */
  getProviderInfo(): Promise<string>;

  /**
   * List available AI models
   * @returns Array of model names
   */
  listAvailableModels(): Promise<string[]>;

  /**
   * Refresh AI models for a provider
   * @param provider Provider name
   * @param apiKey API key
   */
  refreshAIModels(provider: string, apiKey: string): Promise<void>;

  /**
   * Analyze task and collect relevant context
   * @param task Task description
   * @param allFilesJson All files as JSON string
   * @param rootDir Root directory
   * @returns Analysis result
   */
  analyzeTaskAndCollectContext(
    task: string,
    allFilesJson: string,
    rootDir: string
  ): Promise<string>;

  /**
   * Suggest context files based on task
   * @param task Task description
   * @param allFiles All available files
   * @returns Array of suggested file paths
   */
  suggestContextFiles(
    task: string,
    allFiles: DomainFileNode[]
  ): Promise<string[]>;

  /**
   * Test backend connectivity
   * @param allFiles All files
   * @param rootDir Root directory
   * @returns Test result
   */
  testBackend(allFiles: DomainFileNode[], rootDir: string): Promise<string>;
}