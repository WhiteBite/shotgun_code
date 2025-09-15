<template>
  <BasePanel
    title="Context"
    :icon="ContextIcon"
    :collapsible="true"
    :is-collapsed="isCollapsed"
    :scrollable="true"
    :loading="contextBuilderStore.contextLoading"
    :error="contextError"
    @toggle="handleToggle"
    @resize="handleResize"
    @retry="handleRetry"
    class="context-panel"
  >
    <div class="context-panel-content">
      <!-- Context Builder Section -->
      <div class="context-builder-section">
        <div class="section-header">
          <h4 class="section-title">{{ $t('context.builder', 'Context Builder') }}</h4>
          <div class="header-actions">
            <button
              @click="clearSelection"
              :disabled="selectedFilesCount === 0"
              class="action-btn secondary small"
              :title="$t('context.clearSelection', 'Clear selection')"
            >
              <XIcon class="btn-icon" />
            </button>
          </div>
        </div>

        <!-- File Selection -->
        <div class="file-selection">
          <textarea
            ref="fileTextarea"
            v-model="selectedFilesText"
            :placeholder="$t('context.selectedFilesPlaceholder', 'Enter file paths, one per line')"
            class="file-textarea"
            rows="4"
          />
          
          <div v-if="selectedFilesCount > 0" class="selection-info">
            <span class="file-count">
              {{ selectedFilesCount }} {{ $t('context.filesSelected', 'files selected') }}
            </span>
            <span v-if="contextBuilderStore.validationErrors.length > 0" class="validation-errors">
              {{ contextBuilderStore.validationErrors.length }} {{ $t('context.errors', 'errors') }}
            </span>
          </div>
        </div>

        <!-- Context Actions -->
        <ContextActions 
          :selected-files="selectedFiles"
          :is-building="contextBuilderStore.isBuilding"
          :build-status="contextBuilderStore.buildStatus"
          :context-summary="contextBuilderStore.contextSummaryState"
          @build="handleBuild"
          @clear="handleClear"
          @export="handleExport"
          @import="handleImport"
        />
      </div>

      <!-- Context Preview Section -->
      <div v-if="hasContextContent" class="context-preview-section">
        <div class="section-header">
          <h4 class="section-title">{{ $t('context.preview', 'Context Preview') }}</h4>
          <div class="context-stats">
            <span class="stat-item">
              {{ $t('context.tokens', 'Tokens') }}: {{ estimatedTokens }}
            </span>
          </div>
        </div>

        <!-- Context Content Display -->
        <div class="context-content content-scrollable">
          <SplitViewManager 
            v-if="enableSplit"
            :context="contextData"
            :chunks="contextChunks"
            :active-chunk="activeChunk"
            :active-chunk-index="activeChunkIndex"
            :split-settings="splitSettings"
            @chunk-select="handleChunkSelect"
            @chunk-hover="handleChunkHover"
            @copy-chunk="handleCopyChunk"
            @copy-all="handleCopyAll"
          />
          <div v-else class="context-simple-view">
            <!-- CRITICAL OOM FIX: Use paginated content instead of full content -->
            <ContextViewer
              :content="paginatedContextContent"
              :highlight="true"
              :line-numbers="true"
              :virtual-scroll="true"
              :show-chunk-boundaries="false"
              @copy-selection="$emit('copy-selection', $event)"
              class="context-viewer"
            />
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="context-footer">
        <div class="footer-stats">
          <span class="stat">
            {{ selectedFilesCount }} {{ $t('context.files', 'files') }}
          </span>
          <span v-if="estimatedTokens" class="stat">
            ~{{ estimatedTokens }} {{ $t('context.tokens', 'tokens') }}
          </span>
        </div>
        
        <div class="footer-actions">
          <button
            v-if="hasContextContent"
            @click="toggleSplit"
            class="action-btn secondary"
            :class="{ active: enableSplit }"
          >
            <SplitIcon class="btn-icon" />
            {{ $t('context.split', 'Split') }}
          </button>
          
          <button
            @click="buildContext"
            :disabled="selectedFilesCount === 0 || contextBuilderStore.isBuilding"
            class="action-btn primary"
          >
            <ZapIcon v-if="!contextBuilderStore.isBuilding" class="btn-icon" />
            <LoaderIcon v-else class="btn-icon animate-spin" />
            {{ contextBuilderStore.isBuilding ? $t('context.building', 'Building...') : $t('context.buildContext', 'Build Context') }}
          </button>
        </div>
      </div>
    </template>
  </BasePanel>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { 
  DocumentTextIcon, 
  XIcon, 
  ZapIcon, 
  LoaderIcon, 
  SplitIcon 
} from 'lucide-vue-next'
import BasePanel from '@/presentation/components/BasePanel.vue'
import ContextActions from '@/presentation/components/workspace/ContextActions/ContextActionsBar.vue'
import SplitViewManager from '@/presentation/components/shared/SplitViewManager.vue'
import ContextViewer from '@/presentation/components/shared/ContextViewer.vue'
import { getFileIcon } from '@/utils/fileIcons'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useProjectStore } from '@/stores/project.store'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useSettingsStore } from '@/stores/settings.store'
import usePanelManager from '@/composables/usePanelManager'
import { useContextChunking, type ContextChunk } from '@/infrastructure/context/context-chunking.service'

