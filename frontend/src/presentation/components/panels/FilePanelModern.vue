<template>
  <BasePanel
    :title="$t('filePanel.title', 'File Explorer')"
    :icon="FolderIcon"
    :collapsible="true"
    :is-collapsed="isCollapsed"
    :scrollable="false"
    :loading="fileTreeStore.isLoading"
    :error="fileError"
    :resizable="true"
    :min-width="250"
    :max-width="500"
    :width="320"
    variant="primary"
    size="md"
    @toggle="handleToggle"
    @resize="handleResize"
    @retry="handleRetry"
    class="file-panel-modern"
  >
    <template #header-actions>
      <div class="file-panel-actions">
        <button
          @click="refreshFileTree"
          :disabled="fileTreeStore.isLoading"
          class="action-btn secondary"
          :title="$t('filePanel.refresh', 'Refresh file tree')"
        >
          <RefreshCwIcon class="btn-icon" :class="{ 'animate-spin': fileTreeStore.isLoading }" />
        </button>
        
        <button
          @click="toggleShowHidden"
          class="action-btn secondary"
          :class="{ active: showHidden }"
          :title="$t('filePanel.toggleHidden', 'Toggle hidden files')"
        >
          <EyeOffIcon v-if="showHidden" class="btn-icon" />
          <EyeIcon v-else class="btn-icon" />
        </button>
      </div>
    </template>

    <!-- File Tree Content -->
    <div class="file-tree-container">
      <!-- Search Bar -->
      <div class="file-search-section">
        <div class="search-input-wrapper">
          <SearchIcon class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="$t('filePanel.searchPlaceholder', 'Search files...')"
            class="search-input"
            @input="handleSearch"
          />
          <button
            v-if="searchQuery"
            @click="clearSearch"
            class="clear-search-btn"
          >
            <XIcon class="clear-icon" />
          </button>
        </div>
        
        <div v-if="searchQuery" class="search-results-info">
          {{ filteredFileCount }} {{ $t('filePanel.filesFound', 'files found') }}
        </div>
      </div>

      <!-- File Tree -->
      <div class="file-tree-content content-scrollable">
        <div v-if="!hasFiles" class="empty-state">
          <div class="empty-icon">
            <FolderOpenIcon class="icon" />
          </div>
          <p class="empty-title">{{ $t('filePanel.noFiles', 'No files found') }}</p>
          <p class="empty-description">
            {{ $t('filePanel.noFilesDescription', 'This project appears to be empty or still loading.') }}
          </p>
          <button
            @click="refreshFileTree"
            class="empty-action-btn"
          >
            <RefreshCwIcon class="btn-icon" />
            {{ $t('filePanel.refresh', 'Refresh') }}
          </button>
        </div>

        <FileTreeEnhanced
          v-else
          :nodes="filteredFileTree"
          :search-query="searchQuery"
          :show-hidden="showHidden"
          :selected-files="selectedFiles"
          :expanded-folders="expandedFolders"
          @file-select="handleFileSelect"
          @folder-toggle="handleFolderToggle"
          @file-context-menu="handleFileContextMenu"
          class="file-tree"
        />
      </div>

      <!-- Selection Summary -->
      <div v-if="selectedFiles.length > 0" class="selection-summary">
        <div class="summary-header">
          <span class="summary-title">{{ $t('filePanel.selected', 'Selected') }}</span>
          <button
            @click="clearSelection"
            class="clear-selection-btn"
            :title="$t('filePanel.clearSelection', 'Clear selection')"
          >
            <XIcon class="clear-icon" />
          </button>
        </div>
        <div class="summary-stats">
          <span class="stat-item">
            <FileIcon class="stat-icon" />
            {{ selectedFiles.length }} {{ $t('filePanel.files', 'files') }}
          </span>
          <span class="stat-item">
            <HardDriveIcon class="stat-icon" />
            {{ formatFileSize(totalSelectedSize) }}
          </span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="file-panel-footer">
        <div class="footer-stats">
          <span class="stat">
            {{ totalFileCount }} {{ $t('filePanel.totalFiles', 'files') }}
          </span>
          <span v-if="searchQuery" class="stat">
            {{ filteredFileCount }} {{ $t('filePanel.filtered', 'filtered') }}
          </span>
        </div>
        
        <div class="footer-actions">
          <button
            v-if="selectedFiles.length > 0"
            @click="buildContext"
            class="action-btn primary"
          >
            <ZapIcon class="btn-icon" />
            {{ $t('filePanel.buildContext', 'Build Context') }}
          </button>
        </div>
      </div>
    </template>
  </BasePanel>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  FolderIcon,
  FolderOpenIcon,
  RefreshCwIcon,
  EyeIcon,
  EyeOffIcon,
  SearchIcon,
  XIcon,
  FileIcon,
  HardDriveIcon,
  ZapIcon
} from 'lucide-vue-next'

