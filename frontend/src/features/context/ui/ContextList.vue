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
        <button @click="list.showSettings.value = !list.showSettings.value"
          class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-gray-700/50 transition-colors"
          :class="{ 'text-purple-400 bg-purple-500/10': list.showSettings.value }"
          :title="t('context.settings')">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>
        <button @click="list.refresh" class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-gray-700/50 transition-colors" :title="t('files.refresh')">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Settings Panel -->
    <div v-if="list.showSettings.value" class="p-3 border-b border-gray-700/30 bg-gray-800/30 space-y-3">
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs text-gray-400 mb-1 block">{{ t('context.maxContexts') }}</label>
          <input type="number" v-model.number="list.storageSettings.value.maxContexts" min="5" max="100" class="input text-sm" />
        </div>
        <div>
          <label class="text-xs text-gray-400 mb-1 block">{{ t('context.autoCleanupDays') }}</label>
          <input type="number" v-model.number="list.storageSettings.value.autoCleanupDays" min="0" max="365" class="input text-sm" />
        </div>
      </div>
      <label class="flex items-center gap-2 text-xs text-gray-300 cursor-pointer">
        <input type="checkbox" v-model="list.storageSettings.value.autoCleanupOnLimit" class="rounded border-gray-600 bg-gray-700 text-purple-500 focus:ring-purple-500" />
        {{ t('context.autoCleanupOnLimit') }}
      </label>
    </div>

    <!-- Toolbar -->
    <div v-if="contextStore.contextList.length > 0" class="p-2 border-b border-gray-700/30 space-y-2">
      <div class="relative">
        <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input v-model="list.searchQuery.value" type="text" :placeholder="t('context.search')" class="input pl-8 text-sm" />
      </div>
      
      <div class="flex items-center gap-2 flex-wrap">
        <select v-model="list.sortBy.value" class="input text-xs py-1 px-2 w-auto">
          <option value="date">{{ t('context.sortByDate') }}</option>
          <option value="name">{{ t('context.sortByName') }}</option>
          <option value="size">{{ t('context.sortBySize') }}</option>
        </select>
        
        <label class="flex items-center gap-1.5 text-xs text-gray-400 cursor-pointer">
          <input type="checkbox" v-model="list.showFavoritesOnly.value" class="rounded border-gray-600 bg-gray-700 text-purple-500 focus:ring-purple-500 w-3 h-3" />
          <svg class="w-3 h-3 text-yellow-400" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
          </svg>
        </label>
        
        <div class="flex-1"></div>
        
        <template v-if="list.selectedContexts.value.size > 0">
          <span class="text-xs text-gray-400">{{ list.selectedContexts.value.size }} {{ t('files.selected') }}</span>
          <button @click="list.copySelectedContext" class="text-xs text-indigo-400 hover:text-indigo-300" :title="t('context.shortcutCopy')">
            {{ t('context.copyToClipboard') }}
          </button>
          <button v-if="list.selectedContexts.value.size >= 2" @click="list.mergeSelected" class="text-xs text-purple-400 hover:text-purple-300" :title="t('context.shortcutMerge')">
            {{ t('context.mergeSelected') }}
          </button>
          <button @click="list.deleteSelected" class="text-xs text-red-400 hover:text-red-300" :title="t('context.deleteAction')">
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
      <ContextListItem
        v-for="(context, index) in list.filteredContexts.value"
        :key="context.id"
        :context="context"
        :index="index"
        :is-selected="list.selectedContexts.value.has(context.id)"
        :is-editing="list.editingId.value === context.id"
        :editing-name="list.editingName.value"
        :drag-over-index="list.dragOverIndex.value"
        :search-query="list.searchQuery.value"
        :show-favorites-only="list.showFavoritesOnly.value"
        @toggle-select="list.toggleSelect"
        @toggle-favorite="list.toggleFavorite"
        @load="loadContext"
        @start-rename="handleStartRename"
        @save-rename="list.saveRename"
        @cancel-rename="list.cancelRename"
        @update:editing-name="list.editingName.value = $event"
        @copy="list.copyContext"
        @duplicate="list.duplicateContext"
        @export="list.exportContext"
        @delete="list.confirmDelete"
        @drag-start="list.handleDragStart"
        @drag-over="list.handleDragOver"
        @drag-leave="list.handleDragLeave"
        @drop="list.handleDrop"
        @drag-end="list.handleDragEnd"
      />
    </div>

    <!-- Delete Confirmation Modal -->
    <Teleport to="body">
      <div v-if="list.deleteModal.show" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="list.deleteModal.show = false"></div>
        <div class="relative confirm-modal">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-12 h-12 rounded-xl bg-red-500/20 flex items-center justify-center">
              <svg class="w-6 h-6 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-white">{{ t('context.confirmDelete') }}</h3>
              <p class="text-sm text-gray-400">{{ list.deleteModal.message }}</p>
            </div>
          </div>
          
          <div class="flex justify-end gap-3">
            <button @click="list.deleteModal.show = false" class="btn btn-ghost">{{ t('context.cancel') }}</button>
            <button @click="list.executeDelete" class="btn bg-red-500 hover:bg-red-600 text-white">{{ t('context.delete') }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { onMounted, onUnmounted } from 'vue'
import { useContextList } from '../composables/useContextList'
import { useContextStore, type ContextSummary } from '../model/context.store'
import ContextListItem from './ContextListItem.vue'

const { t } = useI18n()
const contextStore = useContextStore()
const list = useContextList()

const emit = defineEmits<{
  (e: 'switch-to-preview'): void
}>()

async function loadContext(contextId: string) {
  try {
    await contextStore.loadContextContent(contextId, 0, 100)
    emit('switch-to-preview')
  } catch (error) {
    console.error('[ContextList] Failed to load context:', error)
  }
}

function handleStartRename(context: ContextSummary) {
  list.startRename(context)
}

onMounted(() => {
  list.refresh()
  window.addEventListener('keydown', list.handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', list.handleKeydown)
})
</script>
