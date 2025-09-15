<template>
  <div class="loading-state">
    <div class="loading-content">
      <div class="loading-spinner">
        <div class="spinner-ring"></div>
        <div class="spinner-ring"></div>
        <div class="spinner-ring"></div>
      </div>
      <p class="loading-text">{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  message?: string
}

withDefaults(defineProps<Props>(), {
  message: 'Loading...'
})
</script>

<style scoped>
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
  min-height: 120px;
}

.loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.loading-spinner {
  position: relative;
  width: 32px;
  height: 32px;
}

.spinner-ring {
  position: absolute;
  width: 32px;
  height: 32px;
  border: 2px solid transparent;
  border-top-color: rgb(168, 85, 247);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.spinner-ring:nth-child(2) {
  width: 24px;
  height: 24px;
  top: 4px;
  left: 4px;
  border-top-color: rgb(59, 130, 246);
  animation-duration: 0.8s;
  animation-direction: reverse;
}

.spinner-ring:nth-child(3) {
  width: 16px;
  height: 16px;
  top: 8px;
  left: 8px;
  border-top-color: rgb(34, 197, 94);
  animation-duration: 0.6s;
}

.loading-text {
  color: rgb(148, 163, 184);
  font-size: 0.875rem;
  font-weight: 500;
  margin: 0;
  text-align: center;
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