import { reactive } from 'vue';

export interface ErrorContext {
  operation: string;
  component?: string;
  retry?: () => Promise<void>;
  fallback?: unknown;
}

export interface ErrorClassification {
  type: 'network' | 'validation' | 'authorization' | 'business' | 'system';
  severity: 'low' | 'medium' | 'high' | 'critical';
  recoverable: boolean;
  userMessage: string;
  technicalMessage: string;
}

interface ErrorInfo {
  id: string;
  timestamp: Date;
  context: ErrorContext;
  classification: ErrorClassification;
  resolved: boolean;
  retryCount: number;
}

interface RetryableError extends ErrorInfo {
  nextRetryAt: Date;
}

interface ErrorState {
  activeErrors: Map<string, ErrorInfo>;
  errorHistory: ErrorInfo[];
  retryQueue: RetryableError[];
}

interface RetryConfig {
  maxAttempts: number;
  baseDelay: number;
  maxDelay: number;
  backoffMultiplier: number;
  retryableErrors: string[];
}

const defaultRetryConfig: RetryConfig = {
  maxAttempts: 3,
  baseDelay: 1000,
  maxDelay: 30000,
  backoffMultiplier: 2,
  retryableErrors: ['NETWORK_ERROR', 'RATE_LIMIT_EXCEEDED', 'INTERNAL_ERROR']
};

