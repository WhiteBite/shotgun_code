<template>
  <div class="split-view-manager" :class="managerClasses">
    <!-- Split Toolbar -->
    <div class="split-toolbar">
      <div class="toolbar-section">
        <div class="layout-controls">
          <button
            v-for="layoutOption in layoutOptions"
            :key="layoutOption.value"
            @click="handleLayoutChange(layoutOption.value)"
            :class="['layout-btn', { active: currentLayout === layoutOption.value }]"
            :title="layoutOption.title"
          >
            <component :is="layoutOption.icon" class="btn-icon" />
            <span class="btn-label">{{ layoutOption.label }}</span>
          </button>
        </div>
        
        <div v-if="chunks.length > 1" class="chunk-controls">
          <span class="chunk-label">Chunks:</span>
          <div class="chunk-selector">
            <button
              v-for="(chunk, index) in chunks"
              :key="chunk.id"
              @click="selectChunk(index)"
              :class="['chunk-btn', { active: activeChunkIndex === index }]"
              :title="`Chunk ${index + 1}: ${chunk.tokens} tokens`"
            >
              {{ index + 1 }}
            </button>
          </div>
        </div>
      </div>
      
      <div class="toolbar-actions">
        <button
          @click="copyActiveChunk"
          :disabled="!activeChunk"
          class="action-btn secondary"
          title="Copy current chunk"
        >
          <CopyIcon class="btn-icon" />
          Copy Chunk
        </button>
        
        <button
          @click="copyAllContent"
          class="action-btn primary"
          title="Copy all content"
        >
          <FileTextIcon class="btn-icon" />
          Copy All
        </button>
      </div>
    </div>

    <!-- Split Content Area -->
    <div class="split-content-area">
      <component 
        :is="layoutComponent" 
        :context="context"
        :chunks="chunks"
        :active-chunk="activeChunk"
        :active-chunk-index="activeChunkIndex"
        :split-settings="splitSettings"
        @chunk-select="selectChunk"
        @chunk-hover="handleChunkHover"
        @copy-chunk="handleCopyChunk"
        @copy-selection="handleCopySelection"
      />
    </div>
    
    <!-- Split Status Bar -->
    <div class="split-status-bar">
      <div class="status-section">
        <div class="content-stats">
          <span class="stat-item">
            <FileIcon class="stat-icon" />
            {{ context?.fileCount || 0 }} files
          </span>
          <span class="stat-item">
            <HashIcon class="stat-icon" />
            {{ totalTokens }} tokens
          </span>
          <span v-if="chunks.length > 1" class="stat-item">
            <LayersIcon class="stat-icon" />
            {{ chunks.length }} chunks
          </span>
        </div>
        
        <div v-if="activeChunk" class="active-chunk-info">
          <span class="chunk-info">
            Chunk {{ activeChunkIndex + 1 }}: 
            Lines {{ activeChunk.startLine }}-{{ activeChunk.endLine }} 
            ({{ activeChunk.tokens }} tokens)
          </span>
        </div>
      </div>
      
      <div class="copy-history" v-if="copyHistory.length > 0">
        <span class="history-label">Recent:</span>
        <div class="history-items">
          <span
            v-for="item in recentCopyHistory"
            :key="item.id"
            class="history-item"
            :title="`Copied at ${item.timestamp}`"
          >
            {{ item.type === 'chunk' ? `Chunk ${item.chunkIndex + 1}` : 'All' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, defineAsyncComponent } from 'vue'
import {
  LayoutGridIcon,
  SplitIcon,
  SquareIcon,
  ColumnsIcon,
  RowsIcon,
  CopyIcon,
  FileTextIcon,
  FileIcon,
  HashIcon,
  LayersIcon
} from 'lucide-vue-next'

// Lazy load layout components
const SinglePaneLayout = defineAsyncComponent(() => import('./layouts/SinglePaneLayout.vue'))
const VerticalSplitLayout = defineAsyncComponent(() => import('./layouts/VerticalSplitLayout.vue'))
const HorizontalSplitLayout = defineAsyncComponent(() => import('./layouts/HorizontalSplitLayout.vue'))
const GridLayout = defineAsyncComponent(() => import('./layouts/GridLayout.vue'))

// Types
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
  metadata?: any
}

interface SplitSettings {
  enabled: boolean
  layout?: string
  chunkSize?: number
}

