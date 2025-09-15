/**
 * Domain Services Index
 * 
 * Exports all domain services for centralized access.
 * These services encapsulate business logic following DDD principles.
 */

// Export domain services
export * from './StreamingPolicy';
export * from './TokenEstimationService';
export * from './LanguageDetectionService';
export * from './CodeValidationService';
export * from './MemoryManagementPolicy';
export * from './SecurityValidationService';
export * from './PerformanceMonitoringService';
export * from './LayoutManagementPolicy';
export * from './HtmlSanitizationService';
export * from './AuditLoggingService';
export * from './CircuitBreakerService';
export * from './ObservabilityService';

// Export default instances for convenience
export {
  defaultStreamingPolicy,
  StreamingPolicy,
  type MemoryOperation as StreamingMemoryOperation,
  type MemoryCheckResult
} from './StreamingPolicy';

export {
  defaultTokenEstimationService,
  TokenEstimationService,
  type TokenEstimationResult,
  type TokenEstimationOptions
} from './TokenEstimationService';

export {
  defaultLanguageDetectionService,
  LanguageDetectionService,
  type LanguageDetectionResult,
  type LanguageDetectionOptions
} from './LanguageDetectionService';

export {
  defaultCodeValidationService,
  CodeValidationService,
  type ValidationResult,
  type ValidationError,
  type ValidationWarning,
  type ValidationOptions,
  type CodeGenerationOptions
} from './CodeValidationService';

export {
  defaultMemoryManagementPolicy,
  MemoryManagementPolicy,
  type MemoryUsage,
  type MemoryOperation,
  type MemoryAssessment,
  type MemoryOptimizationSuggestion,
  type SystemMemoryInfo
} from './MemoryManagementPolicy';

export {
  defaultSecurityValidationService,
  SecurityValidationService,
  type PathValidationResult,
  type SecurityValidationOptions,
  type SecurityAuditEntry
} from './SecurityValidationService';



export {
  defaultPerformanceMonitoringService,
  PerformanceMonitoringService,
  type PerformanceMetrics,
  type PerformanceAlert,
  type CircuitBreakerConfig,
  type ObservabilityContext
} from './PerformanceMonitoringService';

export {
  createLayoutManagementPolicy,
  LayoutManagementPolicy,
  type WorkspaceMode,
  type PanelDockPosition,
  type LayoutPreset,
  type ViewportSize,
  type PanelState,
  type LayoutState,
  type PanelConfiguration,
  type LayoutValidationResult
} from './LayoutManagementPolicy';

export {
  defaultHtmlSanitizationService,
  HtmlSanitizationService,
  type SanitizationConfig,
  type SanitizationResult,
  type SanitizationOptions
} from './HtmlSanitizationService';

export {
  createAuditLoggingService,
  AuditLoggingService,
  type AuditEvent,
  type AuditEventType,
  type AuditSeverity,
  type AuditQuery,
  type AuditStatistics,
  type AuditRetentionPolicy
} from './AuditLoggingService';

export {
  defaultCircuitBreakerService,
  CircuitBreakerService,
  type CircuitState,
  type CircuitBreakerOptions,
  type CircuitBreakerStatus,
  type CircuitBreakerMetrics,
  type RetryConfig
} from './CircuitBreakerService';

export {
  defaultObservabilityService,
  ObservabilityService,
  type LogLevel,
  type LogEntry,
  type Span,
  type Metric,
  type ObservabilityConfig
} from './ObservabilityService';