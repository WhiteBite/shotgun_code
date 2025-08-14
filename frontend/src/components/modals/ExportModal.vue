<template>
  <transition name="modal-fade">
    <div
      v-if="exportStore.isOpen"
      class="fixed inset-0 z-40 flex items-center justify-center bg-black/50"
      @click.self="exportStore.close()"
    >
      <div
        class="bg-gray-800 border border-gray-700 rounded-lg w-full max-w-5xl max-h-[85vh] overflow-hidden flex flex-col"
      >
        <header
          class="p-4 border-b border-gray-700 flex items-center justify-between"
        >
          <h2 class="text-xl font-semibold">Export Context</h2>
          <button
            class="p-2 rounded hover:bg-gray-700"
            @click="exportStore.close()"
            title="Close"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </header>

        <main
          class="p-4 grid grid-cols-1 lg:grid-cols-3 gap-4 overflow-y-auto overflow-x-hidden"
        >
          <!-- Copy to clipboard -->
          <section
            class="p-3 bg-gray-900/60 border border-gray-700 rounded-md flex flex-col min-w-0"
          >
            <h3 class="font-semibold text-white mb-1">Copy to Clipboard</h3>
            <p class="text-xs text-gray-400 mb-3 break-words">
              For pasting directly into AI chats.
            </p>

            <div class="mb-3">
              <div class="text-xs text-gray-400 mb-1">Format</div>
              <div
                class="inline-flex bg-gray-800/80 border border-gray-700 rounded-md overflow-hidden"
              >
                <button
                  :class="seg('plain')"
                  @click="exportStore.exportFormat = 'plain'"
                  title="Simple headers + content. Most compatible."
                >
                  Plain
                </button>
                <button
                  :class="seg('manifest')"
                  @click="exportStore.exportFormat = 'manifest'"
                  title="Adds file tree manifest, then file contents. Recommended."
                >
                  Manifest + Text
                </button>
                <button
                  :class="seg('json')"
                  @click="exportStore.exportFormat = 'json'"
                  title="Machine-readable JSON array of files"
                >
                  JSON
                </button>
              </div>
            </div>

            <label class="flex items-center gap-2 text-sm mb-1">
              <input
                type="checkbox"
                class="form-checkbox"
                v-model="exportStore.stripComments"
              />
              Strip comments
            </label>
            <label class="flex items-center gap-2 text-sm mb-3">
              <input
                type="checkbox"
                class="form-checkbox"
                v-model="exportStore.includeManifest"
                :disabled="exportStore.exportFormat !== 'manifest'"
              />
              Include file manifest
            </label>
            <button
              :disabled="exportStore.isLoading"
              @click="exportStore.doExportClipboard"
              class="mt-auto py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
            >
              Copy Context to Clipboard
            </button>
          </section>

          <!-- Export for AI -->
          <section
            class="p-3 bg-gray-900/60 border border-gray-700 rounded-md flex flex-col min-w-0"
          >
            <h3 class="font-semibold text-white mb-1">Export for AI</h3>
            <p class="text-xs text-gray-400 mb-3 break-words">
              Generate compact, text-only PDF(s) optimized for model upload.
            </p>

            <label class="block text-xs text-gray-400">AI Profile</label>
            <select
              v-model="exportStore.aiProfile"
              class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm mb-1"
            >
              <option
                value="Claude-3"
                title="Tight PDF, low overhead. Best for Claude family."
              >
                Claude 3
              </option>
              <option value="GPT-4o" title="Balanced PDF. Good for GPT family.">
                GPT-4o
              </option>
              <option
                value="Generic"
                title="Generic compact PDF. Works with most models."
              >
                Generic
              </option>
            </select>
            <div class="text-[11px] text-gray-500 mb-3 min-h-[1rem]">
              {{ exportStore.aiProfileHint }}
            </div>

            <div class="grid grid-cols-2 gap-2 mb-3">
              <div>
                <label class="block text-xs text-gray-400">Token limit</label>
                <input
                  type="number"
                  v-model.number="exportStore.tokenLimit"
                  class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                />
              </div>
              <div>
                <label class="block text-xs text-gray-400"
                  >File size (KB)</label
                >
                <input
                  type="number"
                  v-model.number="exportStore.fileSizeLimitKB"
                  class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                />
              </div>
            </div>
            <button
              :disabled="exportStore.isLoading"
              @click="exportStore.doExportAI"
              class="mt-auto py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
            >
              Generate & Download AI-PDF
            </button>
          </section>

          <!-- Export for Human -->
          <section
            class="p-3 bg-gray-900/60 border border-gray-700 rounded-md flex flex-col min-w-0"
          >
            <h3 class="font-semibold text-white mb-1">Export for Human</h3>
            <p class="text-xs text-gray-400 mb-3 break-words">
              Create a formatted PDF with syntax highlighting for code review.
            </p>
            <div class="grid grid-cols-2 gap-2 mb-2">
              <div>
                <label class="block text-xs text-gray-400">Theme</label>
                <select
                  v-model="exportStore.theme"
                  class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                >
                  <option value="Dark">Dark</option>
                  <option value="Light">Light</option>
                </select>
              </div>
              <div class="flex items-end gap-2">
                <label class="flex items-center gap-2 text-sm"
                  ><input
                    type="checkbox"
                    class="form-checkbox"
                    v-model="exportStore.includeLineNumbers"
                  />
                  Line numbers</label
                >
                <label class="flex items-center gap-2 text-sm"
                  ><input
                    type="checkbox"
                    class="form-checkbox"
                    v-model="exportStore.includePageNumbers"
                  />
                  Page numbers</label
                >
              </div>
            </div>
            <button
              :disabled="exportStore.isLoading"
              @click="exportStore.doExportHuman"
              class="mt-auto py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
            >
              Generate & Download PDF
            </button>
          </section>
        </main>

        <footer class="p-3 border-t border-gray-700 text-right">
          <span v-if="exportStore.isLoading" class="text-xs text-gray-400"
            >Generating...</span
          >
        </footer>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { useExportStore } from "@/stores/export.store";
const exportStore = useExportStore();

function seg(v: string) {
  return [
    "px-2 py-1 text-xs",
    exportStore.exportFormat === v
      ? "bg-blue-600 text-white"
      : "text-gray-300 hover:bg-gray-700",
  ];
}
</script>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.15s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
