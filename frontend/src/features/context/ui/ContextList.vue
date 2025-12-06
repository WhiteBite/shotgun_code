<template>
  <div class="h-full flex flex-col bg-transparent">
    <!-- Header -->
    <div class="flex items-center justify-between p-3 border-b border-gray-700/30">
      <div class="flex items-center gap-2">
        <div class="panel-icon bg-purple-500/20">
          <svg class="w-4 h-4 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
          </svg>
        </div>
        <h3 class="text-sm font-semibold text-white">{{ t('context.saved') }}</h3>
      </div>
      <div class="flex items-center gap-1">
        <button
          @click="showSettings = !showSettings"
          class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-gray-700/50 transition-colors"
          :class="{ 'text-purple-400 bg-purple-500/10': showSettings }"
          :title="t('context.settings')"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>
        <button
          @click="refresh"
          class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-gray-700/50 transition-colors"
          :title="t('files.refresh')"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Settings Panel -->
    <div v-if="showSettings" class="p-3 border-b border-gray-700/30 bg-gray-800/30 space-y-3">
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs text-gray-400 mb-1 block">{{ t('context.maxContexts') }}</label>
          <input type="number" v-model.number="storageSettings.maxContexts" min="5" max="100" class="input text-sm" />
        </div>
        <div>
          <label class="text-xs text-gray-400 mb-1 block">{{ t('context.autoCleanupDays') }}</label>
          <input type="number" v-model.number="storageSettings.autoCleanupDays" min="0" max="365" class="input text-sm" />
        </div>
      </div>
      <label class="flex items-center gap-2 text-xs text-gray-300 cursor-pointer">
        <input type="checkbox" v-model="storageSettings.autoCleanupOnLimit" class="rounded border-gray-600 bg-gray-700 text-purple-500 focus:ring-purple-500" />
        {{ t('context.autoCleanupOnLimit') }}
      </label>
    </div>

    <!-- Toolbar -->
    <div v-if="contextStore.contextList.length > 0" class="p-2 border-b border-gray-700/30 space-y-2">
      <!-- Search -->
      <div class="relative">
        <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="t('context.search')"
          class="input pl-8 text-sm"
        />
      </div>
      
      <!-- Sort & Filter -->
      <div class="flex items-center gap-2 flex-wrap">
        <select v-model="sortBy" class="input text-xs py-1 px-2 w-auto">
          <option value="date">{{ t('context.sortByDate') }}</option>
          <option value="name">{{ t('context.sortByName') }}</option>
          <option value="size">{{ t('context.sortBySize') }}</option>
        </select>
        
        <label class="flex items-center gap-1.5 text-xs text-gray-400 cursor-pointer">
          <input type="checkbox" v-model="showFavoritesOnly" class="rounded border-gray-600 bg-gray-700 text-purple-500 focus:ring-purple-500 w-3 h-3" />
          <svg class="w-3 h-3 text-yellow-400" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
          </svg>
        </label>
        
        <div class="flex-1"></div>
        
        <!-- Multi-select actions -->
        <template v-if="selectedContexts.size > 0">
          <span class="text-xs text-gray-400">{{ selectedContexts.size }} {{ t('files.selected') }}</span>
          <button 
            @click="copySelectedContext" 
            class="text-xs text-indigo-400 hover:text-indigo-300"
            title="Ctrl+C"
          >
            {{ t('context.copyToClipboard') }}
          </button>
          <button 
            v-if="selectedContexts.size >= 2"
            @click="mergeSelected" 
            class="text-xs text-purple-400 hover:text-purple-300"
            title="Ctrl+M"
          >
            {{ t('context.mergeSelected') }}
          </button>
          <button @click="deleteSelected" class="text-xs text-red-400 hover:text-red-300" title="Delete">
            {{ t('context.deleteSelected') }}
          </button>
        </template>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="contextStore.isLoading" class="flex items-center justify-center flex-1">
      <svg class="animate-spin h-8 w-8 text-indigo-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
      </svg>
    </div>

    <!-- Empty State -->
    <div v-else-if="contextStore.contextList.length === 0" class="flex items-center justify-center flex-1 p-4">
      <div class="text-center max-w-xs">
        <div class="w-16 h-16 mx-auto mb-4 bg-gray-800/50 rounded-2xl flex items-center justify-center border border-gray-700/30">
          <svg class="w-8 h-8 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
          </svg>
        </div>
        <p class="text-sm font-medium text-gray-400 mb-1">{{ t('context.noSaved') }}</p>
        <p class="text-xs text-gray-500">{{ t('context.buildToSee') }}</p>
      </div>
    </div>
    <!-- Context List -->
    <div v-else class="flex-1 overflow-auto p-3 space-y-2">
      <div
        v-for="(context, index) in filteredContexts"
        :key="context.id"
        class="bg-gray-800/50 rounded-xl p-3 border transition-all cursor-pointer"
        :class="[
          selectedContexts.has(context.id) ? 'border-purple-500/50 bg-purple-500/10' : 'border-gray-700/30 hover:bg-gray-800/70 hover:border-gray-600/50',
          context.isFavorite ? 'ring-1 ring-yellow-500/20' : '',
          dragOverIndex === index ? 'border-indigo-500 border-dashed' : ''
        ]"
        :draggable="!searchQuery && !showFavoritesOnly"
        @click="handleClick(context)"
        @dblclick="loadContext(context.id)"
        @dragstart="handleDragStart($event, index)"
        @dragover="handleDragOver($event, index)"
        @dragleave="handleDragLeave"
        @drop="handleDrop($event, index)"
        @dragend="handleDragEnd"
      >
        <!-- Header Row -->
        <div class="flex items-start justify-between mb-2">
          <div class="flex items-center gap-2 flex-1 min-w-0">
            <!-- Checkbox for multi-select -->
            <input
              type="checkbox"
              :checked="selectedContexts.has(context.id)"
              @click.stop="toggleSelect(context.id)"
              class="rounded border-gray-600 bg-gray-700 text-purple-500 focus:ring-purple-500 w-3.5 h-3.5 flex-shrink-0"
            />
            
            <!-- Favorite star -->
            <button
              @click.stop="toggleFavorite(context.id)"
              class="flex-shrink-0 p-0.5"
              :title="context.isFavorite ? t('context.unfavorite') : t('context.favorite')"
            >
              <svg class="w-4 h-4" :class="context.isFavorite ? 'text-yellow-400' : 'text-gray-600 hover:text-yellow-400'" :fill="context.isFavorite ? 'currentColor' : 'none'" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
              </svg>
            </button>
            
            <!-- Name (editable) -->
            <div class="flex-1 min-w-0">
              <input
                v-if="editingId === context.id"
                v-model="editingName"
                @blur="saveRename(context.id)"
                @keyup.enter="saveRename(context.id)"
                @keyup.escape="cancelRename"
                @click.stop
                class="input text-sm py-0.5 px-1 w-full"
                ref="renameInput"
              />
              <p v-else class="text-sm text-white font-medium truncate" :title="context.name">
                {{ context.name || context.id }}
              </p>
              <p class="text-xs text-gray-500">{{ formatTimestamp(context.createdAt || '') }}</p>
            </div>
          </div>
          
          <!-- Action buttons -->
          <div class="flex gap-0.5 flex-shrink-0 ml-2">
            <button
              @click.stop="copyContext(context.id)"
              class="p-1.5 rounded-lg text-gray-400 hover:text-indigo-400 hover:bg-indigo-500/10 transition-colors"
              :title="t('context.copyToClipboard')"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
            </button>
            <button
              @click.stop="startRename(context)"
              class="p-1.5 rounded-lg text-gray-400 hover:text-emerald-400 hover:bg-emerald-500/10 transition-colors"
              :title="t('context.rename')"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
            </button>
            <button
              @click.stop="duplicateContext(context.id)"
              class="p-1.5 rounded-lg text-gray-400 hover:text-purple-400 hover:bg-purple-500/10 transition-colors"
              :title="t('context.duplicate')"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2" />
              </svg>
            </button>
            <button
              @click.stop="exportContext(context.id)"
              class="p-1.5 rounded-lg text-gray-400 hover:text-indigo-400 hover:bg-indigo-500/10 transition-colors"
              :title="t('context.export')"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </button>
            <button
              @click.stop="confirmDelete(context)"
              class="p-1.5 rounded-lg text-gray-400 hover:text-red-400 hover:bg-red-500/10 transition-colors"
              :title="t('context.delete')"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>
        
        <!-- Stats -->
        <div class="flex items-center gap-3 text-xs">
          <span class="badge badge-primary">{{ context.fileCount }} {{ t('context.files') }}</span>
          <span class="badge badge-primary">{{ context.lineCount }} {{ t('context.lines') }}</span>
          <span class="badge badge-primary">{{ formatSize(context.totalSize) }}</span>
        </div>
        
        <!-- Double-click hint -->
        <p class="text-xs text-gray-600 mt-2 text-center">{{ t('context.doubleClickToView') }}</p>
      </div>
    </div>
    <!--
 Delete Confirmation Modal -->
    <Teleport to="body">
      <div v-if="deleteModal.show" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="deleteModal.show = false"></div>
        <div class="relative confirm-modal">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-12 h-12 rounded-xl bg-red-500/20 flex items-center justify-center">
              <svg class="w-6 h-6 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-white">{{ t('context.confirmDelete') }}</h3>
              <p class="text-sm text-gray-400">{{ deleteModal.message }}</p>
            </div>
          </div>
          
          <div class="flex justify-end gap-3">
            <button @click="deleteModal.show = false" class="btn btn-ghost">
              {{ t('context.cancel') }}
            </button>
            <button @click="executeDelete" class="btn bg-red-500 hover:bg-red-600 text-white">
              {{ t('context.delete') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useSettingsStore } from '@/stores/settings.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { formatContextSize, formatTimestamp } from '../lib/context-utils'
import { useContextStore, type ContextSummary } from '../model/context.store'

const { t } = useI18n()
const contextStore = useContextStore()
const settingsStore = useSettingsStore()
const uiStore = useUIStore()

// State
const showSettings = ref(false)
const searchQuery = ref('')
const sortBy = ref<'date' | 'name' | 'size'>('date')
const showFavoritesOnly = ref(false)
const selectedContexts = ref(new Set<string>())
const editingId = ref<string | null>(null)
const editingName = ref('')
const renameInput = ref<HTMLInputElement | null>(null)

// Drag & drop state
const dragIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

// Delete modal
const deleteModal = reactive({
  show: false,
  contextId: null as string | null,
  contextIds: [] as string[],
  message: ''
})

// Storage settings (reactive binding)
const storageSettings = computed({
  get: () => settingsStore.settings.contextStorage,
  set: (val) => settingsStore.updateContextStorageSettings(val)
})

// Watch for settings changes - auto-saved by store
watch(() => settingsStore.settings.contextStorage, () => {}, { deep: true })

// Filtered and sorted contexts
const filteredContexts = computed(() => {
  let list = [...contextStore.contextList]
  
  // Filter by search
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(c => 
      (c.name || c.id).toLowerCase().includes(q)
    )
  }
  
  // Filter favorites
  if (showFavoritesOnly.value) {
    list = list.filter(c => c.isFavorite)
  }
  
  // Sort
  list.sort((a, b) => {
    // Favorites always first
    if (a.isFavorite && !b.isFavorite) return -1
    if (!a.isFavorite && b.isFavorite) return 1
    
    switch (sortBy.value) {
      case 'name':
        return (a.name || a.id).localeCompare(b.name || b.id)
      case 'size':
        return b.totalSize - a.totalSize
      case 'date':
      default:
        return new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime()
    }
  })
  
  return list
})

