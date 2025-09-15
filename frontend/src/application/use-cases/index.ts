// Use Cases for Clean Architecture Application Layer

import { 
  Project, 
  FileNode, 
  Context, 
  ContextChunk 
} from '../domain/entities'
import { 
  ProjectRepository,
  FileSystemRepository,
  ContextRepository,
  SelectionRepository,
  SettingsRepository,
  ValidationResult
} from '../domain/repositories'

// Export all use cases from separate files
export * from './BuildContextUseCase';
export * from './GetContextContentUseCase';
export * from './CreateStreamingContextUseCase';
export * from './ProjectUseCases';
export * from './GitUseCases';
export * from './AIUseCases';
export * from './SettingsUseCases';
export * from './ReportsUseCases';

/**
 * Project Management Use Cases
 */

export class CreateProjectUseCase {
  constructor(private projectRepo: ProjectRepository) {}

  async execute(data: any): Promise<Project> {
    // Validate project data
    this.validateProjectData(data)
    
    // Check if project already exists
    const existingProject = await this.projectRepo.getProjectByPath(data.path)
    if (existingProject) {
      throw new Error(`Project already exists at path: ${data.path}`)
    }
    
    // Create and save project
    const project = await this.projectRepo.createProject(data)
    return project
  }

  private validateProjectData(data: any): void {
    if (!data.name?.trim()) {
      throw new Error('Project name is required')
    }
    if (!data.path?.trim()) {
      throw new Error('Project path is required')
    }
  }
}

export class LoadProjectUseCase {
  constructor(
    private projectRepo: ProjectRepository,
    private fileSystemRepo: FileSystemRepository
  ) {}

  async execute(projectId: string): Promise<Project> {
    const project = await this.projectRepo.getProject(projectId)
    if (!project) {
      throw new Error(`Project not found: ${projectId}`)
    }

    // Update last accessed time
    const updatedProject = project.updateLastAccessed()
    await this.projectRepo.updateProject(updatedProject)
    
    // Set as current project
    await this.projectRepo.setCurrentProject(updatedProject)
    
    return updatedProject
  }
}

export class GetRecentProjectsUseCase {
  constructor(private projectRepo: ProjectRepository) {}

  async execute(limit: number = 10): Promise<Project[]> {
    return await this.projectRepo.getRecentProjects(limit)
  }
}

/**
 * File System Use Cases
 */

export class LoadFileTreeUseCase {
  constructor(
    private fileSystemRepo: FileSystemRepository,
    private projectRepo: ProjectRepository
  ) {}

  async execute(refresh: boolean = false): Promise<FileNode[]> {
    const currentProject = await this.projectRepo.getCurrentProject()
    if (!currentProject) {
      throw new Error('No current project selected')
    }

    if (refresh) {
      return await this.fileSystemRepo.refreshFileTree(currentProject.path)
    } else {
      return await this.fileSystemRepo.getFileTree(currentProject.path)
    }
  }
}

export class SearchFilesUseCase {
  constructor(
    private fileSystemRepo: FileSystemRepository,
    private projectRepo: ProjectRepository
  ) {}

  async execute(query: string): Promise<FileNode[]> {
    if (!query.trim()) {
      return []
    }

    const currentProject = await this.projectRepo.getCurrentProject()
    if (!currentProject) {
      throw new Error('No current project selected')
    }

    return await this.fileSystemRepo.searchFiles(currentProject.path, query)
  }
}

export class ValidateFilePathsUseCase {
  constructor(private fileSystemRepo: FileSystemRepository) {}

  async execute(filePaths: string[]): Promise<ValidationResult[]> {
    return await this.fileSystemRepo.validateFilePaths(filePaths)
  }
}

/**
 * File Selection Use Cases
 */

export class UpdateFileSelectionUseCase {
  constructor(
    private selectionRepo: SelectionRepository,
    private fileSystemRepo: FileSystemRepository
  ) {}

