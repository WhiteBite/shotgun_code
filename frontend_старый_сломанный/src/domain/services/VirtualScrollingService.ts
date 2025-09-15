import { APP_CONFIG } from '@/config/app-config';

export interface VirtualItem {
  index: number;
  top: number;
  height: number;
  visible: boolean;
}

export interface VirtualScrollConfig {
  itemHeight: number;
  containerHeight: number;
  buffer: number;
  totalItems: number;
  dynamicHeight?: boolean;
  overscan?: number;
}

export interface ScrollMetrics {
  scrollTop: number;
  scrollHeight: number;
  clientHeight: number;
  visibleStart: number;
  visibleEnd: number;
  renderStart: number;
  renderEnd: number;
}

export interface PerformanceMetrics {
  renderTime: number;
  itemCount: number;
  visibleCount: number;
  lastUpdate: number;
  renderCalls: number;
}

/**
 * Domain service for virtual scrolling calculations and optimization
 * Encapsulates business rules for efficient rendering of large lists
 * Follows DDD principles by containing virtual scrolling domain knowledge
 */
export class VirtualScrollingService {
  private readonly defaultBuffer: number;
  private readonly defaultOverscan: number;
  private readonly performanceThreshold: number;
  
  constructor() {
    this.defaultBuffer = APP_CONFIG.ui.virtualScroll.DEFAULT_BUFFER;
    this.defaultOverscan = APP_CONFIG.ui.virtualScroll.DEFAULT_OVERSCAN;
    this.performanceThreshold = APP_CONFIG.ui.virtualScroll.PERFORMANCE_THRESHOLD;
  }

  /**
   * Calculates visible items based on scroll position
   * Business rule: Render only items that are visible plus buffer
   */
  calculateVisibleItems(config: VirtualScrollConfig, scrollTop: number): VirtualItem[] {
    const { itemHeight, containerHeight, buffer = this.defaultBuffer, totalItems } = config;
    
    // Calculate visible range
    const visibleStart = Math.max(0, Math.floor(scrollTop / itemHeight) - buffer);
    const visibleCount = Math.ceil(containerHeight / itemHeight);
    const visibleEnd = Math.min(totalItems - 1, visibleStart + visibleCount + buffer * 2);
    
    const items: VirtualItem[] = [];
    
    for (let i = visibleStart; i <= visibleEnd; i++) {
      items.push({
        index: i,
        top: i * itemHeight,
        height: itemHeight,
        visible: i >= Math.floor(scrollTop / itemHeight) && 
                i <= Math.floor(scrollTop / itemHeight) + visibleCount
      });
    }
    
    return items;
  }

  /**
   * Calculates scroll metrics for performance optimization
   * Business rule: Provide detailed metrics for scroll optimization decisions
   */
  calculateScrollMetrics(config: VirtualScrollConfig, scrollTop: number): ScrollMetrics {
    const { itemHeight, containerHeight, buffer = this.defaultBuffer, totalItems } = config;
    
    const scrollHeight = totalItems * itemHeight;
    const clientHeight = containerHeight;
    
    const visibleStart = Math.floor(scrollTop / itemHeight);
    const visibleCount = Math.ceil(containerHeight / itemHeight);
    const visibleEnd = Math.min(totalItems - 1, visibleStart + visibleCount - 1);
    
    const renderStart = Math.max(0, visibleStart - buffer);
    const renderEnd = Math.min(totalItems - 1, visibleEnd + buffer);
    
    return {
      scrollTop,
      scrollHeight,
      clientHeight,
      visibleStart,
      visibleEnd,
      renderStart,
      renderEnd
    };
  }

  /**
   * Determines optimal buffer size based on performance metrics
   * Business rule: Adapt buffer size to maintain smooth scrolling
   */
  optimizeBufferSize(metrics: PerformanceMetrics, currentBuffer: number): number {
    const { renderTime, renderCalls, lastUpdate } = metrics;
    
    // If rendering is taking too long, reduce buffer
    if (renderTime > this.performanceThreshold) {
      return Math.max(1, Math.floor(currentBuffer * 0.8));
    }
    
    // If rendering is fast and we have frequent updates, increase buffer
    if (renderTime < this.performanceThreshold / 2 && renderCalls > 10) {
      return Math.min(20, Math.floor(currentBuffer * 1.2));
    }
    
    return currentBuffer;
  }

  /**
   * Calculates scroll position to bring item into view
   * Business rule: Smooth scrolling with configurable alignment
   */
  calculateScrollToItem(
    itemIndex: number, 
    config: VirtualScrollConfig, 
    scrollTop: number,
    alignment: 'start' | 'center' | 'end' | 'auto' = 'auto'
  ): number {
    const { itemHeight, containerHeight, totalItems } = config;
    
    if (itemIndex < 0 || itemIndex >= totalItems) {
      return scrollTop; // Invalid index, don't scroll
    }
    
    const itemTop = itemIndex * itemHeight;
    const itemBottom = itemTop + itemHeight;
    const viewportTop = scrollTop;
    const viewportBottom = scrollTop + containerHeight;
    
    switch (alignment) {
      case 'start':
        return itemTop;
      
      case 'center':
        return itemTop - (containerHeight - itemHeight) / 2;
      
      case 'end':
        return itemBottom - containerHeight;
      
      case 'auto':
      default:
        // Only scroll if item is not visible
        if (itemTop < viewportTop) {
          return itemTop; // Scroll to start
        } else if (itemBottom > viewportBottom) {
          return itemBottom - containerHeight; // Scroll to end
        }
        return scrollTop; // No scroll needed
    }
  }

