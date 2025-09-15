<template>
  <div
    class="animate-spin rounded-full border-2 border-current border-t-transparent"
    :class="[
      size === 'xs' ? 'w-3 h-3' :
      size === 'sm' ? 'w-4 h-4' :
      size === 'md' ? 'w-6 h-6' :
      size === 'lg' ? 'w-8 h-8' :
      size === 'xl' ? 'w-12 h-12' : 'w-6 h-6',
      colorClass
    ]"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  color?: 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'current'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  color: 'current'
})

const colorClass = computed(() => {
  switch (props.color) {
    case 'primary':
      return 'text-blue-500'
    case 'secondary':
      return 'text-gray-500'
    case 'success':
      return 'text-green-500'
    case 'warning':
      return 'text-yellow-500'
    case 'error':
      return 'text-red-500'
    default:
      return 'text-current'
  }
})
</script>