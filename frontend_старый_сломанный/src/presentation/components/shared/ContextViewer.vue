<template>
  <div class="context-viewer">
    <!-- Memory usage indicator and context info -->
    <div class="context-header">
      <div 
        v-if="contextSummary" 
        class="context-info"
      >
        <span class="file-count">{{ contextSummary.fileCount }} files</span>
        <span class="size-info">{{ formatSize(contextSummary.totalSize) }}</span>
        <span class="token-count">~{{ contextSummary.tokenCount }} tokens</span>
      </div>
      
      <div 
        v-if="showMemoryInfo && memoryUsage > 0" 
        :class="['memory-indicator', memoryStatusClass]"
        :title="memoryTooltip"
      >
        Memory: {{ formatMemoryUsage }} MB
      </div>
    </div>
    
    <!-- Paginated content display - CRITICAL OOM FIX -->
    <div class="content-container">
      <!-- Loading state -->
      <div v-if="isLoading" class="loading-state">
        <div class="loading-spinner"></div>
        <span>Loading content...</span>
      </div>
      
      <!-- Error state -->
      <div v-else-if="error" class="error-state">
        <span class="error-message">{{ error }}</span>
        <button class="retry-button" @click="retry">Retry</button>
      </div>
      
      <!-- Virtual scroll viewer for paginated content -->
      <div v-else-if="currentChunk" class="paginated-viewer">
        <!-- Pagination controls -->
        <div class="pagination-controls">
          <button 
            :disabled="currentPage <= 1" 
            class="nav-button"
            @click="loadPreviousChunk"
          >
            ← Previous
          </button>
          
          <span class="page-info">
            Page {{ currentPage }} of {{ totalPages }}
            (Lines {{ currentChunk.startLine + 1 }}-{{ currentChunk.endLine + 1 }})
          </span>
          
          <button 
            :disabled="!hasMore" 
            class="nav-button"
            @click="loadNextChunk"
          >
            Next →
          </button>
        </div>
        
        <!-- Content display with syntax highlighting -->
        <pre class="content-display">
          <code 
            v-for="(line, index) in currentChunk.lines" 
            :key="currentChunk.startLine + index"
            class="code-line"
            :data-line="currentChunk.startLine + index + 1"
          >{{ line }}</code>
        </pre>
        
        <!-- Progress indicator -->
        <div class="progress-bar">
          <div 
            class="progress-fill" 
            :style="{ width: progress + '%' }"
          ></div>
        </div>
      </div>
      
      <!-- No context state -->
      <div v-else class="no-content-state">
        <span>No context available</span>
        <p class="text-sm text-gray-500">Build a context to view content here</p>
      </div>
    </div>
    
    <!-- Controls -->
    <div class="viewer-controls">
      <div class="left-controls">
        <button 
          v-if="currentChunk" 
          class="control-button"
          @click="goToLine"
        >
          Go to Line
        </button>
        
        <button 
          v-if="currentChunk" 
          class="control-button"
          @click="searchInContent"
        >
          Search
        </button>
      </div>
      
      <div class="right-controls">
        <button 
          v-if="currentChunk" 
          :disabled="isLoading"
          class="control-button copy-button"
          @click="copyCurrentChunk"
        >
          Copy Chunk
        </button>
        
        <button 
          v-if="contextSummary" 
          :disabled="isLoading"
          class="control-button"
          @click="exportContext"
        >
          Export Full
        </button>
      </div>
    </div>
    
    <!-- Search modal -->
    <div v-if="showSearchModal" class="search-modal-overlay" @click="closeSearchModal">
      <div class="search-modal" @click.stop>
        <input 
          v-model="searchQuery" 
          placeholder="Search in current chunk..."
          class="search-input"
          @keyup.enter="performSearch"
        />
        <div class="search-results">
          <div 
            v-for="result in searchResults" 
            :key="result.line"
            class="search-result"
            @click="goToSearchResult(result)"
          >
            <span class="line-number">Line {{ result.line }}:</span>
            <span class="line-text">{{ result.text }}</span>
          </div>
        </div>
        <div class="search-actions">
          <button class="search-button" @click="performSearch">Search</button>
          <button class="close-button" @click="closeSearchModal">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { usePaginatedContext } from '@/composables/usePaginatedContext'
