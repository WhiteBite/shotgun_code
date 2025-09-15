<template>
  <div class="grid-layout">
    <slot />
  </div>
</template>

<script setup lang="ts">
// Grid layout component
</script>

<style scoped>
.grid-layout {
  height: 100%;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 1px;
  background: rgba(148, 163, 184, 0.1);
}
</style>