  async execute(filePaths: string[]): Promise<any> {
    // Validate file paths
    const validationResults = await this.fileSystemRepo.validateFilePaths(filePaths)
    const validFiles: any[] = []
    const errors: string[] = []

    for (const result of validationResults) {
      if (result.isValid) {
        const fileInfo = await this.fileSystemRepo.getFileInfo(result.filePath)
        if (fileInfo) {
          validFiles.push({
            path: fileInfo.path,
            name: fileInfo.name,
            size: fileInfo.size,
            lastModified: fileInfo.lastModified,
            language: this.detectLanguage(fileInfo.name)
          })
        }
      } else {
        errors.push(`${result.filePath}: ${result.error}`)
      }
    }

    if (errors.length > 0) {
      console.warn('Some files could not be selected:', errors)
    }

    const selection = Selection.create(validFiles)
    await this.selectionRepo.setSelection(selection)
    return selection
  }

  private detectLanguage(fileName: string): string | undefined {
    const extension = fileName.split('.').pop()?.toLowerCase()
    const languageMap: Record<string, string> = {
      'ts': 'typescript',
      'js': 'javascript',
      'vue': 'vue',
      'py': 'python',
      'java': 'java',
      'cpp': 'cpp',
      'c': 'c',
      'cs': 'csharp',
      'php': 'php',
      'rb': 'ruby',
      'go': 'go',
      'rs': 'rust',
      'swift': 'swift',
      'kt': 'kotlin',
      'scala': 'scala',
      'md': 'markdown',
      'json': 'json',
      'xml': 'xml',
      'html': 'html',
      'css': 'css',
      'scss': 'scss',
      'sass': 'sass',
      'less': 'less'
    }
    
    return extension ? languageMap[extension] : undefined
  }
}

export class AddFileToSelectionUseCase {
  constructor(private selectionRepo: SelectionRepository) {}

  async execute(file: any): Promise<any> {
    return await this.selectionRepo.addFileToSelection(file)
  }
}

export class RemoveFileFromSelectionUseCase {
  constructor(private selectionRepo: SelectionRepository) {}

  async execute(filePath: string): Promise<any> {
    return await this.selectionRepo.removeFileFromSelection(filePath)
  }
}

export class ClearFileSelectionUseCase {
  constructor(private selectionRepo: SelectionRepository) {}

  async execute(): Promise<void> {
    await this.selectionRepo.clearSelection()
  }
}

/**
 * Context Building Use Cases
 */

export class BuildContextUseCase {
  constructor(
    private contextRepo: ContextRepository,
    private selectionRepo: SelectionRepository,
    private projectRepo: ProjectRepository,
    private settingsRepo: SettingsRepository
  ) {}

  async execute(): Promise<Context> {
    // Get current selection
    const selection = await this.selectionRepo.getSelection()
    if (selection.isEmpty()) {
      throw new Error('No files selected for context building')
    }

    // Get current project
    const currentProject = await this.projectRepo.getCurrentProject()
    if (!currentProject) {
      throw new Error('No current project selected')
    }

    // Build context
    const context = await this.contextRepo.buildContext(
      selection.files,
      currentProject.path
    )

    // Auto-chunk if enabled and context is large
    const contextSettings = await this.settingsRepo.getContextSettings()
    if (contextSettings.autoChunk && context.shouldAutoChunk()) {
      // This would trigger chunking - implementation would depend on chunking service
    }

    // Save as current context
    await this.contextRepo.setCurrentContext(context)
    
    return context
  }
}

export class BuildContextFromPathsUseCase {
  constructor(
    private contextRepo: ContextRepository,
    private projectRepo: ProjectRepository
  ) {}

  async execute(filePaths: string[]): Promise<Context> {
    if (filePaths.length === 0) {
      throw new Error('No file paths provided for context building')
    }

    const currentProject = await this.projectRepo.getCurrentProject()
    if (!currentProject) {
      throw new Error('No current project selected')
    }

    const context = await this.contextRepo.buildContextFromPaths(
      filePaths,
      currentProject.path
    )

    await this.contextRepo.setCurrentContext(context)
    return context
  }
}

export class GetContextHistoryUseCase {
  constructor(private contextRepo: ContextRepository) {}

