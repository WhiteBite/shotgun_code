<template>
  <div class="horizontal-split-layout">
    <SplitPane 
      :initial-ratio="0.4"
      direction="vertical"
      :storage-key="`context-horizontal-split-${contextId}`"
      @resize="handleResize"
      class="split-container"
    >
      <template #left>
        <div class="chunk-navigator-panel">
          <div class="panel-header">
            <h4 class="panel-title">Chunk Navigator</h4>
            <div class="panel-actions">
              <button
                @click="toggleAutoSelect"
                :class="['action-btn', { active: autoSelect }]"
                title="Auto-select chunks on hover"
              >
                <MousePointerIcon class="btn-icon" />
              </button>
              
              <button
                @click="copyAllChunks"
                class="action-btn primary"
                title="Copy all chunks"
              >
                <CopyIcon class="btn-icon" />
              </button>
            </div>
          </div>
          
          <div class="chunk-navigator-content">
            <div class="navigator-grid">
              <div
                v-for="(chunk, index) in chunks"
                :key="chunk.id"
                @click="selectChunk(index)"
                @mouseenter="handleChunkHover(index)"
                @mouseleave="handleChunkHover(null)"
                :class="[
                  'chunk-card',
                  {
                    active: activeChunkIndex === index,
                    hovered: hoveredChunk === index
                  }
                ]"
              >
                <div class="chunk-header">
                  <span class="chunk-number">{{ index + 1 }}</span>
                  <div class="chunk-actions">
                    <button
                      @click.stop="copyChunk(chunk)"
                      class="chunk-action-btn"
                      title="Copy this chunk"
                    >
                      <CopyIcon class="action-icon" />
                    </button>
                  </div>
                </div>
                
                <div class="chunk-info">
                  <div class="info-row">
                    <span class="info-label">Lines:</span>
                    <span class="info-value">
                      {{ chunk.startLine }}-{{ chunk.endLine }}
                    </span>
                  </div>
                  
                  <div class="info-row">
                    <span class="info-label">Tokens:</span>
                    <span class="info-value">{{ chunk.tokens }}</span>
                  </div>
                  
                  <div class="info-row">
                    <span class="info-label">Significance:</span>
                    <span 
                      :class="[
                        'info-value',
                        'significance-badge',
                        `significance-${chunk.metadata.significance}`
                      ]"
                    >
                      {{ chunk.metadata.significance }}
                    </span>
                  </div>
                </div>
                
                <div class="chunk-preview">
                  <code class="preview-text">
                    {{ getChunkPreview(chunk.content) }}
                  </code>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
      
      <template #right>
        <div class="chunk-detail-panel">
          <div class="panel-header">
            <h4 class="panel-title">
              {{ activeChunk ? `Chunk ${activeChunk.metadata.chunkIndex + 1} Details` : 'Select a Chunk' }}
            </h4>
            <div v-if="activeChunk" class="panel-actions">
              <button
                @click="previousChunk"
                :disabled="!canGoPrevious"
                class="action-btn"
                title="Previous chunk"
              >
                <ChevronUpIcon class="btn-icon" />
              </button>
              
              <button
                @click="nextChunk"
                :disabled="!canGoNext"
                class="action-btn"
                title="Next chunk"
              >
                <ChevronDownIcon class="btn-icon" />
              </button>
              
              <button
                @click="copyCurrentChunk"
                class="action-btn primary"
                title="Copy current chunk"
              >
                <CopyIcon class="btn-icon" />
              </button>
            </div>
          </div>
          
          <div v-if="activeChunk" class="chunk-detail-content">
            <div class="detail-metadata">
              <div class="metadata-grid">
                <div class="metadata-item">
                  <span class="metadata-label">Lines</span>
                  <span class="metadata-value">
                    {{ activeChunk.startLine }}-{{ activeChunk.endLine }}
                    ({{ activeChunk.metadata.linesCount }} total)
                  </span>
                </div>
                
                <div class="metadata-item">
                  <span class="metadata-label">Tokens</span>
                  <span class="metadata-value">{{ activeChunk.tokens }}</span>
                </div>
                
                <div class="metadata-item">
                  <span class="metadata-label">Characters</span>
                  <span class="metadata-value">{{ activeChunk.content.length }}</span>
                </div>
                
                <div class="metadata-item">
                  <span class="metadata-label">Significance</span>
                  <span 
                    :class="[
                      'metadata-value',
                      'significance-badge',
                      `significance-${activeChunk.metadata.significance}`
                    ]"
                  >
                    {{ activeChunk.metadata.significance }}
                  </span>
                </div>
                
                <div v-if="activeChunk.metadata.language" class="metadata-item">
                  <span class="metadata-label">Language</span>
                  <span class="metadata-value">{{ activeChunk.metadata.language }}</span>
                </div>
                
                <div class="metadata-item">
                  <span class="metadata-label">Position</span>
                  <span class="metadata-value">
                    {{ Math.round((activeChunk.metadata.chunkIndex / chunks.length) * 100) }}% through content
                  </span>
                </div>
              </div>
            </div>
            
            <div class="detail-content">
              <ContextViewer
                :content="activeChunk.content"
                :highlight="true"
                :line-numbers="true"
                :virtual-scroll="false"
                :show-chunk-boundaries="false"
                @copy-selection="$emit('copy-selection', $event)"
                class="detail-viewer"
              />
            </div>
          </div>
          
          <div v-else class="no-chunk-detail">
            <div class="empty-state">
              <div class="empty-icon">
                <MousePointerClickIcon class="icon" />
              </div>
              <p class="empty-message">Click on a chunk above to view its details</p>
              <p class="empty-hint">You can also use arrow keys to navigate</p>
            </div>
          </div>
        </div>
      </template>
    </SplitPane>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  CopyIcon,
  ChevronUpIcon,
  ChevronDownIcon,
  MousePointerIcon,
  MousePointerClickIcon
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
const { copyChunk: copyChunkService, copyAll } = useContextChunking()

