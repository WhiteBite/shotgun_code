import { ref, computed } from 'vue'
import { ReadFileContent, GetFileStats } from '@/wailsjs/go/main/App'

/**
 * QuickLook modal composable for file preview
 * Manages file loading, stats, and modal state
 */
export function useQuickLook() {
  const isOpen = ref(false)
  const isLoading = ref(false)
  const currentFile = ref<string | null>(null)
  const fileContent = ref<string>('')
  const fileStats = ref<any>(null)
  const error = ref<string | null>(null)

  const hasFile = computed(() => currentFile.value !== null)

  /**
   * Open QuickLook with a file
   */
  async function open(filePath: string) {
    currentFile.value = filePath
    isOpen.value = true
    isLoading.value = true
    error.value = null

    try {
      // Load file content and stats in parallel
      const [content, statsJson] = await Promise.all([
        ReadFileContent(filePath),
        GetFileStats(filePath)
      ])

      fileContent.value = content
      fileStats.value = JSON.parse(statsJson)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load file'
      console.error('QuickLook error:', err)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Close QuickLook and reset state
   */
  function close() {
    isOpen.value = false
    // Delay reset to allow close animation
    setTimeout(() => {
      currentFile.value = null
      fileContent.value = ''
      fileStats.value = null
      error.value = null
    }, 300)
  }

  /**
   * Copy file content to clipboard
   */
  async function copyToClipboard() {
    if (!fileContent.value) return

    try {
      await navigator.clipboard.writeText(fileContent.value)
      return true
    } catch (err) {
      console.error('Failed to copy to clipboard:', err)
      return false
    }
  }

  return {
    // State
    isOpen,
    isLoading,
    currentFile,
    fileContent,
    fileStats,
    error,
    hasFile,
    
    // Actions
    open,
    close,
    copyToClipboard
  }
}
