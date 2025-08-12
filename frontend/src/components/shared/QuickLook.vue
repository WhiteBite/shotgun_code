<template>
  <transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="transform opacity-0 scale-95"
      enter-to-class="transform opacity-100 scale-100"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="transform opacity-100 scale-100"
      leave-to-class="transform opacity-0 scale-95"
  >
    <div
        v-if="uiStore.quickLook.visible"
        class="fixed z-50 flex flex-col bg-gray-800/95 backdrop-blur-sm shadow-2xl rounded-lg border border-gray-600 overflow-hidden"
        :style="quickLookStyle"
        @mouseenter="uiStore.keepQuickLookOpen()"
        @mouseleave="uiStore.requestHideQuickLook()"
    >
      <div class="flex-shrink-0 p-2 border-b border-gray-700 bg-gray-800/80 flex justify-between items-center cursor-move">
        <span class="text-xs text-gray-400 font-mono truncate">{{ uiStore.quickLook.filePath }}</span>
        <div class="flex items-center gap-2">
          <button @click="uiStore.toggleQuickLookPin()" class="p-1 rounded hover:bg-gray-700 transition-colors" :title="isPinned ? 'Unpin' : 'Pin'">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="{ 'text-blue-400': isPinned, 'text-gray-400': !isPinned }"><path d="M12 17v5m-4.5-5.23a7.5 7.5 0 0111.46.24l-2.23 2.23a1.5 1.5 0 01-2.12 0l-1.88-1.88a1.5 1.5 0 00-2.12 0l-1.88 1.88a1.5 1.5 0 01-2.12 0l-2.23-2.23a7.5 7.5 0 012.46-2.46zM12 2a5 5 0 015 5c0 5-5 9-5 9s-5-4-5-9a5 5 0 015-5z"></path></svg>
          </button>
          <button @click="uiStore.hideQuickLook()" class="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-white transition-colors" title="Close">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>
      </div>
      <div class="flex-grow overflow-auto" style="max-height: 60vh;">
        <div v-if="uiStore.quickLook.isLoading" class="flex items-center justify-center h-32 text-gray-400">
          <div class="flex items-center gap-2">
            <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path></svg>
            Loading...
          </div>
        </div>
        <div v-else-if="uiStore.quickLook.error" class="flex items-center justify-center h-32 text-red-400 p-4 text-center text-sm">
          {{ uiStore.quickLook.error }}
        </div>
        <div v-else class="p-4">
          <pre class="text-sm leading-relaxed"><code class="hljs" :class="`language-${uiStore.quickLook.language}`" v-html="uiStore.quickLook.content"></code></pre>
        </div>
      </div>
      <div class="flex-shrink-0 p-2 border-t border-gray-700 bg-gray-800/50 flex justify-between items-center text-xs text-gray-400">
        <span>{{ lineCount }} lines</span>
        <span>{{ fileSize }}</span>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useUiStore } from '@/stores/ui.store';
import { ClipboardSetText } from '../../../wailsjs/runtime/runtime';

const uiStore = useUiStore();
const isPinned = computed(() => uiStore.quickLook.isPinned);

const fileSize = computed(() => {
  const bytes = new Blob([uiStore.quickLook.rawContent]).size;
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
});

const lineCount = computed(() => uiStore.quickLook.rawContent.split('\n').length);

const quickLookStyle = computed(() => {
  const base = {
    width: '50vw',
    maxWidth: '900px',
    minWidth: '500px',
  };
  if (isPinned.value) {
    return { ...base, top: '10vh', left: '25vw', };
  }
  const x = Math.min(uiStore.quickLook.x, window.innerWidth - 600);
  const y = Math.min(uiStore.quickLook.y, window.innerHeight - 400);
  return { ...base, top: `${y}px`, left: `${x}px` };
});
</script>