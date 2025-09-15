/**
 * WailsApiAdapter - Enhanced API Adapter with Automatic JSON Handling
 * 
 * Provides centralized JSON serialization/deserialization for all Wails API calls.
 * Eliminates manual JSON.parse/JSON.stringify throughout the codebase.
 * Follows Clean Architecture principles by providing a clean abstraction layer.
 */

import { ApiError, ValidationError } from '@/types/api';
import { APP_CONFIG } from '@/config/app-config';
import type { CircuitBreakerService } from '@/domain/services/CircuitBreakerService';

export interface WailsApiOptions {
  readonly timeout?: number;
  readonly retries?: number;
  readonly validateInput?: boolean;
  readonly validateOutput?: boolean;
  readonly useCircuitBreaker?: boolean;
  readonly circuitBreakerName?: string;
}

export interface ApiCallResult<T> {
  readonly success: boolean;
  readonly data?: T;
  readonly error?: string;
  readonly metadata?: {
    readonly duration: number;
    readonly retries: number;
    readonly timestamp: number;
  };
}

/**
 * Enhanced WailsApiAdapter with automatic JSON handling
 * 
 * Features:
 * - Automatic JSON serialization/deserialization
 * - Type-safe API calls
 * - Centralized error handling
 * - Request/response validation
 * - Retry logic with exponential backoff
 * - Performance monitoring
 * - Circuit breaker pattern integration
 */
export class WailsApiAdapter {
  private readonly defaultOptions: WailsApiOptions = {
    timeout: APP_CONFIG.events.API_TIMEOUT_MS, // 30 seconds
    retries: APP_CONFIG.events.ERROR_RETRY_ATTEMPTS,
    validateInput: true,
    validateOutput: true,
    useCircuitBreaker: true
  };
  
  private circuitBreakerService?: CircuitBreakerService;

  constructor(circuitBreakerService?: CircuitBreakerService) {
    this.circuitBreakerService = circuitBreakerService;
  }

  /**
   * Makes a type-safe API call with automatic JSON handling
   */
  async callApi<TInput, TOutput>(
    method: (...args: any[]) => Promise<any>,
    input?: TInput,
    options?: WailsApiOptions
  ): Promise<TOutput> {
    // Check if we're in a Wails environment
    const isWailsEnvironment = typeof window !== 'undefined' && 
                              typeof (window as any).runtime !== 'undefined';
    
    if (!isWailsEnvironment) {
      console.warn('⚠️  Not in Wails environment, returning mock response for:', method.name);
      // Return a mock response for development/testing
      return Promise.resolve({} as TOutput);
    }
    
    const config = { ...this.defaultOptions, ...options };
    const circuitName = config.circuitBreakerName || method.name || 'unknown_api';
    
    // Use circuit breaker if enabled and available
    if (config.useCircuitBreaker && this.circuitBreakerService) {
      return this.circuitBreakerService.executeWithRetry(
        circuitName,
        () => this.executeApiCall<TInput, TOutput>(method, input, config),
        {
          maxRetries: config.retries || 0,
          baseDelayMs: 1000,
          exponentialBackoff: true,
          jitter: true
        },
        {
          timeoutMs: config.timeout,
          retryCount: config.retries
        }
      );
    } else {
      // Fallback to direct execution without circuit breaker
      return this.executeApiCall<TInput, TOutput>(method, input, config);
    }
  }
  
  /**
   * Internal method to execute API call with retry logic
   */
  private async executeApiCall<TInput, TOutput>(
    method: (...args: any[]) => Promise<any>,
    input?: TInput,
    options?: WailsApiOptions
  ): Promise<TOutput> {
    const config = { ...this.defaultOptions, ...options };
    const startTime = performance.now();
    let retryCount = 0;

    while (retryCount <= config.retries!) {
      try {
        // Validate input if enabled
        if (config.validateInput && input !== undefined) {
          this.validateInput(input);
        }

        // Serialize input to JSON if it's an object
        const serializedInput = this.serializeInput(input);
        
        // Make the API call with timeout
        const rawResult = await this.callWithTimeout(method, serializedInput, config.timeout!);
        
        // Deserialize output from JSON if it's a string
        const deserializedOutput = this.deserializeOutput<TOutput>(rawResult);
        
        // Validate output if enabled
        if (config.validateOutput) {
          this.validateOutput(deserializedOutput);
        }

        return deserializedOutput;
      } catch (error) {
        retryCount++;
        
        if (retryCount > config.retries!) {
          const duration = performance.now() - startTime;
          throw this.enhanceError(error, {
            method: method.name,
            input,
            duration,
            retries: retryCount - 1
          });
        }
        
        // Exponential backoff before retry (only if not using circuit breaker)
        if (!config.useCircuitBreaker) {
          await this.delay(Math.pow(2, retryCount) * 1000);
        }
      }
    }

    throw new ApiError('Max retries exceeded');
  }

