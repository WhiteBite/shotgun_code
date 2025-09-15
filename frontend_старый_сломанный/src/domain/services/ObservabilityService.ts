/**
 * Observability Service
 * 
 * Domain service that provides comprehensive observability patterns including
 * structured logging, distributed tracing, and metrics collection following DDD principles.
 */

import { APP_CONFIG } from '@/config/app-config';

export type LogLevel = 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal';

export interface LogEntry {
  timestamp: number;
  level: LogLevel;
  message: string;
  service: string;
  traceId?: string;
  spanId?: string;
  userId?: string;
  sessionId?: string;
  context?: Record<string, unknown>;
  metadata?: {
    userAgent?: string;
    url?: string;
    referrer?: string;
    viewport?: { width: number; height: number };
    memory?: { used: number; total: number };
    performance?: { navigation?: number; paint?: number };
  };
  tags?: Record<string, string | number | boolean>;
  error?: {
    name: string;
    message: string;
    stack?: string;
    code?: string | number;
  };
}

export interface Span {
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  operationName: string;
  startTime: number;
  endTime?: number;
  duration?: number;
  tags: Record<string, string | number | boolean>;
  logs: Array<{ timestamp: number; fields: Record<string, unknown> }>;
  status: 'ok' | 'error' | 'cancelled' | 'unknown';
  error?: Error;
}

export interface Metric {
  name: string;
  type: 'counter' | 'gauge' | 'histogram' | 'timer';
  value: number;
  timestamp: number;
  tags?: Record<string, string>;
  unit?: string;
  help?: string;
}

export interface ObservabilityConfig {
  logging: {
    level: LogLevel;
    enableConsole: boolean;
    enableStorage: boolean;
    maxLogEntries: number;
    enableStructuredLogging: boolean;
  };
  tracing: {
    enabled: boolean;
    sampleRate: number;
    maxSpans: number;
    enableAutoInstrumentation: boolean;
  };
  metrics: {
    enabled: boolean;
    collectionInterval: number;
    maxMetrics: number;
    enableSystemMetrics: boolean;
  };
}

export class ObservabilityService {
  private config: ObservabilityConfig;
  private logs: LogEntry[] = [];
  private spans: Map<string, Span> = new Map();
  private activeSpans: Map<string, Span> = new Map();
  private metrics: Metric[] = [];
  private sessionId: string;
  private logListeners: ((entry: LogEntry) => void)[] = [];
  private metricsCollectionTimer?: number;

  constructor(config?: Partial<ObservabilityConfig>) {
    this.config = {
      logging: {
        level: 'info',
        enableConsole: APP_CONFIG.performance.monitoring.PERFORMANCE_LOG_ENABLED,
        enableStorage: true,
        maxLogEntries: 1000,
        enableStructuredLogging: true
      },
      tracing: {
        enabled: true,
        sampleRate: 1.0,
        maxSpans: 500,
        enableAutoInstrumentation: true
      },
      metrics: {
        enabled: true,
        collectionInterval: 30000, // 30 seconds
        maxMetrics: 1000,
        enableSystemMetrics: true
      },
      ...config
    };

    this.sessionId = this.generateSessionId();
    this.initializeSystemMetrics();
  }

  // Structured Logging Methods

  /**
   * Log a message with structured context
   */
  log(level: LogLevel, message: string, context?: Record<string, unknown>, error?: Error): void {
    if (!this.shouldLog(level)) {
      return;
    }

    const entry: LogEntry = {
      timestamp: Date.now(),
      level,
      message,
      service: 'shotgun-code-frontend',
      sessionId: this.sessionId,
      context,
      metadata: this.collectMetadata(),
      tags: {
        environment: 'development', // Would be configurable
        version: '1.0.0' // Would come from package.json
      }
    };

    // Add trace context if available
    const currentSpan = this.getCurrentSpan();
    if (currentSpan) {
      entry.traceId = currentSpan.traceId;
      entry.spanId = currentSpan.spanId;
    }

    // Add error information
    if (error) {
      entry.error = {
        name: error.name,
        message: error.message,
        stack: error.stack,
        code: (error as any).code
      };
    }

    this.addLogEntry(entry);
  }

  trace(message: string, context?: Record<string, unknown>): void {
    this.log('trace', message, context);
  }

