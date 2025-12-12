<template>
  <div class="h-full flex flex-col bg-transparent relative"
       @dragover.prevent="handleDragOver"
       @dragleave="handleDragLeave"
       @drop.prevent="handleDrop">
    
    <!-- Drop Zone Overlay -->
    <div v-if="isDragging" class="drop-zone-overlay">
      <div class="drop-zone-content">
        <svg class="w-12 h-12 text-indigo-400 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
            d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
        </svg>
        <p class="text-lg font-semibold text-white">{{ t('context.dropFiles') }}</p>
        <p class="text-sm text-gray-400">{{ t('context.dropFilesHint') }}</p>
      </div>
    </div>
    
    <!-- Header -->
    <div class="panel-header-unified">
      <div class="panel-header-unified-title">
        <div class="section-icon section-icon-indigo">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
        </div>
        <span>{{ t('context.preview') }}</span>
      </div>

      <div v-if="contextStore.hasContext" class="flex items-center gap-1.5 ml-2 overflow-x-auto">
        <!-- Search toggle -->
        <button @click="showSearch = !showSearch" class="icon-btn"
          :class="{ 'text-indigo-400 bg-indigo-500/10': showSearch }" :title="t('context.search')">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </button>

        <!-- Stats (compact badges) -->
        <div class="hidden xl:flex items-center gap-1.5 flex-shrink-0">
          <span class="chip-unified chip-unified-accent">{{ contextStore.fileCount }} {{ t('context.files') }}</span>
          <span class="chip-unified chip-unified-accent">{{ contextStore.lineCount }} {{ t('context.lines') }}</span>
          <span class="chip-unified chip-unified-accent">{{ contextStore.tokenCount }} {{ t('context.tokens') }}</span>
        </div>

        <!-- Current Format Badge -->
        <span class="chip-unified chip-unified-primary" :title="t('context.currentFormat')">
          {{ formatLabels[settingsStore.settings.context.outputFormat] || 'XML' }}
        </span>

        <!-- Actions (icon only) -->
        <div class="flex gap-0.5 flex-shrink-0">
          <button @click="handleCopyText" class="icon-btn" :title="t('context.copy')">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          </button>
          <button @click="handleExport" class="icon-btn" :title="t('context.export')">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
          </button>
          <button @click="contextStore.clearContext" class="icon-btn icon-btn-danger" :title="t('context.clear')">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Search Bar -->
    <div v-if="showSearch" class="search-bar">
      <div class="relative">
        <svg class="search-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input v-model="searchQuery" type="text" :placeholder="t('context.search')" class="input pl-8 pr-20 text-sm"
          @keyup.enter="searchNext" @keyup.escape="showSearch = false" ref="searchInput" />
        <div v-if="searchQuery && searchResults.length > 0" class="search-nav">
          <span class="search-count">{{ currentSearchIndex + 1 }}/{{ searchResults.length }}</span>
          <button @click="searchPrev" class="search-nav-btn">
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
            </svg>
          </button>
          <button @click="searchNext" class="search-nav-btn">
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>
        <div v-else-if="searchQuery && searchResults.length === 0" class="search-no-results">
          <span>{{ t('context.noResults') }}</span>
        </div>
      </div>
    </div>

    <div class="flex-1 scrollable-y p-4" ref="contentContainer">
      <!-- Smart Context Builder - показывает рекомендации файлов -->
      <SmartContextBuilder 
        v-if="selectedFiles.length > 0"
        :selected-files="selectedFiles"
        @add-files="handleAddSuggestedFiles"
      />
      
      <!-- Impact Preview - показывает анализ влияния -->
      <ImpactPreview 
        v-if="selectedFiles.length > 0"
        :selected-files="selectedFiles"
      />

      <div v-if="contextStore.isBuilding" class="flex items-center justify-center h-full">
        <div class="text-center">
          <svg class="loading-spinner mx-auto mb-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          <p class="text-gray-400 mb-2">Building context...</p>
          <div class="w-64 h-2 bg-gray-700 rounded-full overflow-hidden mx-auto">
            <div class="h-full bg-indigo-500 transition-all duration-300"
              :style="{ width: `${contextStore.buildProgress}%` }"></div>
          </div>
          <p class="text-xs text-gray-500 mt-2">{{ contextStore.buildProgress }}%</p>
        </div>
      </div>

      <div
        v-else-if="(contextStore.fileCount === 0 || contextStore.totalSize === 0 || contextStore.lineCount === 0) && contextStore.contextId"
        class="flex items-center justify-center h-full">
        <div class="text-center max-w-md mx-auto px-4">
          <div
            class="w-16 h-16 mx-auto mb-4 bg-amber-500/20 rounded-2xl flex items-center justify-center border border-amber-500/30">
            <svg class="w-8 h-8 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z">
              </path>
            </svg>
          </div>
          <p class="text-lg font-semibold text-amber-400 mb-4">{{ t('error.emptyContext') }}</p>

          <div class="context-stats mb-4">
            <div class="stats-grid">
              <div class="stat-card">
                <div class="stat-card-value">{{ contextStore.fileCount }}</div>
                <div class="stat-label">{{ t('context.files') }}</div>
              </div>
              <div class="stat-card">
                <div class="stat-card-value">{{ contextStore.lineCount }}</div>
                <div class="stat-label">{{ t('context.lines') }}</div>
              </div>
            </div>
          </div>

          <div class="info-box text-left">
            <p class="text-sm font-medium text-gray-300 mb-2">{{ t('error.suggestions') }}:</p>
            <ul class="text-sm text-gray-400 space-y-1.5">
              <li>• {{ t('error.checkFiles') }}</li>
              <li>• {{ t('error.checkPaths') }}</li>
              <li>• {{ t('error.tryRefresh') }}</li>
            </ul>
          </div>
        </div>
      </div>

      <div v-else-if="contextStore.error" class="flex items-center justify-center h-full">
        <div class="text-center max-w-md">
          <div
            class="w-16 h-16 mx-auto mb-4 bg-red-500/20 rounded-2xl flex items-center justify-center border border-red-500/30">
            <svg class="w-8 h-8 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
          </div>
          <p class="text-lg font-semibold text-red-400 mb-2">{{ contextStore.error }}</p>

          <div class="info-box text-left mt-4">
            <p class="text-sm font-medium text-gray-300 mb-2">{{ t('error.suggestions') }}:</p>
            <ul class="text-sm text-gray-400 space-y-1.5">
              <li>• {{ t('error.checkFiles') }}</li>
              <li>• {{ t('error.checkPaths') }}</li>
              <li>• {{ t('error.tryRefresh') }}</li>
            </ul>
          </div>
        </div>
      </div>

      <div v-else-if="!contextStore.hasContext" class="empty-state-enhanced h-full">
        <div class="empty-state-icon-glow mb-4">
          <svg class="w-8 h-8 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z">
            </path>
          </svg>
        </div>
        <p class="text-base font-semibold text-white mb-3">{{ t('context.notBuilt') }}</p>
        
        <!-- Step-by-step instructions -->
        <div class="text-left max-w-xs space-y-2 mb-6">
          <div class="flex items-center gap-3 text-sm">
            <span class="flex-shrink-0 w-5 h-5 rounded-full bg-indigo-500/20 text-indigo-400 text-xs font-bold flex items-center justify-center">1</span>
            <span class="text-gray-400">{{ t('context.step1') }}</span>
          </div>
          <div class="flex items-center gap-3 text-sm">
            <span class="flex-shrink-0 w-5 h-5 rounded-full bg-indigo-500/20 text-indigo-400 text-xs font-bold flex items-center justify-center">2</span>
            <span class="text-gray-400">{{ t('context.step2') }}</span>
          </div>
        </div>

        <!-- Hint arrows -->
        <div class="flex items-center gap-6 text-gray-500">
          <div class="flex items-center gap-2">
            <svg class="w-4 h-4 animate-[pulse_3s_ease-in-out_infinite]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            <span class="text-xs">{{ t('context.selectHint') }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs">{{ t('context.chatHint') }}</span>
            <svg class="w-4 h-4 animate-[pulse_3s_ease-in-out_infinite]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
            </svg>
          </div>
        </div>
      </div>

      <div v-else class="code-editor context-content">
        <div v-if="contextStore.currentChunk && contextStore.currentChunk.lines">
          <div v-for="(line, index) in contextStore.currentChunk.lines" :key="index"
            :ref="el => setLineRef(el, contextStore.currentChunk?.startLine != null ? contextStore.currentChunk.startLine + index : null)"
            class="code-line" :class="{
              'code-line-highlight': contextStore.currentChunk?.startLine != null && isLineHighlighted(contextStore.currentChunk.startLine + index)
            }">
            <span class="line-number">{{ (contextStore.currentChunk?.startLine ?? 0) + index + 1 }}</span>
            <span v-html="highlightLine(line)"></span>
          </div>
          <div v-if="contextStore.currentChunk.hasMore" class="text-center text-gray-500 py-4">
            <button @click="loadMore" class="btn btn-secondary btn-sm">
              {{ t('context.loadMore') }}
            </button>
          </div>
        </div>
        <div v-else-if="contextStore.isLoading" class="text-center py-8">
          <svg class="animate-spin h-6 w-6 text-blue-500 mx-auto mb-2" xmlns="http://www.w3.org/2000/svg" fill="none"
            viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          <p class="text-gray-400">{{ t('context.loading') }}</p>
        </div>
        <div v-else class="text-center py-8">
          <p v-if="contextStore.contextId" class="text-gray-400">{{ t('context.loading') }}</p>
          <p v-else class="text-gray-400">{{ t('context.notBuilt') }}</p>
        </div>
      </div>
      <ExportModal ref="exportModalRef" />
    </div>
  </div>
