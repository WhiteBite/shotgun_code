<template>
  <BasePanel
    title="Context"
    :icon="ContextIcon"
    :collapsible="true"
    :is-collapsed="isCollapsed"
    :scrollable="true"
    :loading="contextLoading || isBuilding"
    :error="contextError"
    class="context-panel"
    @toggle="handleToggle"
    @resize="handleResize"
    @retry="handleRetry"
  >
    <div class="context-panel-content">
      <!-- Context Builder Section -->
      <div class="context-builder-section">
        <div class="section-header">
          <h4 class="section-title">Context Builder</h4>
          <div class="header-actions">
            <button
              :disabled="selectedFilesCount === 0"
              class="action-btn secondary small"
              title="Clear selection"
              @click="clearSelection"
            >
              <XIcon class="btn-icon" />
            </button>
          </div>
        </div>

        <!-- File Selection - REPLACED TEXTAREA WITH FILE LIST -->
        <div class="file-selection">
          <div v-if="selectedFiles.length === 0" class="no-files-message">
            No files selected. Select files in the File Explorer to build context.
          </div>
          
          <div v-else class="selected-files-list">
            <div 
              v-for="filePath in selectedFiles" 
              :key="filePath"
              class="selected-file-item"
            >
              <div class="file-info">
                <span class="file-icon">{{ getFileIcon(filePath) }}</span>
                <span class="file-name" :title="filePath">{{ getFileName(filePath) }}</span>
                <span class="file-path" :title="filePath">{{ getFilePath(filePath) }}</span>
              </div>
              <div class="file-actions">
                <button 
                  class="action-btn secondary small"
                  title="Preview file"
                  @click="previewFile(filePath)"
                >
                  <EyeIcon class="btn-icon" />
                </button>
                <button 
                  class="action-btn secondary small"
                  title="Remove file"
                  @click="removeFile(filePath)"
                >
                  <XIcon class="btn-icon" />
                </button>
              </div>
            </div>
          </div>

          <div v-if="selectedFilesCount > 0" class="selection-info">
            <span class="file-count">
              {{ selectedFilesCount }} files selected
            </span>
            <span v-if="validationErrors.length > 0" class="validation-errors">
              {{ validationErrors.length }} errors
            </span>
          </div>
        </div>

        <!-- Context Actions -->
        <ContextActions 
          :has-content="hasContextContent"
          :selected-files="selectedFiles"
          :is-building="isBuilding"
          :build-status="buildStatus"
          :context-summary="contextSummaryState"
          @build="handleBuild"
          @clear="handleClear"
          @export="handleExport"
          @import="handleImport"
        />
      </div>

      <!-- Context Preview Section -->
      <div v-if="hasContextContent" class="context-preview-section">
        <div class="section-header">
          <h4 class="section-title">Context Preview</h4>
          <div class="context-stats">
            <span class="stat-item">
              Tokens: {{ estimatedTokens }}
            </span>
          </div>
        </div>

        <!-- Context Content Display -->
        <div class="context-content content-scrollable">
          <SplitViewManager 
            v-if="contextPanelService.enableSplit"
            :context="contextData"
            :chunks="contextChunks"
            :active-chunk="activeChunk"
            :active-chunk-index="contextPanelService.activeChunkIndex"
            :split-settings="splitSettings"
            @chunk-select="contextPanelService.handleChunkSelect"
            @chunk-hover="contextPanelService.handleChunkHover"
            @copy-chunk="(chunk: ContextChunk) => contextPanelService.handleCopyChunk(chunk)"
            @copy-all="contextPanelService.handleCopyAll"
          />
          <div v-else class="context-simple-view">
            <!-- CRITICAL OOM FIX: Use paginated content instead of full content -->
            <ContextViewer
              :content="contextPanelService.paginatedContextContent"
              :highlight="true"
              :line-numbers="true"
              :virtual-scroll="true"
              :show-chunk-boundaries="false"
              class="context-viewer"
              @copy-selection="$emit('copy-selection', $event)"
            />
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="context-footer">
        <div class="footer-stats">
          <span class="stat">
            {{ selectedFilesCount }} files
          </span>
          <span v-if="estimatedTokens" class="stat">
            ~{{ estimatedTokens }} tokens
          </span>
        </div>
        
        <div class="footer-actions">
          <button
            v-if="hasContextContent"
            class="action-btn secondary"
            :class="{ active: contextPanelService.enableSplit }"
            @click="contextPanelService.toggleSplit"
          >
            <SplitIcon class="btn-icon" />
            Split
          </button>
          
          <button
            :disabled="selectedFilesCount === 0 || isBuilding"
            class="action-btn primary"
            @click="buildContext"
          >
            <ZapIcon v-if="!isBuilding" class="btn-icon" />
            <LoaderIcon v-else class="btn-icon animate-spin" />
            {{ isBuilding ? 'Building...' : 'Build Context' }}
          </button>
        </div>
      </div>
    </template>
  </BasePanel>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { 
  FileTextIcon, 
  XIcon, 
  ZapIcon, 
  LoaderIcon, 
  SplitIcon,
  EyeIcon
} from 'lucide-vue-next'
import BasePanel from '@/presentation/components/BasePanel.vue'
import ContextActions from '@/presentation/components/workspace/ContextActions/ContextActionsBar.vue'
import SplitViewManager from '@/presentation/components/shared/SplitViewManager.vue'
import ContextViewer from '@/presentation/components/shared/ContextViewer.vue'
import { getFileIcon } from '@/utils/fileIcons'
import { useContextPanelService } from '@/composables/useContextPanelService'
import type { ContextChunk } from '@/infrastructure/context/context-chunking.service'

