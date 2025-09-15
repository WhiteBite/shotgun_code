/**
 * MemoryManagementPolicy Domain Service
 * 
 * Centralizes memory management rules, thresholds, and policies.
 * Provides intelligent memory usage assessment and optimization
 * recommendations following DDD principles.
 */

import type { 
  RiskLevel, 
  MemoryManagementPolicyConfiguration 
} from '@/types/configuration';
import { APP_CONFIG, calculateRiskLevel, isSecurityPatternMatch } from '@/config/app-config';

export interface MemoryUsage {
  readonly current: number;
  readonly peak: number;
  readonly available: number;
  readonly usedPercent: number;
  readonly timestamp: number;
}

export interface SystemMemoryInfo {
  readonly totalMemory: number;
  readonly usedMemory: number;
  readonly availableMemory: number;
  readonly memoryPressure: 'low' | 'medium' | 'high' | 'critical';
}

export interface MemoryOperation {
  readonly type: 'file_load' | 'context_build' | 'ui_render' | 'cache_operation' | 'streaming';
  readonly estimatedSize: number;
  readonly fileCount?: number;
  readonly priority: 'low' | 'medium' | 'high' | 'critical';
  readonly description: string;
}

export interface MemoryAssessment {
  readonly isAllowed: boolean;
  readonly riskLevel: RiskLevel;
  readonly currentUsage: MemoryUsage;
  readonly estimatedImpact: number;
  readonly recommendations: readonly string[];
  readonly alternativeStrategies: readonly string[];
  readonly maxAllowedSize: number;
}

export interface MemoryOptimizationSuggestion {
  readonly type: 'cleanup' | 'streaming' | 'chunking' | 'caching' | 'lazy_loading';
  readonly priority: 'low' | 'medium' | 'high' | 'urgent';
  readonly description: string;
  readonly estimatedSavings: number;
  readonly implementation: string;
}

/**
 * MemoryManagementPolicy domain service
 * 
 * Provides comprehensive memory management including:
 * - Memory limit enforcement
 * - Risk assessment and mitigation
 * - Optimization recommendations
 * - System memory monitoring
 * - Memory leak detection
 */
export class MemoryManagementPolicy {
  private readonly config: MemoryManagementPolicyConfiguration;
  private memoryHistory: MemoryUsage[] = [];
  private readonly maxHistorySize = 100;

  constructor(config?: Partial<MemoryManagementPolicyConfiguration>) {
    this.config = {
      thresholds: APP_CONFIG.performance.memory,
      limits: APP_CONFIG.performance.limits,
      riskAssessment: APP_CONFIG.context.estimation.RISK_THRESHOLDS,
      ...config
    };
  }

  /**
   * Checks if a memory operation is allowed under current conditions
   */
  checkMemoryLimits(operation: MemoryOperation): MemoryAssessment {
    const currentUsage = this.getCurrentMemoryUsage();
    const riskLevel = this.assessOperationRisk(operation, currentUsage);
    const estimatedImpact = this.estimateMemoryImpact(operation);
    
    // Check against absolute limits
    const exceedsAbsoluteLimit = estimatedImpact > this.config.thresholds.MAX_MEMORY_INCREASE;
    
    // Check against file count limits
    const exceedsFileLimit = operation.fileCount && 
      operation.fileCount > this.config.thresholds.MAX_FILE_LIMIT;
    
    // Check against individual file size limits
    const avgFileSize = operation.fileCount && operation.fileCount > 0 
      ? operation.estimatedSize / operation.fileCount 
      : operation.estimatedSize;
    const exceedsFileSizeLimit = avgFileSize > this.config.thresholds.MAX_INDIVIDUAL_FILE_SIZE;
    
    // Check against system memory pressure
    const systemMemory = this.getSystemMemoryInfo();
    const wouldExceedSystemLimit = currentUsage.current + estimatedImpact > 
      systemMemory.totalMemory * this.config.thresholds.MAX_PEAK_USAGE_PERCENT;

    // Determine if operation is allowed
    const isAllowed = !exceedsAbsoluteLimit && 
                     !exceedsFileLimit && 
                     !exceedsFileSizeLimit && 
                     !wouldExceedSystemLimit;

    // Generate recommendations
    const recommendations = this.generateRecommendations(operation, currentUsage, riskLevel);
    const alternativeStrategies = this.generateAlternativeStrategies(operation, riskLevel);
    
    // Calculate maximum allowed size for this operation
    const maxAllowedSize = this.calculateMaxAllowedSize(operation, currentUsage);

    return {
      isAllowed,
      riskLevel,
      currentUsage,
      estimatedImpact,
      recommendations,
      alternativeStrategies,
      maxAllowedSize
    };
  }

