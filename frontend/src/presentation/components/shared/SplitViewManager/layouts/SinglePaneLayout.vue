<template>
  <div class="single-pane-layout">
    <div class="content-viewer">
      <ContextViewer
        :content="context.content"
        :chunks="chunks"
        :active-chunk="activeChunk"
        :highlight="true"
        :line-numbers="true"
        :virtual-scroll="true"
        :show-chunk-boundaries="chunks.length > 1"
        @chunk-hover="$emit('chunk-hover', $event)"
        @chunk-select="$emit('chunk-select', $event)"
        @copy-selection="$emit('copy-selection', $event)"
        class="full-height"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
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

interface Props {
  context: ContextData
  chunks: ContextChunk[]
  activeChunk: ContextChunk | null
  activeChunkIndex: number
  splitSettings: any
}

defineProps<Props>()

defineEmits<{
  'chunk-select': [index: number]
  'chunk-hover': [index: number | null]
  'copy-chunk': [chunk: ContextChunk]
  'copy-selection': [selection: string]
}>()
</script>

<style scoped>
.single-pane-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
}

.content-viewer {
  flex: 1;
  min-height: 0;
  position: relative;
}

.full-height {
  height: 100%;
}
</style>