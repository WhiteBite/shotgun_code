import { computed } from 'vue'
import type { ContextMetrics } from '@/types/dto'
import { APP_CONFIG } from '@/config/app-config'

export class StatusBarService {
  constructor(
    private getAiProvider: () => string | null,
    private getBuildStatus: () => string,
    private getContextMetrics: () => ContextMetrics,
    private getSelectedFilesCount: () => number,
    private getLastContextGeneration: () => Date | null,
    private getNotificationCount: () => number
  ) {}

  // AI Provider Status
  get aiProviderName() {
    return this.getAiProvider() || 'No Provider'
  }

  get aiProviderStatusColor() {
    // This would check actual AI provider connectivity
    return 'bg-green-400' // Connected
    // return 'bg-yellow-400' // Warning
    // return 'bg-red-400' // Disconnected
  }

  // Git Status
  get gitStatus() {
    // This would show current git status
    return null // 'main • 2 ahead'
  }

  // Build Status
  get buildStatus() {
    if (this.getBuildStatus() === 'building') {
      return {
        isRunning: true,
        message: 'Building context...'
      }
    }
    
    if (this.getBuildStatus() === 'error') {
      return {
        isRunning: false,
        message: 'Build failed'
      }
    }
    
    return null
  }

  get buildStatusIconType() {
    if (this.getBuildStatus() === 'complete') return 'check'
    if (this.getBuildStatus() === 'error') return 'error'
    return 'warning'
  }

  // Context Progress
  get contextProgress() {
    if (this.getBuildStatus() === 'building') {
      return {
        visible: true,
        message: 'Building context',
        percentage: 45 // This would be calculated based on actual progress
      }
    }
    
    return { visible: false, message: '', percentage: 0 }
  }

  // Generation Progress
  get generationProgress() {
    // This would track AI generation progress
    return { visible: false, message: '', percentage: 0 }
  }

  // File and Token Stats
  get selectedFilesText() {
    const count = this.getSelectedFilesCount()
    if (count === 0) return '0 files'
    return `${count} file${count === 1 ? '' : 's'}`
  }

  get tokenCount() {
    return this.getContextMetrics().tokenCount || 0
  }

  get formattedTokenCount() {
    const count = this.tokenCount
    if (count >= 1000) {
      return `${(count / 1000).toFixed(1)}k tokens`
    }
    return `${count} tokens`
  }

  get estimatedCost() {
    return this.getContextMetrics().estimatedCost || 0
  }

  get formattedCost() {
    return this.estimatedCost.toFixed(4)
  }

  // Last Action
  get lastAction() {
    const lastGen = this.getLastContextGeneration()
    if (lastGen) {
      const timeAgo = this.getTimeAgo(lastGen)
      return {
        message: 'Context built',
        timeAgo
      }
    }
    return null
  }

  // Notifications
  get notificationCount() {
    return this.getNotificationCount()
  }

  // Helper functions
  private getTimeAgo(date: Date): string {
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    
    if (minutes < 1) return 'just now'
    if (minutes < 60) return `${minutes}m ago`
    
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `${hours}h ago`
    
    const days = Math.floor(hours / 24)
    return `${days}d ago`
  }
}