<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { MemoryErrorHandler, type ErrorCategory } from '@/utils/memory-error-handler';

interface Props {
  retryHandler?: () => void;
  recoveryHandler?: () => void;
  resetHandler?: () => void;
  monitorMemory?: boolean;
  maxRetries?: number; // New prop for max retry attempts
}

const props = withDefaults(defineProps<Props>(), {
  monitorMemory: true,
  maxRetries: 3
});

const emit = defineEmits<{
  (e: 'error', error: Error, category: ErrorCategory): void;
  (e: 'retry'): void;
  (e: 'recover'): void;
  (e: 'reset'): void;
  (e: 'max-retries-exceeded'): void; // New event for when max retries are exceeded
}>();

// State
const hasError = ref(false);
const error = ref<Error | null>(null);
const errorCategory = ref<ErrorCategory>('unknown');
const memoryDetails = ref<{ used: number; total: number; percentage: number } | null>(null);
const retryCount = ref(0); // Track retry attempts
let memoryMonitorInterval: number | null = null;

// Computed properties
const errorTitle = computed(() => {
  switch (errorCategory.value) {
    case 'memory':
      return 'Memory Error';
    case 'context':
      return 'Context Processing Error';
    default:
      return 'Error Occurred';
  }
});

const errorMessage = computed(() => {
  return error.value?.message || 'An unexpected error occurred';
});

const suggestion = computed(() => {
  switch (errorCategory.value) {
    case 'memory':
      return 'Try selecting fewer files, closing other applications, or refreshing the page to free up memory. Consider using the chunked processing feature for large contexts.';
    case 'context':
      return 'The context is too large. Try selecting fewer or smaller files, or use the split context feature. The application has automatic chunked processing for large contexts.';
    default:
      return 'Try refreshing the page or contact support if the problem persists.';
  }
});

const canRetry = computed(() => {
  return retryCount.value < props.maxRetries && (
    !!props.retryHandler || 
    (errorCategory.value === 'memory' && 
    memoryDetails.value && 
    memoryDetails.value.percentage < 85) // Reduced from 90
  );
});

const canRecover = computed(() => {
  return !!props.recoveryHandler || errorCategory.value === 'memory';
});

const memoryBarColorClass = computed(() => {
  if (!memoryDetails.value) return 'bg-gray-500';
  
  const percentage = memoryDetails.value.percentage;
  if (percentage > 95) return 'bg-red-600';
  if (percentage > 85) return 'bg-orange-500';
  if (percentage > 70) return 'bg-yellow-500';
  return 'bg-green-500';
});

// Error handler method
const handleError = (err: Error) => {
  error.value = err;
  errorCategory.value = MemoryErrorHandler.categorizeError(err);
  hasError.value = true;
  
  // Update memory stats
  updateMemoryStats();
  
  // Start memory monitoring if not already started
  if (props.monitorMemory && !memoryMonitorInterval) {
    memoryMonitorInterval = window.setInterval(updateMemoryStats, 2000);
  }
  
  // Emit error event
  emit('error', err, errorCategory.value);
};

// Actions
const retry = () => {
  if (!canRetry.value) {
    emit('max-retries-exceeded');
    return;
  }
  
  retryCount.value++;
  
  if (props.retryHandler) {
    props.retryHandler();
  }
  
  // Reset error state
  hasError.value = false;
  error.value = null;
  
  emit('retry');
};

const attemptRecovery = () => {
  if (props.recoveryHandler) {
    props.recoveryHandler();
  } else {
    // Default recovery for memory errors
    if (errorCategory.value === 'memory') {
      // Force garbage collection if available
      if (window.gc) {
        try {
          window.gc();
        } catch (e) {
          console.warn('Failed to trigger garbage collection', e);
        }
      }
      
      // Clear any temporary data
      clearTemporaryData();
    }
  }
  
  emit('recover');
};

const reset = () => {
  // Reset error state
  hasError.value = false;
  error.value = null;
  retryCount.value = 0; // Reset retry count
  
  if (props.resetHandler) {
    props.resetHandler();
  }
  
  emit('reset');
};