function formatSize(bytes: number): string {
  return formatContextSize(bytes)
}

async function refresh() {
  try {
    await contextStore.listProjectContexts()
  } catch (error) {
    console.error('[ContextList] Failed to refresh:', error)
  }
}

const emit = defineEmits<{
  (e: 'switch-to-preview'): void
}>()

// Single click - toggle selection
function handleClick(_context: ContextSummary) {
  // Don't toggle if editing
  if (editingId.value) return
}

// Double click - load context
async function loadContext(contextId: string) {
  try {
    await contextStore.loadContextContent(contextId, 0, 100)
    emit('switch-to-preview')
  } catch (error) {
    console.error('[ContextList] Failed to load context:', error)
  }
}

function toggleSelect(contextId: string) {
  if (selectedContexts.value.has(contextId)) {
    selectedContexts.value.delete(contextId)
  } else {
    selectedContexts.value.add(contextId)
  }
  selectedContexts.value = new Set(selectedContexts.value) // Trigger reactivity
}

// Favorite
function toggleFavorite(contextId: string) {
  contextStore.toggleFavorite(contextId)
}

// Rename
function startRename(context: ContextSummary) {
  editingId.value = context.id
  editingName.value = context.name || context.id
  nextTick(() => {
    renameInput.value?.focus()
    renameInput.value?.select()
  })
}

