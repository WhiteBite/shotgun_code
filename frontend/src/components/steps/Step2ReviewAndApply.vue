<template>
  <div class="p-4 h-full flex flex-col">
    <div class="mb-4">
      <h3 class="text-md font-medium text-gray-700 mb-2">Review & Apply</h3>
      <p class="text-sm text-gray-500">Paste the `git diff` output from the LLM below to review and split it for application.</p>
    </div>

    <div class="flex-grow grid grid-cols-2 gap-4 overflow-hidden">
      <!-- Left side: Input -->
      <div class="flex flex-col">
        <label for="diff-input" class="block text-sm font-medium text-gray-700 mb-1">Git Diff Input:</label>
        <textarea
            id="diff-input"
            v-model="diff.gitDiffInput"
            class="w-full flex-grow p-2 border border-gray-300 rounded-md shadow-sm font-mono text-xs"
            placeholder="Paste the diff here..."
        ></textarea>
      </div>

      <!-- Right side: Output -->
      <div class="flex flex-col overflow-hidden">
        <div class="flex items-center justify-between mb-1">
          <label class="block text-sm font-medium text-gray-700">Split Diffs:</label>
          <div class="flex items-center space-x-2">
            <label for="line-limit" class="text-sm text-gray-600">Line limit:</label>
            <input
                type="number"
                id="line-limit"
                v-model.number="diff.splitLineLimit"
                class="w-20 p-1 border border-gray-300 rounded-md text-sm"
                step="50"
            />
          </div>
        </div>
        <div class="flex-grow border border-gray-200 rounded-md bg-gray-50 overflow-y-auto p-2 space-y-4">
          <div v-if="diff.isLoading" class="text-center text-gray-500">
            <p>Splitting diff...</p>
          </div>
          <div v-else-if="diff.splitDiffs.length > 0" v-for="(split, index) in diff.splitDiffs" :key="index" class="bg-white p-2 border rounded">
            <div class="flex justify-between items-center mb-1">
              <h4 class="font-semibold text-gray-800">Part {{ index + 1 }} / {{ diff.splitDiffs.length }}</h4>
              <button
                  @click="copyPart(split, index)"
                  class="px-3 py-1 bg-gray-200 text-gray-700 text-xs font-semibold rounded-md hover:bg-gray-300"
              >
                {{ copyStatus[index] === 'success' ? 'Copied!' : 'Copy' }}
              </button>
            </div>
            <pre class="w-full p-2 bg-gray-100 rounded-md font-mono text-xs overflow-x-auto">{{ split }}</pre>
          </div>
          <div v-else class="text-center text-gray-500 pt-10">
            <p>Output will appear here after processing.</p>
          </div>
        </div>
      </div>
    </div>

    <div class="mt-4 flex justify-end">
      <button
          @click="diff.processAndSplitDiff"
          :disabled="!diff.gitDiffInput.trim() || diff.isLoading"
          class="px-6 py-2 bg-blue-600 text-white font-semibold rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-wait"
      >
        {{ diff.isLoading ? 'Processing...' : 'Process and Split Diff' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';
import { useDiffStore } from '../../stores/diffStore.js';
import { useClipboard } from '../../composables/useClipboard';
import { useNotificationsStore } from '../../stores/notificationsStore';

const diff = useDiffStore();
const notifications = useNotificationsStore();
const { copy: copyText, status: globalCopyStatus } = useClipboard();
const copyStatus = ref({});

watch(() => diff.splitDiffs, () => {
  copyStatus.value = {}; // Reset status when diffs change
});

async function copyPart(text, index) {
  await copyText(text);
  const currentStatus = globalCopyStatus.value;
  if (currentStatus === 'success') {
    copyStatus.value[index] = 'success';
    notifications.addLog(`Часть ${index + 1} скопирована в буфер обмена.`, 'success');
    setTimeout(() => {
      copyStatus.value[index] = 'idle';
    }, 2000);
  } else if (currentStatus === 'error') {
    notifications.addLog(`Не удалось скопировать часть ${index + 1}.`, 'error');
  }
}
</script>