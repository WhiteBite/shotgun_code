import { ContextBuilderSettingsService, ContextBuilderSettings } from '@/domain/services/ContextBuilderSettingsService'

export function useContextBuilderSettingsService() {
  // Create service instance
  const contextBuilderSettingsService = new ContextBuilderSettingsService()
  
  return {
    contextBuilderSettingsService
  }
}