  async execute(limit: number = 20): Promise<Context[]> {
    return await this.contextRepo.getContextHistory(limit)
  }
}

export class ClearContextUseCase {
  constructor(private contextRepo: ContextRepository) {}

  async execute(): Promise<void> {
    await this.contextRepo.clearCurrentContext()
  }
}

/**
 * Settings Use Cases
 */

export class UpdateUISettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(settings: Partial<import('../domain/repositories').UISettings>): Promise<void> {
    await this.settingsRepo.updateUISettings(settings)
  }
}

export class UpdateContextSettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(settings: Partial<import('../domain/repositories').ContextSettings>): Promise<void> {
    await this.settingsRepo.updateContextSettings(settings)
  }
}

export class UpdatePanelSettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(settings: Partial<import('../domain/repositories').PanelSettings>): Promise<void> {
    await this.settingsRepo.updatePanelSettings(settings)
  }
}

/**
 * Workspace Use Cases
 */

export class InitializeWorkspaceUseCase {
  constructor(
    private projectRepo: ProjectRepository,
    private fileSystemRepo: FileSystemRepository,
    private settingsRepo: SettingsRepository
  ) {}

  async execute(projectPath?: string): Promise<{
    project: Project | null
    fileTree: FileNode[]
    settings: any
  }> {
    let project: Project | null = null
    let fileTree: FileNode[] = []

    // Load project if path provided or get current project
    if (projectPath) {
      project = await this.projectRepo.getProjectByPath(projectPath)
      if (project) {
        await this.projectRepo.setCurrentProject(project)
      }
    } else {
      project = await this.projectRepo.getCurrentProject()
    }

    // Load file tree if we have a project
    if (project) {
      try {
        fileTree = await this.fileSystemRepo.getFileTree(project.path)
      } catch (error) {
        console.error('Failed to load file tree:', error)
        // Continue without file tree
      }
    }

    // Load settings
    const settings = await this.settingsRepo.getAllSettings()

    return {
      project,
      fileTree,
      settings
    }
  }
}

export class SwitchWorkspaceModeUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(mode: 'manual' | 'autonomous' | 'reports'): Promise<void> {
    await this.settingsRepo.updateUISettings({ workspaceMode: mode })
  }
}

/**
 * Command Pattern - Command Interface for Use Cases
 */

export interface Command<TInput = void, TOutput = void> {
  execute(input: TInput): Promise<TOutput>
}

/**
 * Use Case Factory for Dependency Injection
 */

export interface UseCaseFactory {
  // Project Use Cases
  createProjectUseCase(): CreateProjectUseCase
  loadProjectUseCase(): LoadProjectUseCase
  getRecentProjectsUseCase(): GetRecentProjectsUseCase
  
  // File System Use Cases
  loadFileTreeUseCase(): LoadFileTreeUseCase
  searchFilesUseCase(): SearchFilesUseCase
  validateFilePathsUseCase(): ValidateFilePathsUseCase
  
  // Selection Use Cases
  updateFileSelectionUseCase(): UpdateFileSelectionUseCase
  addFileToSelectionUseCase(): AddFileToSelectionUseCase
  removeFileFromSelectionUseCase(): RemoveFileFromSelectionUseCase
  clearFileSelectionUseCase(): ClearFileSelectionUseCase
  
  // Context Use Cases
  buildContextUseCase(): BuildContextUseCase
  buildContextFromPathsUseCase(): BuildContextFromPathsUseCase
  getContextHistoryUseCase(): GetContextHistoryUseCase
  clearContextUseCase(): ClearContextUseCase
  
  // Settings Use Cases
  updateUISettingsUseCase(): UpdateUISettingsUseCase
  updateContextSettingsUseCase(): UpdateContextSettingsUseCase
  updatePanelSettingsUseCase(): UpdatePanelSettingsUseCase
  
  // Workspace Use Cases
  initializeWorkspaceUseCase(): InitializeWorkspaceUseCase
  switchWorkspaceModeUseCase(): SwitchWorkspaceModeUseCase
}