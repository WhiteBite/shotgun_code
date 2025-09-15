<template>
  <div 
    class="w-2 h-2 rounded-full transition-all duration-300"
    :class="statusClasses"
    :title="statusTooltip"
  >
    <div 
      v-if="status === 'healthy'"
      class="w-full h-full rounded-full animate-pulse"
      :class="statusClasses"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type StatusType = 'none' | 'idle' | 'ready' | 'healthy' | 'error' | 'warning'

interface Props {
  status: StatusType
}

const props = defineProps<Props>()

const statusClasses = computed(() => {
  switch (props.status) {
    case 'healthy':
      return 'bg-green-400 shadow-lg shadow-green-400/50'
    case 'ready':
      return 'bg-blue-400 shadow-lg shadow-blue-400/50'
    case 'warning':
      return 'bg-yellow-400 shadow-lg shadow-yellow-400/50'
    case 'error':
      return 'bg-red-400 shadow-lg shadow-red-400/50'
    case 'idle':
      return 'bg-gray-400'
    default:
      return 'bg-gray-600'
  }
})

const statusTooltip = computed(() => {
  switch (props.status) {
    case 'healthy':
      return 'Project is healthy with active context'
    case 'ready':
      return 'Project is ready with selected files'
    case 'warning':
      return 'Project has warnings or issues'
    case 'error':
      return 'Project has errors'
    case 'idle':
      return 'Project is loaded but inactive'
    default:
      return 'No project loaded'
  }
})
</script>