  /**
   * Makes a batch API call with automatic JSON handling
   */
  async callApiBatch<TInput, TOutput>(
    calls: Array<{
      method: (...args: any[]) => Promise<any>;
      input?: TInput;
      options?: WailsApiOptions;
    }>
  ): Promise<ApiCallResult<TOutput>[]> {
    const promises = calls.map(async (call, index) => {
      try {
        const data = await this.callApi<TInput, TOutput>(call.method, call.input, call.options);
        return {
          success: true,
          data,
          metadata: {
            duration: 0, // Would be calculated per call
            retries: 0,
            timestamp: Date.now()
          }
        };
      } catch (error) {
        return {
          success: false,
          error: error instanceof Error ? error.message : String(error),
          metadata: {
            duration: 0,
            retries: 0,
            timestamp: Date.now()
          }
        };
      }
    });

    return Promise.all(promises);
  }

  /**
   * Makes a streaming API call with automatic JSON handling
   */
  async callApiStreaming<TInput, TOutput>(
    method: (...args: any[]) => Promise<any>,
    input?: TInput,
    onData?: (chunk: TOutput) => void,
    options?: WailsApiOptions
  ): Promise<void> {
    const config = { ...this.defaultOptions, ...options };
    
    try {
      const serializedInput = this.serializeInput(input);
      const rawResult = await this.callWithTimeout(method, serializedInput, config.timeout!);
      
      // For streaming, we expect the result to be a string that can be parsed as chunks
      if (typeof rawResult === 'string') {
        const chunks = rawResult.split('\n').filter(line => line.trim());
        
        for (const chunk of chunks) {
          try {
            const deserializedChunk = this.deserializeOutput<TOutput>(chunk);
            onData?.(deserializedChunk);
          } catch (parseError) {
            console.warn('Failed to parse streaming chunk:', chunk, parseError);
          }
        }
      } else {
        // Single result, treat as one chunk
        const deserializedOutput = this.deserializeOutput<TOutput>(rawResult);
        onData?.(deserializedOutput);
      }
    } catch (error) {
      throw this.enhanceError(error, {
        method: method.name,
        input,
        streaming: true
      });
    }
  }

