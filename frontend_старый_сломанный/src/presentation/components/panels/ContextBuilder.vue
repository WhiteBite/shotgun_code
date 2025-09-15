<template>
  <BasePanel
    title="Context Builder"
    :icon="FileTextIcon"
    :collapsible="true"
    :is-collapsed="isCollapsed"
    :scrollable="false"
    :loading="contextBuilderStore.isBuilding"
    :error="contextBuilderStore.error"
    :resizable="true"
    :min-width="320"
    :max-width="600"
    :width="400"
    variant="secondary"
    size="md"
    class="context-builder-panel"
    @toggle="handleToggle"
    @resize="handleResize"
    @retry="handleRetry"
  >
    <div class="context-builder-content">
      <!-- Header with context stats -->
      <div class="context-header">
        <div class="context-stats">
          <span class="stat-item">
            <FileIcon class="stat-icon" />
            {{ contextBuilderStore.selectedFilesList?.length || 0 }} files
          </span>
          <span class="stat-item">
            <HashIcon class="stat-icon" />
            {{ contextBuilderStore.contextMetrics.tokenCount }} tokens
          </span>
        </div>
      </div>

      <!-- Selected Files List -->
      <div class="selected-files-container">
        <div
          v-if="!contextBuilderStore.selectedFilesList || contextBuilderStore.selectedFilesList.length === 0"
          class="empty-state"
        >
          <div class="empty-icon">
            <FileTextIcon class="icon" />
          </div>
          <p class="empty-title">No files selected</p>
          <p class="empty-description">
            Select files in the File Explorer to build context
          </p>
        </div>

        <div
          v-else
          class="selected-files-list content-scrollable"
        >
          <div
            v-for="filePath in contextBuilderStore.selectedFilesList"
            :key="filePath"
            class="selected-file-item"
          >
            <div class="file-info">
              <span
                class="file-icon"
                :class="getFileIconColor(getFileExtension(filePath))"
              >{{ getFileIcon(getFileExtension(filePath)) }}</span>
              <div class="file-details">
                <span class="file-name">{{ getFileName(filePath) }}</span>
                <span class="file-path">{{ getFilePath(filePath) }}</span>
              </div>
            </div>

            <div class="file-actions">
              <button
                class="action-btn secondary small"
                :title="'Preview file'"
                @click="previewFile(filePath)"
              >
                <EyeIcon class="btn-icon" />
              </button>
              <button
                class="action-btn secondary small"
                :title="'Remove file'"
                @click="removeFile(filePath)"
              >
                <XIcon class="btn-icon" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Context Preview Area -->
      <div
        v-if="showContextPreview && contextBuilderStore.contextSummaryState"
        class="context-preview-container"
      >
        <div class="preview-header">
          <h4 class="preview-title">Context Preview</h4>
          <button
            class="action-btn secondary small"
            @click="showContextPreview = false"
          >
            <XIcon class="btn-icon" />
          </button>
        </div>
        <div class="preview-content content-scrollable">
          <!-- We'll use a simple textarea for now, but this could be enhanced -->
          <textarea
            v-if="contextPreviewContent"
            :value="contextPreviewContent"
            readonly
            class="preview-textarea"
          ></textarea>
          <div v-else class="preview-placeholder">
            Context preview will appear here after building
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="context-footer">
        <button
          class="action-btn secondary"
          :disabled="!contextBuilderStore.selectedFilesList || contextBuilderStore.selectedFilesList.length === 0"
          @click="clearSelection"
        >
          <XIcon class="btn-icon" />
          Clear
        </button>

        <button
          class="action-btn primary"
          :disabled="!contextBuilderStore.selectedFilesList || contextBuilderStore.selectedFilesList.length === 0 || contextBuilderStore.isBuilding"
          @click="buildContext"
        >
          <ZapIcon v-if="!contextBuilderStore.isBuilding" class="btn-icon" />
          <LoaderIcon v-else class="btn-icon animate-spin" />
          {{ contextBuilderStore.isBuilding ? 'Building...' : 'Build Context' }}
        </button>

        <button
          class="action-btn secondary"
          :disabled="!contextBuilderStore.selectedFilesList || contextBuilderStore.selectedFilesList.length === 0 || !contextBuilderStore.contextSummaryState"
          @click="showContextPreview = !showContextPreview"
        >
          <EyeIcon class="btn-icon" />
          {{ showContextPreview ? 'Hide Preview' : 'Preview' }}
        </button>
      </div>
    </template>
  </BasePanel>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  FileTextIcon,
  FileIcon,
  HashIcon,
  EyeIcon,
  XIcon,
  ZapIcon,
  LoaderIcon
} from 'lucide-vue-next'

import BasePanel from '@/presentation/components/BasePanel.vue'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useProjectStore } from '@/stores/project.store'
import { getFileIcon } from '@/utils/fileIcons'

