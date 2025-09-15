<template>
  <button
    class="p-2 rounded-lg transition-all duration-200 group"
    :class="[
      disabled 
        ? 'opacity-50 cursor-not-allowed text-gray-500' 
        : 'text-gray-400 hover:text-white hover:bg-gray-700 active:bg-gray-600',
      size === 'sm' ? 'p-1.5' : size === 'lg' ? 'p-3' : 'p-2'
    ]"
    :disabled="disabled"
    :title="tooltip"
    @click="handleClick"
  >
    <component 
      :is="iconComponent" 
      :class="[
        size === 'sm' ? 'w-4 h-4' : size === 'lg' ? 'w-6 h-6' : 'w-5 h-5'
      ]"
    />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import * as HeroIcons from '@heroicons/vue/24/outline'

interface Props {
  icon: string
  tooltip?: string
  disabled?: boolean
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  disabled: false
})

const emit = defineEmits<{
  click: []
}>()

const iconComponent = computed(() => {
  // @ts-expect-error - Dynamic icon import
  return HeroIcons[props.icon] || HeroIcons.QuestionMarkCircleIcon
})

const handleClick = () => {
  if (!props.disabled) {
    emit('click')
  }
}
</script>