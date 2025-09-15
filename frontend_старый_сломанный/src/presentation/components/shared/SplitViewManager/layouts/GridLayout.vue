<template>
  <div class="grid-layout">
    <div class="grid-container">
      <div class="grid-panel panel-top-left">
        <div class="panel-header">
          <h4 class="panel-title">Full Context</h4>
        </div>
        <ContextViewer
          :content="context.content"
          :chunks="chunks"
          :highlight="true"
          :line-numbers="true"
          :show-chunk-boundaries="true"
          class="panel-content"
          @chunk-hover="$emit('chunk-hover', $event)"
          @chunk-select="$emit('chunk-select', $event)"
          @copy-selection="$emit('copy-selection', $event)"
        />
      </div>
      
      <div class="grid-panel panel-top-right">
        <div class="panel-header">
          <h4 class="panel-title">
            {{ activeChunk ? `Chunk ${activeChunk.metadata.chunkIndex + 1}` : 'Select Chunk' }}
          </h4>
        </div>
        <div v-if="activeChunk" class="panel-content">
          <div class="chunk-metadata">
            <div class="metadata-grid">
              <div class="metadata-item">
                <span class="metadata-label">Lines</span>
                <span class="metadata-value">
                  {{ activeChunk.startLine }}-{{ activeChunk.endLine }}
                </span>
              </div>
              
              <div class="metadata-item">
                <span class="metadata-label">Tokens</span>
                <span class="metadata-value">{{ activeChunk.tokens }}</span>
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
            </div>
          </div>
          
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
        <div v-else class="panel-content empty-state">
          <div class="empty-message">Select a chunk to view details</div>
        </div>
      </div>
      
      <div class="grid-panel panel-bottom-left">
        <div class="panel-header">
          <h4 class="panel-title">Chunk Navigator</h4>
        </div>
        <div class="chunk-navigator">
          <div 
            v-for="(chunk, index) in chunks"
            :key="chunk.id"
            :class="[
              'chunk-item',
              {
                active: activeChunkIndex === index,
                hovered: hoveredChunk === index
              }
            ]"
            @click="$emit('chunk-select', index)"
          >
            <div class="chunk-info">
              <span class="chunk-number">{{ index + 1 }}</span>
              <span class="chunk-tokens">{{ chunk.tokens }} tokens</span>
            </div>
            <div class="chunk-preview">
              {{ getChunkPreview(chunk.content) }}
            </div>
          </div>
        </div>
      </div>
      
      <div class="grid-panel panel-bottom-right">
        <div class="panel-header">
          <h4 class="panel-title">Actions</h4>
        </div>
        <div class="actions-panel">
          <div class="action-buttons">
            <button
              :disabled="!activeChunk"
              class="action-btn primary"
              @click="copyActiveChunk"
            >
              <CopyIcon class="btn-icon" />
              Copy Chunk
            </button>
            
            <button
              class="action-btn secondary"
              @click="copyAllContent"
            >
              <FileTextIcon class="btn-icon" />
              Copy All
            </button>
          </div>
          
          <div class="content-stats">
            <div class="stat-item">
              <FileIcon class="stat-icon" />
              <span>{{ context?.fileCount || 0 }} files</span>
            </div>
            <div class="stat-item">
              <HashIcon class="stat-icon" />
              <span>{{ totalTokens }} tokens</span>
            </div>
            <div class="stat-item">
              <LayersIcon class="stat-icon" />
              <span>{{ chunks.length }} chunks</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  CopyIcon,
  FileTextIcon,
  FileIcon,
  HashIcon,
  LayersIcon
} from 'lucide-vue-next'

import ContextViewer from '@/presentation/components/shared/ContextViewer.vue'

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
  'copy-all': [content: string]
}>()

// State
const hoveredChunk = ref<number | null>(null)

// Computed
const totalTokens = computed(() => {
  return props.chunks.reduce((sum, chunk) => sum + chunk.tokens, 0)
})

// Methods
function copyActiveChunk() {
  if (props.activeChunk) {
    emit('copy-chunk', props.activeChunk)
  }
}

function copyAllContent() {
  if (props.context?.content) {
    emit('copy-all', props.context.content)
  }
}

function getChunkPreview(content: string): string {
  const lines = content.split('\n')
  const previewLines = lines.slice(0, 2)
  let preview = previewLines.join(' ').trim()
  
  if (preview.length > 60) {
    preview = preview.substring(0, 57) + '...'
  } else if (lines.length > 2) {
    preview += '...'
  }
  
  return preview || 'Empty chunk'
}
</script>

<style scoped>
.grid-layout {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.grid-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 16px;
  flex: 1;
  min-height: 0;
  padding: 16px;
}

.grid-panel {
  background: rgba(15, 23, 42, 0.6);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.panel-header {
  background: linear-gradient(145deg, rgba(51, 65, 85, 0.9), rgba(30, 41, 59, 0.8));
  border-bottom: 1px solid rgba(148, 163, 184, 0.15);
  padding: 12px 16px;
  flex-shrink: 0;
}

.panel-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(203, 213, 225);
  margin: 0;
}

.panel-content {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.chunk-metadata {
  padding: 16px;
  background: rgba(30, 41, 59, 0.4);
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  flex-shrink: 0;
}

.metadata-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
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

.chunk-viewer {
  height: 100%;
}

.chunk-navigator {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.chunk-item {
  padding: 12px;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 6px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.chunk-item:hover {
  background: rgba(51, 65, 85, 0.8);
  border-color: rgba(148, 163, 184, 0.3);
}

.chunk-item.active {
  background: rgba(59, 130, 246, 0.2);
  border-color: rgba(59, 130, 246, 0.5);
}

.chunk-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
}

.chunk-number {
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(168, 85, 247);
}

.chunk-tokens {
  font-size: 0.75rem;
  color: rgb(148, 163, 184);
}

.chunk-preview {
  font-size: 0.75rem;
  color: rgb(148, 163, 184);
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.actions-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 24px;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border: 1px solid transparent;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
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

.action-btn.secondary:hover {
  background: rgba(51, 65, 85, 0.8);
  border-color: rgba(148, 163, 184, 0.4);
}

.btn-icon {
  width: 16px;
  height: 16px;
}

.content-stats {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.875rem;
  color: rgb(148, 163, 184);
}

.stat-icon {
  width: 16px;
  height: 16px;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-message {
  font-size: 0.875rem;
  color: rgb(100, 116, 139);
}

/* Responsive design */
@media (max-width: 768px) {
  .grid-container {
    grid-template-columns: 1fr;
    grid-template-rows: auto;
    gap: 12px;
    padding: 12px;
  }
  
  .panel-header {
    padding: 8px 12px;
  }
  
  .chunk-navigator {
    padding: 4px;
  }
  
  .chunk-item {
    padding: 8px;
    margin-bottom: 6px;
  }
  
  .actions-panel {
    padding: 12px;
    gap: 16px;
  }
}
</style>