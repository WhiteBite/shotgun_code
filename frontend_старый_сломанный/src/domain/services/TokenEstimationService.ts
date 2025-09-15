/**
 * TokenEstimationService Domain Service
 * 
 * Provides accurate token estimation for different programming languages
 * and content types. Encapsulates business logic for token calculation
 * following DDD principles.
 */

import type { TokenEstimationConfiguration } from '@/types/configuration';
import { APP_CONFIG } from '@/config/app-config';

export interface TokenEstimationResult {
  readonly estimatedTokens: number;
  readonly actualLength: number;
  readonly charsPerToken: number;
  readonly language: string;
  readonly confidence: number;
}

export interface TokenEstimationOptions {
  readonly language?: string;
  readonly stripComments?: boolean;
  readonly includeWhitespace?: boolean;
  readonly applyLanguageMultiplier?: boolean;
}

/**
 * TokenEstimationService domain service
 * 
 * Provides sophisticated token estimation based on:
 * - Content analysis
 * - Language-specific patterns
 * - Whitespace and comment handling
 * - Statistical modeling
 */
export class TokenEstimationService {
  private readonly config: TokenEstimationConfiguration;
  
  // Language-specific multipliers for more accurate estimation
  private readonly languageMultipliers: Record<string, number> = {
    'javascript': 1.0,
    'typescript': 1.1,
    'python': 0.9,
    'go': 1.0,
    'java': 1.2,
    'csharp': 1.2,
    'cpp': 1.1,
    'c': 1.0,
    'rust': 1.1,
    'json': 0.8,
    'yaml': 0.7,
    'markdown': 0.6,
    'text': 0.8,
    'html': 0.9,
    'css': 0.8,
    'xml': 0.9,
    'sql': 0.9
  };

  constructor(config?: Partial<TokenEstimationConfiguration>) {
    this.config = {
      charsPerToken: APP_CONFIG.performance.tokens,
      languageMultipliers: this.languageMultipliers,
      ...config
    };
  }

  /**
   * Estimates tokens for the given text with language-specific optimizations
   */
  estimateTokens(text: string, options: TokenEstimationOptions = {}): TokenEstimationResult {
    const {
      language = 'text',
      stripComments = false,
      includeWhitespace = true,
      applyLanguageMultiplier = true
    } = options;

    // Preprocess text if needed
    let processedText = text;
    if (stripComments) {
      processedText = this.stripComments(processedText, language);
    }
    if (!includeWhitespace) {
      processedText = this.normalizeWhitespace(processedText);
    }

    // Get base characters per token
    const baseCharsPerToken = this.getCharsPerToken(language);
    
    // Apply language-specific analysis
    const languageAnalysis = this.analyzeLanguageSpecificPatterns(processedText, language);
    
    // Calculate estimated tokens
    const baseEstimate = Math.ceil(processedText.length / baseCharsPerToken);
    
    // Apply language multiplier if enabled
    let finalEstimate = baseEstimate;
    if (applyLanguageMultiplier) {
      const multiplier = this.getLanguageMultiplier(language);
      finalEstimate = Math.ceil(baseEstimate * multiplier);
    }
    
    // Apply pattern-based adjustments
    finalEstimate = Math.ceil(finalEstimate * languageAnalysis.complexityMultiplier);
    
    // Apply global estimation multiplier for safety margin
    finalEstimate = Math.ceil(finalEstimate * this.config.charsPerToken.ESTIMATION_MULTIPLIER);

    return {
      estimatedTokens: Math.max(1, finalEstimate), // Minimum 1 token
      actualLength: text.length,
      charsPerToken: baseCharsPerToken,
      language,
      confidence: languageAnalysis.confidence
    };
  }

  /**
   * Gets characters per token for a specific language
   */
  getCharsPerToken(language: string): number {
    // Use language-specific chars per token if available
    const languageSpecific = this.getLanguageSpecificCharsPerToken(language);
    if (languageSpecific > 0) {
      return languageSpecific;
    }
    
    return this.config.charsPerToken.DEFAULT_CHARS_PER_TOKEN;
  }

  /**
   * Estimates tokens for multiple text segments
   */
  estimateTokensForSegments(segments: { text: string; language?: string }[]): TokenEstimationResult[] {
    return segments.map(segment => 
      this.estimateTokens(segment.text, { language: segment.language })
    );
  }

