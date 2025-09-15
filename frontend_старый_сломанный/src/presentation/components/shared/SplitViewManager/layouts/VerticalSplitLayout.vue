<template>
  <div class="vertical-split-layout">
    <SplitPane 
      :initial-ratio="0.6"
      direction="horizontal"
      :storage-key="`context-vertical-split-${contextId}`"
      class="split-container"
      @resize="handleResize"
    >
      <template #left>
        <div class="original-content-panel">
          <div class="panel-header">
            <h4 class="panel-title">Original Context</h4>
            <div class="panel-actions">
              <button
                class="action-btn"
                title="Copy original content"
                @click="copyOriginal"
              >
                <CopyIcon class="btn-icon" />
              </button>
            </div>
          </div>
          
          <ContextViewer
            :content="context.content"
            :chunks="chunks"
            :active-chunk="activeChunk"
            :highlight="true"
            :line-numbers="true"
            :show-chunk-boundaries="true"
            class="context-viewer"
            @chunk-hover="$emit('chunk-hover', $event)"
            @chunk-select="$emit('chunk-select', $event)"
            @copy-selection="$emit('copy-selection', $event)"
          />
        </div>
      </template>
      
      <template #right>
        <div class="chunk-preview-panel">
          <div class="panel-header">
            <h4 class="panel-title">
              Chunk {{ activeChunk ? activeChunk.metadata.chunkIndex + 1 : 1 }}
              <span v-if="chunks.length > 1" class="chunk-count">
                of {{ chunks.length }}
              </span>
            </h4>
            <div class="panel-actions">
              <button
                v-if="chunks.length > 1"
                :disabled="!canGoPrevious"
                class="action-btn"
                title="Previous chunk"
                @click="previousChunk"
              >
                <ChevronLeftIcon class="btn-icon" />
              </button>
              
              <button
                v-if="chunks.length > 1"
                :disabled="!canGoNext"
                class="action-btn"
                title="Next chunk"
                @click="nextChunk"
              >
                <ChevronRightIcon class="btn-icon" />
              </button>
              
              <button
                :disabled="!activeChunk"
                class="action-btn primary"
                title="Copy current chunk"
                @click="copyCurrentChunk"
              >
                <CopyIcon class="btn-icon" />
              </button>
            </div>
          </div>
          
          <div v-if="activeChunk" class="chunk-content">
            <div class="chunk-metadata">
              <div class="metadata-row">
                <span class="metadata-label">Lines:</span>
                <span class="metadata-value">
                  {{ activeChunk.startLine }}-{{ activeChunk.endLine }}
                  ({{ activeChunk.metadata.linesCount }} lines)
                </span>
              </div>
              
              <div class="metadata-row">
                <span class="metadata-label">Tokens:</span>
                <span class="metadata-value">{{ activeChunk.tokens }}</span>
              </div>
              
              <div class="metadata-row">
                <span class="metadata-label">Significance:</span>
                <span 
                  class="metadata-value significance-badge"
                  :class="`significance-${activeChunk.metadata.significance}`"
                >
                  {{ activeChunk.metadata.significance }}
                </span>
              </div>
              
              <div v-if="activeChunk.metadata.language" class="metadata-row">
                <span class="metadata-label">Language:</span>
                <span class="metadata-value">{{ activeChunk.metadata.language }}</span>
              </div>
            </div>
            
            <div class="chunk-preview">
              <ContextViewer
                :content="activeChunk.content"
                :highlight="true"
                :line-numbers="true"
                :virtual-scroll="false"
                :show-chunk-boundaries="false"
                class="chunk-viewer"
                @copy-selection="$emit('copy-selection', $event)"
              />
            </div>
          </div>
          
          <div v-else class="no-chunk-selected">
            <div class="empty-state">
              <div class="empty-icon">
                <FileTextIcon class="icon" />
              </div>
              <p class="empty-message">Select a chunk to preview</p>
            </div>
          </div>
        </div>
      </template>
    </SplitPane>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  CopyIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  FileTextIcon
} from 'lucide-vue-next'

import SplitPane from '@/presentation/components/shared/SplitPane.vue'
import ContextViewer from '@/presentation/components/shared/ContextViewer.vue'
import { useContextChunking } from '@/infrastructure/context/context-chunking.service'

