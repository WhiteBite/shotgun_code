import { computed } from 'vue'
import { APP_CONFIG } from '@/config/app-config'

export interface ContextSettings {
  maxContextSize: number
  maxFilesInContext: number
  includeDependencies: boolean
  includeTests: boolean
}

export class ContextSettingsService {
  constructor() {}

  // Get default context settings
  getDefaultSettings(): ContextSettings {
    return {
      maxContextSize: 10000,
      maxFilesInContext: 20,
      includeDependencies: true,
      includeTests: false
    }
  }

  // Sanitize max context size value
  sanitizeMaxContextSize(value: number): number {
    if (isNaN(value)) {
      return this.getDefaultSettings().maxContextSize
    }
    
    return Math.min(100000, Math.max(1000, value))
  }

  // Sanitize max files value
  sanitizeMaxFiles(value: number): number {
    if (isNaN(value)) {
      return this.getDefaultSettings().maxFilesInContext
    }
    
    return Math.min(100, Math.max(1, value))
  }

  // Calculate estimated cost based on tokens
  calculateEstimatedCost(tokensPerRequest: number): string {
    const costPerThousandTokens = 0.01 // Approximate cost for GPT-4
    return ((tokensPerRequest / 1000) * costPerThousandTokens).toFixed(4)
  }

  // Calculate efficiency score based on context size and file count
  calculateEfficiency(maxSize: number, maxFiles: number): number {
    let score = 100
    
    // Penalize very small contexts
    if (maxSize < 4000) score -= 20
    
    // Penalize very large contexts
    if (maxSize > 32000) score -= 30
    
    // Penalize too few files
    if (maxFiles < 5) score -= 15
    
    // Penalize too many files
    if (maxFiles > 50) score -= 25
    
    return Math.max(0, score)
  }

  // Get efficiency class based on score
  getEfficiencyClass(efficiency: number): string {
    if (efficiency >= 80) return 'text-green-400'
    if (efficiency >= 60) return 'text-yellow-400'
    return 'text-red-400'
  }

  // Calculate memory usage estimate
  calculateMemoryUsage(tokens: number, files: number): string {
    // Rough estimate: 4 chars per token, plus overhead for files
    const estimatedBytes = (tokens * 4) + (files * 1024) // 1KB overhead per file
    
    if (estimatedBytes < 1024 * 1024) {
      return `${Math.round(estimatedBytes / 1024)} KB`
    } else {
      return `${(estimatedBytes / (1024 * 1024)).toFixed(1)} MB`
    }
  }

  // Get memory usage class based on estimated bytes
  getMemoryUsageClass(tokens: number, files: number): string {
    const estimatedBytes = (tokens * 4) + (files * 1024)
    
    if (estimatedBytes > 10 * 1024 * 1024) return 'text-red-400' // > 10MB
    if (estimatedBytes > 5 * 1024 * 1024) return 'text-yellow-400' // > 5MB
    return 'text-green-400'
  }

  // Get optimization suggestions based on settings
  getSuggestions(maxSize: number, maxFiles: number, includeDependencies: boolean, includeTests: boolean): string[] {
    const suggestions: string[] = []
    
    if (maxSize < 4000) {
      suggestions.push('Consider increasing max context size to 8000+ tokens for better AI performance')
    }
    
    if (maxSize > 32000) {
      suggestions.push('Very large contexts may be slower and more expensive. Consider splitting into smaller chunks')
    }
    
    if (maxFiles > 50) {
      suggestions.push('Too many files may overwhelm the AI. Consider focusing on core files')
    }
    
    if (maxFiles < 5) {
      suggestions.push('Very few files may not provide enough context. Consider including related files')
    }
    
    if (!includeDependencies && maxFiles > 10) {
      suggestions.push('Enable "Include dependencies" to automatically include relevant imports')
    }
    
    if (includeTests && maxFiles > 20) {
      suggestions.push('Consider disabling test files inclusion to focus on implementation code')
    }
    
    return suggestions
  }

  // Computed properties for UI elements
  getEstimatedCost(tokensPerRequest: number) {
    return computed(() => this.calculateEstimatedCost(tokensPerRequest))
  }

  getEfficiency(maxSize: number, maxFiles: number) {
    return computed(() => this.calculateEfficiency(maxSize, maxFiles))
  }

  getEfficiencyClassForEfficiency(efficiency: number) {
    return computed(() => this.getEfficiencyClass(efficiency))
  }

  getMemoryUsage(tokens: number, files: number) {
    return computed(() => this.calculateMemoryUsage(tokens, files))
  }

  getMemoryUsageClassForUsage(tokens: number, files: number) {
    return computed(() => this.getMemoryUsageClass(tokens, files))
  }

  getSuggestionsForSettings(maxSize: number, maxFiles: number, includeDependencies: boolean, includeTests: boolean) {
    return computed(() => this.getSuggestions(maxSize, maxFiles, includeDependencies, includeTests))
  }
}