  /**
   * Gets the total token estimate for multiple segments
   */
  getTotalTokenEstimate(segments: { text: string; language?: string }[]): number {
    const results = this.estimateTokensForSegments(segments);
    return results.reduce((total, result) => total + result.estimatedTokens, 0);
  }

  /**
   * Checks if content exceeds token limits
   */
  exceedsTokenLimit(text: string, options: TokenEstimationOptions = {}): boolean {
    const result = this.estimateTokens(text, options);
    return result.estimatedTokens > this.config.charsPerToken.MAX_TOKENS_PER_REQUEST;
  }

  /**
   * Applies language-specific rules for token estimation
   */
  private applyLanguageSpecificRules(text: string, language: string): number {
    switch (language.toLowerCase()) {
      case 'json':
        return this.estimateJsonTokens(text);
      case 'yaml':
      case 'yml':
        return this.estimateYamlTokens(text);
      case 'markdown':
      case 'md':
        return this.estimateMarkdownTokens(text);
      case 'csv':
        return this.estimateCsvTokens(text);
      default:
        return this.estimateCodeTokens(text, language);
    }
  }

  /**
   * Analyzes language-specific patterns to adjust estimation
   */
  private analyzeLanguageSpecificPatterns(text: string, language: string): {
    complexityMultiplier: number;
    confidence: number;
  } {
    let complexityMultiplier = 1.0;
    let confidence = 0.8;

    // Analyze different aspects of the code
    const analysis = {
      hasComplexStructures: this.hasComplexStructures(text, language),
      hasLongIdentifiers: this.hasLongIdentifiers(text),
      hasStringLiterals: this.hasStringLiterals(text),
      hasComments: this.hasComments(text, language),
      indentationLevel: this.getAverageIndentationLevel(text)
    };

    // Adjust multiplier based on complexity
    if (analysis.hasComplexStructures) {
      complexityMultiplier *= 1.1;
      confidence += 0.1;
    }
    
    if (analysis.hasLongIdentifiers) {
      complexityMultiplier *= 1.05;
    }
    
    if (analysis.hasStringLiterals) {
      complexityMultiplier *= 0.95; // String literals are often more token-efficient
    }
    
    if (analysis.hasComments) {
      complexityMultiplier *= 0.9; // Comments are typically token-efficient
    }
    
    // High indentation suggests complex structure
    if (analysis.indentationLevel > 4) {
      complexityMultiplier *= 1.05;
    }

    return {
      complexityMultiplier: Math.max(0.5, Math.min(2.0, complexityMultiplier)),
      confidence: Math.max(0.5, Math.min(1.0, confidence))
    };
  }

  /**
   * Gets language-specific characters per token
   */
  private getLanguageSpecificCharsPerToken(language: string): number {
    const baseChars = this.config.charsPerToken.DEFAULT_CHARS_PER_TOKEN;
    
    switch (language.toLowerCase()) {
      case 'json':
        return baseChars * 0.8; // JSON is more token-dense
      case 'yaml':
      case 'yml':
        return baseChars * 1.2; // YAML has more whitespace
      case 'markdown':
      case 'md':
        return baseChars * 1.5; // Markdown has lots of formatting
      case 'python':
        return baseChars * 1.1; // Python tends to be more verbose
      case 'java':
      case 'csharp':
        return baseChars * 0.9; // More concise syntax
      default:
        return 0; // Use default
    }
  }

  /**
   * Gets language multiplier for estimation adjustment
   */
  private getLanguageMultiplier(language: string): number {
    return this.config.languageMultipliers[language.toLowerCase()] || 1.0;
  }

