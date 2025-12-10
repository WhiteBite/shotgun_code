<template>
  <div class="h-full flex flex-col bg-transparent">
    <!-- Header -->
    <div class="border-b border-gray-700/30">
      <div class="flex items-center justify-between p-3">
        <div class="section-title">
          <div class="section-icon section-icon-indigo">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"></path>
            </svg>
          </div>
          <h2 class="section-title-text">{{ t('files.title') }}</h2>
        </div>

        <div class="flex items-center gap-2">
          <div class="flex items-center gap-2 text-xs">
            <div class="relative group">
              <span class="chip-unified chip-unified-accent cursor-help">
                {{ fileStore.selectedCount }}
              </span>
              <!-- Tooltip with stats -->
              <div
                class="absolute right-0 top-full mt-2 hidden group-hover:block z-50 px-4 py-3 bg-gray-800/95 backdrop-blur-sm border border-gray-700/50 rounded-xl shadow-2xl whitespace-nowrap">
                <div class="text-xs space-y-1.5">
                  <div class="text-white font-semibold mb-2">Selection Stats</div>
                  <div class="text-gray-400">Files: <span class="text-white">{{ fileStore.selectedCount }}</span></div>
                  <div class="text-gray-400">Total: <span class="text-white">{{ totalFileCount }}</span> files</div>
                  <div class="text-gray-400">Progress: <span class="text-indigo-400">{{ selectionProgress }}%</span>
                  </div>
                  <div class="text-gray-400">Est. Size: <span class="text-white">{{
                    Math.round(fileStore.estimatedContextSize * 100) / 100 }}MB</span></div>
                  <div class="text-gray-400">Est. Tokens: <span class="text-emerald-400">~{{
                    Math.round(fileStore.estimatedTokenCount / 1000) }}K</span></div>
                </div>
              </div>
            </div>
            <span class="text-gray-400">{{ t('files.selected') }}</span>
          </div>
          <button @click="showSettings = !showSettings" class="p-2 rounded-lg transition-colors"
            :class="showSettings ? 'bg-indigo-500/20 text-indigo-400' : 'hover:bg-gray-700/50 text-gray-400 hover:text-white'"
            title="Settings">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>
          <button @click="handleRefresh"
            class="p-2 rounded-lg hover:bg-gray-700/50 text-gray-400 hover:text-white transition-colors"
            :title="t('files.refresh')">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Breadcrumbs -->
      <div v-if="fileStore.breadcrumbs.length > 0" class="px-3 pb-3">
        <BreadcrumbsNav :segments="fileStore.breadcrumbs" :root-name="fileStore.projectName" />
      </div>
    </div>

    <!-- Settings Panel with animation -->
    <Transition name="settings-panel">
      <div v-if="showSettings" class="border-b border-gray-700 bg-gray-800/50 p-3 space-y-2 settings-panel">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-semibold text-gray-300">{{ t('settings.title') }}</span>
        </div>

        <label class="flex items-center gap-2 text-xs text-gray-300 cursor-pointer hover:text-white transition-colors">
          <input v-model="settingsStore.settings.fileExplorer.useGitignore" type="checkbox"
            class="w-3.5 h-3.5 rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-0"
            @change="handleSettingsChange" />
          {{ t('settings.useGitignore') }}
        </label>

        <label class="flex items-center gap-2 text-xs text-gray-300 cursor-pointer hover:text-white transition-colors">
          <input v-model="settingsStore.settings.fileExplorer.useCustomIgnore" type="checkbox"
            class="w-3.5 h-3.5 rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-0"
            @change="handleSettingsChange" />
          {{ t('settings.useCustomIgnore') }}
        </label>

        <label class="flex items-center gap-2 text-xs text-gray-300 cursor-pointer hover:text-white transition-colors">
          <input v-model="settingsStore.settings.fileExplorer.autoSaveSelection" type="checkbox"
            class="w-3.5 h-3.5 rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-0" />
          {{ t('settings.autoSaveSelection') }}
        </label>

        <label class="flex items-center gap-2 text-xs text-gray-300 cursor-pointer hover:text-white transition-colors">
          <input v-model="settingsStore.settings.fileExplorer.compactNestedFolders" type="checkbox"
            class="w-3.5 h-3.5 rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-0" />
          {{ t('settings.compactFolders') }}
        </label>

        <div class="flex gap-2">
          <button @click="openIgnoreRulesModal" class="settings-btn">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
            </svg>
            {{ t('settings.ignoreRules') }}
          </button>
          <button @click="openPresetsModal" class="settings-btn">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
            </svg>
            {{ t('settings.presets') }}
          </button>
        </div>

        <!-- Selection History Indicator with animation -->
        <div v-if="hasSelectionHistory" class="mt-2 pt-2 border-t border-gray-700">
          <button @click="restorePreviousSelection" class="selection-indicator">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            {{ t('settings.restoreSelection').replace('{count}', selectionHistoryCount.toString()) }}
          </button>
        </div>
      </div>
    </Transition>

    <!-- Quick Filters -->
    <QuickFiltersBar />

    <!-- Search and Filter -->
    <div class="p-2 border-b border-gray-700/30 space-y-2">
      <div class="relative">
        <input v-model="searchQuery" type="text" :placeholder="t('files.search')" class="search-input"
          @input="handleSearch" />
        <svg class="input-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>
      <FileFilterDropdown v-if="availableExtensions.length > 0" :extensions="availableExtensions"
        v-model:selected="filterExtensions" @update:selected="handleFilterUpdate" />
    </div>

    <!-- File Tree -->
    <div class="flex-1 scrollable-y p-2">
      <div v-if="fileStore.isLoading" class="flex items-center justify-center h-full">
        <svg class="loading-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
      </div>

      <div v-else-if="fileStore.nodes.length === 0" class="empty-state h-full">
        <p class="empty-state-text">{{ t('files.noFiles') }}</p>
      </div>

      <div v-else-if="searchQuery && fileStore.searchResults.length > 0">
        <div class="text-xs text-gray-400 mb-2 px-2">
          {{ fileStore.searchResults.length }} {{ t('files.results') }}
        </div>
        <FileTreeNode v-for="node in fileStore.searchResults" :key="node.path" :node="node"
          :compact-mode="settingsStore.settings.fileExplorer.compactNestedFolders" @toggle-select="handleToggleSelect"
          @toggle-expand="handleToggleExpand" @contextmenu="handleContextMenu" @quicklook="handleQuickLook" />
      </div>

      <!-- TEMPORARY: Limit rendering to first 50 nodes to test memory leak -->
      <div v-if="fileStore.filteredNodes.length > 50" class="text-xs text-yellow-400 mb-2 px-2">
        Warning: Showing only first 50 of {{ fileStore.filteredNodes.length }} items (memory optimization test)
      </div>
      <FileTreeNode v-else v-for="node in fileStore.filteredNodes.slice(0, 50)" :key="node.path" :node="node"
        :compact-mode="settingsStore.settings.fileExplorer.compactNestedFolders" @toggle-select="handleToggleSelect"
        @toggle-expand="handleToggleExpand" @contextmenu="handleContextMenu" @quicklook="handleQuickLook" />
    </div>

    <!-- Context Menu -->
    <FileContextMenu :node="contextMenu.targetNode.value" :position="contextMenu.position.value"
      :visible="contextMenu.isVisible.value" @action="handleContextMenuAction" @close="contextMenu.hide" />

    <!-- Ignore Rules Modal -->
    <IgnoreRulesModal ref="ignoreRulesModalRef" />

    <!-- Selection Presets Modal -->
    <SelectionPresetsModal ref="presetsModalRef" />

    <!-- QuickLook Modal -->
    <QuickLookModal v-model="quickLookVisible" :file-path="quickLookPath" @add-to-context="handleAddToContext" />

    <!-- Bottom Panel - Compact Actions -->
    <div class="border-t border-gray-700/30 p-2 bg-gray-900/30 space-y-2">
      <!-- Build Context Button -->
      <button @click="$emit('build-context')" :disabled="fileStore.selectedPaths.size === 0 || contextStore.isBuilding"
        class="w-full btn-unified btn-unified-primary py-2.5">
        <svg v-if="!contextStore.isBuilding" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <svg v-else class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        {{ contextStore.isBuilding ? t('context.building') : t('context.build') }}
        <span v-if="fileStore.selectedPaths.size > 0" class="opacity-75">({{ fileStore.selectedPaths.size }})</span>
      </button>

      <!-- Clear & Expand/Collapse -->
      <div class="flex gap-2">
        <button @click="fileStore.clearSelection" :disabled="!fileStore.hasSelectedFiles"
          class="flex-1 btn-unified btn-unified-secondary py-2">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          {{ t('files.clear') }}
        </button>
        <button @click="fileStore.expandAll" class="btn-unified btn-unified-ghost btn-unified-icon"
          :title="t('files.expandAll')">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        <button @click="fileStore.collapseAll" class="btn-unified btn-unified-ghost btn-unified-icon"
          :title="t('files.collapseAll')">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import QuickLookModal from '@/components/QuickLookModal.vue'
