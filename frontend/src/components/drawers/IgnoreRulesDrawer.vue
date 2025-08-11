<template>
  <transition name="slide-fade">
    <aside v-if="uiStore.activeDrawer === 'ignore'" class="absolute top-0 right-0 h-full w-96 bg-gray-800 border-l border-gray-700 shadow-2xl z-30 p-4 flex flex-col">
      <div class="flex items-center justify-between mb-4 flex-shrink-0">
        <h2 class="text-xl font-semibold text-white">Правила игнорирования</h2>
        <button @click="uiStore.closeDrawer()" class="p-2 rounded-md hover:bg-gray-700">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-gray-400"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
        </button>
      </div>

      <div class="flex-grow flex flex-col min-h-0">
        <p class="text-sm text-gray-400 mb-2 flex-shrink-0">
          Используйте синтаксис `.gitignore`. Правила применяются, когда включена опция "Кастомные правила".
        </p>
        <textarea
            v-model="settingsStore.settings.customIgnoreRules"
            class="w-full flex-grow bg-gray-900 border border-gray-600 rounded-md p-3 font-mono text-sm resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="node_modules/&#10;*.log&#10;dist/"
        ></textarea>
        <div class="mt-4 flex-shrink-0">
          <button @click="settingsStore.saveSettings()" class="w-full py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50" :disabled="settingsStore.isLoading">
            {{ settingsStore.isLoading ? 'Сохранение...' : 'Сохранить' }}
          </button>
        </div>
      </div>
    </aside>
  </transition>
</template>

<script setup lang="ts">
import { useUiStore } from '@/stores/uiStore';
import { useSettingsStore } from '@/stores/settingsStore';
const uiStore = useUiStore();
const settingsStore = useSettingsStore();
</script>

<style scoped>
.slide-fade-enter-active, .slide-fade-leave-active {
  transition: transform 0.22s cubic-bezier(0.165, 0.84, 0.44, 1);
}
.slide-fade-enter-from, .slide-fade-leave-to {
  transform: translateX(100%);
}
</style>