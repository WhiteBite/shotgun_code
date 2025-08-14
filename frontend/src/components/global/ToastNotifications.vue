<template>
  <div
    aria-live="assertive"
    class="fixed bottom-4 left-1/2 -translate-x-1/2 flex flex-col items-center space-y-3 w-full max-w-md z-50"
  >
    <transition-group
      tag="div"
      enter-active-class="transform ease-out duration-300 transition"
      enter-from-class="opacity-0 translate-y-4"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <Toast
        v-for="toast in uiStore.toasts"
        :key="toast.id"
        :toast="toast"
        :duration="5000"
        @close="uiStore.removeToast(toast.id)"
      />
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { useUiStore } from "@/stores/ui.store";
import Toast from "./Toast.vue";

const uiStore = useUiStore();
</script>

<style scoped>
div[aria-live="assertive"] {
  pointer-events: none;
}
div[aria-live="assertive"] > div {
  pointer-events: all;
}
</style>
