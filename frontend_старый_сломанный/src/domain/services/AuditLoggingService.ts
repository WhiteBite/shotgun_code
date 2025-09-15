/**
 * Audit Logging Service
 * 
 * Domain service that provides comprehensive security audit trail functionality
 * following DDD principles and least privilege principle.
 */

import { APP_CONFIG } from '@/config/app-config';
import type { LocalStorageService } from './LocalStorageService';
import { defaultLocalStorageService } from './LocalStorageService';

export type AuditEventType = 
  | 'authentication' 
  | 'authorization' 
  | 'data_access' 
  | 'data_modification' 
  | 'configuration_change' 
  | 'security_violation' 
  | 'system_event' 
  | 'user_action'
  | 'api_call'
  | 'file_operation'
  | 'validation_failure'
  | 'suspicious_activity';

export type AuditSeverity = 'info' | 'warning' | 'error' | 'critical';

export interface AuditEvent {
  id: string;
  timestamp: number;
  eventType: AuditEventType;
  severity: AuditSeverity;
  source: string;
  action: string;
  resource?: string;
  userId?: string;
  sessionId?: string;
  ipAddress?: string;
  userAgent?: string;
  details: Record<string, unknown>;
  outcome: 'success' | 'failure' | 'pending';
  riskLevel?: 'low' | 'medium' | 'high' | 'critical';
  compliance?: {
    regulation?: string;
    requirement?: string;
    satisfied: boolean;
  };
}

export interface AuditQuery {
  startTime?: number;
  endTime?: number;
  eventTypes?: AuditEventType[];
  severity?: AuditSeverity[];
  source?: string;
  userId?: string;
  outcome?: ('success' | 'failure' | 'pending')[];
  riskLevel?: ('low' | 'medium' | 'high' | 'critical')[];
  searchText?: string;
  limit?: number;
  offset?: number;
}

export interface AuditStatistics {
  totalEvents: number;
  eventsByType: Record<AuditEventType, number>;
  eventsBySeverity: Record<AuditSeverity, number>;
  eventsByOutcome: Record<'success' | 'failure' | 'pending', number>;
  securityViolations: number;
  suspiciousActivities: number;
  averageEventsPerDay: number;
  topSources: Array<{ source: string; count: number }>;
  topActions: Array<{ action: string; count: number }>;
  complianceViolations: number;
}

export interface AuditRetentionPolicy {
  defaultRetentionDays: number;
  criticalRetentionDays: number;
  complianceRetentionDays: number;
  maxStorageSize: number;
  compressionEnabled: boolean;
  encryptionEnabled: boolean;
}

export class AuditLoggingService {
  private events: AuditEvent[] = [];
  private storageService: LocalStorageService;
  private sessionId: string;
  private eventIdCounter = 0;

  constructor(storageService: LocalStorageService) {
    this.storageService = storageService;
    this.sessionId = this.generateSessionId();
    this.loadPersistedEvents();
    this.initializeCleanupScheduler();
  }

  /**
   * Log a security or system event
   */
  async logEvent(event: Omit<AuditEvent, 'id' | 'timestamp' | 'sessionId'>): Promise<string> {
    const auditEvent: AuditEvent = {
      ...event,
      id: this.generateEventId(),
      timestamp: Date.now(),
      sessionId: this.sessionId
    };

    // Validate event before logging
    this.validateEvent(auditEvent);

    // Add to in-memory store
    this.events.push(auditEvent);

    // Persist to storage if enabled
    if (APP_CONFIG.security.audit.LOG_FAILED_VALIDATIONS || 
        APP_CONFIG.security.audit.LOG_SUSPICIOUS_ACTIVITY) {
      await this.persistEvent(auditEvent);
    }

    // Handle critical events immediately
    if (auditEvent.severity === 'critical' || auditEvent.riskLevel === 'critical') {
      await this.handleCriticalEvent(auditEvent);
    }

    // Enforce storage limits
    this.enforceStorageLimits();

    return auditEvent.id;
  }

  /**
   * Log authentication event
   */
  async logAuthentication(
    action: 'login' | 'logout' | 'session_start' | 'session_end' | 'token_refresh',
    outcome: 'success' | 'failure',
    details: Record<string, unknown> = {}
  ): Promise<string> {
    return this.logEvent({
      eventType: 'authentication',
      severity: outcome === 'failure' ? 'warning' : 'info',
      source: 'auth_system',
      action,
      outcome,
      details: {
        ...details,
        userAgent: navigator.userAgent
      }
    });
  }

