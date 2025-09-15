/**
 * Application Configuration
 * 
 * Centralized configuration following DDD and Clean Architecture principles.
 * This eliminates configuration fragmentation and provides type-safe access
 * to all application settings.
 */

import type { AppConfiguration, RiskLevel } from '@/types/configuration';

export const APP_CONFIG: AppConfiguration = {
  // Primary workspace view
  PRIMARY_WORKSPACE_VIEW: 'EnhancedWorkspaceView',
  
  // UI Configuration
  ui: {
    components: {
      HEADER_BAR: 'HeaderBar',
      BOTTOM_CONSOLE: 'BottomConsole',
      CONTEXT_VIEWER: 'ContextViewer',
      FILE_TREE: 'FilePanelModern'
    },
    
    features: {
      ENABLE_CONSOLE: true,
      ENABLE_SPLIT: true,
      ENABLE_VIRTUAL_MODE: true,
      ENABLE_RESIZE_HANDLES: true,
      ENABLE_MEMORY_MONITORING: true
    },
    
    layout: {
      DEFAULT_CONTEXT_PANEL_WIDTH: 400,
      DEFAULT_RESULTS_PANEL_WIDTH: 500,
      MIN_PANEL_WIDTH: 200,
      MAX_PANEL_WIDTH_PERCENTAGE: 0.4,
      MIN_CENTER_AREA_WIDTH: 300,
      DEFAULT_GAP: 16,
      MOBILE_GAP: 8,
      MIN_ITEM_WIDTH: '250px',
      SIDEBAR_WIDTH: 300,
      HEADER_HEIGHT: 60,
      FOOTER_HEIGHT: 40,
      // Margin configuration for responsive design
      RESPONSIVE_MARGIN: 40,
      GRID: {
        NAVIGATOR_COLUMNS: 'minmax(280px, 1fr)',
        MIN_NAVIGATOR_WIDTH: 280
      }
    },
    
    tooltips: {
      MAX_CONCURRENT_TOOLTIPS: 5,
      CACHE_SIZE: 100,
      DEFAULT_DELAY: 500,
      HIDE_DELAY: 200,
      THROTTLE_DELAY: 100,
      MAX_WIDTH: 300,
      VIRTUALIZATION_THRESHOLD: 20,
      DEBOUNCE_DELAY: 150,
      ARROW_SIZE: 8,
      SCALE_INITIAL: 0.95,
      SCALE_FINAL: 1.0,
      HIGH_CONTRAST_BORDER_WIDTH: 2,
      
      // Positioning service configuration
      DEFAULT_OFFSET: 8,
      VIEWPORT_PADDING: 8,
      
      // Smart tooltip configuration
      DEFAULT_ALLOW_HTML: true,
      DEFAULT_INTERACTIVE: true,
      DEFAULT_SIZE: 'md'
    },
    
    selection: {
      MAX_CASCADE_FILES: 100,
      MAX_RECURSION_DEPTH: 5
    },
    
    virtualScroll: {
      MAX_CACHE_SIZE: 1000,
      MEMORY_CLEANUP_INTERVAL: 30000,
      CRITICAL_MEMORY_THRESHOLD: 0.85,
      BUFFER_SIZE: 10,
      ITEM_HEIGHT: 20,
      LINE_HEIGHT: 20,
      // Domain service configuration
      DEFAULT_BUFFER: 5,
      DEFAULT_OVERSCAN: 3,
      PERFORMANCE_THRESHOLD: 16, // 16ms target for 60fps
      VIRTUALIZATION_THRESHOLD: 100,
      // Virtual scroll viewer specific parameters
      VIEWER_MAX_CACHE_SIZE: 1, // Critical: Only keep 1 chunk in memory
      VIEWER_MEMORY_CLEANUP_INTERVAL: 1000, // Cleanup every 1 second
      VIEWER_CRITICAL_MEMORY_THRESHOLD: 10, // 10MB threshold
      VIEWER_MAX_LINES: 5000, // Hard limit to prevent memory issues
      VIEWER_MAX_VISIBLE_LINES: 200, // Limit visible lines to prevent rendering issues
      VIEWER_CHUNK_SIZE: 50, // Group lines into chunks for batch loading
      VIEWER_OVERSCAN: 1, // Reduced overscan for better performance
      VIEWER_MEMORY_CHECK_INTERVAL: 500, // Check memory every 500ms
      VIEWER_MEMORY_CLEANUP_THRESHOLD: 5 // Start cleanup at 50% of threshold
    },
    
    quicklook: {
      MAX_CHARS: 400000,
      SCROLL_INDICATOR_HIDE_MS: 2000
    },
    
    responsive: {
      BREAKPOINTS: {
        XS: 480,
        SM: 768,
        MD: 1024,
        LG: 1280,
        XL: 1536
      }
    },
    
    keyboard: {
      DEBOUNCE_DELAY: 150,
      LONG_PRESS_DURATION: 500
    },
    
    splitPane: {
      DEFAULT_RATIO: 0.5,
      MIN_SIZE: 100,
      SNAP_THRESHOLD: 10
    },
    
    panels: {
      MIN_WIDTH: 200,
      MIN_HEIGHT: 150,
      MAX_WIDTH_RATIO: 0.8,
      RESIZE_HANDLE_SIZE: 4,
      // Workspace panel constants
      CONTEXT_PANEL_MIN_WIDTH: 300,
      CONTEXT_PANEL_MAX_WIDTH: 600,
      RESULTS_PANEL_MIN_WIDTH: 400,
      RESULTS_PANEL_MAX_WIDTH: 700,
      
      // Panel layout constants
      BORDER_RADIUS: 12,
      HEADER_PADDING: 16,
      CONTENT_PADDING: 16,
      FOOTER_PADDING: 16,
      COLLAPSED_WIDTH: 48,
      RESIZE_HANDLE_WIDTH: 4,
      HEADER_HEIGHT: 60,
      
      // Icon sizes
      ICON_SIZE: 20,
      BUTTON_ICON_SIZE: 16,
      ACTION_BUTTON_SIZE: 32,
      
      // Gaps and spacing
      GAP_SM: 8,
      GAP_MD: 12,
      GAP_LG: 16,
      
      // Scrollbar dimensions
      SCROLLBAR_WIDTH: 8,
      SCROLLBAR_TRACK_RADIUS: 4,
      SCROLLBAR_THUMB_RADIUS: 4,
      
      // Resize handle configuration
      RESIZE_INDICATOR_WIDTH: 2,
      RESIZE_INDICATOR_HEIGHT: 24,
      
      // Mobile responsive
      MOBILE_PADDING: 12,
      MOBILE_BORDER_RADIUS: 8,
      MOBILE_HEADER_RADIUS: 7
    },
    
    workspace: {
      DEFAULT_CONTEXT_PANEL_WIDTH: 400,
      DEFAULT_RESULTS_PANEL_WIDTH: 500,
      DEFAULT_LAYOUT_PRESET: 'default',
      ENABLE_ANIMATIONS: true
    }
  },
  
  // Performance and Memory Management
  performance: {
    memory: {
      MAX_FILE_LIMIT: 10,
      MAX_INDIVIDUAL_FILE_SIZE: 32 * 1024, // 32KB
      CRITICAL_MEMORY_THRESHOLD: 0.85,
      MEMORY_CLEANUP_INTERVAL: 30000,
      MAX_MEMORY_INCREASE: 100 * 1024 * 1024, // 100MB
      MAX_PEAK_USAGE_PERCENT: 0.90,
      DEFAULT_CHARS_PER_TOKEN: 4,
      // Memory monitoring constants
      WARNING_THRESHOLD_MB: 20,
      CRITICAL_THRESHOLD_MB: 30,
      POLLING_INTERVAL_MS: 5000,
      SHOW_TOASTS: false,
      AUTO_CLEANUP: true
    },
    
    tokens: {
      DEFAULT_CHARS_PER_TOKEN: 4,
      MAX_TOKENS_PER_REQUEST: 10000,
      ESTIMATION_MULTIPLIER: 1.2
    },
    
    streaming: {
      HIGH_RISK_THRESHOLD: 'high' as RiskLevel,
      CRITICAL_RISK_THRESHOLD: 'critical' as RiskLevel,
      CHUNK_SIZE: 1000,
      STRIP_COMMENTS: true,
      INCLUDE_MANIFEST: true
    },
    
    limits: {
      MAX_CONSOLE_LOGS: 200,
      MAX_SELECTED_PATHS: 500,
      MAX_EXPANDED_PATHS: 200,
      CLEANUP_INTERVAL: 60000,
      MAX_CACHE_SIZE: 1000
    },
    
    monitoring: {
      METRICS_RETENTION_COUNT: 100,
      ALERTS_RETENTION_COUNT: 20,
      TRACE_CLEANUP_THRESHOLD_MS: 5 * 60 * 1000, // 5 minutes
      PERFORMANCE_LOG_ENABLED: true,
      SLOW_RESPONSE_THRESHOLD_MS: 5000 // 5 seconds
    },
    
    circuitBreaker: {
      DEFAULT_FAILURE_THRESHOLD: 5,
      DEFAULT_RESET_TIMEOUT_MS: 60000, // 1 minute
      DEFAULT_MONITORING_PERIOD_MS: 10000, // 10 seconds
      HEALTH_CHECK_INTERVAL_MS: 30000 // 30 seconds
    }
  },
  
  // Context Configuration
  context: {
    estimation: {
      DEFAULT_ESTIMATED_SIZE: 1024 * 1024, // 1MB
      RISK_THRESHOLDS: {
        LOW: 10 * 1024 * 1024,    // 10MB
        MEDIUM: 50 * 1024 * 1024, // 50MB
        HIGH: 100 * 1024 * 1024,  // 100MB
        CRITICAL: 500 * 1024 * 1024 // 500MB
      }
    },
    
    streaming: {
      STRIP_COMMENTS: true,
      INCLUDE_MANIFEST: true,
      CHUNK_SIZE: 1000,
      MAX_LINES: 5000
    },
    
    // Context chunking configuration
    chunking: {
      COPY_HISTORY_LIMIT: 50
    },
    
    // Context splitter configuration
    splitter: {
      AVG_CHARS_PER_TOKEN: 4,
      MIN_STEP_CHARS: 1,
      MIN_OVERLAP_CHARS: 0
    },
    
    pagination: {
      DEFAULT_PAGE_SIZE: 50,
      MAX_PAGE_SIZE: 200
    }
  },
  
  // Security Configuration
  security: {
    sanitization: {
      ALLOWED_HTML_TAGS: ['strong', 'em', 'code', 'pre', 'br'],
      ALLOWED_ATTRIBUTES: ['class', 'id'],
      STRIP_SCRIPTS: true
    },
    
    validation: {
      MAX_PATH_LENGTH: 2048,
      MAX_FILENAME_LENGTH: 255,
      MAX_DIRECTORY_DEPTH: 10,
      PATH_VALIDATION_ENABLED: true,
      
      FORBIDDEN_PATTERNS: [
        /\.\.[\/\\]/,  // Parent directory traversal
        /[<>:"|?*]/,   // Windows invalid characters
        /\x00/,        // Null bytes
        /^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])$/i, // Windows reserved names
        /[\/\\]\s/,    // Paths with trailing spaces
        /\s[\/\\]/,    // Paths with leading spaces
        /^\s/,         // Leading whitespace
        /\s$/,         // Trailing whitespace
        /[\x00-\x1f\x7f-\x9f]/, // Control characters
        /javascript:/i, // JavaScript protocol
        /vbscript:/i,  // VBScript protocol
        /data:.*script/i, // Data URI with script
        /\.\.\/\.\.\//, // Multiple parent directory traversals
        /~[^/\\]*[\/\\]/, // User directory access
        /proc[\/\\]/,  // Linux proc filesystem
        /dev[\/\\]/,   // Device files
        /etc[\/\\]/,   // System configuration
        /var[\/\\]log/, // Log files
        /tmp[\/\\]/,   // Temporary files
        /system32[\/\\]/i, // Windows system directory
        /windows[\/\\]/i, // Windows directory
        /program files/i, // Program files directory
        /\$[A-Z_]+/,   // Environment variables
        /%[A-Z_]+%/,   // Windows environment variables
      ],
      
      FORBIDDEN_EXTENSIONS: [
        // Executable files
        '.exe', '.bat', '.cmd', '.scr', '.com', '.pif', '.msi', '.msu',
        '.dll', '.sys', '.drv', '.vxd', '.386', '.cpl', '.ocx',
        
        // Script files
        '.vbs', '.vbe', '.jse', '.wsf', '.wsh', '.ps1', '.ps1xml',
        '.ps2', '.ps2xml', '.psc1', '.psc2', '.msh', '.msh1', '.msh2',
        '.mshxml', '.msh1xml', '.msh2xml',
        
        // Archive with executables (potentially dangerous)
        '.application', '.gadget', '.msp', '.mst',
        
        // Other potentially dangerous
        '.hta', '.inf', '.ins', '.isp', '.its', '.job', '.lnk',
        '.mde', '.msc', '.pcd', '.prf', '.reg', '.scf', '.sct',
        '.shb', '.shs', '.url', '.vb', '.wsc', '.ws'
      ],
      
      ALLOWED_EXTENSIONS: [
        // Source code
        '.js', '.ts', '.jsx', '.tsx', '.vue', '.html', '.css', '.scss',
        '.go', '.py', '.java', '.cs', '.cpp', '.c', '.h', '.hpp',
        '.rb', '.php', '.rs', '.kt', '.swift', '.m', '.scala',
        '.clj', '.hs', '.ml', '.fs', '.sh', '.bash', '.zsh',
        
        // Documentation
        '.md', '.txt', '.rst', '.adoc', '.tex',
        
        // Configuration
        '.json', '.yaml', '.yml', '.xml', '.toml', '.ini', '.conf',
        '.env', '.properties', '.cfg',
        
        // Images (safe)
        '.png', '.jpg', '.jpeg', '.gif', '.svg', '.webp', '.ico',
        
        // Other safe files
        '.log', '.csv', '.tsv', '.sql', '.lock', '.gitignore',
        '.gitattributes', '.editorconfig', '.prettierrc'
      ],
      
      WORKSPACE_BOUNDARY_CHECKS: {
        ENFORCE_WORKSPACE_BOUNDARY: true,
        ALLOW_SYMLINKS: false,
        ALLOW_HIDDEN_FILES: false,
        ALLOW_SYSTEM_FILES: false,
        MAX_SYMLINK_DEPTH: 3
      }
    },
    
    errorPatterns: {
      MEMORY_ERROR_PATTERNS: [
        /out of memory/i,
        /maximum call stack/i,
        /heap out of memory/i,
        /memory allocation failed/i,
        /stack overflow/i,
        /heap overflow/i,
        /buffer overflow/i,
        /segmentation fault/i,
        /access violation/i
      ],
      CONTEXT_SIZE_ERROR_PATTERNS: [
        /context too large/i,
        /token limit exceeded/i,
        /request too large/i,
        /payload too large/i,
        /content length exceeded/i
      ]
    },
    
    audit: {
      LOG_FAILED_VALIDATIONS: true,
      LOG_SUSPICIOUS_ACTIVITY: true,
      MAX_AUDIT_LOG_SIZE: 10 * 1024 * 1024, // 10MB
      AUDIT_LOG_RETENTION_DAYS: 30
    }
  },
  
  // File Tree Configuration
  fileTree: {
    limits: {
      MAX_SELECTED_FILES: 500,
      MAX_NODES: 5000,
      MAX_DEPTH: 20
    },
    
    virtualization: {
      ENABLED: true,
      ITEM_HEIGHT: 24,
      BUFFER_SIZE: 10
    },
    
    // Root node configuration
    ROOT_NODE_PATH: ""
  },
  
  // Language Detection
  languageDetection: {
    EXTENSION_MAP: {
      '.js': 'javascript',
      '.ts': 'typescript',
      '.jsx': 'javascript',
      '.tsx': 'typescript',
      '.vue': 'vue',
      '.go': 'go',
      '.py': 'python',
      '.java': 'java',
      '.cs': 'csharp',
      '.cpp': 'cpp',
      '.c': 'c',
      '.h': 'c',
      '.hpp': 'cpp',
      '.rb': 'ruby',
      '.php': 'php',
      '.rs': 'rust',
      '.kt': 'kotlin',
      '.swift': 'swift',
      '.m': 'objectivec',
      '.scala': 'scala',
      '.clj': 'clojure',
      '.hs': 'haskell',
      '.ml': 'ocaml',
      '.fs': 'fsharp'
    },
    DEFAULT_LANGUAGE: 'text',
    FALLBACK_DETECTION: true
  },
  
  // AI and Model Configuration
  ai: {
    defaultSettings: {
      provider: 'openai',
      model: 'gpt-4',
      temperature: 0.7,
      maxTokens: 4096
    },
    
    systemPrompts: {
      DEFAULT: 'You are a helpful AI assistant for code generation.',
      CODE_GENERATION: 'Generate clean, maintainable code following best practices.',
      CODE_REVIEW: 'Review the following code for potential issues and improvements.'
    },
    
    validation: {
      SYNTAX_CHECK: true,
      STRUCTURE_CHECK: true,
      SECURITY_CHECK: true
    }
  },
  
  // Autonomous Mode Configuration (from YAML)
  autonomous: {
    slaPolicies: {
      lite: {
        maxConcurrentTasks: 1,
        maxRetries: 2,
        timeoutPerTask: 900, // 15 minutes
        qualityChecks: ['basic']
      },
      standard: {
        maxConcurrentTasks: 2,
        maxRetries: 3,
        timeoutPerTask: 1800, // 30 minutes
        qualityChecks: ['basic', 'syntax', 'testing']
      },
      strict: {
        maxConcurrentTasks: 1,
        maxRetries: 5,
        timeoutPerTask: 3600, // 60 minutes
        qualityChecks: ['basic', 'syntax', 'testing', 'security', 'performance']
      }
    },
    DEFAULT_SLA: 'standard'
  },
  
  // Reports Configuration
  reports: {
    idGeneration: {
      PREFIX: 'RPT',
      TIMESTAMP_FORMAT: 'YYYYMMDD-HHmmss',
      RANDOM_SUFFIX_LENGTH: 6
    },
    
    export: {
      DEFAULT_FORMAT: 'json',
      SUPPORTED_FORMATS: ['json', 'csv', 'pdf', 'html'],
      MAX_EXPORT_SIZE: 50 * 1024 * 1024 // 50MB
    }
  },
  
  // Project Configuration
  project: {
    extraction: {
      NAME_FROM_PATH: true,
      FALLBACK_NAME: 'Unnamed Project'
    },
    
    recentProjects: {
      MAX_COUNT: 10,
      STORAGE_KEY: 'recent-projects'
    }
  },
  
  // QuickLook Configuration
  quicklook: {
    DEFAULT_LANGUAGE: 'text',
    SYNTAX_HIGHLIGHTING: true,
    MAX_FILE_SIZE: 1024 * 1024, // 1MB
    PREVIEW_LINES: 100
  },
  
  // Events and Notifications
  events: {
    DEBOUNCE_DELAY: 300,
    MAX_EVENT_QUEUE_SIZE: 1000,
    ERROR_RETRY_ATTEMPTS: 3,
    // API timeout configuration
    API_TIMEOUT_MS: 30000 // 30 seconds
  },
  
  // Storage Configuration
  storage: {
    UI_PREFERENCES_PREFIX: 'ui-',
    SESSION_STORAGE_PREFIX: 'session-',
    CACHE_EXPIRY: 24 * 60 * 60 * 1000, // 24 hours
    CLEANUP_INTERVAL: 60 * 60 * 1000    // 1 hour
  },
  
  // Theme Configuration
  theme: {
    STORAGE_KEY: 'shotgun-theme-settings'
  }
} as const;

