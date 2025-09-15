import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { EnhancedPanelService } from '@/domain/services/EnhancedPanelService'

export function useEnhancedPanelService() {
  // Create service instance
  const enhancedPanelService = new EnhancedPanelService()
  
  return {
    enhancedPanelService
  }
}