
<template>
  <!-- ИСПРАВЛЕНО: Восстановлена полная верстка (template) -->
  <div class="p-6 flex flex-col h-full">
    <h2 class="text-xl font-semibold text-gray-800 mb-4">Step 4: Apply Patch</h2>

    <div v-if="diff.isLoading" class="flex-grow flex justify-center items-center">
      <p class="text-gray-600">Loading split diffs...</p>
    </div>

    <div v-else-if="diff.splitDiffs.length > 0" class="flex-grow overflow-y-auto space-y-6">
      <p class="text-gray-600 mb-2 text-xs">
        The original diff has been split into {{ diff.splitDiffs.length }} smaller diffs.
        Copy each part and apply it using your preferred tool.
      </p>
      <div v-for="(split, index) in diff.splitDiffs" :key="index" class="border p-4 rounded-md bg-gray-50">
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-medium text-gray-700">Split {{ index + 1 }} of {{ diff.splitDiffs.length }}</h3>
          <button
              @click="copy(split, index)"
              class="px-3 py-1 bg-gray-200 text-gray-700 text-xs font-semibold rounded-md"
          >
            {{ copyStatus[index] === 'success' ? 'Copied!' : 'Copy' }}
          </button>
        </div>
        <textarea
            :value="split"
            rows="10"
            readonly
            class="w-full p-2 mt-2 border border-gray-200 rounded-md bg-white font-mono text-xs"
            style="min-height: 150px;"
        ></textarea>
      </div>
    </div>

    <div v-else class="flex-grow flex justify-center items-center">
      <p class="text-gray-500">No split diffs to display. Go to Step 3 to split a diff.</p>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';
import { useDiffStore } from '../../stores/diff';
import { useClipboard } from '../../composables/useClipboard';

const diff = useDiffStore();
const { copy: copyText } = useClipboard();
const copyStatus = ref({});

watch(() => diff.splitDiffs, () => {
  copyStatus.value = {};
});

async function copy(text, index) {
  await copyText(text);
  copyStatus.value[index] = 'success';
  setTimeout(() => {
    copyStatus.value[index] = 'idle';
  }, 2000);
}
</script>