<template>
  <div class="space-y-6">
    <h3 class="text-lg font-semibold text-white flex items-center gap-2">
      <svg
        class="w-5 h-5 text-cyan-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
        />
      </svg>
      Context Builder
    </h3>

    <div class="space-y-4">
      <label class="block">
        <span class="text-sm font-medium text-gray-300">
          Max Token Limit (0 = unlimited)
        </span>
        <input
          v-model.number="localSettings.maxTokenLimit"
          type="number"
          min="0"
          max="100000"
          step="1000"
          class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          placeholder="0 for unlimited"
          @input="onSettingsChange"
        />
        <p class="text-xs text-gray-500 mt-1">
          Maximum tokens allowed in context (0 for no limit). Overrides per-model limits.
        </p>
      </label>

      <label class="block">
        <span class="text-sm font-medium text-gray-300">
          Max File Limit (0 = unlimited)
        </span>
        <input
          v-model.number="localSettings.maxFileLimit"
          type="number"
          min="0"
          max="1000"
          step="1"
          class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          placeholder="0 for unlimited"
          @input="onSettingsChange"
        />
        <p class="text-xs text-gray-500 mt-1">
          Maximum files allowed in context (0 for no limit). Helps prevent memory issues.
        </p>
      </label>

      <div class="space-y-3">
        <label class="flex items-center">
          <input
            v-model="localSettings.excludeBinaryFiles"
            type="checkbox"
            class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="onSettingsChange"
          />
          <span class="ml-2 text-sm text-gray-300">
            Exclude binary files by default
          </span>
        </label>
        <p class="text-xs text-gray-500 ml-6">
          Automatically exclude images, videos, executables, and other binary files
        </p>

        <label class="flex items-center">
          <input
            v-model="localSettings.autoBuildEnabled"
            type="checkbox"
            class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="onSettingsChange"
          />
          <span class="ml-2 text-sm text-gray-300">
            Auto-build context on file selection
          </span>
        </label>
        <p class="text-xs text-gray-500 ml-6">
          Automatically rebuild context when files are selected or deselected
        </p>

        <label class="flex items-center">
          <input
            v-model="localSettings.smartSuggestionsEnabled"
            type="checkbox"
            class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="onSettingsChange"
          />
          <span class="ml-2 text-sm text-gray-300">
            Enable smart file suggestions
          </span>
        </label>
        <p class="text-xs text-gray-500 ml-6">
          AI-powered suggestions for relevant files based on your selection
        </p>
      </div>

      <div class="bg-gray-800 p-3 rounded-lg">
        <label class="block">
          <span class="text-sm font-medium text-gray-300">
            Allowed File Extensions
          </span>
          <textarea
            v-model="allowedExtensionsText"
            rows="3"
            class="mt-1 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder=".ts, .js, .vue, .py, .md, .txt..."
            @input="onAllowedExtensionsChange"
          />
          <p class="text-xs text-gray-500 mt-1">
            Comma-separated list of file extensions to include when excluding binary files
          </p>
        </label>
      </div>

      <!-- Memory Management Settings -->
      <div class="bg-gray-800 p-3 rounded-lg">
        <h4 class="text-sm font-medium text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3"/>
          </svg>
          Memory Management
        </h4>
        
        <div class="space-y-3">
          <label class="block">
            <span class="text-sm font-medium text-gray-300">Max Individual File Size</span>
            <div class="flex items-center gap-2 mt-1">
              <input
                v-model.number="maxIndividualFileSizeKB"
                type="number"
                min="1"
                max="10240"
                step="1"
                class="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                @input="onMemorySettingsChange"
              />
              <span class="text-sm text-gray-400">KB</span>
            </div>
            <p class="text-xs text-gray-500 mt-1">
              Skip files larger than this size to prevent memory issues
            </p>
          </label>

          <label class="block">
            <span class="text-sm font-medium text-gray-300">Max Total Context Size</span>
            <div class="flex items-center gap-2 mt-1">
              <input
                v-model.number="maxTotalContextSizeMB"
                type="number"
                min="1"
                max="100"
                step="1"
                class="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                @input="onMemorySettingsChange"
              />
              <span class="text-sm text-gray-400">MB</span>
            </div>
            <p class="text-xs text-gray-500 mt-1">
              Maximum total size of all context content combined
            </p>
          </label>
        </div>
      </div>

      <!-- Performance Indicators -->
      <div class="grid grid-cols-2 gap-3">
        <div class="bg-gray-800 p-3 rounded-lg">
          <div class="text-xs text-gray-400">Memory Safety</div>
          <div class="flex items-center gap-2 mt-1">
            <div class="w-2 h-2 rounded-full" :class="memorySafetyClass"></div>
            <span class="text-sm" :class="memorySafetyTextClass">{{ memorySafetyText }}</span>
          </div>
        </div>
        
        <div class="bg-gray-800 p-3 rounded-lg">
          <div class="text-xs text-gray-400">Performance</div>
          <div class="flex items-center gap-2 mt-1">
            <div class="w-2 h-2 rounded-full" :class="performanceClass"></div>
            <span class="text-sm" :class="performanceTextClass">{{ performanceText }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';

interface ContextBuilderSettings {
  maxTokenLimit: number;
  maxFileLimit: number;
  excludeBinaryFiles: boolean;
  autoBuildEnabled: boolean;
  smartSuggestionsEnabled: boolean;
  allowedExtensions: string[];
  maxIndividualFileSize?: number; // in bytes
  maxTotalContextSize?: number; // in bytes
}

const props = defineProps<{
  settings: ContextBuilderSettings;
}>();

const emit = defineEmits<{
  'update:settings': [settings: ContextBuilderSettings];
  'settings-changed': [settings: ContextBuilderSettings];
}>();

const localSettings = ref<ContextBuilderSettings>({ ...props.settings });

// Convert bytes to user-friendly units for display
const maxIndividualFileSizeKB = computed({
  get: () => Math.round((localSettings.value.maxIndividualFileSize || 128 * 1024) / 1024),
  set: (value: number) => {
    localSettings.value.maxIndividualFileSize = value * 1024;
    onSettingsChange();
  }
});

const maxTotalContextSizeMB = computed({
  get: () => Math.round((localSettings.value.maxTotalContextSize || 1 * 1024 * 1024) / (1024 * 1024)),
  set: (value: number) => {
    localSettings.value.maxTotalContextSize = value * 1024 * 1024;
    onSettingsChange();
  }
});

const allowedExtensionsText = computed({
  get: () => localSettings.value.allowedExtensions.join(', '),
  set: (value: string) => {
    localSettings.value.allowedExtensions = value
      .split(',')
      .map(ext => ext.trim())
      .filter(ext => ext.length > 0 && ext.startsWith('.'))
      .filter((ext, index, arr) => arr.indexOf(ext) === index); // Remove duplicates
    onSettingsChange();
  }
});

// Performance and safety indicators
const memorySafetyClass = computed(() => {
  const fileSizeKB = maxIndividualFileSizeKB.value;
  const totalSizeMB = maxTotalContextSizeMB.value;
  
  if (fileSizeKB > 1024 || totalSizeMB > 10) return 'bg-red-400'; // High risk
  if (fileSizeKB > 512 || totalSizeMB > 5) return 'bg-yellow-400'; // Medium risk
  return 'bg-green-400'; // Low risk
});

const memorySafetyTextClass = computed(() => {
  const fileSizeKB = maxIndividualFileSizeKB.value;
  const totalSizeMB = maxTotalContextSizeMB.value;
  
  if (fileSizeKB > 1024 || totalSizeMB > 10) return 'text-red-400';
  if (fileSizeKB > 512 || totalSizeMB > 5) return 'text-yellow-400';
  return 'text-green-400';
});

const memorySafetyText = computed(() => {
  const fileSizeKB = maxIndividualFileSizeKB.value;
  const totalSizeMB = maxTotalContextSizeMB.value;
  
  if (fileSizeKB > 1024 || totalSizeMB > 10) return 'High Risk';
  if (fileSizeKB > 512 || totalSizeMB > 5) return 'Medium Risk';
  return 'Safe';
});

const performanceClass = computed(() => {
  const tokenLimit = localSettings.value.maxTokenLimit;
  const fileLimit = localSettings.value.maxFileLimit;
  const autoEnabled = localSettings.value.autoBuildEnabled;
  
  // Performance impact factors
  let score = 100;
  
  if (tokenLimit === 0 || tokenLimit > 32000) score -= 30; // No limit or very high limit
  if (fileLimit === 0 || fileLimit > 50) score -= 20; // No limit or many files
  if (autoEnabled) score -= 10; // Auto-build can impact performance
  
  if (score >= 80) return 'bg-green-400';
  if (score >= 60) return 'bg-yellow-400';
  return 'bg-red-400';
});

const performanceTextClass = computed(() => {
  const tokenLimit = localSettings.value.maxTokenLimit;
  const fileLimit = localSettings.value.maxFileLimit;
  const autoEnabled = localSettings.value.autoBuildEnabled;
  
  let score = 100;
  if (tokenLimit === 0 || tokenLimit > 32000) score -= 30;
  if (fileLimit === 0 || fileLimit > 50) score -= 20;
  if (autoEnabled) score -= 10;
  
  if (score >= 80) return 'text-green-400';
  if (score >= 60) return 'text-yellow-400';
  return 'text-red-400';
});

const performanceText = computed(() => {
  const tokenLimit = localSettings.value.maxTokenLimit;
  const fileLimit = localSettings.value.maxFileLimit;
  const autoEnabled = localSettings.value.autoBuildEnabled;
  
  let score = 100;
  if (tokenLimit === 0 || tokenLimit > 32000) score -= 30;
  if (fileLimit === 0 || fileLimit > 50) score -= 20;
  if (autoEnabled) score -= 10;
  
  if (score >= 80) return 'Optimal';
  if (score >= 60) return 'Good';
  return 'Slow';
});

// Watch for external settings changes
watch(
  () => props.settings,
  (newSettings) => {
    localSettings.value = { ...newSettings };
  },
  { deep: true }
);

// Watch for local settings changes and emit updates
watch(
  localSettings,
  (newSettings) => {
    emit('update:settings', newSettings);
  },
  { deep: true }
);

const onSettingsChange = () => {
  emit('settings-changed', localSettings.value);
  emit('update:settings', localSettings.value);
};

const onAllowedExtensionsChange = () => {
  // Trigger computed setter
  allowedExtensionsText.value = allowedExtensionsText.value;
};

const onMemorySettingsChange = () => {
  onSettingsChange();
};
</script>