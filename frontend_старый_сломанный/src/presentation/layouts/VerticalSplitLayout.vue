<template>
  <div class="vertical-split-layout">
    <slot />
  </div>
</template>

<script setup lang="ts">
// Vertical split layout component
</script>

<style scoped>
.vertical-split-layout {
  height: 100%;
  display: flex;
  flex-direction: row;
  gap: 1px;
  background: rgba(148, 163, 184, 0.1);
}
</style>