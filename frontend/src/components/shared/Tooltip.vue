<template>
  <div class="relative group">
    <slot></slot>
    <div
      class="absolute bottom-full mb-2 px-3 py-1.5 text-xs font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none w-max max-w-xs"
      :class="positionClass"
    >
      {{ text }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  text: string;
  position?: "top" | "top-start" | "top-end";
}>();

const positionClass = computed(() => {
  switch (props.position) {
    case "top-start":
      return "left-0";
    case "top-end":
      return "right-0";
    case "top":
    default:
      return "left-1/2 -translate-x-1/2";
  }
});
</script>
