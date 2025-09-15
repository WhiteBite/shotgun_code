<template>
  <div class="single-pane-layout">
    <slot />
  </div>
</template>

<script setup lang="ts">
// Single pane layout component
</script>

<style scoped>
.single-pane-layout {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>