function saveRename(contextId: string) {
  if (editingName.value.trim()) {
    contextStore.renameContext(contextId, editingName.value.trim())
    uiStore.addToast(t('context.renamed'), 'success')
  }
  cancelRename()
}

function cancelRename() {
  editingId.value = null
  editingName.value = ''
}

// Copy
async function copyContext(contextId: string) {
  try {
    // Load this specific context first if needed
    if (contextStore.contextId !== contextId) {
      await contextStore.loadContextContent(contextId, 0, 100)
    }
    const fullContent = await contextStore.getFullContextContent()
    await navigator.clipboard.writeText(fullContent)
    uiStore.addToast(t('toast.contextCopied'), 'success')
  } catch (error) {
    console.error('[ContextList] Failed to copy:', error)
    uiStore.addToast(t('toast.copyError'), 'error')
  }
}

// Duplicate
async function duplicateContext(contextId: string) {
  try {
    const newId = await contextStore.duplicateContext(contextId)
    if (newId) {
      uiStore.addToast(t('context.duplicated'), 'success')
    }
  } catch (error) {
    console.error('[ContextList] Failed to duplicate:', error)
  }
}

// Export
async function exportContext(contextId: string) {
  try {
    await contextStore.exportContext(contextId)
  } catch (error) {
    console.error('[ContextList] Export failed:', error)
  }
}

