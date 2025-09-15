import { 
  defaultLocalStorageService, 
  type LocalStorageService 
} from '@/domain/services/LocalStorageService';
import { 
  defaultSelectionManagementService, 
  type SelectionManagementService 
} from '@/domain/services/SelectionManagementService';
import { 
  defaultVirtualScrollingService, 
  type VirtualScrollingService 
} from '@/domain/services/VirtualScrollingService';
import { 
  defaultTooltipPositioningService, 
  type TooltipPositioningService 
} from '@/domain/services/TooltipPositioningService';
import { 
  defaultPerformanceMonitoringService, 
  type PerformanceMonitoringService 
} from '@/domain/services/PerformanceMonitoringService';
import { 
  defaultLayoutManagementPolicy, 
  type LayoutManagementPolicy 
} from '@/domain/services/LayoutManagementPolicy';
import { 
  defaultStreamingPolicy, 
  type StreamingPolicy 
} from '@/domain/services/StreamingPolicy';
import { 
  defaultTokenEstimationService, 
  type TokenEstimationService 
} from '@/domain/services/TokenEstimationService';
import { 
  defaultLanguageDetectionService, 
  type LanguageDetectionService 
} from '@/domain/services/LanguageDetectionService';
import { 
  defaultCodeValidationService, 
  type CodeValidationService 
} from '@/domain/services/CodeValidationService';
import { 
  defaultMemoryManagementPolicy, 
  type MemoryManagementPolicy 
} from '@/domain/services/MemoryManagementPolicy';
import { 
  defaultSecurityValidationService, 
  type SecurityValidationService 
} from '@/domain/services/SecurityValidationService';
import { 
  defaultHtmlSanitizationService, 
  type HtmlSanitizationService 
} from '@/domain/services/HtmlSanitizationService';
import { 
  defaultAuditLoggingService, 
  type AuditLoggingService 
} from '@/domain/services/AuditLoggingService';
import { 
  defaultCircuitBreakerService, 
  type CircuitBreakerService 
} from '@/domain/services/CircuitBreakerService';
import { 
  defaultObservabilityService, 
  type ObservabilityService 
} from '@/domain/services/ObservabilityService';
import type { StorageRepository } from '@/domain/repositories/StorageRepository';
import { LocalStorageAdapter } from '@/infrastructure/storage/LocalStorageAdapter';
import { container } from '@/infrastructure/container';
import type { SettingsRepository } from '@/domain/repositories/SettingsRepository';
import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';
import type { ContextRepository } from '@/domain/repositories/ContextRepository';

/**
 * Store Dependencies Interface
 * Defines all domain services available for dependency injection into stores
 * Following DDD and Clean Architecture principles
 */
export interface StoreDependencies {
  // Core services
  localStorageService: StorageRepository;
  selectionManagementService: SelectionManagementService;
  virtualScrollingService: VirtualScrollingService;
  tooltipPositioningService: TooltipPositioningService;
  
  // Performance and monitoring
  performanceMonitoringService: PerformanceMonitoringService;
  layoutManagementPolicy: LayoutManagementPolicy;
  streamingPolicy: StreamingPolicy;
  memoryManagementPolicy: MemoryManagementPolicy;
  
  // Language and validation
  tokenEstimationService: TokenEstimationService;
  languageDetectionService: LanguageDetectionService;
  codeValidationService: CodeValidationService;
  
  // Security services
  securityValidationService: SecurityValidationService;
  htmlSanitizationService: HtmlSanitizationService;
  auditLoggingService: AuditLoggingService;
  
  // Infrastructure services
  circuitBreakerService: CircuitBreakerService;
  observabilityService: ObservabilityService;
  
  // Repositories
  settingsRepository: SettingsRepository;
  projectRepository: ProjectRepository;
  contextRepository: ContextRepository;
  storageRepository: StorageRepository;
}

/**
 * Store Dependency Injection Container
 * Provides centralized dependency injection for all stores
 * Ensures consistent service instances across the application
 */
export class StoreDependencyContainer {
  private readonly dependencies: StoreDependencies;