</template>

<script setup lang="ts">
import ExportModal from '@/components/ExportModal.vue'
import { useI18n } from '@/composables/useI18n'
import { useLogger } from '@/composables/useLogger'
import { useFileStore } from '@/features/files/model/file.store'
import { useSettingsStore } from '@/stores/settings.store'
import type { ComponentPublicInstance } from 'vue'
import { computed, nextTick, ref, watch } from 'vue'
import { formatContextSize } from '../lib/context-utils'
import { useContextStore } from '../model/context.store'
import ImpactPreview from './ImpactPreview.vue'
import SmartContextBuilder from './SmartContextBuilder.vue'

const logger = useLogger('ContextPanel')
const contextStore = useContextStore()
const settingsStore = useSettingsStore()
const fileStore = useFileStore()
const { t } = useI18n()
const exportModalRef = ref<InstanceType<typeof ExportModal> | null>(null)

// Selected files for SmartContextBuilder and ImpactPreview
const selectedFiles = computed(() => Array.from(fileStore.selectedPaths))

// Handle adding suggested files
function handleAddSuggestedFiles(files: string[]) {
  files.forEach(path => {
    if (!fileStore.selectedPaths.has(path)) {
      fileStore.toggleSelect(path)
    }
  })
}

// Search state
const showSearch = ref(false)
const searchQuery = ref('')
const searchResults = ref<number[]>([]) // Line numbers with matches
const currentSearchIndex = ref(0)
const searchInput = ref<HTMLInputElement | null>(null)
const lineRefs = ref<Map<number, HTMLElement>>(new Map())

