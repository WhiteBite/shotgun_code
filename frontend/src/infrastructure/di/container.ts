/**
 * Dependency Injection Container
 * Implements IoC container for Clean Architecture dependencies
 */

interface ServiceDefinition<T = any> {
  factory: (...deps: any[]) => T;
  dependencies?: string[];
  singleton?: boolean;
  instance?: T;
}

interface Container {
  register<T>(token: string, definition: ServiceDefinition<T>): void;
  resolve<T>(token: string): T;
  registerSingleton<T>(token: string, factory: (...deps: any[]) => T, dependencies?: string[]): void;
  registerTransient<T>(token: string, factory: (...deps: any[]) => T, dependencies?: string[]): void;
  clear(): void;
}

class DIContainer implements Container {
  private services = new Map<string, ServiceDefinition>();
  private resolving = new Set<string>();

  register<T>(token: string, definition: ServiceDefinition<T>): void {
    this.services.set(token, definition);
  }

  registerSingleton<T>(token: string, factory: (...deps: any[]) => T, dependencies: string[] = []): void {
    this.register(token, {
      factory,
      dependencies,
      singleton: true
    });
  }

  registerTransient<T>(token: string, factory: (...deps: any[]) => T, dependencies: string[] = []): void {
    this.register(token, {
      factory,
      dependencies,
      singleton: false
    });
  }

  resolve<T>(token: string): T {
    // Prevent circular dependencies
    if (this.resolving.has(token)) {
      throw new Error(`Circular dependency detected: ${token}`);
    }

    const service = this.services.get(token);
    if (!service) {
      throw new Error(`Service not registered: ${token}`);
    }

    // Return singleton instance if available
    if (service.singleton && service.instance) {
      return service.instance as T;
    }

    // Resolve dependencies
    this.resolving.add(token);
    const dependencies = (service.dependencies || []).map(dep => this.resolve(dep));
    this.resolving.delete(token);

    // Create instance
    const instance = service.factory(...dependencies);

    // Store singleton instance
    if (service.singleton) {
      service.instance = instance;
    }

    return instance as T;
  }

  clear(): void {
    this.services.clear();
    this.resolving.clear();
  }
}

// Global container instance
export const container = new DIContainer();

// Service tokens
export const TOKENS = {
  // Repositories
  PROJECT_REPOSITORY: 'ProjectRepository',
  FILE_SYSTEM_REPOSITORY: 'FileSystemRepository',
  SETTINGS_REPOSITORY: 'SettingsRepository',
  
  // Use Cases
  LOAD_PROJECT_USE_CASE: 'LoadProjectUseCase',
  SELECT_FILES_USE_CASE: 'SelectFilesUseCase',
  BUILD_CONTEXT_USE_CASE: 'BuildContextUseCase',
  GENERATE_CODE_USE_CASE: 'GenerateCodeUseCase',
  
  // Services
  API_CLIENT: 'ApiClient',
  EVENT_BUS: 'EventBus',
  FILE_WATCHER_SERVICE: 'FileWatcherService',
  TOKEN_ESTIMATOR_SERVICE: 'TokenEstimatorService',
  
  // Infrastructure
  STORAGE_ADAPTER: 'StorageAdapter',
  EXTERNAL_API_ADAPTER: 'ExternalApiAdapter'
} as const;

export type ServiceToken = typeof TOKENS[keyof typeof TOKENS];

// Helper functions
export function inject<T>(token: ServiceToken): T {
  return container.resolve<T>(token);
}

export function registerService<T>(
  token: ServiceToken,
  factory: (...deps: any[]) => T,
  dependencies: ServiceToken[] = [],
  singleton = true
): void {
  if (singleton) {
    container.registerSingleton(token, factory, dependencies);
  } else {
    container.registerTransient(token, factory, dependencies);
  }
}