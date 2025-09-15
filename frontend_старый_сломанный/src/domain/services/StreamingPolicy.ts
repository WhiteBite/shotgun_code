/**
 * StreamingPolicy Domain Service
 * 
 * Encapsulates the business logic for determining when to use streaming mode
 * based on context size, risk assessment, and memory management policies.
 * 
 * This follows DDD principles by keeping business logic in the domain layer.
 */

import type { RiskLevel, StreamingPolicyConfiguration } from '@/types/configuration';
import { APP_CONFIG, calculateRiskLevel } from '@/config/app-config';

export interface MemoryOperation {
  readonly type: 'context_build' | 'file_read' | 'chunk_process';
  readonly estimatedSize: number;
  readonly fileCount: number;
  readonly projectPath?: string;
}

export interface MemoryCheckResult {
  readonly isAllowed: boolean;
  readonly riskLevel: RiskLevel;
  readonly recommendStreaming: boolean;
  readonly estimatedMemoryImpact: number;
  readonly reasons: readonly string[];
}

/**
 * StreamingPolicy domain service
 * 
 * Determines whether streaming mode should be used based on:
 * - Context size and file count
 * - Memory impact assessment
 * - Risk level calculation
 * - Performance thresholds
 */
export class StreamingPolicy {
  private readonly config: StreamingPolicyConfiguration;

  constructor(config?: Partial<StreamingPolicyConfiguration>) {
    this.config = {
      highRiskThreshold: APP_CONFIG.performance.streaming.HIGH_RISK_THRESHOLD,
      criticalRiskThreshold: APP_CONFIG.performance.streaming.CRITICAL_RISK_THRESHOLD,
      memoryThresholds: APP_CONFIG.performance.memory,
      riskThresholds: APP_CONFIG.context.estimation.RISK_THRESHOLDS,
      ...config
    };
  }

  /**
   * Determines if streaming mode should be used for the given operation
   */
  shouldUseStreamingMode(operation: MemoryOperation): boolean {
    const checkResult = this.checkMemoryLimits(operation);
    return checkResult.recommendStreaming;
  }

  /**
   * Calculates the risk level for a given operation
   */
  calculateRiskLevel(operation: MemoryOperation): RiskLevel {
    const { estimatedSize, fileCount } = operation;
    
    // Check against size thresholds
    const sizeRisk = calculateRiskLevel(estimatedSize);
    
    // Check against file count limits
    const fileCountRisk = this.assessFileCountRisk(fileCount);
    
    // Return the higher risk level
    return this.getHigherRiskLevel(sizeRisk, fileCountRisk);
  }

  /**
   * Performs comprehensive memory limit checking
   */
  checkMemoryLimits(operation: MemoryOperation): MemoryCheckResult {
    const { estimatedSize, fileCount } = operation;
    const reasons: string[] = [];
    
    // Calculate risk level
    const riskLevel = this.calculateRiskLevel(operation);
    
    // Check file count limits
    const exceedsFileLimit = fileCount > this.config.memoryThresholds.MAX_FILE_LIMIT;
    if (exceedsFileLimit) {
      reasons.push(`File count (${fileCount}) exceeds limit (${this.config.memoryThresholds.MAX_FILE_LIMIT})`);
    }
    
    // Check individual file size limits (estimated average)
    const avgFileSize = fileCount > 0 ? estimatedSize / fileCount : estimatedSize;
    const exceedsFileSizeLimit = avgFileSize > this.config.memoryThresholds.MAX_INDIVIDUAL_FILE_SIZE;
    if (exceedsFileSizeLimit) {
      reasons.push(`Average file size exceeds limit`);
    }
    
    // Calculate memory impact
    const estimatedMemoryImpact = this.estimateMemoryImpact(estimatedSize, fileCount);
    const exceedsMemoryIncrease = estimatedMemoryImpact > this.config.memoryThresholds.MAX_MEMORY_INCREASE;
    if (exceedsMemoryIncrease) {
      reasons.push(`Estimated memory impact (${Math.round(estimatedMemoryImpact / 1024 / 1024)}MB) exceeds threshold`);
    }
    
    // Determine if streaming is recommended
    const shouldStream = this.shouldStreamBasedOnRisk(riskLevel) || 
                        exceedsFileLimit || 
                        exceedsFileSizeLimit || 
                        exceedsMemoryIncrease;
    
    if (shouldStream) {
      reasons.push(`Risk level: ${riskLevel}`);
    }
    
    // Operation is allowed but may require streaming
    const isAllowed = !exceedsMemoryIncrease || shouldStream;
    
    return {
      isAllowed,
      riskLevel,
      recommendStreaming: shouldStream,
      estimatedMemoryImpact,
      reasons
    };
  }

  /**
   * Gets the streaming threshold based on configuration
   */
  getStreamingThreshold(): number {
    return this.config.riskThresholds.HIGH;
  }

  /**
   * Evaluates memory impact for the operation
   */
  private evaluateMemoryImpact(operation: MemoryOperation): boolean {
    const impact = this.estimateMemoryImpact(operation.estimatedSize, operation.fileCount);
    return impact > this.config.memoryThresholds.MAX_MEMORY_INCREASE;
  }

  /**
   * Estimates memory impact based on size and file count
   */
  private estimateMemoryImpact(size: number, fileCount: number): number {
    // Base memory for content
    let memoryEstimate = size;
    
    // Add overhead for object creation and DOM manipulation
    const objectOverhead = fileCount * 1024; // 1KB per file object
    const domOverhead = size * 0.1; // 10% DOM overhead
    
    memoryEstimate += objectOverhead + domOverhead;
    
    return memoryEstimate;
  }

  /**
   * Assesses risk based on file count
   */
  private assessFileCountRisk(fileCount: number): RiskLevel {
    const maxFiles = this.config.memoryThresholds.MAX_FILE_LIMIT;
    
    if (fileCount >= maxFiles * 2) return 'critical';
    if (fileCount >= maxFiles * 1.5) return 'high';
    if (fileCount >= maxFiles) return 'medium';
    return 'low';
  }

  /**
   * Determines if streaming should be used based on risk level
   */
  private shouldStreamBasedOnRisk(riskLevel: RiskLevel): boolean {
    return riskLevel === this.config.highRiskThreshold || 
           riskLevel === this.config.criticalRiskThreshold;
  }

  /**
   * Returns the higher of two risk levels
   */
  private getHigherRiskLevel(risk1: RiskLevel, risk2: RiskLevel): RiskLevel {
    const riskOrder: Record<RiskLevel, number> = {
      'low': 1,
      'medium': 2,
      'high': 3,
      'critical': 4
    };
    
    return riskOrder[risk1] >= riskOrder[risk2] ? risk1 : risk2;
  }
}

/**
 * Default streaming policy instance using application configuration
 */
export const defaultStreamingPolicy = new StreamingPolicy();