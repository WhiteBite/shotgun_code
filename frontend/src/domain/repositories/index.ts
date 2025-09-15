// Repository Interfaces for Clean Architecture

import { 
  Project, 
  FileNode, 
  Context, 
  Selection, 
  SelectedFile,
  CreateProjectData,
  CreateContextData 
} from '../entities'

/**
 * Project Repository Interface
 * Handles project persistence and retrieval
 */
export interface ProjectRepository {
  // Project Management
  createProject(data: CreateProjectData): Promise<Project>
  getProject(id: string): Promise<Project | null>
  getProjectByPath(path: string): Promise<Project | null>
  getAllProjects(): Promise<Project[]>
  getRecentProjects(limit?: number): Promise<Project[]>
  updateProject(project: Project): Promise<Project>
  deleteProject(id: string): Promise<void>
  
  // Project State
  getCurrentProject(): Promise<Project | null>
  setCurrentProject(project: Project): Promise<void>
  clearCurrentProject(): Promise<void>
}

/**
 * File System Repository Interface
 * Handles file system operations and file tree management
 */
export interface FileSystemRepository {
  // File Tree Operations
  getFileTree(projectPath: string): Promise<FileNode[]>
  refreshFileTree(projectPath: string): Promise<FileNode[]>
  searchFiles(projectPath: string, query: string): Promise<FileNode[]>
  
  // File Operations
  getFileContent(filePath: string): Promise<string>
  getFileInfo(filePath: string): Promise<FileNode | null>
  checkFileExists(filePath: string): Promise<boolean>
  getFileStats(filePath: string): Promise<FileStats>
  
  // Directory Operations
  getDirectoryContents(directoryPath: string): Promise<FileNode[]>
  createDirectory(directoryPath: string): Promise<void>
  
  // Bulk Operations
  getMultipleFileContents(filePaths: string[]): Promise<FileContentMap>
  validateFilePaths(filePaths: string[]): Promise<ValidationResult[]>
}

/**
 * Context Repository Interface
 * Handles context building and management
 */
export interface ContextRepository {
  // Context Building
  buildContext(files: SelectedFile[], projectPath: string): Promise<Context>
  buildContextFromPaths(filePaths: string[], projectPath: string): Promise<Context>
  
  // Context Management
  saveContext(context: Context): Promise<void>
  getContext(id: string): Promise<Context | null>
  getCurrentContext(): Promise<Context | null>
  setCurrentContext(context: Context): Promise<void>
  clearCurrentContext(): Promise<void>
  
  // Context History
  getContextHistory(limit?: number): Promise<Context[]>
  deleteContext(id: string): Promise<void>
  clearContextHistory(): Promise<void>
}

/**
 * Selection Repository Interface
 * Handles file selection state management
 */
export interface SelectionRepository {
  // Selection Management
  getSelection(): Promise<Selection>
  setSelection(selection: Selection): Promise<void>
  addFileToSelection(file: SelectedFile): Promise<Selection>
  removeFileFromSelection(filePath: string): Promise<Selection>
  clearSelection(): Promise<void>
  
  // Selection Validation
  validateSelection(files: SelectedFile[]): Promise<ValidationResult[]>
  filterValidFiles(files: SelectedFile[]): Promise<SelectedFile[]>
}

/**
 * Settings Repository Interface
 * Handles application settings and preferences
 */
export interface SettingsRepository {
  // General Settings
  getSetting<T>(key: string): Promise<T | null>
  setSetting<T>(key: string, value: T): Promise<void>
  getAllSettings(): Promise<Record<string, any>>
  resetSettings(): Promise<void>
  
  // UI Settings
  getUISettings(): Promise<UISettings>
  updateUISettings(settings: Partial<UISettings>): Promise<UISettings>
  
  // Context Settings
  getContextSettings(): Promise<ContextSettings>
  updateContextSettings(settings: Partial<ContextSettings>): Promise<ContextSettings>
  
  // Panel Settings
  getPanelSettings(): Promise<PanelSettings>
  updatePanelSettings(settings: Partial<PanelSettings>): Promise<PanelSettings>
}

