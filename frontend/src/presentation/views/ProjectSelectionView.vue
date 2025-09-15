<template>
  <div class="min-h-screen bg-gray-900 flex items-center justify-center p-4">
    <div class="w-full max-w-2xl">
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-white mb-2">Shotgun</h1>
        <p class="text-gray-400">Select a project to get started</p>
      </div>

      <div class="space-y-4">
        <!-- Open Project Button -->
        <button
          :disabled="projectStore.isLoading"
          class="w-full p-4 bg-blue-600 hover:bg-blue-500 disabled:bg-blue-800 disabled:cursor-not-allowed rounded-lg text-white font-semibold transition-colors"
          @click="openProject"
        >
          <div class="flex items-center justify-center gap-2">
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 6v6m0 0v6m0-6h6m-6 0H6"
              />
            </svg>
            {{ projectStore.isLoading ? "Opening..." : "Open Project" }}
          </div>
        </button>

        <!-- Recent Projects -->
        <div v-if="projectStore.hasRecentProjects" class="space-y-2">
          <h2 class="text-lg font-semibold text-white">Recent Projects</h2>
          <div class="space-y-2">
            <div
              v-for="project in projectStore.recentProjects"
              :key="project.path"
              class="p-3 bg-gray-800 hover:bg-gray-700 rounded-lg cursor-pointer transition-colors group"
              @click="openRecentProject(project)"
            >
              <div class="flex items-center justify-between">
                <div class="flex-1 min-w-0">
                  <h3 class="text-white font-medium truncate">
                    {{ project.name }}
                  </h3>
                  <p class="text-sm text-gray-400 truncate">
                    {{ normalizePath(project.path) }}
                  </p>
                  <p v-if="project.lastOpened" class="text-xs text-gray-500">
                    Last opened: {{ formatDate(project.lastOpened) }}
                  </p>
                </div>
                <div class="flex items-center gap-2 ml-4">
                  <button
                    class="p-1 text-gray-400 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                    title="Remove from recent"
                    @click.stop="removeProject(project.path)"
                  >
                    <svg
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Settings -->
        <div class="mt-8 p-4 bg-gray-800 rounded-lg">
          <h3 class="text-white font-medium mb-3">Settings</h3>
          <label class="flex items-center gap-2 text-gray-300 cursor-pointer">
            <input
              v-model="settingsStore.autoOpenLastProject"
              type="checkbox"
              class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500"
              @change="settingsStore.toggleAutoOpenLastProject"
            />
            <span>Automatically open last project on startup</span>
          </label>
        </div>

        <!-- Error Display -->
        <div
          v-if="projectStore.error"
          class="p-3 bg-red-900/50 border border-red-700 rounded-lg"
        >
          <p class="text-red-300 text-sm">{{ projectStore.error }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useProjectStore } from "@/stores/project.store";
import { useSettingsStore } from "@/stores/settings.store";

const router = useRouter();
const route = useRoute();
const projectStore = useProjectStore();
const settingsStore = useSettingsStore();
function normalizePath(p: string): string { return p ? p.replace(/\\/g, '/') : p }

// Навигация выполняется императивно после успешного открытия проекта

onMounted(async () => {
  console.log("ProjectSelectionView mounted");
  
  // Initialize store
  projectStore.initialize();
  console.log("Store initialized, isProjectLoaded:", projectStore.isProjectLoaded);

  // Try to auto-open last project
  const manualPick = route.query.manual === '1';
  if (settingsStore.autoOpenLastProject && !manualPick) {
    console.log("Trying to auto-open last project");
    await projectStore.tryAutoOpenLastProject();
    console.log("Auto-open result, isProjectLoaded:", projectStore.isProjectLoaded);
  }

  // If no recent projects and auto-open is disabled, try to load current directory
  if (!projectStore.hasRecentProjects && !manualPick) {
    console.log("No recent projects, trying to load current directory");
    await projectStore.tryLoadCurrentDirectory();
    console.log("Load current directory result, isProjectLoaded:", projectStore.isProjectLoaded);
  }
});

async function openProject() {
  console.log("Opening project...");
  const ok = await projectStore.openProject();
  console.log("Project opened, isProjectLoaded:", projectStore.isProjectLoaded);
  if (ok) {
    await router.push("/workspace");
  }
}

async function openRecentProject(project: any) {
  console.log("Opening recent project:", project.name);
  const ok = await projectStore.openRecentProject(project);
  console.log("Recent project opened, isProjectLoaded:", projectStore.isProjectLoaded);
  if (ok) {
    await router.push("/workspace");
  }
}

function removeProject(path: string) {
  projectStore.removeRecent(path);
}

function formatDate(dateString: string): string {
  try {
    const date = new Date(dateString);
    const now = new Date();
    const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);

    if (diffInHours < 1) {
      return "Just now";
    } else if (diffInHours < 24) {
      const hours = Math.floor(diffInHours);
      return `${hours} hour${hours > 1 ? "s" : ""} ago`;
    } else if (diffInHours < 24 * 7) {
      const days = Math.floor(diffInHours / 24);
      return `${days} day${days > 1 ? "s" : ""} ago`;
    } else {
      return date.toLocaleDateString();
    }
  } catch {
    return "Unknown";
  }
}
</script>