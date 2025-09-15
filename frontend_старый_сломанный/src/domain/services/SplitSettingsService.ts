import { APP_CONFIG } from '@/config/app-config'

export interface SplitSettings {
  maxTokensPerChunk: number
  enableAutoSplitting: boolean
  overlapSize: number
  splitStrategy: 'semantic' | 'balanced' | 'priority'
  preserveCodeBlocks: boolean
}

export class SplitSettingsService {
  constructor() {}

  // Get default split settings
  getDefaultSettings(): SplitSettings {
    return {
      maxTokensPerChunk: 8000,
      enableAutoSplitting: true,
      overlapSize: 100,
      splitStrategy: 'semantic',
      preserveCodeBlocks: true
    }
  }

  // Sanitize max tokens per chunk value
  sanitizeMaxTokensPerChunk(value: number): number {
    if (isNaN(value)) {
      return this.getDefaultSettings().maxTokensPerChunk
    }
    
    return Math.min(50000, Math.max(1000, value))
  }

  // Sanitize overlap size value
  sanitizeOverlapSize(value: number): number {
    if (isNaN(value)) {
      return this.getDefaultSettings().overlapSize
    }
    
    return Math.min(2000, Math.max(0, value))
  }
}