import { useContextBuilderStore } from '@/stores/context-builder.store'

// CRITICAL OOM FIX: New interface using ContextSummary instead of full content
interface Props {
  contextId?: string
  containerHeight?: number
  linesPerPage?: number
  showMemoryInfo?: boolean
  autoLoad?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  containerHeight: 400,
  linesPerPage: 100, // Reduced chunk size for better memory management
  showMemoryInfo: true,
  autoLoad: true
})

const emit = defineEmits<{
  (e: 'copy'): void
  (e: 'export'): void
  (e: 'memory-warning', usage: number): void
  (e: 'line-navigate', line: number): void
}>()

const contextStore = useContextBuilderStore()

// Use the new paginated context composable - MEMORY SAFE
const {
  currentChunk,
  isLoading,
  error,
  currentPage,
  totalPages,
  progress,
  hasMore,
  loadNextChunk,
  loadPreviousChunk,
  goToLine: navigateToLine,
  searchInChunk
} = usePaginatedContext(props.contextId, props.linesPerPage)

// Get context summary from store instead of full content
const contextSummary = computed(() => contextStore.contextSummaryState)

// Memory monitoring (much lighter now)
const memoryUsage = ref(0)
const memoryWarning = ref(false)

// Search functionality
const showSearchModal = ref(false)
const searchQuery = ref('')
const searchResults = ref<{ line: number; text: string }[]>([])

// Computed properties
const formatMemoryUsage = computed(() => memoryUsage.value.toFixed(1))
const memoryStatusClass = computed(() => {
  if (memoryUsage.value > 60) return 'bg-red-600 text-white'
  if (memoryUsage.value > 40) return 'bg-orange-600 text-white' 
  if (memoryUsage.value > 25) return 'bg-yellow-600 text-white'
  return 'bg-green-600 text-white'
})

const memoryTooltip = computed(() => {
  return `Memory usage: ${formatMemoryUsage.value} MB`
})

// Methods
const formatSize = (bytes: number): string => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

const copyCurrentChunk = async () => {
  if (!currentChunk.value) return
  
  try {
    const content = currentChunk.value.lines.join('\n')
    await navigator.clipboard.writeText(content)
    emit('copy')
  } catch (error) {
    console.error('Failed to copy chunk:', error)
  }
}

const exportContext = () => {
  emit('export')
}

const retry = async () => {
  if (props.contextId) {
    // Retry loading the current chunk
    await navigateToLine(currentChunk.value?.startLine || 0)
  }
}

const goToLine = () => {
  const lineNumber = prompt('Enter line number:')
  if (lineNumber) {
    const line = parseInt(lineNumber, 10) - 1 // Convert to 0-based
    if (!isNaN(line) && line >= 0) {
      navigateToLine(line)
      emit('line-navigate', line)
    }
  }
}

const searchInContent = () => {
  showSearchModal.value = true
}

const performSearch = () => {
  if (!searchQuery.value || !currentChunk.value) {
    searchResults.value = []
    return
  }
  
  searchResults.value = searchInChunk(searchQuery.value)
}

const goToSearchResult = (result: { line: number; text: string }) => {
  navigateToLine(result.line)
  emit('line-navigate', result.line)
  closeSearchModal()
}

const closeSearchModal = () => {
  showSearchModal.value = false
  searchQuery.value = ''
  searchResults.value = []
}

const updateMemoryUsage = () => {
  if ('performance' in window && 'memory' in (performance as any)) {
    const memory = (performance as any).memory
    const newUsage = memory.usedJSHeapSize / 1024 / 1024 // MB
    memoryUsage.value = newUsage
    
    // Emit warning if memory usage is high
    if (newUsage > 60) {
      memoryWarning.value = true
      emit('memory-warning', newUsage)
    }
  }
}

// Memory monitoring
let memoryMonitorInterval: number | null = null

onMounted(() => {
  updateMemoryUsage()
  memoryMonitorInterval = window.setInterval(updateMemoryUsage, 3000)
})

