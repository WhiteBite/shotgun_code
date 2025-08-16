<template>
  <header
      class="h-14 bg-gray-900/80 backdrop-blur-sm text-white flex items-center justify-between px-4 border-b border-gray-700 flex-shrink-0 z-40"
  >
    <div class="flex items-center gap-4 min-w-0">
      <div class="flex items-center gap-2 cursor-pointer" @click="goToSelection()">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-blue-500">
          <path d="m12 19-7-7 7-7" />
          <path d="m19 12-7 7" />
        </svg>
        <span class="font-bold text-lg">Shotgun</span>
      </div>
      <div v-if="projectStore.isProjectLoaded" class="w-px h-6 bg-gray-700"></div>
      <div v-if="projectStore.isProjectLoaded" class="flex items-center gap-2 min-w-0">
        <button class="px-3 py-1 bg-gray-700/50 rounded-md hover:bg-gray-700 text-sm">
          <span class="text-gray-400">Project:</span>
          <span class="font-semibold ml-1 truncate max-w-[180px]">{{ projectStore.currentProject?.name }}</span>
        </button>
        <button @click="changeProject" class="px-2 py-1 text-xs text-gray-400 hover:text-white hover:bg-gray-700 rounded-md" title="Change Project">
          Change
        </button>
      </div>
    </div>

    <div class="flex items-center gap-3">
      <button @click="uiStore.openDrawer('settings')" class="p-2 rounded-md hover:bg-gray-700" title="Settings">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="4" y1="21" x2="4" y2="14" /><line x1="4" y1="10" x2="4" y2="3" />
          <line x1="12" y1="21" x2="12" y2="12" /><line x1="12" y1="8" x2="12" y2="3" />
          <line x1="20" y1="21" x2="20" y2="16" /><line x1="20" y1="12" x2="20" y2="3" />
          <line x1="1" y1="14" x2="7" y2="14" /><line x1="9" y1="8" x2="15" y2="8" /><line x1="17" y1="16" x2="23" y2="16" />
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
  await projectStore.openProject();
}
function goToSelection() {
  router.push("/");
}
</script>
