<template>
  <div
    v-if="uiStore.quickLook.visible"
    class="fixed inset-0 z-40"
    :class="{
      'bg-black/50 pointer-events-auto': isPinned,
      'bg-transparent pointer-events-none': !isPinned,
    }"
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
        v-if="uiStore.quickLook.visible"
        class="fixed z-50 flex flex-col bg-gray-800/95 backdrop-blur-sm shadow-2xl rounded-lg border border-gray-600 overflow-hidden pointer-events-auto"
        :style="panelStyle"
      >
        <div
          class="flex-shrink-0 p-2 border-b border-gray-700 bg-gray-800/80 flex justify-between items-center"
        >
          <span class="text-xs text-gray-400 font-mono truncate max-w-[70vw]">{{
            uiStore.quickLook.filePath
          }}</span>
          <div class="flex items-center gap-1">
            <button
              @click="uiStore.togglePinQuickLook()"
              class="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-white transition-colors"
              :title="isPinned ? 'Unpin' : 'Pin'"
            >
              <span v-if="!isPinned">📌</span>
              <span v-else>📍</span>
            </button>
            <button
              @click="uiStore.hideQuickLook()"
              class="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-white transition-colors"
              title="Close (Esc)"
            >
              ✖
            </button>
          </div>
        </div>

        <div class="flex-grow p-3 overflow-auto">
          <div
            v-if="uiStore.quickLook.isLoading"
            class="flex items-center justify-center h-32 text-gray-400"
          >
            Loading...
          </div>

          <div
            v-else-if="uiStore.quickLook.error"
            class="flex items-center justify-center h-32 text-red-400 p-4 text-center text-sm"
          >
            {{ uiStore.quickLook.error }}
          </div>

          <div v-else>
            <div
              v-if="uiStore.quickLook.truncated"
              class="mb-2 text-xs text-amber-300 bg-amber-900/30 border border-amber-700/40 px-2 py-1 rounded"
            >
              Preview truncated for performance.
            </div>
            <pre
              class="text-sm leading-relaxed"
            ><code class="hljs" :class="`language-${uiStore.quickLook.language}`" v-html="uiStore.quickLook.content"></code></pre>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useUiStore } from "@/stores/ui.store";

const uiStore = useUiStore();
const isPinned = computed(() => uiStore.quickLook.isPinned);

const panelStyle = computed(() => {
  const base: Record<string, string> = {
    width: "600px",
    maxWidth: "70vw",
    maxHeight: "70vh",
  };

  const safeTop = (y: number) =>
    `${Math.max(8, Math.min(y, window.innerHeight - 160))}px`;
  const safeLeft = (x: number) =>
    `${Math.max(8, Math.min(x, window.innerWidth - 620))}px`;

  const top = safeTop(uiStore.quickLook.y || 80);
  const left = safeLeft(uiStore.quickLook.x || window.innerWidth / 2 - 300);

  return { ...base, top, left };
});

function hideIfUnpinned() {
  if (!isPinned.value) {
    uiStore.hideQuickLook();
  }
}
</script>