// Drag & drop state
const isDragging = ref(false)
let dragCounter = 0

// Format labels for display
const formatLabels: Record<string, string> = {
  xml: 'XML',
  markdown: 'Markdown',
  plain: 'Plain',
  json: 'JSON'
}

// Search functionality
watch(searchQuery, (query) => {
  if (!query || !contextStore.currentChunk?.lines) {
    searchResults.value = []
    currentSearchIndex.value = 0
    return
  }

  const results: number[] = []
  const lowerQuery = query.toLowerCase()

  contextStore.currentChunk.lines.forEach((line, index) => {
    if (line.toLowerCase().includes(lowerQuery)) {
      results.push(contextStore.currentChunk!.startLine + index)
    }
  })

  searchResults.value = results
  currentSearchIndex.value = 0

  if (results.length > 0) {
    scrollToLine(results[0])
  }
})

watch(showSearch, (show) => {
  if (show) {
    nextTick(() => searchInput.value?.focus())
  } else {
    searchQuery.value = ''
  }
})

function setLineRef(el: Element | ComponentPublicInstance | null, lineNum: number | null) {
  if (el && lineNum !== null && lineNum !== undefined) {
    lineRefs.value.set(lineNum, el as HTMLElement)
  }
}

function isLineHighlighted(lineNum: number): boolean {
  return searchResults.value.includes(lineNum) &&
    searchResults.value[currentSearchIndex.value] === lineNum
}