  debug(message: string, context?: Record<string, unknown>): void {
    this.log('debug', message, context);
  }

  info(message: string, context?: Record<string, unknown>): void {
    this.log('info', message, context);
  }

  warn(message: string, context?: Record<string, unknown>, error?: Error): void {
    this.log('warn', message, context, error);
  }

  error(message: string, context?: Record<string, unknown>, error?: Error): void {
    this.log('error', message, context, error);
  }

  fatal(message: string, context?: Record<string, unknown>, error?: Error): void {
    this.log('fatal', message, context, error);
  }

  // Distributed Tracing Methods

  /**
   * Start a new trace
   */
  startTrace(operationName: string, tags?: Record<string, string | number | boolean>): Span {
    const traceId = this.generateTraceId();
    const spanId = this.generateSpanId();

    const span: Span = {
      traceId,
      spanId,
      operationName,
      startTime: performance.now(),
      tags: { ...tags, service: 'shotgun-code-frontend' },
      logs: [],
      status: 'unknown'
    };

    this.spans.set(spanId, span);
    this.activeSpans.set(traceId, span);

    this.debug(`Started trace: ${operationName}`, { traceId, spanId });
    return span;
  }

  /**
   * Start a child span
   */
  startSpan(operationName: string, parentSpan?: Span, tags?: Record<string, string | number | boolean>): Span {
    const parent = parentSpan || this.getCurrentSpan();
    const spanId = this.generateSpanId();

    const span: Span = {
      traceId: parent?.traceId || this.generateTraceId(),
      spanId,
      parentSpanId: parent?.spanId,
      operationName,
      startTime: performance.now(),
      tags: { ...tags, service: 'shotgun-code-frontend' },
      logs: [],
      status: 'unknown'
    };

    this.spans.set(spanId, span);
    if (!parent) {
      this.activeSpans.set(span.traceId, span);
    }

    return span;
  }

  /**
   * Finish a span
   */
  finishSpan(span: Span, status: Span['status'] = 'ok', error?: Error): void {
    span.endTime = performance.now();
    span.duration = span.endTime - span.startTime;
    span.status = status;

    if (error) {
      span.error = error;
      span.tags.error = true;
    }

    // Remove from active spans if this is the root span
    if (!span.parentSpanId) {
      this.activeSpans.delete(span.traceId);
    }

    this.debug(`Finished span: ${span.operationName}`, {
      traceId: span.traceId,
      spanId: span.spanId,
      duration: span.duration,
      status
    });

    // Cleanup old spans
    this.cleanupSpans();
  }

  /**
   * Add log to current span
   */
  addSpanLog(fields: Record<string, unknown>, span?: Span): void {
    const target = span || this.getCurrentSpan();
    if (target) {
      target.logs.push({
        timestamp: performance.now(),
        fields
      });
    }
  }

  /**
   * Set span tags
   */
  setSpanTags(tags: Record<string, string | number | boolean>, span?: Span): void {
    const target = span || this.getCurrentSpan();
    if (target) {
      Object.assign(target.tags, tags);
    }
  }

  // Metrics Collection Methods

  /**
   * Record a counter metric
   */
  counter(name: string, value: number = 1, tags?: Record<string, string>): void {
    this.recordMetric({
      name,
      type: 'counter',
      value,
      timestamp: Date.now(),
      tags
    });
  }

  /**
   * Record a gauge metric
   */
  gauge(name: string, value: number, tags?: Record<string, string>): void {
    this.recordMetric({
      name,
      type: 'gauge',
      value,
      timestamp: Date.now(),
      tags
    });
  }

  /**
   * Record a histogram metric
   */
  histogram(name: string, value: number, tags?: Record<string, string>): void {
    this.recordMetric({
      name,
      type: 'histogram',
      value,
      timestamp: Date.now(),
      tags
    });
  }

  /**
   * Time an operation
   */
  timer<T>(name: string, operation: () => T | Promise<T>, tags?: Record<string, string>): Promise<T> {
    const startTime = performance.now();
    
    const finish = (result: T, error?: Error) => {
      const duration = performance.now() - startTime;
      this.recordMetric({
        name,
        type: 'timer',
        value: duration,
        timestamp: Date.now(),
        tags: { ...tags, success: !error },
        unit: 'ms'
      });
      
      if (error) {
        this.counter(`${name}.errors`, 1, tags);
      }
      
      return result;
    };

    try {
      const result = operation();
      
      if (result instanceof Promise) {
        return result.then(
          (res) => finish(res),
          (err) => finish(result, err)
        );
      } else {
        return Promise.resolve(finish(result));
      }
    } catch (error) {
      return Promise.resolve(finish(undefined as any, error as Error));
    }
  }

