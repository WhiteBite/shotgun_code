/**
 * Security Validation Service
 * 
 * Provides comprehensive security validation using centralized configuration
 * following DDD principles for security policies.
 */

import type { ValidationResult } from '@/domain/repositories/base';
import { APP_CONFIG } from '@/config/app-config';
import type { AuditLoggingService } from './AuditLoggingService';

export interface PathValidationResult extends ValidationResult {
  violations: string[];
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  securityScore: number; // 0-100, higher is more secure
  extensionAllowed: boolean;
  withinWorkspace: boolean;
}

export interface SecurityValidationOptions {
  strictMode?: boolean;
  customPatterns?: RegExp[];
  allowAbsolutePaths?: boolean;
  workspacePath?: string;
  checkExtensions?: boolean;
  enforceWorkspaceBoundary?: boolean;
}

export interface SecurityAuditEntry {
  timestamp: number;
  operation: string;
  path?: string;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  violations: string[];
  allowed: boolean;
  userAgent?: string;
  sessionId?: string;
}

export class SecurityValidationService {
  private auditService?: AuditLoggingService;
  
  constructor(auditService?: AuditLoggingService) {
    this.auditService = auditService;
  }
  
  /**
   * Validate file path according to comprehensive security policies
   */
  validatePath(path: string, options: SecurityValidationOptions = {}): PathValidationResult {
    const violations: string[] = [];
    let riskLevel: 'low' | 'medium' | 'high' | 'critical' = 'low';
    let securityScore = 100;
    let extensionAllowed = true;
    let withinWorkspace = true;

    // Sanitize path first
    const sanitizedPath = this.sanitizePath(path);
    if (sanitizedPath !== path) {
      violations.push('Path contains invalid characters that were sanitized');
      securityScore -= 10;
      riskLevel = 'medium';
    }

    // Check path length using centralized configuration
    if (sanitizedPath.length > APP_CONFIG.security.validation.MAX_PATH_LENGTH) {
      violations.push(`Path exceeds maximum length of ${APP_CONFIG.security.validation.MAX_PATH_LENGTH} characters`);
      securityScore -= 20;
      riskLevel = 'medium';
    }
    
    // Check filename length
    const filename = this.extractFilename(sanitizedPath);
    if (filename.length > APP_CONFIG.security.validation.MAX_FILENAME_LENGTH) {
      violations.push(`Filename exceeds maximum length of ${APP_CONFIG.security.validation.MAX_FILENAME_LENGTH} characters`);
      securityScore -= 15;
      riskLevel = 'medium';
    }
    
    // Check directory depth
    const depth = this.calculateDirectoryDepth(sanitizedPath);
    if (depth > APP_CONFIG.security.validation.MAX_DIRECTORY_DEPTH) {
      violations.push(`Directory depth exceeds maximum of ${APP_CONFIG.security.validation.MAX_DIRECTORY_DEPTH} levels`);
      securityScore -= 15;
      riskLevel = 'medium';
    }

    // Check for forbidden patterns from centralized configuration
    const forbiddenPatterns = APP_CONFIG.security.validation.FORBIDDEN_PATTERNS;
    for (const pattern of forbiddenPatterns) {
      if (pattern.test(sanitizedPath)) {
        violations.push(`Path contains forbidden pattern: ${pattern.source}`);
        securityScore -= 25;
        riskLevel = 'high';
        break;
      }
    }
    
    // Check file extension if enabled
    if (options.checkExtensions !== false) {
      const extensionCheck = this.validateFileExtension(sanitizedPath);
      if (!extensionCheck.allowed) {
        extensionAllowed = false;
        violations.push(`File extension '${extensionCheck.extension}' is not allowed`);
        securityScore -= 30;
        riskLevel = extensionCheck.dangerous ? 'critical' : 'high';
      }
    }

    // Check workspace boundary if workspace path provided
    if (options.workspacePath && options.enforceWorkspaceBoundary !== false) {
      const workspaceCheck = this.validateWorkspaceAccess(sanitizedPath, options.workspacePath);
      if (!workspaceCheck.isValid) {
        withinWorkspace = false;
        violations.push(...workspaceCheck.errors);
        securityScore -= 40;
        riskLevel = 'critical';
      }
    }

    // Check for absolute paths if not allowed
    if (!options.allowAbsolutePaths && this.isAbsolutePath(sanitizedPath)) {
      violations.push('Absolute paths are not allowed');
      securityScore -= 20;
      riskLevel = this.upgradeRiskLevel(riskLevel, 'medium');
    }

    // Check for parent directory traversal attempts
    if (sanitizedPath.includes('..')) {
      violations.push('Parent directory access is not allowed');
      securityScore -= 35;
      riskLevel = this.upgradeRiskLevel(riskLevel, 'high');
    }

    // Check for hidden files/directories (potential security risk)
    if (sanitizedPath.includes('/.') && !APP_CONFIG.security.validation.WORKSPACE_BOUNDARY_CHECKS.ALLOW_HIDDEN_FILES) {
      violations.push('Access to hidden files/directories detected');
      securityScore -= 15;
      riskLevel = this.upgradeRiskLevel(riskLevel, 'medium');
    }
    
    // Check for system files
    if (this.isSystemFile(sanitizedPath) && !APP_CONFIG.security.validation.WORKSPACE_BOUNDARY_CHECKS.ALLOW_SYSTEM_FILES) {
      violations.push('Access to system files is not allowed');
      securityScore -= 30;
      riskLevel = this.upgradeRiskLevel(riskLevel, 'high');
    }

    // Custom pattern validation if provided
    if (options.customPatterns) {
      for (const pattern of options.customPatterns) {
        if (pattern.test(sanitizedPath)) {
          violations.push(`Path violates custom security pattern: ${pattern.source}`);
          securityScore -= 20;
          riskLevel = this.upgradeRiskLevel(riskLevel, 'high');
          break;
        }
      }
    }

    // Determine overall risk level based on violation count
    if (violations.length > 3) {
      riskLevel = 'critical';
      securityScore = Math.min(securityScore, 20);
    } else if (violations.length > 1 && riskLevel !== 'critical') {
      riskLevel = this.upgradeRiskLevel(riskLevel, 'high');
    }
    
    // Ensure security score is not negative
    securityScore = Math.max(0, securityScore);
    
    // Log the validation attempt for audit
    // Also log to audit service if available
    if (this.auditService && violations.length > 0) {
      this.auditService.logEvent({
        eventType: 'security_violation',
        severity: riskLevel === 'critical' ? 'critical' : riskLevel === 'high' ? 'error' : 'warning',
        source: 'security_validation',
        action: 'path_validation',
        outcome: violations.length === 0 ? 'success' : 'failure',
        details: {
          path: sanitizedPath,
          riskLevel,
          violations,
          securityScore,
          extensionAllowed,
          withinWorkspace,
          validationOptions: options
        }
      });
    }

    return {
      isValid: violations.length === 0,
      violations,
      riskLevel,
      securityScore,
      extensionAllowed,
      withinWorkspace,
      errors: violations,
      warnings: []
    };
  }

