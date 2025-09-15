import { ContextSettingsService, ContextSettings } from '@/domain/services/ContextSettingsService'

export function useContextSettingsService() {
  // Create service instance
  const contextSettingsService = new ContextSettingsService()
  
  return {
    contextSettingsService
  }
}