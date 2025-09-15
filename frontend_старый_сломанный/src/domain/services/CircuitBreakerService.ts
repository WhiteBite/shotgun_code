/**
 * Circuit Breaker Service
 * 
 * Domain service that implements the Circuit Breaker pattern for API calls
 * with timeouts, retries, and health checks following DDD principles.
 */

import { APP_CONFIG } from '@/config/app-config';

export type CircuitState = 'CLOSED' | 'OPEN' | 'HALF_OPEN';

export interface CircuitBreakerOptions {
  failureThreshold: number;
  resetTimeoutMs: number;
  monitoringPeriodMs: number;
  timeoutMs?: number;
  retryCount?: number;
  retryDelayMs?: number;
  healthCheckIntervalMs?: number;
  healthCheckPath?: string;
}

export interface CircuitBreakerStatus {
  name: string;
  state: CircuitState;
  failures: number;
  totalCalls: number;
  successfulCalls: number;
  lastFailureTime: number;
  nextAttemptTime: number;
  lastStateChange: number;
  healthCheckStatus?: 'healthy' | 'unhealthy' | 'unknown';
  averageResponseTime?: number;
  uptime: number;
}

export interface CircuitBreakerMetrics {
  totalCircuits: number;
  openCircuits: number;
  halfOpenCircuits: number;
  closedCircuits: number;
  totalCalls: number;
  totalFailures: number;
  successRate: number;
  circuitStatuses: CircuitBreakerStatus[];
}

export interface RetryConfig {
  maxRetries: number;
  baseDelayMs: number;
  exponentialBackoff: boolean;
  maxDelayMs: number;
  jitter: boolean;
}

interface CircuitBreaker {
  name: string;
  state: CircuitState;
  failures: number;
  totalCalls: number;
  successfulCalls: number;
  lastFailureTime: number;
  nextAttemptTime: number;
  lastStateChange: number;
  options: CircuitBreakerOptions;
  responseTimes: number[];
  healthCheckTimer?: number;
  healthCheckStatus: 'healthy' | 'unhealthy' | 'unknown';
}

export class CircuitBreakerService {
  private circuits: Map<string, CircuitBreaker> = new Map();
  private globalHealthCheckTimer?: number;

  constructor() {
    this.startGlobalHealthCheck();
  }

  /**
   * Execute operation with circuit breaker protection
   */
  async execute<T>(
    circuitName: string,
    operation: () => Promise<T>,
    options: Partial<CircuitBreakerOptions> = {}
  ): Promise<T> {
    const circuit = this.getOrCreateCircuit(circuitName, options);
    
    // Check if circuit is open
    if (circuit.state === 'OPEN') {
      if (Date.now() >= circuit.nextAttemptTime) {
        circuit.state = 'HALF_OPEN';
        circuit.lastStateChange = Date.now();
        console.log(`Circuit breaker ${circuitName} transitioning to HALF_OPEN`);
      } else {
        throw new Error(`Circuit breaker ${circuitName} is OPEN. Next attempt at ${new Date(circuit.nextAttemptTime).toISOString()}`);
      }
    }

    const startTime = performance.now();
    circuit.totalCalls++;

    try {
      // Add timeout if specified
      let result: T;
      if (circuit.options.timeoutMs) {
        result = await this.withTimeout(operation, circuit.options.timeoutMs);
      } else {
        result = await operation();
      }

      // Record success
      const responseTime = performance.now() - startTime;
      this.recordSuccess(circuit, responseTime);

      return result;
    } catch (error) {
      const responseTime = performance.now() - startTime;
      this.recordFailure(circuit, error, responseTime);
      throw error;
    }
  }

  /**
   * Execute operation with retries and circuit breaker protection
   */
  async executeWithRetry<T>(
    circuitName: string,
    operation: () => Promise<T>,
    retryConfig: Partial<RetryConfig> = {},
    circuitOptions: Partial<CircuitBreakerOptions> = {}
  ): Promise<T> {
    const config: RetryConfig = {
      maxRetries: circuitOptions.retryCount || 3,
      baseDelayMs: circuitOptions.retryDelayMs || 1000,
      exponentialBackoff: true,
      maxDelayMs: 30000, // 30 seconds max delay
      jitter: true,
      ...retryConfig
    };

    let lastError: unknown;
    
    for (let attempt = 0; attempt <= config.maxRetries; attempt++) {
      try {
        return await this.execute(circuitName, operation, circuitOptions);
      } catch (error) {
        lastError = error;
        
        // Don't retry if circuit is open
        const circuit = this.circuits.get(circuitName);
        if (circuit?.state === 'OPEN') {
          break;
        }
        
        // Don't retry on last attempt
        if (attempt === config.maxRetries) {
          break;
        }
        
        // Calculate delay with exponential backoff and jitter
        let delay = config.baseDelayMs;
        if (config.exponentialBackoff) {
          delay = Math.min(delay * Math.pow(2, attempt), config.maxDelayMs);
        }
        
        if (config.jitter) {
          delay = delay + (Math.random() * delay * 0.1); // Add up to 10% jitter
        }
        
        console.log(`Retry attempt ${attempt + 1}/${config.maxRetries} for ${circuitName} after ${delay}ms`);
        await this.sleep(delay);
      }
    }
    
    throw lastError;
  }

