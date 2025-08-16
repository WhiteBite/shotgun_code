<template>
  <MainLayout>
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </MainLayout>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import MainLayout from "@/components/layout/MainLayout.vue";
import { attachKeyboardState, detachKeyboardState } from "@/composables/useKeyboardState";
import { eventService } from "@/services/event.service";

onMounted(() => {
  attachKeyboardState();
  // Ensure events are subscribed when DOM is ready
  eventService.initialize();
});

onUnmounted(() => {
  detachKeyboardState();
});
</script>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>