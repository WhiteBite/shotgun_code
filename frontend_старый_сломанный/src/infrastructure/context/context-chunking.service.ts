import { ref } from 'vue'
import { 
  defaultTokenEstimationService, 
  defaultLanguageDetectionService,
  type TokenEstimationService,
  type LanguageDetectionService
} from '@/domain/services'
import { APP_CONFIG } from '@/config/app-config'

export interface ContextChunk {
  id: string
  content: string
  tokens: number
  startLine: number
  endLine: number
  startChar: number
  endChar: number
  metadata: {
    chunkIndex: number
    totalChunks: number
    linesCount: number
    language?: string
    fileSegments?: FileSegment[]
    significance: 'high' | 'medium' | 'low'
  }
}

export interface FileSegment {
  filePath: string
  startLine: number
  endLine: number
  percentage: number
}

export interface ChunkingOptions {
  maxTokens: number
  strategy: 'balanced' | 'natural' | 'aggressive'
  preserveCodeBlocks: boolean
  preserveMarkdownStructure: boolean
  minChunkSize: number
  overlapSize: number
}

export interface CopyAction {
  id: string
  type: 'chunk' | 'selection' | 'all'
  content: string
  tokens: number
  chunkId?: string
  chunkIndex?: number
  timestamp: Date
  source: 'manual' | 'hotkey' | 'context-menu'
  metadata?: unknown
}

export interface ContentStructure {
  language: string | null
  hasCodeBlocks: boolean
  hasMarkdownHeaders: boolean
  hasFunctionDefinitions: boolean
  hasClassDefinitions: boolean
  naturalBreakPoints: number[]
}

export class ContextChunkingService {
  private static instance: ContextChunkingService
  private copyHistory = ref<CopyAction[]>([])
  private activeChunks = ref<ContextChunk[]>([])
  
  private readonly tokenEstimationService: TokenEstimationService
  private readonly languageDetectionService: LanguageDetectionService
  
  constructor(
    tokenEstimationService: TokenEstimationService = defaultTokenEstimationService,
    languageDetectionService: LanguageDetectionService = defaultLanguageDetectionService
  ) {
    this.tokenEstimationService = tokenEstimationService
    this.languageDetectionService = languageDetectionService
  }
  
  static getInstance(): ContextChunkingService {
    if (!ContextChunkingService.instance) {
      ContextChunkingService.instance = new ContextChunkingService()
    }
    return ContextChunkingService.instance
  }

  /**
   * Chunk content using intelligent algorithms
   */
  async chunkContent(
    content: string, 
    options: Partial<ChunkingOptions> = {}
  ): Promise<ContextChunk[]> {
    const config: ChunkingOptions = {
      maxTokens: APP_CONFIG.performance.streaming.CHUNK_SIZE,
      strategy: 'balanced',
      preserveCodeBlocks: true,
      preserveMarkdownStructure: true,
      minChunkSize: 100,
      overlapSize: 50,
      ...options
    }

    const chunks: ContextChunk[] = []
    const lines = content.split('\n')
    
    // Detect content type and structure
    const contentStructure = await this.analyzeContentStructure(content)
    
    switch (config.strategy) {
      case 'natural':
        return await this.chunkByNaturalBreaks(content, lines, config, contentStructure)
      case 'aggressive':
        return await this.chunkByTokenLimit(content, lines, config)
      default:
        return await this.chunkBalanced(content, lines, config, contentStructure)
    }
  }

  /**
   * Analyze content structure to identify natural break points
   */
  private async analyzeContentStructure(content: string): Promise<ContentStructure> {
    const languageResult = await this.languageDetectionService.detectLanguageFromContent(content)
    
    const structure: ContentStructure = {
      language: languageResult.language,
      hasCodeBlocks: /```[\s\S]*?```/.test(content),
      hasMarkdownHeaders: /^#{1,6}\s+.+$/gm.test(content),
      hasFunctionDefinitions: false,
      hasClassDefinitions: false,
      naturalBreakPoints: []
    }

    // Detect programming language constructs
    if (structure.language && structure.language !== 'text') {
      structure.hasFunctionDefinitions = this.detectFunctions(content)
      structure.hasClassDefinitions = this.detectClasses(content)
    }

    // Find natural break points
    structure.naturalBreakPoints = this.findNaturalBreakPoints(content)

    return structure
  }

