<template>
  <transition name="slide-fade">
    <div v-if="uiStore.activeDrawer === 'ignore'" class="fixed inset-0 z-30" role="dialog" aria-modal="true" @click.self="uiStore.closeDrawer()">
      <aside class="absolute top-0 right-0 h-full w-96 bg-gray-800 border-l border-gray-700 shadow-2xl p-4 flex flex-col">
        <div class="flex items-center justify-between mb-4 flex-shrink-0">
          <h2 class="text-xl font-semibold text-white">Правила игнорирования</h2>
          <button @click="uiStore.closeDrawer()" class="p-2 rounded-md hover:bg-gray-700" aria-label="Close">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>

        <div class="flex-grow flex flex-col min-h-0 space-y-3">
          <p class="text-xs text-gray-400">
            Используйте glob-паттерны для исключения файлов и папок. Каждое правило с новой строки.
          </p>
          <textarea
            v-model="settingsStore.settings.customIgnoreRules"
            class="w-full flex-grow bg-gray-900 border border-gray-600 rounded-md p-3 font-mono text-xs resize-none"
            placeholder="node_modules/&#10;**/*.log&#10;dist/"
          ></textarea>

          <div class="mt-2">
            <button @click="toggleGitignore" class="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 rounded-md">
              {{ showGitignore ? 'Скрыть .gitignore' : 'Показать .gitignore' }}
            </button>
            <div v-if="showGitignore" class="mt-2 bg-gray-900 border border-gray-700 rounded-md p-2 overflow-auto max-h-40">
              <pre class="text-[11px] text-gray-300 whitespace-pre-wrap">{{ gitignoreText || 'Нет данных .gitignore' }}</pre>
            </div>
          </div>
        </div>

        <div class="mt-4 flex-shrink-0">
          <button @click="handleSave" class="w-full py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50" :disabled="settingsStore.isLoading">
            {{ settingsStore.isLoading ? "Сохранение..." : "Сохранить и применить" }}
          </button>
        </div>
      </aside>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from "vue";
import { useUiStore } from "@/stores/ui.store";
import { useSettingsStore } from "@/stores/settings.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useProjectStore } from "@/stores/project.store";
import { apiService } from "@/services/api.service";

const uiStore = useUiStore();
const settingsStore = useSettingsStore();
const fileTreeStore = useFileTreeStore();
const projectStore = useProjectStore();

const showGitignore = ref(false);
const gitignoreText = ref("");

async function loadGitignore() {
  const root = projectStore.currentProject?.path || "";
  gitignoreText.value = root ? await apiService.getGitignoreContent(root) : "";
}
function toggleGitignore() {
  showGitignore.value = !showGitignore.value;
  if (showGitignore.value) loadGitignore();
}
async function handleSave() {
  await settingsStore.saveIgnoreSettings();
  uiStore.closeDrawer();
  await fileTreeStore.fetchFileTree();
}
watch(() => uiStore.activeDrawer, (v) => { if (v === "ignore" && showGitignore.value) loadGitignore(); });

function onKeydown(e: KeyboardEvent) { if (e.key === "Escape") uiStore.closeDrawer(); }
onMounted(() => document.addEventListener("keydown", onKeydown));
onUnmounted(() => document.removeEventListener("keydown", onKeydown));
</script>

<style scoped>
.slide-fade-enter-active,.slide-fade-leave-active{ transition: transform 0.22s cubic-bezier(0.165,0.84,0.44,1); }
.slide-fade-enter-from,.slide-fade-leave-to{ transform: translateX(100%); }
</style>