import BasePanel from '@/presentation/components/BasePanel.vue'
import FileTreeEnhanced from './FileTreeEnhanced.vue'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useProjectStore } from '@/stores/project.store'
import { useContextBuilderStore } from '@/stores/context-builder.store'

interface Props {
  isCollapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  isCollapsed: false
})

const emit = defineEmits<{
  toggle: [collapsed: boolean]
  resize: [width: number]
}>()

// Stores
const fileTreeStore = useFileTreeStore()
const projectStore = useProjectStore()
const contextBuilderStore = useContextBuilderStore()

// State
const searchQuery = ref('')
const showHidden = ref(false)
const expandedFolders = ref<Set<string>>(new Set())

// Computed
const fileError = computed(() => {
  if (fileTreeStore.error) {
    return {
      message: fileTreeStore.error,
      code: 'FILE_TREE_ERROR'
    }
  }
  return null
})

const selectedFiles = computed(() => fileTreeStore.selectedFiles)
const fileTree = computed(() => fileTreeStore.fileTree)
const hasFiles = computed(() => fileTree.value.length > 0)
const totalFileCount = computed(() => fileTreeStore.totalFileCount)

const filteredFileTree = computed(() => {
  if (!searchQuery.value) return fileTree.value
  
  return filterFileTree(fileTree.value, searchQuery.value.toLowerCase())
})

const filteredFileCount = computed(() => {
  if (!searchQuery.value) return totalFileCount.value
  return countFilteredFiles(filteredFileTree.value)
})

const totalSelectedSize = computed(() => {
  return selectedFiles.value.reduce((total, file) => {
    // This would need to be implemented with actual file size data
    return total + 1024 // Placeholder
  }, 0)
})

// Methods
function handleToggle(collapsed: boolean) {
  emit('toggle', collapsed)
}

function handleResize(width: number) {
  emit('resize', width)
}

function handleRetry() {
  refreshFileTree()
}

function refreshFileTree() {
  fileTreeStore.refreshFileTree()
}

function toggleShowHidden() {
  showHidden.value = !showHidden.value
}

function handleSearch() {
  // Search is reactive through computed property
}

function clearSearch() {
  searchQuery.value = ''
}

function handleFileSelect(filePath: string, isSelected: boolean) {
  if (isSelected) {
    fileTreeStore.addSelectedFile(filePath)
  } else {
    fileTreeStore.removeSelectedFile(filePath)
  }
}

function handleFolderToggle(folderPath: string, isExpanded: boolean) {
  if (isExpanded) {
    expandedFolders.value.add(folderPath)
  } else {
    expandedFolders.value.delete(folderPath)
  }
}

function handleFileContextMenu(filePath: string, event: MouseEvent) {
  // Handle context menu
  console.log('Context menu for:', filePath, event)
}

function clearSelection() {
  fileTreeStore.clearSelection()
}

function buildContext() {
  if (projectStore.currentProject?.path) {
    contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
  }
}

function filterFileTree(nodes: any[], query: string): any[] {
  return nodes.filter(node => {
    if (node.type === 'file') {
      return node.name.toLowerCase().includes(query) || 
             node.path.toLowerCase().includes(query)
    } else {
      // For folders, include if any children match
      const filteredChildren = filterFileTree(node.children || [], query)
      return filteredChildren.length > 0 || 
             node.name.toLowerCase().includes(query)
    }
  }).map(node => {
    if (node.type === 'folder') {
      return {
        ...node,
        children: filterFileTree(node.children || [], query)
      }
    }
    return node
  })
}

