/**
 * LanguageDetectionService Domain Service
 * 
 * Provides intelligent language detection for files and content.
 * Encapsulates business logic for language identification following DDD principles.
 */

import type { LanguageDetectionServiceConfiguration } from '@/types/configuration';
import { APP_CONFIG } from '@/config/app-config';

export interface LanguageDetectionResult {
  readonly language: string;
  readonly confidence: number;
  readonly detectionMethod: 'extension' | 'content' | 'filename' | 'fallback';
  readonly alternativeLanguages?: readonly string[];
}

export interface LanguageDetectionOptions {
  readonly enableContentAnalysis?: boolean;
  readonly enableFilenameAnalysis?: boolean;
  readonly enableFallbackDetection?: boolean;
  readonly preferredLanguages?: readonly string[];
}

/**
 * LanguageDetectionService domain service
 * 
 * Provides sophisticated language detection based on:
 * - File extension mapping
 * - Content analysis and patterns
 * - Filename conventions
 * - Statistical analysis
 */
export class LanguageDetectionService {
  private readonly config: LanguageDetectionServiceConfiguration;
  
  // Content patterns for language detection
  private readonly contentPatterns: Record<string, RegExp[]> = {
    javascript: [
      /\bfunction\s+\w+\s*\(/,
      /\bconst\s+\w+\s*=/,
      /\blet\s+\w+\s*=/,
      /=>\s*{?/,
      /require\s*\(/,
      /module\.exports/,
      /console\.log\s*\(/
    ],
    typescript: [
      /:\s*\w+(\[\])?(\s*\|\s*\w+)*\s*[=;]/,
      /interface\s+\w+/,
      /type\s+\w+\s*=/,
      /import.*from\s*['"].*['"];?/,
      /export\s+(interface|type|class|function)/,
      /<.*>/
    ],
    python: [
      /def\s+\w+\s*\(/,
      /class\s+\w+.*:/,
      /import\s+\w+/,
      /from\s+\w+\s+import/,
      /if\s+__name__\s*==\s*['"]__main__['"]:/,
      /@\w+/,
      /print\s*\(/
    ],
    go: [
      /package\s+\w+/,
      /func\s+\w+\s*\(/,
      /import\s*\(/,
      /type\s+\w+\s+struct/,
      /fmt\.Print/,
      /goroutine/
    ],
    java: [
      /public\s+class\s+\w+/,
      /import\s+[\w.]+;/,
      /public\s+static\s+void\s+main/,
      /@\w+/,
      /System\.out\.print/,
      /extends\s+\w+/,
      /implements\s+\w+/
    ],
    csharp: [
      /using\s+[\w.]+;/,
      /namespace\s+[\w.]+/,
      /public\s+class\s+\w+/,
      /Console\.Write/,
      /\[.*\]/,
      /var\s+\w+\s*=/
    ],
    cpp: [
      /#include\s*<.*>/,
      /std::/,
      /cout\s*<<|cin\s*>>/,
      /class\s+\w+/,
      /template\s*<.*>/,
      /namespace\s+\w+/
    ],
    c: [
      /#include\s*<.*\.h>/,
      /int\s+main\s*\(/,
      /printf\s*\(/,
      /scanf\s*\(/,
      /malloc\s*\(/,
      /free\s*\(/
    ],
    rust: [
      /fn\s+\w+\s*\(/,
      /let\s+mut\s+\w+/,
      /use\s+[\w:]+;/,
      /match\s+\w+/,
      /println!\s*\(/,
      /impl\s+\w+/
    ],
    php: [
      /<\?php/,
      /\$\w+/,
      /echo\s+/,
      /function\s+\w+\s*\(/,
      /class\s+\w+/,
      /require_once|include_once/
    ],
    ruby: [
      /def\s+\w+/,
      /class\s+\w+/,
      /require\s+['"].*['"]$/,
      /puts\s+/,
      /end$/m,
      /@\w+/
    ],
    swift: [
      /func\s+\w+\s*\(/,
      /var\s+\w+:\s*\w+/,
      /let\s+\w+:\s*\w+/,
      /import\s+\w+/,
      /class\s+\w+/,
      /print\s*\(/
    ],
    kotlin: [
      /fun\s+\w+\s*\(/,
      /val\s+\w+:\s*\w+/,
      /var\s+\w+:\s*\w+/,
      /class\s+\w+/,
      /package\s+[\w.]+/,
      /println\s*\(/
    ],
    html: [
      /<html.*>/i,
      /<head.*>/i,
      /<body.*>/i,
      /<div.*>/i,
      /<script.*>/i,
      /<!DOCTYPE\s+html>/i
    ],
    css: [
      /[.\#]?\w+\s*{/,
      /[\w-]+:\s*[\w\s#%-]+;/,
      /@media/,
      /@import/,
      /rgba?\s*\(/,
      /!important/
    ],
    sql: [
      /SELECT\s+.*\s+FROM/i,
      /INSERT\s+INTO/i,
      /UPDATE\s+.*\s+SET/i,
      /DELETE\s+FROM/i,
      /CREATE\s+TABLE/i,
      /ALTER\s+TABLE/i
    ],
    json: [
      /^\s*{/,
      /^\s*\[/,
      /"[\w\s]+"\s*:/,
      /}\s*,?\s*$/m,
      /]\s*,?\s*$/m
    ],
    yaml: [
      /^\s*\w+:\s*/m,
      /^\s*-\s+/m,
      /---\s*$/m,
      /\.\.\.\s*$/m
    ],
    xml: [
      /<\?xml.*\?>/,
      /<\w+.*>/,
      /<\/\w+>/,
      /xmlns/,
      /<!\[CDATA\[/
    ],
    markdown: [
      /^#+ /m,
      /\*\*.*\*\*/,
      /\[.*\]\(.*\)/,
      /```/,
      /^\* /m,
      /^\d+\. /m
    ]
  };

  // Filename patterns for additional detection
  private readonly filenamePatterns: Record<string, RegExp[]> = {
    dockerfile: [/^Dockerfile$/i, /\.dockerfile$/i],
    makefile: [/^Makefile$/i, /\.mk$/i],
    cmake: [/^CMakeLists\.txt$/i, /\.cmake$/i],
    gitignore: [/^\.gitignore$/i],
    yaml: [/\.ya?ml$/i],
    json: [/\.json$/i],
    xml: [/\.xml$/i, /\.plist$/i, /\.xsd$/i],
    properties: [/\.properties$/i],
    ini: [/\.ini$/i, /\.cfg$/i, /\.conf$/i],
    shell: [/\.sh$/i, /\.bash$/i, /\.zsh$/i],
    batch: [/\.bat$/i, /\.cmd$/i],
    powershell: [/\.ps1$/i]
  };

  constructor(config?: Partial<LanguageDetectionServiceConfiguration>) {
    this.config = {
      extensionMap: APP_CONFIG.languageDetection.EXTENSION_MAP,
      defaultLanguage: APP_CONFIG.languageDetection.DEFAULT_LANGUAGE,
      fallbackDetection: APP_CONFIG.languageDetection.FALLBACK_DETECTION,
      ...config
    };
  }

  /**
   * Detects language from file path with comprehensive analysis
   */
  detectLanguage(filePath: string, options: LanguageDetectionOptions = {}): LanguageDetectionResult {
    const {
      enableContentAnalysis = false,
      enableFilenameAnalysis = true,
      enableFallbackDetection = this.config.fallbackDetection,
      preferredLanguages = []
    } = options;

    // Try extension-based detection first
    const extensionResult = this.detectLanguageFromExtension(filePath);
    if (extensionResult.confidence > 0.8) {
      return extensionResult;
    }

    // Try filename pattern detection
    if (enableFilenameAnalysis) {
      const filenameResult = this.detectLanguageFromFilename(filePath);
      if (filenameResult.confidence > 0.9) {
        return filenameResult;
      }
    }

    // If we have a reasonable extension match, return it
    if (extensionResult.confidence > 0.5) {
      return extensionResult;
    }

    // Use fallback detection if enabled
    if (enableFallbackDetection) {
      return {
        language: this.config.defaultLanguage,
        confidence: 0.3,
        detectionMethod: 'fallback'
      };
    }

    return extensionResult;
  }

  /**
   * Detects language from content analysis
   */
  detectLanguageFromContent(content: string, options: LanguageDetectionOptions = {}): LanguageDetectionResult {
    const { preferredLanguages = [] } = options;
    
    const scores: Record<string, number> = {};
    const alternatives: string[] = [];

    // Score each language based on pattern matches
    for (const [language, patterns] of Object.entries(this.contentPatterns)) {
      let score = 0;
      let matchCount = 0;
      
      for (const pattern of patterns) {
        const matches = content.match(pattern);
        if (matches) {
          matchCount++;
          score += matches.length;
        }
      }
      
      // Calculate normalized score
      if (matchCount > 0) {
        scores[language] = (score / patterns.length) * (matchCount / patterns.length);
        alternatives.push(language);
      }
    }

    // Apply preference boost
    for (const preferred of preferredLanguages) {
      if (scores[preferred]) {
        scores[preferred] *= 1.2;
      }
    }

    // Find best match
    const sortedLanguages = Object.entries(scores)
      .sort(([, a], [, b]) => b - a);

    if (sortedLanguages.length === 0) {
      return {
        language: this.config.defaultLanguage,
        confidence: 0.1,
        detectionMethod: 'fallback'
      };
    }

    const [bestLanguage, bestScore] = sortedLanguages[0];
    const confidence = Math.min(0.95, Math.max(0.1, bestScore));

    return {
      language: bestLanguage,
      confidence,
      detectionMethod: 'content',
      alternativeLanguages: alternatives.slice(0, 3)
    };
  }

  /**
   * Gets list of supported languages
   */
  getSupportedLanguages(): string[] {
    const extensionLanguages = Object.values(this.config.extensionMap);
    const contentLanguages = Object.keys(this.contentPatterns);
    const filenameLanguages = Object.keys(this.filenamePatterns);
    
    return [...new Set([...extensionLanguages, ...contentLanguages, ...filenameLanguages])];
  }

  /**
   * Checks if a language is supported
   */
  isLanguageSupported(language: string): boolean {
    return this.getSupportedLanguages().includes(language.toLowerCase());
  }

  /**
   * Gets language display name
   */
  getLanguageDisplayName(language: string): string {
    const displayNames: Record<string, string> = {
      javascript: 'JavaScript',
      typescript: 'TypeScript',
      python: 'Python',
      go: 'Go',
      java: 'Java',
      csharp: 'C#',
      cpp: 'C++',
      c: 'C',
      rust: 'Rust',
      php: 'PHP',
      ruby: 'Ruby',
      swift: 'Swift',
      kotlin: 'Kotlin',
      html: 'HTML',
      css: 'CSS',
      sql: 'SQL',
      json: 'JSON',
      yaml: 'YAML',
      xml: 'XML',
      markdown: 'Markdown',
      shell: 'Shell Script',
      batch: 'Batch',
      powershell: 'PowerShell',
      dockerfile: 'Dockerfile',
      makefile: 'Makefile',
      cmake: 'CMake'
    };

    return displayNames[language.toLowerCase()] || 
           language.charAt(0).toUpperCase() + language.slice(1);
  }

  /**
   * Detects language from file extension
   */
  private detectLanguageFromExtension(filePath: string): LanguageDetectionResult {
    const extension = this.extractExtension(filePath);
    const language = this.config.extensionMap[extension];
    
    if (language) {
      return {
        language,
        confidence: 0.9,
        detectionMethod: 'extension'
      };
    }

    // Try case-insensitive match
    const lowerExtension = extension.toLowerCase();
    for (const [ext, lang] of Object.entries(this.config.extensionMap)) {
      if (ext.toLowerCase() === lowerExtension) {
        return {
          language: lang,
          confidence: 0.8,
          detectionMethod: 'extension'
        };
      }
    }

    return {
      language: this.config.defaultLanguage,
      confidence: 0.2,
      detectionMethod: 'fallback'
    };
  }

  /**
   * Detects language from filename patterns
   */
  private detectLanguageFromFilename(filePath: string): LanguageDetectionResult {
    const filename = this.extractFilename(filePath);
    
    for (const [language, patterns] of Object.entries(this.filenamePatterns)) {
      for (const pattern of patterns) {
        if (pattern.test(filename)) {
          return {
            language,
            confidence: 0.95,
            detectionMethod: 'filename'
          };
        }
      }
    }

    return {
      language: this.config.defaultLanguage,
      confidence: 0.1,
      detectionMethod: 'fallback'
    };
  }

  /**
   * Extracts file extension from path
   */
  private extractExtension(filePath: string): string {
    const lastDot = filePath.lastIndexOf('.');
    const lastSlash = Math.max(filePath.lastIndexOf('/'), filePath.lastIndexOf('\\'));
    
    if (lastDot > lastSlash && lastDot < filePath.length - 1) {
      return filePath.substring(lastDot);
    }
    
    return '';
  }

  /**
   * Extracts filename from path
   */
  private extractFilename(filePath: string): string {
    const lastSlash = Math.max(filePath.lastIndexOf('/'), filePath.lastIndexOf('\\'));
    return filePath.substring(lastSlash + 1);
  }
}

/**
 * Default language detection service instance
 */
export const defaultLanguageDetectionService = new LanguageDetectionService();