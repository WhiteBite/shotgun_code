<template>
  <transition name="modal-fade">
    <div v-if="gitStore.isHistoryVisible" class="fixed inset-0 bg-black/60 z-40 flex items-center justify-center" @click.self="gitStore.hideHistory">
      <div class="bg-gray-800 rounded-lg shadow-2xl w-full max-w-4xl max-h-[85vh] flex flex-col border border-gray-700">
        <header class="p-4 border-b border-gray-700 flex-shrink-0 flex justify-between items-center">
          <h2 class="text-xl font-semibold text-white">Commit History</h2>
          <button @click="gitStore.hideHistory" class="p-2 rounded-md hover:bg-gray-700">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </header>

        <main class="p-4 overflow-y-auto flex-grow">
          <div v-if="gitStore.isLoading" class="text-center text-gray-400">Loading history...</div>
          <div v-else-if="gitStore.commits.length === 0" class="text-center text-gray-400">No commits found.</div>
          <div v-else class="space-y-2">
            <div v-for="commit in gitStore.commits" :key="commit.hash" class="p-3 bg-gray-900/50 rounded-md border border-gray-700 flex items-start gap-3">
              <input type="checkbox" :value="commit.hash" v-model="selectedCommits" class="form-checkbox mt-1.5 bg-gray-700 border-gray-500 rounded text-blue-500">
              <div class="flex-grow">
                <p class="font-semibold text-white font-mono text-sm">{{ commit.subject }}</p>
                <div class="flex items-center gap-2 text-xs text-gray-400 mt-1">
                  <span>{{ commit.author }}</span>
                  <span class="text-gray-600">&bull;</span>
                  <span>{{ new Date(commit.date).toLocaleString() }}</span>
                  <span class="text-gray-600">&bull;</span>
                  <span class="font-mono">{{ commit.hash.substring(0, 7) }}</span>
                </div>
                <div class="flex flex-wrap gap-1 mt-2">
                  <span v-for="file in commit.files" :key="file"
                        class="px-2 py-0.5 bg-gray-700 text-xs text-gray-300 rounded-full cursor-pointer hover:bg-gray-600"
                        @mouseenter="showFilePreview(file, commit.hash, $event)"
                        @mouseleave="hideFilePreview">
                    {{ file }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </main>

        <footer class="p-4 border-t border-gray-700 flex justify-between items-center flex-shrink-0">
          <span class="text-sm text-gray-400">{{ selectedCommits.length }} commit(s) selected</span>
          <div class="flex gap-3">
            <button @click="gitStore.hideHistory" class="px-4 py-2 bg-gray-600 hover:bg-gray-500 rounded-md">Cancel</button>
            <button @click="applySelection" :disabled="selectedCommits.length === 0" class="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-md disabled:opacity-50">
              Add Files to Context
            </button>
          </div>
        </footer>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useGitStore } from '@/stores/git.store';
import { useContextStore } from '@/stores/context.store';
import { useUiStore } from '@/stores/ui.store';

const gitStore = useGitStore();
const contextStore = useContextStore();
const uiStore = useUiStore();
const selectedCommits = ref<string[]>([]);

function applySelection() {
  const filesToSelect = new Set<string>();
  selectedCommits.value.forEach(hash => {
    const commit = gitStore.commits.find(c => c.hash === hash);
    if (commit) {
      commit.files.forEach(file => filesToSelect.add(file));
    }
  });

  contextStore.selectFilesByRelPaths(Array.from(filesToSelect), 'git');
  gitStore.hideHistory();
  selectedCommits.value = [];
}

function showFilePreview(file: string, commitHash: string, event: MouseEvent) {
  if (event.ctrlKey || event.metaKey) {
    uiStore.showQuickLook({ path: file, type: 'git', commitHash, event });
  }
}

function hideFilePreview() {
  uiStore.hideQuickLook();
}
</script>

<style scoped>
.modal-fade-enter-active, .modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from, .modal-fade-leave-to {
  opacity: 0;
}
</style>