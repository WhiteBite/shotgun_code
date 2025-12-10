<template>
  <label class="toggle-label">
    <div 
      class="toggle-switch"
      :class="{ 'toggle-switch-active': modelValue }"
      @click.prevent="handleToggle"
    >
      <div 
        class="toggle-switch-thumb"
        :class="{ 'toggle-switch-thumb-active': modelValue }"
      />
    </div>
    <span class="toggle-text">{{ label }}</span>
  </label>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps<{
  modelValue: boolean
  label: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const isAnimating = ref(false)

function handleToggle() {
  if (isAnimating.value) return
  isAnimating.value = true
  emit('update:modelValue', !props.modelValue)
  setTimeout(() => {
    isAnimating.value = false
  }, 250)
}
</script>
