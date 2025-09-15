/**
 * Configuration Type Definitions
 * 
 * Comprehensive TypeScript interfaces for all configuration aspects
 * to ensure type safety across the application following DDD principles.
 */

// ========== Risk and Memory Management Types ==========

export type RiskLevel = 'low' | 'medium' | 'high' | 'critical';

export interface MemoryThresholds {
  readonly MAX_FILE_LIMIT: number;
  readonly MAX_INDIVIDUAL_FILE_SIZE: number;
  readonly CRITICAL_MEMORY_THRESHOLD: number;
  readonly MEMORY_CLEANUP_INTERVAL: number;
  readonly MAX_MEMORY_INCREASE: number;
  readonly MAX_PEAK_USAGE_PERCENT: number;
  readonly DEFAULT_CHARS_PER_TOKEN: number;
}

export interface TokenConfiguration {
  readonly DEFAULT_CHARS_PER_TOKEN: number;
  readonly MAX_TOKENS_PER_REQUEST: number;
  readonly ESTIMATION_MULTIPLIER: number;
}

export interface StreamingConfiguration {
  readonly HIGH_RISK_THRESHOLD: RiskLevel;
  readonly CRITICAL_RISK_THRESHOLD: RiskLevel;
  readonly CHUNK_SIZE: number;
  readonly STRIP_COMMENTS: boolean;
  readonly INCLUDE_MANIFEST: boolean;
}

export interface PerformanceLimits {
  readonly MAX_CONSOLE_LOGS: number;
  readonly MAX_SELECTED_PATHS: number;
  readonly MAX_EXPANDED_PATHS: number;
  readonly CLEANUP_INTERVAL: number;
  readonly MAX_CACHE_SIZE: number;
}

export interface MonitoringConfiguration {
  readonly METRICS_RETENTION_COUNT: number;
  readonly ALERTS_RETENTION_COUNT: number;
  readonly TRACE_CLEANUP_THRESHOLD_MS: number;
  readonly PERFORMANCE_LOG_ENABLED: boolean;
  readonly SLOW_RESPONSE_THRESHOLD_MS: number;
}

export interface CircuitBreakerConfiguration {
  readonly DEFAULT_FAILURE_THRESHOLD: number;
  readonly DEFAULT_RESET_TIMEOUT_MS: number;
  readonly DEFAULT_MONITORING_PERIOD_MS: number;
  readonly HEALTH_CHECK_INTERVAL_MS: number;
}

export interface PerformanceConfiguration {
  readonly memory: MemoryThresholds;
  readonly tokens: TokenConfiguration;
  readonly streaming: StreamingConfiguration;
  readonly limits: PerformanceLimits;
  readonly monitoring: MonitoringConfiguration;
  readonly circuitBreaker: CircuitBreakerConfiguration;
}

// ========== Context Configuration Types ==========

export interface RiskThresholds {
  readonly LOW: number;
  readonly MEDIUM: number;
  readonly HIGH: number;
  readonly CRITICAL: number;
}

export interface ContextEstimation {
  readonly DEFAULT_ESTIMATED_SIZE: number;
  readonly RISK_THRESHOLDS: RiskThresholds;
}

export interface ContextStreamingSettings {
  readonly STRIP_COMMENTS: boolean;
  readonly INCLUDE_MANIFEST: boolean;
  readonly CHUNK_SIZE: number;
  readonly MAX_LINES: number;
}

export interface ContextPagination {
  readonly DEFAULT_PAGE_SIZE: number;
  readonly MAX_PAGE_SIZE: number;
}

export interface ContextConfiguration {
  readonly estimation: ContextEstimation;
  readonly streaming: ContextStreamingSettings;
  readonly pagination: ContextPagination;
}

// ========== Security Configuration Types ==========

export interface SanitizationSettings {
  readonly ALLOWED_HTML_TAGS: readonly string[];
  readonly ALLOWED_ATTRIBUTES: readonly string[];
  readonly STRIP_SCRIPTS: boolean;
}

export interface WorkspaceBoundaryChecks {
  readonly ENFORCE_WORKSPACE_BOUNDARY: boolean;
  readonly ALLOW_SYMLINKS: boolean;
  readonly ALLOW_HIDDEN_FILES: boolean;
  readonly ALLOW_SYSTEM_FILES: boolean;
  readonly MAX_SYMLINK_DEPTH: number;
}

