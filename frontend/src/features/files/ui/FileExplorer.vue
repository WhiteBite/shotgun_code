<template>
  <div class="h-full flex flex-col bg-transparent">
    <!-- Header -->
    <div class="border-b" style="border-color: var(--border-default)">
      <div class="panel-header-unified">
        <div class="panel-header-unified-title">
          <div class="panel-header-unified-icon panel-header-unified-icon-indigo">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"></path>
            </svg>
          </div>
          <h2>{{ t('files.title') }}</h2>
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
                  <div class="text-gray-400">Total: <span class="text-white">{{ explorer.totalFileCount.value }}</span> files</div>
                  <div class="text-gray-400">Progress: <span class="text-indigo-400">{{ explorer.selectionProgress.value }}%</span></div>
                  <div class="text-gray-400">Est. Size: <span class="text-white">{{ Math.round(fileStore.estimatedContextSize * 100) / 100 }}MB</span></div>
                  <div class="text-gray-400">Est. Tokens: <span class="text-emerald-400">~{{ Math.round(fileStore.estimatedTokenCount / 1000) }}K</span></div>
                </div>
              </div>
            </div>
            <span class="text-gray-400">{{ t('files.selected') }}</span>
          </div>
          <button @click="explorer.showSettings.value = !explorer.showSettings.value" class="p-2 rounded-lg transition-colors"
            :class="explorer.showSettings.value ? 'bg-indigo-500/20 text-indigo-400' : 'hover:bg-gray-700/50 text-gray-400 hover:text-white'"
            :title="t('files.settings')"
            :aria-label="t('files.settings')"
            :aria-expanded="explorer.showSettings.value">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>
          <button @click="explorer.handleRefresh"
            class="p-2 rounded-lg hover:bg-gray-700/50 text-gray-400 hover:text-white transition-colors"
            :title="t('files.refresh')"
            :aria-label="t('files.refresh')">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Breadcrumbs -->
      <div v-if="fileStore.breadcrumbs.length > 0" class="px-3 pb-3">
        <BreadcrumbsNav 
          :segments="fileStore.breadcrumbs" 
          :root-name="fileStore.projectName"
          @navigate="handleBreadcrumbNavigate"
          @open-in-explorer="handleOpenInExplorer"
        />
      </div>
    </div>

    <!-- Settings Panel -->
    <Transition name="settings-panel">
      <div v-if="explorer.showSettings.value" class="border-b border-gray-700 bg-gray-800/50 p-3 space-y-2 settings-panel">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-semibold text-gray-300">{{ t('settings.title') }}</span>
        </div>

        <label class="flex items-center gap-2 text-xs text-gray-300 cursor-pointer hover:text-white transition-colors">
          <input v-model="settingsStore.settings.fileExplorer.useGitignore" type="checkbox"
            class="w-3.5 h-3.5 rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-0"
            @change="explorer.handleSettingsChange" />
          {{ t('settings.useGitignore') }}
        </label>

        <label class="flex items-center gap-2 text-xs text-gray-300 cursor-pointer hover:text-white transition-colors">
          <input v-model="settingsStore.settings.fileExplorer.useCustomIgnore" type="checkbox"
            class="w-3.5 h-3.5 rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-0"
            @change="explorer.handleSettingsChange" />
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

        <label class="flex items-center gap-2 text-xs text-gray-300 cursor-pointer hover:text-white transition-colors">
          <input v-model="settingsStore.settings.fileExplorer.foldersFirst" type="checkbox"
            class="w-3.5 h-3.5 rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-0" />
          {{ t('settings.foldersFirst') }}
        </label>

        <div class="flex gap-2">
          <button @click="ignoreRulesModalRef?.open()" class="settings-btn">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
            </svg>
            {{ t('settings.ignoreRules') }}
          </button>
          <button @click="presetsModalRef?.open()" class="settings-btn">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
            </svg>
            {{ t('settings.presets') }}
          </button>
        </div>

        <!-- Selection History Indicator -->
        <div v-if="explorer.hasSelectionHistory.value" class="mt-2 pt-2 border-t border-gray-700">
          <button @click="explorer.restorePreviousSelection" class="selection-indicator">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            {{ t('settings.restoreSelection').replace('{count}', explorer.selectionHistoryCount.value.toString()) }}
          </button>
        </div>
      </div>
    </Transition>

    <!-- Quick Filters -->
    <QuickFiltersBar />

    <!-- Search -->
    <div class="p-2 border-b" style="border-color: var(--border-default)">
      <div class="relative group">
        <input 
          ref="searchInputRef"
          v-model="explorer.searchQuery.value" 
          type="text" 
          :placeholder="t('files.searchShort')" 
          :aria-label="t('files.searchShort')"
          class="search-input pr-8"
          :class="{ 'search-input-active': explorer.searchQuery.value }"
          @input="explorer.handleSearch" 
          @keydown.escape="clearSearch"
        />
        <svg class="input-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <!-- Clear button -->
        <button 
          v-if="explorer.searchQuery.value"
          @click="clearSearch"
          class="search-clear-btn"
          :title="t('files.clear')"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- File Tree -->
    <div class="flex-1 min-h-0 scrollable-y p-2" data-tour="file-tree">
      <div v-if="fileStore.isLoading" class="flex items-center justify-center h-full">
        <svg class="loading-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
      </div>

      <div v-else-if="fileStore.nodes.length === 0" class="empty-state h-full">
        <p class="empty-state-text">{{ t('files.noFiles') }}</p>
      </div>

      <!-- Hint when tree is collapsed -->
      <div v-else-if="isTreeCollapsed" class="tree-hint">
        <p class="tree-hint-text">{{ t('tree.expandHint') }}</p>
      </div>

      <!-- Search Results Header -->
      <div v-if="explorer.searchQuery.value && fileStore.searchResults.length > 0" class="text-xs text-gray-400 mb-2 px-2">
        {{ fileStore.searchResults.length }} {{ t('files.results') }}
      </div>

      <!-- Virtualized File Tree (for both normal view and search results) -->
      <VirtualFileTree
        v-if="explorer.searchQuery.value ? fileStore.searchResults.length > 0 : true"
        :nodes="explorer.searchQuery.value ? fileStore.searchResults : fileStore.filteredNodes"
        :compact-mode="settingsStore.settings.fileExplorer.compactNestedFolders"
        @toggle-select="explorer.handleToggleSelect"
        @toggle-expand="explorer.handleToggleExpand"
        @contextmenu="handleContextMenu"
        @quicklook="explorer.handleQuickLook"
      />
    </div>

    <!-- Context Menu -->
    <FileContextMenu :node="contextMenu.targetNode.value" :position="contextMenu.position.value"
      :visible="contextMenu.isVisible.value" @action="explorer.handleContextMenuAction" @close="contextMenu.hide" />

    <!-- Modals -->
    <IgnoreRulesModal ref="ignoreRulesModalRef" />
    <SelectionPresetsModal ref="presetsModalRef" />
    <QuickLookModal v-model="explorer.quickLookVisible.value" :file-path="explorer.quickLookPath.value" @add-to-context="explorer.handleAddToContext" />

    <!-- Bottom Panel -->
    <div class="border-t p-2 bg-gray-900/30 space-y-2" style="border-color: var(--border-default)">
      <button 
        data-tour="build-button"
        @click="$emit('build-context')" 
        :disabled="fileStore.selectedPaths.size === 0 || contextStore.isBuilding"
        :title="fileStore.selectedPaths.size === 0 ? t('context.selectFilesFirst') : t('context.buildTooltip')"
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
      <!-- Hint when no files selected -->
      <p v-if="fileStore.selectedPaths.size === 0" class="text-xs text-gray-400 text-center">
        {{ t('context.selectFilesHint') }}
      </p>

      <div class="flex gap-2">
        <button @click="handleClearSelection" :disabled="!fileStore.hasSelectedFiles"
          class="flex-1 btn-unified btn-unified-secondary py-2">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          {{ t('files.clear') }}
        </button>
        <button @click="fileStore.expandAll" class="btn-unified btn-unified-ghost btn-unified-icon" :title="t('files.expandAll')" :aria-label="t('files.expandAll')">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        <button @click="fileStore.collapseAll" class="btn-unified btn-unified-ghost btn-unified-icon" :title="t('files.collapseAll')" :aria-label="t('files.collapseAll')">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useConfirm } from '@/composables/useConfirm'
