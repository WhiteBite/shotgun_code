<template>
  <div class="context-memory-panel">
    <div class="panel-header">
      <div class="flex items-center gap-2">
        <svg class="panel-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        <span class="panel-title">{{ t('context.savedContexts') }}</span>
      </div>
      <button @click="refreshContexts" class="action-btn" :disabled="isLoading">
        <svg class="w-4 h-4" :class="{ 'animate-spin': isLoading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
      </button>
    </div>

    <!-- Search -->
    <div class="px-3 py-2">
      <input 
        v-model="searchQuery" 
        type="text" 
        :placeholder="t('context.searchContexts')"
        class="search-input w-full"
        @input="debouncedSearch"
      />
    </div>

    <!-- Contexts List -->
    <div class="contexts-list">
      <div v-if="isLoading" class="flex items-center justify-center py-8">
        <svg class="animate-spin h-6 w-6 text-indigo-400" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      </div>

      <div v-else-if="filteredContexts.length === 0" class="empty-state">
        <svg class="w-12 h-12 text-gray-600 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" 
            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        <p class="text-sm text-gray-500">{{ t('context.noSavedContexts') }}</p>
      </div>

      <div v-else class="space-y-2 p-2">
        <div 
          v-for="ctx in filteredContexts" 
          :key="ctx.id"
          class="context-card"
          @click="restoreContext(ctx)"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1 min-w-0">
              <h4 class="context-topic">{{ ctx.topic }}</h4>
              <p class="context-summary">{{ ctx.summary }}</p>
            </div>
          </div>
          <div class="context-meta">
            <span class="context-files">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              {{ ctx.files.length }} {{ t('context.files') }}
            </span>
            <span class="context-date">{{ formatDate(ctx.createdAt) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Save Current Context -->
    <div class="save-section">
      <button 
        @click="showSaveDialog = true" 
        class="btn btn-primary w-full"
        :disabled="!canSave"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
        </svg>
        {{ t('context.saveCurrentContext') }}
      </button>
    </div>

    <!-- Save Dialog -->
    <Teleport to="body">
      <div v-if="showSaveDialog" class="modal-overlay" @click.self="showSaveDialog = false">
        <div class="modal-content">
          <h3 class="modal-title">{{ t('context.saveContext') }}</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm text-gray-400 mb-1">{{ t('context.topic') }}</label>
              <input v-model="saveTopic" type="text" class="input w-full" :placeholder="t('context.topicPlaceholder')" />
            </div>
            <div>
              <label class="block text-sm text-gray-400 mb-1">{{ t('context.summary') }}</label>
              <textarea v-model="saveSummary" class="input w-full h-20 resize-none" :placeholder="t('context.summaryPlaceholder')" />
            </div>
            <div class="text-sm text-gray-500">
              {{ selectedFiles.length }} {{ t('context.filesSelected') }}
            </div>
          </div>
          <div class="modal-actions">
            <button @click="showSaveDialog = false" class="btn btn-ghost">{{ t('common.cancel') }}</button>
            <button @click="saveContext" class="btn btn-primary" :disabled="!saveTopic.trim()">{{ t('common.save') }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useFileStore } from '@/features/files/model/file.store'
import { apiService, type ContextMemoryEntry } from '@/services/api.service'
import { useProjectStore } from '@/stores/project.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, onMounted, ref, watch } from 'vue'

const { t } = useI18n()
const fileStore = useFileStore()
const projectStore = useProjectStore()
const uiStore = useUIStore()

const contexts = ref<ContextMemoryEntry[]>([])
const isLoading = ref(false)
const searchQuery = ref('')
const showSaveDialog = ref(false)
const saveTopic = ref('')
const saveSummary = ref('')

const selectedFiles = computed(() => Array.from(fileStore.selectedPaths))
const canSave = computed(() => selectedFiles.value.length > 0 && projectStore.currentPath)

const filteredContexts = computed(() => {
  if (!searchQuery.value.trim()) return contexts.value
  const query = searchQuery.value.toLowerCase()
  return contexts.value.filter(ctx => 
    ctx.topic.toLowerCase().includes(query) || 
    ctx.summary.toLowerCase().includes(query)
  )
})

let searchTimeout: number | null = null
function debouncedSearch() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = window.setTimeout(() => {
    if (searchQuery.value.trim()) {
      searchByTopic()
    } else {
      refreshContexts()
    }
  }, 300)
}

async function refreshContexts() {
  if (!projectStore.currentPath) return
  isLoading.value = true
  try {
    contexts.value = await apiService.getRecentContexts(projectStore.currentPath, 20)
  } finally {
    isLoading.value = false
  }
}

async function searchByTopic() {
  if (!projectStore.currentPath || !searchQuery.value.trim()) return
  isLoading.value = true
  try {
    contexts.value = await apiService.findContextByTopic(projectStore.currentPath, searchQuery.value)
  } finally {
    isLoading.value = false
  }
}

async function saveContext() {
  if (!projectStore.currentPath || !saveTopic.value.trim()) return
  try {
    await apiService.saveContextMemory(
      projectStore.currentPath,
      saveTopic.value.trim(),
      saveSummary.value.trim(),
      selectedFiles.value
    )
    uiStore.addToast(t('context.contextSaved'), 'success')
    showSaveDialog.value = false
    saveTopic.value = ''
    saveSummary.value = ''
    await refreshContexts()
  } catch {
    uiStore.addToast(t('context.saveError'), 'error')
  }
}

function restoreContext(ctx: ContextMemoryEntry) {
  fileStore.clearSelection()
  ctx.files.forEach(file => fileStore.toggleSelect(file))
  uiStore.addToast(t('context.contextRestored'), 'success')
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (days === 0) return t('common.today')
  if (days === 1) return t('common.yesterday')
  if (days < 7) return `${days} ${t('common.daysAgo')}`
  return date.toLocaleDateString()
}

watch(() => projectStore.currentPath, refreshContexts, { immediate: true })
onMounted(refreshContexts)
</script>

<style scoped>
.context-memory-panel {
  @apply flex flex-col h-full;
  background: var(--bg-1);
}

.contexts-list {
  @apply flex-1 overflow-y-auto;
}

.empty-state {
  @apply flex flex-col items-center justify-center py-8 text-center;
}

.context-card {
  @apply p-3 rounded-lg cursor-pointer transition-colors;
  background: var(--bg-2);
  border: 1px solid var(--border-default);
}

.context-card:hover {
  background: var(--bg-3);
  border-color: var(--border-hover);
}

.context-topic {
  @apply text-sm font-medium truncate;
  color: var(--text-primary);
}

.context-summary {
  @apply text-xs mt-1 line-clamp-2;
  color: var(--text-secondary);
}

.context-meta {
  @apply flex items-center justify-between mt-2 text-xs;
  color: var(--text-tertiary);
}

.context-files {
  @apply flex items-center gap-1;
}

.save-section {
  @apply p-3 border-t;
  border-color: var(--border-default);
}

.modal-overlay {
  @apply fixed inset-0 flex items-center justify-center z-50;
  background: rgba(0, 0, 0, 0.7);
}

.modal-content {
  @apply p-6 rounded-lg w-full max-w-md;
  background: var(--bg-1);
  border: 1px solid var(--border-default);
}

.modal-title {
  @apply text-lg font-semibold mb-4;
  color: var(--text-primary);
}

.modal-actions {
  @apply flex justify-end gap-2 mt-6;
}
</style>
