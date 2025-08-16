#!/usr/bin/env python3
"""
Автоматический рефакторинг (PATCH): Batch 1/4 — инфраструктура сплита контента, контракты, UI-компоненты панели действий
Сгенерировано: 2025-08-15
Режим: PATCH (пере-запись целевых файлов, идемпотентно)
"""

from pathlib import Path
from datetime import datetime

class ProjectRefactor:
    def __init__(self, dry_run=False, start_from_step=None):
        self.dry_run = dry_run
        self.start_from_step = start_from_step
        self.current_step = 0

    # ---------- helpers ----------
    def log(self, message):
        ts = datetime.now().strftime('%H:%M:%S')
        print(f"[{ts}] {message}")

    def ensure_directory(self, path: str):
        p = Path(path)
        if self.dry_run:
            self.log(f"[DRY-RUN] ensure_directory: {path}")
            return
        p.mkdir(parents=True, exist_ok=True)

    def update_file(self, path: str, content: str):
        if self.dry_run:
            self.log(f"[DRY-RUN] Обновить/создать файл: {path}")
            return
        self.ensure_directory(str(Path(path).parent))
        Path(path).write_text(content, encoding="utf-8", newline="\n")
        self.log(f"✅ Обновлён файл: {path}")

    def should_exec(self, step: int) -> bool:
        return self.start_from_step is None or step >= self.start_from_step

    # ---------- file contents ----------
    def c_types_splitter(self) -> str:
        return """\
/* Auto-generated: types for content splitting and copy actions */
export type ClipboardFormat = "plain" | "manifest" | "json";
export type SplitStrategy = "smart" | "file" | "token";

export interface SplitSettings {
  enableAutoSplit: boolean;
  maxTokensPerChunk: number;
  overlapTokens: number;
  splitStrategy: SplitStrategy;
}

export interface TokenEstimateOptions {
  estimator?: (text: string) => number; // optional precise estimator (tiktoken etc.)
}

export interface Segment {
  id: string;
  title?: string;
  start: number; // inclusive
  end: number;   // exclusive
  isFile?: boolean;
  relPath?: string;
}

export interface SplitInput {
  text: string;
  segments?: Segment[];
}

export interface SplitChunkRef {
  segmentId: string;
  from: number;
  to: number;
}

export interface SplitChunk {
  index: number;
  start: number;
  end: number;
  text: string;
  tokens: number;
  chars: number;
  refs?: SplitChunkRef[];
  note?: string;
}

export interface SplitPreview {
  totalTokens: number;
  totalChars: number;
  chunkCount: number;
  chunks: SplitChunk[];
  warnings: string[];
}

export interface CopyRequest {
  target: "all" | "chunk";
  chunkIndex?: number;
  format: ClipboardFormat;
  stripComments: boolean;
}

export interface CopyPlan {
  request: CopyRequest;
  parts: string[];
}
"""

    def c_service_splitter(self) -> str:
        return """\
/* Auto-generated: content splitter service (framework-agnostic) */
import type {
  SplitInput, SplitSettings, SplitPreview, TokenEstimateOptions,
  Segment, SplitChunk
} from "@/types/splitter";

export interface SplitterService {
  estimateTokens(text: string, opts?: TokenEstimateOptions): number;
  detectSegmentsFromContextText(text: string): Segment[];
  split(input: SplitInput, settings: SplitSettings, opts?: TokenEstimateOptions): SplitPreview;
}

function defaultEstimateTokens(text: string): number {
  // Fallback heuristic; can be replaced by precise estimator (tiktoken etc.)
  if (!text) return 0;
  return Math.max(1, Math.ceil(text.length / 4));
}

function clamp(n: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, n));
}

function makeChunk(text: string, start: number, end: number, estimator: (t: string)=>number): SplitChunk {
  const slice = text.slice(start, end);
  return {
    index: 0, // will be assigned later
    start, end,
    text: slice,
    tokens: estimator(slice),
    chars: slice.length,
  };
}

function packSegmentsByTokens(
  text: string,
  segments: Segment[],
  limitTokens: number,
  estimator: (t: string)=>number,
  overlapTokens: number
): { chunks: SplitChunk[]; warnings: string[] } {
  const chunks: SplitChunk[] = [];
  const warnings: string[] = [];
  let currentStart = segments.length ? segments[0].start : 0;
  let accTokens = 0;
  let accStart = currentStart;
  let lastEnd = currentStart;

  const avgCharsPerToken = (text.length || 1) / Math.max(1, estimator(text));
  const overlapChars = Math.floor(overlapTokens * avgCharsPerToken);

  const pushChunk = (from: number, to: number) => {
    const s = clamp(from, 0, text.length);
    const e = clamp(to, s, text.length);
    const chunk = makeChunk(text, s, e, estimator);
    chunks.push(chunk);
  };

  for (const seg of segments) {
    const segText = text.slice(seg.start, seg.end);
    const segTokens = estimator(segText);
    if (segTokens <= limitTokens) {
      if (accTokens + segTokens > limitTokens && lastEnd > accStart) {
        pushChunk(accStart, lastEnd);
        accStart = Math.max(accStart, lastEnd - overlapChars);
        accTokens = estimator(text.slice(accStart, accStart)); // 0
      }
      if (accStart === 0 && chunks.length === 0) {
        accStart = seg.start;
      } else if (lastEnd === 0) {
        accStart = seg.start;
      }
      lastEnd = seg.end;
      accTokens += segTokens;
    } else {
      if (lastEnd > accStart) {
        pushChunk(accStart, lastEnd);
        accStart = Math.max(accStart, lastEnd - overlapChars);
      }
      const charsPerToken = avgCharsPerToken || 4;
      const stepChars = Math.max(1, Math.floor(limitTokens * charsPerToken));
      let pos = seg.start;
      while (pos < seg.end) {
        const endPos = Math.min(seg.end, pos + stepChars);
        pushChunk(pos, endPos);
        pos = endPos - overlapChars;
        if (pos <= seg.start) pos = endPos; // guard
      }
      warnings.push(`Segment too large, split by token within: ${seg.title || seg.relPath || seg.id}`);
      accTokens = 0;
      lastEnd = 0;
      accStart = segments.find(s => s.start >= seg.end)?.start ?? seg.end;
    }
  }

  if (lastEnd > accStart) {
    pushChunk(accStart, lastEnd);
  }

  chunks.forEach((c, i) => c.index = i);
  return { chunks, warnings };
}

export function createSplitterService(): SplitterService {
  return {
    estimateTokens(text: string, opts?: TokenEstimateOptions): number {
      if (opts?.estimator) return Math.max(0, opts.estimator(text));
      return defaultEstimateTokens(text);
    },

    detectSegmentsFromContextText(text: string): Segment[] {
      if (!text) return [];
      const lines = text.split(/\\r?\\n/);
      const segments: Segment[] = [];
      let currentStart = 0;
      let currentTitle = "Context";
      let currentRel = "";

      const push = (start: number, end: number, title: string, rel?: string) => {
        if (end <= start) return;
        segments.push({
          id: `${start}:${end}`,
          title,
          start, end,
          isFile: !!rel,
          relPath: rel,
        });
      };

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        const m1 = /^File:\\s+(.+)$/.exec(line);
        const m2 = /^\\+\\+\\+\\s.+?\\s(.+)$/.exec(line);
        const m3 = /^\\s*[-=]{3,}\\s*(.+?)\\s*[-=]{3,}\\s*$/.exec(line);
        if (m1 || m2 || m3) {
          const prevEnd = text.split(/\\r?\\n/).slice(0, i).join("\\n").length;
          push(currentStart, prevEnd, currentTitle, currentRel || undefined);
          currentStart = prevEnd + (i > 0 ? 1 : 0);
          currentTitle = "File";
          currentRel = (m1?.[1] || m2?.[1] || m3?.[1] || "").trim();
        }
      }
      push(currentStart, text.length, currentTitle, currentRel || undefined);

      if (segments.length === 0) {
        return [{ id: "0:len", title: "Context", start: 0, end: text.length }];
      }
      return segments.filter(s => s.end > s.start).sort((a, b) => a.start - b.start);
    },

    split(input: SplitInput, settings: SplitSettings, opts?: TokenEstimateOptions): SplitPreview {
      const text = input.text || "";
      const estimator = (t: string) => (opts?.estimator ? opts.estimator(t) : defaultEstimateTokens(t));
      const totalTokens = estimator(text);
      const totalChars = text.length;

      if (!settings.enableAutoSplit || !text) {
        const chunk: SplitChunk = {
          index: 0, start: 0, end: text.length, text,
          tokens: totalTokens, chars: totalChars,
        };
        return { totalTokens, totalChars, chunkCount: 1, chunks: [chunk], warnings: [] };
      }

      const limit = Math.max(1, settings.maxTokensPerChunk);
      const overlap = Math.max(0, settings.overlapTokens || 0);
      const strategy = settings.splitStrategy || "smart";

      let segments: Segment[] | undefined = input.segments;
      if ((!segments || !segments.length) && (strategy === "file" || strategy === "smart")) {
        segments = this.detectSegmentsFromContextText(text);
      }

      if ((strategy === "file" || strategy === "smart") && segments && segments.length) {
        const { chunks, warnings } = packSegmentsByTokens(text, segments, limit, estimator, overlap);
        return { totalTokens, totalChars, chunkCount: chunks.length, chunks, warnings };
      }

      const avgCharsPerToken = totalTokens > 0 ? totalChars / totalTokens : 4;
      const stepChars = Math.max(1, Math.floor(limit * avgCharsPerToken));
      const overlapChars = Math.floor(overlap * avgCharsPerToken);
      const chunks: SplitChunk[] = [];
      let i = 0;
      while (i < text.length) {
        const end = Math.min(text.length, i + stepChars);
        chunks.push(makeChunk(text, i, end, estimator));
        i = end - overlapChars;
        if (i <= 0) i = end;
        if (i <= 0) break;
      }
      chunks.forEach((c, idx) => (c.index = idx));
      return { totalTokens, totalChars, chunkCount: chunks.length, chunks, warnings: [] };
    },
  };
}
"""

    def c_use_context_actions(self) -> str:
        return """\
// Auto-generated: composable for context actions (copy/split)
import { shallowRef, watch, computed } from "vue";
import type { ClipboardFormat, CopyRequest, SplitSettings, SplitPreview } from "@/types/splitter";
import { createSplitterService } from "@/services/splitter.service";
import { useExportStore } from "@/stores/export.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useUiStore } from "@/stores/ui.store";
import { apiService } from "@/services/api.service";

const splitter = createSplitterService();

export function useContextActions() {
  const exportStore = useExportStore();
  const ctxStore = useContextBuilderStore();
  const ui = useUiStore();

  const exportFormat = computed<ClipboardFormat>({
    get: () => exportStore.exportFormat,
    set: (v) => (exportStore.exportFormat = v),
  });
  const stripComments = computed<boolean>({
    get: () => exportStore.stripComments,
    set: (v) => (exportStore.stripComments = v),
  });
  const splitSettings = computed<SplitSettings>({
    get: () => ({
      enableAutoSplit: exportStore.enableAutoSplit,
      maxTokensPerChunk: exportStore.maxTokensPerChunk,
      overlapTokens: exportStore.overlapTokens,
      splitStrategy: exportStore.splitStrategy,
    }),
    set: (v) => {
      exportStore.enableAutoSplit = v.enableAutoSplit;
      exportStore.maxTokensPerChunk = v.maxTokensPerChunk;
      exportStore.overlapTokens = v.overlapTokens;
      exportStore.splitStrategy = v.splitStrategy;
    },
  });

  const splitPreview = shallowRef<SplitPreview | null>(null);
  let debounceId: number | null = null;

  async function computePreviewNow() {
    const text = ctxStore.shotgunContextText || "";
    if (!text) { splitPreview.value = null; return; }
    try {
      const preview = splitter.split({ text }, splitSettings.value);
      splitPreview.value = preview;
    } catch (e: any) {
      ui.addToast(`Split preview failed: ${e?.message || e}`, "error");
      splitPreview.value = null;
    }
  }

  function refreshPreview() {
    if (debounceId) window.clearTimeout(debounceId);
    debounceId = window.setTimeout(() => { computePreviewNow(); debounceId = null; }, 200);
  }

  watch(() => [ctxStore.shotgunContextText, splitSettings.value], refreshPreview, { deep: true });

  async function copy(req: CopyRequest) {
    const text = ctxStore.shotgunContextText || "";
    if (!text) { ui.addToast("No context to copy", "error"); return; }

    let payload = text;
    if (splitSettings.value.enableAutoSplit && splitPreview.value && splitPreview.value.chunks.length > 1) {
      if (req.target === "chunk" && typeof req.chunkIndex === "number") {
        const c = splitPreview.value.chunks[req.chunkIndex];
        payload = c ? c.text : text;
      } else {
        const parts = splitPreview.value.chunks.map((c, i, arr) => `=== Part ${i+1}/${arr.length} ===\\n${c.text}`);
        payload = parts.join("\\n\\n");
      }
    }

    try {
      const result = await apiService.exportContext({
        mode: "clipboard",
        context: payload,
        exportFormat: req.format,
        stripComments: req.stripComments,
        includeManifest: req.format === "manifest",
      });
      if (result?.text) {
        await navigator.clipboard.writeText(result.text);
      } else {
        await navigator.clipboard.writeText(payload);
      }
      ui.addToast("Copied to clipboard", "success");
    } catch (e: any) {
      ui.addToast(`Copy failed: ${e?.message || e}`, "error");
    }
  }

  function openExportModal() { const s = useExportStore(); s.open(); }

  return {
    exportFormat, stripComments, splitSettings, splitPreview,
    refreshPreview, copy, openExportModal,
  };
}

export type UseContextActions = ReturnType<typeof useContextActions>;
"""

    def c_ctx_actions_bar(self) -> str:
        return """\
<template>
  <div
    v-if="hasContent"
    class="flex items-center gap-2 px-2 py-1 border-b border-gray-700 bg-gray-900/60"
    :aria-disabled="disabled ? 'true' : 'false'"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{ hasContent: boolean; disabled?: boolean }>();
</script>
"""

    def c_copy_menu_btn(self) -> str:
        return """\
<template>
  <div class="relative">
    <button
      type="button"
      class="px-2 py-1 text-xs bg-gray-800 hover:bg-gray-700 rounded-md border border-gray-700"
      :aria-expanded="open ? 'true' : 'false'"
      :aria-haspopup="'menu'"
      :disabled="disabled"
      @click="open = !open"
    >
      {{ labels.copy }} ▾
    </button>

    <div
      v-if="open"
      class="absolute mt-1 left-0 z-50 min-w-[220px] bg-gray-800 border border-gray-700 rounded shadow-lg p-2"
      role="menu"
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
import { computed, ref } from 'vue';
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

<style scoped>
/* close on outside click can be added by parent; here keep SRP */
</style>
"""

    def c_split_settings_popover(self) -> str:
        return """\
<template>
  <div class="relative">
    <button
      type="button"
      class="px-2 py-1 text-xs bg-gray-800 hover:bg-gray-700 rounded-md border border-gray-700"
      :disabled="disabled"
      @click="open = !open"
    >
      ⫿ Split
    </button>

    <div
      v-if="open"
      class="absolute mt-1 left-0 z-50 w-80 bg-gray-800 border border-gray-700 rounded shadow-lg p-3"
      role="dialog"
      aria-modal="true"
    >
      <div class="text-sm font-semibold text-white mb-2">{{ t.title }}</div>

      <label class="flex items-center gap-2 text-sm mb-2">
        <input type="checkbox" class="form-checkbox" v-model="local.enableAutoSplit" />
        <span>{{ t.enableAutoSplit }}</span>
      </label>

      <div class="grid grid-cols-2 gap-2 mb-2" :class="{ 'opacity-50 pointer-events-none': !local.enableAutoSplit }">
        <div>
          <div class="text-[11px] text-gray-400 mb-1">{{ t.maxTokens }}</div>
          <input type="number" min="100" class="w-full bg-gray-700 border-gray-600 rounded px-2 py-1 text-xs" v-model.number="local.maxTokensPerChunk" />
        </div>
        <div>
          <div class="text-[11px] text-gray-400 mb-1">{{ t.overlap }}</div>
          <input type="number" min="0" class="w-full bg-gray-700 border-gray-600 rounded px-2 py-1 text-xs" v-model.number="local.overlapTokens" />
        </div>
      </div>

      <div :class="{ 'opacity-50 pointer-events-none': !local.enableAutoSplit }">
        <div class="text-[11px] text-gray-400 mb-1">{{ t.strategy }}</div>
        <select v-model="local.splitStrategy" class="w-full bg-gray-700 border-gray-600 rounded px-2 py-1 text-xs">
          <option value="smart">{{ t.strategySmart }}</option>
          <option value="file">{{ t.strategyFile }}</option>
          <option value="token">{{ t.strategyToken }}</option>
        </select>
      </div>

      <div class="text-[11px] text-gray-400 mt-2 min-h-[1rem]">
        <template v-if="preview && local.enableAutoSplit">
          {{ t.willSplitInto(preview.chunkCount) }}
        </template>
      </div>

      <div class="flex justify-end gap-2 mt-2">
        <button type="button" class="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 rounded" @click="emit('refresh')">
          Refresh
        </button>
        <button type="button" class="px-2 py-1 text-xs bg-blue-600 hover:bg-blue-500 rounded text-white"
          @click="apply">
          Apply
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, computed } from 'vue';
import type { SplitSettings, SplitPreview } from '@/types/splitter';

const props = defineProps<{
  modelValue: SplitSettings;
  preview: SplitPreview | null;
  disabled?: boolean;
  i18n?: {
    title: string; enableAutoSplit: string; maxTokens: string; overlap: string; strategy: string;
    strategySmart: string; strategyFile: string; strategyToken: string; willSplitInto: (n: number)=>string;
  };
}>();
const emit = defineEmits<{
  (e: 'update:modelValue', v: SplitSettings): void;
  (e: 'apply'): void;
  (e: 'refresh'): void;
}>();

const t = computed(() => ({
  title: props.i18n?.title ?? 'Split settings',
  enableAutoSplit: props.i18n?.enableAutoSplit ?? 'Auto-split large contexts',
  maxTokens: props.i18n?.maxTokens ?? 'Max tokens per chunk',
  overlap: props.i18n?.overlap ?? 'Overlap tokens',
  strategy: props.i18n?.strategy ?? 'Strategy',
  strategySmart: props.i18n?.strategySmart ?? 'Smart (recommended)',
  strategyFile: props.i18n?.strategyFile ?? 'By files',
  strategyToken: props.i18n?.strategyToken ?? 'By tokens',
  willSplitInto: props.i18n?.willSplitInto ?? ((n:number)=>`Will split into ${n} parts`),
}));

const local = reactive<SplitSettings>({ ...props.modelValue });
const open = ref(false);

watch(() => props.modelValue, (v) => { Object.assign(local, v); }, { deep: true });

function apply() {
  emit('update:modelValue', { ...local });
  emit('apply');
  open.value = false;
}
</script>
"""

    def c_hash_util(self) -> str:
        return """\
// Auto-generated: small non-crypto hash (FNV-1a like) for storage keys
export function hashStringFNV1a(str: string): string {
  let h = 0x811c9dc5;
  for (let i = 0; i < str.length; i++) {
    h ^= str.charCodeAt(i);
    h += (h << 1) + (h << 4) + (h << 7) + (h << 8) + (h << 24);
  }
  return (h >>> 0).toString(36);
}
"""

    # ---------- steps ----------
    def step_1_types(self):
        self.current_step = 1
        self.log("📄 Шаг 1: types/splitter.ts")
        self.update_file("frontend/src/types/splitter.ts", self.c_types_splitter())

    def step_2_service(self):
        self.current_step = 2
        self.log("🧠 Шаг 2: services/splitter.service.ts")
        self.update_file("frontend/src/services/splitter.service.ts", self.c_service_splitter())

    def step_3_composable(self):
        self.current_step = 3
        self.log("🧩 Шаг 3: composables/useContextActions.ts")
        self.update_file("frontend/src/composables/useContextActions.ts", self.c_use_context_actions())

    def step_4_components(self):
        self.current_step = 4
        self.log("🖼️  Шаг 4: ContextActions components")
        base = "frontend/src/components/workspace/ContextActions"
        self.ensure_directory(base)
        self.update_file(f"{base}/ContextActionsBar.vue", self.c_ctx_actions_bar())
        self.update_file(f"{base}/CopyMenuButton.vue", self.c_copy_menu_btn())
        self.update_file(f"{base}/SplitSettingsPopover.vue", self.c_split_settings_popover())

    def step_5_utils(self):
        self.current_step = 5
        self.log("🧰 Шаг 5: utils/hash.ts")
        self.update_file("frontend/src/utils/hash.ts", self.c_hash_util())

    def execute(self):
        try:
            self.log("🚀 Запуск PATCH Batch 1/4")
            if self.should_exec(1): self.step_1_types()
            if self.should_exec(2): self.step_2_service()
            if self.should_exec(3): self.step_3_composable()
            if self.should_exec(4): self.step_4_components()
            if self.should_exec(5): self.step_5_utils()
            self.log("✅ Готово: Batch 1/4 перезаписан")
        except Exception as e:
            self.log(f"❌ Ошибка на шаге {self.current_step}: {e}")
            self.log(f"💡 Повтор: python fix.py --continue-from {self.current_step}")
            raise

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument("--dry-run", action="store_true", help="Предпросмотр без изменений")
    parser.add_argument("--continue-from", type=int, help="Продолжить с указанного шага")
    args = parser.parse_args()
    ProjectRefactor(dry_run=args.dry_run, start_from_step=args.continue_from).execute()