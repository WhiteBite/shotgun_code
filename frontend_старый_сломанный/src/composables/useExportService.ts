import { onMounted, onUnmounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useExportStore } from '@/stores/export.store'
import { ExportService } from '@/domain/services/ExportService'

export function useExportService() {
  // Get store
  const exportStore = useExportStore()
  
  // Get reactive refs
  const { 
    isOpen, 
    isLoading,
    exportFormat,
    stripComments,
    includeManifest,
    aiProfile,
    enableAutoSplit,
    maxTokensPerChunk,
    splitStrategy,
    overlapTokens,
    tokenLimit,
    fileSizeLimitKB,
    theme,
    includeLineNumbers,
    includePageNumbers
  } = storeToRefs(exportStore)
  
  // Create service instance
  const exportService = new ExportService(
    () => isOpen.value,
    () => isLoading.value,
    () => exportStore.doExportClipboard(),
    () => exportStore.doExportAI(),
    () => exportStore.doExportHuman(),
    () => exportStore.close()
  )
  
  // Sync store refs with service
  // Export format
  Object.defineProperty(exportService, 'exportFormat', {
    get: () => exportFormat.value,
    set: (value) => { exportFormat.value = value },
    enumerable: true,
    configurable: true
  })
  
  Object.defineProperty(exportService, 'stripComments', {
    get: () => stripComments.value,
    set: (value) => { stripComments.value = value },
    enumerable: true,
    configurable: true
  })
  
  Object.defineProperty(exportService, 'includeManifest', {
    get: () => includeManifest.value,
    set: (value) => { includeManifest.value = value },
    enumerable: true,
    configurable: true
  })
  
  // AI Profile
  Object.defineProperty(exportService, 'aiProfile', {
    get: () => aiProfile.value,
    set: (value) => { aiProfile.value = value },
    enumerable: true,
    configurable: true
  })
  
  // Auto-split settings
  Object.defineProperty(exportService, 'enableAutoSplit', {
    get: () => enableAutoSplit.value,
    set: (value) => { enableAutoSplit.value = value },
    enumerable: true,
    configurable: true
  })
  
  Object.defineProperty(exportService, 'maxTokensPerChunk', {
    get: () => maxTokensPerChunk.value,
    set: (value) => { maxTokensPerChunk.value = value },
    enumerable: true,
    configurable: true
  })
  
  Object.defineProperty(exportService, 'splitStrategy', {
    get: () => splitStrategy.value,
    set: (value) => { splitStrategy.value = value },
    enumerable: true,
    configurable: true
  })
  
  Object.defineProperty(exportService, 'overlapTokens', {
    get: () => overlapTokens.value,
    set: (value) => { overlapTokens.value = value },
    enumerable: true,
    configurable: true
  })
  
  // Limits
  Object.defineProperty(exportService, 'tokenLimit', {
    get: () => tokenLimit.value,
    set: (value) => { tokenLimit.value = value },
    enumerable: true,
    configurable: true
  })
  
  Object.defineProperty(exportService, 'fileSizeLimitKB', {
    get: () => fileSizeLimitKB.value,
    set: (value) => { fileSizeLimitKB.value = value },
    enumerable: true,
    configurable: true
  })
  
  // Human-readable export settings
  Object.defineProperty(exportService, 'theme', {
    get: () => theme.value,
    set: (value) => { theme.value = value },
    enumerable: true,
    configurable: true
  })
  
  Object.defineProperty(exportService, 'includeLineNumbers', {
    get: () => includeLineNumbers.value,
    set: (value) => { includeLineNumbers.value = value },
    enumerable: true,
    configurable: true
  })
  
  Object.defineProperty(exportService, 'includePageNumbers', {
    get: () => includePageNumbers.value,
    set: (value) => { includePageNumbers.value = value },
    enumerable: true,
    configurable: true
  })
  
  // Keyboard event handling
  const onKeydown = (e: KeyboardEvent) => {
    if (e.key === "Escape") exportStore.close()
  }

  onMounted(() => document.addEventListener("keydown", onKeydown))
  onUnmounted(() => document.removeEventListener("keydown", onKeydown))
  
  return {
    exportService,
    exportStore
  }
}