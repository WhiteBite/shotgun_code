<template>
  <div class="p-3 border-t border-gray-700/30">
    <!-- Context Toggle (only in manual mode) -->
    <div v-if="chatMode === 'manual'" class="flex items-center justify-between mb-2">
      <!-- When context exists - show interactive checkbox -->
      <label 
        v-if="contextStore.hasContext"
        class="flex items-center gap-2 text-xs cursor-pointer text-gray-400 hover:text-gray-300"
      >
        <input
          type="checkbox"
          :checked="includeContext"
          @change="$emit('update:include-context', ($event.target as HTMLInputElement).checked)"
          :aria-label="t('chat.useContext')"
          class="w-3.5 h-3.5 rounded border-gray-600 bg-gray-800 text-indigo-500 focus:ring-indigo-500 focus:ring-offset-0"
        />
        <span>{{ t('chat.useContext') }}</span>
      </label>
      <!-- When no context - show informational hint -->
      <div 
        v-else
        class="flex items-center gap-1.5 text-xs text-gray-500"
        :title="t('chat.buildContextFirst')"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{{ t('chat.noContextHint') }}</span>
      </div>
      <div class="flex items-center gap-2">
        <span v-if="contextStore.hasContext" class="text-[10px] text-indigo-400">
          {{ contextStore.fileCount }} {{ t('context.files') }}
        </span>
        <button
          v-if="hasMessages"
          @click="$emit('clear')"
          class="text-[10px] text-gray-500 hover:text-gray-300"
        >
          {{ t('chat.clear') }}
        </button>
      </div>
    </div>

    <!-- Smart/Agentic mode hint -->
    <div v-else class="flex items-center justify-between mb-2">
      <span class="text-[10px]" :class="chatMode === 'agentic' ? 'text-emerald-400/70' : 'text-purple-400/70'">
        {{ chatMode === 'agentic' ? t('chat.modeAgenticHint') : t('chat.modeSmartHint') }}
      </span>
      <button
        v-if="hasMessages"
        @click="$emit('clear')"
        class="text-[10px] text-gray-500 hover:text-gray-300"
      >
        {{ t('chat.clear') }}
      </button>
    </div>

    <!-- Input -->
    <div class="flex gap-2">
      <textarea
        :value="modelValue"
        @input="handleInput"
        @keydown.enter.exact="handleEnter"
        @keydown.shift.enter.prevent="handleShiftEnter"
        class="input-chat text-xs flex-1"
        :placeholder="t('chat.placeholder')"
        :disabled="isThinking || isAnalyzing"
        rows="1"
        style="min-height: 36px; max-height: 120px;"
        ref="inputRef"
      ></textarea>
      <button
        @click="$emit('send')"
        :disabled="!modelValue.trim() || isThinking || isAnalyzing"
        class="btn-cta !px-3 !py-2 self-end"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';
import { useContextStore } from '@/features/context';

const props = defineProps<{
  modelValue: string
  chatMode: 'manual' | 'smart' | 'agentic'
  isThinking: boolean
  isAnalyzing: boolean
  includeContext: boolean
  hasMessages: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:include-context': [value: boolean]
  send: []
  clear: []
}>()

const { t } = useI18n()
const contextStore = useContextStore()


function handleInput(event: Event) {
  const textarea = event.target as HTMLTextAreaElement
  emit('update:modelValue', textarea.value)
  autoResize(textarea)
}

function handleEnter(event: KeyboardEvent) {
  event.preventDefault()
  emit('send')
}

function handleShiftEnter() {
  emit('update:modelValue', props.modelValue + '\n')
}

function autoResize(textarea: HTMLTextAreaElement) {
  textarea.style.height = '36px'
  textarea.style.height = Math.min(textarea.scrollHeight, 120) + 'px'
}
</script>
