import type { StorageOptions } from '@/domain/services/LocalStorageService';

/**
 * Repository interface for storage operations
 * Provides abstraction over localStorage and sessionStorage implementations
 */
export interface StorageRepository {
  /**
   * Set item in storage
   * @param key Storage key
   * @param value Value to store
   * @param options Storage options
   * @returns Boolean indicating success
   */
  set<T>(key: string, value: T, options?: Partial<StorageOptions>): boolean;

  /**
   * Get item from storage
   * @param key Storage key
   * @param defaultValue Default value if key doesn't exist
   * @returns Stored value or default
   */
  get<T>(key: string, defaultValue?: T): T | null;

  /**
   * Set item in session storage
   * @param key Storage key
   * @param value Value to store
   * @returns Boolean indicating success
   */
  setSession<T>(key: string, value: T): boolean;

  /**
   * Get item from session storage
   * @param key Storage key
   * @param defaultValue Default value if key doesn't exist
   * @returns Stored value or default
   */
  getSession<T>(key: string, defaultValue?: T): T | null;

  /**
   * Remove item from storage
   * @param key Storage key
   * @returns Boolean indicating success
   */
  remove(key: string): boolean;

  /**
   * Remove item from session storage
   * @param key Storage key
   * @returns Boolean indicating success
   */
  removeSession(key: string): boolean;

  /**
   * Clear all items with current prefix
   * @returns Boolean indicating success
   */
  clear(): boolean;

  /**
   * Check if an item exists
   * @param key Storage key
   * @returns Boolean indicating existence
   */
  has(key: string): boolean;

  /**
   * Get all keys with current prefix
   * @returns Array of keys
   */
  getKeys(): string[];

  /**
   * Get multiple items at once
   * @param keys Array of keys
   * @returns Record of key-value pairs
   */
  getMultiple<T>(keys: string[]): Record<string, T | null>;

  /**
   * Set multiple items at once
   * @param items Record of key-value pairs
   * @param options Storage options
   * @returns Boolean indicating success
   */
  setMultiple<T>(items: Record<string, T>, options?: Partial<StorageOptions>): boolean;
}