import { useContextMenu } from '@/composables/useContextMenu'
import { useI18n } from '@/composables/useI18n'
import { useContextStore } from '@/features/context'
import { apiService } from '@/services/api.service'
import { useProjectStore } from '@/stores/project.store'
import { useSettingsStore } from '@/stores/settings.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { filesApi } from '../api/files.api'
import { copyToClipboard, getRelativePath } from '../lib/file-utils'
import { useFileStore, type FileNode } from '../model/file.store'
import BreadcrumbsNav from './BreadcrumbsNav.vue'
import FileContextMenu from './FileContextMenu.vue'
import FileFilterDropdown from './FileFilterDropdown.vue'
import FileTreeNode from './FileTreeNode.vue'
import IgnoreRulesModal from './IgnoreRulesModal.vue'
import QuickFiltersBar from './QuickFiltersBar.vue'
import SelectionPresetsModal from './SelectionPresetsModal.vue'

const fileStore = useFileStore()
const contextStore = useContextStore()
const projectStore = useProjectStore()
const uiStore = useUIStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()
const contextMenu = useContextMenu()
const searchQuery = ref('')
const showSettings = ref(false)
const ignoreRulesModalRef = ref<InstanceType<typeof IgnoreRulesModal>>()
const presetsModalRef = ref<InstanceType<typeof SelectionPresetsModal>>()
const filterExtensions = ref<string[]>([])

