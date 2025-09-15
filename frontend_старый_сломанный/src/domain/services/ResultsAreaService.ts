import { APP_CONFIG } from '@/config/app-config'

export class ResultsAreaService {
  constructor() {}

  // Get log color based on level
  getLogColor(level: string) {
    switch (level) {
      case 'error': return 'text-red-400'
      case 'warning': return 'text-yellow-400'
      case 'success': return 'text-green-400'
      case 'info': return 'text-blue-400'
      default: return 'text-gray-300'
    }
  }

  // Clear results
  clearResults(generatedOutput: string, lastGenerated: string, diffStats: any, addToast: (message: string, type: string) => void) {
    generatedOutput = ''
    lastGenerated = ''
    diffStats = { additions: 0, deletions: 0, files: 0 }
    addToast('Results cleared', 'info')
  }

  // Export results
  exportResults(hasResults: boolean, openExport: () => void, addToast: (message: string, type: string) => void) {
    if (!hasResults) {
      addToast('No results to export', 'warning')
      return
    }
    
    // Open export modal
    openExport()
    addToast('Export dialog opened', 'info')
  }

  // Clear console
  clearConsole(consoleLogs: any[], addToast: (message: string, type: string) => void) {
    consoleLogs = []
    addToast('Console cleared', 'info')
  }

  // Export console
  exportConsole(consoleLogs: any[], addToast: (message: string, type: string) => void) {
    if (consoleLogs.length === 0) {
      addToast('No console logs to export', 'warning')
      return
    }
    
    // Create log text
    const logText = consoleLogs
      .map(log => `[${log.timestamp}] ${log.level.toUpperCase()}: ${log.message}`)
      .join('\n')
    
    // Copy to clipboard
    navigator.clipboard.writeText(logText).then(() => {
      addToast('Console logs copied to clipboard', 'success')
    }).catch(() => {
      addToast('Failed to copy console logs', 'error')
    })
  }
}