// Icons
const ContextIcon = DocumentTextIcon

// Composables
const { isCollapsed, toggleCollapse } = usePanelManager('context', 320)
const { chunkContent, copyChunk, copyAll, getCopyHistory } = useContextChunking()

// Stores
const fileTreeStore = useFileTreeStore()
const contextBuilderStore = useContextBuilderStore()
const projectStore = useProjectStore()
const settingsStore = useSettingsStore()

// Refs
const fileTextarea = ref<HTMLTextAreaElement>()

// State
const enableSplit = ref(false)
const chunkedContext = ref<ContextChunk[]>([])
const activeChunkIndex = ref(0)
// CRITICAL OOM FIX: Add state for paginated content
const paginatedContextContent = ref<string>('')

// Computed properties
const selectedFiles = computed(() => fileTreeStore.selectedFiles)
const selectedFilesCount = computed(() => selectedFiles.value.length)
// CRITICAL OOM FIX: Update contextData to use context summary instead of full content
const contextData = computed(() => ({
  content: paginatedContextContent.value,
  fileCount: contextBuilderStore.contextSummaryState?.fileCount || 0,
  metadata: contextBuilderStore.contextSummaryState?.metadata
}))
const hasContextContent = computed(() => !!contextBuilderStore.contextSummaryState?.id)
const estimatedTokens = computed(() => contextBuilderStore.estimatedTokens)

const contextError = computed(() => {
  if (contextBuilderStore.error) {
    return {
      message: contextBuilderStore.error,
      code: 'CONTEXT_BUILD_ERROR'
    }
  }
  return null
})

const contextChunks = computed(() => {
  if (!enableSplit.value || !paginatedContextContent.value) {
    return []
  }
  
  if (chunkedContext.value.length === 0) {
    // Auto-chunk the content when split is enabled
    chunkedContext.value = chunkContent(paginatedContextContent.value, {
      maxTokens: settingsStore.chunkSize || 1000,
      strategy: 'balanced',
      preserveCodeBlocks: true,
      preserveMarkdownStructure: true
    })
  }
  
  return chunkedContext.value
})

const activeChunk = computed(() => {
  return contextChunks.value[activeChunkIndex.value] || null
})

const splitSettings = computed(() => ({
  enabled: enableSplit.value,
  layout: settingsStore.splitLayout || 'vertical',
  chunkSize: settingsStore.chunkSize || 1000,
  chunks: contextChunks.value,
  activeChunkIndex: activeChunkIndex.value
}))

// Text field for selected files
const selectedFilesText = computed({
  get: () => selectedFiles.value.join('\n'),
  set: (value: string) => {
    const newFiles = value
      .split('\n')
      .map(v => v.trim().replace(/\\/g, '/'))
      .filter(Boolean)
    fileTreeStore.setSelectedFiles(newFiles)
  }
})

// CRITICAL OOM FIX: Load paginated content when context changes
watch(() => contextBuilderStore.contextSummaryState?.id, async (newContextId) => {
  if (newContextId) {
    try {
      // Load first chunk of context content
      const chunk = await contextBuilderStore.getContextContent(0, 1000);
      if (chunk) {
        paginatedContextContent.value = chunk.lines.join('\n');
      }
    } catch (error) {
      console.error('Error loading context content:', error);
    }
  } else {
    paginatedContextContent.value = '';
  }
}, { immediate: true });

// Methods
function handleToggle(collapsed: boolean) {
  toggleCollapse()
}

function handleResize(width: number) {
  // Handle panel resize if needed
  console.log('Panel resized to:', width)
}

function handleRetry() {
  if (projectStore.currentProject?.path) {
    contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
  }
}

function handleCopy() {
  if (paginatedContextContent.value) {
    navigator.clipboard.writeText(paginatedContextContent.value)
  }
}

function handleBuild() {
  if (projectStore.currentProject?.path) {
    contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
  }
}

function handleClear() {
  contextBuilderStore.resetContext()
  fileTreeStore.clearSelection()
}

function handleExport() {
  // Export context functionality
  console.log('Export context')
}

function handleImport() {
  // Import context functionality
  console.log('Import context')
}

function clearSelection() {
  fileTreeStore.clearSelection()
}

function buildContext() {
  if (projectStore.currentProject?.path) {
    contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
  }
}

function toggleSplit() {
  enableSplit.value = !enableSplit.value
}

function handleChunkSelect(index: number) {
  activeChunkIndex.value = index
}

function handleChunkHover(index: number | null) {
  // Handle chunk hover if needed
  console.log('Chunk hover:', index)
}

function handleCopyChunk(chunk: ContextChunk) {
  copyChunk(chunk, 'manual')
}

function handleCopyAll(content: string) {
  copyAll(content, 'manual')
}
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

.file-textarea {
  width: 100%;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 6px;
  padding: 12px;
  color: rgb(203, 213, 225);
  font-size: 0.875rem;
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
  resize: vertical;
  min-height: 80px;
}

.file-textarea:focus {
  outline: none;
  border-color: rgba(59, 130, 246, 0.5);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
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