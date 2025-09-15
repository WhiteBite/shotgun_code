import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { ContextCollectionService } from '@/domain/services/ContextCollectionService'

export function useContextCollectionService() {
  // Get stores
  const contextBuilderStore = useContextBuilderStore()
  
  // Get reactive refs from store
  const {
    selectedFilesList,
    isBuilding,
    canBuildContext,
    contextMetrics,
    suggestions,
    buildStatus,
    error
  } = storeToRefs(contextBuilderStore)
  
  // Create service instance
  const contextCollectionService = new ContextCollectionService()
  
  // Computed properties that depend on store state
  const hasSelectedFiles = computed(() => selectedFilesList.value.length > 0)
  
  const contextSummary = computed(() => ({
    files: contextMetrics.value.fileCount,
    characters: contextMetrics.value.characterCount,
    tokens: contextMetrics.value.tokenCount,
    cost: contextMetrics.value.estimatedCost
  }))
  
  const suggestedFiles = computed(() => 
    suggestions.value.map((suggestion: ContextSuggestion) => ({
      path: suggestion.filePath,
      reason: suggestion.reason
    }))
  )
  
  const isBuildInProgress = computed(() => 
    buildStatus.value === 'building' || buildStatus.value === 'validating'
  )
  
  return {
    contextCollectionService,
    contextBuilderStore,
    // Store refs
    selectedFilesList,
    isBuilding,
    canBuildContext,
    contextMetrics,
    suggestions,
    buildStatus,
    error,
    // Computed properties
    hasSelectedFiles,
    contextSummary,
    suggestedFiles,
    isBuildInProgress
  }
}