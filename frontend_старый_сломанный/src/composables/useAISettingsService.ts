import { AISettingsService, CodeGenerationSettings } from '@/domain/services/AISettingsService'

export function useAISettingsService() {
  // Create service instance
  const aiSettingsService = new AISettingsService()
  
  return {
    aiSettingsService
  }
}