function countFilteredFiles(nodes: any[]): number {
  return nodes.reduce((count, node) => {
    if (node.type === 'file') {
      return count + 1
    } else {
      return count + countFilteredFiles(node.children || [])
    }
  }, 0)
}

function formatFileSize(bytes: number): string {
  const sizes = ['B', 'KB', 'MB', 'GB']
  if (bytes === 0) return '0 B'
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
}

// Watch for project changes to refresh file tree
watch(
  () => projectStore.currentProject?.path,
  (newPath) => {
    if (newPath) {
      refreshFileTree()
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.file-panel-modern {
  min-width: 250px;
  max-width: 500px;
}

.file-panel-actions {
  display: flex;
  gap: 4px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 6px;
  color: rgb(148, 163, 184);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-btn:hover:not(:disabled) {
  background: rgba(51, 65, 85, 0.8);
  border-color: rgba(148, 163, 184, 0.4);
  color: rgb(203, 213, 225);
}

.action-btn.active {
  background: rgba(59, 130, 246, 0.2);
  border-color: rgba(59, 130, 246, 0.4);
  color: rgb(147, 197, 253);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  width: 14px;
  height: 14px;
}

.file-tree-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  min-height: 0;
}

.file-search-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  width: 14px;
  height: 14px;
  color: rgb(100, 116, 139);
  z-index: 1;
}

.search-input {
  width: 100%;
  padding: 10px 12px 10px 36px;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  color: rgb(248, 250, 252);
  font-size: 0.875rem;
  transition: all var(--transition-fast);
}

.search-input:focus {
  outline: none;
  border-color: rgb(59, 130, 246);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.search-input::placeholder {
  color: rgb(100, 116, 139);
}

.clear-search-btn {
  position: absolute;
  right: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: rgb(100, 116, 139);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.clear-search-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: rgb(239, 68, 68);
}

.clear-icon {
  width: 12px;
  height: 12px;
}

.search-results-info {
  font-size: 0.75rem;
  color: rgb(100, 116, 139);
  padding-left: 4px;
}

.file-tree-content {
  flex: 1;
  min-height: 0;
  max-height: 400px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.6);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 40px 20px;
  text-align: center;
}

.empty-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 50%;
}

.empty-icon .icon {
  width: 24px;
  height: 24px;
  color: rgb(59, 130, 246);
}

.empty-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(203, 213, 225);
  margin: 0;
}

.empty-description {
  font-size: 0.75rem;
  color: rgb(100, 116, 139);
  margin: 0;
  line-height: 1.5;
}

.empty-action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(59, 130, 246, 0.2);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 6px;
  color: rgb(147, 197, 253);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.empty-action-btn:hover {
  background: rgba(59, 130, 246, 0.3);
  border-color: rgba(59, 130, 246, 0.5);
}

.selection-summary {
  padding: 12px;
  background: rgba(168, 85, 247, 0.1);
  border: 1px solid rgba(168, 85, 247, 0.2);
  border-radius: 8px;
}

.summary-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.summary-title {
  font-size: 0.75rem;
  font-weight: 600;
  color: rgb(196, 181, 253);
}

.clear-selection-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: rgb(148, 163, 184);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.clear-selection-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: rgb(239, 68, 68);
}

.summary-stats {
  display: flex;
  gap: 12px;
  font-size: 0.75rem;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: rgb(203, 213, 225);
}

.stat-icon {
  width: 12px;
  height: 12px;
}

.file-panel-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.footer-stats {
  display: flex;
  gap: 12px;
  font-size: 0.75rem;
  color: rgb(100, 116, 139);
}

.stat {
  padding: 4px 8px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 6px;
}

.footer-actions {
  display: flex;
  gap: 8px;
}

.action-btn.primary {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  padding: 8px 12px;
  width: auto;
  height: auto;
  gap: 6px;
  font-size: 0.75rem;
  font-weight: 500;
}

.action-btn.primary:hover {
  background: linear-gradient(135deg, #2563eb, #7c3aed);
  transform: translateY(-1px);
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>