  /**
   * Determines if virtualization should be enabled
   * Business rule: Use virtualization only when beneficial for performance
   */
  shouldUseVirtualization(
    totalItems: number, 
    itemHeight: number, 
    containerHeight: number
  ): boolean {
    const visibleItems = Math.ceil(containerHeight / itemHeight);
    const virtualizationThreshold = APP_CONFIG.ui.virtualScroll.VIRTUALIZATION_THRESHOLD;
    
    // Use virtualization if total items significantly exceed visible items
    return totalItems > Math.max(virtualizationThreshold, visibleItems * 3);
  }

  /**
   * Calculates range for item pre-loading
   * Business rule: Pre-load items around visible area for smooth scrolling
   */
  calculatePreloadRange(
    metrics: ScrollMetrics, 
    scrollDirection: 'up' | 'down' | 'none',
    preloadFactor: number = 1.5
  ): { start: number; end: number } {
    const { renderStart, renderEnd, visibleStart, visibleEnd } = metrics;
    const visibleCount = visibleEnd - visibleStart + 1;
    const preloadCount = Math.floor(visibleCount * preloadFactor);
    
    let start = renderStart;
    let end = renderEnd;
    
    // Extend range based on scroll direction
    if (scrollDirection === 'down') {
      end = Math.min(end + preloadCount, metrics.scrollHeight);
    } else if (scrollDirection === 'up') {
      start = Math.max(start - preloadCount, 0);
    }
    
    return { start, end };
  }

  /**
   * Calculates dynamic item heights for variable-height items
   * Business rule: Support dynamic content while maintaining performance
   */
  calculateDynamicHeights(
    items: Array<{ height?: number }>, 
    defaultHeight: number
  ): Array<{ top: number; height: number }> {
    const result: Array<{ top: number; height: number }> = [];
    let currentTop = 0;
    
    for (let i = 0; i < items.length; i++) {
      const height = items[i].height || defaultHeight;
      
      result.push({
        top: currentTop,
        height
      });
      
      currentTop += height;
    }
    
    return result;
  }

  /**
   * Detects scroll direction and velocity
   * Business rule: Optimize rendering based on scroll behavior
   */
  detectScrollBehavior(
    currentScrollTop: number, 
    previousScrollTop: number, 
    timestamp: number,
    previousTimestamp: number
  ): {
    direction: 'up' | 'down' | 'none';
    velocity: number;
    acceleration: number;
  } {
    const deltaScroll = currentScrollTop - previousScrollTop;
    const deltaTime = timestamp - previousTimestamp;
    
    const direction = deltaScroll > 0 ? 'down' : deltaScroll < 0 ? 'up' : 'none';
    const velocity = deltaTime > 0 ? Math.abs(deltaScroll) / deltaTime : 0;
    
    // For acceleration, we would need to track previous velocity
    // For now, return 0 as placeholder
    const acceleration = 0;
    
    return { direction, velocity, acceleration };
  }

  /**
   * Validates virtual scroll configuration
   * Business rule: Ensure configuration is valid for safe operation
   */
  validateConfiguration(config: VirtualScrollConfig): { valid: boolean; errors: string[] } {
    const errors: string[] = [];
    
    if (config.itemHeight <= 0) {
      errors.push('Item height must be greater than 0');
    }
    
    if (config.containerHeight <= 0) {
      errors.push('Container height must be greater than 0');
    }
    
    if (config.totalItems < 0) {
      errors.push('Total items must be non-negative');
    }
    
    if (config.buffer < 0) {
      errors.push('Buffer must be non-negative');
    }
    
    if (config.overscan && config.overscan < 0) {
      errors.push('Overscan must be non-negative');
    }
    
    return {
      valid: errors.length === 0,
      errors
    };
  }

  /**
   * Gets performance recommendations based on metrics
   * Business rule: Provide actionable performance optimization suggestions
   */
  getPerformanceRecommendations(metrics: PerformanceMetrics): string[] {
    const recommendations: string[] = [];
    
    if (metrics.renderTime > this.performanceThreshold) {
      recommendations.push('Consider reducing buffer size to improve render performance');
    }
    
    if (metrics.visibleCount > metrics.itemCount * 0.8) {
      recommendations.push('Virtualization may not be beneficial with this item count');
    }
    
    if (metrics.renderCalls > 100) {
      recommendations.push('Consider debouncing scroll events to reduce render calls');
    }
    
    const timeSinceLastUpdate = Date.now() - metrics.lastUpdate;
    if (timeSinceLastUpdate < 16) { // Less than one frame
      recommendations.push('Scroll updates are very frequent, consider throttling');
    }
    
    return recommendations;
  }
}

// Default instance for dependency injection
export const defaultVirtualScrollingService = new VirtualScrollingService();