import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useWorkspaceStore } from '@/stores/workspace.store'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useSettingsStore } from '@/stores/settings.store'
import { useUiStore } from '@/stores/ui.store'
import { StatusBarService } from '@/domain/services/StatusBarService'

export function useStatusBarService() {
  // Get stores
  const workspaceStore = useWorkspaceStore()
  const contextStore = useContextBuilderStore()
  const settingsStore = useSettingsStore()
  const uiStore = useUiStore()
  
  // Get reactive refs
  const { buildStatus, contextMetrics, selectedFilesCount, lastContextGeneration } = storeToRefs(contextStore)
  const { toasts } = storeToRefs(uiStore)
  
  // Create service instance
  const statusBarService = new StatusBarService(
    () => settingsStore.settings?.aiProvider || null,
    () => buildStatus.value,
    () => contextMetrics.value,
    () => selectedFilesCount.value,
    () => lastContextGeneration.value,
    () => toasts.value.length
  )
  
  return {
    statusBarService,
    contextStore,
    settingsStore,
    uiStore
  }
}