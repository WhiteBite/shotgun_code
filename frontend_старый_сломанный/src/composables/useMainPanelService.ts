import { storeToRefs } from 'pinia'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { MainPanelService } from '@/domain/services/MainPanelService'

export function useMainPanelService() {
  // Get stores
  const contextBuilderStore = useContextBuilderStore()
  
  // Get reactive refs from stores
  const { 
    currentContext
  } = storeToRefs(contextBuilderStore)
  
  // Create service instance
  const mainPanelService = new MainPanelService()
  
  return {
    mainPanelService,
    contextBuilderStore,
    // Store refs
    currentContext
  }
}