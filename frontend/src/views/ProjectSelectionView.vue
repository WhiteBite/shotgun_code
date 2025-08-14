<template>
  <div
    class="w-full h-full overflow-y-auto bg-gray-900 flex justify-center p-8"
  >
    <div class="w-full max-w-xl mx-auto space-y-8">
      <div>
        <h1 class="text-3xl font-bold text-center mb-2 text-white">
          Select a Project
        </h1>
        <p class="text-center text-gray-400">
          Choose a recent project or open a new folder to get started.
        </p>
      </div>

      <div
        class="bg-gray-800/50 p-6 rounded-lg border border-gray-700 space-y-4"
      >
        <button
          @click="projectStore.selectProjectFromDialog()"
          class="w-full flex items-center justify-center gap-3 p-4 bg-blue-600 hover:bg-blue-500 rounded-lg text-white font-semibold text-lg transition-colors shadow-lg"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path
              d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"
            ></path>
          </svg>
          Open Project Folder
        </button>

        <div v-if="projectStore.recentProjects.length > 0">
          <div class="w-full border-t border-gray-700 my-4"></div>
          <h3 class="text-sm font-semibold text-gray-400 mb-2 px-1">
            Recent Projects
          </h3>
          <div class="space-y-2">
            <div
              v-for="proj in projectStore.recentProjects"
              :key="proj.id"
              @click="projectStore.selectProject(proj)"
              class="p-3 bg-gray-700/50 hover:bg-gray-700 rounded-md cursor-pointer border border-transparent hover:border-blue-500 transition-all"
            >
              <p class="font-semibold text-white">{{ proj.name }}</p>
              <p class="text-xs text-gray-400">{{ proj.path }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useProjectStore } from "@/stores/project.store";
const projectStore = useProjectStore();
</script>