// Stores
const contextBuilderStore = useContextBuilderStore()
const projectStore = useProjectStore()

// Component state
const showContextPreview = ref(false)
const contextPreviewContent = ref('')

// Props
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

// Methods
function handleToggle(collapsed: boolean) {
  emit('toggle', collapsed)
}

function handleResize(width: number) {
  emit('resize', width)
}

function handleRetry() {
  // Retry logic if needed
}

async function loadPreviewFromContext() {
  try {
    // Не пытаемся грузить, если контекст ещё не построен
    // @ts-ignore: readonly unwrap
    if (!contextBuilderStore.currentContextId) return
    // Загружаем первый чанκ текущего контекста постранично
    const pageSize = contextBuilderStore.contextPageSize || 200;
    const chunk = await contextBuilderStore.getContextContent?.(0, pageSize);
    if (chunk) {
      // Универсальная распаковка: поддержим content или lines
      const content = (chunk as any).content
        ?? (Array.isArray((chunk as any).lines) ? (chunk as any).lines.join('\n') : '')
      contextPreviewContent.value = content || contextPreviewContent.value
    }
  } catch (e) {
    console.error('Failed to load preview chunk:', e)
  }
}

function getFileExtension(filePath: string): string {
  return filePath.split('.').pop()?.toLowerCase() || ''
}

function getFileName(filePath: string): string {
  return filePath.split('/').pop() || filePath
}

function getFilePath(filePath: string): string {
  const parts = filePath.split('/')
  parts.pop() // Remove filename
  return parts.join('/') || '/'
}

// Отдаём строковый эмодзи-икон, используем как текст, а не как компонент

function getFileIconColor(extension: string) {
  switch (extension) {
    case 'ts':
    case 'tsx':
      return 'text-blue-400'
    case 'js':
    case 'jsx':
      return 'text-yellow-400'
    case 'vue':
      return 'text-green-400'
    case 'json':
      return 'text-orange-400'
    default:
      return 'text-gray-400'
  }
}

function previewFile(filePath: string) {
  // For now, just show in console
  console.log('Preview file:', filePath)
  showContextPreview.value = true
  // Если нет контента — попробуем подгрузить первый чанκ из контекста
  // Только если контекст уже построен
  // @ts-ignore: readonly unwrap
  if (!contextPreviewContent.value && contextBuilderStore.currentContextId) {
    loadPreviewFromContext()
  } else if (!contextBuilderStore.contextSummaryState) {
    contextPreviewContent.value = 'No context yet. Click "Build Context" to generate preview.'
  }
}

function removeFile(filePath: string) {
  contextBuilderStore.removeSelectedFile(filePath)
}

function clearSelection() {
  contextBuilderStore.clearSelectedFiles()
}

async function buildContext() {
  if (projectStore.currentProject?.path) {
    try {
      await contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path)
      // After building, show the preview
      showContextPreview.value = true
      // Подгрузим первый чанκ текста для предпросмотра
      await loadPreviewFromContext()
    } catch (error) {
      console.error('Failed to build context:', error)
    }
  }
}

// Watch for context changes to update preview
watch(
  () => contextBuilderStore.contextSummaryState,
  (newContext) => {
    if (newContext) {
      contextPreviewContent.value = `Context ID: ${newContext.id}
Files: ${newContext.fileCount}
Tokens: ${newContext.tokenCount}
Size: ${newContext.totalSize} characters`
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.context-builder-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 16px;
}

.context-header {
  padding: 0 12px;
}

.context-stats {
  display: flex;
  gap: 16px;
  font-size: 0.75rem;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: rgb(148, 163, 184);
}

.stat-icon {
  width: 12px;
  height: 12px;
}

.selected-files-container {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 40px 20px;
  text-align: center;
  flex: 1;
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

.selected-files-list {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 12px;
  max-height: 300px;
}

.selected-file-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  background: rgba(15, 23, 42, 0.6);
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.file-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.file-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.file-details {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.file-name {
  font-size: 0.875rem;
  color: rgb(203, 213, 225);
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-path {
  font-size: 0.75rem;
  color: rgb(148, 163, 184);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.context-preview-container {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  padding-top: 16px;
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px 12px;
}

.preview-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(203, 213, 225);
  margin: 0;
}

.preview-content {
  flex: 1;
  min-height: 0;
}

.preview-textarea {
  width: 100%;
  height: 100%;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 6px;
  color: rgb(203, 213, 225);
  font-size: 0.875rem;
  font-family: monospace;
  padding: 12px;
  resize: none;
}

.preview-placeholder {
  padding: 20px;
  text-align: center;
  color: rgb(100, 116, 139);
  font-size: 0.875rem;
}

.context-footer {
  display: flex;
  gap: 8px;
  padding: 12px 0;
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
  flex: 1;
  justify-content: center;
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

.btn-icon {
  width: 14px;
  height: 14px;
}
</style>
