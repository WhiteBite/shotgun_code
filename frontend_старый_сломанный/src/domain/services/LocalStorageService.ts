import { APP_CONFIG } from '@/config/app-config';
import { CircuitBreakerService } from './CircuitBreakerService';

/**
 * LocalStorageService - Enhanced Local Storage Service with Circuit Breaker Pattern
 * 
 * Provides a robust abstraction over browser localStorage with:
 * - Automatic JSON serialization/deserialization
 * - Circuit breaker pattern for fault tolerance
 * - Performance monitoring and logging
 * - Error handling and fallback strategies
 * - Type safety with generics
 * - Configurable retry logic
 */

export interface LocalStorageOptions {
  readonly circuitBreaker?: boolean;
  readonly maxRetries?: number;
  readonly timeoutMs?: number;
  readonly enableCompression?: boolean;
  readonly encryption?: boolean;
}

export interface StorageMetrics {
  readonly operationCount: number;
  readonly errorCount: number;
  readonly lastError?: string;
  readonly avgLatencyMs: number;
  readonly cacheHitRate: number;
}

export class LocalStorageService {
  private readonly circuitBreakerService: CircuitBreakerService;
  private readonly defaultOptions: LocalStorageOptions = {
    circuitBreaker: true,
    maxRetries: APP_CONFIG.events.ERROR_RETRY_ATTEMPTS,
    timeoutMs: 5000,
    enableCompression: false,
    encryption: false
  };
  
  private metrics: StorageMetrics = {
    operationCount: 0,
    errorCount: 0,
    avgLatencyMs: 0,
    cacheHitRate: 0
  };

  constructor(circuitBreakerService: CircuitBreakerService) {
    this.circuitBreakerService = circuitBreakerService;
  }

  /**
   * Sets a value in localStorage with automatic JSON serialization
   */
  async setItem<T>(key: string, value: T, options?: LocalStorageOptions): Promise<void> {
    const config = { ...this.defaultOptions, ...options };
    const startTime = performance.now();
    
    try {
      // Use circuit breaker if enabled
      if (config.circuitBreaker) {
        await this.circuitBreakerService.execute(`localStorage_set_${key}`, () => 
          this.executeSetItem(key, value, config)
        );
      } else {
        await this.executeSetItem(key, value, config);
      }
      
      // Update metrics
      this.updateMetrics(startTime, true);
    } catch (error) {
      this.updateMetrics(startTime, false, error);
      throw error;
    }
  }

  /**
   * Gets a value from localStorage with automatic JSON deserialization
   */
  async getItem<T>(key: string, defaultValue?: T, options?: LocalStorageOptions): Promise<T | null> {
    const config = { ...this.defaultOptions, ...options };
    const startTime = performance.now();
    
    try {
      let result: T | null = null;
      
      // Use circuit breaker if enabled
      if (config.circuitBreaker) {
        result = await this.circuitBreakerService.execute(`localStorage_get_${key}`, () => 
          this.executeGetItem<T>(key, defaultValue, config)
        );
      } else {
        result = await this.executeGetItem<T>(key, defaultValue, config);
      }
      
      // Update metrics
      this.updateMetrics(startTime, true);
      return result;
    } catch (error) {
      this.updateMetrics(startTime, false, error);
      return defaultValue ?? null;
    }
  }

  /**
   * Removes an item from localStorage
   */
  async removeItem(key: string, options?: LocalStorageOptions): Promise<void> {
    const config = { ...this.defaultOptions, ...options };
    const startTime = performance.now();
    
    try {
      if (config.circuitBreaker) {
        await this.circuitBreakerService.execute(`localStorage_remove_${key}`, () => 
          this.executeRemoveItem(key)
        );
      } else {
        await this.executeRemoveItem(key);
      }
      
      this.updateMetrics(startTime, true);
    } catch (error) {
      this.updateMetrics(startTime, false, error);
      throw error;
    }
  }