// Icons
const ContextIcon = FileTextIcon

// Use the service
const { 
  contextPanelService,
  contextBuilderStore,
  projectStore,
  fileTreeStore,
  settingsStore,
  // Store refs
  contextLoading,
  isBuilding,
  buildStatus,
  contextSummaryState,
  estimatedTokens,
  error,
  validationErrors,
  selectedFiles,
  // Computed properties
  isCollapsed,
  selectedFilesCount,
  contextData,
  hasContextContent,
  contextError,
  contextChunks,
  activeChunk,
  splitSettings,
  // Methods
  handleToggle,
  handleResize,
  handleRetry,
  handleCopy,
  handleBuild,
  handleClear,
  handleExport,
  handleImport,
  clearSelection,
  buildContext,
  toggleCollapse
} = useContextPanelService()

// Refs
const fileTextarea = ref<HTMLTextAreaElement>()

// Helper functions for file display
const getFileName = (filePath: string): string => {
  return filePath.split('/').pop() || filePath
}

const getFilePath = (filePath: string): string => {
  const parts = filePath.split('/')
  parts.pop() // Remove filename
  return parts.join('/') || '/'
}

// Action methods
const previewFile = (filePath: string) => {
  // TODO: Implement file preview functionality
  console.log('Preview file:', filePath)
}

const removeFile = (filePath: string) => {
  fileTreeStore.removeSelectedFile(filePath)
}

// CRITICAL OOM FIX: Load paginated content when context changes
watch(() => contextSummaryState.value?.id, async (newContextId) => {
  if (newContextId) {
    try {
      // Load first chunk of context content
      const chunk = await contextBuilderStore.getContextContent(0, 1000);
      if (chunk) {
        contextPanelService.setPaginatedContextContent(chunk.lines.join('\n'));
      }
    } catch (error) {
      console.error('Error loading context content:', error);
    }
  } else {
    contextPanelService.setPaginatedContextContent('');
  }
}, { immediate: true });
</script>

<style scoped>
.context-panel-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 20px;
}

.context-builder-section,
.context-preview-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

.section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(203, 213, 225);
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 4px;
}

.file-selection {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.no-files-message {
  padding: 12px;
  text-align: center;
  color: rgb(100, 116, 139);
  font-size: 0.875rem;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 6px;
}

.selected-files-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 200px;
  overflow-y: auto;
  padding: 4px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 6px;
}

.selected-file-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: rgba(15, 23, 42, 0.6);
  border-radius: 4px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.file-icon {
  flex-shrink: 0;
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.file-name {
  font-size: 0.875rem;
  color: rgb(203, 213, 225);
  font-weight: 500;
  flex-shrink: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-path {
  font-size: 0.75rem;
  color: rgb(148, 163, 184);
  flex-shrink: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.selection-info {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.75rem;
}

.file-count {
  color: rgb(148, 163, 184);
}

.validation-errors {
  color: rgb(248, 113, 113);
}

.context-stats {
  display: flex;
  align-items: center;
  gap: 12px;
}

.stat-item {
  font-size: 0.75rem;
  color: rgb(148, 163, 184);
  display: flex;
  align-items: center;
  gap: 4px;
}

.context-content {
  flex: 1;
  min-height: 200px;
  background: rgba(15, 23, 42, 0.4);
  border-radius: 6px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  overflow: hidden;
}

.context-simple-view {
  height: 100%;
}

.context-viewer {
  height: 100%;
}

.context-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
}

.footer-stats {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 0.75rem;
  color: rgb(148, 163, 184);
}

.footer-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: 1px solid transparent;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-btn.small {
  padding: 6px 8px;
}

.action-btn.primary {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
}

.action-btn.primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #2563eb, #7c3aed);
  transform: translateY(-1px);
}

.action-btn.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.action-btn.secondary {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(148, 163, 184, 0.2);
  color: rgb(203, 213, 225);
}

.action-btn.secondary:hover:not(:disabled) {
  background: rgba(51, 65, 85, 0.8);
  border-color: rgba(148, 163, 184, 0.4);
}

.action-btn.secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn.active {
  background: rgba(59, 130, 246, 0.2);
  border-color: rgba(59, 130, 246, 0.5);
  color: rgb(147, 197, 253);
}

.btn-icon {
  width: 14px;
  height: 14px;
}

/* Responsive design */
@media (max-width: 768px) {
  .context-panel-content {
    gap: 16px;
  }
  
  .context-footer {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .footer-stats {
    justify-content: center;
  }
  
  .footer-actions {
    justify-content: center;
  }
}
</style>