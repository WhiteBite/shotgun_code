/**
 * Memory Monitor Utility
 * 
 * This utility helps prevent Out Of Memory (OOM) errors by:
 * 1. Monitoring memory usage
 * 2. Providing early warnings
 * 3. Cleaning up large objects when memory usage is high
 * 4. Forcing garbage collection when possible
 */

import { APP_CONFIG } from '@/config/app-config';
import type { NotificationPort } from '@/application/ports';

export interface MemoryStats {
  used: number;        // Used memory in MB
  total: number;       // Total available memory in MB
  percentage: number;  // Usage percentage (0-100)
}

export interface MemoryWarningOptions {
  warningThreshold?: number;   // Absolute MB threshold for warning (default: 15MB)
  criticalThreshold?: number;  // Absolute MB threshold for critical warning (default: 25MB)
  pollingInterval?: number;    // How often to check memory in ms (default: 1000ms)
  showToasts?: boolean;        // Whether to show toast notifications (default: true)
  autoCleanup?: boolean;       // Whether to auto cleanup when critical (default: true)
}

const DEFAULT_OPTIONS: MemoryWarningOptions = {
  warningThreshold: 15,  // 15MB
  criticalThreshold: 25,  // 25MB
  pollingInterval: 1000, // 1 second
  showToasts: true,
  autoCleanup: true
};

export class MemoryMonitor {
  private static instance: MemoryMonitor;
  private options: MemoryWarningOptions;
  private monitorInterval: number | null = null;
  private warningIssued = false;
  private criticalIssued = false;
  private notificationPort: NotificationPort;

  private constructor(notificationPort: NotificationPort, options: MemoryWarningOptions = {}) {
    this.notificationPort = notificationPort;
    this.options = { ...DEFAULT_OPTIONS, ...options };
  }

  /**
   * Get the singleton instance of MemoryMonitor
   */
  public static getInstance(notificationPort: NotificationPort, options?: MemoryWarningOptions): MemoryMonitor {
    if (!MemoryMonitor.instance) {
      MemoryMonitor.instance = new MemoryMonitor(notificationPort, options);
    } else if (options) {
      MemoryMonitor.instance.updateOptions(options);
    }
    return MemoryMonitor.instance;
  }

  /**
   * Update monitoring options
   */
  public updateOptions(options: Partial<MemoryWarningOptions>): void {
    this.options = { ...this.options, ...options };
    
    // If monitoring is active, restart it with new options
    if (this.monitorInterval) {
      this.stopMonitoring();
      this.startMonitoring();
    }
  }

  /**
   * Get current memory stats
   */
  public getMemoryStats(): MemoryStats | null {
    if (!('performance' in window) || !('memory' in (performance as unknown as { memory?: unknown }))) {
      return null;
    }

    const memory = (performance as unknown as { memory: { usedJSHeapSize: number; jsHeapSizeLimit: number } }).memory;
    const used = Math.round(memory.usedJSHeapSize / (1024 * 1024));
    const total = Math.round(memory.jsHeapSizeLimit / (1024 * 1024));
    const percentage = Math.round((memory.usedJSHeapSize / memory.jsHeapSizeLimit) * 100);
    
    return { used, total, percentage };
  }

  /**
   * Start memory monitoring
   */
  public startMonitoring(): void {
    if (this.monitorInterval) {
      return; // Already monitoring
    }

    this.checkMemory();
    this.monitorInterval = window.setInterval(
      () => this.checkMemory(), 
      this.options.pollingInterval
    );
    
    console.log(`Memory monitoring started (interval: ${this.options.pollingInterval}ms)`);
  }

  /**
   * Stop memory monitoring
   */
  public stopMonitoring(): void {
    if (this.monitorInterval) {
      clearInterval(this.monitorInterval);
      this.monitorInterval = null;
      console.log('Memory monitoring stopped');
    }
  }