// Enhanced helper functions with type safety

/**
 * Get primary workspace view component name
 */
export function getPrimaryWorkspaceView(): string {
  return APP_CONFIG.PRIMARY_WORKSPACE_VIEW;
}

/**
 * Get primary UI component
 */
export function getPrimaryComponent(component: keyof typeof APP_CONFIG.ui.components): string {
  return APP_CONFIG.ui.components[component];
}

/**
 * Check if a feature is enabled
 */
export function isFeatureEnabled(feature: keyof typeof APP_CONFIG.ui.features): boolean {
  return APP_CONFIG.ui.features[feature];
}

/**
 * Get layout setting
 */
export function getLayoutSetting(setting: keyof typeof APP_CONFIG.ui.layout): number | string | object {
  return APP_CONFIG.ui.layout[setting];
}

/**
 * Get performance limit
 */
export function getPerformanceLimit(limitType: keyof typeof APP_CONFIG.performance.limits): number {
  return APP_CONFIG.performance.limits[limitType];
}

/**
 * Get memory threshold
 */
export function getMemoryThreshold(thresholdType: keyof typeof APP_CONFIG.performance.memory): number {
  return APP_CONFIG.performance.memory[thresholdType];
}

/**
 * Get risk threshold
 */
export function getRiskThreshold(riskLevel: keyof typeof APP_CONFIG.context.estimation.RISK_THRESHOLDS): number {
  return APP_CONFIG.context.estimation.RISK_THRESHOLDS[riskLevel];
}

