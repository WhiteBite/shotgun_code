/**
 * Use Case Result Types
 * Common types for use case implementations
 */

import type { Project } from '@/domain/entities';

/**
 * Generic use case result wrapper
 */
export interface UseCaseResult<T> {
  isSuccess: boolean;
  data?: T;
  error?: string;
}

/**
 * Project loading result
 */
export interface ProjectLoadResult {
  project: Project;
  workspace: unknown; // Simple workspace interface for now
}

/**
 * Helper function to create successful result
 */
export function createSuccessResult<T>(data: T): UseCaseResult<T> {
  return {
    isSuccess: true,
    data
  };
}

/**
 * Helper function to create error result
 */
export function createErrorResult<T>(error: string): UseCaseResult<T> {
  return {
    isSuccess: false,
    error
  };
}