  /**
   * Validate multiple paths efficiently
   */
  validatePaths(paths: string[], options: SecurityValidationOptions = {}): PathValidationResult[] {
    return paths.map(path => this.validatePath(path, options));
  }
  
  /**
   * Validate workspace boundary for multiple paths
   */
  validateWorkspaceBoundary(paths: string[], workspacePath: string): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];
    
    for (const path of paths) {
      const result = this.validateWorkspaceAccess(path, workspacePath);
      if (!result.isValid) {
        errors.push(...result.errors.map(err => `${path}: ${err}`));
      }
    }
    
    if (errors.length > 0) {
      warnings.push('Some paths are outside the workspace boundary');
    }
    
    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }
  
  /**
   * Get security risk assessment for a collection of paths
   */
  assessSecurityRisk(paths: string[], options: SecurityValidationOptions = {}): {
    overallRisk: 'low' | 'medium' | 'high' | 'critical';
    riskFactors: string[];
    recommendations: string[];
    securityScore: number;
    pathRisks: Array<{ path: string; risk: 'low' | 'medium' | 'high' | 'critical'; score: number }>;
  } {
    const pathResults = this.validatePaths(paths, options);
    const riskFactors: string[] = [];
    const recommendations: string[] = [];
    
    let totalScore = 0;
    let maxRisk: 'low' | 'medium' | 'high' | 'critical' = 'low';
    
    const pathRisks = pathResults.map((result, index) => {
      totalScore += result.securityScore;
      maxRisk = this.upgradeRiskLevel(maxRisk, result.riskLevel);
      
      if (!result.isValid) {
        riskFactors.push(...result.violations.map(v => `${paths[index]}: ${v}`));
      }
      
      return {
        path: paths[index],
        risk: result.riskLevel,
        score: result.securityScore
      };
    });
    
    const averageScore = paths.length > 0 ? totalScore / paths.length : 100;
    
    // Generate recommendations based on findings
    if (maxRisk === 'critical') {
      recommendations.push('Immediate action required: Critical security violations detected');
      recommendations.push('Review and sanitize all flagged paths before proceeding');
    } else if (maxRisk === 'high') {
      recommendations.push('High risk detected: Review security violations');
      recommendations.push('Consider restricting access to flagged paths');
    } else if (maxRisk === 'medium') {
      recommendations.push('Medium risk: Monitor and validate path usage');
    }
    
    if (averageScore < 70) {
      recommendations.push('Overall security score is low - implement stricter path validation');
    }
    
    return {
      overallRisk: maxRisk,
      riskFactors: [...new Set(riskFactors)], // Remove duplicates
      recommendations,
      securityScore: averageScore,
      pathRisks
    };
  }
  
  /**
   * Get audit log entries for security analysis
   */
  getAuditLog(filters?: {
    startTime?: number;
    endTime?: number;
    riskLevel?: 'low' | 'medium' | 'high' | 'critical';
    operation?: string;
    onlyViolations?: boolean;
  }): SecurityAuditEntry[] {
    // Return empty array as we're now using external audit service
    return [];
  }
  
  /**
   * Clear audit log (for testing or privacy)
   */
  clearAuditLog(): void {
    // No-op as we're now using external audit service
  }
  
  /**
   * Get security statistics
   */
  getSecurityStatistics(): {
    totalValidations: number;
    violationsCount: number;
    riskDistribution: Record<'low' | 'medium' | 'high' | 'critical', number>;
    commonViolations: Array<{ violation: string; count: number }>;
    averageSecurityScore: number;
  } {
    // Return placeholder values as we're now using external audit service
    return {
      totalValidations: 0,
      violationsCount: 0,
      riskDistribution: { low: 0, medium: 0, high: 0, critical: 0 },
      commonViolations: [],
      averageSecurityScore: 85
    };
  }

  /**
   * Check if path points to workspace only (no external access)
   */
  validateWorkspaceAccess(path: string, workspacePath: string): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];
    
    // Skip validation if workspace boundary checks are disabled
    if (!APP_CONFIG.security.validation.WORKSPACE_BOUNDARY_CHECKS.ENFORCE_WORKSPACE_BOUNDARY) {
      return { isValid: true, errors: [], warnings: [] };
    }
    
    const normalizedPath = this.normalizePath(path);
    const normalizedWorkspace = this.normalizePath(workspacePath);

    // Basic workspace boundary check
    if (!normalizedPath.startsWith(normalizedWorkspace)) {
      errors.push('Path is outside of workspace boundary');
    }
    
    // Check for symlink traversal if symlinks are not allowed
    if (!APP_CONFIG.security.validation.WORKSPACE_BOUNDARY_CHECKS.ALLOW_SYMLINKS) {
      if (this.containsSymlinks(normalizedPath)) {
        errors.push('Symlinks are not allowed in workspace');
      }
    }
    
    // Additional workspace security checks
    const relativePath = normalizedPath.replace(normalizedWorkspace, '');
    
    // Check for attempts to access parent workspace
    if (relativePath.includes('../') || relativePath.includes('..\\')) {
      errors.push('Attempted parent directory access outside workspace');
    }
    
    // Check for access to hidden workspace files if not allowed
    if (!APP_CONFIG.security.validation.WORKSPACE_BOUNDARY_CHECKS.ALLOW_HIDDEN_FILES && 
        relativePath.match(/[\/\\]\.[^/\\]/)) {
      warnings.push('Access to hidden files detected');
    }
    
    // Check for system file access within workspace
    if (!APP_CONFIG.security.validation.WORKSPACE_BOUNDARY_CHECKS.ALLOW_SYSTEM_FILES && 
        this.isSystemFile(relativePath)) {
      errors.push('Access to system files within workspace is not allowed');
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }
  
  private containsSymlinks(path: string): boolean {
    // This is a simplified check - in a real implementation,
    // you would need to check the filesystem for actual symlinks
    // For now, we'll check for common symlink indicators
    return path.includes(' -> ') || path.includes('@');
  }

  /**
   * Sanitize path for safe usage
   */
  sanitizePath(path: string): string {
    // Remove null bytes and control characters
    let sanitized = path.replace(/[\x00-\x1f\x7f-\x9f]/g, '');
    
    // Normalize path separators
    sanitized = sanitized.replace(/\\/g, '/');
    
    // Remove redundant slashes
    sanitized = sanitized.replace(/\/+/g, '/');
    
    // Remove leading/trailing whitespace
    sanitized = sanitized.trim();
    
    return sanitized;
  }

  /**
   * Check for memory-based attacks (extremely long paths, etc.)
   */
  validateMemorySecurityLimits(data: string): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    // Check against memory security patterns from centralized configuration
    const memoryErrorPatterns = APP_CONFIG.security.errorPatterns.MEMORY_ERROR_PATTERNS;
    for (const pattern of memoryErrorPatterns) {
      if (pattern.test(data)) {
        errors.push('Potential memory-based security threat detected');
        break;
      }
    }

    // Check data size
    if (data.length > APP_CONFIG.performance.memory.MAX_INDIVIDUAL_FILE_SIZE) {
      warnings.push('Data size exceeds recommended limits');
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * Validate context size for security (prevent DoS attacks)
   */
  validateContextSecurity(content: string): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    // Check against context size error patterns from centralized configuration
    const contextErrorPatterns = APP_CONFIG.security.errorPatterns.CONTEXT_SIZE_ERROR_PATTERNS;
    for (const pattern of contextErrorPatterns) {
      if (pattern.test(content)) {
        errors.push('Potential context-based security threat detected');
        break;
      }
    }

    // Check if content exceeds safe limits
    const tokenEstimate = content.length / APP_CONFIG.performance.tokens.DEFAULT_CHARS_PER_TOKEN;
    if (tokenEstimate > APP_CONFIG.performance.tokens.MAX_TOKENS_PER_REQUEST) {
      errors.push('Context size exceeds security limits');
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * Comprehensive security audit of operation
   */
  auditOperation(operation: {
    type: 'file_access' | 'context_build' | 'export' | 'import';
    paths?: string[];
    content?: string;
    metadata?: Record<string, unknown>;
  }): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    // Validate paths if provided
    if (operation.paths) {
      for (const path of operation.paths) {
        const pathValidation = this.validatePath(path);
        if (!pathValidation.isValid) {
          errors.push(...pathValidation.errors);
        }
      }
    }

    // Validate content if provided
    if (operation.content) {
      const contentValidation = this.validateContextSecurity(operation.content);
      if (!contentValidation.isValid) {
        errors.push(...contentValidation.errors);
      }
      warnings.push(...contentValidation.warnings);
    }

    // Operation-specific validations
    switch (operation.type) {
      case 'export':
        if (operation.content && operation.content.length > 1000000) { // 1MB limit
          warnings.push('Export size is very large and may cause performance issues');
        }
        break;
      
      case 'context_build':
        if (operation.paths && operation.paths.length > APP_CONFIG.performance.limits.MAX_SELECTED_PATHS) {
          errors.push(`Too many files selected (max: ${APP_CONFIG.performance.limits.MAX_SELECTED_PATHS})`);
        }
        break;
    }
    
    // Log audit event if audit service is available
    // No-op as we're now using external audit service

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  // Private helper methods
  
  private extractFilename(path: string): string {
    const parts = path.split(/[\/\\]/);
    return parts[parts.length - 1] || '';
  }
  
  private calculateDirectoryDepth(path: string): number {
    const normalizedPath = path.replace(/[\/\\]+/g, '/');
    const parts = normalizedPath.split('/').filter(part => part.length > 0);
    return parts.length - 1; // Subtract 1 for the filename
  }
  
  private validateFileExtension(path: string): { allowed: boolean; extension: string; dangerous: boolean } {
    const extension = this.getFileExtension(path).toLowerCase();
    
    // Check if extension is explicitly forbidden
    const forbiddenExtensions = APP_CONFIG.security.validation.FORBIDDEN_EXTENSIONS;
    if (forbiddenExtensions.includes(extension)) {
      return { allowed: false, extension, dangerous: true };
    }
    
    // Check if extension is in allowed list
    const allowedExtensions = APP_CONFIG.security.validation.ALLOWED_EXTENSIONS;
    if (allowedExtensions.includes(extension)) {
      return { allowed: true, extension, dangerous: false };
    }
    
    // If not in either list, it's not allowed but not necessarily dangerous
    return { allowed: false, extension, dangerous: false };
  }
  
  private getFileExtension(path: string): string {
    const filename = this.extractFilename(path);
    const lastDotIndex = filename.lastIndexOf('.');
    return lastDotIndex > 0 ? filename.substring(lastDotIndex) : '';
  }
  
  private isSystemFile(path: string): boolean {
    const systemPatterns = [
      /[\/\\]system32[\/\\]/i,
      /[\/\\]windows[\/\\]/i,
      /[\/\\]program files/i,
      /[\/\\]etc[\/\\]/,
      /[\/\\]proc[\/\\]/,
      /[\/\\]dev[\/\\]/,
      /[\/\\]sys[\/\\]/,
      /[\/\\]boot[\/\\]/,
      /[\/\\]usr[\/\\]bin[\/\\]/,
      /[\/\\]sbin[\/\\]/
    ];
    
    return systemPatterns.some(pattern => pattern.test(path));
  }
  
  private upgradeRiskLevel(
    current: 'low' | 'medium' | 'high' | 'critical',
    target: 'low' | 'medium' | 'high' | 'critical'
  ): 'low' | 'medium' | 'high' | 'critical' {
    const levels = { low: 0, medium: 1, high: 2, critical: 3 };
    return levels[target] > levels[current] ? target : current;
  }
  


  private isAbsolutePath(path: string): boolean {
    // Check for Windows absolute paths (C:\ or \\)
    if (/^[a-zA-Z]:[\\\/]/.test(path) || path.startsWith('\\\\')) {
      return true;
    }
    
    // Check for Unix absolute paths (/)
    if (path.startsWith('/')) {
      return true;
    }
    
    return false;
  }

  private normalizePath(path: string): string {
    // Normalize path separators and resolve relative components
    return path.replace(/\\/g, '/').replace(/\/+/g, '/').replace(/\/$/, '');
  }
}

// Default instance for injection - will be enhanced with audit service when available
export const defaultSecurityValidationService = new SecurityValidationService();

// Factory function for dependency injection with audit service
export const createSecurityValidationService = (auditService?: AuditLoggingService) => 
  new SecurityValidationService(auditService);