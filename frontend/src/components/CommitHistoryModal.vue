<template>
  <div v-if="isVisible" class="fixed inset-0 bg-gray-600 bg-opacity-75 z-50 flex justify-center items-center" @click.self="close">
    <div class="bg-white rounded-lg shadow-xl w-full max-w-3xl max-h-[80vh] flex flex-col">
      <header class="p-4 border-b">
        <h2 class="text-xl font-semibold">Select Commits</h2>
        <div class="mt-2 flex items-center gap-2">
          <input
              type="text"
              v-model="branchName"
              placeholder="Enter branch name (e.g., main, feature/auth)"
              class="w-full p-2 border border-gray-300 rounded-md shadow-sm"
              @keyup.enter="handleSearch"
          />
          <button @click="handleSearch" :disabled="isLoading" class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
            {{ isLoading ? '...' : 'Search' }}
          </button>
        </div>
      </header>

      <main class="p-4 overflow-y-auto flex-grow">
        <div v-if="isLoading" class="text-center text-gray-500">
          <p>Loading commits...</p>
        </div>
        <div v-else-if="commits.length === 0" class="text-center text-gray-500">
          <p>No recent commits found.</p>
        </div>
        <ul v-else class="space-y-1">
          <li v-for="commit in commits" :key="commit.hash"
              class="border rounded-md"
              :class="{'bg-blue-50 border-blue-200': selectedHashes.has(commit.hash)}">
            <div
                class="p-2 flex items-center cursor-pointer hover:bg-gray-100"
                @click="toggleCommitSelection(commit.hash)"
            >
              <input
                  type="checkbox"
                  :checked="selectedHashes.has(commit.hash)"
                  class="h-4 w-4 rounded border-gray-300 mr-3"
              />
              <div class="flex-grow">
                <p class="font-mono text-xs text-gray-500 flex items-center gap-2">
                  <span>{{ commit.hash.substring(0, 8) }}</span>
                  <span v-if="commit.isMerge" class="px-1.5 py-0.5 text-xs font-semibold bg-gray-200 text-gray-700 rounded-full">Merge</span>
                </p>
                <p class="font-medium">{{ commit.subject }}</p>
              </div>
              <button
                  v-if="commit.files && commit.files.length > 0"
                  @click.stop="toggleCommitExpansion(commit.hash)"
                  class="p-1 text-xs text-gray-600 hover:bg-gray-200 rounded"
              >
                ({{ commit.files.length }} files) {{ expandedCommits.has(commit.hash) ? '▲' : '▼' }}
              </button>
            </div>
            <div v-if="expandedCommits.has(commit.hash)" class="px-4 pb-2 pt-1 border-t bg-gray-50">
              <ul class="list-disc list-inside text-xs text-gray-700 max-h-48 overflow-y-auto">
                <li v-for="file in commit.files" :key="file">{{ file }}</li>
              </ul>
            </div>
          </li>
        </ul>
      </main>

      <footer class="p-4 border-t flex justify-end space-x-4">
        <button @click="close" class="px-4 py-2 bg-gray-200 rounded-md hover:bg-gray-300">Cancel</button>
        <button @click="applySelection" :disabled="selectedHashes.size === 0" class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
          Apply Selection ({{ selectedHashes.size }})
        </button>
      </footer>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  isVisible: Boolean,
  commits: Array,
  isLoading: Boolean,
});

const emit = defineEmits(['close', 'apply', 'search']);

const selectedHashes = ref(new Set());
const expandedCommits = ref(new Set());
const branchName = ref('');

watch(() => props.isVisible, (newValue) => {
  if (newValue) {
    selectedHashes.value.clear();
    expandedCommits.value.clear();
    branchName.value = '';
  }
});

function toggleCommitSelection(hash) {
  if (selectedHashes.value.has(hash)) {
    selectedHashes.value.delete(hash);
  } else {
    selectedHashes.value.add(hash);
  }
}

function toggleCommitExpansion(hash) {
  if (expandedCommits.value.has(hash)) {
    expandedCommits.value.delete(hash);
  } else {
    expandedCommits.value.add(hash);
  }
}

function handleSearch() {
  emit('search', branchName.value.trim());
}

function close() {
  emit('close');
}

function applySelection() {
  emit('apply', Array.from(selectedHashes.value));
}
</script>