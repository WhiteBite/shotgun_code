import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useExportStore } from '@/stores/export.store'
import { useUiStore } from '@/stores/ui.store'
import { ResultsAreaService } from '@/domain/services/ResultsAreaService'

export function useResultsAreaService() {
  // Get stores
  const exportStore = useExportStore()
  const uiStore = useUiStore()
  
  // Get reactive refs from stores
  const { } = storeToRefs(exportStore)
  const { addToast } = uiStore
  
  // Create service instance
  const resultsAreaService = new ResultsAreaService()
  
  return {
    resultsAreaService,
    exportStore,
    uiStore,
    // Methods from uiStore
    addToast
  }
}