// QuickLook state
const quickLookVisible = ref(false)
const quickLookPath = ref('')

// Declare global window properties for hover tracking
declare global {
  interface Window {
    __hoveredFilePath: string | null
    __hoveredFileIsDir: boolean | null
  }
}

const availableExtensions = computed(() => fileStore.getAvailableExtensions())

const hasSelectionHistory = computed(() => {
  if (!projectStore.currentPath) return false
  const stats = fileStore.getSelectionStats()
  return stats[projectStore.currentPath] > 0
})

const selectionHistoryCount = computed(() => {
  if (!projectStore.currentPath) return 0
  const stats = fileStore.getSelectionStats()
  return stats[projectStore.currentPath] || 0
})

// Cache total file count to avoid recalculating on every render
const totalFileCount = ref(0)

watch(() => fileStore.nodes, (newNodes) => {
  let count = 0
  const countFiles = (nodes: any[]) => {
    nodes.forEach(node => {
      if (!node.isDir) count++
      if (node.children) countFiles(node.children)
    })
  }
  countFiles(newNodes)
  totalFileCount.value = count
}, { immediate: true, deep: false })

const selectionProgress = computed(() => {
  if (totalFileCount.value === 0) return 0
  return Math.round((fileStore.selectedCount / totalFileCount.value) * 100)
})

defineEmits<{
  (e: 'preview-file', filePath: string): void
  (e: 'build-context'): void
}>()

function handleToggleSelect(path: string) {
  fileStore.toggleSelect(path)
}

function handleToggleExpand(path: string) {
  fileStore.toggleExpand(path)
}

function handleSearch() {
  fileStore.setSearchQuery(searchQuery.value)
}



