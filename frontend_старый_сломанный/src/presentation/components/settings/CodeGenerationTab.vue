<template>
  <div class="space-y-6">
    <h3 class="text-lg font-semibold text-white flex items-center gap-2">
      <svg
        class="w-5 h-5 text-yellow-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
        />
      </svg>
      Code Generation
    </h3>

    <div class="space-y-4">
      <label class="block">
        <span class="text-sm font-medium text-gray-300">Temperature</span>
        <div class="mt-1 flex items-center space-x-2">
          <input
            v-model.number="localSettings.temperature"
            type="range"
            min="0"
            max="2"
            step="0.1"
            class="flex-1 h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer slider"
            @input="onSettingsChange"
          />
          <span class="text-sm text-gray-400 w-12">{{ localSettings.temperature }}</span>
        </div>
        <p class="text-xs text-gray-500 mt-1">
          Lower = more focused and deterministic, Higher = more creative and varied
        </p>
      </label>

      <label class="block">
        <span class="text-sm font-medium text-gray-300">Max Tokens</span>
        <input
          v-model.number="localSettings.maxTokens"
          type="number"
          min="100"
          max="8000"
          step="100"
          class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          @blur="sanitizeMaxTokens"
          @input="onSettingsChange"
        />
        <p class="text-xs text-gray-500 mt-1">
          Maximum tokens for generated code. Higher values allow longer responses.
        </p>
      </label>

      <div class="space-y-3">
        <label class="flex items-center">
          <input
            v-model="localSettings.autoFormat"
            type="checkbox"
            class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="onSettingsChange"
          />
          <span class="ml-2 text-sm text-gray-300">Auto-format generated code</span>
        </label>
        <p class="text-xs text-gray-500 ml-6">
          Automatically format code using Prettier or language-specific formatters
        </p>

        <label class="flex items-center">
          <input
            v-model="localSettings.includeComments"
            type="checkbox"
            class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="onSettingsChange"
          />
          <span class="ml-2 text-sm text-gray-300">Include comments in generated code</span>
        </label>
        <p class="text-xs text-gray-500 ml-6">
          Add explanatory comments to generated code for better understanding
        </p>
      </div>

      <!-- Quality Indicators -->
      <div class="bg-gray-800 p-3 rounded-lg">
        <h4 class="text-sm font-medium text-gray-300 mb-2">Generation Quality</h4>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <div class="text-xs text-gray-400">Creativity</div>
            <div class="flex items-center gap-2 mt-1">
              <div class="w-2 h-2 rounded-full" :class="creativityClass"></div>
              <span class="text-sm" :class="creativityTextClass">{{ creativityText }}</span>
            </div>
          </div>
          <div>
            <div class="text-xs text-gray-400">Reliability</div>
            <div class="flex items-center gap-2 mt-1">
              <div class="w-2 h-2 rounded-full" :class="reliabilityClass"></div>
              <span class="text-sm" :class="reliabilityTextClass">{{ reliabilityText }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useAISettingsService } from '@/composables/useAISettingsService'
import type { CodeGenerationSettings } from '@/domain/services/AISettingsService'

const props = defineProps<{
  settings: CodeGenerationSettings
}>()

const emit = defineEmits<{
  'update:settings': [settings: CodeGenerationSettings]
  'settings-changed': [settings: CodeGenerationSettings]
}>()

const { aiSettingsService } = useAISettingsService()

const localSettings = ref<CodeGenerationSettings>({ ...props.settings })

// Quality indicators
const creativityClass = computed(() => {
  return aiSettingsService.calculateCreativityClass(localSettings.value.temperature)
})

const creativityTextClass = computed(() => {
  return aiSettingsService.calculateCreativityTextClass(localSettings.value.temperature)
})

const creativityText = computed(() => {
  return aiSettingsService.calculateCreativityText(localSettings.value.temperature)
})

const reliabilityClass = computed(() => {
  return aiSettingsService.calculateReliabilityClass(localSettings.value.temperature)
})

const reliabilityTextClass = computed(() => {
  return aiSettingsService.calculateReliabilityTextClass(localSettings.value.temperature)
})

const reliabilityText = computed(() => {
  return aiSettingsService.calculateReliabilityText(localSettings.value.temperature)
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

const sanitizeMaxTokens = () => {
  const val = Number(localSettings.value.maxTokens)
  localSettings.value.maxTokens = aiSettingsService.sanitizeMaxTokens(val)
  onSettingsChange()
}
</script>

<style scoped>
.slider::-webkit-slider-thumb {
  appearance: none;
  height: 16px;
  width: 16px;
  border-radius: 50%;
  background: #3b82f6;
  cursor: pointer;
  border: 2px solid #1e40af;
}

.slider::-moz-range-thumb {
  height: 16px;
  width: 16px;
  border-radius: 50%;
  background: #3b82f6;
  cursor: pointer;
  border: 2px solid #1e40af;
}
</style>