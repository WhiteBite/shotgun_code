<template>
  <div class="space-y-6">
    <h3 class="text-lg font-semibold text-white flex items-center gap-2">
      <svg 
        class="w-5 h-5 text-red-400" 
        fill="none" 
        stroke="currentColor" 
        viewBox="0 0 24 24"
      >
        <path 
          stroke-linecap="round" 
          stroke-linejoin="round" 
          stroke-width="2" 
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
        />
      </svg>
      Safety & Validation
    </h3>
    
    <div class="space-y-4">
      <!-- Enable Guardrails -->
      <label class="flex items-start gap-3">
        <input
          v-model="localSettings.enableGuardrails"
          type="checkbox"
          class="mt-1 w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
          @change="updateSettings"
        />
        <div>
          <span class="text-sm text-gray-300">Enable Safety Guardrails</span>
          <p class="text-xs text-gray-400 mt-1">
            Enforce safety policies and validation rules for AI-generated content
          </p>
        </div>
      </label>

      <!-- Memory Limits -->
      <div class="space-y-3">
        <h4 class="text-sm font-medium text-gray-200">Memory Safety Limits</h4>
        
        <label class="block">
          <span class="text-sm text-gray-300">Max Memory Usage Warning (MB)</span>
          <p class="text-xs text-gray-400 mt-1">
            Show warning when memory usage exceeds this threshold
          </p>
          <input
            v-model.number="localSettings.maxMemoryWarning"
            type="number"
            min="10"
            max="1000"
            step="10"
            class="mt-2 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            @blur="sanitizeMemoryWarning"
            @input="updateSettings"
          />
        </label>

        <label class="block">
          <span class="text-sm text-gray-300">Max Context Size (MB)</span>
          <p class="text-xs text-gray-400 mt-1">
            Maximum total size for context content to prevent OOM issues
          </p>
          <input
            v-model.number="localSettings.maxContextSizeMB"
            type="number"
            min="1"
            max="100"
            step="1"
            class="mt-2 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            @blur="sanitizeContextSize"
            @input="updateSettings"
          />
        </label>
      </div>

      <!-- Content Validation -->
      <div class="space-y-3">
        <h4 class="text-sm font-medium text-gray-200">Content Validation</h4>
        
        <label class="flex items-start gap-3">
          <input
            v-model="localSettings.validateFileTypes"
            type="checkbox"
            class="mt-1 w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="updateSettings"
          />
          <div>
            <span class="text-sm text-gray-300">Validate File Types</span>
            <p class="text-xs text-gray-400 mt-1">
              Only allow known file types in context building
            </p>
          </div>
        </label>

        <label class="flex items-start gap-3">
          <input
            v-model="localSettings.excludeBinaryFiles"
            type="checkbox"
            class="mt-1 w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="updateSettings"
          />
          <div>
            <span class="text-sm text-gray-300">Exclude Binary Files</span>
            <p class="text-xs text-gray-400 mt-1">
              Automatically exclude binary files from context building
            </p>
          </div>
        </label>

        <label class="flex items-start gap-3">
          <input
            v-model="localSettings.scanForSecrets"
            type="checkbox"
            class="mt-1 w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="updateSettings"
          />
          <div>
            <span class="text-sm text-gray-300">Scan for Secrets</span>
            <p class="text-xs text-gray-400 mt-1">
              Check for potential secrets or sensitive data before including in context
            </p>
          </div>
        </label>
      </div>

      <!-- AI Safety -->
      <div class="space-y-3">
        <h4 class="text-sm font-medium text-gray-200">AI Safety</h4>
        
        <label class="flex items-start gap-3">
          <input
            v-model="localSettings.enableContentFiltering"
            type="checkbox"
            class="mt-1 w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="updateSettings"
          />
          <div>
            <span class="text-sm text-gray-300">Enable Content Filtering</span>
            <p class="text-xs text-gray-400 mt-1">
              Filter AI responses for potentially harmful or inappropriate content
            </p>
          </div>
        </label>

        <label class="flex items-start gap-3">
          <input
            v-model="localSettings.requireUserConfirmation"
            type="checkbox"
            class="mt-1 w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
            @change="updateSettings"
          />
          <div>
            <span class="text-sm text-gray-300">Require User Confirmation</span>
            <p class="text-xs text-gray-400 mt-1">
              Require explicit confirmation before executing AI-generated code changes
            </p>
          </div>
        </label>

        <label class="block">
          <span class="text-sm text-gray-300">Safety Level</span>
          <p class="text-xs text-gray-400 mt-1">
            Level of safety restrictions to apply
          </p>
          <select
            v-model="localSettings.safetyLevel"
            class="mt-2 w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            @change="updateSettings"
          >
            <option value="relaxed">Relaxed - Minimal restrictions</option>
            <option value="balanced">Balanced - Moderate safety checks</option>
            <option value="strict">Strict - Maximum safety</option>
          </select>
        </label>
      </div>

      <!-- Warning Box -->
      <div class="p-4 bg-yellow-900/30 border border-yellow-700 rounded-lg">
        <div class="flex items-start gap-2">
          <svg
            class="w-5 h-5 text-yellow-400 mt-0.5 flex-shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
            />
          </svg>
          <div>
            <p class="text-yellow-300 text-sm font-medium">Safety Considerations</p>
            <p class="text-yellow-200 text-xs mt-1">
              These settings help protect against memory exhaustion, security risks, and inappropriate AI behavior. 
              Stricter settings provide better safety but may limit functionality.
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';

interface SafetySettings {
  enableGuardrails: boolean;
  maxMemoryWarning: number;
  maxContextSizeMB: number;
  validateFileTypes: boolean;
  excludeBinaryFiles: boolean;
  scanForSecrets: boolean;
  enableContentFiltering: boolean;
  requireUserConfirmation: boolean;
  safetyLevel: 'relaxed' | 'balanced' | 'strict';
}

const props = defineProps<{
  settings: SafetySettings;
}>();

const emit = defineEmits<{
  'update:settings': [settings: SafetySettings];
}>();

const localSettings = ref<SafetySettings>({ ...props.settings });

// Watch for prop changes
watch(
  () => props.settings,
  (newSettings) => {
    localSettings.value = { ...newSettings };
  },
  { deep: true }
);

const updateSettings = () => {
  emit('update:settings', { ...localSettings.value });
};

const sanitizeMemoryWarning = () => {
  const val = Number(localSettings.value.maxMemoryWarning);
  if (Number.isNaN(val)) {
    localSettings.value.maxMemoryWarning = 30;
  } else {
    localSettings.value.maxMemoryWarning = Math.min(1000, Math.max(10, val));
  }
  updateSettings();
};

const sanitizeContextSize = () => {
  const val = Number(localSettings.value.maxContextSizeMB);
  if (Number.isNaN(val)) {
    localSettings.value.maxContextSizeMB = 1;
  } else {
    localSettings.value.maxContextSizeMB = Math.min(100, Math.max(1, val));
  }
  updateSettings();
};
</script>