<template>
  <div class="flex flex-col h-full" ref="rootRef">
    <div :style="{ height: topHeight + 'px' }" class="min-h-[160px] overflow-hidden">
      <slot name="top" />
    </div>
    <div
        class="h-2 cursor-row-resize bg-gray-700 hover:bg-gray-600 flex items-center justify-center select-none flex-shrink-0"
        @mousedown.prevent="startDrag"
        title="Drag to resize"
    >
      <div class="w-24 h-0.5 bg-gray-400 rounded"></div>
    </div>
    <div class="flex-1 min-h-0 overflow-hidden">
      <slot name="bottom" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";

const props = withDefaults(defineProps<{ initialTop?: number; minTop?: number; minBottom?: number; storageKey?: string }>(), {
  initialTop: 320,
  minTop: 160,
  minBottom: 160,
  storageKey: undefined,
});

const rootRef = ref<HTMLElement | null>(null);
const topHeight = ref(props.initialTop);
let dragging = false;
let startY = 0;
let startHeight = 0;

// Функции обработчиков событий
let onMouseMove: ((e: MouseEvent) => void) | null = null;
let onMouseUp: (() => void) | null = null;

onMounted(() => {
  if (props.storageKey) {
    try {
      const raw = localStorage.getItem(props.storageKey);
      const v = raw ? parseInt(raw, 10) : NaN;
      if (!Number.isNaN(v) && v > 0) topHeight.value = v;
    } catch { /* ignore */ }
  }
});

// Очистка listeners при unmount
onUnmounted(() => {
  if (onMouseMove) {
    document.removeEventListener("mousemove", onMouseMove);
    onMouseMove = null;
  }
  if (onMouseUp) {
    document.removeEventListener("mouseup", onMouseUp);
    onMouseUp = null;
  }
});

function startDrag(e: MouseEvent) {
  if (!rootRef.value) return;

  dragging = true;
  startY = e.clientY;
  startHeight = topHeight.value;

  // Создаем новые функции для этой сессии drag
  onMouseMove = (e: MouseEvent) => {
    if (!dragging || !rootRef.value) return;
    const total = rootRef.value.clientHeight || 0;
    const dy = e.clientY - startY;
    const minTop = props.minTop;
    const minBottom = props.minBottom;
    const maxTop = Math.max(minTop, total - minBottom);
    const next = Math.max(minTop, Math.min(maxTop, startHeight + dy));
    topHeight.value = next;
  };

  onMouseUp = () => {
    dragging = false;

    // Удаляем listeners
    if (onMouseMove) {
      document.removeEventListener("mousemove", onMouseMove);
      onMouseMove = null;
    }
    if (onMouseUp) {
      document.removeEventListener("mouseup", onMouseUp);
      onMouseUp = null;
    }

    // Сохраняем в localStorage
    if (props.storageKey) {
      try {
        localStorage.setItem(props.storageKey, String(topHeight.value));
      } catch { /* ignore */ }
    }
  };

  // Добавляем listeners
  document.addEventListener("mousemove", onMouseMove);
  document.addEventListener("mouseup", onMouseUp);
}
</script>