/**
 * Cache Repository Interface
 * Handles caching for performance optimization
 */
export interface CacheRepository {
  // Generic Cache Operations
  get<T>(key: string): Promise<T | null>
  set<T>(key: string, value: T, ttl?: number): Promise<void>
  delete(key: string): Promise<void>
  clear(): Promise<void>
  has(key: string): Promise<boolean>
  
  // Specialized Cache Operations
  cacheFileContent(filePath: string, content: string): Promise<void>
  getCachedFileContent(filePath: string): Promise<string | null>
  invalidateFileCache(filePath: string): Promise<void>
  
  cacheFileTree(projectPath: string, tree: FileNode[]): Promise<void>
  getCachedFileTree(projectPath: string): Promise<FileNode[] | null>
  invalidateFileTreeCache(projectPath: string): Promise<void>
}

// Supporting Types and Interfaces

export interface FileStats {
  size: number
  lastModified: Date
  created: Date
  isDirectory: boolean
  isFile: boolean
  permissions: FilePermissions
}

export interface FilePermissions {
  readable: boolean
  writable: boolean
  executable: boolean
}

export interface FileContentMap {
  [filePath: string]: string
}

export interface ValidationResult {
  filePath: string
  isValid: boolean
  error?: string
  warnings?: string[]
}

export interface UISettings {
  theme: 'dark' | 'light' | 'auto'
  workspaceMode: 'manual' | 'autonomous' | 'reports'
  panels: {
    file: PanelConfig
    context: PanelConfig
  }
  animations: boolean
  compactMode: boolean
}

export interface ContextSettings {
  chunkSize: number
  splitLayout: 'vertical' | 'horizontal' | 'grid'
  autoChunk: boolean
  preserveCodeBlocks: boolean
  preserveMarkdown: boolean
  chunkStrategy: 'balanced' | 'natural' | 'aggressive'
}

export interface PanelSettings {
  file: PanelConfig
  context: PanelConfig
  reports: PanelConfig
}

export interface PanelConfig {
  width: number
  collapsed: boolean
  position: 'left' | 'right'
  resizable: boolean
}

// Error Types for Repository Operations

export class RepositoryError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly details?: any
  ) {
    super(message)
    this.name = 'RepositoryError'
  }
}

export class ProjectNotFoundError extends RepositoryError {
  constructor(projectId: string) {
    super(`Project not found: ${projectId}`, 'PROJECT_NOT_FOUND', { projectId })
  }
}

export class FileNotFoundError extends RepositoryError {
  constructor(filePath: string) {
    super(`File not found: ${filePath}`, 'FILE_NOT_FOUND', { filePath })
  }
}

export class FileAccessError extends RepositoryError {
  constructor(filePath: string, reason: string) {
    super(`Cannot access file: ${filePath} - ${reason}`, 'FILE_ACCESS_ERROR', { filePath, reason })
  }
}

export class ContextBuildError extends RepositoryError {
  constructor(reason: string, files?: string[]) {
    super(`Failed to build context: ${reason}`, 'CONTEXT_BUILD_ERROR', { reason, files })
  }
}

export class ValidationError extends RepositoryError {
  constructor(message: string, violations: ValidationResult[]) {
    super(message, 'VALIDATION_ERROR', { violations })
  }
}

// Repository Factory Interface

export interface RepositoryFactory {
  createProjectRepository(): ProjectRepository
  createFileSystemRepository(): FileSystemRepository
  createContextRepository(): ContextRepository
  createSelectionRepository(): SelectionRepository
  createSettingsRepository(): SettingsRepository
  createCacheRepository(): CacheRepository
}

// Repository interfaces for Clean Architecture implementation
export type { ContextRepository } from './ContextRepository';
export type { ProjectRepository, FileTreeOptions, FileChangeEvent, FileChangeCallback } from './ProjectRepository';
export type { GitRepository } from './GitRepository';
export type { AIRepository, AIGenerationOptions, AIProviderInfo } from './AIRepository';
export type { SettingsRepository } from './SettingsRepository';
export type { 
  ReportsRepository, 
  AutonomousRepository, 
  ExportRepository 
} from './RepositoryInterfaces';