  /**
   * Get circuit breaker status
   */
  getCircuitStatus(circuitName: string): CircuitBreakerStatus | null {
    const circuit = this.circuits.get(circuitName);
    if (!circuit) {
      return null;
    }

    return {
      name: circuitName,
      state: circuit.state,
      failures: circuit.failures,
      totalCalls: circuit.totalCalls,
      successfulCalls: circuit.successfulCalls,
      lastFailureTime: circuit.lastFailureTime,
      nextAttemptTime: circuit.nextAttemptTime,
      lastStateChange: circuit.lastStateChange,
      healthCheckStatus: circuit.healthCheckStatus,
      averageResponseTime: this.calculateAverageResponseTime(circuit),
      uptime: Date.now() - circuit.lastStateChange
    };
  }

  /**
   * Get all circuit breaker metrics
   */
  getMetrics(): CircuitBreakerMetrics {
    const statuses = Array.from(this.circuits.keys())
      .map(name => this.getCircuitStatus(name))
      .filter(status => status !== null) as CircuitBreakerStatus[];

    const totalCalls = statuses.reduce((sum, status) => sum + status.totalCalls, 0);
    const totalFailures = statuses.reduce((sum, status) => sum + status.failures, 0);
    const successRate = totalCalls > 0 ? ((totalCalls - totalFailures) / totalCalls) * 100 : 100;

    return {
      totalCircuits: this.circuits.size,
      openCircuits: statuses.filter(s => s.state === 'OPEN').length,
      halfOpenCircuits: statuses.filter(s => s.state === 'HALF_OPEN').length,
      closedCircuits: statuses.filter(s => s.state === 'CLOSED').length,
      totalCalls,
      totalFailures,
      successRate,
      circuitStatuses: statuses
    };
  }

  /**
   * Reset a circuit breaker to CLOSED state
   */
  resetCircuit(circuitName: string): boolean {
    const circuit = this.circuits.get(circuitName);
    if (!circuit) {
      return false;
    }

    circuit.state = 'CLOSED';
    circuit.failures = 0;
    circuit.lastFailureTime = 0;
    circuit.nextAttemptTime = 0;
    circuit.lastStateChange = Date.now();
    circuit.healthCheckStatus = 'unknown';

    console.log(`Circuit breaker ${circuitName} manually reset to CLOSED`);
    return true;
  }

  /**
   * Remove a circuit breaker
   */
  removeCircuit(circuitName: string): boolean {
    const circuit = this.circuits.get(circuitName);
    if (!circuit) {
      return false;
    }

    // Clear health check timer
    if (circuit.healthCheckTimer) {
      clearInterval(circuit.healthCheckTimer);
    }

    this.circuits.delete(circuitName);
    console.log(`Circuit breaker ${circuitName} removed`);
    return true;
  }

  /**
   * Update circuit breaker options
   */
  updateCircuitOptions(circuitName: string, options: Partial<CircuitBreakerOptions>): boolean {
    const circuit = this.circuits.get(circuitName);
    if (!circuit) {
      return false;
    }

    circuit.options = { ...circuit.options, ...options };
    
    // Restart health check if health check options changed
    if (options.healthCheckIntervalMs || options.healthCheckPath) {
      this.setupHealthCheck(circuit);
    }

    return true;
  }

  /**
   * Check health of all circuits
   */
  async checkAllCircuitsHealth(): Promise<void> {
    const promises = Array.from(this.circuits.entries()).map(([name, circuit]) => {
      if (circuit.options.healthCheckPath) {
        return this.performHealthCheck(circuit);
      }
    });

    await Promise.allSettled(promises);
  }

  /**
   * Cleanup resources
   */
  cleanup(): void {
    // Clear all health check timers
    this.circuits.forEach(circuit => {
      if (circuit.healthCheckTimer) {
        clearInterval(circuit.healthCheckTimer);
      }
    });

    if (this.globalHealthCheckTimer) {
      clearInterval(this.globalHealthCheckTimer);
    }

    this.circuits.clear();
  }

  // Private helper methods

