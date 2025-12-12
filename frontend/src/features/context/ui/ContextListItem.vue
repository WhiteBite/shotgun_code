<template>
  <div
    class="bg-gray-800/50 rounded-xl p-3 border transition-all cursor-pointer"
    :class="[
      isSelected ? 'border-purple-500/50 bg-purple-500/10' : 'border-gray-700/30 hover:bg-gray-800/70 hover:border-gray-600/50',
      context.isFavorite ? 'ring-1 ring-yellow-500/20' : '',
      dragOverIndex === index ? 'border-indigo-500 border-dashed' : ''
    ]"
    :draggable="!searchQuery && !showFavoritesOnly"
    @click="$emit('toggle-select', context.id)"
    @dblclick="$emit('load', context.id)"
    @dragstart="$emit('drag-start', $event, index)"
    @dragover="$emit('drag-over', $event, index)"
    @dragleave="$emit('drag-leave')"
    @drop="$emit('drop', $event, index)"
    @dragend="$emit('drag-end')"
  >
    <!-- Header Row -->
    <div class="flex items-start justify-between mb-2">
      <div class="flex items-center gap-2 flex-1 min-w-0">
        <!-- Checkbox -->
        <input
          type="checkbox"
          :checked="isSelected"
          @click.stop="$emit('toggle-select', context.id)"
          class="rounded border-gray-600 bg-gray-700 text-purple-500 focus:ring-purple-500 w-3.5 h-3.5 flex-shrink-0"
        />
        
        <!-- Favorite star -->
        <button
          @click.stop="$emit('toggle-favorite', context.id)"
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
            v-if="isEditing"
            :value="editingName"
            @input="$emit('update:editing-name', ($event.target as HTMLInputElement).value)"
            @blur="$emit('save-rename', context.id)"
            @keyup.enter="$emit('save-rename', context.id)"
            @keyup.escape="$emit('cancel-rename')"
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
        <button @click.stop="$emit('copy', context.id)" class="p-1.5 rounded-lg text-gray-400 hover:text-indigo-400 hover:bg-indigo-500/10 transition-colors" :title="t('context.copyToClipboard')">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </button>
        <button @click.stop="$emit('start-rename', context)" class="p-1.5 rounded-lg text-gray-400 hover:text-emerald-400 hover:bg-emerald-500/10 transition-colors" :title="t('context.rename')">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
        </button>
        <button @click.stop="$emit('duplicate', context.id)" class="p-1.5 rounded-lg text-gray-400 hover:text-purple-400 hover:bg-purple-500/10 transition-colors" :title="t('context.duplicate')">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2" />
          </svg>
        </button>
        <button @click.stop="$emit('export', context.id)" class="p-1.5 rounded-lg text-gray-400 hover:text-indigo-400 hover:bg-indigo-500/10 transition-colors" :title="t('context.export')">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
        </button>
        <button @click.stop="$emit('delete', context)" class="p-1.5 rounded-lg text-gray-400 hover:text-red-400 hover:bg-red-500/10 transition-colors" :title="t('context.delete')">
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
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';
import { formatContextSize, formatTimestamp } from '../lib/context-utils';
import type { ContextSummary } from '../model/context.store';

const { t } = useI18n()

defineProps<{
  context: ContextSummary
  index: number
  isSelected: boolean
  isEditing: boolean
  editingName: string
  dragOverIndex: number | null
  searchQuery: string
  showFavoritesOnly: boolean
}>()

defineEmits<{
  (e: 'toggle-select', id: string): void
  (e: 'toggle-favorite', id: string): void
  (e: 'load', id: string): void
  (e: 'start-rename', context: ContextSummary): void
  (e: 'save-rename', id: string): void
  (e: 'cancel-rename'): void
  (e: 'update:editing-name', value: string): void
  (e: 'copy', id: string): void
  (e: 'duplicate', id: string): void
  (e: 'export', id: string): void
  (e: 'delete', context: ContextSummary): void
  (e: 'drag-start', event: DragEvent, index: number): void
  (e: 'drag-over', event: DragEvent, index: number): void
  (e: 'drag-leave'): void
  (e: 'drop', event: DragEvent, index: number): void
  (e: 'drag-end'): void
}>()

function formatSize(bytes: number): string {
  return formatContextSize(bytes)
}
</script>
