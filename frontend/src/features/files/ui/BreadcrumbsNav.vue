<template>
  <div class="flex items-center gap-1 text-xs text-gray-400">
    <template v-for="(segment, index) in segments" :key="index">
      <div
        :class="[
          'px-2 py-1 rounded flex items-center gap-1',
          index === segments.length - 1
            ? 'text-white bg-gray-700'
            : 'text-gray-400'
        ]"
        :title="getFullPath(index)"
      >
        <svg v-if="index === 0" class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
          <path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z" />
        </svg>
        <span class="truncate max-w-[120px]">{{ segment }}</span>
      </div>
      <svg v-if="index < segments.length - 1" class="w-3 h-3 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
        <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
      </svg>
    </template>
    
    <!-- Copy path button -->
    <button
      v-if="segments.length > 0"
      @click="copyPath"
      class="ml-2 p-1 hover:bg-gray-700 rounded transition-colors"
      title="Copy full path"
    >
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
interface Props {
  segments: string[]
  rootName?: string
}

const props = withDefaults(defineProps<Props>(), {
  rootName: 'Project'
})

function getFullPath(index: number): string {
  return props.segments.slice(0, index + 1).join(' / ')
}

async function copyPath() {
  const fullPath = props.segments.join(' / ')
  try {
    await navigator.clipboard.writeText(fullPath)
  } catch (err) {
    console.warn('Failed to copy path:', err)
  }
}
</script>
