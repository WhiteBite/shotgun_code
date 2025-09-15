<template>
  <div class="error-state">
    <div class="error-content">
      <div class="error-icon">
        <AlertTriangleIcon class="icon" />
      </div>
      <div class="error-details">
        <h4 class="error-title">Something went wrong</h4>
        <p class="error-message">{{ error.message }}</p>
        <p v-if="error.code" class="error-code">Error code: {{ error.code }}</p>
      </div>
      <button 
        v-if="showRetry"
        class="error-retry-btn"
        @click="$emit('retry')"
      >
        <RefreshCwIcon class="retry-icon" />
        Try Again
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { AlertTriangleIcon, RefreshCwIcon } from 'lucide-vue-next'

interface PanelError {
  message: string
  code?: string
}

interface Props {
  error: PanelError
  showRetry?: boolean
}

withDefaults(defineProps<Props>(), {
  showRetry: true
})

defineEmits<{
  retry: []
}>()
</script>

<style scoped>
.error-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
  min-height: 120px;
}

.error-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  max-width: 320px;
  text-align: center;
}

.error-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 50%;
}

.error-icon .icon {
  width: 24px;
  height: 24px;
  color: rgb(239, 68, 68);
}

.error-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.error-title {
  font-size: 1rem;
  font-weight: 600;
  color: rgb(248, 250, 252);
  margin: 0;
}

.error-message {
  font-size: 0.875rem;
  color: rgb(148, 163, 184);
  margin: 0;
  line-height: 1.5;
}

.error-code {
  font-size: 0.75rem;
  color: rgb(100, 116, 139);
  margin: 0;
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
}

.error-retry-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 8px;
  color: rgb(147, 197, 253);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.error-retry-btn:hover {
  background: rgba(59, 130, 246, 0.2);
  border-color: rgba(59, 130, 246, 0.5);
  color: rgb(147, 197, 253);
}

.retry-icon {
  width: 16px;
  height: 16px;
}
</style>