  /**
   * Balanced chunking strategy that respects content structure
   */
  private async chunkBalanced(
    content: string,
    lines: string[],
    config: ChunkingOptions,
    structure: ContentStructure
  ): Promise<ContextChunk[]> {
    const chunks: ContextChunk[] = []
    let currentChunk = ''
    let currentTokens = 0
    let startLine = 1
    let startChar = 0
    let chunkIndex = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const lineTokens = await this.estimateTokens(line)
      
      // Check if we should break here
      const shouldBreak = await this.shouldBreakAtLine(
        i, 
        lines, 
        currentTokens + lineTokens, 
        config, 
        structure
      )

      if (shouldBreak && currentChunk.trim() && currentTokens >= config.minChunkSize) {
        // Create chunk
        const chunk = this.createChunk(
          chunkIndex,
          currentChunk.trim(),
          currentTokens,
          startLine,
          i,
          startChar,
          startChar + currentChunk.length,
          structure.language || undefined
        )
        chunks.push(chunk)

        // Start new chunk with overlap if configured
        if (config.overlapSize > 0) {
          const overlapLines = this.getOverlapLines(lines, i, config.overlapSize)
          currentChunk = overlapLines.join('\n') + '\n'
          currentTokens = await this.estimateTokens(currentChunk)
        } else {
          currentChunk = ''
          currentTokens = 0
        }
        
        startLine = i + 1
        startChar += currentChunk.length
        chunkIndex++
      }

      currentChunk += line + '\n'
      currentTokens += lineTokens
    }

    // Add final chunk
    if (currentChunk.trim()) {
      const chunk = this.createChunk(
        chunkIndex,
        currentChunk.trim(),
        currentTokens,
        startLine,
        lines.length,
        startChar,
        content.length,
        structure.language || undefined
      )
      chunks.push(chunk)
    }

    // Update metadata with total chunks count
    chunks.forEach(chunk => {
      chunk.metadata.totalChunks = chunks.length
    })

