<template>
  <div class="space-y-6">
    <h3 class="text-lg font-semibold text-white flex items-center gap-2">
      <svg
        class="w-5 h-5 text-green-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
        />
      </svg>
      Context Settings
    </h3>

    <div class="space-y-4">
      <label class="block">
        <span class="text-sm font-medium text-gray-300">Max Context Size (tokens)</span>
        <input
          v-model.number="localSettings.maxContextSize"
          type="number"
          min="1000"
          max="100000"
          step="1000"
          class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          @blur="sanitizeMaxContextSize"
          @input="onSettingsChange"
        />
        <p class="text-xs text-gray-500 mt-1">
          Maximum tokens allowed in context. Recommended: 8000-16000 for most models.
        </p>
      </label>

      <label class="block">
        <span class="text-sm font-medium text-gray-300">Max Files in Context</span>
        <input
          v-model.number="localSettings.maxFilesInContext"
          type="number"
          min="1"
          max="100"
          class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          @blur="sanitizeMaxFiles"
          @input="onSettingsChange"
        />
        <p class="text-xs text-gray-500 mt-1">
          Maximum number of files to include in context. Helps prevent overwhelming the AI.
        </p>
      </label>

      <div class="space-y-3">
        <label class="flex items-center">
          <input
            v-model="localSettings.includeDependencies"
            type="checkbox"
            class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="onSettingsChange"
          />
          <span class="ml-2 text-sm text-gray-300">Include dependencies in context</span>
        </label>
        <p class="text-xs text-gray-500 ml-6">
          Automatically include import dependencies when building context
        </p>

        <label class="flex items-center">
          <input
            v-model="localSettings.includeTests"
            type="checkbox"
            class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="onSettingsChange"
          />
          <span class="ml-2 text-sm text-gray-300">Include test files in context</span>
        </label>
        <p class="text-xs text-gray-500 ml-6">
          Include *.test.* and *.spec.* files in context building
        </p>
      </div>

      <!-- Context Preview -->
      <div class="bg-gray-800 p-3 rounded-lg">
        <h4 class="text-sm font-medium text-gray-300 mb-2">Context Preview</h4>
        <div class="space-y-1 text-xs text-gray-400">
          <div class="flex justify-between">
            <span>Estimated cost per request:</span>
            <span class="text-green-400">${{ estimatedCost }}</span>
          </div>
          <div class="flex justify-between">
            <span>Token efficiency:</span>
            <span :class="efficiencyClass">{{ efficiency }}%</span>
          </div>
          <div class="flex justify-between">
            <span>Memory usage estimate:</span>
            <span :class="memoryUsageClass">{{ memoryUsage }}</span>
          </div>
        </div>
      </div>

      <!-- Optimization Suggestions -->
      <div v-if="suggestions.length > 0" class="bg-blue-900/30 border border-blue-700 rounded-lg p-3">
        <h4 class="text-sm font-medium text-blue-300 mb-2 flex items-center gap-2">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          Optimization Suggestions
        </h4>
        <ul class="space-y-1 text-xs text-blue-200">
          <li v-for="suggestion in suggestions" :key="suggestion" class="flex items-start gap-2">
            <span class="text-blue-400 mt-0.5">•</span>
            <span>{{ suggestion }}</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useContextSettingsService } from '@/composables/useContextSettingsService'
import type { ContextSettings } from '@/domain/services/ContextSettingsService'

const props = defineProps<{
  settings: ContextSettings
}>()

const emit = defineEmits<{
  'update:settings': [settings: ContextSettings]
  'settings-changed': [settings: ContextSettings]
}>()

const { contextSettingsService } = useContextSettingsService()

const localSettings = ref<ContextSettings>({ ...props.settings })

// Computed properties for analysis
const estimatedCost = computed(() => {
  return contextSettingsService.calculateEstimatedCost(localSettings.value.maxContextSize)
})

const efficiency = computed(() => {
  return contextSettingsService.calculateEfficiency(
    localSettings.value.maxContextSize,
    localSettings.value.maxFilesInContext
  )
})

const efficiencyClass = computed(() => {
  return contextSettingsService.getEfficiencyClass(efficiency.value)
})

const memoryUsage = computed(() => {
  return contextSettingsService.calculateMemoryUsage(
    localSettings.value.maxContextSize,
    localSettings.value.maxFilesInContext
  )
})

const memoryUsageClass = computed(() => {
  return contextSettingsService.getMemoryUsageClass(
    localSettings.value.maxContextSize,
    localSettings.value.maxFilesInContext
  )
})

const suggestions = computed(() => {
  return contextSettingsService.getSuggestions(
    localSettings.value.maxContextSize,
    localSettings.value.maxFilesInContext,
    localSettings.value.includeDependencies,
    localSettings.value.includeTests
  )
})

// Watch for external settings changes
watch(
  () => props.settings,
  (newSettings) => {
    localSettings.value = { ...newSettings }
  },
  { deep: true }
)

// Watch for local settings changes and emit updates
watch(
  localSettings,
  (newSettings) => {
    emit('update:settings', newSettings)
  },
  { deep: true }
)

const onSettingsChange = () => {
  emit('settings-changed', localSettings.value)
  emit('update:settings', localSettings.value)
}

const sanitizeMaxContextSize = () => {
  const val = Number(localSettings.value.maxContextSize)
  localSettings.value.maxContextSize = contextSettingsService.sanitizeMaxContextSize(val)
  onSettingsChange()
}

const sanitizeMaxFiles = () => {
  const val = Number(localSettings.value.maxFilesInContext)
  localSettings.value.maxFilesInContext = contextSettingsService.sanitizeMaxFiles(val)
  onSettingsChange()
}
</script>