<template>
  <div class="context-stats">
    <div class="flex items-center justify-between mb-2">
      <label class="text-xs text-gray-400">{{ t('export.tokenLimit') }}</label>
      <div class="flex items-center gap-2">
        <input 
          type="number"
          :value="modelValue"
          @input="handleInput"
          min="1000"
          max="1000000"
          step="1000"
          class="input text-xs py-0.5 px-2 w-20 text-right"
        />
        <span class="text-xs text-gray-500">{{ t('limits.tokens') }}</span>
      </div>
    </div>
    <input 
      type="range" 
      :value="modelValue"
      @input="$emit('update:modelValue', Number(($event.target as HTMLInputElement).value))" 
      min="1000" 
      max="500000"
      step="1000" 
      class="w-full h-1.5 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-indigo-500" 
    />
    <div class="flex justify-between text-2xs text-gray-600 mt-1">
      <span>1K</span>
      <span class="text-gray-500">{{ t('tokenLimit.hint') }}</span>
      <span>500K</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';

const { t } = useI18n()

defineProps<{
  modelValue: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number]
}>()

function handleInput(event: Event) {
  const value = Number((event.target as HTMLInputElement).value)
  // Clamp between 1000 and 1000000
  const clamped = Math.max(1000, Math.min(1000000, value))
  emit('update:modelValue', clamped)
}
</script>