import { useContextMenu } from '@/composables/useContextMenu'
import { useI18n } from '@/composables/useI18n'
import { useLogger } from '@/composables/useLogger'
import { useContextStore } from '@/features/context'
import { useProjectStore } from '@/stores/project.store'
import { useSettingsStore } from '@/stores/settings.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref, watch } from 'vue'
import { useFileExplorer } from '../composables/useFileExplorer'
import { provideHoveredFile } from '../composables/useHoveredFile'
import { useFileStore, type FileNode } from '../model/file.store'
import BreadcrumbsNav from './BreadcrumbsNav.vue'
import FileContextMenu from './FileContextMenu.vue'
import QuickFiltersBar from './QuickFiltersBar.vue'
import VirtualFileTree from './VirtualFileTree.vue'

// Lazy-loaded modals (loaded on demand)
const QuickLookModal = defineAsyncComponent(() => import('@/components/QuickLookModal.vue'))
const IgnoreRulesModal = defineAsyncComponent(() => import('./IgnoreRulesModal.vue'))
const SelectionPresetsModal = defineAsyncComponent(() => import('./SelectionPresetsModal.vue'))

// IMPORTANT: provideHoveredFile MUST be called first in setup, before any other composables
// This provides the hovered file state to child components via Vue's provide/inject
provideHoveredFile()

