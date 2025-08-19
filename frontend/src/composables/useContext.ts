import { ref, computed, readonly } from 'vue'
import type { Context } from '../types/api'
import { apiService } from '../services/api.service'
import { useErrorHandler } from './useErrorHandler'

export function useContext() {
  const { handleError } = useErrorHandler()
  
  const currentContext = ref<Context | null>(null)
  const contexts = ref<Context[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const hasContext = computed(() => currentContext.value !== null)
  const contextFiles = computed(() => currentContext.value?.files || [])
  const contextTokenCount = computed(() => currentContext.value?.tokenCount || 0)

  async function buildContext(projectPath: string, includedPaths: string[], options?: {
    includeGitStatus?: boolean
    includeCommitHistory?: boolean
    maxTokens?: number
  }) {
    isLoading.value = true
    error.value = null
    
    try {
      // Add safety limits to prevent OOM
      const maxFiles = 1000 // Maximum files to process
      const maxTokens = options?.maxTokens || 100000 // Maximum tokens
      
      if (includedPaths.length > maxFiles) {
        throw new Error(`Too many files selected (${includedPaths.length}). Maximum allowed: ${maxFiles}`)
      }
      
      console.log('Building context for:', includedPaths.length, 'files')
      await apiService.requestShotgunContextGeneration(projectPath, includedPaths)
      
      // Don't create a placeholder context - wait for the real one from backend
      // The real context will be set via the shotgunContextGenerated event
      console.log('Context generation requested, waiting for backend response...')
      
      // Return a minimal context object that will be updated by the event
      const context: Context = {
        id: Date.now().toString(),
        name: `Контекст ${new Date().toLocaleString('ru-RU')}`,
        description: `Запрос контекста для ${includedPaths.length} файлов`,
        content: `Запрос контекста отправлен в backend. Ожидание ответа...`,
        files: includedPaths,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        projectPath,
        tokenCount: 0
      }
      
      currentContext.value = context
      return context
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err)
      error.value = errorMessage
      handleError(err, 'Context building')
      throw new Error(errorMessage)
    } finally {
      isLoading.value = false
    }
  }

  async function loadContexts(projectPath: string) {
    isLoading.value = true
    error.value = null
    
    try {
      // For now, return empty array since we don't have persistence yet
      contexts.value = []
      return contexts.value
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err)
      error.value = errorMessage
      handleError(err, 'Loading contexts')
      throw new Error(errorMessage)
    } finally {
      isLoading.value = false
    }
  }

  async function loadContext(id: string) {
    isLoading.value = true
    error.value = null
    
    try {
      // For now, return null since we don't have persistence yet
      currentContext.value = null
      return null
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err)
      error.value = errorMessage
      handleError(err, 'Loading context')
      throw new Error(errorMessage)
    } finally {
      isLoading.value = false
    }
  }

  async function deleteContext(id: string) {
    try {
      // For now, just remove from local state
      contexts.value = contexts.value.filter((ctx: Context) => ctx.id !== id)
      if (currentContext.value?.id === id) {
        currentContext.value = null
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err)
      error.value = errorMessage
      handleError(err, 'Deleting context')
      throw new Error(errorMessage)
    }
  }

  function clearContext() {
    currentContext.value = null
    error.value = null
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    currentContext, // Убираю readonly - нужно изменять извне
    contexts: readonly(contexts),
    isLoading: readonly(isLoading),
    error: readonly(error),
    
    // Computed
    hasContext,
    contextFiles,
    contextTokenCount,
    
    // Methods
    buildContext,
    loadContexts,
    loadContext,
    deleteContext,
    clearContext,
    clearError
  }
}
