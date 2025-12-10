<template>
  <div class="context-stats">
    <label class="text-xs text-gray-400 mb-2 block">{{ t('settings.selectProvider') }}</label>
    <div class="relative">
      <select
        :value="modelValue"
        @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
        class="input input-md pr-10 appearance-none cursor-pointer"
      >
        <option v-for="provider in providers" :key="provider.id" :value="provider.id">
          {{ provider.icon }} {{ provider.name }}
        </option>
      </select>
      <div class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-gray-400">
        <ChevronDown class="w-4 h-4" />
      </div>
    </div>
    <p v-if="description" class="text-2xs text-gray-500 mt-1.5">{{ description }}</p>
  </div>
</template>

<script setup lang="ts">
import type { AIProvider } from '@/composables/useAISettings';
import { useI18n } from '@/composables/useI18n';
import { ChevronDown } from 'lucide-vue-next';

const { t } = useI18n()

defineProps<{
  modelValue: string
  providers: AIProvider[]
  description?: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