/**
 * Get language from file extension
 */
export function getLanguageFromExtension(extension: string): string {
  return APP_CONFIG.languageDetection.EXTENSION_MAP[extension] || APP_CONFIG.languageDetection.DEFAULT_LANGUAGE;
}

/**
 * Get SLA policy configuration
 */
export function getSLAPolicy(policyName: keyof typeof APP_CONFIG.autonomous.slaPolicies) {
  return APP_CONFIG.autonomous.slaPolicies[policyName];
}

/**
 * Get UI component limit
 */
export function getUILimit(component: string, limitType: string): number {
  const limits = (APP_CONFIG.ui as any)[component];
  return (limits as any)?.[limitType] || 0;
}

/**
 * Check if text matches security error patterns
 */
export function isSecurityPatternMatch(text: string, patternType: 'MEMORY_ERROR' | | 'CONTEXT_SIZE_ERROR'): boolean {
  const patterns = APP_CONFIG.security.errorPatterns[
    `${patternType}_PATTERNS` as keyof typeof APP_CONFIG.security.errorPatterns
  ];
  return patterns.some(pattern => pattern.test(text));
}

/**
 * Validate file path against security policies
 */
export function validatePath(path: string): boolean {
  if (path.length > APP_CONFIG.security.validation.MAX_PATH_LENGTH) {
    return false;
  }
  
  
  return !APP_CONFIG.security.validation.FORBIDDEN_PATTERNS.some(pattern => 
    path.includes(pattern)
  );
}