  // Query and Management Methods

  /**
   * Get logs with filtering
   */
  getLogs(filter?: {
    level?: LogLevel;
    service?: string;
    traceId?: string;
    since?: number;
    limit?: number;
  }): LogEntry[] {
    let filtered = [...this.logs];

    if (filter) {
      if (filter.level) {
        const levelOrder = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'];
        const minLevelIndex = levelOrder.indexOf(filter.level);
        filtered = filtered.filter(log => levelOrder.indexOf(log.level) >= minLevelIndex);
      }

      if (filter.service) {
        filtered = filtered.filter(log => log.service === filter.service);
      }

      if (filter.traceId) {
        filtered = filtered.filter(log => log.traceId === filter.traceId);
      }

      if (filter.since) {
        filtered = filtered.filter(log => log.timestamp >= filter.since!);
      }

      if (filter.limit) {
        filtered = filtered.slice(-filter.limit);
      }
    }

    return filtered.sort((a, b) => b.timestamp - a.timestamp);
  }

  /**
   * Get span by ID
   */
  getSpan(spanId: string): Span | undefined {
    return this.spans.get(spanId);
  }

  /**
   * Get all spans for a trace
   */
  getTrace(traceId: string): Span[] {
    return Array.from(this.spans.values()).filter(span => span.traceId === traceId);
  }

  /**
   * Get metrics
   */
  getMetrics(filter?: {
    name?: string;
    type?: Metric['type'];
    since?: number;
    limit?: number;
  }): Metric[] {
    let filtered = [...this.metrics];

    if (filter) {
      if (filter.name) {
        filtered = filtered.filter(metric => metric.name.includes(filter.name!));
      }

      if (filter.type) {
        filtered = filtered.filter(metric => metric.type === filter.type);
      }

      if (filter.since) {
        filtered = filtered.filter(metric => metric.timestamp >= filter.since!);
      }

      if (filter.limit) {
        filtered = filtered.slice(-filter.limit);
      }
    }

    return filtered.sort((a, b) => b.timestamp - a.timestamp);
  }

  /**
   * Add log listener
   */
  onLog(listener: (entry: LogEntry) => void): () => void {
    this.logListeners.push(listener);
    return () => {
      const index = this.logListeners.indexOf(listener);
      if (index > -1) {
        this.logListeners.splice(index, 1);
      }
    };
  }

  /**
   * Clear all observability data
   */
  clear(): void {
    this.logs = [];
    this.spans.clear();
    this.activeSpans.clear();
    this.metrics = [];
  }

  /**
   * Export observability data
   */
  export(): {
    logs: LogEntry[];
    spans: Span[];
    metrics: Metric[];
    metadata: {
      sessionId: string;
      exportTime: number;
      config: ObservabilityConfig;
    };
  } {
    return {
      logs: this.logs,
      spans: Array.from(this.spans.values()),
      metrics: this.metrics,
      metadata: {
        sessionId: this.sessionId,
        exportTime: Date.now(),
        config: this.config
      }
    };
  }

  /**
   * Cleanup resources
   */
  cleanup(): void {
    if (this.metricsCollectionTimer) {
      clearInterval(this.metricsCollectionTimer);
    }
    this.clear();
  }

  // Private helper methods

  private shouldLog(level: LogLevel): boolean {
    const levels = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'];
    const minLevel = levels.indexOf(this.config.logging.level);
    const currentLevel = levels.indexOf(level);
    return currentLevel >= minLevel;
  }

