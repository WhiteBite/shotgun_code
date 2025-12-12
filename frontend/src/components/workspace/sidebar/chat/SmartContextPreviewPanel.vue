<template>
  <div class="mx-3 mb-2 p-3 bg-indigo-500/10 border border-indigo-500/30 rounded-xl">
    <div class="flex items-center justify-between mb-2">
      <span class="text-xs font-medium text-indigo-300">
        {{ t('chat.smartContext.title') }}
      </span>
      <span class="text-[10px] text-gray-400">
        {{ preview.files.length }} {{ t('context.files') }}
      </span>
    </div>
    
    <div class="max-h-24 overflow-y-auto mb-2">
      <div 
        v-for="file in preview.files.slice(0, 5)" 
        :key="file"
        class="text-[10px] text-gray-400 truncate"
      >
        {{ file }}
      </div>
      <div v-if="preview.files.length > 5" class="text-[10px] text-gray-500">
        +{{ preview.files.length - 5 }} {{ t('common.more') }}
      </div>
    </div>
    
    <div class="flex gap-2">
      <button 
        class="btn btn-sm btn-primary flex-1"
        @click="$emit('confirm')"
      >
        {{ t('common.confirm') }}
      </button>
      <button 
        class="btn btn-sm btn-ghost"
        @click="$emit('cancel')"
      >
        {{ t('common.cancel') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';

interface SmartContextPreview {
  files: string[]
  totalTokens?: number
}

defineProps<{
  preview: SmartContextPreview
}>()

defineEmits<{
  confirm: []
  cancel: []
}>()

const { t } = useI18n()
</script>
