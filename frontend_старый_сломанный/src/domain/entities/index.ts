// Domain Entities for Clean Architecture Implementation

// Context Summary interfaces for memory optimization
export type {
  ContextSummary,
  ContextMetadata as ContextSummaryMetadata,
  ContextBuildOptions,
  ContextBuildStatus,
  ContextChunk as ContextSummaryChunk,
  ContextValidationResult,
  MemoryImpact,
  ContextMetrics,
  ContextSuggestion,
  PaginationOptions,
  StreamingContext,
  StreamingProgress
} from './ContextSummary';

/**
 * Project Entity - Core business concept
 */
export class Project {
  constructor(
    public readonly id: string,
    public readonly name: string,
    public readonly path: string,
    public readonly type: ProjectType,
    public readonly metadata: ProjectMetadata,
    public readonly createdAt: Date,
    public readonly lastAccessedAt: Date
  ) {}

  static create(data: CreateProjectData): Project {
    return new Project(
      data.id || generateId(),
      data.name,
      data.path,
      data.type || 'general',
      data.metadata || {},
      new Date(),
      new Date()
    )
  }

  updateLastAccessed(): Project {
    return new Project(
      this.id,
      this.name,
      this.path,
      this.type,
      this.metadata,
      this.createdAt,
      new Date()
    )
  }

  updateMetadata(metadata: Partial<ProjectMetadata>): Project {
    return new Project(
      this.id,
      this.name,
      this.path,
      this.type,
      { ...this.metadata, ...metadata },
      this.createdAt,
      this.lastAccessedAt
    )
  }
}

/**
 * FileNode Entity - Represents a file or directory in the project
 */
export class FileNode {
  constructor(
    public readonly id: string,
    public readonly name: string,
    public readonly path: string,
    public readonly type: FileNodeType,
    public readonly size?: number,
    public readonly lastModified?: Date,
    public readonly children?: FileNode[],
    public readonly metadata?: FileNodeMetadata
  ) {}

  static createFile(data: CreateFileData): FileNode {
    return new FileNode(
      data.id || generateId(),
      data.name,
      data.path,
      'file',
      data.size,
      data.lastModified || new Date(),
      undefined,
      data.metadata
    )
  }

  static createDirectory(data: CreateDirectoryData): FileNode {
    return new FileNode(
      data.id || generateId(),
      data.name,
      data.path,
      'directory',
      undefined,
      data.lastModified || new Date(),
      data.children || [],
      data.metadata
    )
  }

  isFile(): boolean {
    return this.type === 'file'
  }

  isDirectory(): boolean {
    return this.type === 'directory'
  }

  hasChildren(): boolean {
    return this.children && this.children.length > 0
  }

  getFileExtension(): string | undefined {
    if (this.isFile()) {
      const lastDot = this.name.lastIndexOf('.')
      return lastDot > 0 ? this.name.substring(lastDot + 1) : undefined
    }
    return undefined
  }
}

/**
 * Context Entity - Represents built context from selected files
 */
export class Context {
  constructor(
    public readonly id: string,
    public readonly content: string,
    public readonly files: SelectedFile[],
    public readonly tokens: number,
    public readonly metadata: ContextMetadata,
    public readonly createdAt: Date,
    public readonly chunks?: ContextChunk[]
  ) {}

  static create(data: CreateContextData): Context {
    const content = data.content
    const tokens = estimateTokens(content)
    
    return new Context(
      data.id || generateId(),
      content,
      data.files,
      tokens,
      {
        buildDuration: data.buildDuration || 0,
        language: detectPrimaryLanguage(data.files),
        totalLines: content.split('\n').length,
        ...data.metadata
      },
      new Date(),
      data.chunks
    )
  }

  isEmpty(): boolean {
    return this.content.trim().length === 0
  }

  isLarge(): boolean {
    return this.tokens > 5000
  }

  shouldAutoChunk(): boolean {
    return this.isLarge() && !this.chunks
  }

  getFileCount(): number {
    return this.files.length
  }

  getTotalSize(): number {
    return this.files.reduce((total, file) => total + (file.size || 0), 0)
  }
}

/**
 * ContextChunk Entity - Represents a portion of context content
 */
export class ContextChunk {
  constructor(
    public readonly id: string,
    public readonly content: string,
    public readonly tokens: number,
    public readonly startLine: number,
    public readonly endLine: number,
    public readonly startChar: number,
    public readonly endChar: number,
    public readonly significance: ChunkSignificance,
    public readonly metadata: ChunkMetadata
  ) {}

