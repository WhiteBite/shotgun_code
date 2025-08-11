<template>
  <div class="w-full h-full overflow-y-auto bg-gray-900 flex justify-center p-8">
    <div class="w-full max-w-3xl mx-auto">
      <div class="space-y-8">
        <ProjectList />
        <RecentTaskList />
      </div>

      <!-- This global loading overlay is no longer needed -->
      <!-- <div v-if="projectStore.isLoading" ... ></div> -->

      <div v-if="projectStore.recentProjects.length === 0"
           class="mt-8 p-6 bg-yellow-900/50 border border-yellow-700 text-yellow-300 rounded-lg text-center">
        <h3 class="font-semibold">Добро пожаловать!</h3>
        <p class="text-sm">Чтобы начать работу, выберите папку с вашим проектом.</p>
      </div>
      <div v-else class="mt-8">
        <button @click="projectStore.selectMostRecentProject()" class="w-full p-3 bg-gray-700/50 hover:bg-gray-700 rounded-md text-white font-semibold transition-colors text-center">
          Открыть последний проект: {{ projectStore.recentProjects[0].name }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import ProjectList from '../components/project/ProjectList.vue';
import RecentTaskList from '../components/project/RecentTaskList.vue';
import { useProjectStore } from '../stores/projectStore';

const projectStore = useProjectStore();
</script>