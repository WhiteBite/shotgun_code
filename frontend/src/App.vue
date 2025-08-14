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
import { watch, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useProjectStore } from "@/stores/project.store";
import MainLayout from "./components/layout/MainLayout.vue";
import {
  attachKeyboardState,
  detachKeyboardState,
} from "@/composables/useKeyboardState";
import {
  attachShortcuts,
  detachShortcuts,
} from "@/composables/useKeyboardShortcuts";

const projectStore = useProjectStore();
const router = useRouter();

// Watch for a project being loaded and handle navigation here.
watch(
  () => projectStore.isProjectLoaded,
  (isLoaded, wasLoaded) => {
    if (isLoaded && !wasLoaded) {
      router.push("/workspace");
    }
  },
);

onMounted(() => {
  attachKeyboardState();
  attachShortcuts();
});

onUnmounted(() => {
  detachShortcuts();
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