  /**
   * Calculates risk level for memory operations
   */
  calculateRiskLevel(operation: MemoryOperation): RiskLevel {
    return this.assessOperationRisk(operation, this.getCurrentMemoryUsage());
  }

  /**
   * Gets maximum file limit based on current memory state
   */
  getMaxFileLimit(): number {
    const currentUsage = this.getCurrentMemoryUsage();
    const memoryPressure = this.assessMemoryPressure(currentUsage);
    
    switch (memoryPressure) {
      case 'critical':
        return Math.max(1, Math.floor(this.config.thresholds.MAX_FILE_LIMIT * 0.2));
      case 'high':
        return Math.max(2, Math.floor(this.config.thresholds.MAX_FILE_LIMIT * 0.5));
      case 'medium':
        return Math.floor(this.config.thresholds.MAX_FILE_LIMIT * 0.8);
      default:
        return this.config.thresholds.MAX_FILE_LIMIT;
    }
  }

  /**
   * Gets maximum file size based on current memory state
   */
  getMaxFileSize(): number {
    const currentUsage = this.getCurrentMemoryUsage();
    const memoryPressure = this.assessMemoryPressure(currentUsage);
    
    switch (memoryPressure) {
      case 'critical':
        return Math.floor(this.config.thresholds.MAX_INDIVIDUAL_FILE_SIZE * 0.1);
      case 'high':
        return Math.floor(this.config.thresholds.MAX_INDIVIDUAL_FILE_SIZE * 0.25);
      case 'medium':
        return Math.floor(this.config.thresholds.MAX_INDIVIDUAL_FILE_SIZE * 0.5);
      default:
        return this.config.thresholds.MAX_INDIVIDUAL_FILE_SIZE;
    }
  }

  /**
   * Gets critical memory threshold
   */
  getCriticalThreshold(): number {
    return this.config.thresholds.CRITICAL_MEMORY_THRESHOLD;
  }

  /**
   * Provides memory optimization suggestions
   */
  getOptimizationSuggestions(): MemoryOptimizationSuggestion[] {
    const currentUsage = this.getCurrentMemoryUsage();
    const suggestions: MemoryOptimizationSuggestion[] = [];
    
    // Check for cleanup opportunities
    if (currentUsage.usedPercent > 0.7) {
      suggestions.push({
        type: 'cleanup',
        priority: 'high',
        description: 'Clear unused caches and temporary data',
        estimatedSavings: currentUsage.current * 0.2,
        implementation: 'Run memory cleanup routine'
      });
    }
    
    // Check for streaming opportunities
    if (this.hasLargeOperationsPending()) {
      suggestions.push({
        type: 'streaming',
        priority: 'medium',
        description: 'Use streaming for large file operations',
        estimatedSavings: this.config.thresholds.MAX_MEMORY_INCREASE * 0.6,
        implementation: 'Enable streaming mode for large contexts'
      });
    }
    
    // Check for chunking opportunities
    if (this.hasHighFileCount()) {
      suggestions.push({
        type: 'chunking',
        priority: 'medium',
        description: 'Process files in smaller chunks',
        estimatedSavings: currentUsage.current * 0.3,
        implementation: 'Implement file chunking strategy'
      });
    }
    
    // Check for lazy loading opportunities
    suggestions.push({
      type: 'lazy_loading',
      priority: 'low',
      description: 'Implement lazy loading for UI components',
      estimatedSavings: currentUsage.current * 0.15,
      implementation: 'Use virtual scrolling and lazy rendering'
    });

    return suggestions.sort((a, b) => this.getPriorityWeight(b.priority) - this.getPriorityWeight(a.priority));
  }