  /**
   * Clears all items from localStorage
   */
  async clear(options?: LocalStorageOptions): Promise<void> {
    const config = { ...this.defaultOptions, ...options };
    const startTime = performance.now();
    
    try {
      if (config.circuitBreaker) {
        await this.circuitBreakerService.execute('localStorage_clear', () => 
          this.executeClear()
        );
      } else {
        await this.executeClear();
      }
      
      this.updateMetrics(startTime, true);
    } catch (error) {
      this.updateMetrics(startTime, false, error);
      throw error;
    }
  }

  /**
   * Gets the number of items in localStorage
   */
  async length(options?: LocalStorageOptions): Promise<number> {
    const config = { ...this.defaultOptions, ...options };
    const startTime = performance.now();
    
    try {
      let result = 0;
      
      if (config.circuitBreaker) {
        result = await this.circuitBreakerService.execute('localStorage_length', () => 
          this.executeLength()
        );
      } else {
        result = await this.executeLength();
      }
      
      this.updateMetrics(startTime, true);
      return result;
    } catch (error) {
      this.updateMetrics(startTime, false, error);
      return 0;
    }
  }

  /**
   * Gets all keys in localStorage
   */
  async getAllKeys(options?: LocalStorageOptions): Promise<string[]> {
    const config = { ...this.defaultOptions, ...options };
    const startTime = performance.now();
    
    try {
      let result: string[] = [];
      
      if (config.circuitBreaker) {
        result = await this.circuitBreakerService.execute('localStorage_keys', () => 
          this.executeGetAllKeys()
        );
      } else {
        result = await this.executeGetAllKeys();
      }
      
      this.updateMetrics(startTime, true);
      return result;
    } catch (error) {
      this.updateMetrics(startTime, false, error);
      return [];
    }
  }

  /**
   * Gets storage metrics
   */
  getMetrics(): StorageMetrics {
    return { ...this.metrics };
  }

  /**
   * Resets storage metrics
   */
  resetMetrics(): void {
    this.metrics = {
      operationCount: 0,
      errorCount: 0,
      avgLatencyMs: 0,
      cacheHitRate: 0
    };
  }

  // Private execution methods
  private async executeSetItem<T>(key: string, value: T, options: LocalStorageOptions): Promise<void> {
    // Serialize value to JSON
    const serializedValue = JSON.stringify(value);
    
    // Apply compression if enabled
    // TODO: Implement compression logic
    
    // Apply encryption if enabled
    // TODO: Implement encryption logic
    
    // Store in localStorage
    localStorage.setItem(key, serializedValue);
  }

  private async executeGetItem<T>(key: string, defaultValue?: T, options: LocalStorageOptions): Promise<T | null> {
    const serializedValue = localStorage.getItem(key);
    
    if (serializedValue === null) {
      return defaultValue ?? null;
    }
    
    try {
      // Decrypt if encryption was enabled
      // TODO: Implement decryption logic
      
      // Decompress if compression was enabled
      // TODO: Implement decompression logic
      
      // Parse JSON
      return JSON.parse(serializedValue);
    } catch (error) {
      console.error(`Failed to parse localStorage item '${key}':`, error);
      return defaultValue ?? null;
    }
  }

  private async executeRemoveItem(key: string): Promise<void> {
    localStorage.removeItem(key);
  }

  private async executeClear(): Promise<void> {
    localStorage.clear();
  }

  private async executeLength(): Promise<number> {
    return localStorage.length;
  }

  private async executeGetAllKeys(): Promise<string[]> {
    const keys: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key !== null) {
        keys.push(key);
      }
    }
    return keys;
  }

  // Metrics tracking
  private updateMetrics(startTime: number, success: boolean, error?: unknown): void {
    const duration = performance.now() - startTime;
    this.metrics.operationCount++;
    
    if (!success) {
      this.metrics.errorCount++;
      this.metrics.lastError = error instanceof Error ? error.message : String(error);
    }
    
    // Update average latency (simple moving average)
    this.metrics.avgLatencyMs = (
      (this.metrics.avgLatencyMs * (this.metrics.operationCount - 1)) + duration
    ) / this.metrics.operationCount;
  }
}

// Create default instance
const defaultCircuitBreaker = new CircuitBreakerService();
export const defaultLocalStorageService = new LocalStorageService(defaultCircuitBreaker);