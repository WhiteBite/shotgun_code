/**
 * Performance Monitoring Service
 * 
 * Domain service that centralizes performance monitoring, memory management,
 * and error pattern detection following DDD principles and observability patterns.
 */

import { APP_CONFIG } from '@/config/app-config';
import type { ObservabilityService } from './ObservabilityService';

export interface PerformanceMetrics {
  timestamp: number;
  memoryUsage: {
    used: number;
    total: number;
    percentage: number;
    jsHeapSize?: number;
    jsHeapSizeLimit?: number;
  };
  responseTime?: number;
  operationType: string;
  success: boolean;
  error?: string;
  context?: Record<string, unknown>;
}

export interface PerformanceAlert {
  id: string;
  level: 'warning' | 'critical' | 'error';
  type: 'memory' | 'response_time' | 'error_rate' | 'context_size';
  message: string;
  timestamp: number;
  metrics: PerformanceMetrics;
  suggestions?: string[];
}

export interface CircuitBreakerConfig {
  failureThreshold: number;
  resetTimeout: number;
  monitoringPeriod: number;
}

export interface CircuitBreakerState {
  state: 'CLOSED' | 'OPEN' | 'HALF_OPEN';
  failures: number;
  lastFailureTime: number;
  nextAttemptTime: number;
}

export interface ObservabilityContext {
  traceId: string;
  spanId: string;
  operation: string;
  startTime: number;
  tags: Record<string, string | number | boolean>;
}

export class PerformanceMonitoringService {
  private metrics: PerformanceMetrics[] = [];
  private alerts: PerformanceAlert[] = [];
  private circuitBreakers: Map<string, CircuitBreakerState> = new Map();
  private activeTraces: Map<string, ObservabilityContext> = new Map();
  private monitoringInterval: number | null = null;
  private alertListeners: ((alert: PerformanceAlert) => void)[] = [];
  private observabilityService?: ObservabilityService;

  constructor(observabilityService?: ObservabilityService) {
    this.observabilityService = observabilityService;
    this.startMonitoring();
  }

  /**
   * Start performance monitoring
   */
  startMonitoring(): void {
    if (this.monitoringInterval) {
      return;
    }

    this.monitoringInterval = window.setInterval(
      () => this.collectMetrics(),
      APP_CONFIG.performance.memory.POLLING_INTERVAL_MS
    );

    console.log('Performance monitoring started');
  }