  private addLogEntry(entry: LogEntry): void {
    this.logs.push(entry);

    // Trim logs if exceeding limit
    if (this.logs.length > this.config.logging.maxLogEntries) {
      this.logs = this.logs.slice(-this.config.logging.maxLogEntries);
    }

    // Console logging
    if (this.config.logging.enableConsole) {
      const consoleMethod = this.getConsoleMethod(entry.level);
      if (this.config.logging.enableStructuredLogging) {
        consoleMethod(`[${entry.level.toUpperCase()}] ${entry.message}`, entry);
      } else {
        consoleMethod(`[${entry.level.toUpperCase()}] ${entry.message}`);
      }
    }

    // Notify listeners
    this.logListeners.forEach(listener => {
      try {
        listener(entry);
      } catch (error) {
        console.error('Error in log listener:', error);
      }
    });
  }

  private getConsoleMethod(level: LogLevel): (message: string, ...args: any[]) => void {
    switch (level) {
      case 'trace':
      case 'debug':
        return console.debug;
      case 'info':
        return console.info;
      case 'warn':
        return console.warn;
      case 'error':
      case 'fatal':
        return console.error;
      default:
        return console.log;
    }
  }

  private recordMetric(metric: Metric): void {
    if (!this.config.metrics.enabled) {
      return;
    }

    this.metrics.push(metric);

    // Trim metrics if exceeding limit
    if (this.metrics.length > this.config.metrics.maxMetrics) {
      this.metrics = this.metrics.slice(-this.config.metrics.maxMetrics);
    }
  }

  private getCurrentSpan(): Span | undefined {
    // Return the most recently created active span
    const spans = Array.from(this.activeSpans.values());
    return spans[spans.length - 1];
  }

  private cleanupSpans(): void {
    if (this.spans.size > this.config.tracing.maxSpans) {
      const sortedSpans = Array.from(this.spans.entries())
        .sort(([, a], [, b]) => (a.endTime || a.startTime) - (b.endTime || b.startTime));

      // Keep only the most recent spans
      const toKeep = sortedSpans.slice(-Math.floor(this.config.tracing.maxSpans * 0.8));
      this.spans.clear();
      toKeep.forEach(([id, span]) => this.spans.set(id, span));
    }
  }

  private collectMetadata(): LogEntry['metadata'] {
    return {
      userAgent: navigator.userAgent,
      url: window.location.href,
      referrer: document.referrer,
      viewport: {
        width: window.innerWidth,
        height: window.innerHeight
      },
      memory: this.getMemoryInfo(),
      performance: this.getPerformanceInfo()
    };
  }

  private getMemoryInfo(): { used: number; total: number } | undefined {
    if ('memory' in performance) {
      const memory = (performance as any).memory;
      return {
        used: Math.round(memory.usedJSHeapSize / (1024 * 1024)),
        total: Math.round(memory.jsHeapSizeLimit / (1024 * 1024))
      };
    }
    return undefined;
  }

  private getPerformanceInfo(): { navigation?: number; paint?: number } | undefined {
    try {
      const navigation = performance.getEntriesByType('navigation')[0] as any;
      const paint = performance.getEntriesByType('paint').find(entry => entry.name === 'first-contentful-paint') as any;

      return {
        navigation: navigation?.loadEventEnd - navigation?.navigationStart,
        paint: paint?.startTime
      };
    } catch {
      return undefined;
    }
  }

  private initializeSystemMetrics(): void {
    if (!this.config.metrics.enableSystemMetrics) {
      return;
    }

    this.metricsCollectionTimer = window.setInterval(() => {
      this.collectSystemMetrics();
    }, this.config.metrics.collectionInterval);
  }

  private collectSystemMetrics(): void {
    // Memory metrics
    const memory = this.getMemoryInfo();
    if (memory) {
      this.gauge('system.memory.used', memory.used, { unit: 'MB' });
      this.gauge('system.memory.total', memory.total, { unit: 'MB' });
      this.gauge('system.memory.usage_percent', (memory.used / memory.total) * 100);
    }

    // Performance metrics
    this.gauge('system.spans.active', this.activeSpans.size);
    this.gauge('system.spans.total', this.spans.size);
    this.gauge('system.logs.count', this.logs.length);
    this.gauge('system.metrics.count', this.metrics.length);
  }

  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
  }

  private generateTraceId(): string {
    return `trace_${Date.now()}_${Math.random().toString(36).substring(2, 15)}`;
  }

  private generateSpanId(): string {
    return `span_${Math.random().toString(36).substring(2, 11)}`;
  }
}

// Default instance for dependency injection
export const defaultObservabilityService = new ObservabilityService();