interface ContextData {
  content: string
  fileCount?: number
  files?: string[]
  metadata?: any
}

interface ContextChunk {
  id: string
  content: string
  tokens: number
  startLine: number
  endLine: number
  startChar: number
  endChar: number
  metadata: any
}

interface Props {
  context: ContextData
  chunks: ContextChunk[]
  activeChunk: ContextChunk | null
  activeChunkIndex: number
  splitSettings: any
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'chunk-select': [index: number]
  'chunk-hover': [index: number | null]
  'copy-chunk': [chunk: ContextChunk]
  'copy-selection': [selection: string]
}>()

// Composables
const { copyChunk, copyAll } = useContextChunking()

// Computed
const contextId = computed(() => {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
})

const canGoPrevious = computed(() => {
  return props.activeChunkIndex > 0
})

const canGoNext = computed(() => {
  return props.activeChunkIndex < props.chunks.length - 1
})

// Methods
function handleResize(ratio: number) {
  // Handle split pane resize if needed
  console.log('Split pane resized:', ratio)
}

function previousChunk() {
  if (canGoPrevious.value) {
    emit('chunk-select', props.activeChunkIndex - 1)
  }
}

function nextChunk() {
  if (canGoNext.value) {
    emit('chunk-select', props.activeChunkIndex + 1)
  }
}

async function copyCurrentChunk() {
  if (props.activeChunk) {
    const success = await copyChunk(props.activeChunk, 'manual')
    if (success) {
      emit('copy-chunk', props.activeChunk)
    }
  }
}

async function copyOriginal() {
  const success = await copyAll(props.context.content, 'manual')
  if (success) {
    // Show success feedback
    console.log('Original content copied')
  }
}
</script>

<style scoped>
.vertical-split-layout {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.split-container {
  flex: 1;
  min-height: 0;
}

.original-content-panel,
.chunk-preview-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: rgba(15, 23, 42, 0.6);
  border-radius: 8px;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(145deg, rgba(51, 65, 85, 0.9), rgba(30, 41, 59, 0.8));
  border-bottom: 1px solid rgba(148, 163, 184, 0.15);
  flex-shrink: 0;
}

.panel-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(203, 213, 225);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.chunk-count {
  font-size: 0.75rem;
  color: rgb(100, 116, 139);
  font-weight: normal;
}

.panel-actions {
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

.action-btn.primary {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  border-color: transparent;
  color: white;
}

.action-btn.primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #2563eb, #7c3aed);
  transform: translateY(-1px);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.btn-icon {
  width: 14px;
  height: 14px;
}

.context-viewer {
  flex: 1;
  min-height: 0;
}

.chunk-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.chunk-metadata {
  padding: 16px;
  background: rgba(30, 41, 59, 0.4);
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  flex-shrink: 0;
}

.metadata-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.metadata-row:last-child {
  margin-bottom: 0;
}

.metadata-label {
  font-size: 0.75rem;
  color: rgb(100, 116, 139);
  font-weight: 500;
}

.metadata-value {
  font-size: 0.75rem;
  color: rgb(203, 213, 225);
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
}

.significance-badge {
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
  text-transform: uppercase;
  font-size: 0.625rem;
}

.significance-high {
  background: rgba(239, 68, 68, 0.2);
  color: rgb(248, 113, 113);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.significance-medium {
  background: rgba(251, 191, 36, 0.2);
  color: rgb(252, 211, 77);
  border: 1px solid rgba(251, 191, 36, 0.3);
}

.significance-low {
  background: rgba(34, 197, 94, 0.2);
  color: rgb(74, 222, 128);
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.chunk-preview {
  flex: 1;
  min-height: 0;
}

.chunk-viewer {
  height: 100%;
}

.no-chunk-selected {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  text-align: center;
  padding: 40px 20px;
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

.empty-message {
  font-size: 0.875rem;
  color: rgb(100, 116, 139);
  margin: 0;
}

/* Responsive design */
@media (max-width: 768px) {
  .panel-header {
    padding: 8px 12px;
  }
  
  .chunk-metadata {
    padding: 12px;
  }
  
  .panel-title {
    font-size: 0.75rem;
  }
  
  .chunk-count {
    display: none;
  }
  
  .metadata-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    margin-bottom: 12px;
  }
}
</style>