  constructor(dependencies?: Partial<StoreDependencies>) {
    this.dependencies = {
      // Core services
      localStorageService: dependencies?.localStorageService || new LocalStorageAdapter(),
      selectionManagementService: dependencies?.selectionManagementService || defaultSelectionManagementService,
      virtualScrollingService: dependencies?.virtualScrollingService || defaultVirtualScrollingService,
      tooltipPositioningService: dependencies?.tooltipPositioningService || defaultTooltipPositioningService,
      
      // Performance and monitoring
      performanceMonitoringService: dependencies?.performanceMonitoringService || defaultPerformanceMonitoringService,
      layoutManagementPolicy: dependencies?.layoutManagementPolicy || defaultLayoutManagementPolicy,
      streamingPolicy: dependencies?.streamingPolicy || defaultStreamingPolicy,
      memoryManagementPolicy: dependencies?.memoryManagementPolicy || defaultMemoryManagementPolicy,
      
      // Language and validation
      tokenEstimationService: dependencies?.tokenEstimationService || defaultTokenEstimationService,
      languageDetectionService: dependencies?.languageDetectionService || defaultLanguageDetectionService,
      codeValidationService: dependencies?.codeValidationService || defaultCodeValidationService,
      
      // Security services
      securityValidationService: dependencies?.securityValidationService || defaultSecurityValidationService,
      htmlSanitizationService: dependencies?.htmlSanitizationService || defaultHtmlSanitizationService,
      auditLoggingService: dependencies?.auditLoggingService || defaultAuditLoggingService,
      
      // Infrastructure services
      circuitBreakerService: dependencies?.circuitBreakerService || defaultCircuitBreakerService,
      observabilityService: dependencies?.observabilityService || defaultObservabilityService,
      
      // Repositories
      settingsRepository: dependencies?.settingsRepository || container.settingsRepository,
      projectRepository: dependencies?.projectRepository || container.projectRepository,
      contextRepository: dependencies?.contextRepository || container.contextRepository,
      storageRepository: dependencies?.storageRepository || container.storageRepository
    };
  }

  /**
   * Gets all dependencies for injection into stores
   */
  getDependencies(): StoreDependencies {
    return this.dependencies;
  }

  /**
   * Gets a specific service by name
   */
  getService<K extends keyof StoreDependencies>(serviceName: K): StoreDependencies[K] {
    return this.dependencies[serviceName];
  }

  /**
   * Creates a new container with different service implementations
   * Useful for testing with mocked services
   */
  withServices(overrides: Partial<StoreDependencies>): StoreDependencyContainer {
    return new StoreDependencyContainer({
      ...this.dependencies,
      ...overrides
    });
  }

  /**
   * Validates that all required services are available
   */
  validate(): { valid: boolean; missing: string[] } {
    const required: (keyof StoreDependencies)[] = [
      'localStorageService',
      'performanceMonitoringService',
      'observabilityService'
    ];

    const missing = required.filter(serviceName => !this.dependencies[serviceName]);

    return {
      valid: missing.length === 0,
      missing
    };
  }

  /**
   * Gets performance monitoring service with additional context
   */
  getPerformanceMonitoringWithContext(storeId: string) {
    const service = this.dependencies.performanceMonitoringService;
    return {
      ...service,
      // Add store-specific context to performance monitoring
      logPerformance: (operation: string, duration: number, metadata?: any) => {
        service.logPerformance(operation, duration, {
          storeId,
          ...metadata
        });
      }
    };
  }

  /**
   * Gets observability service with store-specific tracing
   */
  getObservabilityWithTracing(storeId: string) {
    const service = this.dependencies.observabilityService;
    return {
      ...service,
      // Add store-specific tracing context
      startTrace: (operationName: string, metadata?: any) => {
        return service.startTrace(operationName, {
          storeId,
          component: 'store',
          ...metadata
        });
      },
      // Include finishSpan method
      finishSpan: (span: any, status: any, error?: any) => {
        return service.finishSpan(span, status, error);
      }
    };
  }

  /**
   * Gets audit logging service with store context
   */
  getAuditLoggingWithContext(storeId: string, userId?: string) {
    const service = this.dependencies.auditLoggingService;
    return {
      ...service,
      // Add store-specific audit context
      logOperation: (operation: string, details: any, severity: 'low' | 'medium' | 'high' = 'medium') => {
        service.logOperation(operation, {
          storeId,
          userId,
          ...details
        }, severity);
      }
    };
  }
}

