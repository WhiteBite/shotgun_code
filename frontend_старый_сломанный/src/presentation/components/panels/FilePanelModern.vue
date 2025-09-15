<template>
  <div class="file-panel-wrapper">
    <BasePanel
      title="File Explorer"
      :icon="FolderIcon"
      :collapsible="true"
      :is-collapsed="isCollapsed"
      :scrollable="false"
      :loading="isLoading"
      :error="fileError"
      :resizable="true"
      :min-width="250"
      :max-width="500"
      :width="320"
      variant="primary"
      size="md"
      class="file-panel-modern"
      @toggle="handleToggle"
      @resize="handleResize"
      @retry="handleRetry"
    >
      <template #header-actions>
        <div class="file-panel-actions">
          <button
            :disabled="isLoading"
            class="action-btn secondary"
            :title="'Refresh file tree'"
            @click="refreshFileTree"
          >
            <RefreshCwIcon class="btn-icon" :class="{ 'animate-spin': isLoading }" />
          </button>
          
          <button
            class="action-btn secondary"
            :class="{ active: showHidden }"
            :title="'Toggle hidden files'"
            @click="toggleShowHidden"
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
              :placeholder="'Search files...'"
              class="search-input"
              @input="handleSearch"
            />
            <button
              v-if="searchQuery"
              class="clear-search-btn"
              @click="clearSearch"
            >
              <XIcon class="clear-icon" />
            </button>
          </div>
          
          <div v-if="searchQuery" class="search-results-info">
            {{ filteredFileCount }} files found
          </div>
        </div>

        <!-- File Tree -->
        <div class="file-tree-content content-scrollable">
          <div v-if="!hasFiles" class="empty-state">
            <div class="empty-icon">
              <FolderOpenIcon class="icon" />
            </div>
            <p class="empty-title">No files found</p>
            <p class="empty-description">
              This project appears to be empty or still loading.
            </p>
            <button
              class="empty-action-btn"
              @click="refreshFileTree"
            >
              <RefreshCwIcon class="btn-icon" />
              Refresh
            </button>
          </div>

          <FileTreeEnhanced
            v-else
            :nodes="filteredFileTree"
            :search-query="searchQuery"
            :show-hidden="showHidden"
            :selected-files="selectedFiles"
            :expanded-folders="expandedFolders"
            class="file-tree"
            @file-select="handleFileSelect"
            @folder-toggle="handleFolderToggle"
            @file-context-menu="handleFileContextMenu"
          />
        </div>

        <!-- Selection Summary -->
        <div v-if="selectedFiles.length > 0" class="selection-summary">
          <div class="summary-header">
            <span class="summary-title">Selected</span>
            <button
              class="clear-selection-btn"
              :title="'Clear selection'"
              @click="clearSelection"
            >
              <XIcon class="clear-icon" />
            </button>
          </div>
          <div class="summary-stats">
            <span class="stat-item">
              <FileIcon class="stat-icon" />
              {{ selectedFiles.length }} files
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
              {{ totalFileCount }} files
            </span>
            <span v-if="searchQuery" class="stat">
              {{ filteredFileCount }} filtered
            </span>
          </div>
          
          <div class="footer-actions">
            <button
              v-if="selectedFiles.length > 0"
              class="action-btn primary"
              @click="buildContext"
            >
              <ZapIcon class="btn-icon" />
              Build Context
            </button>
          </div>
        </div>
      </template>
    </BasePanel>

    <!-- File Selection Warning Modal -->
    <FileSelectionWarningModal
      v-if="showFileWarningModal"
      :file-count="fileWarningCount"
      @close="showFileWarningModal = false"
      @confirm="confirmFileSelection"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
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
import FileSelectionWarningModal from '@/presentation/components/modals/FileSelectionWarningModal.vue'
import { useFilePanelModernService } from '@/composables/useFilePanelModernService'

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

// Warning modal state
const showFileWarningModal = ref(false)
const fileWarningCount = ref(0)
const allowExcessiveSelection = ref(false)

// Use the service
const { 
  filePanelModernService,
  fileTreeStore,
  projectStore,
  contextBuilderStore,
  // Store refs
  isLoading,
  error,
  selectedFiles,
  fileTree,
  totalFileCount,
  currentProject,
  // Local state
  searchQuery,
  showHidden,
  expandedFolders,
  // Computed properties
  fileError,
  hasFiles,
  filteredFileTree,
  filteredFileCount,
  totalSelectedSize,
  // Methods
  handleToggle,
  handleResize,
  handleRetry,
  refreshFileTree,
  toggleShowHidden,
  handleSearch,
  clearSearch,
  handleFileSelect,
  handleFolderToggle,
  handleFileContextMenu,
  clearSelection,
  buildContext,
  formatFileSize
} = useFilePanelModernService()

// Watch for file selection warnings
const handleFileSelectionWarning = (event: Event) => {
  const customEvent = event as CustomEvent<{ fileCount: number }>
  fileWarningCount.value = customEvent.detail.fileCount
  showFileWarningModal.value = true
}

const confirmFileSelection = () => {
  allowExcessiveSelection.value = true
  showFileWarningModal.value = false
}

// Lifecycle
onMounted(() => {
  window.addEventListener('show-file-selection-warning', handleFileSelectionWarning as EventListener)
})

onUnmounted(() => {
  window.removeEventListener('show-file-selection-warning', handleFileSelectionWarning as EventListener)
})
</script>

<style scoped>
.file-panel-wrapper {
  flex: 1;
  height: 100%;
}

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