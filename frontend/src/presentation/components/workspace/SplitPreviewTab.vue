<template>
  <div class="h-full flex flex-col">
    <!-- Split Info Header -->
    <div class="p-3 border-b border-gray-700 bg-gray-800/30 flex-shrink-0">
      <div class="text-xs text-gray-400 mb-2">
        Split into {{ splitPreview?.chunkCount || 0 }} parts ({{
          splitPreview?.totalTokens?.toLocaleString() || 0
        }}
        tokens total)
      </div>
      <div class="flex gap-2 flex-wrap">
        <button
          v-for="(chunk, index) in splitPreview?.chunks || []"
          :key="index"
          :class="[
            'px-2 py-1 text-xs rounded border transition-colors',
            selectedChunk === index
              ? 'bg-blue-600 border-blue-500 text-white'
              : 'bg-gray-700 border-gray-600 text-gray-300 hover:bg-gray-600',
          ]"
          @click="selectedChunk = index"
        >
          Part {{ index + 1 }} (~{{ chunk.tokens }}t)
        </button>
      </div>
    </div>

    <!-- Selected Chunk Content -->
    <div class="flex-1 min-h-0 relative">
      <div
        v-if="!selectedChunkContent"
        class="flex items-center justify-center h-full text-gray-500"
      >
        <div class="text-center">
          <p>Select a part to preview</p>
        </div>
      </div>
      <div v-else class="h-full relative">
        <!-- Copy button for current chunk -->
        <div class="absolute top-2 right-2 z-10">
          <button
            class="p-2 bg-gray-800/80 hover:bg-gray-700 rounded transition-colors border border-gray-600"
            title="Copy this part"
            aria-label="Copy current part"
            @click="copyCurrentChunk"
          >
            <CopyIcon class="w-4 h-4" />
          </button>
        </div>

        <!-- Navigation buttons -->
        <div class="absolute top-2 left-2 z-10 flex gap-1">
          <button
            :disabled="selectedChunk <= 0"
            class="p-2 bg-gray-800/80 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed rounded transition-colors border border-gray-600"
            title="Previous part"
            @click="previousChunk"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 19l-7-7 7-7"
              />
            </svg>
          </button>
          <button
            :disabled="selectedChunk >= (splitPreview?.chunks?.length || 0) - 1"
            class="p-2 bg-gray-800/80 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed rounded transition-colors border border-gray-600"
            title="Next part"
            @click="nextChunk"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 5l7 7-7 7"
              />
            </svg>
          </button>
        </div>

        <ScrollableContent
          :content="selectedChunkContent"
          :language="'plaintext'"
          :highlight="false"
          class="h-full pt-12"
        />
      </div>
    </div>

    <!-- Chunk Stats Footer -->
    <div
      v-if="selectedChunkData"
      class="p-2 border-t border-gray-700 bg-gray-800/30 flex-shrink-0"
    >
      <div class="flex items-center justify-between text-xs text-gray-400">
        <div>
          Part {{ selectedChunk + 1 }} of
          {{ splitPreview?.chunks?.length || 0 }}:
          {{ selectedChunkData.chars.toLocaleString() }} chars, ~{{
            selectedChunkData.tokens.toLocaleString()
          }}
          tokens
        </div>
        <div class="flex items-center gap-2">
          <span class="text-gray-500">Copy:</span>
          <button
            class="px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-xs transition-colors"
            @click="copyCurrentChunk"
          >
            Part {{ selectedChunk + 1 }}
          </button>
          <button
            class="px-2 py-1 bg-blue-600 hover:bg-blue-700 rounded text-xs transition-colors"
            @click="copyAllChunks"
          >
            All Parts
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { SplitPreview } from "@/types/splitter";
import ScrollableContent from "@/presentation/components/shared/ScrollableContent.vue";
import { CopyIcon } from "@/presentation/components/icons/index";
import { useNotificationsStore } from "@/stores/notifications.store";

const props = defineProps<{
  splitPreview: SplitPreview | null;
}>();

const emit = defineEmits<{
  (e: "copy-chunk", index: number): void;
  (e: "copy-all-chunks"): void;
}>();

const notifications = useNotificationsStore();
const selectedChunk = ref(0);

const selectedChunkData = computed(() => {
  if (
    !props.splitPreview?.chunks ||
    selectedChunk.value >= props.splitPreview.chunks.length
  ) {
    return null;
  }
  return props.splitPreview.chunks[selectedChunk.value];
});

const selectedChunkContent = computed(() => {
  return selectedChunkData.value?.text || "";
});

function copyCurrentChunk() {
  if (!selectedChunkData.value) return;

  navigator.clipboard.writeText(selectedChunkData.value.text);
  notifications.addLog(
    `Part ${selectedChunk.value + 1} copied to clipboard`,
    "success",
  );
  emit("copy-chunk", selectedChunk.value);
}

function copyAllChunks() {
  if (!props.splitPreview?.chunks) return;

  const allText = props.splitPreview.chunks
    .map((chunk, index) => `=== Part ${index + 1} ===\n${chunk.text}\n`)
    .join("\n");

  navigator.clipboard.writeText(allText);
  notifications.addLog("All parts copied to clipboard", "success");
  emit("copy-all-chunks");
}

function previousChunk() {
  if (selectedChunk.value > 0) {
    selectedChunk.value--;
  }
}

function nextChunk() {
  if (
    props.splitPreview?.chunks &&
    selectedChunk.value < props.splitPreview.chunks.length - 1
  ) {
    selectedChunk.value++;
  }
}
</script>