  private getOrCreateCircuit(name: string, options: Partial<CircuitBreakerOptions>): CircuitBreaker {
    if (!this.circuits.has(name)) {
      const defaultOptions: CircuitBreakerOptions = {
        failureThreshold: APP_CONFIG.performance.circuitBreaker.DEFAULT_FAILURE_THRESHOLD,
        resetTimeoutMs: APP_CONFIG.performance.circuitBreaker.DEFAULT_RESET_TIMEOUT_MS,
        monitoringPeriodMs: APP_CONFIG.performance.circuitBreaker.DEFAULT_MONITORING_PERIOD_MS,
        healthCheckIntervalMs: APP_CONFIG.performance.circuitBreaker.HEALTH_CHECK_INTERVAL_MS,
        timeoutMs: 30000, // 30 seconds default timeout
        retryCount: 3,
        retryDelayMs: 1000
      };

      const circuit: CircuitBreaker = {
        name,
        state: 'CLOSED',
        failures: 0,
        totalCalls: 0,
        successfulCalls: 0,
        lastFailureTime: 0,
        nextAttemptTime: 0,
        lastStateChange: Date.now(),
        options: { ...defaultOptions, ...options },
        responseTimes: [],
        healthCheckStatus: 'unknown'
      };

      this.circuits.set(name, circuit);
      this.setupHealthCheck(circuit);

      console.log(`Circuit breaker ${name} created with options:`, circuit.options);
    }

    return this.circuits.get(name)!;
  }

  private recordSuccess(circuit: CircuitBreaker, responseTime: number): void {
    circuit.successfulCalls++;
    circuit.responseTimes.push(responseTime);
    
    // Keep only recent response times for average calculation
    if (circuit.responseTimes.length > 100) {
      circuit.responseTimes = circuit.responseTimes.slice(-50);
    }

    // Reset failures and state if in HALF_OPEN
    if (circuit.state === 'HALF_OPEN') {
      circuit.state = 'CLOSED';
      circuit.failures = 0;
      circuit.lastStateChange = Date.now();
      console.log(`Circuit breaker ${circuit.name} recovered and transitioned to CLOSED`);
    }
  }

  private recordFailure(circuit: CircuitBreaker, error: unknown, responseTime: number): void {
    circuit.failures++;
    circuit.lastFailureTime = Date.now();
    circuit.responseTimes.push(responseTime);

    // Check if we should open the circuit
    if (circuit.failures >= circuit.options.failureThreshold) {
      if (circuit.state !== 'OPEN') {
        circuit.state = 'OPEN';
        circuit.nextAttemptTime = Date.now() + circuit.options.resetTimeoutMs;
        circuit.lastStateChange = Date.now();
        console.warn(`Circuit breaker ${circuit.name} opened due to ${circuit.failures} failures. Next attempt at ${new Date(circuit.nextAttemptTime).toISOString()}`);
      }
    }

    console.warn(`Circuit breaker ${circuit.name} recorded failure:`, error);
  }

  private async withTimeout<T>(operation: () => Promise<T>, timeoutMs: number): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      const timer = setTimeout(() => {
        reject(new Error(`Operation timed out after ${timeoutMs}ms`));
      }, timeoutMs);

      operation()
        .then(result => {
          clearTimeout(timer);
          resolve(result);
        })
        .catch(error => {
          clearTimeout(timer);
          reject(error);
        });
    });
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  private calculateAverageResponseTime(circuit: CircuitBreaker): number {
    if (circuit.responseTimes.length === 0) {
      return 0;
    }
    
    const sum = circuit.responseTimes.reduce((acc, time) => acc + time, 0);
    return sum / circuit.responseTimes.length;
  }

  private setupHealthCheck(circuit: CircuitBreaker): void {
    // Clear existing health check
    if (circuit.healthCheckTimer) {
      clearInterval(circuit.healthCheckTimer);
    }

    // Setup new health check if configured
    if (circuit.options.healthCheckPath && circuit.options.healthCheckIntervalMs) {
      circuit.healthCheckTimer = window.setInterval(
        () => this.performHealthCheck(circuit),
        circuit.options.healthCheckIntervalMs
      );
    }
  }

  private async performHealthCheck(circuit: CircuitBreaker): Promise<void> {
    if (!circuit.options.healthCheckPath) {
      return;
    }

    try {
      // Simple health check - in a real implementation, this would call the actual health endpoint
      const response = await fetch(circuit.options.healthCheckPath, {
        method: 'GET',
        timeout: 5000
      } as RequestInit);

      if (response.ok) {
        circuit.healthCheckStatus = 'healthy';
        
        // If circuit is open but health check passes, consider transitioning to half-open
        if (circuit.state === 'OPEN' && Date.now() >= circuit.nextAttemptTime) {
          circuit.state = 'HALF_OPEN';
          circuit.lastStateChange = Date.now();
          console.log(`Circuit breaker ${circuit.name} health check passed, transitioning to HALF_OPEN`);
        }
      } else {
        circuit.healthCheckStatus = 'unhealthy';
      }
    } catch (error) {
      circuit.healthCheckStatus = 'unhealthy';
      console.warn(`Health check failed for circuit ${circuit.name}:`, error);
    }
  }

  private startGlobalHealthCheck(): void {
    // Global health check every 30 seconds
    this.globalHealthCheckTimer = window.setInterval(
      () => this.checkAllCircuitsHealth(),
      30000
    );
  }
}

// Default instance for dependency injection
export const defaultCircuitBreakerService = new CircuitBreakerService();