export interface ValidationPolicies {
  readonly MAX_PATH_LENGTH: number;
  readonly MAX_FILENAME_LENGTH: number;
  readonly MAX_DIRECTORY_DEPTH: number;
  readonly PATH_VALIDATION_ENABLED: boolean;
  readonly FORBIDDEN_PATTERNS: readonly RegExp[];
  readonly FORBIDDEN_EXTENSIONS: readonly string[];
  readonly ALLOWED_EXTENSIONS: readonly string[];
  readonly WORKSPACE_BOUNDARY_CHECKS: WorkspaceBoundaryChecks;
}

export interface ErrorPatterns {
  readonly MEMORY_ERROR_PATTERNS: readonly RegExp[];
  readonly CONTEXT_SIZE_ERROR_PATTERNS: readonly RegExp[];
}

export interface AuditConfiguration {
  readonly LOG_FAILED_VALIDATIONS: boolean;
  readonly LOG_SUSPICIOUS_ACTIVITY: boolean;
  readonly MAX_AUDIT_LOG_SIZE: number;
  readonly AUDIT_LOG_RETENTION_DAYS: number;
}

export interface SecurityConfiguration {
  readonly sanitization: SanitizationSettings;
  readonly validation: ValidationPolicies;
  readonly errorPatterns: ErrorPatterns;
  readonly audit: AuditConfiguration;
}

// ========== UI Configuration Types ==========

export interface ComponentSelection {
  readonly HEADER_BAR: string;
  readonly BOTTOM_CONSOLE: string;
  readonly CONTEXT_VIEWER: string;
  readonly FILE_TREE: string;
}

export interface FeatureFlags {
  readonly ENABLE_CONSOLE: boolean;
  readonly ENABLE_SPLIT: boolean;
  readonly ENABLE_VIRTUAL_MODE: boolean;
  readonly ENABLE_RESIZE_HANDLES: boolean;
  readonly ENABLE_MEMORY_MONITORING: boolean;
}

export interface LayoutSettings {
  readonly DEFAULT_CONTEXT_PANEL_WIDTH: number;
  readonly DEFAULT_RESULTS_PANEL_WIDTH: number;
  readonly MIN_PANEL_WIDTH: number;
  readonly MAX_PANEL_WIDTH_PERCENTAGE: number;
  readonly MIN_CENTER_AREA_WIDTH: number;
  readonly GRID: {
    readonly NAVIGATOR_COLUMNS: string;
    readonly MIN_NAVIGATOR_WIDTH: number;
  };
}

export interface TooltipConfiguration {
  readonly MAX_CONCURRENT_TOOLTIPS: number;
  readonly CACHE_SIZE: number;
  readonly DEFAULT_DELAY: number;
  readonly HIDE_DELAY: number;
  readonly THROTTLE_DELAY: number;
  readonly MAX_WIDTH: number;
  readonly VIRTUALIZATION_THRESHOLD: number;
  readonly DEBOUNCE_DELAY: number;
  readonly ARROW_SIZE: number;
  readonly SCALE_INITIAL: number;
  readonly SCALE_FINAL: number;
  readonly HIGH_CONTRAST_BORDER_WIDTH: number;
  readonly DEFAULT_OFFSET: number;
  readonly VIEWPORT_PADDING: number;
}

export interface SelectionConfiguration {
  readonly MAX_CASCADE_FILES: number;
  readonly MAX_RECURSION_DEPTH: number;
}

export interface VirtualScrollConfiguration {
  readonly MAX_CACHE_SIZE: number;
  readonly MEMORY_CLEANUP_INTERVAL: number;
  readonly CRITICAL_MEMORY_THRESHOLD: number;
  readonly BUFFER_SIZE: number;
  readonly ITEM_HEIGHT: number;
  readonly LINE_HEIGHT: number;
  readonly DEFAULT_BUFFER: number;
  readonly DEFAULT_OVERSCAN: number;
  readonly PERFORMANCE_THRESHOLD: number;
  readonly VIRTUALIZATION_THRESHOLD: number;
}

export interface ResponsiveBreakpoints {
  readonly XS: number;
  readonly SM: number;
  readonly MD: number;
  readonly LG: number;
  readonly XL: number;
}

export interface KeyboardConfiguration {
  readonly DEBOUNCE_DELAY: number;
  readonly LONG_PRESS_DURATION: number;
}

export interface SplitPaneConfiguration {
  readonly DEFAULT_RATIO: number;
  readonly MIN_SIZE: number;
  readonly SNAP_THRESHOLD: number;
}