    this.activeChunks.value = chunks
    return chunks
  }

  /**
   * Determine if we should break at this line
   */
  private async shouldBreakAtLine(
    lineIndex: number,
    lines: string[],
    currentTokens: number,
    config: ChunkingOptions,
    structure: ContentStructure
  ): Promise<boolean> {
    const line = lines[lineIndex]

    // Hard limit check
    if (currentTokens >= config.maxTokens) {
      return true
    }

    // Don't break if we're under minimum size
    if (currentTokens < config.minChunkSize) {
      return false
    }

    // Check for natural break points
    if (this.isNaturalBreakPoint(lineIndex, lines, structure)) {
      return currentTokens >= config.minChunkSize * 0.7 // 70% of minimum
    }

    // Code block preservation
    if (config.preserveCodeBlocks && this.isInsideCodeBlock(lineIndex, lines)) {
      return false
    }

    // Function/class boundary detection
    if (structure.hasFunctionDefinitions && this.isFunctionBoundary(line)) {
      return currentTokens >= config.minChunkSize * 0.8
    }

    return false
  }

  /**
   * Check for natural break points
   */
  private isNaturalBreakPoint(lineIndex: number, lines: string[], structure: ContentStructure): boolean {
    const line = lines[lineIndex]
    const nextLine = lines[lineIndex + 1]

    // Markdown headers
    if (structure.hasMarkdownHeaders && /^#{1,6}\s+.+$/.test(nextLine)) {
      return true
    }

    // Empty lines followed by content
    if (line.trim() === '' && nextLine && nextLine.trim() !== '') {
      return true
    }

    // Comment blocks
    if (this.isCommentBlockEnd(lineIndex, lines)) {
      return true
    }

    return false
  }

  /**
   * Create a context chunk with metadata
   */
  private createChunk(
    index: number,
    content: string,
    tokens: number,
    startLine: number,
    endLine: number,
    startChar: number,
    endChar: number,
    language?: string
  ): ContextChunk {
    return {
      id: `chunk-${index}-${Date.now()}`,
      content,
      tokens,
      startLine,
      endLine,
      startChar,
      endChar,
      metadata: {
        chunkIndex: index,
        totalChunks: 0, // Will be updated later
        linesCount: endLine - startLine + 1,
        language,
        significance: this.calculateSignificance(content, tokens),
        fileSegments: this.extractFileSegments(content)
      }
    }
  }

  /**
   * Calculate chunk significance based on content analysis
   */
  private calculateSignificance(content: string, tokens: number): 'high' | 'medium' | 'low' {
    let score = 0

    // Token count factor
    if (tokens > 800) score += 2
    else if (tokens > 400) score += 1

    // Function/class definitions
    if (this.detectFunctions(content)) score += 2
    if (this.detectClasses(content)) score += 3

    // Important keywords
    const importantKeywords = [
      'export', 'import', 'class', 'function', 'const', 'let', 'var',
      'interface', 'type', 'enum', 'namespace', 'module'
    ]
    
    const keywordCount = importantKeywords.reduce((count, keyword) => {
      return count + (content.match(new RegExp(`\\b${keyword}\\b`, 'g')) || []).length
    }, 0)
    
    score += Math.min(keywordCount * 0.5, 3)

    // Documentation/comments
    const commentRatio = this.calculateCommentRatio(content)
    if (commentRatio > 0.3) score += 1

    if (score >= 5) return 'high'
    if (score >= 2) return 'medium'
    return 'low'
  }

  /**
   * Copy chunk content with tracking
   */
  async copyChunk(
    chunk: ContextChunk, 
    source: 'manual' | 'hotkey' | 'context-menu' = 'manual'
  ): Promise<boolean> {
    try {
      await navigator.clipboard.writeText(chunk.content)
      
      const copyAction: CopyAction = {
        id: `copy-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        type: 'chunk',
        content: chunk.content,
        tokens: chunk.tokens,
        chunkId: chunk.id,
        chunkIndex: chunk.metadata.chunkIndex,
        timestamp: new Date(),
        source,
        metadata: {
          chunkMetadata: chunk.metadata,
          contentPreview: chunk.content.substring(0, 100) + '...'
        }
      }

      this.addToCopyHistory(copyAction)
      return true
    } catch (error) {
      console.error('Failed to copy chunk:', error)
      return false
    }
  }

  /**
   * Copy all content
   */
  async copyAll(
    content: string,
    source: 'manual' | 'hotkey' | 'context-menu' = 'manual'
  ): Promise<boolean> {
    try {
      await navigator.clipboard.writeText(content)
      
      const copyAction: CopyAction = {
        id: `copy-all-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        type: 'all',
        content,
        tokens: await this.estimateTokens(content),
        timestamp: new Date(),
        source,
        metadata: {
          totalChunks: this.activeChunks.value.length,
          contentPreview: content.substring(0, 100) + '...'
        }
      }

      this.addToCopyHistory(copyAction)
      return true
    } catch (error) {
      console.error('Failed to copy all content:', error)
      return false
    }
  }

  /**
   * Copy selection with tracking
   */
  async copySelection(
    selection: string,
    startLine?: number,
    endLine?: number,
    source: 'manual' | 'hotkey' | 'context-menu' = 'manual'
  ): Promise<boolean> {
    try {
      await navigator.clipboard.writeText(selection)
      
      const copyAction: CopyAction = {
        id: `copy-sel-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        type: 'selection',
        content: selection,
        tokens: await this.estimateTokens(selection),
        timestamp: new Date(),
        source,
        metadata: {
          startLine,
          endLine,
          contentPreview: selection.substring(0, 100) + '...'
        }
      }

      this.addToCopyHistory(copyAction)
      return true
    } catch (error) {
      console.error('Failed to copy selection:', error)
      return false
    }
  }

  /**
   * Get copy history
   */
  getCopyHistory(): CopyAction[] {
    return this.copyHistory.value
  }

  /**
   * Get recent copy actions
   */
  getRecentCopyActions(limit: number = 5): CopyAction[] {
    return this.copyHistory.value
      .slice(-limit)
      .reverse()
  }

  /**
   * Clear copy history
   */
  clearCopyHistory(): void {
    this.copyHistory.value = []
  }

  /**
   * Add to copy history with size management
   */
  private addToCopyHistory(action: CopyAction): void {
    this.copyHistory.value.push(action)
    
    // Keep only last N actions based on centralized config
    if (this.copyHistory.value.length > APP_CONFIG.context.chunking.COPY_HISTORY_LIMIT) {
      this.copyHistory.value = this.copyHistory.value.slice(-APP_CONFIG.context.chunking.COPY_HISTORY_LIMIT)
    }
  }

  /**
   * Estimate tokens in text using domain service
   */
  private async estimateTokens(text: string, language?: string): Promise<number> {
    const result = await this.tokenEstimationService.estimateTokens(text, { language });
    return result.estimatedTokens;
  }

  private detectLanguage(content: string): string | undefined {
    // Simple language detection based on patterns
    if (/^\s*import\s+.*from\s+['"]/.test(content) || /^\s*export\s+/.test(content)) {
      return 'typescript'
    }
    if (/^\s*<template>/.test(content) || /<\/template>\s*$/.test(content)) {
      return 'vue'
    }
    if (/^\s*#\s+/.test(content) || /^\s*##\s+/.test(content)) {
      return 'markdown'
    }
    return undefined
  }

  private detectFunctions(content: string): boolean {
    const functionPatterns = [
      /function\s+\w+\s*\(/,
      /const\s+\w+\s*=\s*\(/,
      /const\s+\w+\s*=\s*async\s*\(/,
      /\w+\s*\([^)]*\)\s*{/,
      /=>\s*{/
    ]
    
    return functionPatterns.some(pattern => pattern.test(content))
  }

  private detectClasses(content: string): boolean {
    return /class\s+\w+/.test(content) || /interface\s+\w+/.test(content)
  }

  private findNaturalBreakPoints(content: string): number[] {
    const lines = content.split('\n')
    const breakPoints: number[] = []
    
    lines.forEach((line, index) => {
      // Empty lines
      if (line.trim() === '') {
        breakPoints.push(index)
      }
      
      // Comments
      if (/^\s*(\/\/|#|<!--)/.test(line)) {
        breakPoints.push(index)
      }
      
      // Function/class definitions
      if (this.isFunctionBoundary(line)) {
        breakPoints.push(index)
      }
    })
    
    return breakPoints
  }

  private isFunctionBoundary(line: string): boolean {
    return /^\s*(function|class|interface|type|const\s+\w+\s*=\s*\()/i.test(line)
  }

  private isInsideCodeBlock(lineIndex: number, lines: string[]): boolean {
    let inCodeBlock = false
    
    for (let i = 0; i <= lineIndex; i++) {
      if (lines[i].includes('```')) {
        inCodeBlock = !inCodeBlock
      }
    }
    
    return inCodeBlock
  }

  private isCommentBlockEnd(lineIndex: number, lines: string[]): boolean {
    const line = lines[lineIndex]
    const nextLine = lines[lineIndex + 1]
    
    return /\*\/\s*$/.test(line) && !!nextLine && !nextLine.trim().startsWith('*')
  }

  private calculateCommentRatio(content: string): number {
    const lines = content.split('\n')
    const commentLines = lines.filter(line => 
      /^\s*(\/\/|#|\/\*|\*|<!--)/.test(line)
    ).length
    
    return commentLines / lines.length
  }

  private extractFileSegments(_content: string): FileSegment[] {
    // This would analyze content to identify which files it came from
    // For now, return empty array - would need integration with file tree data
    return []
  }

  private getOverlapLines(lines: string[], fromIndex: number, overlapSize: number): string[] {
    const start = Math.max(0, fromIndex - overlapSize)
    return lines.slice(start, fromIndex)
  }

  private async chunkByNaturalBreaks(
    content: string,
    lines: string[],
    config: ChunkingOptions,
    structure: ContentStructure
  ): Promise<ContextChunk[]> {
    // Implementation for natural break strategy
    return await this.chunkBalanced(content, lines, config, structure)
  }

  private async chunkByTokenLimit(
    content: string,
    lines: string[],
    config: ChunkingOptions
  ): Promise<ContextChunk[]> {
    // Implementation for aggressive token-based chunking
    const chunks: ContextChunk[] = []
    let currentChunk = ''
    let currentTokens = 0
    let startLine = 1
    let chunkIndex = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const lineTokens = await this.estimateTokens(line)
      
      if (currentTokens + lineTokens > config.maxTokens && currentChunk.trim()) {
        const chunk = this.createChunk(
          chunkIndex,
          currentChunk.trim(),
          currentTokens,
          startLine,
          i,
          0,
          currentChunk.length
        )
        chunks.push(chunk)
        
        currentChunk = ''
        currentTokens = 0
        startLine = i + 1
        chunkIndex++
      }
      
      currentChunk += line + '\n'
      currentTokens += lineTokens
    }

    if (currentChunk.trim()) {
      const chunk = this.createChunk(
        chunkIndex,
        currentChunk.trim(),
        currentTokens,
        startLine,
        lines.length,
        0,
        content.length
      )
      chunks.push(chunk)
    }

    chunks.forEach(chunk => {
      chunk.metadata.totalChunks = chunks.length
    })

    this.activeChunks.value = chunks
    return chunks
  }
}

// Export composable for Vue components
export function useContextChunking() {
  const service = ContextChunkingService.getInstance()
  
  return {
    chunkContent: service.chunkContent.bind(service),
    copyChunk: service.copyChunk.bind(service),
    copyAll: service.copyAll.bind(service),
    copySelection: service.copySelection.bind(service),
    getCopyHistory: service.getCopyHistory.bind(service),
    getRecentCopyActions: service.getRecentCopyActions.bind(service),
    clearCopyHistory: service.clearCopyHistory.bind(service)
  }
}