  /**
   * Strips comments from code based on language
   */
  private stripComments(text: string, language: string): string {
    switch (language.toLowerCase()) {
      case 'javascript':
      case 'typescript':
      case 'java':
      case 'csharp':
      case 'cpp':
      case 'go':
        return text
          .replace(/\/\*[\s\S]*?\*\//g, '') // Block comments
          .replace(/\/\/.*$/gm, ''); // Line comments
      case 'python':
        return text
          .replace(/"""[\s\S]*?"""/g, '') // Triple quotes
          .replace(/#.*$/gm, ''); // Line comments
      case 'html':
      case 'xml':
        return text.replace(/<!--[\s\S]*?-->/g, '');
      case 'css':
        return text.replace(/\/\*[\s\S]*?\*\//g, '');
      default:
        return text;
    }
  }

  /**
   * Normalizes whitespace for more accurate token counting
   */
  private normalizeWhitespace(text: string): string {
    return text
      .replace(/\s+/g, ' ') // Multiple spaces to single space
      .replace(/^\s+|\s+$/gm, '') // Trim lines
      .trim();
  }

  /**
   * Checks for complex language structures
   */
  private hasComplexStructures(text: string, language: string): boolean {
    const complexPatterns = {
      javascript: [/class\s+\w+/, /function\*/, /async\s+function/, /=>/],
      typescript: [/interface\s+\w+/, /type\s+\w+/, /<.*>/],
      python: [/class\s+\w+/, /def\s+\w+/, /lambda/, /@\w+/],
      java: [/@\w+/, /public\s+class/, /implements/, /extends/],
      go: [/func\s+\w+/, /interface/, /struct/, /goroutine/]
    };

    const patterns = complexPatterns[language.toLowerCase() as keyof typeof complexPatterns] || [];
    return patterns.some(pattern => pattern.test(text));
  }

  /**
   * Checks for long identifiers that affect tokenization
   */
  private hasLongIdentifiers(text: string): boolean {
    const identifiers = text.match(/\b[a-zA-Z_][a-zA-Z0-9_]{8,}\b/g);
    return identifiers ? identifiers.length > 5 : false;
  }

  /**
   * Checks for string literals
   */
  private hasStringLiterals(text: string): boolean {
    return /"[^"]*"|'[^']*'|`[^`]*`/.test(text);
  }

  /**
   * Checks for comments
   */
  private hasComments(text: string, language: string): boolean {
    switch (language.toLowerCase()) {
      case 'javascript':
      case 'typescript':
      case 'java':
      case 'csharp':
      case 'cpp':
      case 'go':
        return /\/\/|\/\*/.test(text);
      case 'python':
        return /#|"""/.test(text);
      case 'html':
      case 'xml':
        return /<!--/.test(text);
      default:
        return false;
    }
  }

  /**
   * Gets average indentation level
   */
  private getAverageIndentationLevel(text: string): number {
    const lines = text.split('\n');
    const indentations = lines
      .map(line => line.match(/^\s*/)?.[0]?.length || 0)
      .filter(indent => indent > 0);
    
    if (indentations.length === 0) return 0;
    
    const average = indentations.reduce((sum, indent) => sum + indent, 0) / indentations.length;
    return Math.round(average / 2); // Assuming 2-space indentation
  }

  // Specialized estimation methods for different formats
  private estimateJsonTokens(text: string): number {
    // JSON tokens are typically shorter due to structure
    return Math.ceil(text.length / (this.config.charsPerToken.DEFAULT_CHARS_PER_TOKEN * 0.8));
  }

  private estimateYamlTokens(text: string): number {
    // YAML has more whitespace, so fewer tokens per character
    return Math.ceil(text.length / (this.config.charsPerToken.DEFAULT_CHARS_PER_TOKEN * 1.2));
  }

  private estimateMarkdownTokens(text: string): number {
    // Markdown formatting affects tokenization
    return Math.ceil(text.length / (this.config.charsPerToken.DEFAULT_CHARS_PER_TOKEN * 1.5));
  }

  private estimateCsvTokens(text: string): number {
    // CSV is structured, count fields and rows
    const rows = text.split('\n');
    const fields = rows.reduce((total, row) => total + row.split(',').length, 0);
    return Math.max(fields, Math.ceil(text.length / this.config.charsPerToken.DEFAULT_CHARS_PER_TOKEN));
  }

  private estimateCodeTokens(text: string, language: string): number {
    // Generic code token estimation with language multiplier
    const baseTokens = Math.ceil(text.length / this.config.charsPerToken.DEFAULT_CHARS_PER_TOKEN);
    const multiplier = this.getLanguageMultiplier(language);
    return Math.ceil(baseTokens * multiplier);
  }
}

/**
 * Default token estimation service instance
 */
export const defaultTokenEstimationService = new TokenEstimationService();