  /**
   * Log authorization event
   */
  async logAuthorization(
    action: string,
    resource: string,
    outcome: 'success' | 'failure',
    details: Record<string, unknown> = {}
  ): Promise<string> {
    return this.logEvent({
      eventType: 'authorization',
      severity: outcome === 'failure' ? 'warning' : 'info',
      source: 'auth_system',
      action,
      resource,
      outcome,
      details
    });
  }

  /**
   * Log data access event
   */
  async logDataAccess(
    resource: string,
    action: 'read' | 'list' | 'search' | 'export',
    outcome: 'success' | 'failure',
    details: Record<string, unknown> = {}
  ): Promise<string> {
    return this.logEvent({
      eventType: 'data_access',
      severity: 'info',
      source: 'data_system',
      action,
      resource,
      outcome,
      details
    });
  }

  /**
   * Log data modification event
   */
  async logDataModification(
    resource: string,
    action: 'create' | 'update' | 'delete' | 'move' | 'copy',
    outcome: 'success' | 'failure',
    details: Record<string, unknown> = {}
  ): Promise<string> {
    return this.logEvent({
      eventType: 'data_modification',
      severity: action === 'delete' ? 'warning' : 'info',
      source: 'data_system',
      action,
      resource,
      outcome,
      details
    });
  }

  /**
   * Log security violation
   */
  async logSecurityViolation(
    violation: string,
    resource?: string,
    riskLevel: 'low' | 'medium' | 'high' | 'critical' = 'medium',
    details: Record<string, unknown> = {}
  ): Promise<string> {
    return this.logEvent({
      eventType: 'security_violation',
      severity: riskLevel === 'critical' ? 'critical' : 'error',
      source: 'security_system',
      action: violation,
      resource,
      outcome: 'failure',
      riskLevel,
      details
    });
  }

  /**
   * Log validation failure
   */
  async logValidationFailure(
    validation: string,
    resource?: string,
    details: Record<string, unknown> = {}
  ): Promise<string> {
    return this.logEvent({
      eventType: 'validation_failure',
      severity: 'warning',
      source: 'validation_system',
      action: validation,
      resource,
      outcome: 'failure',
      details
    });
  }

  /**
   * Log file operation
   */
  async logFileOperation(
    operation: 'open' | 'close' | 'read' | 'write' | 'delete' | 'move' | 'copy',
    filePath: string,
    outcome: 'success' | 'failure',
    details: Record<string, unknown> = {}
  ): Promise<string> {
    return this.logEvent({
      eventType: 'file_operation',
      severity: operation === 'delete' ? 'warning' : 'info',
      source: 'file_system',
      action: operation,
      resource: filePath,
      outcome,
      details
    });
  }

  /**
   * Log API call
   */
  async logApiCall(
    endpoint: string,
    method: string,
    outcome: 'success' | 'failure',
    responseTime?: number,
    details: Record<string, unknown> = {}
  ): Promise<string> {
    return this.logEvent({
      eventType: 'api_call',
      severity: outcome === 'failure' ? 'warning' : 'info',
      source: 'api_system',
      action: `${method} ${endpoint}`,
      outcome,
      details: {
        ...details,
        responseTime,
        endpoint,
        method
      }
    });
  }

  /**
   * Query audit events
   */
  queryEvents(query: AuditQuery = {}): AuditEvent[] {
    let filtered = [...this.events];

    // Apply time filters
    if (query.startTime) {
      filtered = filtered.filter(event => event.timestamp >= query.startTime!);
    }
    if (query.endTime) {
      filtered = filtered.filter(event => event.timestamp <= query.endTime!);
    }

    // Apply type filters
    if (query.eventTypes && query.eventTypes.length > 0) {
      filtered = filtered.filter(event => query.eventTypes!.includes(event.eventType));
    }

    // Apply severity filters
    if (query.severity && query.severity.length > 0) {
      filtered = filtered.filter(event => query.severity!.includes(event.severity));
    }

    // Apply source filter
    if (query.source) {
      filtered = filtered.filter(event => event.source === query.source);
    }

    // Apply user filter
    if (query.userId) {
      filtered = filtered.filter(event => event.userId === query.userId);
    }

    // Apply outcome filters
    if (query.outcome && query.outcome.length > 0) {
      filtered = filtered.filter(event => query.outcome!.includes(event.outcome));
    }

    // Apply risk level filters
    if (query.riskLevel && query.riskLevel.length > 0) {
      filtered = filtered.filter(event => 
        event.riskLevel && query.riskLevel!.includes(event.riskLevel)
      );
    }

    // Apply text search
    if (query.searchText) {
      const searchLower = query.searchText.toLowerCase();
      filtered = filtered.filter(event =>
        event.action.toLowerCase().includes(searchLower) ||
        event.source.toLowerCase().includes(searchLower) ||
        (event.resource && event.resource.toLowerCase().includes(searchLower)) ||
        JSON.stringify(event.details).toLowerCase().includes(searchLower)
      );
    }

    // Sort by timestamp (most recent first)
    filtered.sort((a, b) => b.timestamp - a.timestamp);

    // Apply pagination
    const offset = query.offset || 0;
    const limit = query.limit || 100;
    
    return filtered.slice(offset, offset + limit);
  }

