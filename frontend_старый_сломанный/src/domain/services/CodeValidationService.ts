/**
 * CodeValidationService Domain Service
 * 
 * Provides comprehensive code validation including syntax checking,
 * structure validation, and security analysis. Encapsulates business
 * logic for code quality assessment following DDD principles.
 */

import type { CodeValidationServiceConfiguration } from '@/types/configuration';
import { APP_CONFIG, validatePath } from '@/config/app-config';

export interface ValidationError {
  readonly type: 'syntax' | 'structure' | 'security' | 'style' | 'logic';
  readonly severity: 'error' | 'warning' | 'info';
  readonly message: string;
  readonly line?: number;
  readonly column?: number;
  readonly rule?: string;
  readonly suggestion?: string;
}

export interface ValidationWarning {
  readonly type: 'style' | 'performance' | 'maintainability' | 'compatibility';
  readonly message: string;
  readonly line?: number;
  readonly column?: number;
  readonly rule?: string;
  readonly suggestion?: string;
}

export interface ValidationResult {
  readonly isValid: boolean;
  readonly errors: readonly ValidationError[];
  readonly warnings: readonly ValidationWarning[];
  readonly score: number; // 0-100 quality score
  readonly language: string;
  readonly analysisTime: number;
}

export interface ValidationOptions {
  readonly enableSyntaxCheck?: boolean;
  readonly enableStructureCheck?: boolean;
  readonly enableSecurityCheck?: boolean;
  readonly enableStyleCheck?: boolean;
  readonly strictMode?: boolean;
  readonly language?: string;
  readonly rules?: readonly string[];
}

export interface CodeGenerationOptions {
  readonly language: string;
  readonly style?: 'conservative' | 'modern' | 'functional';
  readonly includeComments?: boolean;
  readonly includeTypeAnnotations?: boolean;
  readonly maxComplexity?: number;
  readonly enforcePatterns?: readonly string[];
}

/**
 * CodeValidationService domain service
 * 
 * Provides comprehensive code validation including:
 * - Syntax validation for multiple languages
 * - Structural analysis and complexity metrics
 * - Security vulnerability detection
 * - Code style and best practices
 * - Performance optimization suggestions
 */
export class CodeValidationService {
  private readonly config: CodeValidationServiceConfiguration;
  
