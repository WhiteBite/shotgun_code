import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useUiStore } from '@/stores/ui.store'
import usePanelManager from '@/composables/usePanelManager'
import { FilePanelService } from '@/domain/services/FilePanelService'

export function useFilePanelService() {
  // Get stores
  const fileTreeStore = useFileTreeStore()
  const uiStore = useUiStore()
  
  // Get reactive refs from stores
  const { 
    selectedFiles,
    totalFiles,
    useGitignore,
    useCustomIgnore
  } = storeToRefs(fileTreeStore)
  
  // Composables
  const { panelClasses, panelStyle, isCollapsed, toggleCollapse } = usePanelManager('files', 320)
  
  // Create service instance
  const filePanelService = new FilePanelService()
  
  // Computed properties that depend on store state
  const selectionPercent = computed(() => {
    return filePanelService.calculateSelectionPercent(selectedFiles.value.length, totalFiles.value || 1)
  })
  
  const estimatedTokens = computed(() => {
    return filePanelService.estimateTokens(selectedFiles.value)
  })
  
  const selectedChars = computed(() => {
    return filePanelService.calculateSelectedChars(selectedFiles.value)
  })
  
  // Methods
  async function updateIgnoreRules() {
    await filePanelService.updateIgnoreRules(() => fileTreeStore.refreshFiles())
  }
  
  function openIgnoreDrawer() {
    filePanelService.openIgnoreDrawer((drawer: string) => uiStore.openDrawer(drawer))
  }
  
  return {
    filePanelService,
    fileTreeStore,
    uiStore,
    // Store refs
    selectedFiles,
    totalFiles,
    useGitignore,
    useCustomIgnore,
    // Computed properties
    panelClasses,
    panelStyle,
    isCollapsed,
    toggleCollapse,
    selectionPercent,
    estimatedTokens,
    selectedChars,
    // Methods
    updateIgnoreRules,
    openIgnoreDrawer
  }
}