  /**
   * Detects potential memory leaks
   */
  detectMemoryLeaks(): boolean {
    if (this.memoryHistory.length < 10) {
      return false; // Not enough data
    }

    // Check for consistent memory growth
    const recentHistory = this.memoryHistory.slice(-10);
    const growthTrend = this.calculateGrowthTrend(recentHistory);
    
    // Check for memory not being released
    const hasConsistentGrowth = growthTrend > 0.05; // 5% consistent growth
    const exceedsGrowthThreshold = recentHistory[recentHistory.length - 1].current > 
      recentHistory[0].current * 1.5; // 50% increase
    
    return hasConsistentGrowth && exceedsGrowthThreshold;
  }

  /**
   * Records memory usage for tracking
   */
  recordMemoryUsage(usage: MemoryUsage): void {
    this.memoryHistory.push(usage);
    
    // Maintain history size limit
    if (this.memoryHistory.length > this.maxHistorySize) {
      this.memoryHistory = this.memoryHistory.slice(-this.maxHistorySize);
    }

    // Check for immediate memory issues
    if (usage.usedPercent > this.config.thresholds.CRITICAL_MEMORY_THRESHOLD) {
      this.handleCriticalMemoryPressure(usage);
    }
  }

  /**
   * Gets memory usage history for analysis
   */
  getMemoryHistory(): readonly MemoryUsage[] {
    return [...this.memoryHistory];
  }

  /**
   * Evaluates system memory state
   */
  private evaluateSystemMemory(): SystemMemoryInfo {
    return this.getSystemMemoryInfo();
  }

  /**
   * Assesses risk level for a specific operation
   */
  private assessOperationRisk(operation: MemoryOperation, currentUsage: MemoryUsage): RiskLevel {
    const estimatedImpact = this.estimateMemoryImpact(operation);
    const projectedUsage = currentUsage.current + estimatedImpact;
    
    // Use the centralized risk calculation
    return calculateRiskLevel(projectedUsage);
  }

  /**
   * Estimates memory impact of an operation
   */
  private estimateMemoryImpact(operation: MemoryOperation): number {
    let baseImpact = operation.estimatedSize;
    
    // Add overhead based on operation type
    switch (operation.type) {
      case 'file_load':
        baseImpact *= 1.5; // File parsing and DOM overhead
        break;
      case 'context_build':
        baseImpact *= 2.0; // Context processing overhead
        break;
      case 'ui_render':
        baseImpact *= 1.2; // Rendering overhead
        break;
      case 'cache_operation':
        baseImpact *= 1.1; // Minimal overhead
        break;
      case 'streaming':
        baseImpact *= 0.3; // Streaming reduces memory impact
        break;
    }
    
    // Add file count overhead
    if (operation.fileCount) {
      const fileOverhead = operation.fileCount * 1024; // 1KB per file object
      baseImpact += fileOverhead;
    }
    
    return baseImpact;
  }

  /**
   * Gets current memory usage (mock implementation)
   */
  private getCurrentMemoryUsage(): MemoryUsage {
    // In a real implementation, this would use performance.memory or similar APIs
    const mockMemory = {
      current: 50 * 1024 * 1024, // 50MB
      peak: 80 * 1024 * 1024,    // 80MB
      available: 200 * 1024 * 1024, // 200MB
      usedPercent: 0.2,           // 20%
      timestamp: Date.now()
    };
    
    return mockMemory;
  }

  /**
   * Gets system memory information (mock implementation)
   */
  private getSystemMemoryInfo(): SystemMemoryInfo {
    // Mock implementation - would use actual system APIs
    const totalMemory = 8 * 1024 * 1024 * 1024; // 8GB
    const usedMemory = totalMemory * 0.3; // 30% used
    const availableMemory = totalMemory - usedMemory;
    
    let memoryPressure: 'low' | 'medium' | 'high' | 'critical' = 'low';
    const usageRatio = usedMemory / totalMemory;
    
    if (usageRatio > 0.9) memoryPressure = 'critical';
    else if (usageRatio > 0.7) memoryPressure = 'high';
    else if (usageRatio > 0.5) memoryPressure = 'medium';
    
    return {
      totalMemory,
      usedMemory,
      availableMemory,
      memoryPressure
    };
  }

  /**
   * Assesses memory pressure level
   */
  private assessMemoryPressure(usage: MemoryUsage): 'low' | 'medium' | 'high' | 'critical' {
    if (usage.usedPercent >= 0.9) return 'critical';
    if (usage.usedPercent >= 0.7) return 'high';
    if (usage.usedPercent >= 0.5) return 'medium';
    return 'low';
  }

