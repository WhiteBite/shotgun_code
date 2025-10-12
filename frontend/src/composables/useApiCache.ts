/**
 * Composable for caching API requests
 * Reduces unnecessary network calls and improves performance
 */

import { ref } from 'vue'

interface CacheEntry<T> {
  data: T
  timestamp: number
}

// Global cache shared across all instances
const cache = new Map<string, CacheEntry<any>>()

const DEFAULT_TTL = 5 * 60 * 1000 // 5 minutes

export function useApiCache<T>(
  key: string,
  fetcher: () => Promise<T>,
  ttl = DEFAULT_TTL
) {
  const data = ref<T | null>(null)
  const isLoading = ref(false)
  const error = ref<Error | null>(null)

  const load = async (force = false): Promise<T | null> => {
    // Check cache first (unless force refresh)
    if (!force) {
      const cached = cache.get(key)
      if (cached && Date.now() - cached.timestamp < ttl) {
        data.value = cached.data
        return cached.data
      }
    }

    // Fetch fresh data
    isLoading.value = true
    error.value = null

    try {
      const result = await fetcher()
      data.value = result
      
      // Store in cache
      cache.set(key, {
        data: result,
        timestamp: Date.now()
      })
      
      return result
    } catch (e) {
      error.value = e as Error
      console.error(`API cache error for key "${key}":`, e)
      throw e
    } finally {
      isLoading.value = false
    }
  }

  const invalidate = () => {
    cache.delete(key)
    data.value = null
  }

  const invalidateAll = () => {
    cache.clear()
  }

  return {
    data,
    isLoading,
    error,
