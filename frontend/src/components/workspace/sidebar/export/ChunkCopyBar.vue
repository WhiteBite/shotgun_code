<template>
  <div v-if="contextStore.hasContext && needsSplitting" class="chunk-copy-bar">
    <!-- Header with warning -->
    <div class="chunk-copy-header">
      <AlertTriangle class="w-4 h-4" />
      <span>{{ formatTokens(totalTokens) }} → {{ chunkCount }} {{ t('chunks.parts') }}</span>
    </div>

    <!-- Main copy button -->
    <button 
      v-if="!allCopied" 
      class="chunk-copy-main-btn"
      :disabled="isProcessing"
      @click="handleCopyNext"
    >
      <Copy class="w-4 h-4" />
      <span>{{ t('chunks.copyPart') }} {{ nextUncopiedIndex + 1 }}/{{ chunkCount }}</span>
    </button>
    <div v-else class="chunk-copy-done">
      <CheckCircle class="w-4 h-4" />
      <span>{{ t('chunks.allCopied') }}</span>
    </div>

    <!-- Progress bar -->
    <div class="chunk-copy-progress">
      <div class="chunk-copy-progress-bar" :style="{ width: `${progress}%` }" />
    </div>

    <!-- Chunk buttons grid -->
    <div class="chunk-copy-grid">
      <button
        v-for="chunk in chunks"
        :key="chunk.index"
        class="chunk-btn"
        :class="{
          'chunk-btn-copied': chunk.copied,
          'chunk-btn-next': !chunk.copied && chunk.index === nextUncopiedIndex,
          'chunk-btn-pending': !chunk.copied && chunk.index !== nextUncopiedIndex
        }"
        :title="getChunkTooltip(chunk)"
        @click="handleCopyChunk(chunk.index)"
      >
        <Check v-if="chunk.copied" class="chunk-btn-icon" />
        <span v-else class="chunk-btn-index">{{ chunk.index + 1 }}</span>
        <span class="chunk-btn-tokens">{{ formatTokens(chunk.tokens) }}</span>
      </button>
    </div>

    <!-- Footer with reset -->
    <div v-if="progress > 0" class="chunk-copy-footer">
      <span class="chunk-copy-status">
        {{ copiedCount }}/{{ chunkCount }} {{ t('chunks.copied') }}
      </span>
      <button class="chunk-copy-reset" @click="handleReset">
        <RotateCcw class="w-3.5 h-3.5" />
        <span>{{ t('chunks.reset') }}</span>
      </button>
    </div>
  </div>

  <!-- Generate button when no chunks yet -->
  <button
    v-else-if="contextStore.hasContext && enableAutoSplit"
    class="chunk-generate-btn"
    :disabled="isProcessing"
    @click="generateChunks"
  >
    <Layers class="w-4 h-4" />
    <span>{{ t('chunks.generate') }}</span>
  </button>
</template>

<script setup lang="ts">
import { useChunkedCopy, type Chunk } from '@/composables/useChunkedCopy'
import { useI18n } from '@/composables/useI18n'
import { useContextStore } from '@/features/context/model/context.store'
import { AlertTriangle, Check, CheckCircle, Copy, Layers, RotateCcw } from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'

const { t } = useI18n()
const contextStore = useContextStore()
const showCopiedFeedback = ref(false)

const {
  chunks,
  isProcessing,
  totalTokens,
  needsSplitting,
  chunkCount,
  allCopied,
  progress,
  enableAutoSplit,
  generateChunks,
  copyChunk,
  copyNext,
  resetCopied
} = useChunkedCopy()

const nextUncopiedIndex = computed(() => {
  const idx = chunks.value.findIndex(c => !c.copied)
  return idx === -1 ? 0 : idx
})

const copiedCount = computed(() => chunks.value.filter(c => c.copied).length)

function formatTokens(tokens: number): string {
  if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}K`
  return tokens.toString()
}

function getChunkTooltip(chunk: Chunk): string {
  const status = chunk.copied ? `✓ ${t('chunks.copied')}` : t('chunks.clickToCopy')
  return `${t('chunks.part')} ${chunk.index + 1}: ${formatTokens(chunk.tokens)} ${t('context.tokens')}\n${chunk.files} ${t('context.files')}\n${status}`
}

async function handleCopyChunk(index: number) {
  const success = await copyChunk(index)
  if (success) {
    showFeedback()
  }
}

async function handleCopyNext() {
  const success = await copyNext()
  if (success) {
    showFeedback()
  }
}

function handleReset() {
  resetCopied()
}

function showFeedback() {
  showCopiedFeedback.value = true
  setTimeout(() => {
    showCopiedFeedback.value = false
  }, 1500)
}

// Auto-generate chunks when context changes and auto-split is enabled
watch(
  () => [contextStore.contextId, enableAutoSplit.value],
  async ([ctxId, autoSplit]) => {
    if (ctxId && autoSplit && needsSplitting.value) {
      await generateChunks()
    }
  },
  { immediate: true }
)
</script>