// State
const hoveredChunk = ref<number | null>(null)
const autoSelect = ref(false)

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
  console.log('Horizontal split resized:', ratio)
}

function selectChunk(index: number) {
  emit('chunk-select', index)
}

function handleChunkHover(index: number | null) {
  hoveredChunk.value = index
  emit('chunk-hover', index)
  
  if (autoSelect.value && index !== null) {
    selectChunk(index)
  }
}

function toggleAutoSelect() {
  autoSelect.value = !autoSelect.value
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

async function copyChunk(chunk: ContextChunk) {
  const success = await copyChunkService(chunk, 'manual')
  if (success) {
    emit('copy-chunk', chunk)
  }
}

async function copyCurrentChunk() {
  if (props.activeChunk) {
    await copyChunk(props.activeChunk)
  }
}

async function copyAllChunks() {
  const allContent = props.chunks.map(chunk => chunk.content).join('\n\n---\n\n')
  const success = await copyAll(allContent, 'manual')
  if (success) {
    console.log('All chunks copied')
  }
}

function getChunkPreview(content: string): string {
  const lines = content.split('\n')
  const previewLines = lines.slice(0, 3)
  let preview = previewLines.join(' ').trim()
  
  if (preview.length > 80) {
    preview = preview.substring(0, 77) + '...'
  } else if (lines.length > 3) {
    preview += '...'
  }
  
  return preview || 'Empty chunk'
}
</script>

<style scoped>
.horizontal-split-layout {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.split-container {
  flex: 1;
  min-height: 0;
}

.chunk-navigator-panel,
.chunk-detail-panel {
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

.action-btn.active {
  background: rgba(59, 130, 246, 0.2);
  border-color: rgba(59, 130, 246, 0.4);
  color: rgb(147, 197, 253);
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

.chunk-navigator-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.navigator-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.chunk-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.chunk-card:hover {
  background: rgba(51, 65, 85, 0.8);
  border-color: rgba(148, 163, 184, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.chunk-card.active {
  background: rgba(59, 130, 246, 0.2);
  border-color: rgba(59, 130, 246, 0.5);
}

.chunk-card.hovered {
  background: rgba(168, 85, 247, 0.2);
  border-color: rgba(168, 85, 247, 0.5);
}

.chunk-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.chunk-number {
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(168, 85, 247);
}

.chunk-actions {
  display: flex;
  gap: 4px;
}

.chunk-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 4px;
  color: rgb(148, 163, 184);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.chunk-action-btn:hover {
  background: rgba(51, 65, 85, 0.8);
  border-color: rgba(148, 163, 184, 0.4);
  color: rgb(203, 213, 225);
}

.action-icon {
  width: 12px;
  height: 12px;
}

.chunk-info {
  margin-bottom: 12px;
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  font-size: 0.75rem;
  color: rgb(100, 116, 139);
}

.info-value {
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
  background: rgba(15, 23, 42, 0.4);
  border-radius: 4px;
  padding: 8px;
  font-size: 0.75rem;
}

.preview-text {
  color: rgb(148, 163, 184);
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
  word-break: break-all;
}

.chunk-detail-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.detail-metadata {
  padding: 16px;
  background: rgba(30, 41, 59, 0.4);
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  flex-shrink: 0;
}

.metadata-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
}

.metadata-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
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

.detail-content {
  flex: 1;
  min-height: 0;
}

.detail-viewer {
  height: 100%;
}

.no-chunk-detail {
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
  color: rgb(203, 213, 225);
  margin: 0;
}

.empty-hint {
  font-size: 0.75rem;
  color: rgb(100, 116, 139);
  margin: 0;
}

/* Responsive design */
@media (max-width: 768px) {
  .navigator-grid {
    grid-template-columns: 1fr;
  }
  
  .metadata-grid {
    grid-template-columns: 1fr 1fr;
  }
  
  .panel-header {
    padding: 8px 12px;
  }
  
  .chunk-navigator-content {
    padding: 12px;
  }
  
  .chunk-card {
    padding: 10px;
  }
}
</style>