onUnmounted(() => {
  if (memoryMonitorInterval) {
    clearInterval(memoryMonitorInterval)
  }
})

defineExpose({
  goToLine: navigateToLine,
  searchInChunk,
  memoryUsage: computed(() => memoryUsage.value),
  currentChunk: computed(() => currentChunk.value)
})
</script>

<style scoped>
.context-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  overflow: hidden;
  background-color: #ffffff;
}

.context-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem;
  background-color: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.context-info {
  display: flex;
  gap: 1rem;
  font-size: 0.875rem;
  color: #64748b;
}

.file-count, .size-info, .token-count {
  font-weight: 500;
}

.memory-indicator {
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 0.25rem;
}

.content-container {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.loading-state, .error-state, .no-content-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 2rem;
  color: #64748b;
}

.loading-spinner {
  width: 2rem;
  height: 2rem;
  border: 2px solid #e2e8f0;
  border-top: 2px solid #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 0.5rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-message {
  color: #dc2626;
  margin-bottom: 1rem;
}

.retry-button {
  padding: 0.5rem 1rem;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.25rem;
  cursor: pointer;
}

.retry-button:hover {
  background-color: #2563eb;
}

.paginated-viewer {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.pagination-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background-color: #f1f5f9;
  border-bottom: 1px solid #e2e8f0;
  font-size: 0.875rem;
}

.nav-button {
  padding: 0.25rem 0.75rem;
  background-color: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s;
}

.nav-button:hover:not(:disabled) {
  background-color: #f8fafc;
  border-color: #94a3b8;
}

.nav-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  font-weight: 500;
  color: #475569;
}

.content-display {
  flex: 1;
  margin: 0;
  padding: 1rem;
  overflow: auto;
  white-space: pre;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
  background-color: #fafafa;
  color: #1e293b;
}

.code-line {
  display: block;
  position: relative;
  padding-left: 3rem;
}

.code-line::before {
  content: attr(data-line);
  position: absolute;
  left: 0;
  width: 2.5rem;
  text-align: right;
  color: #94a3b8;
  font-size: 0.75rem;
  user-select: none;
}

.progress-bar {
  height: 2px;
  background-color: #e2e8f0;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: #3b82f6;
  transition: width 0.3s ease;
}

.viewer-controls {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem;
  border-top: 1px solid #e2e8f0;
  background-color: #f8fafc;
}

.left-controls, .right-controls {
  display: flex;
  gap: 0.5rem;
}

.control-button {
  padding: 0.375rem 0.75rem;
  font-size: 0.875rem;
  border-radius: 0.25rem;
  border: 1px solid #cbd5e1;
  background-color: #ffffff;
  cursor: pointer;
  transition: all 0.2s;
}

.control-button:hover:not(:disabled) {
  background-color: #f8fafc;
  border-color: #94a3b8;
}

.control-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.copy-button {
  background-color: #10b981;
  color: white;
  border-color: #10b981;
}

.copy-button:hover:not(:disabled) {
  background-color: #059669;
  border-color: #059669;
}

/* Search Modal */
.search-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.search-modal {
  background-color: white;
  border-radius: 0.5rem;
  padding: 1.5rem;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.search-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
  font-size: 1rem;
  margin-bottom: 1rem;
}

.search-results {
  max-height: 300px;
  overflow-y: auto;
  margin-bottom: 1rem;
}

.search-result {
  display: flex;
  flex-direction: column;
  padding: 0.5rem;
  border-bottom: 1px solid #e2e8f0;
  cursor: pointer;
  transition: background-color 0.2s;
}

.search-result:hover {
  background-color: #f8fafc;
}

.line-number {
  font-size: 0.75rem;
  color: #64748b;
  font-weight: 500;
}

.line-text {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.875rem;
  white-space: pre-wrap;
  margin-top: 0.25rem;
}

.search-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.search-button {
  padding: 0.5rem 1rem;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.25rem;
  cursor: pointer;
}

.search-button:hover {
  background-color: #2563eb;
}

.close-button {
  padding: 0.5rem 1rem;
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 0.25rem;
  cursor: pointer;
}

.close-button:hover {
  background-color: #4b5563;
}
</style>