  /**
   * Stop performance monitoring
   */
  stopMonitoring(): void {
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
      this.monitoringInterval = null;
      console.log('Performance monitoring stopped');
    }
  }

  /**
   * Record a performance metric
   */
  recordMetric(operationType: string, startTime: number, success: boolean, error?: string, context?: Record<string, unknown>): PerformanceMetrics {
    const endTime = performance.now();
    const responseTime = endTime - startTime;
    const memoryUsage = this.getMemoryUsage();

    const metric: PerformanceMetrics = {
      timestamp: Date.now(),
      memoryUsage,
      responseTime,
      operationType,
      success,
      error,
      context
    };

    this.metrics.push(metric);
    this.trimMetrics();
    this.analyzePerformance(metric);
    
    // Also record in observability service if available
    if (this.observabilityService) {
      this.observabilityService.histogram(`performance.${operationType}.duration`, responseTime, {
        success: success.toString(),
        operation: operationType
      });
      
      if (!success) {
        this.observabilityService.counter(`performance.${operationType}.errors`, 1, {
          operation: operationType
        });
      }
      
      this.observabilityService.gauge('performance.memory.usage', memoryUsage.used, { unit: 'MB' });
    }

    return metric;
  }

  /**
   * Get current memory usage
   */
  getMemoryUsage(): PerformanceMetrics['memoryUsage'] {
    const memoryInfo = this.getBrowserMemoryInfo();
    
    if (memoryInfo) {
      const used = Math.round(memoryInfo.usedJSHeapSize / (1024 * 1024));
      const total = Math.round(memoryInfo.jsHeapSizeLimit / (1024 * 1024));
      const percentage = Math.round((memoryInfo.usedJSHeapSize / memoryInfo.jsHeapSizeLimit) * 100);

      return {
        used,
        total,
        percentage,
        jsHeapSize: memoryInfo.usedJSHeapSize,
        jsHeapSizeLimit: memoryInfo.jsHeapSizeLimit
      };
    }

    // Fallback for browsers without memory API
    return {
      used: 0,
      total: 0,
      percentage: 0
    };
  }

  /**
   * Check if error matches known patterns
   */
  isKnownErrorPattern(error: string): { isKnown: boolean; type?: string; severity?: string } {
    const memoryPatterns = APP_CONFIG.security.errorPatterns.MEMORY_ERROR_PATTERNS;
    const contextPatterns = APP_CONFIG.security.errorPatterns.CONTEXT_SIZE_ERROR_PATTERNS;

    for (const pattern of memoryPatterns) {
      if (pattern.test(error)) {
        return { isKnown: true, type: 'memory', severity: 'critical' };
      }
    }

    for (const pattern of contextPatterns) {
      if (pattern.test(error)) {
        return { isKnown: true, type: 'context_size', severity: 'warning' };
      }
    }

    return { isKnown: false };
  }

  /**
   * Circuit breaker pattern implementation
   */
  async executeWithCircuitBreaker<T>(
    key: string,
    operation: () => Promise<T>,
    config: Partial<CircuitBreakerConfig> = {}
  ): Promise<T> {
    const fullConfig: CircuitBreakerConfig = {
      failureThreshold: APP_CONFIG.performance.circuitBreaker.DEFAULT_FAILURE_THRESHOLD,
      resetTimeout: APP_CONFIG.performance.circuitBreaker.DEFAULT_RESET_TIMEOUT_MS,
      monitoringPeriod: APP_CONFIG.performance.circuitBreaker.DEFAULT_MONITORING_PERIOD_MS,
      ...config
    };

    const state = this.getCircuitBreakerState(key, fullConfig);

    // Check if circuit is open
    if (state.state === 'OPEN') {
      if (Date.now() >= state.nextAttemptTime) {
        state.state = 'HALF_OPEN';
        console.log(`Circuit breaker ${key} entering HALF_OPEN state`);
      } else {
        throw new Error(`Circuit breaker ${key} is OPEN. Operation not allowed.`);
      }
    }

    const startTime = performance.now();
    let success = false;
    let error: string | undefined;

    try {
      const result = await operation();
      success = true;

      // Reset failures on success
      if (state.state === 'HALF_OPEN') {
        state.state = 'CLOSED';
        state.failures = 0;
        console.log(`Circuit breaker ${key} reset to CLOSED state`);
      }

      return result;
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      state.failures++;
      state.lastFailureTime = Date.now();

      // Open circuit if failure threshold exceeded
      if (state.failures >= fullConfig.failureThreshold) {
        state.state = 'OPEN';
        state.nextAttemptTime = Date.now() + fullConfig.resetTimeout;
        console.warn(`Circuit breaker ${key} opened due to ${state.failures} failures`);
      }

      throw err;
    } finally {
      // Record the metric
      this.recordMetric(`circuit_breaker_${key}`, startTime, success, error, { circuitState: state.state });
    }
  }

  /**
   * Start a distributed trace
   */
  startTrace(operation: string, tags: Record<string, string | number | boolean> = {}): string {
    const traceId = this.generateTraceId();
    const spanId = this.generateSpanId();

    const context: ObservabilityContext = {
      traceId,
      spanId,
      operation,
      startTime: performance.now(),
      tags: { ...tags, service: 'shotgun-code-frontend' }
    };

    this.activeTraces.set(traceId, context);
    console.debug(`Started trace ${traceId} for operation: ${operation}`);

    return traceId;
  }

  /**
   * End a distributed trace
   */
  endTrace(traceId: string, success: boolean, error?: string, additionalTags: Record<string, string | number | boolean> = {}): void {
    const context = this.activeTraces.get(traceId);
    if (!context) {
      console.warn(`Trace ${traceId} not found`);
      return;
    }

    const duration = performance.now() - context.startTime;
    const allTags = { ...context.tags, ...additionalTags, success, duration };

    // Record as structured log
    const logEntry = {
      timestamp: Date.now(),
      traceId,
      spanId: context.spanId,
      operation: context.operation,
      duration,
      success,
      error,
      tags: allTags
    };

    console.log('Trace completed:', logEntry);
    this.activeTraces.delete(traceId);

    // Also record as performance metric
    this.recordMetric(context.operation, context.startTime, success, error, allTags);
  }

  /**
   * Add an alert listener
   */
  onAlert(listener: (alert: PerformanceAlert) => void): () => void {
    this.alertListeners.push(listener);
    
    // Return unsubscribe function
    return () => {
      const index = this.alertListeners.indexOf(listener);
      if (index > -1) {
        this.alertListeners.splice(index, 1);
      }
    };
  }

  /**
   * Get performance statistics
   */
  getPerformanceStats(): {
    totalMetrics: number;
    averageResponseTime: number;
    successRate: number;
    memoryUsage: PerformanceMetrics['memoryUsage'];
    errorRate: number;
    recentAlerts: PerformanceAlert[];
  } {
    const totalMetrics = this.metrics.length;
    const recentMetrics = this.metrics.slice(-APP_CONFIG.performance.monitoring.METRICS_RETENTION_COUNT);

    const averageResponseTime = recentMetrics
      .filter(m => m.responseTime !== undefined)
      .reduce((sum, m) => sum + (m.responseTime || 0), 0) / Math.max(1, recentMetrics.length);

    const successfulMetrics = recentMetrics.filter(m => m.success);
    const successRate = successfulMetrics.length / Math.max(1, recentMetrics.length);
    const errorRate = 1 - successRate;

    const memoryUsage = this.getMemoryUsage();
    const recentAlerts = this.alerts.slice(-APP_CONFIG.performance.monitoring.ALERTS_RETENTION_COUNT);

    return {
      totalMetrics,
      averageResponseTime,
      successRate,
      memoryUsage,
      errorRate,
      recentAlerts
    };
  }

  /**
   * Force cleanup for memory optimization
   */
  forceCleanup(): { metricsRemoved: number; alertsRemoved: number; tracesCleared: number } {
    const initialMetricsCount = this.metrics.length;
    const initialAlertsCount = this.alerts.length;
    const initialTracesCount = this.activeTraces.size;

    // Keep only recent metrics using centralized configuration
    this.metrics = this.metrics.slice(-APP_CONFIG.performance.monitoring.METRICS_RETENTION_COUNT);
    
    // Keep only recent alerts using centralized configuration
    this.alerts = this.alerts.slice(-APP_CONFIG.performance.monitoring.ALERTS_RETENTION_COUNT);

    // Clear stale traces (older than configured threshold)
    const now = performance.now();
    const staleThreshold = APP_CONFIG.performance.monitoring.TRACE_CLEANUP_THRESHOLD_MS;
    
    for (const [traceId, context] of this.activeTraces) {
      if (now - context.startTime > staleThreshold) {
        this.activeTraces.delete(traceId);
      }
    }

    const result = {
      metricsRemoved: initialMetricsCount - this.metrics.length,
      alertsRemoved: initialAlertsCount - this.alerts.length,
      tracesCleared: initialTracesCount - this.activeTraces.size
    };

    console.log('Performance monitoring cleanup completed:', result);
    return result;
  }

  // Private methods

  private collectMetrics(): void {
    const memoryUsage = this.getMemoryUsage();
    
    const metric: PerformanceMetrics = {
      timestamp: Date.now(),
      memoryUsage,
      operationType: 'monitoring_cycle',
      success: true
    };

    this.metrics.push(metric);
    this.trimMetrics();
    this.analyzePerformance(metric);
  }

  private analyzePerformance(metric: PerformanceMetrics): void {
    // Memory analysis
    if (metric.memoryUsage.percentage >= APP_CONFIG.performance.memory.CRITICAL_MEMORY_THRESHOLD * 100) {
      this.createAlert('critical', 'memory', 
        `Critical memory usage: ${metric.memoryUsage.percentage}%`, 
        metric,
        ['Consider closing unnecessary tabs', 'Clear browser cache', 'Restart application']
      );
    } else if (metric.memoryUsage.used >= APP_CONFIG.performance.memory.WARNING_THRESHOLD_MB) {
      this.createAlert('warning', 'memory',
        `High memory usage: ${metric.memoryUsage.used}MB`,
        metric,
        ['Monitor memory usage', 'Consider reducing file selections']
      );
    }

    // Response time analysis
    if (metric.responseTime && metric.responseTime > APP_CONFIG.performance.monitoring.SLOW_RESPONSE_THRESHOLD_MS) {
      this.createAlert('warning', 'response_time',
        `Slow operation: ${metric.operationType} took ${metric.responseTime.toFixed(0)}ms`,
        metric,
        ['Check network connection', 'Reduce operation size']
      );
    }

    // Error analysis
    if (!metric.success && metric.error) {
      const errorPattern = this.isKnownErrorPattern(metric.error);
      if (errorPattern.isKnown) {
        this.createAlert(errorPattern.severity as any, errorPattern.type as any,
          `Known error pattern detected: ${metric.error}`,
          metric
        );
      }
    }
  }

  private createAlert(
    level: PerformanceAlert['level'],
    type: PerformanceAlert['type'],
    message: string,
    metrics: PerformanceMetrics,
    suggestions?: string[]
  ): void {
    const alert: PerformanceAlert = {
      id: this.generateAlertId(),
      level,
      type,
      message,
      timestamp: Date.now(),
      metrics,
      suggestions
    };

    this.alerts.push(alert);
    this.trimAlerts();

    // Notify listeners
    this.alertListeners.forEach(listener => {
      try {
        listener(alert);
      } catch (error) {
        console.error('Error in alert listener:', error);
      }
    });

    console.warn(`Performance alert [${level}/${type}]:`, message);
  }

  private trimMetrics(): void {
    const maxMetrics = APP_CONFIG.performance.limits.MAX_CACHE_SIZE;
    if (this.metrics.length > maxMetrics) {
      this.metrics = this.metrics.slice(-maxMetrics);
    }
  }

  private trimAlerts(): void {
    const maxAlerts = APP_CONFIG.performance.monitoring.ALERTS_RETENTION_COUNT * 5; // Keep more in runtime
    if (this.alerts.length > maxAlerts) {
      this.alerts = this.alerts.slice(-maxAlerts);
    }
  }

  private getBrowserMemoryInfo(): { usedJSHeapSize: number; jsHeapSizeLimit: number } | null {
    if ('performance' in window && 'memory' in (performance as any)) {
      return (performance as any).memory;
    }
    return null;
  }

  private getCircuitBreakerState(key: string, config: CircuitBreakerConfig): CircuitBreakerState {
    if (!this.circuitBreakers.has(key)) {
      this.circuitBreakers.set(key, {
        state: 'CLOSED',
        failures: 0,
        lastFailureTime: 0,
        nextAttemptTime: 0
      });
    }
    return this.circuitBreakers.get(key)!;
  }

  private generateTraceId(): string {
    return `trace_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
  }

  private generateSpanId(): string {
    return `span_${Math.random().toString(36).substring(2, 9)}`;
  }

  private generateAlertId(): string {
    return `alert_${Date.now()}_${Math.random().toString(36).substring(2, 6)}`;
  }
}

// Default instance for dependency injection
export const defaultPerformanceMonitoringService = new PerformanceMonitoringService();

// Factory function for dependency injection with observability service
export const createPerformanceMonitoringService = (observabilityService?: ObservabilityService) => 
  new PerformanceMonitoringService(observabilityService);