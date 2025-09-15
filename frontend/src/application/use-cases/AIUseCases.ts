import type { AIRepository } from '@/domain/repositories/AIRepository';
import type { 
  AICodeGenerationRequest, 
  AICodeGenerationResponse,
  AutonomousTaskRequest,
  AutonomousTaskResponse,
  AutonomousTaskStatus
} from '@/types/dto';

/**
 * AI Code Generation Use Cases
 */

export class GenerateCodeUseCase {
  constructor(private aiRepo: AIRepository) {}

  async execute(request: AICodeGenerationRequest): Promise<AICodeGenerationResponse> {
    this.validateCodeGenerationRequest(request);
    
    try {
      const response = await this.aiRepo.generateCode(request);
      
      // Validate response
      this.validateCodeGenerationResponse(response);
      
      return response;
    } catch (error) {
      throw new Error(`Code generation failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateCodeGenerationRequest(request: AICodeGenerationRequest): void {
    if (!request.prompt?.trim()) {
      throw new Error('Prompt is required for code generation');
    }
    
    if (request.prompt.length > 10000) {
      throw new Error('Prompt is too long (maximum 10,000 characters)');
    }
    
    if (request.maxTokens && (request.maxTokens < 1 || request.maxTokens > 8000)) {
      throw new Error('Max tokens must be between 1 and 8000');
    }
    
    if (request.temperature && (request.temperature < 0 || request.temperature > 2)) {
      throw new Error('Temperature must be between 0 and 2');
    }
  }

  private validateCodeGenerationResponse(response: AICodeGenerationResponse): void {
    if (!response.code && !response.analysis) {
      throw new Error('Invalid response: no code or analysis generated');
    }
    
    if (response.code && response.code.length > 50000) {
      console.warn('Generated code is very large, consider breaking it into smaller parts');
    }
  }
}

export class AnalyzeCodeUseCase {
  constructor(private aiRepo: AIRepository) {}

  async execute(
    code: string, 
    analysisType: 'security' | 'performance' | 'quality' | 'bugs' | 'general' = 'general',
    options?: {
      language?: string;
      frameworks?: string[];
      focus?: string[];
    }
  ): Promise<{
    analysis: string;
    suggestions: string[];
    issues: Array<{
      type: 'error' | 'warning' | 'info';
      message: string;
      line?: number;
      severity: 'high' | 'medium' | 'low';
    }>;
    score?: number;
  }> {
    this.validateCodeAnalysisInputs(code, analysisType);
    
    try {
      const response = await this.aiRepo.analyzeCode(code, analysisType, options);
      
      return {
        analysis: response.analysis || 'No analysis available',
        suggestions: response.suggestions || [],
        issues: response.issues || [],
        score: response.score
      };
    } catch (error) {
      throw new Error(`Code analysis failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateCodeAnalysisInputs(code: string, analysisType: string): void {
    if (!code?.trim()) {
      throw new Error('Code is required for analysis');
    }
    
    if (code.length > 100000) {
      throw new Error('Code is too large for analysis (maximum 100,000 characters)');
    }
    
    const validAnalysisTypes = ['security', 'performance', 'quality', 'bugs', 'general'];
    if (!validAnalysisTypes.includes(analysisType)) {
      throw new Error(`Invalid analysis type: ${analysisType}`);
    }
  }
}

export class ExplainCodeUseCase {
  constructor(private aiRepo: AIRepository) {}

  async execute(
    code: string,
    options?: {
      language?: string;
      focus?: 'overview' | 'detailed' | 'beginner';
      includeExamples?: boolean;
    }
  ): Promise<{
    explanation: string;
    keyPoints: string[];
    complexity: 'low' | 'medium' | 'high';
    technologies: string[];
    examples?: string[];
  }> {
    this.validateCodeExplanationInputs(code);
    
    try {
      const response = await this.aiRepo.explainCode(code, options);
      
      return {
        explanation: response.explanation || 'No explanation available',
        keyPoints: response.keyPoints || [],
        complexity: response.complexity || 'medium',
        technologies: response.technologies || [],
        examples: response.examples
      };
    } catch (error) {
      throw new Error(`Code explanation failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateCodeExplanationInputs(code: string): void {
    if (!code?.trim()) {
      throw new Error('Code is required for explanation');
    }
    
    if (code.length > 50000) {
      throw new Error('Code is too large for explanation (maximum 50,000 characters)');
    }
  }
}

/**
 * AI Task Suggestion Use Cases
 */

export class SuggestContextFilesUseCase {
  constructor(private aiRepo: AIRepository) {}

  async execute(
    task: string,
    availableFiles: string[],
    options?: {
      maxSuggestions?: number;
      includeTests?: boolean;
      includeConfig?: boolean;
    }
  ): Promise<{
    suggestedFiles: Array<{
      path: string;
      reason: string;
      priority: 'high' | 'medium' | 'low';
    }>;
    reasoning: string;
  }> {
    this.validateSuggestionInputs(task, availableFiles);
    
    try {
      const maxSuggestions = options?.maxSuggestions || 10;
      const suggestions = await this.aiRepo.suggestContextFiles(task, availableFiles, options);
      
      // Limit and prioritize suggestions
      const limitedSuggestions = suggestions.slice(0, maxSuggestions);
      
      return {
        suggestedFiles: limitedSuggestions,
        reasoning: `Based on the task "${task}", these files are most relevant for providing context.`
      };
    } catch (error) {
      throw new Error(`File suggestion failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateSuggestionInputs(task: string, availableFiles: string[]): void {
    if (!task?.trim()) {
      throw new Error('Task description is required');
    }
    
    if (!availableFiles || availableFiles.length === 0) {
      throw new Error('Available files list is required');
    }
    
    if (availableFiles.length > 1000) {
      throw new Error('Too many files to analyze (maximum 1000)');
    }
  }
}

export class AnalyzeTaskAndCollectContextUseCase {
  constructor(private aiRepo: AIRepository) {}

  async execute(
    task: string,
    projectPath: string,
    options?: {
      includeGitHistory?: boolean;
      maxContextFiles?: number;
      analysisDepth?: 'shallow' | 'medium' | 'deep';
    }
  ): Promise<{
    taskAnalysis: {
      complexity: 'low' | 'medium' | 'high';
      estimatedTime: string;
      requiredSkills: string[];
      risks: string[];
    };
    suggestedContext: {
      files: string[];
      reasoning: string;
    };
    recommendations: string[];
  }> {
    this.validateTaskAnalysisInputs(task, projectPath);
    
    try {
      const response = await this.aiRepo.analyzeTaskAndCollectContext(task, projectPath, options);
      
      return {
        taskAnalysis: {
          complexity: response.complexity || 'medium',
          estimatedTime: response.estimatedTime || 'Unknown',
          requiredSkills: response.requiredSkills || [],
          risks: response.risks || []
        },
        suggestedContext: {
          files: response.suggestedFiles || [],
          reasoning: response.reasoning || 'No reasoning provided'
        },
        recommendations: response.recommendations || []
      };
    } catch (error) {
      throw new Error(`Task analysis failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateTaskAnalysisInputs(task: string, projectPath: string): void {
    if (!task?.trim()) {
      throw new Error('Task description is required');
    }
    
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
    
    if (task.length > 5000) {
      throw new Error('Task description is too long (maximum 5000 characters)');
    }
  }
}

/**
 * Autonomous AI Task Management Use Cases
 */

export class StartAutonomousTaskUseCase {
  constructor(private aiRepo: AIRepository) {}

  async execute(request: AutonomousTaskRequest): Promise<AutonomousTaskResponse> {
    this.validateAutonomousTaskRequest(request);
    
    try {
      const response = await this.aiRepo.startAutonomousTask(request);
      
      // Validate response
      if (!response.taskId) {
        throw new Error('Invalid response: no task ID returned');
      }
      
      return response;
    } catch (error) {
      throw new Error(`Failed to start autonomous task: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateAutonomousTaskRequest(request: AutonomousTaskRequest): void {
    if (!request.description?.trim()) {
      throw new Error('Task description is required');
    }
    
    if (!request.projectPath?.trim()) {
      throw new Error('Project path is required');
    }
    
    if (request.description.length > 10000) {
      throw new Error('Task description is too long (maximum 10,000 characters)');
    }
    
    if (request.maxDuration && (request.maxDuration < 60 || request.maxDuration > 7200)) {
      throw new Error('Max duration must be between 1 minute and 2 hours');
    }
  }
}

export class MonitorAutonomousTaskUseCase {
  constructor(private aiRepo: AIRepository) {}

  async execute(taskId: string): Promise<AutonomousTaskStatus> {
    this.validateTaskId(taskId);
    
    try {
      const status = await this.aiRepo.getAutonomousTaskStatus(taskId);
      
      // Add additional monitoring logic
      if (status.status === 'error' && status.error) {
        console.error(`Autonomous task ${taskId} failed:`, status.error);
      }
      
      if (status.status === 'completed' && status.results) {
        console.log(`Autonomous task ${taskId} completed successfully`);
      }
      
      return status;
    } catch (error) {
      throw new Error(`Failed to get task status: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateTaskId(taskId: string): void {
    if (!taskId?.trim()) {
      throw new Error('Task ID is required');
    }
  }
}

export class CancelAutonomousTaskUseCase {
  constructor(private aiRepo: AIRepository) {}

  async execute(taskId: string, reason?: string): Promise<void> {
    this.validateTaskId(taskId);
    
    try {
      await this.aiRepo.cancelAutonomousTask(taskId);
      
      if (reason) {
        console.log(`Autonomous task ${taskId} cancelled: ${reason}`);
      }
    } catch (error) {
      throw new Error(`Failed to cancel task: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateTaskId(taskId: string): void {
    if (!taskId?.trim()) {
      throw new Error('Task ID is required');
    }
  }
}

/**
 * Composite AI Use Cases
 */

export class SmartCodeAssistantUseCase {
  constructor(private aiRepo: AIRepository) {}

  async execute(
    action: 'generate' | 'analyze' | 'explain' | 'optimize' | 'test',
    input: {
      code?: string;
      prompt?: string;
      language?: string;
      context?: string[];
    },
    options?: {
      quality?: 'fast' | 'balanced' | 'high';
      includeTests?: boolean;
      includeComments?: boolean;
    }
  ): Promise<{
    result: string;
    confidence: number;
    suggestions: string[];
    additionalInfo?: {
      estimatedTime?: string;
      complexity?: string;
      technologies?: string[];
    };
  }> {
    this.validateSmartAssistantInputs(action, input);
    
    try {
      let result: string;
      let additionalInfo: any = {};
      
      switch (action) {
        case 'generate':
          if (!input.prompt) throw new Error('Prompt required for code generation');
          const genResponse = await this.aiRepo.generateCode({
            prompt: input.prompt,
            language: input.language,
            includeTests: options?.includeTests,
            includeComments: options?.includeComments
          });
          result = genResponse.code || genResponse.analysis || '';
          additionalInfo = { complexity: genResponse.complexity };
          break;
          
        case 'analyze':
          if (!input.code) throw new Error('Code required for analysis');
          const analysisResponse = await this.aiRepo.analyzeCode(input.code, 'general', { language: input.language });
          result = analysisResponse.analysis || '';
          additionalInfo = { score: analysisResponse.score };
          break;
          
        case 'explain':
          if (!input.code) throw new Error('Code required for explanation');
          const explainResponse = await this.aiRepo.explainCode(input.code, { language: input.language });
          result = explainResponse.explanation || '';
          additionalInfo = { 
            complexity: explainResponse.complexity,
            technologies: explainResponse.technologies 
          };
          break;
          
        default:
          throw new Error(`Unsupported action: ${action}`);
      }
      
      return {
        result,
        confidence: this.calculateConfidence(result, action),
        suggestions: this.generateSuggestions(action, input),
        additionalInfo
      };
    } catch (error) {
      throw new Error(`Smart assistant failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateSmartAssistantInputs(action: string, input: any): void {
    const validActions = ['generate', 'analyze', 'explain', 'optimize', 'test'];
    if (!validActions.includes(action)) {
      throw new Error(`Invalid action: ${action}`);
    }
    
    if (!input || typeof input !== 'object') {
      throw new Error('Input object is required');
    }
  }

  private calculateConfidence(result: string, action: string): number {
    // Simple confidence calculation based on result length and action type
    if (!result || result.length < 10) return 0.1;
    
    const baseConfidence = Math.min(result.length / 1000, 1) * 0.7;
    const actionBonus = action === 'generate' ? 0.2 : 0.1;
    
    return Math.min(baseConfidence + actionBonus, 0.95);
  }

  private generateSuggestions(action: string, input: any): string[] {
    const suggestions: string[] = [];
    
    if (action === 'generate' && input.language) {
      suggestions.push(`Consider adding unit tests for the generated ${input.language} code`);
    }
    
    if (action === 'analyze') {
      suggestions.push('Review the identified issues and consider refactoring');
    }
    
    if (action === 'explain' && input.code && input.code.length > 1000) {
      suggestions.push('Consider breaking down large functions into smaller, more focused ones');
    }
    
    return suggestions;
  }
}