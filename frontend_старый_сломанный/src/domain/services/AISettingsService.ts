import { computed } from 'vue'
import { APP_CONFIG } from '@/config/app-config'

export interface CodeGenerationSettings {
  temperature: number
  maxTokens: number
  autoFormat: boolean
  includeComments: boolean
}

export class AISettingsService {
  constructor() {}

  // Get default AI settings from configuration
  getDefaultSettings(): CodeGenerationSettings {
    return {
      temperature: APP_CONFIG.ai.defaultSettings.temperature,
      maxTokens: APP_CONFIG.ai.defaultSettings.maxTokens,
      autoFormat: true,
      includeComments: true
    }
  }

  // Sanitize max tokens value based on configuration limits
  sanitizeMaxTokens(value: number): number {
    if (isNaN(value)) {
      return APP_CONFIG.ai.defaultSettings.maxTokens
    }
    
    return Math.min(
      APP_CONFIG.performance.tokens.MAX_TOKENS_PER_REQUEST,
      Math.max(APP_CONFIG.context.streaming.CHUNK_SIZE, value)
    )
  }

  // Calculate creativity class based on temperature
  calculateCreativityClass(temperature: number): string {
    if (temperature >= 1.5) return 'bg-purple-400'
    if (temperature >= 1.0) return 'bg-blue-400'
    if (temperature >= 0.5) return 'bg-green-400'
    return 'bg-gray-400'
  }

  // Calculate creativity text class based on temperature
  calculateCreativityTextClass(temperature: number): string {
    if (temperature >= 1.5) return 'text-purple-400'
    if (temperature >= 1.0) return 'text-blue-400'
    if (temperature >= 0.5) return 'text-green-400'
    return 'text-gray-400'
  }

  // Calculate creativity text based on temperature
  calculateCreativityText(temperature: number): string {
    if (temperature >= 1.5) return 'Very High'
    if (temperature >= 1.0) return 'High'
    if (temperature >= 0.5) return 'Medium'
    return 'Low'
  }

  // Calculate reliability class based on temperature
  calculateReliabilityClass(temperature: number): string {
    if (temperature <= 0.3) return 'bg-green-400'
    if (temperature <= 0.7) return 'bg-blue-400'
    if (temperature <= 1.2) return 'bg-yellow-400'
    return 'bg-red-400'
  }

  // Calculate reliability text class based on temperature
  calculateReliabilityTextClass(temperature: number): string {
    if (temperature <= 0.3) return 'text-green-400'
    if (temperature <= 0.7) return 'text-blue-400'
    if (temperature <= 1.2) return 'text-yellow-400'
    return 'text-red-400'
  }

  // Calculate reliability text based on temperature
  calculateReliabilityText(temperature: number): string {
    if (temperature <= 0.3) return 'Very High'
    if (temperature <= 0.7) return 'High'
    if (temperature <= 1.2) return 'Medium'
    return 'Low'
  }

  // Computed properties for quality indicators (these would typically be used in the component)
  getCreativityClass(temperature: number) {
    return computed(() => this.calculateCreativityClass(temperature))
  }

  getCreativityTextClass(temperature: number) {
    return computed(() => this.calculateCreativityTextClass(temperature))
  }

  getCreativityText(temperature: number) {
    return computed(() => this.calculateCreativityText(temperature))
  }

  getReliabilityClass(temperature: number) {
    return computed(() => this.calculateReliabilityClass(temperature))
  }

  getReliabilityTextClass(temperature: number) {
    return computed(() => this.calculateReliabilityTextClass(temperature))
  }

  getReliabilityText(temperature: number) {
    return computed(() => this.calculateReliabilityText(temperature))
  }
}