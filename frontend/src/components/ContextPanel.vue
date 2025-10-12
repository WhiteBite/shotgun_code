<template>
  <div class="h-full flex flex-col bg-gray-900">
    <div class="flex items-center justify-between p-4 border-b border-gray-700">
      <h2 class="text-lg font-semibold text-white">Context Preview</h2>
      
      <div v-if="contextStore.hasContext" class="flex items-center gap-4 text-sm text-gray-400">
        <span>{{ contextStore.fileCount }} files</span>
        <span>{{ contextStore.totalLines }} lines</span>
        <span>{{ formatSize(contextStore.totalSize) }}</span>
        
        <button
          @click="contextStore.clearContext"
          class="ml-2 px-3 py-1 text-xs bg-red-600 hover:bg-red-700 text-white rounded transition-colors"
        >
          Clear
        </button>
      </div>
    </div>

    <div class="flex-1 overflow-auto p-4">
      <div v-if="contextStore.isLoading" class="flex items-center justify-center h-full">
        <div class="text-center">
          <svg class="animate-spin h-8 w-8 text-blue-500 mx-auto mb-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          <p class="text-gray-400">Building context...</p>
        </div>
      </div>

      <div v-else-if="contextStore.error" class="flex items-center justify-center h-full">
        <div class="text-center text-red-400">
          <svg class="h-12 w-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <p>{{ contextStore.error }}</p>
        </div>
      </div>

      <div v-else-if="!contextStore.hasContext" class="flex items-center justify-center h-full">
        <div class="text-center text-gray-500">
          <svg class="h-16 w-16 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
          </svg>
          <p class="text-lg mb-2">No context built yet</p>
          <p class="text-sm">Select files and click "Build Context" to see the preview.</p>
        </div>
      </div>

      <div v-else class="bg-gray-800 rounded-lg p-4 font-mono text-sm text-gray-300">
        <div v-if="contextStore.currentChunk">
          <div v-for="(line, index) in contextStore.currentChunk.lines" :key="index" class="hover:bg-gray-700 px-2 py-0.5 rounded">
            <span class="text-gray-500 mr-4">{{ contextStore.currentChunk.startLine + index }}</span>
            <span>{{ line }}</span>
          </div>
        </div>
        <div v-else class="text-center text-gray-500 py-8">
          <p>Context built successfully</p>
          <p class="text-xs mt-2">Preview not available</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useContextStore } from '@/stores/context.store'

const contextStore = useContextStore()

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}
</script>
