/**
 * Dependency Injection Setup
 * Configures all dependencies for Clean Architecture
 */

import { container, TOKENS } from './container';

// Repository Implementations
import { ProjectRepositoryImpl } from '../repositories/ProjectRepositoryImpl';
import { FileSystemRepositoryImpl } from '../repositories/FileSystemRepositoryImpl';
import { ProjectApiAdapter } from '../repositories/ProjectApiAdapter';

// Use Case Implementations
import { 
  LoadProjectFilesUseCase,
  SelectProjectDirectoryUseCase,
  ReadFileContentUseCase,
  StartFileWatchingUseCase,
  StopFileWatchingUseCase,
  GetFileStatsUseCase,
  type AIService
} from '../../application/use-cases';

// Add the new LoadProjectUseCase
import { LoadProjectUseCase } from '../../application/use-cases/LoadProjectUseCase';

// Services
import { apiService } from '../api/api.service';
import { eventService } from '../events/event.service';

// AI Service Implementation
class AIServiceImpl implements AIService {
  async generateCode(request: any) {
    try {
      const result = await apiService.generateCodeWithOptions(
        request.systemPrompt,
        request.userPrompt,
        JSON.stringify(request.options)
      );
      
      return {
        content: result,
        model: request.options.model || 'default',
        tokensUsed: Math.ceil(result.length / 4) // Rough estimation
      };
    } catch (error) {
      throw new Error(`AI generation failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  async listModels(): Promise<string[]> {
    try {
      return await apiService.listAvailableModels();
    } catch (error) {
      console.warn('Failed to list models:', error);
      return ['default'];
    }
  }

  async getProviderInfo() {
    try {
      const info = await apiService.getProviderInfo();
      return JSON.parse(info);
    } catch (error) {
      return {
        name: 'Default Provider',
        version: '1.0.0',
        supportedModels: ['default']
      };
    }
  }
}

// Event Bus Implementation
class EventBusImpl {
  private listeners = new Map<string, Function[]>();

  emit(event: any) {
    const eventName = event.constructor.name;
    const handlers = this.listeners.get(eventName) || [];
    handlers.forEach(handler => {
      try {
        handler(event);
      } catch (error) {
        console.error(`Event handler error for ${eventName}:`, error);
      }
    });
  }

  on(eventName: string, handler: Function) {
    const handlers = this.listeners.get(eventName) || [];
    handlers.push(handler);
    this.listeners.set(eventName, handlers);
  }

  off(eventName: string, handler: Function) {
    const handlers = this.listeners.get(eventName) || [];
    const filtered = handlers.filter(h => h !== handler);
    this.listeners.set(eventName, filtered);
  }

  clear() {
    this.listeners.clear();
  }
}

// Context Repository Implementation (simplified)
class ContextRepositoryImpl {
  private contexts = new Map<string, any>();

  async save(context: any): Promise<void> {
    this.contexts.set(context.id, context);
  }

  async findById(id: string): Promise<any | null> {
    return this.contexts.get(id) || null;
  }

  async getContextsForProject(projectId: string): Promise<any[]> {
    const contexts = Array.from(this.contexts.values());
    return contexts.filter(ctx => ctx.project.id.value === projectId);
  }

  async delete(id: string): Promise<void> {
    this.contexts.delete(id);
  }
}

// Combined Project Repository that implements both interfaces
class CombinedProjectRepository extends ProjectApiAdapter {
  private projectRepositoryImpl: ProjectRepositoryImpl;

  constructor() {
    super();
    this.projectRepositoryImpl = new ProjectRepositoryImpl();
  }

  // Delegate to ProjectRepositoryImpl for these methods
  async getRecentProjects(limit?: number) {
    return this.projectRepositoryImpl.getRecentProjects(limit);
  }

  async removeFromRecent(projectId: string) {
    return this.projectRepositoryImpl.removeFromRecent(projectId);
  }

  async save(project: any) {
    return this.projectRepositoryImpl.save(project);
  }

  async findByPath(path: any) {
    return this.projectRepositoryImpl.findByPath(path);
  }

  async exists(path: any) {
    return this.projectRepositoryImpl.exists(path);
  }
}

export function setupDependencies() {
  // Clear existing registrations
  container.clear();

  // Register repositories
  container.registerSingleton(
    TOKENS.PROJECT_REPOSITORY,
    () => new CombinedProjectRepository() // Use the combined implementation
  );

  container.registerSingleton(
    TOKENS.FILE_SYSTEM_REPOSITORY,
    () => new FileSystemRepositoryImpl()
  );

  container.registerSingleton(
    'ContextRepository',
    () => new ContextRepositoryImpl()
  );

  // Register services
  container.registerSingleton(
    TOKENS.EVENT_BUS,
    () => new EventBusImpl()
  );

  container.registerSingleton(
    'AIService',
    () => new AIServiceImpl()
  );

  // Register use cases with their dependencies
  container.registerTransient(
    TOKENS.LOAD_PROJECT_USE_CASE,
    (projectRepo) => 
      new LoadProjectUseCase(projectRepo),
    [TOKENS.PROJECT_REPOSITORY]
  );

  container.registerTransient(
    TOKENS.SELECT_FILES_USE_CASE,
    (projectRepo) => 
      new SelectProjectDirectoryUseCase(projectRepo),
    [TOKENS.PROJECT_REPOSITORY]
  );

  container.registerTransient(
    TOKENS.BUILD_CONTEXT_USE_CASE,
    (projectRepo, fileSystemRepo, contextRepo) => 
      new BuildContextUseCase(projectRepo, fileSystemRepo),
    [TOKENS.PROJECT_REPOSITORY, TOKENS.FILE_SYSTEM_REPOSITORY]
  );

  container.registerTransient(
    TOKENS.GENERATE_CODE_USE_CASE,
    (aiService) => new GenerateCodeUseCase(aiService),
    ['AIService']
  );

  console.log('✅ Dependency injection configured successfully');
}

/**
 * Get a configured use case instance
 */
export function getUseCase<T>(token: string): T {
  return container.resolve<T>(token);
}

/**
 * Get repository instance
 */
export function getRepository<T>(token: string): T {
  return container.resolve<T>(token);
}

/**
 * Get service instance
 */
export function getService<T>(token: string): T {
  return container.resolve<T>(token);
}

// Initialize dependencies on module load
if (typeof window !== 'undefined') {
  setupDependencies();
}