/**
 * Get tokens estimation based on text and language
 */
export function estimateTokens(text: string, language?: string): number {
  const charsPerToken = APP_CONFIG.performance.tokens.DEFAULT_CHARS_PER_TOKEN;
  const baseEstimate = Math.ceil(text.length / charsPerToken);
  
  // Apply multiplier for more accurate estimation
  return Math.ceil(baseEstimate * APP_CONFIG.performance.tokens.ESTIMATION_MULTIPLIER);
}

/**
 * Calculate risk level based on context size and file count
 */
export function calculateRiskLevel(contextSize: number): RiskLevel {
  const thresholds = APP_CONFIG.context.estimation.RISK_THRESHOLDS;
  
  if (contextSize >= thresholds.CRITICAL) return 'critical';
  if (contextSize >= thresholds.HIGH) return 'high';
  if (contextSize >= thresholds.MEDIUM) return 'medium';
  return 'low';
}

/**
 * Check if streaming should be used based on risk level
 */
export function shouldUseStreamingMode(riskLevel: RiskLevel): boolean {
  return riskLevel === APP_CONFIG.performance.streaming.HIGH_RISK_THRESHOLD ||
         riskLevel === APP_CONFIG.performance.streaming.CRITICAL_RISK_THRESHOLD;
}

/**
 * Generate unique report ID
 */
export function generateReportId(): string {
  const config = APP_CONFIG.reports.idGeneration;
  const timestamp = new Date().toISOString().replace(/[^0-9]/g, '').slice(0, 14);
  const randomSuffix = Math.random().toString(36).substring(2, 2 + config.RANDOM_SUFFIX_LENGTH);
  return `${config.PREFIX}-${timestamp}-${randomSuffix}`;
}

/**
 * Extract project name from path
 */
export function extractProjectName(projectPath: string): string {
  if (!APP_CONFIG.project.extraction.NAME_FROM_PATH) {
    return APP_CONFIG.project.extraction.FALLBACK_NAME;
  }
  
  const pathParts = projectPath.split(/[\\/]/);
  const projectName = pathParts[pathParts.length - 1] || pathParts[pathParts.length - 2];
  return projectName || APP_CONFIG.project.extraction.FALLBACK_NAME;
}

/**
 * Get storage key with prefix
 */
export function getStorageKey(key: string, isSession = false): string {
  const prefix = isSession 
    ? APP_CONFIG.storage.SESSION_STORAGE_PREFIX 
    : APP_CONFIG.storage.UI_PREFERENCES_PREFIX;
  return `${prefix}${key}`;
}