// Delete
function confirmDelete(context: ContextSummary) {
  deleteModal.contextId = context.id
  deleteModal.contextIds = []
  deleteModal.message = t('context.confirmDeleteMessage').replace('{name}', context.name || context.id)
  deleteModal.show = true
}

function deleteSelected() {
  if (selectedContexts.value.size === 0) return
  
  deleteModal.contextId = null
  deleteModal.contextIds = [...selectedContexts.value]
  deleteModal.message = t('context.confirmDeleteMultiple').replace('{count}', String(selectedContexts.value.size))
  deleteModal.show = true
}

async function executeDelete() {
  try {
    if (deleteModal.contextId) {
      await contextStore.deleteContext(deleteModal.contextId)
      uiStore.addToast(t('context.deleted'), 'success')
    } else if (deleteModal.contextIds.length > 0) {
      for (const id of deleteModal.contextIds) {
        await contextStore.deleteContext(id)
      }
      selectedContexts.value.clear()
      uiStore.addToast(t('context.deleted'), 'success')
    }
  } catch (error) {
    console.error('[ContextList] Delete failed:', error)
  } finally {
    deleteModal.show = false
    deleteModal.contextId = null
    deleteModal.contextIds = []
  }
}

// Merge selected contexts
async function mergeSelected() {
  if (selectedContexts.value.size < 2) {
    uiStore.addToast(t('context.selectToMerge'), 'warning')
    return
  }
  
  try {
    const ids = [...selectedContexts.value]
    const newId = await contextStore.mergeContexts(ids)
    if (newId) {
      selectedContexts.value.clear()
      selectedContexts.value = new Set(selectedContexts.value)
      uiStore.addToast(t('context.merged'), 'success')
    }
  } catch (error) {
    console.error('[ContextList] Failed to merge:', error)
  }
}

// Drag & drop handlers
function handleDragStart(e: DragEvent, index: number) {
  dragIndex.value = index
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(index))
  }
}

function handleDragOver(e: DragEvent, index: number) {
  e.preventDefault()
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = 'move'
  }
  dragOverIndex.value = index
}

function handleDragLeave() {
  dragOverIndex.value = null
}

function handleDrop(e: DragEvent, toIndex: number) {
  e.preventDefault()
  if (dragIndex.value !== null && dragIndex.value !== toIndex) {
    contextStore.reorderContexts(dragIndex.value, toIndex)
  }
  dragIndex.value = null
  dragOverIndex.value = null
}

function handleDragEnd() {
  dragIndex.value = null
  dragOverIndex.value = null
}

// Copy selected context with Ctrl+C
async function copySelectedContext() {
  if (selectedContexts.value.size === 1) {
    const contextId = [...selectedContexts.value][0]
    await copyContext(contextId)
  } else if (selectedContexts.value.size > 1) {
    // Copy multiple - merge and copy
    try {
      const contents: string[] = []
      for (const ctxId of selectedContexts.value) {
        if (contextStore.contextId !== ctxId) {
          await contextStore.loadContextContent(ctxId, 0, 100)
        }
        const content = await contextStore.getFullContextContent()
        contents.push(content)
      }
      const merged = contents.join('\n\n' + '='.repeat(80) + '\n\n')
      await navigator.clipboard.writeText(merged)
      uiStore.addToast(t('toast.contextCopied'), 'success')
    } catch (error) {
      console.error('[ContextList] Failed to copy multiple:', error)
      uiStore.addToast(t('toast.copyError'), 'error')
    }
  }
}

// Keyboard shortcuts
function handleKeydown(e: KeyboardEvent) {
  // Delete selected contexts
  if (e.key === 'Delete' && selectedContexts.value.size > 0) {
    e.preventDefault()
    deleteSelected()
  }
  // Ctrl+A - select all
  if (e.key === 'a' && (e.ctrlKey || e.metaKey) && filteredContexts.value.length > 0) {
    e.preventDefault()
    filteredContexts.value.forEach(c => selectedContexts.value.add(c.id))
    selectedContexts.value = new Set(selectedContexts.value)
  }
  // Ctrl+C - copy selected
  if (e.key === 'c' && (e.ctrlKey || e.metaKey) && selectedContexts.value.size > 0) {
    e.preventDefault()
    copySelectedContext()
  }
  // Ctrl+M - merge selected
  if (e.key === 'm' && (e.ctrlKey || e.metaKey) && selectedContexts.value.size >= 2) {
    e.preventDefault()
    mergeSelected()
  }
  // Escape - clear selection
  if (e.key === 'Escape') {
    selectedContexts.value.clear()
    selectedContexts.value = new Set(selectedContexts.value)
    editingId.value = null
  }
}

onMounted(() => {
  refresh()
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>