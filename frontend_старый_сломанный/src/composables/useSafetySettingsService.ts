import { SafetySettingsService, SafetySettings } from '@/domain/services/SafetySettingsService'

export function useSafetySettingsService() {
  // Create service instance
  const safetySettingsService = new SafetySettingsService()
  
  return {
    safetySettingsService
  }
}