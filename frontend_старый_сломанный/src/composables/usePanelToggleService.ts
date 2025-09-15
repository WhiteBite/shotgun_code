import { PanelToggleService } from '@/domain/services/PanelToggleService'

export function usePanelToggleService() {
  // Create service instance
  const panelToggleService = new PanelToggleService()
  
  return {
    panelToggleService
  }
}