  /**
   * Generates recommendations for memory management
   */
  private generateRecommendations(
    operation: MemoryOperation, 
    currentUsage: MemoryUsage, 
    riskLevel: RiskLevel
  ): string[] {
    const recommendations: string[] = [];
    
    if (riskLevel === 'critical' || riskLevel === 'high') {
      recommendations.push('Consider using streaming mode to reduce memory usage');
      recommendations.push('Break operation into smaller chunks');
    }
    
    if (currentUsage.usedPercent > this.config.thresholds.CRITICAL_MEMORY_THRESHOLD) {
      recommendations.push('Clear browser cache and temporary data');
      recommendations.push('Close unused browser tabs');
    }
    
    if (operation.fileCount && operation.fileCount > this.config.thresholds.MAX_FILE_LIMIT) {
      recommendations.push(`Reduce file count to ${this.getMaxFileLimit()} or fewer`);
    }
    
    if (operation.type === 'context_build') {
      recommendations.push('Use selective file inclusion');
      recommendations.push('Enable comment stripping');
    }
    
    return recommendations;
  }

  /**
   * Generates alternative strategies for memory-intensive operations
   */
  private generateAlternativeStrategies(operation: MemoryOperation, riskLevel: RiskLevel): string[] {
    const strategies: string[] = [];
    
    if (riskLevel === 'critical' || riskLevel === 'high') {
      strategies.push('streaming_mode');
      strategies.push('chunked_processing');
      strategies.push('progressive_loading');
    }
    
    if (operation.type === 'context_build') {
      strategies.push('selective_inclusion');
      strategies.push('compressed_context');
      strategies.push('lazy_evaluation');
    }
    
    if (operation.fileCount && operation.fileCount > 10) {
      strategies.push('virtualized_lists');
      strategies.push('pagination');
      strategies.push('background_processing');
    }
    
    return strategies;
  }

  /**
   * Calculates maximum allowed size for an operation
   */
  private calculateMaxAllowedSize(operation: MemoryOperation, currentUsage: MemoryUsage): number {
    const systemMemory = this.getSystemMemoryInfo();
    const availableMemory = systemMemory.availableMemory;
    const safetyMargin = 0.8; // 80% of available memory
    
    const maxBySystem = availableMemory * safetyMargin;
    const maxByConfig = this.config.thresholds.MAX_MEMORY_INCREASE;
    const maxByFileLimit = operation.fileCount 
      ? this.config.thresholds.MAX_FILE_LIMIT * this.config.thresholds.MAX_INDIVIDUAL_FILE_SIZE
      : this.config.thresholds.MAX_INDIVIDUAL_FILE_SIZE;
    
    return Math.min(maxBySystem, maxByConfig, maxByFileLimit);
  }

  /**
   * Handles critical memory pressure situations
   */
  private handleCriticalMemoryPressure(usage: MemoryUsage): void {
    // In a real implementation, this would trigger cleanup routines
    console.warn('Critical memory pressure detected:', usage);
    
    // Could trigger:
    // - Cache cleanup
    // - Garbage collection hints
    // - Operation deferrals
    // - User notifications
  }

  /**
   * Checks if there are large operations pending
   */
  private hasLargeOperationsPending(): boolean {
    // Mock implementation - would check actual operation queue
    return false;
  }

  /**
   * Checks if there are operations with high file counts
   */
  private hasHighFileCount(): boolean {
    // Mock implementation - would check actual operation queue
    return false;
  }

  /**
   * Gets priority weight for sorting
   */
  private getPriorityWeight(priority: 'low' | 'medium' | 'high' | 'urgent'): number {
    const weights = { urgent: 4, high: 3, medium: 2, low: 1 };
    return weights[priority];
  }

  /**
   * Calculates growth trend from memory history
   */
  private calculateGrowthTrend(history: MemoryUsage[]): number {
    if (history.length < 2) return 0;
    
    const first = history[0].current;
    const last = history[history.length - 1].current;
    
    return (last - first) / first;
  }
}

/**
 * Default memory management policy instance
 */
export const defaultMemoryManagementPolicy = new MemoryManagementPolicy();