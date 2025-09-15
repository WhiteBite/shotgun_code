<template>
  <div class="horizontal-split-layout">
    <slot />
  </div>
</template>

<script setup lang="ts">
// Horizontal split layout component
</script>

<style scoped>
.horizontal-split-layout {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: rgba(148, 163, 184, 0.1);
}
</style>