// Stores
const fileStore = useFileStore()
const contextStore = useContextStore()
const projectStore = useProjectStore()
const uiStore = useUIStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()

// Composables
const explorer = useFileExplorer()
const contextMenu = useContextMenu()
const { confirm } = useConfirm()
const logger = useLogger('FileExplorer')

// Refs
const ignoreRulesModalRef = ref<InstanceType<typeof IgnoreRulesModal>>()
const presetsModalRef = ref<InstanceType<typeof SelectionPresetsModal>>()
const searchInputRef = ref<HTMLInputElement | null>(null)

function clearSearch() {
  explorer.clearSearch()
}

async function handleClearSelection() {
  const count = fileStore.selectedPaths.size
  // Show confirmation only if many files selected
  if (count > 5) {
    const confirmed = await confirm({
      title: t('files.clearConfirmTitle'),
      message: t('files.clearConfirmMessage').replace('{count}', count.toString()),
      confirmText: t('common.clear'),
      cancelText: t('common.cancel'),
      variant: 'warning',
    })
    if (!confirmed) return
  }
  fileStore.clearSelection()
}

// Check if all root folders are collapsed
const isTreeCollapsed = computed(() => {
  if (fileStore.nodes.length === 0) return false
  // Check if any root node is expanded
  return !fileStore.nodes.some(node => node.isDir && node.isExpanded)
})

// Expose refs for template binding
defineExpose({ searchInputRef })



defineEmits<{
  (e: 'preview-file', filePath: string): void
  (e: 'build-context'): void
}>()

function handleContextMenu(node: FileNode, event: MouseEvent) {
  contextMenu.show(node, event)
}

function handleBreadcrumbNavigate(path: string) {
  // Expand the path in the tree and scroll to it
  fileStore.expandPath(path)
}

async function handleOpenInExplorer(_path: string) {
  // Open the project folder in system file explorer
  if (projectStore.currentPath) {
    try {
      // Use Wails runtime to open URL - dynamic import to avoid build issues
      const runtime = await import('#wailsjs/runtime/runtime')
      runtime.BrowserOpenURL('file://' + projectStore.currentPath)
    } catch (error) {
      logger.error('Failed to open in explorer:', error)
      uiStore.addToast('Failed to open in file explorer', 'error')
    }
  }
}

onMounted(async () => {
  explorer.initialize()

  try {
    if (!projectStore.currentPath) {
      uiStore.addToast('No project selected', 'warning')
      return
    }
    await fileStore.loadFileTree(projectStore.currentPath)
  } catch (error) {
    logger.error('Failed to load file tree:', error)
    uiStore.addToast('Failed to load project files. Please try again.', 'error')
  }
})

watch(() => projectStore.currentPath, async (newPath, oldPath) => {
  if (newPath && newPath !== oldPath) {
    fileStore.clearSelection()
    contextStore.clearContext()
    try {
      await fileStore.loadFileTree(newPath)
    } catch (error) {
      logger.error('Failed to load file tree after project change:', error)
      uiStore.addToast('Failed to load project files', 'error')
    }
  }
})

onUnmounted(() => {
  explorer.cleanup()
})
</script>
