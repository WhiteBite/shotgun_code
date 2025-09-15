import { APP_CONFIG } from '@/config/app-config'

export interface SafetySettings {
  enableGuardrails: boolean
  maxMemoryWarning: number
  maxContextSizeMB: number
  validateFileTypes: boolean
  excludeBinaryFiles: boolean
  scanForSecrets: boolean
  enableContentFiltering: boolean
  requireUserConfirmation: boolean
  safetyLevel: 'relaxed' | 'balanced' | 'strict'
}

export class SafetySettingsService {
  constructor() {}

  // Get default safety settings
  getDefaultSettings(): SafetySettings {
    return {
      enableGuardrails: true,
      maxMemoryWarning: APP_CONFIG.performance.memory.WARNING_THRESHOLD_MB,
      maxContextSizeMB: APP_CONFIG.context.estimation.RISK_THRESHOLDS.LOW / (1024 * 1024),
      validateFileTypes: true,
      excludeBinaryFiles: true,
      scanForSecrets: true,
      enableContentFiltering: true,
      requireUserConfirmation: true,
      safetyLevel: 'balanced'
    }
  }

  // Sanitize memory warning value
  sanitizeMemoryWarning(value: number): number {
    if (isNaN(value)) {
      return this.getDefaultSettings().maxMemoryWarning
    }
    
    return Math.min(1000, Math.max(10, value))
  }

  // Sanitize context size value
  sanitizeContextSize(value: number): number {
    if (isNaN(value)) {
      return this.getDefaultSettings().maxContextSizeMB
    }
    
    return Math.min(100, Math.max(1, value))
  }
}