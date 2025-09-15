import { SplitSettingsService, SplitSettings } from '@/domain/services/SplitSettingsService'

export function useSplitSettingsService() {
  // Create service instance
  const splitSettingsService = new SplitSettingsService()
  
  return {
    splitSettingsService
  }
}