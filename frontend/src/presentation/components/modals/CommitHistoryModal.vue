<template>
  <transition name="modal-fade">
    <div
      v-if="gitStore.isHistoryVisible"
      class="fixed inset-0 bg-black/60 z-40 flex items-center justify-center"
      role="dialog"
      aria-modal="true"
      @click.self="gitStore.hideHistory"
    >
      <div
        class="bg-gray-800 rounded-lg shadow-2xl w-full max-w-4xl max-h-[85vh] flex flex-col border border-gray-700"
      >
        <header
          class="p-4 border-b border-gray-700 flex-shrink-0 flex justify-between items-center"
        >
          <h2 class="text-xl font-semibold text-white">
            {{ t("commits.title") }}
          </h2>
          <button
            class="p-2 rounded-md hover:bg-gray-700"
            aria-label="Close"
            @click="gitStore.hideHistory"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
            >
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </header>

        <main class="p-4 overflow-y-auto flex-grow">
          <div v-if="gitStore.isLoading" class="text-center text-gray-400">
            {{ t("commits.loading") }}
          </div>
          <div
            v-else-if="gitStore.commits.length === 0"
            class="text-center text-gray-400"
          >
            {{ t("commits.empty") }}
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="commit in gitStore.commits"
              :key="commit.hash"
              class="p-3 bg-gray-900/50 rounded-md border border-gray-700"
            >
              <div class="flex gap-3">
                <input
                  v-model="selectedCommits"
                  type="checkbox"
                  :value="commit.hash"
                  class="form-checkbox mt-1.5 bg-gray-700 border-gray-500 rounded text-blue-500"
                />
                <div class="flex-grow min-w-0">
                  <p
                    class="font-semibold text-white font-mono text-sm truncate"
                  >
                    {{ commit.subject }}
                  </p>
                  <div
                    class="flex items-center gap-2 text-xs text-gray-400 mt-1"
                  >
                    <span>{{ commit.author }}</span>
                    <span class="text-gray-600">&bull;</span>
                    <span>{{ commit.date }}</span>
                    <span class="text-gray-600">&bull;</span>
                    <span class="font-mono">{{
                      commit.hash.substring(0, 7)
                    }}</span>
                  </div>
                  <div class="mt-2 space-y-0.5 text-xs text-gray-300">
                    <div
                      v-for="file in commit.files"
                      :key="file"
                      class="truncate cursor-pointer hover:text-white"
                      @click="openPinnedFile(file, commit.hash, $event)"
                      @mouseenter="showFilePreview(file, commit.hash, $event)"
                      @mouseleave="hideFilePreview"
                    >
                      {{ file }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </main>

        <footer
          class="p-4 border-t border-gray-700 flex justify-between items-center flex-shrink-0"
        >
          <span class="text-sm text-gray-400"
            >{{ selectedCommits.length }} commit(s) selected</span
          >
          <div class="flex gap-3">
            <button
              class="px-4 py-2 bg-gray-600 hover:bg-gray-500 rounded-md"
              @click="gitStore.hideHistory"
            >
              {{ t("button.cancel") }}
            </button>
            <button
              :disabled="selectedCommits.length === 0"
              class="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-md disabled:opacity-50"
              @click="applySelection"
            >
              {{ t("button.addFiles") }}
            </button>
          </div>
        </footer>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { t } from "@/lib/i18n";
import { useGitStore } from "@/stores/git.store";
import { useUiStore } from "@/stores/ui.store";
import { useProjectStore } from "@/stores/project.store";

const gitStore = useGitStore();
const uiStore = useUiStore();
const projectStore = useProjectStore();
const selectedCommits = ref<string[]>([]);

function applySelection() {
  gitStore.applyCommitSelection(selectedCommits.value);
  gitStore.hideHistory();
  selectedCommits.value = [];
}

function showFilePreview(file: string, commitHash: string, event: MouseEvent) {
  const rootDir = projectStore.currentProject?.path || "";
  if (!rootDir) return;
  if (event.ctrlKey || event.metaKey) {
    uiStore.showQuickLook({
      rootDir,
      path: file,
      type: "git",
      commitHash,
      position: { x: event.clientX + 10, y: event.clientY + 10 },
      isPinned: false,
    });
  }
}
function hideFilePreview() {
  uiStore.hideQuickLook();
}
function openPinnedFile(file: string, commitHash: string, event: MouseEvent) {
  const rootDir = projectStore.currentProject?.path || "";
  if (!rootDir) return;
  uiStore.showQuickLook({
    rootDir,
    path: file,
    type: "git",
    commitHash,
    // Центрируем закреплённый предпросмотр для удобства чтения
    position: { x: window.innerWidth / 2, y: window.innerHeight / 2 },
    isPinned: true,
  });
}
function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") gitStore.hideHistory();
}

onMounted(() => document.addEventListener("keydown", onKeydown));
onUnmounted(() => document.removeEventListener("keydown", onKeydown));
</script>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>