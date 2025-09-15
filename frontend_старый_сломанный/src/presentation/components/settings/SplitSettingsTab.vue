<template>
  <div class="space-y-6">
    <h3 class="text-lg font-semibold text-white flex items-center gap-2">
      <svg 
        class="w-5 h-5 text-orange-400" 
        fill="none" 
        stroke="currentColor" 
        viewBox="0 0 24 24"
      >
        <path 
          stroke-linecap="round" 
          stroke-linejoin="round" 
          stroke-width="2" 
          d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4"
        />
      </svg>
      Context Splitting
    </h3>
    
    <div class="space-y-4">
      <!-- Max Tokens Per Chunk -->
      <label class="block">
        <span class="text-sm font-medium text-gray-300">Max Tokens Per Chunk</span>
        <p class="text-xs text-gray-400 mt-1">
          Maximum number of tokens allowed in each context chunk when splitting large contexts
        </p>
        <input
          v-model.number="localSettings.maxTokensPerChunk"
          type="number"
          min="1000"
          max="50000"
          step="500"
          class="mt-2 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          @blur="sanitizeMaxTokensPerChunk"
          @input="updateSettings"
        />
      </label>

      <!-- Enable Auto-Splitting -->
      <label class="flex items-start gap-3">
        <input
          v-model="localSettings.enableAutoSplitting"
          type="checkbox"
          class="mt-1 w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
          @change="updateSettings"
        />
        <div>
          <span class="text-sm text-gray-300">Enable Auto-Splitting</span>
          <p class="text-xs text-gray-400 mt-1">
            Automatically split large contexts that exceed the token limit
          </p>
        </div>
      </label>

      <!-- Overlap Size -->
      <label class="block">
        <span class="text-sm font-medium text-gray-300">Overlap Size (tokens)</span>
        <p class="text-xs text-gray-400 mt-1">
          Number of tokens to overlap between chunks to maintain context continuity
        </p>
        <input
          v-model.number="localSettings.overlapSize"
          type="number"
          min="0"
          max="2000"
          step="50"
          class="mt-2 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          :disabled="!localSettings.enableAutoSplitting"
          @blur="sanitizeOverlapSize"
          @input="updateSettings"
        />
      </label>

      <!-- Split Strategy -->
      <label class="block">
        <span class="text-sm font-medium text-gray-300">Split Strategy</span>
        <p class="text-xs text-gray-400 mt-1">
          How to split large contexts when they exceed the token limit
        </p>
        <select
          v-model="localSettings.splitStrategy"
          class="mt-2 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          @change="updateSettings"
        >
          <option value="semantic">Semantic (preserve code blocks)</option>
          <option value="balanced">Balanced (equal chunks)</option>
          <option value="priority">Priority (important files first)</option>
        </select>
      </label>

      <!-- Preserve Code Blocks -->
      <label class="flex items-start gap-3">
        <input
          v-model="localSettings.preserveCodeBlocks"
          type="checkbox"
          class="mt-1 w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
          @change="updateSettings"
        />
        <div>
          <span class="text-sm text-gray-300">Preserve Code Blocks</span>
          <p class="text-xs text-gray-400 mt-1">
            Avoid splitting in the middle of functions, classes, or other code blocks
          </p>
        </div>
      </label>

      <!-- Info Box -->
      <div class="p-4 bg-blue-900/30 border border-blue-700 rounded-lg">
        <div class="flex items-start gap-2">
          <svg
            class="w-5 h-5 text-blue-400 mt-0.5 flex-shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <div>
            <p class="text-blue-300 text-sm font-medium">Context Splitting</p>
            <p class="text-blue-200 text-xs mt-1">
              Context splitting helps handle large codebases by breaking them into manageable chunks. 
              This prevents memory issues and allows for better AI processing of large projects.
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useSplitSettingsService } from '@/composables/useSplitSettingsService'
import type { SplitSettings } from '@/domain/services/SplitSettingsService'

const props = defineProps<{
  settings: SplitSettings
}>()

const emit = defineEmits<{
  'update:settings': [settings: SplitSettings]
}>()

const { splitSettingsService } = useSplitSettingsService()

const localSettings = ref<SplitSettings>({ ...props.settings })

// Watch for prop changes
watch(
  () => props.settings,
  (newSettings) => {
    localSettings.value = { ...newSettings }
  },
  { deep: true }
)

const updateSettings = () => {
  emit('update:settings', { ...localSettings.value })
}

const sanitizeMaxTokensPerChunk = () => {
  const val = Number(localSettings.value.maxTokensPerChunk)
  localSettings.value.maxTokensPerChunk = splitSettingsService.sanitizeMaxTokensPerChunk(val)
  updateSettings()
}

const sanitizeOverlapSize = () => {
  const val = Number(localSettings.value.overlapSize)
  localSettings.value.overlapSize = splitSettingsService.sanitizeOverlapSize(val)
  updateSettings()
}
</script>