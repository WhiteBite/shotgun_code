import type { 
  AIRepository, 
  AIGenerationOptions, 
  AIProviderInfo 
} from '@/domain/repositories/AIRepository';
import type { DomainFileNode } from '@/types/dto';
import { 
  GenerateCode,
  GenerateIntelligentCode,
  GenerateCodeWithOptions,
  GetProviderInfo,
  ListAvailableModels,
  RefreshAIModels,
  AnalyzeTaskAndCollectContext,
  SuggestContextFiles,
  TestBackend
} from '../../../wailsjs/go/main/App';

/**
 * AI API Adapter - Infrastructure implementation of AIRepository
 * This handles AI operations while conforming to Clean Architecture
 */
export class AIApiAdapter implements AIRepository {
  async generateCode(
    systemPrompt: string,
    userPrompt: string,
    options?: AIGenerationOptions
  ): Promise<string> {
    try {
      if (options) {
        const optionsJson = JSON.stringify(options);
        return await GenerateCodeWithOptions(systemPrompt, userPrompt, optionsJson);
      } else {
        return await GenerateCode(systemPrompt, userPrompt);
      }
    } catch (error) {
      throw this.handleError(error, 'Failed to generate code');
    }
  }

  async generateIntelligentCode(
    task: string,
    context: string,
    options?: string
  ): Promise<string> {
    try {
      return await GenerateIntelligentCode(task, context, options || '{}');
    } catch (error) {
      throw this.handleError(error, 'Failed to generate intelligent code');
    }
  }

  async generateCodeWithOptions(
    systemPrompt: string,
    userPrompt: string,
    options?: string
  ): Promise<string> {
    try {
      return await GenerateCodeWithOptions(systemPrompt, userPrompt, options || '{}');
    } catch (error) {
      throw this.handleError(error, 'Failed to generate code with options');
    }
  }

  async getProviderInfo(): Promise<string> {
    try {
      return await GetProviderInfo();
    } catch (error) {
      throw this.handleError(error, 'Failed to get provider info');
    }
  }

  async listAvailableModels(): Promise<string[]> {
    try {
      return await ListAvailableModels();
    } catch (error) {
      throw this.handleError(error, 'Failed to list available models');
    }
  }

  async refreshAIModels(provider: string, apiKey: string): Promise<void> {
    try {
      await RefreshAIModels(provider, apiKey);
    } catch (error) {
      throw this.handleError(error, 'Failed to refresh AI models');
    }
  }

  async analyzeTaskAndCollectContext(
    task: string,
    allFilesJson: string,
    rootDir: string
  ): Promise<string> {
    try {
      return await AnalyzeTaskAndCollectContext(task, allFilesJson, rootDir);
    } catch (error) {
      throw this.handleError(error, 'Failed to analyze task and collect context');
    }
  }

  async suggestContextFiles(
    task: string,
    allFiles: DomainFileNode[]
  ): Promise<string[]> {
    try {
      return await SuggestContextFiles(task, allFiles as unknown as unknown[]);
    } catch (error) {
      throw this.handleError(error, 'Failed to suggest context files');
    }
  }

  async testBackend(allFiles: DomainFileNode[], rootDir: string): Promise<string> {
    try {
      return await TestBackend(JSON.stringify(allFiles), rootDir);
    } catch (error) {
      throw this.handleError(error, 'Failed to test backend');
    }
  }

  // Helper method to parse provider info
  async getParsedProviderInfo(): Promise<AIProviderInfo[]> {
    try {
      const providerInfoJson = await this.getProviderInfo();
      const parsed = JSON.parse(providerInfoJson);
      
      // Convert to standardized format
      if (Array.isArray(parsed)) {
        return parsed.map(provider => ({
          name: provider.name || 'Unknown',
          status: provider.status || 'unavailable',
          models: provider.models || [],
          capabilities: provider.capabilities || [],
          endpoint: provider.endpoint
        }));
      } else if (parsed.name) {
        // Single provider format
        return [{
          name: parsed.name,
          status: parsed.status || 'available',
          models: parsed.models || [],
          capabilities: parsed.capabilities || [],
          endpoint: parsed.endpoint
        }];
      } else {
        return [];
      }
    } catch (error) {
      throw this.handleError(error, 'Failed to parse provider info');
    }
  }

  // Helper method for enhanced code generation with validation
  async generateValidatedCode(
    systemPrompt: string,
    userPrompt: string,
    options?: AIGenerationOptions & {
      validateSyntax?: boolean;
      maxRetries?: number;
    }
  ): Promise<string> {
    const maxRetries = options?.maxRetries || 3;
    let lastError: Error | null = null;
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        const code = await this.generateCode(systemPrompt, userPrompt, options);
        
        // Basic validation if requested
        if (options?.validateSyntax) {
          await this.validateCodeSyntax(code);
        }
        
        return code;
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));
        
        if (attempt < maxRetries) {
          console.warn(`Code generation attempt ${attempt} failed, retrying...`, error);
          // Add retry context to prompt
          systemPrompt += `\n\nPrevious attempt failed with error: ${lastError.message}. Please fix the issue.`;
        }
      }
    }
    
    throw lastError || new Error('Code generation failed after all retries');
  }

  // Basic syntax validation (can be extended)
  private async validateCodeSyntax(code: string): Promise<void> {
    // Basic checks for common syntax issues
    const lines = code.split('\n');
    const issues: string[] = [];
    
    let braceCount = 0;
    let parenCount = 0;
    let bracketCount = 0;
    
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      
      // Count brackets
      for (const char of line) {
        switch (char) {
          case '{': braceCount++; break;
          case '}': braceCount--; break;
          case '(': parenCount++; break;
          case ')': parenCount--; break;
          case '[': bracketCount++; break;
          case ']': bracketCount--; break;
        }
      }
    }
    
    if (braceCount !== 0) issues.push('Unmatched curly braces');
    if (parenCount !== 0) issues.push('Unmatched parentheses');
    if (bracketCount !== 0) issues.push('Unmatched square brackets');
    
    if (issues.length > 0) {
      throw new Error(`Syntax validation failed: ${issues.join(', ')}`);
    }
  }

  // Private helper methods
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