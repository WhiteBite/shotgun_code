<template>
  <div class="h-full flex flex-col border-r border-gray-700 bg-gray-900">
    <FileExplorer @preview-file="handlePreviewFile" />
  </div>
</template>

<script setup lang="ts">
import FileExplorer from '../FileExplorer.vue'

const emit = defineEmits<{
  (e: 'preview-file', filePath: string): void
}>()

function handlePreviewFile(filePath: string) {
  emit('preview-file', filePath)
}
</script>