<template>
  <div
    v-if="hasContent"
    class="flex items-center gap-2 px-2 py-1 border-b border-gray-700 bg-gray-900/60"
    :aria-disabled="disabled ? 'true' : 'false'"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{ hasContent: boolean; disabled?: boolean }>();
</script>
