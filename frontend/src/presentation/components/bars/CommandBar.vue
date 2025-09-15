<template>
  <header
    class="h-14 bg-gray-900/80 backdrop-blur-sm text-white flex items-center justify-between px-6 border-b border-gray-700 flex-shrink-0 z-40"
  >
    <div class="flex items-center gap-4 min-w-0">
      <div
        class="flex items-center gap-2 cursor-pointer"
        @click="goToSelection()"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          class="text-blue-500"
        >
          <path d="m12 19-7-7 7-7" />
          <path d="m19 12-7 7" />
        </svg>
        <span class="font-bold text-xl text-gradient-primary">Shotgun</span>
      </div>
      <div
        v-if="projectStore.isProjectLoaded"
        class="w-px h-6 bg-gray-700"
      ></div>
      <div
        v-if="projectStore.isProjectLoaded"
        class="flex items-center gap-2 min-w-0"
      >
        <button
          class="px-3 py-1 bg-gray-700/50 rounded-md hover:bg-gray-700 text-sm"
        >
          <span class="text-gray-400">Project:</span>
          <span class="font-semibold ml-1 truncate max-w-[180px]">{{
            projectStore.currentProject?.name
          }}</span>
        </button>
        <button
          class="px-2 py-1 text-xs text-gray-400 hover:text-white hover:bg-gray-700 rounded-md"
          title="Change Project"
          @click="changeProject"
        >
          Change
        </button>
      </div>
    </div>

    <div class="flex items-center gap-3 relative">
      <!-- Recent Projects Button -->
      <button
        v-if="projectStore.hasRecentProjects"
        class="p-2 rounded-md hover:bg-gray-700"
        title="Недавние проекты"
        @click="toggleRecent"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <path d="M3 3v18h18" />
          <path d="M18 17V9" />
          <path d="M13 17V5" />
          <path d="M8 17v-3" />
        </svg>
      </button>

      <!-- Recent Projects Popover -->
      <div v-if="showRecent" class="absolute right-12 top-12 bg-gray-800 border border-gray-700 rounded shadow-xl w-72 z-50">
        <div class="p-2 border-b border-gray-700 text-xs text-gray-400">Недавние проекты</div>
        <div class="max-h-64 overflow-auto">
          <div
            v-for="proj in projectStore.recentProjects"
            :key="proj.path"
            class="flex items-center justify-between px-3 py-2 text-sm hover:bg-gray-700/50"
          >
            <div class="min-w-0">
              <div class="font-semibold truncate">{{ proj.name }}</div>
              <div class="text-xs text-gray-400 truncate">{{ normalizePath(proj.path) }}</div>
            </div>
            <div class="flex items-center gap-2">
              <button class="px-2 py-1 bg-blue-600 hover:bg-blue-500 rounded text-xs" @click="openRecent(proj)">Открыть</button>
              <button class="px-2 py-1 bg-gray-600 hover:bg-gray-500 rounded text-xs" @click="removeRecent(proj.path)">Удалить</button>
            </div>
          </div>
        </div>
      </div>

      <button
        class="p-2 rounded-md hover:bg-gray-700"
        title="Settings"
        @click="uiStore.openDrawer('settings')"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <line x1="4" y1="21" x2="4" y2="14" />
          <line x1="4" y1="10" x2="4" y2="3" />
          <line x1="12" y1="21" x2="12" y2="12" />
          <line x1="12" y1="8" x2="12" y2="3" />
          <line x1="20" y1="21" x2="20" y2="16" />
          <line x1="20" y1="12" x2="20" y2="3" />
          <line x1="1" y1="14" x2="7" y2="14" />
          <line x1="9" y1="8" x2="15" y2="8" />
          <line x1="17" y1="16" x2="23" y2="16" />
        </svg>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useUiStore } from "@/stores/ui.store";
import { useProjectStore } from "@/stores/project.store";
import { useRouter } from "vue-router";

const uiStore = useUiStore();
const projectStore = useProjectStore();
const router = useRouter();

async function changeProject() {
  // Перенаправляем на экран выбора с флагом ручного выбора
  router.push({ path: "/", query: { manual: "1" } });
}

import { ref, onMounted, onUnmounted } from "vue";
function normalizePath(p: string): string { return p ? p.replace(/\\/g, '/') : p }
const showRecent = ref(false);
function toggleRecent() { showRecent.value = !showRecent.value; }
function onDocClick(e: MouseEvent) {
  const target = e.target as HTMLElement;
  if (!target.closest('header')) showRecent.value = false;
}
onMounted(() => document.addEventListener('click', onDocClick));
onUnmounted(() => document.removeEventListener('click', onDocClick));

async function openRecent(project: any) {
  const ok = await projectStore.openRecentProject(project);
  if (ok) {
    showRecent.value = false;
    router.push("/workspace");
  }
}
function removeRecent(path: string) { projectStore.removeRecent(path); }

function goToSelection() {
  router.push({ path: "/", query: { manual: "1" } });
}
</script>