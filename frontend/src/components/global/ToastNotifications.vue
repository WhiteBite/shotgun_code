<template>
  <div aria-live="assertive" class="fixed inset-0 flex items-end px-4 py-6 pointer-events-none sm:p-6 sm:items-start z-50">
    <div class="w-full flex flex-col items-center space-y-4 sm:items-end">
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
  </div>
</template>

<script setup lang="ts">
import { useUiStore } from '@/stores/uiStore';
import Toast from './Toast.vue';

const uiStore = useUiStore();
</script>