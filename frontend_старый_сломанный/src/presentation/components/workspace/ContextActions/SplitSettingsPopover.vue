<template>
  <div class="relative">
    <button
      ref="triggerRef"
      type="button"
      class="px-2 py-1 text-xs bg-gray-800 hover:bg-gray-700 rounded-md border border-gray-700"
      :disabled="disabled"
      aria-haspopup="dialog"
      :aria-expanded="open ? 'true' : 'false'"
      @click="toggle"
    >
      ⫿ Split
    </button>

    <div
      v-if="open"
      ref="popoverRef"
      class="z-50 w-80 bg-gray-800 border border-gray-700 rounded shadow-lg p-3"
      role="dialog"
      aria-modal="true"
      :style="floatingStyle"
    >
      <div class="text-sm font-semibold text-white mb-2">{{ t.title }}</div>

      <label class="flex items-center gap-2 text-sm mb-2">
        <input
          v-model="local.enableAutoSplit"
          type="checkbox"
          class="form-checkbox"
        />
        <span>{{ t.enableAutoSplit }}</span>
      </label>

      <div
        class="grid grid-cols-2 gap-2 mb-2"
        :class="{ 'opacity-50 pointer-events-none': !local.enableAutoSplit }"
      >
        <div>
          <div class="text-[11px] text-gray-400 mb-1">{{ t.maxTokens }}</div>
          <input
            v-model.number="local.maxTokensPerChunk"
            type="number"
            min="100"
            class="w-full bg-gray-700 border-gray-600 rounded px-2 py-1 text-xs"
          />
        </div>
        <div>
          <div class="text-[11px] text-gray-400 mb-1">{{ t.overlap }}</div>
          <input
            v-model.number="local.overlapTokens"
            type="number"
            min="0"
            class="w-full bg-gray-700 border-gray-600 rounded px-2 py-1 text-xs"
          />
        </div>
      </div>

      <div
        :class="{ 'opacity-50 pointer-events-none': !local.enableAutoSplit }"
      >
        <div class="text-[11px] text-gray-400 mb-1">{{ t.strategy }}</div>
        <select
          v-model="local.splitStrategy"
          class="w-full bg-gray-700 border-gray-600 rounded px-2 py-1 text-xs"
        >
          <option value="smart">{{ t.strategySmart }}</option>
          <option value="file">{{ t.strategyFile }}</option>
          <option value="token">{{ t.strategyToken }}</option>
        </select>
      </div>

      <!-- Улучшенный preview с детальной информацией -->
      <div
        class="text-[11px] text-gray-400 mt-2 min-h-[2rem] p-2 bg-gray-900/50 rounded border border-gray-600"
      >
        <template v-if="preview && local.enableAutoSplit">
          <div class="font-medium text-gray-300 mb-1">Preview:</div>
          <div>{{ t.willSplitInto(preview.chunkCount) }}</div>
          <div class="mt-1">
            Total: {{ preview.totalTokens.toLocaleString() }} tokens,
            {{ preview.totalChars.toLocaleString() }} chars
          </div>
          <div v-if="preview.chunks.length > 0" class="mt-1">
            Avg per chunk: ~{{
              Math.round(
                preview.totalTokens / preview.chunkCount,
              ).toLocaleString()
            }}
            tokens
          </div>
          <div
            v-if="preview.warnings && preview.warnings.length > 0"
            class="mt-1 text-yellow-400"
          >
            ⚠️ {{ preview.warnings.length }} warning(s)
          </div>
        </template>
        <template v-else-if="!local.enableAutoSplit">
          <div class="text-gray-500">Auto-split disabled</div>
        </template>
        <template v-else>
          <div class="text-gray-500">Build context to see preview</div>
        </template>
      </div>

      <div class="flex justify-end gap-2 mt-2">
        <button
          type="button"
          class="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 rounded"
          @click="emit('refresh')"
        >
          Refresh
        </button>
        <button
          type="button"
          class="px-2 py-1 text-xs bg-blue-600 hover:bg-blue-500 rounded text-white"
          @click="apply"
        >
          Apply
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, computed, onUnmounted } from "vue";
import type { SplitSettings, SplitPreview } from "@/types/splitter";
import {
  computePosition,
  offset,
  flip,
  shift,
  autoUpdate,
} from "@floating-ui/dom";

const props = defineProps<{
  modelValue: SplitSettings;
  preview: SplitPreview | null;
  disabled?: boolean;
  i18n?: {
    title: string;
    enableAutoSplit: string;
    maxTokens: string;
    overlap: string;
    strategy: string;
    strategySmart: string;
    strategyFile: string;
    strategyToken: string;
    willSplitInto: (n: number) => string;
  };
}>();
const emit = defineEmits<{
  (e: "update:modelValue", v: SplitSettings): void;
  (e: "apply"): void;
  (e: "refresh"): void;
}>();

const t = computed(() => ({
  title: props.i18n?.title ?? "Split settings",
  enableAutoSplit: props.i18n?.enableAutoSplit ?? "Auto-split large contexts",
  maxTokens: props.i18n?.maxTokens ?? "Max tokens per chunk",
  overlap: props.i18n?.overlap ?? "Overlap tokens",
  strategy: props.i18n?.strategy ?? "Strategy",
  strategySmart: props.i18n?.strategySmart ?? "Smart (recommended)",
  strategyFile: props.i18n?.strategyFile ?? "By files",
  strategyToken: props.i18n?.strategyToken ?? "By tokens",
  willSplitInto:
    props.i18n?.willSplitInto ?? ((n: number) => `Will split into ${n} parts`),
}));

const local = reactive<SplitSettings>({ ...props.modelValue });
const open = ref(false);

watch(
  () => props.modelValue,
  (v) => {
    Object.assign(local, v);
  },
  { deep: true },
);

function apply() {
  emit("update:modelValue", { ...local });
  emit("apply");
  open.value = false;
}

function toggle() {
  open.value = !open.value;
}

const triggerRef = ref<HTMLElement | null>(null);
const popoverRef = ref<HTMLElement | null>(null);
const floatingStyle = ref<{ left: string; top: string; position: "fixed" }>({
  left: "0px",
  top: "0px",
  position: "fixed",
});
let cleanup: (() => void) | null = null;

async function updatePosition() {
  if (!triggerRef.value || !popoverRef.value) return;
  const { x, y } = await computePosition(triggerRef.value, popoverRef.value, {
    placement: "bottom-start",
    strategy: "fixed",
    middleware: [offset(6), flip(), shift({ padding: 8 })],
  });
  floatingStyle.value = { left: `${x}px`, top: `${y}px`, position: "fixed" };
}

watch(open, (v) => {
  if (v) {
    updatePosition();
    if (triggerRef.value && popoverRef.value) {
      cleanup = autoUpdate(triggerRef.value, popoverRef.value, updatePosition);
    }
  } else {
    if (cleanup) {
      cleanup();
      cleanup = null;
    }
  }
});

onUnmounted(() => {
  if (cleanup) cleanup();
});
</script>
