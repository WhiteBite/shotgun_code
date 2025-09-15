import { computed } from 'vue'
import { APP_CONFIG } from '@/config/app-config'

export interface ContextBuilderSettings {
  maxTokenLimit: number
  maxFileLimit: number
  excludeBinaryFiles: boolean
  autoBuildEnabled: boolean
  smartSuggestionsEnabled: boolean
  allowedExtensions: string[]
  maxIndividualFileSize?: number // in bytes
  maxTotalContextSize?: number // in bytes
}

export class ContextBuilderSettingsService {
  constructor() {}

  // Get default context builder settings from configuration
  getDefaultSettings(): ContextBuilderSettings {
    return {
      maxTokenLimit: 0, // 0 = unlimited
      maxFileLimit: 0, // 0 = unlimited
      excludeBinaryFiles: true,
      autoBuildEnabled: true,
      smartSuggestionsEnabled: true,
      allowedExtensions: APP_CONFIG.security.validation.ALLOWED_EXTENSIONS,
      maxIndividualFileSize: APP_CONFIG.performance.memory.MAX_INDIVIDUAL_FILE_SIZE,
      maxTotalContextSize: APP_CONFIG.context.estimation.RISK_THRESHOLDS.LOW
    }
  }

  // Convert bytes to KB for display
  bytesToKB(bytes: number): number {
    return Math.round(bytes / 1024)
  }

  // Convert KB to bytes
  kbToBytes(kb: number): number {
    return kb * 1024
  }

  // Convert bytes to MB for display
  bytesToMB(bytes: number): number {
    return Math.round(bytes / (1024 * 1024))
  }

  // Convert MB to bytes
  mbToBytes(mb: number): number {
    return mb * 1024 * 1024
  }

  // Process allowed extensions text
  processAllowedExtensionsText(text: string): string[] {
    return text
      .split(',')
      .map(ext => ext.trim())
      .filter(ext => ext.length > 0 && ext.startsWith('.'))
      .filter((ext, index, arr) => arr.indexOf(ext) === index) // Remove duplicates
  }

  // Calculate memory safety class based on settings
  calculateMemorySafetyClass(fileSizeKB: number, totalSizeMB: number): string {
    if (fileSizeKB > 1024 || totalSizeMB > 10) return 'bg-red-400' // High risk
    if (fileSizeKB > 512 || totalSizeMB > 5) return 'bg-yellow-400' // Medium risk
    return 'bg-green-400' // Low risk
  }

  // Calculate memory safety text class based on settings
  calculateMemorySafetyTextClass(fileSizeKB: number, totalSizeMB: number): string {
    if (fileSizeKB > 1024 || totalSizeMB > 10) return 'text-red-400'
    if (fileSizeKB > 512 || totalSizeMB > 5) return 'text-yellow-400'
    return 'text-green-400'
  }

  // Calculate memory safety text based on settings
  calculateMemorySafetyText(fileSizeKB: number, totalSizeMB: number): string {
    if (fileSizeKB > 1024 || totalSizeMB > 10) return 'High Risk'
    if (fileSizeKB > 512 || totalSizeMB > 5) return 'Medium Risk'
    return 'Safe'
  }

  // Calculate performance class based on settings
  calculatePerformanceClass(
    tokenLimit: number,
    fileLimit: number,
    autoEnabled: boolean
  ): string {
    let score = 100
    
    if (tokenLimit === 0 || tokenLimit > APP_CONFIG.performance.tokens.MAX_TOKENS_PER_REQUEST) score -= 30 // No limit or very high limit
    if (fileLimit === 0 || fileLimit > APP_CONFIG.fileTree.limits.MAX_SELECTED_FILES) score -= 20 // No limit or many files
    if (autoEnabled) score -= 10 // Auto-build can impact performance
    
    if (score >= 80) return 'bg-green-400'
    if (score >= 60) return 'bg-yellow-400'
    return 'bg-red-400'
  }

  // Calculate performance text class based on settings
  calculatePerformanceTextClass(
    tokenLimit: number,
    fileLimit: number,
    autoEnabled: boolean
  ): string {
    let score = 100
    
    if (tokenLimit === 0 || tokenLimit > APP_CONFIG.performance.tokens.MAX_TOKENS_PER_REQUEST) score -= 30
    if (fileLimit === 0 || fileLimit > APP_CONFIG.fileTree.limits.MAX_SELECTED_FILES) score -= 20
    if (autoEnabled) score -= 10
    
    if (score >= 80) return 'text-green-400'
    if (score >= 60) return 'text-yellow-400'
    return 'text-red-400'
  }

  // Calculate performance text based on settings
  calculatePerformanceText(
    tokenLimit: number,
    fileLimit: number,
    autoEnabled: boolean
  ): string {
    let score = 100
    
    if (tokenLimit === 0 || tokenLimit > APP_CONFIG.performance.tokens.MAX_TOKENS_PER_REQUEST) score -= 30
    if (fileLimit === 0 || fileLimit > APP_CONFIG.fileTree.limits.MAX_SELECTED_FILES) score -= 20
    if (autoEnabled) score -= 10
    
    if (score >= 80) return 'Optimal'
    if (score >= 60) return 'Good'
    return 'Slow'
  }

  // Computed properties for UI elements
  getMemorySafetyClass(fileSizeKB: number, totalSizeMB: number) {
    return computed(() => this.calculateMemorySafetyClass(fileSizeKB, totalSizeMB))
  }

  getMemorySafetyTextClass(fileSizeKB: number, totalSizeMB: number) {
    return computed(() => this.calculateMemorySafetyTextClass(fileSizeKB, totalSizeMB))
  }

  getMemorySafetyText(fileSizeKB: number, totalSizeMB: number) {
    return computed(() => this.calculateMemorySafetyText(fileSizeKB, totalSizeMB))
  }

  getPerformanceClass(tokenLimit: number, fileLimit: number, autoEnabled: boolean) {
    return computed(() => this.calculatePerformanceClass(tokenLimit, fileLimit, autoEnabled))
  }

  getPerformanceTextClass(tokenLimit: number, fileLimit: number, autoEnabled: boolean) {
    return computed(() => this.calculatePerformanceTextClass(tokenLimit, fileLimit, autoEnabled))
  }

  getPerformanceText(tokenLimit: number, fileLimit: number, autoEnabled: boolean) {
    return computed(() => this.calculatePerformanceText(tokenLimit, fileLimit, autoEnabled))
  }
}