async function handleRefresh() {
  if (!projectStore.currentPath) return

  try {
    fileStore.clearSelection()
    await fileStore.loadFileTree(projectStore.currentPath)
    uiStore.addToast(t('toast.refreshed'), 'success')
  } catch (error) {
    console.error('Failed to refresh file tree:', error)
    uiStore.addToast(t('toast.refreshError'), 'error')
  }
}

// Refresh file tree while preserving expanded folders state
async function handleRefreshPreserveState() {
  if (!projectStore.currentPath) return

  try {
    // Save current expanded state
    const expandedPaths = fileStore.getExpandedPaths()
    const selectedPaths = fileStore.selectedFilesList

    // Clear backend cache first (important when ignore rules change)
    await apiService.clearFileTreeCache()
    
    // Clear frontend cache
    filesApi.clearCache()

    // Reload tree
    await fileStore.loadFileTree(projectStore.currentPath)

    // Restore expanded state
    fileStore.restoreExpandedPaths(expandedPaths)

    // Restore selection (only for paths that still exist)
    for (const path of selectedPaths) {
      if (fileStore.nodeExists(path)) {
        fileStore.toggleSelect(path)
      }
    }
  } catch (error) {
    console.error('Failed to refresh file tree:', error)
    uiStore.addToast(t('toast.refreshError'), 'error')
  }
}

function handleFilterUpdate(selected: string[]) {
  fileStore.setFilterExtensions(selected)
}

async function handleSettingsChange() {
  // Save settings to backend
  try {
    const dto = await apiService.getSettings()
    dto.useGitignore = settingsStore.settings.fileExplorer.useGitignore
    dto.useCustomIgnore = settingsStore.settings.fileExplorer.useCustomIgnore
    await apiService.saveSettings(JSON.stringify(dto))

    // Reload file tree when gitignore/custom ignore settings change
    if (projectStore.currentPath) {
      await handleRefresh()
    }
  } catch (error) {
    console.error('Failed to save settings:', error)
    uiStore.addToast('Failed to save settings', 'error')
  }
}

function openIgnoreRulesModal() {
  ignoreRulesModalRef.value?.open()
}

function openPresetsModal() {
  presetsModalRef.value?.open()
}

function restorePreviousSelection() {
  if (projectStore.currentPath) {
    fileStore.loadSelectionFromStorage(projectStore.currentPath)
    uiStore.addToast('Selection restored', 'success')
  }
}

function handleContextMenu(node: FileNode, event: MouseEvent) {
  contextMenu.show(node, event)
}

function handleQuickLook(path: string) {
  // Toggle: if same file is already open, close it
  if (quickLookVisible.value && quickLookPath.value === path) {
    quickLookVisible.value = false
    return
  }
  quickLookPath.value = path
  quickLookVisible.value = true
}

// Global Space key handler for hover-based QuickLook
function handleGlobalKeydown(event: KeyboardEvent) {
  // Only handle Space key
  if (event.key !== ' ') return
  
  // Don't handle if typing in an input
  const target = event.target as HTMLElement
  if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
    return
  }
  
  // Check if there's a hovered file
  const hoveredPath = window.__hoveredFilePath
  const isDir = window.__hoveredFileIsDir
  
  if (hoveredPath) {
    event.preventDefault()
    
    if (isDir) {
      // For folders, toggle expand
      fileStore.toggleExpand(hoveredPath)
    } else {
      // For files, toggle QuickLook
      handleQuickLook(hoveredPath)
    }
  } else if (quickLookVisible.value) {
    // If no hover but QuickLook is open, close it
    event.preventDefault()
    quickLookVisible.value = false
  }
}

function handleAddToContext(path: string) {
  fileStore.toggleSelect(path)
  uiStore.addToast(t('files.addToContext'), 'success')
}

