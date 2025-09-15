import { ref, onMounted, onUnmounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useGitStore } from '@/stores/git.store'
import { useUiStore } from '@/stores/ui.store'
import { useProjectStore } from '@/stores/project.store'
import { CommitHistoryService } from '@/domain/services/CommitHistoryService'

export function useCommitHistoryService() {
  // Get stores
  const gitStore = useGitStore()
  const uiStore = useUiStore()
  const projectStore = useProjectStore()
  
  // Get reactive refs
  const { commits, isLoading, isHistoryVisible } = storeToRefs(gitStore)
  const { currentProject } = storeToRefs(projectStore)
  
  // Create service instance
  const commitHistoryService = new CommitHistoryService(
    () => commits.value,
    () => isLoading.value,
    () => isHistoryVisible.value,
    () => currentProject.value?.path || null,
    () => gitStore.hideHistory(),
    (selectedCommits) => gitStore.applyCommitSelection(selectedCommits),
    (params) => uiStore.showQuickLook(params),
    () => uiStore.hideQuickLook()
  )
  
  // Keyboard event handling
  const onKeydown = (e: KeyboardEvent) => {
    if (e.key === "Escape") gitStore.hideHistory()
  }

  onMounted(() => document.addEventListener("keydown", onKeydown))
  onUnmounted(() => document.removeEventListener("keydown", onKeydown))
  
  return {
    commitHistoryService,
    gitStore,
    uiStore,
    projectStore
  }
}