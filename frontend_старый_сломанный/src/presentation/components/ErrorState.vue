<template>
  <div class="error-state">
    <div class="error-content">
      <div class="error-icon">
        <AlertTriangleIcon class="icon" />
      </div>
      
      <div class="error-details">
        <h4 class="error-title">Something went wrong</h4>
        <p class="error-message">{{ error.message || 'An unexpected error occurred' }}</p>
        <p v-if="error.code" class="error-code">Error Code: {{ error.code }}</p>
      </div>
      
      <div class="error-actions">
        <button
          class="retry-btn"
          @click="$emit('retry')"
        >
          <RefreshCwIcon class="btn-icon" />
          Try Again
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { AlertTriangleIcon, RefreshCwIcon } from 'lucide-vue-next'

interface Props {
  error: {
    message: string
    code?: string
  }
}

defineProps<Props>()

defineEmits<{
  retry: []
}>()
</script>

<style scoped>
.error-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
}

.error-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  max-width: 300px;
}

.error-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: rgba(239, 68, 68, 0.1);
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
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(239, 68, 68);
  margin: 0;
}

.error-message {
  font-size: 0.75rem;
  color: rgb(148, 163, 184);
  margin: 0;
  line-height: 1.5;
}

.error-code {
  font-size: 0.75rem;
  color: rgb(100, 116, 139);
  font-family: monospace;
  margin: 0;
}

.error-actions {
  display: flex;
  gap: 8px;
}

.retry-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(239, 68, 68, 0.2);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 6px;
  color: rgb(248, 113, 113);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.retry-btn:hover {
  background: rgba(239, 68, 68, 0.3);
  border-color: rgba(239, 68, 68, 0.5);
}

.btn-icon {
  width: 14px;
  height: 14px;
}
</style>