<template>
  <div class="bg-gray-800/50 p-6 rounded-lg border border-gray-700">
    <h2 class="text-xl font-semibold mb-4 text-white">Проекты</h2>
    <div class="space-y-3">
      <!-- Кнопка выбора папки -->
      <button
        @click="projectStore.selectProjectFromDialog()"
        class="w-full flex items-center justify-center gap-2 p-3 bg-blue-600/80 hover:bg-blue-600 rounded-md text-white font-semibold transition-colors"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
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
        Выбрать папку проекта
      </button>

      <div class="w-full border-t border-gray-700 my-4"></div>

      <!-- Список недавних проектов -->
      <h3 class="text-sm font-semibold text-gray-400 pb-1">Недавние проекты</h3>
      <div v-if="projectStore.recentProjects.length > 0" class="space-y-2">
        <div
          v-for="proj in projectStore.recentProjects"
          :key="proj.id"
          @click="projectStore.selectProject(proj)"
          class="p-3 bg-gray-700/50 hover:bg-gray-700/80 rounded-md cursor-pointer border border-transparent hover:border-blue-500 transition-all"
        >
          <p class="font-semibold text-white">{{ proj.name }}</p>
          <p class="text-xs text-gray-400">{{ proj.path }}</p>
        </div>
      </div>
      <div v-else class="p-4 text-center text-gray-500">
        Нет недавних проектов.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useProjectStore } from "@/stores/project.store";
const projectStore = useProjectStore();
</script>