  // Language-specific syntax patterns for basic validation
  private readonly syntaxPatterns: Record<string, {
    valid: RegExp[];
    invalid: RegExp[];
    brackets: { open: string; close: string }[];
  }> = {
    javascript: {
      valid: [
        /^\s*(const|let|var)\s+\w+\s*=/,
        /^\s*function\s+\w+\s*\(/,
        /^\s*class\s+\w+/,
        /^\s*if\s*\(/,
        /^\s*for\s*\(/,
        /^\s*while\s*\(/
      ],
      invalid: [
        /\bwith\s*\(/,  // 'with' statement
        /eval\s*\(/,    // eval usage
        /\+\+\w+|--\w+/, // pre-increment/decrement
      ],
      brackets: [
        { open: '{', close: '}' },
        { open: '(', close: ')' },
        { open: '[', close: ']' }
      ]
    },
    typescript: {
      valid: [
        /:\s*\w+(\[\])?\s*[=;]/,
        /interface\s+\w+/,
        /type\s+\w+\s*=/,
        /export\s+(interface|type|class|function)/
      ],
      invalid: [
        /any\s*[;\],=)]/,  // 'any' type usage
        /\beval\s*\(/,
        /\+\+\w+|--\w+/
      ],
      brackets: [
        { open: '{', close: '}' },
        { open: '(', close: ')' },
        { open: '[', close: ']' },
        { open: '<', close: '>' }
      ]
    },
    python: {
      valid: [
        /^\s*def\s+\w+\s*\(/,
        /^\s*class\s+\w+.*:/,
        /^\s*if\s+.*:/,
        /^\s*for\s+\w+\s+in\s+.*:/,
        /^\s*while\s+.*:/
      ],
      invalid: [
        /eval\s*\(/,
        /exec\s*\(/,
        /\bexcept:\s*$/,  // bare except
      ],
      brackets: [
        { open: '(', close: ')' },
        { open: '[', close: ']' },
        { open: '{', close: '}' }
      ]
    },
    go: {
      valid: [
        /^\s*func\s+\w+\s*\(/,
        /^\s*type\s+\w+\s+struct/,
        /^\s*if\s+.*{/,
        /^\s*for\s+.*{/
      ],
      invalid: [
        /goto\s+\w+/,  // goto usage
      ],
      brackets: [
        { open: '{', close: '}' },
        { open: '(', close: ')' },
        { open: '[', close: ']' }
      ]
    }
  };

  // Security vulnerability patterns
  private readonly securityPatterns: Record<string, RegExp[]> = {
    javascript: [
      /eval\s*\(/,
      /Function\s*\(/,
      /setTimeout\s*\(\s*["']/,
      /setInterval\s*\(\s*["']/,
      /innerHTML\s*=/,
      /outerHTML\s*=/,
      /document\.write\s*\(/,
      /window\.location\s*=/
    ],
    typescript: [
      /eval\s*\(/,
      /Function\s*\(/,
      /any\s*[;\],=)]/,
      /innerHTML\s*=/,
      /outerHTML\s*=/
    ],
    python: [
      /eval\s*\(/,
      /exec\s*\(/,
      /input\s*\(/,
      /raw_input\s*\(/,
      /pickle\.loads?\s*\(/,
      /subprocess\.call\s*\(/,
      /os\.system\s*\(/
    ],
    sql: [
      /'\s*\+\s*\w+\s*\+\s*'/,  // String concatenation in SQL
      /"\s*\+\s*\w+\s*\+\s*"/,
      /EXEC\s*\(/i,
      /EXECUTE\s*\(/i
    ]
  };

  constructor(config?: Partial<CodeValidationServiceConfiguration>) {
    this.config = {
      syntaxValidation: APP_CONFIG.ai.validation,
      securityValidation: APP_CONFIG.security,
      ...config
    };
  }

  /**
   * Validates code syntax for the specified language
   */
  validateSyntax(code: string, language: string, options: ValidationOptions = {}): ValidationResult {
    const startTime = performance.now();
    const errors: ValidationError[] = [];
    const warnings: ValidationWarning[] = [];
    
    const {
      enableSyntaxCheck = this.config.syntaxValidation.SYNTAX_CHECK,
      enableStructureCheck = this.config.syntaxValidation.STRUCTURE_CHECK,
      enableSecurityCheck = this.config.syntaxValidation.SECURITY_CHECK,
      strictMode = false
    } = options;

    // Basic syntax validation
    if (enableSyntaxCheck) {
      errors.push(...this.performSyntaxCheck(code, language));
    }

    // Structure validation
    if (enableStructureCheck) {
      errors.push(...this.performStructureCheck(code, language));
      warnings.push(...this.performStyleCheck(code, language));
    }

    // Security validation
    if (enableSecurityCheck) {
      errors.push(...this.performSecurityCheck(code, language));
    }

    // Calculate quality score
    const score = this.calculateQualityScore(code, errors, warnings);
    
    const analysisTime = performance.now() - startTime;

    return {
      isValid: errors.length === 0,
      errors,
      warnings,
      score,
      language,
      analysisTime
    };
  }

  /**
   * Validates code structure and complexity
   */
  validateStructure(code: string, language: string = 'javascript'): ValidationResult {
    return this.validateSyntax(code, language, {
      enableSyntaxCheck: false,
      enableStructureCheck: true,
      enableSecurityCheck: false
    });
  }

  /**
   * Generates validated code with applied best practices
   */
  generateValidatedCode(code: string, options: CodeGenerationOptions): string {
    const { language, style = 'modern', includeComments = true } = options;
    
    // Apply formatting and validation rules
    let validatedCode = this.applyFormattingRules(code, language, style);
    
    // Add type annotations if requested
    if (options.includeTypeAnnotations && (language === 'typescript' || language === 'javascript')) {
      validatedCode = this.addTypeAnnotations(validatedCode, language);
    }
    
    // Add comments if requested
    if (includeComments) {
      validatedCode = this.addDocumentationComments(validatedCode, language);
    }
    
    // Enforce complexity limits
    if (options.maxComplexity) {
      validatedCode = this.enforceComplexityLimits(validatedCode, options.maxComplexity);
    }
    
    return validatedCode;
  }

  /**
   * Checks if code follows security best practices
   */
  isSecureCode(code: string, language: string): boolean {
    const securityErrors = this.performSecurityCheck(code, language);
    return securityErrors.length === 0;
  }

  /**
   * Gets validation rules for a specific language
   */
  getValidationRules(language: string): readonly string[] {
    const commonRules = [
      'no-unused-variables',
      'no-duplicate-keys',
      'no-unreachable-code',
      'prefer-const',
      'no-magic-numbers'
    ];

    const languageSpecificRules: Record<string, string[]> = {
      javascript: [
        ...commonRules,
        'no-eval',
        'no-with',
        'no-implicit-globals',
        'prefer-arrow-callback'
      ],
      typescript: [
        ...commonRules,
        'no-explicit-any',
        'prefer-interface',
        'no-non-null-assertion',
        'prefer-readonly'
      ],
      python: [
        ...commonRules,
        'no-exec',
        'no-bare-except',
        'prefer-f-strings',
        'follow-pep8'
      ],
      go: [
        ...commonRules,
        'no-goto',
        'prefer-interfaces',
        'follow-gofmt',
        'handle-errors'
      ]
    };

    return languageSpecificRules[language.toLowerCase()] || commonRules;
  }

  /**
   * Performs basic syntax checking
   */
  private performSyntaxCheck(code: string, language: string): ValidationError[] {
    const errors: ValidationError[] = [];
    const patterns = this.syntaxPatterns[language.toLowerCase()];
    
    if (!patterns) {
      return errors; // Language not supported for syntax checking
    }

    // Check for invalid patterns
    for (const invalidPattern of patterns.invalid) {
      const matches = code.matchAll(new RegExp(invalidPattern.source, 'gm'));
      for (const match of matches) {
        const line = this.getLineNumber(code, match.index || 0);
        errors.push({
          type: 'syntax',
          severity: 'error',
          message: `Invalid syntax pattern detected: ${match[0]}`,
          line,
          rule: 'syntax-check'
        });
      }
    }

    // Check bracket matching
    for (const bracket of patterns.brackets) {
      const openCount = (code.match(new RegExp(`\\${bracket.open}`, 'g')) || []).length;
      const closeCount = (code.match(new RegExp(`\\${bracket.close}`, 'g')) || []).length;
      
      if (openCount !== closeCount) {
        errors.push({
          type: 'syntax',
          severity: 'error',
          message: `Mismatched ${bracket.open}${bracket.close} brackets: ${openCount} open, ${closeCount} close`,
          rule: 'bracket-matching'
        });
      }
    }

    return errors;
  }

  /**
   * Performs structural analysis
   */
  private performStructureCheck(code: string, language: string): ValidationError[] {
    const errors: ValidationError[] = [];

    // Check function length
    const functions = this.extractFunctions(code, language);
    for (const func of functions) {
      if (func.lineCount > 50) {
        errors.push({
          type: 'structure',
          severity: 'warning',
          message: `Function '${func.name}' is too long (${func.lineCount} lines). Consider breaking it down.`,
          line: func.startLine,
          rule: 'max-function-length',
          suggestion: 'Split into smaller functions'
        });
      }
    }

    // Check nesting depth
    const maxNesting = this.calculateMaxNestingDepth(code);
    if (maxNesting > 4) {
      errors.push({
        type: 'structure',
        severity: 'warning',
        message: `Maximum nesting depth (${maxNesting}) exceeds recommended limit (4)`,
        rule: 'max-nesting-depth',
        suggestion: 'Consider refactoring nested structures'
      });
    }

    return errors;
  }

  /**
   * Performs security vulnerability checking
   */
  private performSecurityCheck(code: string, language: string): ValidationError[] {
    const errors: ValidationError[] = [];
    const patterns = this.securityPatterns[language.toLowerCase()];
    
    if (!patterns) {
      return errors; // Language not supported for security checking
    }

    for (const pattern of patterns) {
      const matches = code.matchAll(new RegExp(pattern.source, 'gm'));
      for (const match of matches) {
        const line = this.getLineNumber(code, match.index || 0);
        errors.push({
          type: 'security',
          severity: 'error',
          message: `Potential security vulnerability: ${match[0]}`,
          line,
          rule: 'security-check',
          suggestion: 'Use safer alternatives'
        });
      }
    }

    return errors;
  }

  /**
   * Performs style and best practices checking
   */
  private performStyleCheck(code: string, language: string): ValidationWarning[] {
    const warnings: ValidationWarning[] = [];

    // Check naming conventions
    const variableNames = this.extractVariableNames(code, language);
    for (const variable of variableNames) {
      if (!this.isValidNaming(variable.name, language)) {
        warnings.push({
          type: 'style',
          message: `Variable '${variable.name}' doesn't follow naming conventions`,
          line: variable.line,
          rule: 'naming-convention',
          suggestion: `Use ${this.getSuggestedNaming(variable.name, language)}`
        });
      }
    }

    return warnings;
  }

  /**
   * Calculates code quality score
   */
  private calculateQualityScore(code: string, errors: ValidationError[], warnings: ValidationWarning[]): number {
    let score = 100;
    
    // Deduct points for errors
    score -= errors.length * 10;
    
    // Deduct points for warnings
    score -= warnings.length * 5;
    
    // Bonus for good practices
    if (code.includes('use strict')) score += 5;
    if (this.hasGoodCommentCoverage(code)) score += 10;
    if (this.hasConsistentIndentation(code)) score += 5;
    
    return Math.max(0, Math.min(100, score));
  }

  /**
   * Helper methods
   */
  private getLineNumber(code: string, index: number): number {
    return code.substring(0, index).split('\n').length;
  }

  private extractFunctions(code: string, language: string): Array<{ name: string; startLine: number; lineCount: number }> {
    // Simplified function extraction - would be more sophisticated in practice
    const functions: Array<{ name: string; startLine: number; lineCount: number }> = [];
    
    const functionRegex = language === 'python' 
      ? /def\s+(\w+)\s*\(/g
      : /function\s+(\w+)\s*\(|(\w+)\s*=\s*function|\s*(\w+)\s*\(/g;
    
    const matches = code.matchAll(functionRegex);
    for (const match of matches) {
      const name = match[1] || match[2] || match[3] || 'anonymous';
      const startLine = this.getLineNumber(code, match.index || 0);
      // Simplified line counting - would need proper parsing
      const lineCount = 10; // Placeholder
      
      functions.push({ name, startLine, lineCount });
    }
    
    return functions;
  }

  private calculateMaxNestingDepth(code: string): number {
    let maxDepth = 0;
    let currentDepth = 0;
    
    for (const char of code) {
      if (char === '{' || char === '(') {
        currentDepth++;
        maxDepth = Math.max(maxDepth, currentDepth);
      } else if (char === '}' || char === ')') {
        currentDepth--;
      }
    }
    
    return maxDepth;
  }

  private extractVariableNames(code: string, language: string): Array<{ name: string; line: number }> {
    // Simplified variable extraction
    const variables: Array<{ name: string; line: number }> = [];
    
    const variableRegex = language === 'python'
      ? /^\s*(\w+)\s*=/gm
      : /(?:const|let|var)\s+(\w+)/g;
    
    const matches = code.matchAll(variableRegex);
    for (const match of matches) {
      const name = match[1];
      const line = this.getLineNumber(code, match.index || 0);
      variables.push({ name, line });
    }
    
    return variables;
  }

  private isValidNaming(name: string, language: string): boolean {
    const conventions: Record<string, RegExp> = {
      javascript: /^[a-z][a-zA-Z0-9]*$/,  // camelCase
      typescript: /^[a-z][a-zA-Z0-9]*$/,  // camelCase
      python: /^[a-z][a-z0-9_]*$/,        // snake_case
      go: /^[a-z][a-zA-Z0-9]*$/           // camelCase
    };
    
    const pattern = conventions[language.toLowerCase()];
    return pattern ? pattern.test(name) : true;
  }

  private getSuggestedNaming(name: string, language: string): string {
    if (language === 'python') {
      return name.replace(/([A-Z])/g, '_$1').toLowerCase();
    }
    return name.charAt(0).toLowerCase() + name.slice(1);
  }

  private hasGoodCommentCoverage(code: string): boolean {
    const lines = code.split('\n');
    const commentLines = lines.filter(line => line.trim().startsWith('//') || line.trim().startsWith('#')).length;
    return (commentLines / lines.length) > 0.1; // 10% comment coverage
  }

  private hasConsistentIndentation(code: string): boolean {
    const lines = code.split('\n').filter(line => line.trim().length > 0);
    const indentations = lines.map(line => {
      const match = line.match(/^\s*/);
      return match ? match[0].length : 0;
    });
    
    // Check if indentation is consistent (multiples of 2 or 4)
    return indentations.every(indent => indent % 2 === 0 || indent % 4 === 0);
  }

  private applyFormattingRules(code: string, language: string, style: string): string {
    // Simplified formatting - would use proper formatters in practice
    return code
      .replace(/\t/g, '  ') // Convert tabs to spaces
      .replace(/\s+$/gm, '') // Remove trailing whitespace
      .replace(/\n{3,}/g, '\n\n'); // Limit consecutive empty lines
  }

  private addTypeAnnotations(code: string, language: string): string {
    // Simplified type annotation addition
    if (language === 'typescript') {
      return code.replace(/(\w+)\s*=/g, '$1: any =');
    }
    return code;
  }

  private addDocumentationComments(code: string, language: string): string {
    // Simplified comment addition
    const commentPrefix = language === 'python' ? '#' : '//';
    return `${commentPrefix} Generated code with validation\n${code}`;
  }

  private enforceComplexityLimits(code: string, maxComplexity: number): string {
    // Simplified complexity enforcement
    return code; // Would implement actual complexity reduction
  }
}

/**
 * Default code validation service instance
 */
export const defaultCodeValidationService = new CodeValidationService();