interface CopyAction {
  id: string
  type: 'chunk' | 'all'
  chunkIndex?: number
  timestamp: string
  tokens: number
}

// Props
interface Props {
  context: ContextData
  splitSettings: SplitSettings
}

const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  'copy-chunk': [chunk: ContextChunk]
  'copy-all': [content: string]
  'layout-change': [layout: string]
  'chunk-select': [index: number]
}>()

// State
const currentLayout = ref<string>('single')
const activeChunkIndex = ref<number>(0)
const copyHistory = ref<CopyAction[]>([])
const hoveredChunk = ref<number | null>(null)

// Layout options
const layoutOptions = [
  {
    value: 'single',
    label: 'Single',
    title: 'Single pane view',
    icon: SquareIcon
  },
  {
    value: 'vertical',
    label: 'Vertical',
    title: 'Vertical split view',
    icon: ColumnsIcon
  },
  {
    value: 'horizontal',
    label: 'Horizontal',
    title: 'Horizontal split view',
    icon: RowsIcon
  },
  {
    value: 'grid',
    label: 'Grid',
    title: 'Grid layout view',
    icon: LayoutGridIcon
  }
]

// Computed
const managerClasses = computed(() => [
  `split-layout-${currentLayout.value}`,
  {
    'has-chunks': chunks.value.length > 1,
    'split-enabled': props.splitSettings.enabled
  }
])

const layoutComponent = computed(() => {
  switch (currentLayout.value) {
    case 'vertical':
      return VerticalSplitLayout
    case 'horizontal':
      return HorizontalSplitLayout
    case 'grid':
      return GridLayout
    default:
      return SinglePaneLayout
  }
})

const chunks = computed(() => {
  if (!props.context?.content || !props.splitSettings.enabled) {
    return [{
      id: 'full-content',
      content: props.context?.content || '',
      tokens: estimateTokens(props.context?.content || ''),
      startLine: 1,
      endLine: countLines(props.context?.content || ''),
      metadata: props.context?.metadata
    }]
  }
  
  return chunkContent(props.context.content, props.splitSettings.chunkSize || 1000)
})

const activeChunk = computed(() => {
  return chunks.value[activeChunkIndex.value] || null
})

const totalTokens = computed(() => {
  return chunks.value.reduce((sum, chunk) => sum + chunk.tokens, 0)
})

const recentCopyHistory = computed(() => {
  return copyHistory.value.slice(-3).reverse()
})

// Methods
function handleLayoutChange(layout: string) {
  currentLayout.value = layout
  emit('layout-change', layout)
}

function selectChunk(index: number) {
  if (index >= 0 && index < chunks.value.length) {
    activeChunkIndex.value = index
    emit('chunk-select', index)
  }
}

function handleChunkHover(index: number | null) {
  hoveredChunk.value = index
}

function copyActiveChunk() {
  if (activeChunk.value) {
    handleCopyChunk(activeChunk.value)
  }
}

function copyAllContent() {
  if (props.context?.content) {
    emit('copy-all', props.context.content)
    addToCopyHistory('all', undefined, totalTokens.value)
  }
}

function handleCopyChunk(chunk: ContextChunk) {
  emit('copy-chunk', chunk)
  const chunkIndex = chunks.value.findIndex(c => c.id === chunk.id)
  addToCopyHistory('chunk', chunkIndex, chunk.tokens)
}

function handleCopySelection(selection: string) {
  navigator.clipboard.writeText(selection)
  // Could add selection to copy history
}

function addToCopyHistory(type: 'chunk' | 'all', chunkIndex?: number, tokens: number = 0) {
  const action: CopyAction = {
    id: generateId(),
    type,
    chunkIndex,
    timestamp: new Date().toISOString(),
    tokens
  }
  
  copyHistory.value.push(action)
  
  // Keep only last 10 items
  if (copyHistory.value.length > 10) {
    copyHistory.value = copyHistory.value.slice(-10)
  }
}