  /**
   * Get audit statistics
   */
  getStatistics(timeRange?: { start: number; end: number }): AuditStatistics {
    let events = this.events;
    
    if (timeRange) {
      events = events.filter(event => 
        event.timestamp >= timeRange.start && event.timestamp <= timeRange.end
      );
    }

    const eventsByType: Record<AuditEventType, number> = {
      authentication: 0,
      authorization: 0,
      data_access: 0,
      data_modification: 0,
      configuration_change: 0,
      security_violation: 0,
      system_event: 0,
      user_action: 0,
      api_call: 0,
      file_operation: 0,
      validation_failure: 0,
      suspicious_activity: 0
    };

    const eventsBySeverity: Record<AuditSeverity, number> = {
      info: 0,
      warning: 0,
      error: 0,
      critical: 0
    };

    const eventsByOutcome: Record<'success' | 'failure' | 'pending', number> = {
      success: 0,
      failure: 0,
      pending: 0
    };

    const sourceCounts: Record<string, number> = {};
    const actionCounts: Record<string, number> = {};

    let securityViolations = 0;
    let suspiciousActivities = 0;
    let complianceViolations = 0;

    events.forEach(event => {
      eventsByType[event.eventType]++;
      eventsBySeverity[event.severity]++;
      eventsByOutcome[event.outcome]++;

      sourceCounts[event.source] = (sourceCounts[event.source] || 0) + 1;
      actionCounts[event.action] = (actionCounts[event.action] || 0) + 1;

      if (event.eventType === 'security_violation') {
        securityViolations++;
      }
      
      if (event.eventType === 'suspicious_activity') {
        suspiciousActivities++;
      }

      if (event.compliance && !event.compliance.satisfied) {
        complianceViolations++;
      }
    });

    const topSources = Object.entries(sourceCounts)
      .map(([source, count]) => ({ source, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10);

    const topActions = Object.entries(actionCounts)
      .map(([action, count]) => ({ action, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10);

    // Calculate average events per day
    const timeSpanMs = timeRange 
      ? (timeRange.end - timeRange.start)
      : (Date.now() - (events[events.length - 1]?.timestamp || Date.now()));
    const timeSpanDays = Math.max(1, timeSpanMs / (24 * 60 * 60 * 1000));
    const averageEventsPerDay = events.length / timeSpanDays;

    return {
      totalEvents: events.length,
      eventsByType,
      eventsBySeverity,
      eventsByOutcome,
      securityViolations,
      suspiciousActivities,
      averageEventsPerDay,
      topSources,
      topActions,
      complianceViolations
    };
  }

  /**
   * Export audit log for compliance
   */
  exportAuditLog(query: AuditQuery = {}, format: 'json' | 'csv' = 'json'): string {
    const events = this.queryEvents(query);
    
    if (format === 'csv') {
      return this.exportToCsv(events);
    }
    
    return JSON.stringify({
      exportMetadata: {
        timestamp: Date.now(),
        totalEvents: events.length,
        query,
        exportFormat: format
      },
      events
    }, null, 2);
  }

  /**
   * Clear audit log (use with caution)
   */
  clearAuditLog(): void {
    this.events = [];
    this.storageService.remove('audit_events');
  }

  /**
   * Get retention policy
   */
  getRetentionPolicy(): AuditRetentionPolicy {
    return {
      defaultRetentionDays: APP_CONFIG.security.audit.AUDIT_LOG_RETENTION_DAYS,
      criticalRetentionDays: APP_CONFIG.security.audit.AUDIT_LOG_RETENTION_DAYS * 2,
      complianceRetentionDays: APP_CONFIG.security.audit.AUDIT_LOG_RETENTION_DAYS * 3,
      maxStorageSize: APP_CONFIG.security.audit.MAX_AUDIT_LOG_SIZE,
      compressionEnabled: true,
      encryptionEnabled: false // Would need proper encryption implementation
    };
  }

  // Private helper methods

  private generateEventId(): string {
    return `audit_${Date.now()}_${++this.eventIdCounter}`;
  }

  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private validateEvent(event: AuditEvent): void {
    if (!event.eventType || !event.source || !event.action) {
      throw new Error('Invalid audit event: missing required fields');
    }
  }

  private async persistEvent(event: AuditEvent): Promise<void> {
    try {
      const persistedEvents = this.storageService.get<AuditEvent[]>('audit_events', []);
      persistedEvents.push(event);
      
      // Keep only recent events to prevent storage overflow
      const maxEvents = 1000;
      if (persistedEvents.length > maxEvents) {
        persistedEvents.splice(0, persistedEvents.length - maxEvents);
      }
      
      this.storageService.set('audit_events', persistedEvents);
    } catch (error) {
      console.error('Failed to persist audit event:', error);
    }
  }

  private loadPersistedEvents(): void {
    try {
      // Use getItem instead of get
      this.storageService.getItem<AuditEvent[]>('audit_events').then(persistedEvents => {
        this.events = persistedEvents || [];
      }).catch(error => {
        console.error('Failed to load persisted audit events:', error);
        this.events = [];
      });
    } catch (error) {
      console.error('Failed to load persisted audit events:', error);
      this.events = [];
    }
  }

  private async handleCriticalEvent(event: AuditEvent): Promise<void> {
    // In a real implementation, this might:
    // - Send alerts to administrators
    // - Trigger automatic responses
    // - Escalate to security team
    console.error('Critical audit event:', event);
  }

  private enforceStorageLimits(): void {
    const maxEvents = 5000; // Reasonable in-memory limit
    if (this.events.length > maxEvents) {
      // Remove oldest events, but keep critical ones longer
      const criticalEvents = this.events.filter(e => e.severity === 'critical');
      const normalEvents = this.events.filter(e => e.severity !== 'critical');
      
      const keepNormalEvents = normalEvents.slice(-Math.floor(maxEvents * 0.8));
      const keepCriticalEvents = criticalEvents.slice(-Math.floor(maxEvents * 0.2));
      
      this.events = [...keepNormalEvents, ...keepCriticalEvents]
        .sort((a, b) => a.timestamp - b.timestamp);
    }
  }

  private initializeCleanupScheduler(): void {
    // Run cleanup every 6 hours
    setInterval(() => {
      this.cleanupOldEvents();
    }, 6 * 60 * 60 * 1000);
  }

  private cleanupOldEvents(): void {
    const retentionPolicy = this.getRetentionPolicy();
    const now = Date.now();
    
    this.events = this.events.filter(event => {
      const eventAge = now - event.timestamp;
      const retentionMs = this.getRetentionForEvent(event, retentionPolicy) * 24 * 60 * 60 * 1000;
      
      return eventAge < retentionMs;
    });
    
    // Also clean up persisted events
    this.storageService.set('audit_events', this.events);
  }

  private getRetentionForEvent(event: AuditEvent, policy: AuditRetentionPolicy): number {
    if (event.severity === 'critical' || event.riskLevel === 'critical') {
      return policy.criticalRetentionDays;
    }
    
    if (event.compliance) {
      return policy.complianceRetentionDays;
    }
    
    return policy.defaultRetentionDays;
  }

  private exportToCsv(events: AuditEvent[]): string {
    const headers = [
      'ID', 'Timestamp', 'Event Type', 'Severity', 'Source', 'Action', 
      'Resource', 'User ID', 'Session ID', 'Outcome', 'Risk Level', 'Details'
    ];
    
    const rows = events.map(event => [
      event.id,
      new Date(event.timestamp).toISOString(),
      event.eventType,
      event.severity,
      event.source,
      event.action,
      event.resource || '',
      event.userId || '',
      event.sessionId || '',
      event.outcome,
      event.riskLevel || '',
      JSON.stringify(event.details)
    ]);
    
    return [headers, ...rows]
      .map(row => row.map(cell => `"${cell}"`).join(','))
      .join('\n');
  }
}

// Factory function for dependency injection
export const createAuditLoggingService = (storageService: LocalStorageService) => 
  new AuditLoggingService(storageService);

// Add default instance export
export const defaultAuditLoggingService = new AuditLoggingService(defaultLocalStorageService);