function highlightLine(line: string): string {
  let result = escapeHtml(line)
  
  // Apply syntax highlighting based on current format
  const format = settingsStore.settings.context.outputFormat
  result = applySyntaxHighlight(result, format)
  
  // Apply search highlighting on top
  if (searchQuery.value) {
    const query = escapeHtml(searchQuery.value)
    const regex = new RegExp(`(${escapeRegex(query)})`, 'gi')
    result = result.replace(regex, '<mark class="bg-yellow-400/40 text-yellow-200 rounded px-0.5">$1</mark>')
  }
  
  return result
}

function applySyntaxHighlight(line: string, format: string): string {
  switch (format) {
    case 'xml':
      return highlightXml(line)
    case 'markdown':
      return highlightMarkdown(line)
    case 'json':
      return highlightJson(line)
    default:
      return highlightPlain(line)
  }
}

function highlightXml(line: string): string {
  // Highlight XML tags: <file path="..."> </file> <content> </content>
  return line
    .replace(/(&lt;\/?)(file|content)(&gt;)/g, '<span class="syntax-tag">$1$2$3</span>')
    .replace(/(&lt;file\s+)(path)(=)(&quot;[^&]*&quot;)(&gt;)/g, 
      '<span class="syntax-tag">$1</span><span class="syntax-attr">$2</span><span class="syntax-punct">$3</span><span class="syntax-string">$4</span><span class="syntax-tag">$5</span>')
    .replace(/(path)(=)(&quot;[^&]*&quot;)/g,
      '<span class="syntax-attr">$1</span><span class="syntax-punct">$2</span><span class="syntax-string">$3</span>')
}

function highlightMarkdown(line: string): string {
  // Highlight markdown: ## File: path, ```lang, ```
  if (line.startsWith('## File:') || line.startsWith('# ')) {
    return `<span class="syntax-heading">${line}</span>`
  }
  if (line.startsWith('```')) {
    return `<span class="syntax-fence">${line}</span>`
  }
  return line
}

function highlightJson(line: string): string {
  // Highlight JSON keys and strings
  return line
    .replace(/(&quot;[^&]+&quot;)(\s*:)/g, '<span class="syntax-key">$1</span><span class="syntax-punct">$2</span>')
    .replace(/:(\s*)(&quot;[^&]*&quot;)/g, ':<span class="syntax-string">$1$2</span>')
    .replace(/:\s*(true|false|null|\d+)/g, ': <span class="syntax-value">$1</span>')
}