  /**
   * Force cleanup of memory with AGGRESSIVE strategies
   */
  public forceCleanup(): void {
    console.log('EMERGENCY: Forcing aggressive memory cleanup...');
    
    // Attempt to free up memory by clearing large objects in memory
    try {
      // Clear large arrays and objects
      (window as unknown as { largeObjects?: unknown[] }).largeObjects = [];
      (window as unknown as { cachedData?: unknown }).cachedData = null;
      
      // Clear Vue reactive caches if possible
      try {
        // Force cleanup of any global stores
        const stores = ['contextBuilder', 'fileTree', 'treeState', 'ui'];
        stores.forEach(storeName => {
          const storeData = (window as unknown as Record<string, { cleanup?: () => void }>)[storeName];
          if (storeData && typeof storeData.cleanup === 'function') {
            storeData.cleanup();
          }
        });
      } catch (e) {
        console.warn('Could not cleanup Vue stores:', e);
      }
      
      // Force garbage collection multiple times if available
      if (window.gc) {
        try {
          window.gc();
          setTimeout(() => window.gc?.(), 100);
          setTimeout(() => window.gc?.(), 300);
          console.log('Multiple garbage collection cycles triggered');
        } catch (e) {
          console.warn('Failed to trigger garbage collection', e);
        }
      }
      
      // Add a small delay to allow garbage collection to work
      setTimeout(() => {
        const stats = this.getMemoryStats();
        if (stats) {
          console.log(`Memory after aggressive cleanup: ${stats.used}MB / ${stats.total}MB (${stats.percentage}%)`);
        }
      }, 500);
    } catch (e) {
      console.error('Error during emergency memory cleanup:', e);
    }
  }

  /**
   * Check current memory usage with ABSOLUTE thresholds in MB
   */
  private checkMemory(): void {
    const stats = this.getMemoryStats();
    if (!stats) return;
    
    const { used, total, percentage } = stats;
    
    // Log memory usage to console
    console.debug(`Memory usage: ${used}MB / ${total}MB (${percentage}%)`);
    
    // CRITICAL: Use absolute MB values instead of percentages
    const warningThresholdMB = this.options.warningThreshold || 15;
    const criticalThresholdMB = this.options.criticalThreshold || 25;
    
    // Check if we've reached critical threshold (absolute MB)
    if (used >= criticalThresholdMB) {
      if (!this.criticalIssued) {
        console.error(`CRITICAL: Memory usage at ${used}MB >= ${criticalThresholdMB}MB`);
        
        if (this.options.showToasts) {
          this.notificationPort.showError(
            `CRITICAL memory usage (${used}MB). Browser may crash! Cleaning up...`
          );
        }
        
        // Force cleanup immediately
        if (this.options.autoCleanup) {
          this.forceCleanup();
          // Emergency additional cleanup
          setTimeout(() => this.forceCleanup(), 500);
        }
        
        this.criticalIssued = true;
      }
    } 
    // Check if we've reached warning threshold (absolute MB)
    else if (used >= warningThresholdMB) {
      if (!this.warningIssued) {
        console.warn(`WARNING: Memory usage at ${used}MB >= ${warningThresholdMB}MB`);
        
        if (this.options.showToasts) {
          this.notificationPort.showWarning(
            `High memory usage (${used}MB). Reduce file selection or context size.`
          );
        }
        
        this.warningIssued = true;
      }
    } 
    // Reset flags if memory usage drops below thresholds
    else {
      // Only reset if we drop significantly below thresholds
      if (used < warningThresholdMB - 2) {
        this.warningIssued = false;
        this.criticalIssued = false;
      }
    }
  }
}

// Add window.gc interface
declare global {
  interface Window {
    gc?: () => void;
  }
}

// Convenience function to get the monitor instance
export function useMemoryMonitor(notificationPort: NotificationPort, options?: MemoryWarningOptions): MemoryMonitor {
  return MemoryMonitor.getInstance(notificationPort, options);
}

export default MemoryMonitor;