  /**
   * Serializes input to JSON string if needed
   */
  private serializeInput<T>(input?: T): any {
    if (input === undefined || input === null) {
      return input;
    }

    // If input is already a primitive type, return as-is
    if (typeof input === 'string' || typeof input === 'number' || typeof input === 'boolean') {
      return input;
    }

    // If input is an array of primitives, pass directly
    if (Array.isArray(input) && input.every(item => 
      typeof item === 'string' || typeof item === 'number' || typeof item === 'boolean'
    )) {
      return input;
    }

    // For complex objects, serialize to JSON
    try {
      return JSON.stringify(input);
    } catch (error) {
      throw new ValidationError(`Failed to serialize input: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Deserializes output from JSON string if needed
   */
  private deserializeOutput<T>(output: any): T {
    // If output is already an object, return as-is
    if (output !== null && typeof output === 'object') {
      return output as T;
    }

    // If output is a primitive type that's not a string, return as-is
    if (typeof output !== 'string') {
      return output as T;
    }

    // Try to parse JSON string
    try {
      return JSON.parse(output) as T;
    } catch (parseError) {
      // If JSON parsing fails, but it's a valid string response, return as-is
      if (typeof output === 'string') {
        return output as T;
      }
      
      throw new ValidationError(
        `Failed to deserialize output: ${parseError instanceof Error ? parseError.message : String(parseError)}`
      );
    }
  }

  /**
   * Validates input data
   */
  private validateInput<T>(input: T): void {
    if (input === null || input === undefined) {
      return; // Allow null/undefined inputs
    }

    // Check for circular references
    try {
      JSON.stringify(input);
    } catch (error) {
      if (error instanceof Error && error.message.includes('circular')) {
        throw new ValidationError('Input contains circular references');
      }
      throw new ValidationError(`Input validation failed: ${error instanceof Error ? error.message : String(error)}`);
    }

    // Check input size limits
    const inputString = typeof input === 'string' ? input : JSON.stringify(input);
    const maxSize = APP_CONFIG.reports.export.MAX_EXPORT_SIZE; // Reuse this limit
    
    if (inputString.length > maxSize) {
      throw new ValidationError(`Input size (${inputString.length} bytes) exceeds maximum allowed (${maxSize} bytes)`);
    }
  }

  /**
   * Validates output data
   */
  private validateOutput<T>(output: T): void {
    if (output === null || output === undefined) {
      return; // Allow null/undefined outputs
    }

    // For security, check for potentially dangerous content
    if (typeof output === 'string') {
      // Check for XSS patterns
      const dangerousPatterns = [
        /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi,
        /javascript:/gi,
        /on\w+\s*=/gi
      ];

      for (const pattern of dangerousPatterns) {
        if (pattern.test(output)) {
          console.warn('Potentially dangerous content detected in API response');
          break;
        }
      }
    }

    // Check output size
    const outputString = typeof output === 'string' ? output : JSON.stringify(output);
    const maxSize = APP_CONFIG.reports.export.MAX_EXPORT_SIZE;
    
    if (outputString.length > maxSize) {
      console.warn(`Large API response detected: ${outputString.length} bytes`);
    }
  }

  /**
   * Calls method with timeout
   */
  private async callWithTimeout(
    method: (...args: any[]) => Promise<any>,
    input: any,
    timeout: number
  ): Promise<any> {
    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        reject(new ApiError(`API call timed out after ${timeout}ms`));
      }, timeout);

      // Handle different input patterns
      let promise: Promise<any>;
      
      if (input === undefined) {
        promise = method();
      } else if (Array.isArray(input)) {
        promise = method(...input);
      } else {
        promise = method(input);
      }

      promise
        .then(result => {
          clearTimeout(timeoutId);
          resolve(result);
        })
        .catch(error => {
          clearTimeout(timeoutId);
          reject(error);
        });
    });
  }

  /**
   * Enhances errors with additional context
   */
  private enhanceError(error: unknown, context: {
    method?: string;
    input?: any;
    duration?: number;
    retries?: number;
    streaming?: boolean;
  }): Error {
    let enhancedError: Error;

    if (error instanceof ApiError || error instanceof ValidationError) {
      enhancedError = error;
    } else if (error instanceof Error) {
      enhancedError = new ApiError(error.message);
    } else {
      enhancedError = new ApiError(String(error));
    }

    // Add context information
    (enhancedError as any).context = {
      ...context,
      timestamp: new Date().toISOString(),
      userAgent: navigator.userAgent,
      url: window.location.href
    };

    // Log error for debugging
    console.error('Enhanced API Error:', {
      error: enhancedError.message,
      context: (enhancedError as any).context
    });

    return enhancedError;
  }

  /**
   * Delay utility for retry logic
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Utility method to check if a value needs JSON serialization
   */
  private needsSerialization(value: any): boolean {
    return value !== null &&
           value !== undefined &&
           typeof value === 'object' &&
           !Array.isArray(value);
  }

  /**
   * Utility method to check if a value is JSON string
   */
  private isJsonString(value: any): boolean {
    if (typeof value !== 'string') return false;
    
    try {
      const parsed = JSON.parse(value);
      return typeof parsed === 'object' && parsed !== null;
    } catch {
      return false;
    }
  }

  /**
   * Gets API adapter statistics
   */
  getStatistics(): {
    totalCalls: number;
    failedCalls: number;
    averageResponseTime: number;
  } {
    // This would be implemented with actual tracking
    return {
      totalCalls: 0,
      failedCalls: 0,
      averageResponseTime: 0
    };
  }

  /**
   * Clears API adapter cache and statistics
   */
  clearCache(): void {
    // This would clear any internal caching
    console.log('API adapter cache cleared');
  }
}

/**
 * Default WailsApiAdapter instance
 */
export const defaultWailsApiAdapter = new WailsApiAdapter();

/**
 * Factory function to create WailsApiAdapter with circuit breaker
 */
export const createWailsApiAdapter = (circuitBreakerService?: CircuitBreakerService) => 
  new WailsApiAdapter(circuitBreakerService);