function chunkContent(content: string, chunkSize: number): ContextChunk[] {
  const lines = content.split('\n')
  const chunks: ContextChunk[] = []
  let currentChunk = ''
  let currentTokens = 0
  let startLine = 1
  let chunkIndex = 0
  
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    const lineTokens = estimateTokens(line)
    
    // If adding this line would exceed chunk size and we have content
    if (currentTokens + lineTokens > chunkSize && currentChunk) {
      chunks.push({
        id: `chunk-${chunkIndex}`,
        content: currentChunk.trim(),
        tokens: currentTokens,
        startLine,
        endLine: i,
        metadata: {
          chunkIndex,
          linesCount: i - startLine + 1
        }
      })
      
      currentChunk = ''
      currentTokens = 0
      startLine = i + 1
      chunkIndex++
    }
    
    currentChunk += line + '\n'
    currentTokens += lineTokens
  }
  
  // Add remaining content as final chunk
  if (currentChunk.trim()) {
    chunks.push({
      id: `chunk-${chunkIndex}`,
      content: currentChunk.trim(),
      tokens: currentTokens,
      startLine,
      endLine: lines.length,
      metadata: {
        chunkIndex,
        linesCount: lines.length - startLine + 1
      }
    })
  }
  
  return chunks.length > 0 ? chunks : [{
    id: 'chunk-0',
    content: content,
    tokens: estimateTokens(content),
    startLine: 1,
    endLine: lines.length,
    metadata: { chunkIndex: 0, linesCount: lines.length }
  }]
}

function estimateTokens(text: string): number {
  // Simple token estimation: ~4 characters per token
  return Math.ceil(text.length / 4)
}

function countLines(text: string): number {
  return text.split('\n').length
}

function generateId(): string {
  return Math.random().toString(36).substr(2, 9)
}

// Watch for settings changes
watch(
  () => props.splitSettings.layout,
  (newLayout) => {
    if (newLayout && newLayout !== currentLayout.value) {
      currentLayout.value = newLayout
    }
  },
  { immediate: true }
)

// Reset active chunk when chunks change
watch(
  () => chunks.value.length,
  (newLength) => {
    if (activeChunkIndex.value >= newLength) {
      activeChunkIndex.value = Math.max(0, newLength - 1)
    }
  }
)
</script>

<style scoped>
.split-view-manager {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: rgba(15, 23, 42, 0.6);
  border-radius: 8px;
  overflow: hidden;
}

.split-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(145deg, rgba(51, 65, 85, 0.9), rgba(30, 41, 59, 0.8));
  border-bottom: 1px solid rgba(148, 163, 184, 0.15);
  gap: 16px;
}

.toolbar-section {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.layout-controls {
  display: flex;
  gap: 4px;
  background: rgba(15, 23, 42, 0.6);
  border-radius: 8px;
  padding: 4px;
}

.layout-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: rgb(148, 163, 184);
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.layout-btn:hover {
  background: rgba(59, 130, 246, 0.1);
  color: rgb(147, 197, 253);
}

.layout-btn.active {
  background: rgba(59, 130, 246, 0.2);
  color: rgb(147, 197, 253);
}

.chunk-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.chunk-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: rgb(148, 163, 184);
}

.chunk-selector {
  display: flex;
  gap: 2px;
  background: rgba(15, 23, 42, 0.6);
  border-radius: 6px;
  padding: 2px;
}

.chunk-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: rgb(148, 163, 184);
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.chunk-btn:hover {
  background: rgba(168, 85, 247, 0.1);
  color: rgb(196, 181, 253);
}

.chunk-btn.active {
  background: rgba(168, 85, 247, 0.2);
  color: rgb(196, 181, 253);
}

.toolbar-actions {
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

.action-btn.primary {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
}

.action-btn.primary:hover {
  background: linear-gradient(135deg, #2563eb, #7c3aed);
  transform: translateY(-1px);
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

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.btn-icon {
  width: 14px;
  height: 14px;
}

.btn-label {
  display: none;
}

@media (min-width: 640px) {
  .btn-label {
    display: inline;
  }
}

.split-content-area {
  flex: 1;
  min-height: 0;
  position: relative;
}

.split-status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: rgba(15, 23, 42, 0.8);
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  font-size: 0.75rem;
}

.status-section {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.content-stats {
  display: flex;
  gap: 12px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: rgb(100, 116, 139);
}

.stat-icon {
  width: 12px;
  height: 12px;
}

.active-chunk-info {
  color: rgb(148, 163, 184);
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
}

.copy-history {
  display: flex;
  align-items: center;
  gap: 8px;
}

.history-label {
  color: rgb(100, 116, 139);
}

.history-items {
  display: flex;
  gap: 4px;
}

.history-item {
  padding: 2px 6px;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.2);
  border-radius: 4px;
  color: rgb(74, 222, 128);
  font-size: 0.625rem;
}
</style>