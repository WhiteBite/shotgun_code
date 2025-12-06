import { ref, reactive } from 'vue'
import { useContextStore } from '@/features/context/model/context.store'
import { apiService } from '@/services/api.service'

export type ExportMode = 'clipboard' | 'ai' | 'human'

export interface ExportSettings {
  // Clipboard
  exportFormat: 'plain' | 'manifest' | 'json'
  stripComments: boolean
  includeManifest: boolean
  
  // AI
  aiProfile: string
  tokenLimit: number
  enableAutoSplit: boolean
  maxTokensPerChunk: number
  overlapTokens: number
  splitStrategy: 'smart' | 'file' | 'token'
  
  // Human
  theme: string
  includeLineNumbers: boolean
  includePageNumbers: boolean
}

/**
 * Export modal composable for managing export functionality
 * Handles export settings, modal state, and export execution
 */
export function useExport() {
  const contextStore = useContextStore()
  
  const isOpen = ref(false)
  const isExporting = ref(false)
  const selectedMode = ref<ExportMode>('clipboard')
  const exportResult = ref<any>(null)
  const error = ref<string | null>(null)

  const settings = reactive<ExportSettings>({
    // Clipboard defaults
    exportFormat: 'manifest',
    stripComments: false,
    includeManifest: true,
    
    // AI defaults
    aiProfile: 'gpt-4',
    tokenLimit: 8000,
    enableAutoSplit: false,
    maxTokensPerChunk: 4000,
    overlapTokens: 200,
    splitStrategy: 'smart',
    
    // Human defaults
    theme: 'default',
    includeLineNumbers: true,
    includePageNumbers: true
  })

  /**
   * Open export modal
   */
  function open() {
    if (!contextStore.hasContext) {
      error.value = 'Сначала соберите контекст'
      return false
    }
    
    isOpen.value = true
    exportResult.value = null
    error.value = null
    return true
  }

  /**
   * Close export modal
   */
  function close() {
    isOpen.value = false
  }

  /**
   * Execute export with current settings
   */
  async function executeExport() {
    if (!contextStore.hasContext) {
      error.value = 'Контекст не найден'
      return false
    }

    isExporting.value = true
    error.value = null

    try {
      // Get full context content (warning: can be large)
      const contextContent = await contextStore.getFullContextContent()
      
      // Build export settings JSON
      const exportSettingsJson = {
        mode: selectedMode.value,
        context: contextContent,
        
        // Clipboard settings
        stripComments: settings.stripComments,
        includeManifest: settings.includeManifest,
        exportFormat: settings.exportFormat,
        
        // AI settings
        aiProfile: settings.aiProfile,
        tokenLimit: settings.tokenLimit,
        fileSizeLimitKB: 5120, // 5 MB
        enableAutoSplit: settings.enableAutoSplit,
        maxTokensPerChunk: settings.maxTokensPerChunk,
        overlapTokens: settings.overlapTokens,
        splitStrategy: settings.splitStrategy,
        
        // Human settings
        theme: settings.theme,
        includeLineNumbers: settings.includeLineNumbers,
        includePageNumbers: settings.includePageNumbers
      }

      const result = await apiService.exportContext(exportSettingsJson)
      exportResult.value = result

      // Handle result based on mode
      if (result.mode === 'clipboard' && result.text) {
        await navigator.clipboard.writeText(result.text)
        console.log('Content copied to clipboard')
      } else if (result.filePath) {
        console.log('File exported to:', result.filePath)
      } else if (result.dataBase64) {
        // Download file
        downloadBase64File(result.dataBase64, result.fileName || 'export.zip')
      }

      // Auto close after success
      setTimeout(() => {
        close()
      }, 1500)

      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка экспорта'
      console.error('Export failed:', err)
      return false
    } finally {
      isExporting.value = false
    }
  }

  /**
   * Download base64 encoded file
   */
  function downloadBase64File(base64: string, filename: string) {
    const byteCharacters = atob(base64)
    const byteNumbers = new Array(byteCharacters.length)
    for (let i = 0; i < byteCharacters.length; i++) {
      byteNumbers[i] = byteCharacters.charCodeAt(i)
    }
    const byteArray = new Uint8Array(byteNumbers)
    const blob = new Blob([byteArray])
    
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  }

  return {
    // State
    isOpen,
    isExporting,
    selectedMode,
    settings,
    exportResult,
    error,
    // Include context store for use in components
    contextStore,
    
    // Actions
    open,
    close,
    executeExport
  }
}
