<template>
  <transition name="modal-fade">
    <div
      v-if="exportService.isOpen"
      class="fixed inset-0 z-40 flex items-center justify-center bg-black/50"
      role="dialog"
      aria-modal="true"
      @click.self="exportService.handleClose()"
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
            @click="exportService.handleClose()"
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
                  :class="exportService.getSegmentButtonClasses('plain')"
                  @click="exportService.exportFormat = 'plain'"
                >
                  Plain
                </button>
                <button
                  :class="exportService.getSegmentButtonClasses('manifest')"
                  @click="exportService.exportFormat = 'manifest'"
                >
                  Manifest + Text
                </button>
                <button
                  :class="exportService.getSegmentButtonClasses('json')"
                  @click="exportService.exportFormat = 'json'"
                >
                  JSON
                </button>
              </div>
            </div>
            <label class="flex items-center gap-2 text-sm mb-1"
              ><input
                v-model="exportService.stripComments"
                type="checkbox"
                class="form-checkbox"
              />
              Strip comments</label
            >
            <label class="flex items-center gap-2 text-sm mb-3"
              ><input
                v-model="exportService.includeManifest"
                type="checkbox"
                class="form-checkbox"
                :disabled="!exportService.isManifestFormatSelected"
              />
              Include file manifest</label
            >
            <button
              :disabled="exportService.isLoading"
              class="mt-auto py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
              @click="exportService.handleExportClipboard"
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
              v-model="exportService.aiProfile"
              class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm mb-1"
            >
              <option value="Claude-3">Claude 3</option>
              <option value="GPT-4o">GPT-4o</option>
              <option value="Generic">Generic</option>
            </select>
            <div class="text-[11px] text-gray-500 mb-3 min-h-[1rem]">
              {{ exportService.aiProfileHint }}
            </div>

            <div class="mb-3 p-2 bg-gray-800/50 rounded border border-gray-600">
              <label class="flex items-center gap-2 text-sm mb-2">
                <input
                  v-model="exportService.enableAutoSplit"
                  type="checkbox"
                  class="form-checkbox"
                />
                <span title="Enable automatic chunking to respect model limits"
                  >Auto-split large contexts</span
                >
              </label>

              <details
                v-if="exportService.enableAutoSplit"
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
                      v-model.number="exportService.maxTokensPerChunk"
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
                      v-model="exportService.splitStrategy"
                      class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                    >
                      <option value="smart">Smart (recommended)</option>
                      <option value="file">By whole files</option>
                      <option value="token">By token count</option>
                    </select>
                    <div class="text-[10px] text-gray-500 mt-1">
                      {{ exportService.splitStrategyHint }}
                    </div>
                  </div>
                  <div>
                    <label
                      class="block text-xs text-gray-400"
                      title="Overlap to preserve context continuity across chunks"
                      >Overlap tokens</label
                    >
                    <input
                      v-model.number="exportService.overlapTokens"
                      type="number"
                      min="0"
                      class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                    />
                  </div>
                </div>
              </details>

              <div
                v-if="exportService.shouldAutoSplit"
                class="mt-2 text-xs text-yellow-400"
              >
                ⚠️ Context will be auto-split into multiple files
              </div>
            </div>

            <div class="grid grid-cols-2 gap-2 mb-3">
              <div>
                <label class="block text-xs text-gray-400">Token limit</label
                ><input
                  v-model.number="exportService.tokenLimit"
                  type="number"
                  :disabled="exportService.enableAutoSplit"
                  class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm disabled:opacity-50"
                />
              </div>
              <div>
                <label class="block text-xs text-gray-400">File size (KB)</label
                ><input
                  v-model.number="exportService.fileSizeLimitKB"
                  type="number"
                  class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                />
              </div>
            </div>
            <button
              :disabled="exportService.isLoading"
              class="mt-auto py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
              @click="exportService.handleExportAI"
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
                  v-model="exportService.theme"
                  class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
                >
                  <option value="Dark">Dark</option>
                  <option value="Light">Light</option>
                </select>
              </div>
              <div class="flex items-end gap-2">
                <label class="flex items-center gap-2 text-sm"
                  ><input
                    v-model="exportService.includeLineNumbers"
                    type="checkbox"
                    class="form-checkbox"
                  />
                  Line numbers</label
                >
                <label class="flex items-center gap-2 text-sm"
                  ><input
                    v-model="exportService.includePageNumbers"
                    type="checkbox"
                    class="form-checkbox"
                  />
                  Page numbers</label
                >
              </div>
            </div>
            <button
              :disabled="exportService.isLoading"
              class="mt-auto py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
              @click="exportService.handleExportHuman"
            >
              Generate & Download PDF
            </button>
          </section>
        </main>

        <footer class="p-3 border-t border-gray-700 text-right">
          <span v-if="exportService.isLoading" class="text-xs text-gray-400"
            >Generating...</span
          >
        </footer>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { useExportService } from "@/composables/useExportService";

const { exportService, exportStore } = useExportService();
</script>