<template>
  <div v-if="uiStore.quickLook.visible"
       class="fixed inset-0 z-40"
       :class="{'bg-black/50': isPinned}"
       @click.self="hideIfUnpinned"
  >
    <transition
        enter-active-class="transition ease-out duration-200"
        enter-from-class="transform opacity-0 scale-95"
        enter-to-class="transform opacity-100 scale-100"
        leave-active-class="transition ease-in duration-150"
        leave-from-class="transform opacity-100 scale-100"
        leave-to-class="transform opacity-0 scale-95"
    >
      <div
          class="fixed z-50 flex flex-col bg-gray-800/95 backdrop-blur-sm shadow-2xl rounded-lg border border-gray-600 overflow-hidden"
          :style="quickLookStyle"
      >
        <div class="flex-shrink-0 p-2 border-b border-gray-700 bg-gray-800/80 flex justify-between items-center">
          <span class="text-xs text-gray-400 font-mono truncate">{{ uiStore.quickLook.filePath }}</span>
          <button @click="uiStore.hideQuickLook()" class="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-white transition-colors" title="Close (Esc)">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>
        <div class="flex-grow p-3 overflow-auto">
          <div v-if="uiStore.quickLook.isLoading" class="flex items-center justify-center h-32 text-gray-400">
            Loading...
          </div>
          <div v-else-if="uiStore.quickLook.error" class="flex items-center justify-center h-32 text-red-400 p-4 text-center text-sm">
            {{ uiStore.quickLook.error }}
          </div>
          <pre v-else class="text-sm leading-relaxed"><code class="hljs" :class="`language-${uiStore.quickLook.language}`" v-html="uiStore.quickLook.content"></code></pre>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useUiStore } from '@/stores/ui.store';

const uiStore = useUiStore();
const isPinned = computed(() => uiStore.quickLook.isPinned);

const quickLookStyle = computed(() => {
  const base = {
    width: '50vw',
    maxWidth: '900px',
    minWidth: '500px',
    maxHeight: '70vh',
    minHeight: '300px',
  };
  if (isPinned.value) {
    return { ...base, top: '15vh', left: '25vw' };
  }
  return {
    ...base,
    top: `${Math.min(uiStore.quickLook.y, window.innerHeight - 400)}px`,
    left: `${Math.min(uiStore.quickLook.x, window.innerWidth - 600)}px`,
  };
});

function hideIfUnpinned() {
  if (!isPinned.value) {
    uiStore.hideQuickLook();
  }
}
</script>