export interface PanelConfiguration {
  readonly MIN_WIDTH: number;
  readonly MAX_WIDTH_RATIO: number;
  readonly RESIZE_HANDLE_SIZE: number;
  readonly CONTEXT_PANEL_MIN_WIDTH: number;
  readonly CONTEXT_PANEL_MAX_WIDTH: number;
  readonly RESULTS_PANEL_MIN_WIDTH: number;
  readonly RESULTS_PANEL_MAX_WIDTH: number;
  readonly BORDER_RADIUS: number;
  readonly HEADER_PADDING: number;
  readonly CONTENT_PADDING: number;
  readonly FOOTER_PADDING: number;
  readonly COLLAPSED_WIDTH: number;
  readonly RESIZE_HANDLE_WIDTH: number;
  readonly HEADER_HEIGHT: number;
  readonly ICON_SIZE: number;
  readonly BUTTON_ICON_SIZE: number;
  readonly ACTION_BUTTON_SIZE: number;
  readonly GAP_SM: number;
  readonly GAP_MD: number;
  readonly GAP_LG: number;
  readonly SCROLLBAR_WIDTH: number;
  readonly SCROLLBAR_TRACK_RADIUS: number;
  readonly SCROLLBAR_THUMB_RADIUS: number;
  readonly MOBILE_PADDING: number;
  readonly MOBILE_BORDER_RADIUS: number;
  readonly MOBILE_HEADER_RADIUS: number;
}

export interface WorkspaceConfiguration {
  readonly DEFAULT_CONTEXT_PANEL_WIDTH: number;
  readonly DEFAULT_RESULTS_PANEL_WIDTH: number;
  readonly DEFAULT_LAYOUT_PRESET: string;
  readonly ENABLE_ANIMATIONS: boolean;
}

export interface UIConfiguration {
  readonly components: ComponentSelection;
  readonly features: FeatureFlags;
  readonly layout: LayoutSettings;
  readonly tooltips: TooltipConfiguration;
  readonly selection: SelectionConfiguration;
  readonly virtualScroll: VirtualScrollConfiguration;
  readonly responsive: {
    readonly BREAKPOINTS: ResponsiveBreakpoints;
  };
  readonly keyboard: KeyboardConfiguration;
  readonly splitPane: SplitPaneConfiguration;
  readonly panels: PanelConfiguration;
  readonly workspace: WorkspaceConfiguration;
}

// ========== File Tree Configuration Types ==========

export interface FileTreeLimits {
  readonly MAX_SELECTED_FILES: number;
  readonly MAX_NODES: number;
  readonly MAX_DEPTH: number;
}

export interface VirtualizationSettings {
  readonly ENABLED: boolean;
  readonly ITEM_HEIGHT: number;
  readonly BUFFER_SIZE: number;
}

export interface FileTreeConfiguration {
  readonly limits: FileTreeLimits;
  readonly virtualization: VirtualizationSettings;
}

// ========== Language Detection Types ==========

export interface LanguageExtensionMap {
  readonly [extension: string]: string;
}

export interface LanguageDetectionConfiguration {
  readonly EXTENSION_MAP: LanguageExtensionMap;
  readonly DEFAULT_LANGUAGE: string;
  readonly FALLBACK_DETECTION: boolean;
}

// ========== AI Configuration Types ==========

export interface AISettings {
  readonly provider: string;
  readonly model: string;
  readonly temperature: number;
  readonly maxTokens: number;
}

export interface SystemPrompts {
  readonly DEFAULT: string;
  readonly CODE_GENERATION: string;
  readonly CODE_REVIEW: string;
}

export interface AIValidation {
  readonly SYNTAX_CHECK: boolean;
  readonly STRUCTURE_CHECK: boolean;
  readonly SECURITY_CHECK: boolean;
}

export interface AIConfiguration {
  readonly defaultSettings: AISettings;
  readonly systemPrompts: SystemPrompts;
  readonly validation: AIValidation;
}

// ========== Autonomous Configuration Types ==========

export interface SLAPolicy {
  readonly maxConcurrentTasks: number;
  readonly maxRetries: number;
  readonly timeoutPerTask: number;
  readonly qualityChecks: readonly string[];
}

export interface SLAPolicies {
  readonly lite: SLAPolicy;
  readonly standard: SLAPolicy;
  readonly strict: SLAPolicy;
}

export interface AutonomousConfiguration {
  readonly slaPolicies: SLAPolicies;
  readonly defaultSLA: keyof SLAPolicies;
}

// ========== Reports Configuration Types ==========