/**
 * Default store dependency container instance
 * Used by all stores for consistent dependency injection
 */
export const defaultStoreDependencyContainer = new StoreDependencyContainer();

/**
 * Store mixin for consistent dependency injection
 * Provides a standard way for stores to access domain services
 */
export function createStoreWithDependencies<T>(
  storeId: string,
  storeFactory: (dependencies: StoreDependencies) => T,
  container: StoreDependencyContainer = defaultStoreDependencyContainer
): T {
  // Validate container
  const validation = container.validate();
  if (!validation.valid) {
    console.warn(`Store ${storeId} has missing dependencies:`, validation.missing);
  }

  // Get dependencies
  const dependencies = container.getDependencies();

  // Add logging for store creation in development
  if (process.env.NODE_ENV === 'development') {
    const observability = container.getObservabilityWithTracing(storeId);
    const trace = observability.startTrace('store_creation', { storeId });
    
    try {
      const store = storeFactory(dependencies);
      observability.finishSpan(trace, 'ok');
      return store;
    } catch (error) {
      observability.finishSpan(trace, 'error', error as Error);
      throw error;
    }
  }

  return storeFactory(dependencies);
}

/**
 * Store base class with dependency injection
 * Provides common functionality for all stores
 */
export abstract class BaseStore {
  protected readonly dependencies: StoreDependencies;
  protected readonly storeId: string;
  protected readonly performanceMonitoring: ReturnType<StoreDependencyContainer['getPerformanceMonitoringWithContext']>;
  protected readonly observability: ReturnType<StoreDependencyContainer['getObservabilityWithTracing']>;
  protected readonly auditLogging: ReturnType<StoreDependencyContainer['getAuditLoggingWithContext']>;

  constructor(
    storeId: string, 
    container: StoreDependencyContainer = defaultStoreDependencyContainer
  ) {
    this.storeId = storeId;
    this.dependencies = container.getDependencies();
    this.performanceMonitoring = container.getPerformanceMonitoringWithContext(storeId);
    this.observability = container.getObservabilityWithTracing(storeId);
    this.auditLogging = container.getAuditLoggingWithContext(storeId);
  }

  /**
   * Helper method for performance monitoring of store operations
   */
  protected async measureOperation<T>(
    operationName: string,
    operation: () => Promise<T> | T,
    metadata?: any
  ): Promise<T> {
    const startTime = performance.now();
    const trace = this.observability.startTrace(operationName, metadata);

    try {
      const result = await operation();
      const duration = performance.now() - startTime;
      
      this.performanceMonitoring.logPerformance(operationName, duration, metadata);
      this.observability.finishSpan(trace, 'ok');
      
      return result;
    } catch (error) {
      const duration = performance.now() - startTime;
      
      this.performanceMonitoring.logPerformance(operationName, duration, { 
        ...metadata, 
        error: true 
      });
      this.observability.finishSpan(trace, 'error', error as Error);
      
      throw error;
    }
  }

  /**
   * Helper method for auditing store operations
   */
  protected auditOperation(
    operation: string, 
    details: any, 
    severity: 'low' | 'medium' | 'high' = 'medium'
  ): void {
    this.auditLogging.logOperation(operation, details, severity);
  }

  /**
   * Helper method for storing data with audit trail
   */
  protected setWithAudit<T>(
    key: string, 
    value: T, 
    options?: { expiresIn?: number; sessionOnly?: boolean }
  ): void {
    this.dependencies.localStorageService.set(key, value, options);
    this.auditOperation('data_stored', { key, hasValue: !!value }, 'low');
  }

  /**
   * Helper method for retrieving data with audit trail
   */
  protected async getWithAudit<T>(key: string, defaultValue?: T): Promise<T | undefined> {
    try {
      // Use get instead of getItem
      const value = this.dependencies.localStorageService.get<T>(key, defaultValue);
      this.auditOperation('data_retrieved', { key, hasValue: !!value }, 'low');
      return value;
    } catch (error) {
      console.error('Failed to retrieve data with audit trail:', error);
      return defaultValue;
    }
  }
}

// Export type for store factory functions
export type StoreFactory<T> = (dependencies: StoreDependencies) => T;