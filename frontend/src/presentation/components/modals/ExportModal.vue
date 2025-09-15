<template>
  <transition name="modal-fade">
    <div
      v-if="exportStore.isOpen"
      class="fixed inset-0 z-40 flex items-center justify-center bg-black/50"
      role="dialog"
      aria-modal="true"
      @click.self="exportStore.close()"
    >
      <div
        class="bg-gray-800 border border-gray-700 rounded-lg w-full max-w-6xl max-h-[90vh] overflow-hidden flex flex-col"
      >
        <header
          class="p-4 border-b border-gray-700 flex items-center justify-between"
        >
          <h2 class="text-xl font-semibold">Export Context</h2>
          <button
            class="p-2 rounded hover:bg-gray-700"
            title="Close"
            aria-label="Close"
            @click="exportStore.close()"
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
          <!-- Copy to Clipboard -->
          <section
            class="p-3 bg-gray-900/60 border border-gray-700 rounded-md flex flex-col min-w-0"
          >
            <h3 class="font-semibold text-white mb-1">Copy to Clipboard</h3>
            <p class="text-xs text-gray-400 mb-3 break-words">
              Copy the generated context directly to your clipboard. Choose the
              desired format.
            </p>
            <div class="mb-3">
              <div class="text-xs text-gray-400 mb-1">Format</div>
              <div
                class="inline-flex bg-gray-800/80 border border-gray-700 rounded-md overflow-hidden"
              >
                <button
                  :class="seg('plain')"
                  @click="exportStore.exportFormat = 'plain'"
                >
                  Plain
                </button>
                <button
                  :class="seg('manifest')"
                  @click="exportStore.exportFormat = 'manifest'"
                >
                  Manifest + Text
                </button>
                <button
                  :class="seg('json')"
                  @click="exportStore.exportFormat = 'json'"
                >
                  JSON
                </button>
              </div>
            </div>
            <label class="flex items-center gap-2 text-sm mb-1"
              ><input
                v-model="exportStore.stripComments"
                type="checkbox"
                class="form-checkbox"
              />
              Strip comments</label
            >
            <label class="flex items-center gap-2 text-sm mb-3"
              ><input
                v-model="exportStore.includeManifest"
                type="checkbox"
                class="form-checkbox"
                :disabled="exportStore.exportFormat !== 'manifest'"
              />
              Include file manifest</label
            >
            <button
              :disabled="exportStore.isLoading"
              class="mt-auto py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
              @click="exportStore.doExportClipboard"
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
              Generate PDF(s) optimized for AI models. Will split into multiple
              files if context is too large.
            </p>

            <label class="block text-xs text-gray-400">AI Profile</label>
            <select
              v-model="exportStore.aiProfile"
              class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm mb-1"
            >
              <option value="Claude-3">Claude 3</option>
              <option value="GPT-4o">GPT-4o</option>
              <option value="Generic">Generic</option>
            </select>
            <div class="text-[11px] text-gray-500 mb-3 min-h-[1rem]">
              {{ exportStore.aiProfileHint }}
            </div>

            <div class="mb-3 p-2 bg-gray-800/50 rounded border border-gray-600">
              <label class="flex items-center gap-2 text-sm mb-2">
                <input
                  v-model="exportStore.enableAutoSplit"
                  type="checkbox"
                  class="form-checkbox"
                />
                <span title="Enable automatic chunking to respect model limits"
                  >Auto-split large contexts</span
                >
              </label>

              <details
                v-if="exportStore.enableAutoSplit"
                class="rounded bg-gray-900/40 border border-gray-700 p-2"
              >
                <summary class="cursor-pointer text-xs text-gray-300">
                  Advanced Settings
                </summary>
                <div class="space-y-2 mt-2">
                  <div>
                    <label
                      class="block text-xs text-gray-400"
                      title="Upper bound for tokens in each chunk"
                      >Max tokens per chunk</label
                    >
                    <input
                      v-model.number="exportStore.maxTokensPerChunk"
                      type="number"
                      min="1"
                      class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                    />
                  </div>
                  <div>
                    <label
                      class="block text-xs text-gray-400"
                      title="Smart keeps file boundaries; token may split files"
                      >Split strategy</label
                    >
                    <select
                      v-model="exportStore.splitStrategy"
                      class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                    >
                      <option value="smart">Smart (recommended)</option>
                      <option value="file">By whole files</option>
                      <option value="token">By token count</option>
                    </select>
                    <div class="text-[10px] text-gray-500 mt-1">
                      {{ exportStore.splitStrategyHint }}
                    </div>
                  </div>
                  <div>
                    <label
                      class="block text-xs text-gray-400"
                      title="Overlap to preserve context continuity across chunks"
                      >Overlap tokens</label
                    >
                    <input
                      v-model.number="exportStore.overlapTokens"
                      type="number"
                      min="0"
                      class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                    />
                  </div>
                </div>
              </details>

              <div
                v-if="exportStore.shouldAutoSplit"
                class="mt-2 text-xs text-yellow-400"
              >
                ⚠️ Context will be auto-split into multiple files
              </div>
            </div>

            <div class="grid grid-cols-2 gap-2 mb-3">
              <div>
                <label class="block text-xs text-gray-400">Token limit</label
                ><input
                  v-model.number="exportStore.tokenLimit"
                  type="number"
                  :disabled="exportStore.enableAutoSplit"
                  class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm disabled:opacity-50"
                />
              </div>
              <div>
                <label class="block text-xs text-gray-400">File size (KB)</label
                ><input
                  v-model.number="exportStore.fileSizeLimitKB"
                  type="number"
                  class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                />
              </div>
            </div>
            <button
              :disabled="exportStore.isLoading"
              class="mt-auto py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
              @click="exportStore.doExportAI"
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
              Generate a human-readable PDF of the context.
            </p>

            <div class="space-y-3 mb-3">
              <div>
                <label class="block text-xs text-gray-400">Theme</label
                ><select
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
                    v-model="exportStore.includeLineNumbers"
                    type="checkbox"
                    class="form-checkbox"
                  />
                  Line numbers</label
                >
                <label class="flex items-center gap-2 text-sm"
                  ><input
                    v-model="exportStore.includePageNumbers"
                    type="checkbox"
                    class="form-checkbox"
                  />
                  Page numbers</label
                >
              </div>
            </div>
            <button
              :disabled="exportStore.isLoading"
              class="mt-auto py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
              @click="exportStore.doExportHuman"
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
import { onMounted, onUnmounted } from "vue";
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
function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") exportStore.close();
}
onMounted(() => document.addEventListener("keydown", onKeydown));
onUnmounted(() => document.removeEventListener("keydown", onKeydown));
</script>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.3s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>