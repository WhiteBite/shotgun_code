<template>
  <div class="relative">
    <button
      ref="triggerRef"
      type="button"
      class="px-2 py-1 text-xs bg-gray-800 hover:bg-gray-700 rounded-md border border-gray-700"
      :aria-expanded="open ? 'true' : 'false'"
      :aria-haspopup="'menu'"
      :disabled="disabled"
      @click="toggle"
    >
      {{ labels.copy }} ▾
    </button>

    <div
      v-if="open"
      ref="popoverRef"
      class="z-50 min-w-[220px] bg-gray-800 border border-gray-700 rounded shadow-lg p-2"
      role="menu"
      :style="floatingStyle"
    >
      <div class="text-[11px] text-gray-400 mb-1">{{ labels.format }}</div>
      <div class="flex gap-1 mb-2">
        <button v-for="fmt in formats" :key="fmt"
          type="button"
          class="px-2 py-0.5 text-[12px] rounded border"
          :class="exportFormat === fmt ? 'bg-blue-600 text-white border-blue-600' : 'bg-gray-700 text-gray-200 border-gray-600'"
          @click="updateFormat(fmt)"
        >
          {{ fmt }}
        </button>
      </div>

      <label class="flex items-center gap-2 text-xs mb-2">
        <input type="checkbox" class="form-checkbox" :checked="stripComments" @change="toggleStrip" />
        <span>{{ labels.stripComments }}</span>
      </label>

      <div class="h-px bg-gray-700 my-2" />

      <div class="flex flex-col gap-1">
        <button
          type="button"
          class="px-2 py-1 text-xs bg-blue-600 hover:bg-blue-500 rounded text-white"
          @click="doCopyAll"
        >
          {{ labels.copyAll }}
        </button>

        <template v-if="splitEnabled && splitPreview && splitPreview.chunks.length > 1">
          <div class="text-[11px] text-gray-400 mt-1">{{ labels.copyPart }}</div>
          <button
            v-for="c in splitPreview.chunks"
            :key="c.index"
            type="button"
            class="w-full text-left px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 rounded"
            @click="doCopyChunk(c.index)"
          >
            Part {{ c.index + 1 }} / {{ splitPreview.chunks.length }} • ~{{ c.tokens }}t
          </button>
        </template>
      </div>

      <div class="h-px bg-gray-700 my-2" />
      <button type="button" class="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 rounded" @click="emit('open-export')">
        Export options…
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from 'vue';
import { computePosition, offset, flip, shift, autoUpdate } from '@floating-ui/dom';
import type { ClipboardFormat, SplitPreview } from '@/types/splitter';

const props = defineProps<{
  exportFormat: ClipboardFormat;
  stripComments: boolean;
  splitEnabled: boolean;
  splitPreview: SplitPreview | null;
  disabled?: boolean;
  i18n?: { copy: string; copyAll: string; copyPart: string; format: string; stripComments: string; };
}>();

const emit = defineEmits<{
  (e: 'update:exportFormat', v: ClipboardFormat): void;
  (e: 'update:stripComments', v: boolean): void;
  (e: 'do-copy', payload: { target: 'all' | 'chunk'; chunkIndex?: number; format: ClipboardFormat; stripComments: boolean }): void;
  (e: 'open-export'): void;
}>();

const open = ref(false);
const formats: ClipboardFormat[] = ['plain','manifest','json'];

const labels = computed(() => ({
  copy: props.i18n?.copy ?? 'Copy',
  copyAll: props.i18n?.copyAll ?? 'Copy all',
  copyPart: props.i18n?.copyPart ?? 'Copy part',
  format: props.i18n?.format ?? 'Format',
  stripComments: props.i18n?.stripComments ?? 'Strip comments',
}));

// Floating UI setup
const triggerRef = ref<HTMLElement|null>(null);
const popoverRef = ref<HTMLElement|null>(null);
const floatingStyle = ref<{ left: string; top: string; position: 'fixed' }>({ left: '0px', top: '0px', position: 'fixed' });
let cleanup: (() => void) | null = null;

async function updatePosition() {
  if (!triggerRef.value || !popoverRef.value) return;
  const { x, y } = await computePosition(triggerRef.value, popoverRef.value, {
    placement: 'bottom-start',
    strategy: 'fixed',
    middleware: [offset(6), flip(), shift({ padding: 8 })],
  });
  floatingStyle.value = { left: `${x}px`, top: `${y}px`, position: 'fixed' };
}

function toggle() {
  open.value = !open.value;
}

watch(open, (v) => {
  if (v) {
    updatePosition();
    if (triggerRef.value && popoverRef.value) {
      cleanup = autoUpdate(triggerRef.value, popoverRef.value, updatePosition);
    }
  } else {
    if (cleanup) { cleanup(); cleanup = null; }
  }
});

onUnmounted(() => { if (cleanup) cleanup(); });

function updateFormat(f: ClipboardFormat) { emit('update:exportFormat', f); }
function toggleStrip(e: Event) {
  const t = e.target as HTMLInputElement;
  emit('update:stripComments', !!t.checked);
}
function doCopyAll() {
  emit('do-copy', { target: 'all', format: props.exportFormat, stripComments: props.stripComments });
  open.value = false;
}
function doCopyChunk(index: number) {
  emit('do-copy', { target: 'chunk', chunkIndex: index, format: props.exportFormat, stripComments: props.stripComments });
  open.value = false;
}
</script>
