import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface ContextSummary {
  id: string
  fileCount: number
  totalSize: number
  lineCount: number
}

export interface ContextChunk {
  lines: string[]
  startLine: number
  endLine: number
  hasMore: boolean
}

// MAX_CONTEXT_SIZE is defined but available for future use
export const MAX_CONTEXT_SIZE = 50 * 1024 * 1024 // 50 MB OOM-safe limit

export const useContextStore = defineStore('context', () => {
  // State
  const contextId = ref<string | null>(null)
  const summary = ref<ContextSummary | null>(null)
  const currentChunk = ref<ContextChunk | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const hasContext = computed(() => contextId.value !== null)
  const totalLines = computed(() => summary.value?.lineCount || 0)
  const totalSize = computed(() => summary.value?.totalSize || 0)
  const fileCount = computed(() => summary.value?.fileCount || 0)
  const lineCount = computed(() => totalLines.value) // Alias for totalLines
  const tokenCount = computed(() => Math.round(totalLines.value * 2.5)) // Estimated tokens
  const estimatedCost = computed(() => (tokenCount.value / 1000) * 0.002) // Dummy cost

  // Actions
  async function buildContext(filePaths: string[]) {
    if (filePaths.length === 0) {
      error.value = 'No files selected'
      return
    }

    isLoading.value = true
    error.value = null

    try {
      // TODO: Call backend API /api/context/build
      // For now, create dummy summary
      contextId.value = `ctx-${Date.now()}`
      summary.value = {
        id: contextId.value,
        fileCount: filePaths.length,
        totalSize: 0,
        lineCount: 0
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to build context'
    } finally {
      isLoading.value = false
    }
  }

  async function loadChunk(startLine: number, lineCount: number) {
    if (!contextId.value) {
      error.value = 'No context available'
      return
    }

    isLoading.value = true
    error.value = null

    try {
      // TODO: Call backend API /api/context/{id}/lines?start={start}&count={count}
      currentChunk.value = {
        lines: [],
        startLine,
        endLine: startLine + lineCount,
        hasMore: false
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load chunk'
    } finally {
      isLoading.value = false
    }
  }

  async function getContextContent(startLine: number, endLine: number) {
    if (!contextId.value) {
      error.value = 'No context available'
      return ''
    }

    isLoading.value = true
    error.value = null

    try {
      // TODO: Call backend API /api/context/{id}/content?start={start}&end={end}
      console.log(`Fetching content from ${startLine} to ${endLine}`)
      return 'dummy context content' // Return dummy content
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load context'
      return ''
    } finally {
      isLoading.value = false
    }
  }

  function clearContext() {
    contextId.value = null
    summary.value = null
    currentChunk.value = null
    error.value = null
  }

  return {
    // State
    contextId,
    summary,
    currentChunk,
    isLoading,
    error,
    // Computed
    hasContext,
    totalLines,
    totalSize,
    fileCount,
    lineCount,
    tokenCount,
    estimatedCost,
    // Actions
    buildContext,
    loadChunk,
    getContextContent,
    clearContext
  }
})
