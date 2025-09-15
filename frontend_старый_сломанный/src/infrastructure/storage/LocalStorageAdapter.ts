/**
 * LocalStorage Service
 * 
 * Infrastructure service that centralizes localStorage operations.
 * Provides type-safe, error-handled storage operations with automatic data transformation.
 */

import { APP_CONFIG } from '@/config/app-config';
import type { StorageRepository } from '@/domain/repositories/StorageRepository';

export interface StorageOptions {
  prefix?: string;
  encrypt?: boolean;
  expiryMs?: number;
  compress?: boolean;
}

export interface StorageItem<T> {
  data: T;
  timestamp: number;
  expiryMs?: number;
  version: string;
}

export interface CacheMetadata {
  key: string;
  timestamp: number;
  size: number;
  accessed: number;
}

/**
 * LocalStorage adapter that implements StorageRepository interface
 * Provides concrete implementation of storage operations using browser localStorage
 */
export class LocalStorageAdapter implements StorageRepository {
  private prefix: string;
  private isAvailable: boolean;

  constructor(options: StorageOptions = {}) {
    this.prefix = options.prefix || APP_CONFIG.storage.UI_PREFERENCES_PREFIX;
    this.isAvailable = this.checkAvailability();
  }

  /**
   * Check if localStorage is available
   */
  private checkAvailability(): boolean {
    try {
      const test = '__localStorage_test__';
      localStorage.setItem(test, 'test');
      localStorage.removeItem(test);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Generate storage key with prefix
   */
  private generateKey(key: string, isSession = false): string {
    const prefix = isSession 
      ? APP_CONFIG.storage.SESSION_STORAGE_PREFIX 
      : this.prefix;
    return `${prefix}${key}`;
  }

  /**
   * Set item in localStorage with metadata
   */
  set<T>(key: string, value: T, options: Partial<StorageOptions> = {}): boolean {
    if (!this.isAvailable) {
      console.warn('LocalStorage not available');
      return false;
    }

    try {
      const storageKey = this.generateKey(key, false);
      const item: StorageItem<T> = {
        data: value,
        timestamp: Date.now(),
        expiryMs: options.expiryMs || APP_CONFIG.storage.CACHE_EXPIRY,
        version: '1.0'
      };

      const serialized = JSON.stringify(item);
      localStorage.setItem(storageKey, serialized);
      
      this.updateCacheMetadata(storageKey, serialized.length);
      return true;
    } catch (error) {
      console.error('Error setting localStorage item:', error);
      return false;
    }
  }

  /**
   * Get item from localStorage with automatic expiry check
   */
  get<T>(key: string, defaultValue?: T): T | null {
    if (!this.isAvailable) {
      return defaultValue || null;
    }

    try {
      const storageKey = this.generateKey(key, false);
      const serialized = localStorage.getItem(storageKey);
      
      if (!serialized) {
        return defaultValue || null;
      }

      const item: StorageItem<T> = JSON.parse(serialized);
      
      // Check expiry
      if (item.expiryMs && Date.now() - item.timestamp > item.expiryMs) {
        this.remove(key);
        return defaultValue || null;
      }

      // Update access time for cache management
      this.updateAccessTime(storageKey);
      return item.data;
    } catch (error) {
      console.error('Error getting localStorage item:', error);
      return defaultValue || null;
    }
  }

  /**
   * Set item in sessionStorage
   */
  setSession<T>(key: string, value: T): boolean {
    if (!this.isAvailable) {
      return false;
    }

    try {
      const storageKey = this.generateKey(key, true);
      const item: StorageItem<T> = {
        data: value,
        timestamp: Date.now(),
        expiryMs: APP_CONFIG.storage.CACHE_EXPIRY,
        version: '1.0'
      };

      sessionStorage.setItem(storageKey, JSON.stringify(item));
      return true;
    } catch (error) {
      console.error('Error setting sessionStorage item:', error);
      return false;
    }
  }

  /**
   * Get item from sessionStorage
   */
  getSession<T>(key: string, defaultValue?: T): T | null {
    if (!this.isAvailable) {
      return defaultValue || null;
    }

    try {
      const storageKey = this.generateKey(key, true);
      const serialized = sessionStorage.getItem(storageKey);
      
      if (!serialized) {
        return defaultValue || null;
      }

      const item: StorageItem<T> = JSON.parse(serialized);
      return item.data;
    } catch (error) {
      console.error('Error getting sessionStorage item:', error);
      return defaultValue || null;
    }
  }

  /**
   * Remove item from localStorage
   */
  remove(key: string): boolean {
    if (!this.isAvailable) {
      return false;
    }

    try {
      const storageKey = this.generateKey(key, false);
      localStorage.removeItem(storageKey);
      this.removeCacheMetadata(storageKey);
      return true;
    } catch (error) {
      console.error('Error removing localStorage item:', error);
      return false;
    }
  }

  /**
   * Remove item from sessionStorage
   */
  removeSession(key: string): boolean {
    if (!this.isAvailable) {
      return false;
    }

    try {
      const storageKey = this.generateKey(key, true);
      sessionStorage.removeItem(storageKey);
      return true;
    } catch (error) {
      console.error('Error removing sessionStorage item:', error);
      return false;
    }
  }

  /**
   * Clear all items with current prefix
   */
  clear(): boolean {
    if (!this.isAvailable) {
      return false;
    }

    try {
      const keys = Object.keys(localStorage);
      const prefixedKeys = keys.filter(key => key.startsWith(this.prefix));
      
      prefixedKeys.forEach(key => {
        localStorage.removeItem(key);
      });

      return true;
    } catch (error) {
      console.error('Error clearing localStorage:', error);
      return false;
    }
  }

  /**
   * Get all keys with current prefix
   */
  getKeys(): string[] {
    if (!this.isAvailable) {
      return [];
    }

    try {
      const keys = Object.keys(localStorage);
      return keys
        .filter(key => key.startsWith(this.prefix))
        .map(key => key.replace(this.prefix, ''));
    } catch (error) {
      console.error('Error getting localStorage keys:', error);
      return [];
    }
  }

  /**
   * Get storage usage statistics
   */
  getUsageStats(): { totalSize: number; itemCount: number; cacheMetadata: CacheMetadata[] } {
    if (!this.isAvailable) {
      return { totalSize: 0, itemCount: 0, cacheMetadata: [] };
    }

    try {
      const keys = Object.keys(localStorage);
      const prefixedKeys = keys.filter(key => key.startsWith(this.prefix));
      
      let totalSize = 0;
      const cacheMetadata: CacheMetadata[] = [];

      prefixedKeys.forEach(key => {
        const value = localStorage.getItem(key);
        if (value) {
          const size = value.length;
          totalSize += size;
          
          // Try to get metadata if it exists
          const metadata = this.getCacheMetadata(key);
          if (metadata) {
            cacheMetadata.push(metadata);
          }
        }
      });

      return {
        totalSize,
        itemCount: prefixedKeys.length,
        cacheMetadata
      };
    } catch (error) {
      console.error('Error getting storage usage stats:', error);
      return { totalSize: 0, itemCount: 0, cacheMetadata: [] };
    }
  }

  /**
   * Cleanup expired items and large cache
   */
  cleanup(): { removedCount: number; freedSpace: number } {
    if (!this.isAvailable) {
      return { removedCount: 0, freedSpace: 0 };
    }

    let removedCount = 0;
    let freedSpace = 0;

    try {
      const keys = Object.keys(localStorage);
      const prefixedKeys = keys.filter(key => key.startsWith(this.prefix));
      
      const now = Date.now();
      const cacheExpiry = APP_CONFIG.storage.CACHE_EXPIRY;
      const maxCacheSize = APP_CONFIG.performance.limits.MAX_CACHE_SIZE;

      prefixedKeys.forEach(key => {
        const value = localStorage.getItem(key);
        if (value) {
          try {
            const item: StorageItem<any> = JSON.parse(value);
            
            // Check for expiry
            const expired = item.expiryMs 
              ? (now - item.timestamp > item.expiryMs)
              : (now - item.timestamp > cacheExpiry);

            if (expired) {
              freedSpace += value.length;
              localStorage.removeItem(key);
              this.removeCacheMetadata(key);
              removedCount++;
            }
          } catch {
            // Invalid JSON, remove it
            freedSpace += value.length;
            localStorage.removeItem(key);
            removedCount++;
          }
        }
      });

      // Check if we need to remove items due to size limits
      if (freedSpace > maxCacheSize) {
        // Remove oldest items until we're under the limit
        const metadataList = prefixedKeys
          .map(key => {
            const meta = this.getCacheMetadata(key);
            return meta ? { ...meta, key } : null;
          })
          .filter(Boolean) as CacheMetadata[];
        
        // Sort by access time (oldest first)
        metadataList.sort((a, b) => a.accessed - b.accessed);
        
        // Remove items until we're under the limit
        let currentSize = freedSpace;
        for (const meta of metadataList) {
          if (currentSize <= maxCacheSize) break;
          
          const value = localStorage.getItem(meta.key);
          if (value) {
            freedSpace += value.length;
            localStorage.removeItem(meta.key);
            this.removeCacheMetadata(meta.key);
            removedCount++;
            currentSize -= value.length;
          }
        }
      }

      return { removedCount, freedSpace };
    } catch (error) {
      console.error('Error during localStorage cleanup:', error);
      return { removedCount: 0, freedSpace: 0 };
    }
  }

  /**
   * Update cache metadata for LRU management
   */
  private updateCacheMetadata(key: string, size: number): void {
    try {
      const metadata: CacheMetadata = {
        key,
        timestamp: Date.now(),
        size,
        accessed: Date.now()
      };
      
      const metaKey = `${this.prefix}__meta__${key}`;
      localStorage.setItem(metaKey, JSON.stringify(metadata));
    } catch (error) {
      // Metadata is optional, don't fail if we can't set it
      console.debug('Could not update cache metadata:', error);
    }
  }

  /**
   * Update access time for cache item
   */
  private updateAccessTime(key: string): void {
    try {
      const metaKey = `${this.prefix}__meta__${key}`;
      const metaData = localStorage.getItem(metaKey);
      
      if (metaData) {
        const metadata: CacheMetadata = JSON.parse(metaData);
        metadata.accessed = Date.now();
        localStorage.setItem(metaKey, JSON.stringify(metadata));
      }
    } catch (error) {
      // Metadata is optional, don't fail if we can't update it
      console.debug('Could not update access time:', error);
    }
  }

  /**
   * Get cache metadata for a key
   */
  private getCacheMetadata(key: string): CacheMetadata | null {
    try {
      const metaKey = `${this.prefix}__meta__${key}`;
      const metaData = localStorage.getItem(metaKey);
      
      if (metaData) {
        return JSON.parse(metaData);
      }
      return null;
    } catch (error) {
      return null;
    }
  }

  /**
   * Remove cache metadata
   */
  private removeCacheMetadata(key: string): void {
    try {
      const metaKey = `${this.prefix}__meta__${key}`;
      localStorage.removeItem(metaKey);
    } catch (error) {
      // Metadata is optional, don't fail
      console.debug('Could not remove cache metadata:', error);
    }
  }

  /**
   * Check if an item exists
   */
  has(key: string): boolean {
    if (!this.isAvailable) {
      return false;
    }

    const storageKey = this.generateKey(key, false);
    return localStorage.getItem(storageKey) !== null;
  }

  /**
   * Get multiple items at once
   */
  getMultiple<T>(keys: string[]): Record<string, T | null> {
    const result: Record<string, T | null> = {};
    
    keys.forEach(key => {
      result[key] = this.get<T>(key);
    });

    return result;
  }

  /**
   * Set multiple items at once
   */
  setMultiple<T>(items: Record<string, T>, options: Partial<StorageOptions> = {}): boolean {
    let allSuccessful = true;

    Object.entries(items).forEach(([key, value]) => {
      if (!this.set(key, value, options)) {
        allSuccessful = false;
      }
    });

    return allSuccessful;
  }
}

// Default instance for dependency injection
export const defaultLocalStorageService = new LocalStorageAdapter();