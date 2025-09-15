import { UIStateService } from '@/domain/services/UIStateService'

export function useUIStateService() {
  // Create service instance
  const uiStateService = new UIStateService()
  
  return {
    uiStateService
  }
}