// Memory monitoring
const updateMemoryStats = () => {
  if ('performance' in window && 'memory' in (performance as any)) {
    const memory = (performance as any).memory;
    const usedMB = Math.round(memory.usedJSHeapSize / (1024 * 1024));
    const totalMB = Math.round(memory.jsHeapSizeLimit / (1024 * 1024));
    const percentage = Math.min(100, Math.round((usedMB / totalMB) * 100));
    
    memoryDetails.value = {
      used: usedMB,
      total: totalMB,
      percentage
    };
    
    // If memory usage is critically high, attempt recovery automatically
    if (percentage > 90) {
      console.warn(`Critical memory usage (${percentage}%), attempting automatic recovery`);
      attemptRecovery();
    }
  }
};

// Clear temporary data to free memory
const clearTemporaryData = () => {
  // Clear any large objects from memory
  // This is app-specific and would need to be customized
  console.log('Attempting to clear temporary data to free memory');
  
  // Try to access the context store and clear large objects
  try {
    // This is a simplified approach - in a real implementation, we would have a more direct way to access the store
    console.log('Requesting context store to clear large objects');
  } catch (e) {
    console.warn('Could not access context store to clear large objects', e);
  }
  
  // Trigger garbage collection if available
  if (window.gc) {
    try {
      window.gc();
    } catch (e) {
      console.warn('Failed to trigger garbage collection', e);
    }
  }
  
  // Add a small delay to allow garbage collection to work
  setTimeout(() => {
    console.log('Memory recovery attempt completed');
  }, 100);
};

// Cleanup on unmount
onMounted(() => {
  // Initialize memory stats
  if (props.monitorMemory) {
    updateMemoryStats();
  }
});

onUnmounted(() => {
  if (memoryMonitorInterval) {
    clearInterval(memoryMonitorInterval);
    memoryMonitorInterval = null;
  }
});

// Expose the error handler for use by parent
defineExpose({
  handleError,
  retry,
  reset,
  attemptRecovery
});
</script>

<template>
  <div v-if="hasError" class="error-boundary bg-red-900 border border-red-700 rounded-lg p-4">
    <div class="flex items-start justify-between mb-3">
      <h3 class="text-lg font-semibold text-red-200">{{ errorTitle }}</h3>
      <button 
        @click="reset"
        class="text-red-300 hover:text-white"
        title="Close"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
    
    <div class="mb-4">
      <p class="text-red-300 text-sm mb-2">{{ errorMessage }}</p>
      <p class="text-red-400 text-xs">{{ suggestion }}</p>
    </div>
    
    <!-- Memory usage display -->
    <div v-if="memoryDetails" class="mb-4">
      <div class="flex justify-between text-xs text-red-300 mb-1">
        <span>Memory Usage</span>
        <span>{{ memoryDetails.used }}MB / {{ memoryDetails.total }}MB ({{ memoryDetails.percentage }}%)</span>
      </div>
      <div class="w-full bg-red-800 rounded-full h-2">
        <div 
          class="h-2 rounded-full transition-all duration-300"
          :class="memoryBarColorClass"
          :style="{ width: memoryDetails.percentage + '%' }"
        ></div>
      </div>
    </div>
    
    <div class="flex flex-wrap gap-2">
      <button
        v-if="canRetry"
        @click="retry"
        class="px-3 py-1.5 bg-red-700 hover:bg-red-600 text-white text-sm rounded transition-colors"
      >
        Retry ({{ retryCount }}/{{ props.maxRetries }})
      </button>
      
      <button
        v-if="canRecover"
        @click="attemptRecovery"
        class="px-3 py-1.5 bg-orange-600 hover:bg-orange-500 text-white text-sm rounded transition-colors"
      >
        Recover
      </button>
      
      <button
        @click="reset"
        class="px-3 py-1.5 bg-gray-700 hover:bg-gray-600 text-white text-sm rounded transition-colors"
      >
        Reset
      </button>
    </div>
  </div>
  
  <!-- Render children when no error -->
  <slot v-else />
</template>