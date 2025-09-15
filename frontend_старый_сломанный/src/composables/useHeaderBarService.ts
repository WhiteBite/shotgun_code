import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useWorkspaceStore } from '@/stores/workspace.store'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useProjectStore } from '@/stores/project.store'
import { HeaderBarService } from '@/domain/services/HeaderBarService'

export function useHeaderBarService() {
  // Get stores
  const workspaceStore = useWorkspaceStore()
  const contextStore = useContextBuilderStore()
  const projectStore = useProjectStore()
  
  // Get reactive refs
  const { isManualMode, isTransitioning, layoutClasses } = storeToRefs(workspaceStore)
  const { buildStatus, contextMetrics } = storeToRefs(contextStore)
  const { currentProject } = storeToRefs(projectStore)
  
  // Create service instance
  const headerBarService = new HeaderBarService(
    () => currentProject.value,
    () => contextStore.selectedFilesCount,
    () => buildStatus.value,
    () => contextMetrics.value,
    () => isManualMode.value,
    () => isTransitioning.value,
    () => layoutClasses.value
  )
  
  return {
    headerBarService,
    workspaceStore,
    contextStore,
    projectStore
  }
}