async function handleContextMenuAction(payload: { type: string; node: FileNode }) {
  const { type, node } = payload

  try {
    switch (type) {
      case 'quickLook':
        if (!node.isDir) {
          handleQuickLook(node.path)
        }
        break

      case 'selectAll':
        if (node.isDir) {
          fileStore.selectRecursive(node.path)
          uiStore.addToast('All files selected in folder', 'success')
        }
        break

      case 'deselectAll':
        if (node.isDir) {
          fileStore.deselectRecursive(node.path)
          uiStore.addToast('All files deselected in folder', 'success')
        }
        break

      case 'copyPath':
        await copyToClipboard(node.path)
        uiStore.addToast('Path copied to clipboard', 'success')
        break

      case 'copyRelativePath':
        const relativePath = projectStore.currentPath
          ? getRelativePath(node.path, projectStore.currentPath)
          : node.path
        await copyToClipboard(relativePath)
        uiStore.addToast('Relative path copied to clipboard', 'success')
        break

      case 'addToCustomIgnore':
        try {
          const currentRules = settingsStore.getCustomIgnoreRules()
          const newRule = node.isDir ? `${node.name}/` : node.name
          const updatedRules = currentRules ? `${currentRules}\n${newRule}` : newRule
          await apiService.updateCustomIgnoreRules(updatedRules)
          settingsStore.setCustomIgnoreRules(updatedRules)
          
          // Fast path: remove node from tree without full reload
          fileStore.removeNode(node.path)
          // Also invalidate backend cache for next full reload
          await apiService.clearFileTreeCache()
          filesApi.clearCache()
          
          uiStore.addToast('Добавлено в исключения', 'success')
        } catch (error) {
          console.error('Failed to add to ignore:', error)
          uiStore.addToast('Ошибка добавления в исключения', 'error')
        }
        break

      case 'removeFromIgnore':
        try {
          const currentIgnoreRules = settingsStore.getCustomIgnoreRules()
          const pattern = node.isDir ? `${node.name}/` : node.name

          // Remove all lines that contain this pattern
          const lines = currentIgnoreRules.split('\n').filter(line => {
            const trimmed = line.trim()
            return trimmed && !trimmed.includes(pattern) && !trimmed.startsWith('#')
          })

          const updatedRules = lines.join('\n')
          await apiService.updateCustomIgnoreRules(updatedRules)
          settingsStore.setCustomIgnoreRules(updatedRules)
          uiStore.addToast('Удалено из исключений', 'success')
          // Refresh without collapsing - preserve expanded state
          await handleRefreshPreserveState()
        } catch (error) {
          console.error('Failed to remove from ignore:', error)
          uiStore.addToast('Ошибка удаления из исключений', 'error')
        }
        break

      case 'expandAll':
        if (node.isDir) {
          fileStore.expandRecursive(node.path)
          uiStore.addToast('Expanded all folders', 'success')
        }
        break

      case 'collapseAll':
        if (node.isDir) {
          fileStore.collapseRecursive(node.path)
          uiStore.addToast('Collapsed all folders', 'success')
        }
        break
    }
  } catch (error) {
    console.error('Context menu action failed:', error)
    uiStore.addToast('Action failed', 'error')
  }
}

onMounted(async () => {
  // Register global Space key handler
  window.addEventListener('keydown', handleGlobalKeydown)
  
  // Initialize hover tracking
  window.__hoveredFilePath = null
  window.__hoveredFileIsDir = null

  try {
    // CRITICAL FIX: Use projectStore.currentPath instead of getCurrentDirectory()
    if (!projectStore.currentPath) {
      uiStore.addToast('No project selected', 'warning')
      return
    }

    await fileStore.loadFileTree(projectStore.currentPath)
  } catch (error) {
    console.error('Failed to load file tree:', error)
    uiStore.addToast('Failed to load project files. Please try again.', 'error')
  }
})

// Watch for project changes and reload file tree
watch(() => projectStore.currentPath, async (newPath, oldPath) => {
  if (newPath && newPath !== oldPath) {
    // Project changed, reload file tree
    fileStore.clearSelection()
    contextStore.clearContext()
    try {
      await fileStore.loadFileTree(newPath)
    } catch (error) {
      console.error('Failed to load file tree after project change:', error)
      uiStore.addToast('Failed to load project files', 'error')
    }
  }
})

onUnmounted(() => {
  // Remove global Space key handler
  window.removeEventListener('keydown', handleGlobalKeydown)
  
  // Cleanup hover tracking
  window.__hoveredFilePath = null
  window.__hoveredFileIsDir = null
})
</script>
