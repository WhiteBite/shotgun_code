<template>
  <div class="flex flex-col h-screen bg-gray-900 text-white font-sans overflow-hidden">
    <CommandBar />

    <main class="flex-1 overflow-hidden relative">
      <div
          v-if="uiStore.activeDrawer"
          @click="uiStore.closeDrawer()"
          class="absolute inset-0 bg-black/40 z-20"
          aria-hidden="true"
      ></div>

      <slot></slot>

      <SettingsDrawer />
      <IgnoreRulesDrawer />
      <PromptsDrawer />
    </main>

    <BottomConsole v-if="uiStore.isConsoleVisible" :height="200" />
    <ToastNotifications />
  </div>
</template>

<script setup lang="ts">
import { useUiStore } from '@/stores/ui.store';
import CommandBar from '../global/CommandBar.vue';
import SettingsDrawer from '../drawers/SettingsDrawer.vue';
import IgnoreRulesDrawer from '../drawers/IgnoreRulesDrawer.vue';
import PromptsDrawer from '../drawers/PromptsDrawer.vue';
import ToastNotifications from '../global/ToastNotifications.vue';
import BottomConsole from '../BottomConsole.vue';

const uiStore = useUiStore();
</script>