export interface IDGeneration {
  readonly PREFIX: string;
  readonly TIMESTAMP_FORMAT: string;
  readonly RANDOM_SUFFIX_LENGTH: number;
}

export interface ExportSettings {
  readonly DEFAULT_FORMAT: string;
  readonly SUPPORTED_FORMATS: readonly string[];
  readonly MAX_EXPORT_SIZE: number;
}

export interface ReportsConfiguration {
  readonly idGeneration: IDGeneration;
  readonly export: ExportSettings;
}

// ========== Project Configuration Types ==========

export interface ProjectExtraction {
  readonly NAME_FROM_PATH: boolean;
  readonly FALLBACK_NAME: string;
}

export interface RecentProjects {
  readonly MAX_COUNT: number;
  readonly STORAGE_KEY: string;
}

export interface ProjectConfiguration {
  readonly extraction: ProjectExtraction;
  readonly recentProjects: RecentProjects;
}

// ========== QuickLook Configuration Types ==========

export interface QuickLookConfiguration {
  readonly DEFAULT_LANGUAGE: string;
  readonly SYNTAX_HIGHLIGHTING: boolean;
  readonly MAX_FILE_SIZE: number;
  readonly PREVIEW_LINES: number;
}

// ========== Events Configuration Types ==========

export interface EventsConfiguration {
  readonly DEBOUNCE_DELAY: number;
  readonly MAX_EVENT_QUEUE_SIZE: number;
  readonly ERROR_RETRY_ATTEMPTS: number;
}

// ========== Storage Configuration Types ==========

export interface StorageConfiguration {
  readonly UI_PREFERENCES_PREFIX: string;
  readonly SESSION_STORAGE_PREFIX: string;
  readonly CACHE_EXPIRY: number;
  readonly CLEANUP_INTERVAL: number;
}

// ========== Main Application Configuration ==========

export interface AppConfiguration {
  readonly PRIMARY_WORKSPACE_VIEW: string;
  readonly ui: UIConfiguration;
  readonly performance: PerformanceConfiguration;
  readonly context: ContextConfiguration;
  readonly security: SecurityConfiguration;
  readonly fileTree: FileTreeConfiguration;
  readonly languageDetection: LanguageDetectionConfiguration;
  readonly ai: AIConfiguration;
  readonly autonomous: AutonomousConfiguration;
  readonly reports: ReportsConfiguration;
  readonly project: ProjectConfiguration;
  readonly quicklook: QuickLookConfiguration;
  readonly events: EventsConfiguration;
  readonly storage: StorageConfiguration;
}

// ========== Helper Function Types ==========

export type ConfigurationPath = keyof AppConfiguration;
export type UIComponentKey = keyof ComponentSelection;
export type FeatureKey = keyof FeatureFlags;
export type LayoutKey = keyof LayoutSettings;
export type PerformanceLimitKey = keyof PerformanceLimits;
export type MemoryThresholdKey = keyof MemoryThresholds;
export type RiskThresholdKey = keyof RiskThresholds;
export type SLAPolicyKey = keyof SLAPolicies;

// ========== Configuration Validation Types ==========

export interface ConfigurationValidationResult {
  readonly isValid: boolean;
  readonly errors: readonly string[];
  readonly warnings: readonly string[];
}

export interface ConfigurationValidator {
  validate(config: AppConfiguration): ConfigurationValidationResult;
  validateSection<T>(section: T, sectionName: string): ConfigurationValidationResult;
}

// ========== Domain Service Configuration Types ==========

export interface StreamingPolicyConfiguration {
  readonly highRiskThreshold: RiskLevel;
  readonly criticalRiskThreshold: RiskLevel;
  readonly memoryThresholds: MemoryThresholds;
  readonly riskThresholds: RiskThresholds;
}

export interface TokenEstimationConfiguration {
  readonly charsPerToken: TokenConfiguration;
  readonly languageMultipliers: Readonly<Record<string, number>>;
}

export interface LanguageDetectionServiceConfiguration {
  readonly extensionMap: LanguageExtensionMap;
  readonly defaultLanguage: string;
  readonly fallbackDetection: boolean;
}

export interface CodeValidationServiceConfiguration {
  readonly syntaxValidation: AIValidation;
  readonly securityValidation: SecurityConfiguration;
}

export interface MemoryManagementPolicyConfiguration {
  readonly thresholds: MemoryThresholds;
  readonly limits: PerformanceLimits;
  readonly riskAssessment: RiskThresholds;
}