  static create(data: CreateChunkData): ContextChunk {
    return new ContextChunk(
      data.id || generateId(),
      data.content,
      data.tokens || estimateTokens(data.content),
      data.startLine,
      data.endLine,
      data.startChar,
      data.endChar,
      data.significance || 'medium',
      {
        chunkIndex: data.chunkIndex || 0,
        totalChunks: data.totalChunks || 1,
        linesCount: data.endLine - data.startLine + 1,
        language: data.language,
        ...data.metadata
      }
    )
  }

  getContentPreview(maxLength: number = 100): string {
    if (this.content.length <= maxLength) {
      return this.content
    }
    return this.content.substring(0, maxLength - 3) + '...'
  }

  getLinesCount(): number {
    return this.endLine - this.startLine + 1
  }

  isSignificant(): boolean {
    return this.significance === 'high'
  }
}

/**
 * Selection Entity - Represents file selection state
 */
export class Selection {
  constructor(
    public readonly files: SelectedFile[],
    public readonly updatedAt: Date
  ) {}

  static empty(): Selection {
    return new Selection([], new Date())
  }

  static create(files: SelectedFile[]): Selection {
    return new Selection(files, new Date())
  }

  isEmpty(): boolean {
    return this.files.length === 0
  }

  hasFile(filePath: string): boolean {
    return this.files.some(file => file.path === filePath)
  }

  addFile(file: SelectedFile): Selection {
    if (this.hasFile(file.path)) {
      return this
    }
    return new Selection([...this.files, file], new Date())
  }

  removeFile(filePath: string): Selection {
    const filteredFiles = this.files.filter(file => file.path !== filePath)
    return new Selection(filteredFiles, new Date())
  }

  clear(): Selection {
    return Selection.empty()
  }

  getFileCount(): number {
    return this.files.length
  }

  getTotalSize(): number {
    return this.files.reduce((total, file) => total + (file.size || 0), 0)
  }
}

// Types and Interfaces

export type ProjectType = 'frontend' | 'backend' | 'fullstack' | 'mobile' | 'general'
export type FileNodeType = 'file' | 'directory'
export type ChunkSignificance = 'high' | 'medium' | 'low'

export interface ProjectMetadata {
  description?: string
  technologies?: string[]
  version?: string
  author?: string
  lastBuildAt?: Date
  [key: string]: unknown
}

export interface FileNodeMetadata {
  language?: string
  encoding?: string
  isIgnored?: boolean
  [key: string]: unknown
}

export interface ContextMetadata {
  buildDuration: number
  language?: string
  totalLines: number
  [key: string]: unknown
}

export interface ChunkMetadata {
  chunkIndex: number
  totalChunks: number
  linesCount: number
  language?: string
  [key: string]: unknown
}

export interface SelectedFile {
  path: string
  name: string
  size?: number
  lastModified?: Date
  language?: string
}

export interface CreateProjectData {
  id?: string
  name: string
  path: string
  type?: ProjectType
  metadata?: ProjectMetadata
}

export interface CreateFileData {
  id?: string
  name: string
  path: string
  size?: number
  lastModified?: Date
  metadata?: FileNodeMetadata
}

export interface CreateDirectoryData {
  id?: string
  name: string
  path: string
  lastModified?: Date
  children?: FileNode[]
  metadata?: FileNodeMetadata
}

export interface CreateContextData {
  id?: string
  content: string
  files: SelectedFile[]
  buildDuration?: number
  metadata?: Partial<ContextMetadata>
  chunks?: ContextChunk[]
}

export interface CreateChunkData {
  id?: string
  content: string
  tokens?: number
  startLine: number
  endLine: number
  startChar: number
  endChar: number
  significance?: ChunkSignificance
  chunkIndex?: number
  totalChunks?: number
  language?: string
  metadata?: Partial<ChunkMetadata>
}

// Utility Functions

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
}

function estimateTokens(text: string): number {
  // Simple token estimation - can be improved
  const words = text.split(/\s+/).length
  const chars = text.length
  return Math.round((words * 1.3 + chars / 4) / 2)
}

function detectPrimaryLanguage(files: SelectedFile[]): string | undefined {
  const languageCounts: Record<string, number> = {}
  
  files.forEach(file => {
    if (file.language) {
      languageCounts[file.language] = (languageCounts[file.language] || 0) + 1
    }
  })
  
  const mostCommon = Object.entries(languageCounts)
    .sort(([, a], [, b]) => b - a)[0]
  
  return mostCommon?.[0]
}