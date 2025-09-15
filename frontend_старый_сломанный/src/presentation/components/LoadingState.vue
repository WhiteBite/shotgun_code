<template>
  <div class="loading-state">
    <div class="loading-spinner">
      <div class="spinner-icon">
        <svg width="24" height="24" viewBox="0 0 24 24" class="animate-spin">
          <circle 
            cx="12" 
            cy="12" 
            r="10" 
            stroke="currentColor" 
            stroke-width="2" 
            fill="none" 
            stroke-dasharray="32" 
            stroke-dashoffset="32"
            class="opacity-25"
          />
          <circle 
            cx="12" 
            cy="12" 
            r="10" 
            stroke="currentColor" 
            stroke-width="2" 
            fill="none" 
            stroke-dasharray="32" 
            stroke-dashoffset="24"
            class="opacity-75"
          />
        </svg>
      </div>
      <p v-if="message" class="loading-message">{{ message }}</p>
      <p v-else class="loading-message">Loading...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  message?: string
}

withDefaults(defineProps<Props>(), {
  message: ''
})
</script>

<style scoped>
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: rgb(148, 163, 184);
}

.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.spinner-icon {
  width: 24px;
  height: 24px;
  color: rgb(59, 130, 246);
}

.loading-message {
  font-size: 0.875rem;
  text-align: center;
  margin: 0;
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>