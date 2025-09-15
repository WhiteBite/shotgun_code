// Performance Optimization Utilities

import { ref, computed, nextTick } from 'vue'

/**
 * Virtual Scrolling Utilities
 */

export interface VirtualScrollOptions {
  itemHeight: number
  containerHeight: number
  buffer: number
  threshold: number
}

export function useVirtualScroll<T>(
  items: T[],
  options: VirtualScrollOptions
) {
  const scrollTop = ref(0)
  const containerHeight = options.containerHeight
  const itemHeight = options.itemHeight
  const buffer = options.buffer || 5

  const visibleStart = computed(() => {
    return Math.max(0, Math.floor(scrollTop.value / itemHeight) - buffer)
  })

  const visibleEnd = computed(() => {
    const visibleCount = Math.ceil(containerHeight / itemHeight)
    return Math.min(items.length, visibleStart.value + visibleCount + buffer)
  })

  const visibleItems = computed(() => {
    return items.slice(visibleStart.value, visibleEnd.value)
  })

  const totalHeight = computed(() => {
    return items.length * itemHeight
  })

  const offsetY = computed(() => {
    return visibleStart.value * itemHeight
  })

  function handleScroll(event: Event) {
    const target = event.target as HTMLElement
    scrollTop.value = target.scrollTop
  }

  return {
    visibleItems,
    visibleStart,
    visibleEnd,
    totalHeight,
    offsetY,
    handleScroll
  }
}

/**
 * Debounce and Throttle Utilities
 */

export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null

  return (...args: Parameters<T>) => {
    if (timeout) {
      clearTimeout(timeout)
    }
    timeout = setTimeout(() => func(...args), wait)
  }
}

export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle = false

  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}

/**
 * Lazy Loading Utilities
 */

export function useLazyLoading(threshold: number = 100) {
  const observer = ref<IntersectionObserver | null>(null)
  const loadedItems = ref(new Set<string>())

  function observe(element: HTMLElement, id: string, callback: () => void) {
    if (!observer.value) {
      observer.value = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (entry.isIntersecting) {
              const itemId = entry.target.getAttribute('data-id')
              if (itemId && !loadedItems.value.has(itemId)) {
                loadedItems.value.add(itemId)
                callback()
                observer.value?.unobserve(entry.target)
              }
            }
          })
        },
        {
          rootMargin: `${threshold}px`
        }
      )
    }

    element.setAttribute('data-id', id)
    observer.value.observe(element)
  }

  function disconnect() {
    observer.value?.disconnect()
    observer.value = null
    loadedItems.value.clear()
  }

  return {
    observe,
    disconnect,
    loadedItems: computed(() => Array.from(loadedItems.value))
  }
}

/**
 * Memory Management Utilities
 */

export class MemoryManager {
  private cache = new Map<string, any>()
  private maxSize: number
  private ttl: number

  constructor(maxSize: number = 100, ttl: number = 5 * 60 * 1000) {
    this.maxSize = maxSize
    this.ttl = ttl
  }

  set(key: string, value: any): void {
    // Remove oldest entries if cache is full
    if (this.cache.size >= this.maxSize) {
      const firstKey = this.cache.keys().next().value
      this.cache.delete(firstKey)
    }

    this.cache.set(key, {
      value,
      timestamp: Date.now()
    })
  }

  get(key: string): any | null {
    const item = this.cache.get(key)
    if (!item) return null

    // Check if item has expired
    if (Date.now() - item.timestamp > this.ttl) {
      this.cache.delete(key)
      return null
    }

    return item.value
  }

  delete(key: string): boolean {
    return this.cache.delete(key)
  }

  clear(): void {
    this.cache.clear()
  }

  size(): number {
    return this.cache.size
  }

  cleanup(): void {
    const now = Date.now()
    for (const [key, item] of this.cache.entries()) {
      if (now - item.timestamp > this.ttl) {
        this.cache.delete(key)
      }
    }
  }
}

/**
 * Performance Monitoring
 */

export class PerformanceMonitor {
  private measurements = new Map<string, number>()
  private marks = new Map<string, number>()

  startMeasurement(name: string): void {
    this.marks.set(name, performance.now())
  }

  endMeasurement(name: string): number {
    const startTime = this.marks.get(name)
    if (!startTime) {
      console.warn(`No start mark found for measurement: ${name}`)
      return 0
    }

    const duration = performance.now() - startTime
    this.measurements.set(name, duration)
    this.marks.delete(name)
    
    return duration
  }

  getMeasurement(name: string): number | undefined {
    return this.measurements.get(name)
  }

  getAllMeasurements(): Record<string, number> {
    return Object.fromEntries(this.measurements)
  }

  logMeasurement(name: string): void {
    const duration = this.getMeasurement(name)
    if (duration !== undefined) {
      console.log(`Performance [${name}]: ${duration.toFixed(2)}ms`)
    }
  }

  clear(): void {
    this.measurements.clear()
    this.marks.clear()
  }
}

/**
 * Component Performance Wrapper
 */

export function withPerformanceMonitoring<T extends (...args: any[]) => any>(
  func: T,
  name: string,
  monitor: PerformanceMonitor
): T {
  return ((...args: any[]) => {
    monitor.startMeasurement(name)
    const result = func(...args)
    
    if (result && typeof result.then === 'function') {
      // Handle promises
      return result.finally(() => {
        monitor.endMeasurement(name)
        monitor.logMeasurement(name)
      })
    } else {
      monitor.endMeasurement(name)
      monitor.logMeasurement(name)
      return result
    }
  }) as T
}

/**
 * Bundle Size Optimization
 */

export function asyncComponent(loader: () => Promise<any>) {
  return () => {
    return loader().catch(error => {
      console.error('Failed to load component:', error)
      // Return fallback component
      return {
        template: '<div class="error-loading">Failed to load component</div>'
      }
    })
  }
}

/**
 * Request Batching for API Calls
 */

export class RequestBatcher {
  private batches = new Map<string, {
    requests: Array<{
      resolve: (value: any) => void
      reject: (error: any) => void
      data: any
    }>
    timeout: NodeJS.Timeout
  }>()

  private batchDelay: number

  constructor(batchDelay: number = 50) {
    this.batchDelay = batchDelay
  }

  batch<T>(
    key: string,
    data: any,
    batchHandler: (requests: any[]) => Promise<T[]>
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      let batch = this.batches.get(key)

      if (!batch) {
        batch = {
          requests: [],
          timeout: setTimeout(() => {
            this.executeBatch(key, batchHandler)
          }, this.batchDelay)
        }
        this.batches.set(key, batch)
      }

      batch.requests.push({ resolve, reject, data })
    })
  }

  private async executeBatch<T>(
    key: string,
    batchHandler: (requests: any[]) => Promise<T[]>
  ): Promise<void> {
    const batch = this.batches.get(key)
    if (!batch) return

    this.batches.delete(key)
    clearTimeout(batch.timeout)

    try {
      const requestData = batch.requests.map(req => req.data)
      const results = await batchHandler(requestData)

      batch.requests.forEach((request, index) => {
        request.resolve(results[index])
      })
    } catch (error) {
      batch.requests.forEach(request => {
        request.reject(error)
      })
    }
  }
}

// Global instances
export const globalMemoryManager = new MemoryManager()
export const globalPerformanceMonitor = new PerformanceMonitor()
export const globalRequestBatcher = new RequestBatcher()

// Cleanup function for component unmounting
export function cleanupPerformanceUtils(): void {
  globalMemoryManager.cleanup()
  globalPerformanceMonitor.clear()
}