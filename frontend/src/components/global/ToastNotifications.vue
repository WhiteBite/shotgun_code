<template>
  <div aria-live="assertive" class="fixed top-20 right-6 z-50 flex flex-col items-end space-y-3 w-full max-w-sm">
    <!-- Этот контейнер теперь имеет фиксированную позицию и ширину, и не будет растягиваться по вертикали. -->
    <transition-group
        tag="div"
        enter-active-class="transform ease-out duration-300 transition"
        enter-from-class="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2"
        enter-to-class="translate-y-0 opacity-100 sm:translate-x-0"
        leave-active-class="transition ease-in duration-100"
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
import { useUiStore } from '@/stores/ui.store';
import Toast from './Toast.vue';

const uiStore = useUiStore();
</script>

<style scoped>
/* Убираем pointer-events, так как дочерние элементы уже имеют pointer-events-auto через Tailwind */
div[aria-live="assertive"] {
  pointer-events: none;
}
div[aria-live="assertive"] > div {
  pointer-events: all;
}
</style>