export function useAdvancedErrorHandler() {
  const errorState = reactive<ErrorState>({
    activeErrors: new Map(),
    errorHistory: [],
    retryQueue: []
  });

  const classifyError = (error: unknown, _context: ErrorContext): ErrorClassification => {
    const errorMessage = error instanceof Error ? error.message : String(error);
    const errorCode = (error as unknown)?.code || '';
    
    // Network and connectivity errors
    if (errorCode === 'NETWORK_ERROR' || errorMessage.includes('fetch') || errorMessage.includes('connection')) {
      return {
        type: 'network',
        severity: 'medium',
        recoverable: true,
        userMessage: 'Connection issue. Please check your internet connection and try again.',
        technicalMessage: errorMessage
      };
    }
    
    // Validation errors
    if (errorCode === 'VALIDATION_ERROR' || errorMessage.includes('validation') || errorMessage.includes('invalid')) {
      return {
        type: 'validation',
        severity: 'low',
        recoverable: false,
        userMessage: 'Please check your input and try again.',
        technicalMessage: errorMessage
      };
    }
    
    // Authorization errors
    if (errorCode === 'UNAUTHORIZED' || errorCode === 'INVALID_API_KEY' || errorMessage.includes('unauthorized')) {
      return {
        type: 'authorization',
        severity: 'high',
        recoverable: false,
        userMessage: 'Access denied. Please check your credentials.',
        technicalMessage: errorMessage
      };
    }
    
    // Rate limiting
    if (errorCode === 'RATE_LIMIT_EXCEEDED') {
      return {
        type: 'network',
        severity: 'medium',
        recoverable: true,
        userMessage: 'Too many requests. Please wait a moment and try again.',
        technicalMessage: errorMessage
      };
    }
    
    // Business logic errors
    if (errorCode === 'TASK_NOT_FOUND' || errorCode === 'INVALID_TASK_STATE') {
      return {
        type: 'business',
        severity: 'medium',
        recoverable: false,
        userMessage: 'Operation cannot be completed. Please refresh and try again.',
        technicalMessage: errorMessage
      };
    }
    
    // System errors
    if (errorCode === 'INTERNAL_ERROR' || errorCode === 'CONFIGURATION_ERROR') {
      return {
        type: 'system',
        severity: 'critical',
        recoverable: true,
        userMessage: 'A system error occurred. Our team has been notified.',
        technicalMessage: errorMessage
      };
    }
    
    // Default classification
    return {
      type: 'system',
      severity: 'medium',
      recoverable: false,
      userMessage: 'An unexpected error occurred. Please try again.',
      technicalMessage: errorMessage
    };
  };

  const generateErrorId = (): string => {
    return Math.random().toString(36).substring(2, 15);
  };

  const createErrorInfo = (error: unknown, _context: ErrorContext): ErrorInfo => {
    return {
      id: generateErrorId(),
      timestamp: new Date(),
      context: _context,
      classification: classifyError(error, _context),
      resolved: false,
      retryCount: 0
    };
  };

  const isRetryable = (error: Error, config: RetryConfig): boolean => {
    const errorCode = (error as unknown)?.code || '';
    return config.retryableErrors.some(pattern => 
      errorCode === pattern || error.message.includes(pattern)
    );
  };

  const calculateDelay = (attempt: number, config: RetryConfig): number => {
    const delay = config.baseDelay * Math.pow(config.backoffMultiplier, attempt - 1);
    return Math.min(delay, config.maxDelay);
  };

  const delay = (ms: number): Promise<void> => {
    return new Promise(resolve => setTimeout(resolve, ms));
  };

  const executeWithRetry = async <T>(
    operation: () => Promise<T>,
    _context: ErrorContext,
    customConfig?: Partial<RetryConfig>
  ): Promise<T> => {
    const config = { ...defaultRetryConfig, ...customConfig };
    let lastError: Error;

    for (let attempt = 1; attempt <= config.maxAttempts; attempt++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error as Error;

        if (!isRetryable(lastError, config)) {
          throw lastError;
        }

        if (attempt === config.maxAttempts) {
          break;
        }

        const delayMs = calculateDelay(attempt, config);
        console.log(`Retry attempt ${attempt} failed, retrying in ${delayMs}ms`);
        await delay(delayMs);
      }
    }

    throw lastError!;
  };

  const handleStructuredError = (error: unknown, context: ErrorContext) => {
    const errorInfo = createErrorInfo(error, context);
    
    // Store the error
    errorState.activeErrors.set(errorInfo.id, errorInfo);
    errorState.errorHistory.push(errorInfo);
    
    // Auto-retry logic for recoverable errors
    if (errorInfo.classification.recoverable && errorInfo.retryCount < 3 && context.retry) {
      scheduleRetry(errorInfo);
    }
    
    // Route to appropriate handlers based on classification
    switch (errorInfo.classification.severity) {
      case 'critical':
        handleCriticalError(error, context, errorInfo.classification);
        break;
      case 'high':
        handleHighSeverityError(error, context, errorInfo.classification);
        break;
      default:
        handleStandardError(error, context, errorInfo.classification);
    }
    
    return errorInfo;
  };

  const scheduleRetry = (errorInfo: ErrorInfo) => {
    if (!errorInfo.context.retry) return;
    
    const retryDelay = calculateDelay(errorInfo.retryCount + 1, defaultRetryConfig);
    const retryableError: RetryableError = {
      ...errorInfo,
      nextRetryAt: new Date(Date.now() + retryDelay)
    };
    
    errorState.retryQueue.push(retryableError);
    
    setTimeout(async () => {
      try {
        if (errorInfo.context.retry) {
          await errorInfo.context.retry();
          resolveError(errorInfo.id);
        }
      } catch (_retryError) {
        // Using the error for logging even if we don't do anything else with it
        console.error('Retry failed:', _retryError);
        errorInfo.retryCount++;
        if (errorInfo.retryCount < 3) {
          scheduleRetry(errorInfo);
        }
      }
    }, retryDelay);
  };

  const handleCriticalError = (_error: unknown, _context: ErrorContext, _classification: ErrorClassification) => {
    console.error('Critical error occurred:', _error);
    // Could trigger system-wide error state, maintenance mode, etc.
  };

  const handleHighSeverityError = (_error: unknown, _context: ErrorContext, _classification: ErrorClassification) => {
    console.error('High severity error:', _error);
    // Could trigger notifications to support team
  };

  const handleStandardError = (_error: unknown, _context: ErrorContext, _classification: ErrorClassification) => {
    console.warn('Standard error:', _error);
    // Standard logging and user notification
  };

  const resolveError = (errorId: string) => {
    const error = errorState.activeErrors.get(errorId);
    if (error) {
      error.resolved = true;
      errorState.activeErrors.delete(errorId);
    }
  };

  const clearErrors = () => {
    errorState.activeErrors.clear();
    errorState.retryQueue.length = 0;
  };

  const getActiveErrors = () => {
    return Array.from(errorState.activeErrors.values());
  };

  const getErrorHistory = () => {
    return errorState.errorHistory;
  };

  return {
    handleStructuredError,
    classifyError,
    executeWithRetry,
    resolveError,
    clearErrors,
    getActiveErrors,
    getErrorHistory,
    errorState
  };
}