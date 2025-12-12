<template>
  <div class="flex flex-col items-center justify-center h-full text-center p-4">
    <div class="text-4xl mb-4">🤖</div>
    <h3 class="text-lg font-medium text-gray-200 mb-2">
      {{ t('chat.welcome.title') }}
    </h3>
    <p class="text-sm text-gray-400 mb-4">
      {{ isConnected 
        ? t('chat.welcome.connected', { provider: providerName, model: currentModel })
        : t('chat.welcome.disconnected') 
      }}
    </p>
    <div class="flex flex-wrap gap-2 justify-center">
      <button
        v-for="action in quickActions"
        :key="action.id"
        class="btn btn-sm btn-ghost"
        @click="$emit('quick-action', { prompt: action.prompt })"
      >
        {{ action.label }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';
import { computed } from 'vue';

defineProps<{
  isConnected: boolean
  providerName: string
  currentModel: string
}>()

defineEmits<{
  'quick-action': [action: { prompt: string }]
}>()

const { t } = useI18n()

const quickActions = computed(() => [
  { id: 'analyze', label: t('chat.actions.analyze'), prompt: t('chat.prompts.analyze') },
  { id: 'explain', label: t('chat.actions.explain'), prompt: t('chat.prompts.explain') },
  { id: 'refactor', label: t('chat.actions.refactor'), prompt: t('chat.prompts.refactor') },
])
</script>