function highlightPlain(line: string): string {
  // Highlight plain text separators: --- File: path ---
  if (line.startsWith('--- File:') && line.endsWith('---')) {
    return `<span class="syntax-separator">${line}</span>`
  }
  return line
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

function escapeRegex(text: string): string {
  return text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function scrollToLine(lineNum: number) {
  nextTick(() => {
    const el = lineRefs.value.get(lineNum)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  })
}

function searchNext() {
  if (searchResults.value.length === 0) return
  currentSearchIndex.value = (currentSearchIndex.value + 1) % searchResults.value.length
  scrollToLine(searchResults.value[currentSearchIndex.value])
}

function searchPrev() {
  if (searchResults.value.length === 0) return
  currentSearchIndex.value = (currentSearchIndex.value - 1 + searchResults.value.length) % searchResults.value.length
  scrollToLine(searchResults.value[currentSearchIndex.value])
}

// Debug logging to track state changes (dev only)
if (import.meta.env.DEV) {
  watch(() => [contextStore.contextId, contextStore.fileCount, contextStore.totalSize, contextStore.lineCount, contextStore.error, contextStore.isBuilding],
    ([contextId, fileCount, totalSize, lineCount, error, isBuilding]) => {
      logger.debug('State changed:', { contextId, fileCount, totalSize, lineCount, error, isBuilding })
    },
    { immediate: true }
  )
}

// Helper functions available for future use
void formatContextSize // suppress unused import warning

async function loadMore() {
  if (contextStore.contextId && contextStore.currentChunk) {
    try {
      logger.debug('Loading more lines from', contextStore.currentChunk.endLine)
      await contextStore.loadContextContent(
        contextStore.contextId,
        contextStore.currentChunk.endLine,
        1000
      )
      logger.debug('More lines loaded successfully')
    } catch (error) {
      logger.error('Failed to load more lines:', error)
    }
  }
}

async function handleExport() {
  try {
    if (exportModalRef.value) {
      exportModalRef.value.open()
    }
  } catch (error) {
    logger.error('Failed to open export modal:', error)
  }
}

async function handleCopyText() {
  if (contextStore.contextId) {
    try {
      const content = await contextStore.getFullContextContent()
      await navigator.clipboard.writeText(content)
      logger.debug('Context copied, length:', content.length)
    } catch (error) {
      logger.error('Failed to copy context:', error)
    }
  }
}

// Drag & drop handlers
function handleDragOver(e: DragEvent) {
  dragCounter++
  isDragging.value = true
  
  // Check if dragging files
  if (e.dataTransfer?.types.includes('Files') || e.dataTransfer?.types.includes('text/plain')) {
    e.dataTransfer.dropEffect = 'copy'
  }
}

function handleDragLeave() {
  dragCounter--
  if (dragCounter <= 0) {
    dragCounter = 0
    isDragging.value = false
  }
}

async function handleDrop(e: DragEvent) {
  isDragging.value = false
  dragCounter = 0
  
  if (!e.dataTransfer) return
  
  // Handle file paths from internal drag (text/plain with file paths)
  const textData = e.dataTransfer.getData('text/plain')
  if (textData) {
    const paths = textData.split('\n').filter(p => p.trim())
    if (paths.length > 0) {
      logger.debug('Dropped file paths:', paths.length)
      // Emit event to add files to selection and build context
      window.dispatchEvent(new CustomEvent('add-files-to-context', { detail: { paths } }))
      return
    }
  }
  
  // Handle external file drops (from OS file manager)
  if (e.dataTransfer.files.length > 0) {
    const filePaths = Array.from(e.dataTransfer.files).map(f => f.name)
    logger.debug('Dropped external files:', filePaths.length)
    // Note: External files need special handling - for now just log
  }
}
</script>


<style scoped>
/* Search Bar */
.search-bar {
  @apply p-2;
  background: var(--bg-1);
  border-bottom: 1px solid var(--border-default);
}

.search-icon {
  @apply absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4;
  color: var(--text-muted);
}

.search-nav {
  @apply absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1;
}

.search-count {
  @apply text-xs;
  color: var(--text-muted);
}

.search-nav-btn {
  @apply p-1 rounded;
  color: var(--text-muted);
  transition: all 150ms ease-out;
}

.search-nav-btn:hover {
  background: var(--bg-2);
  color: var(--text-primary);
}

.search-no-results {
  @apply absolute right-2 top-1/2 -translate-y-1/2 text-xs;
  color: var(--text-subtle);
}

/* Code Editor */
.code-editor {
  @apply rounded-lg p-4 font-mono text-sm;
  background: var(--bg-1);
  color: var(--text-secondary);
}

.code-line {
  @apply px-2 py-0.5 rounded;
  transition: background-color 150ms ease-out;
}

.code-line:hover:not(.code-line-highlight) {
  background: var(--bg-2);
}

.code-line-highlight {
  background: rgba(250, 204, 21, 0.15);
  border-left: 2px solid var(--color-warning);
}

.line-number {
  @apply mr-4 select-none;
  color: var(--text-subtle);
}
</style>
