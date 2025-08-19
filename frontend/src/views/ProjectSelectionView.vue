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
          @click="openProject"
          :disabled="projectStore.isLoading"
          class="w-full p-4 bg-blue-600 hover:bg-blue-500 disabled:bg-blue-800 disabled:cursor-not-allowed rounded-lg text-white font-semibold transition-colors"
        >
          <div class="flex items-center justify-center gap-2">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            {{ projectStore.isLoading ? 'Opening...' : 'Open Project' }}
          </div>
        </button>

        <!-- Recent Projects -->
        <div v-if="projectStore.hasRecentProjects" class="space-y-2">
          <h2 class="text-lg font-semibold text-white">Recent Projects</h2>
          <div class="space-y-2">
            <div
              v-for="project in projectStore.recentProjects"
              :key="project.path"
              @click="openRecentProject(project)"
              class="p-3 bg-gray-800 hover:bg-gray-700 rounded-lg cursor-pointer transition-colors group"
            >
              <div class="flex items-center justify-between">
                <div class="flex-1 min-w-0">
                  <h3 class="text-white font-medium truncate">{{ project.name }}</h3>
                  <p class="text-sm text-gray-400 truncate">{{ project.path }}</p>
                  <p v-if="project.lastOpened" class="text-xs text-gray-500">
                    Last opened: {{ formatDate(project.lastOpened) }}
                  </p>
                </div>
                <div class="flex items-center gap-2 ml-4">
                  <button
                    @click.stop="removeProject(project.path)"
                    class="p-1 text-gray-400 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                    title="Remove from recent"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
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
              type="checkbox"
              v-model="projectStore.autoOpenLastProject"
              @change="projectStore.toggleAutoOpenLastProject"
              class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500"
            />
            <span>Automatically open last project on startup</span>
          </label>
        </div>

        <!-- Error Display -->
        <div v-if="projectStore.error" class="p-3 bg-red-900/50 border border-red-700 rounded-lg">
          <p class="text-red-300 text-sm">{{ projectStore.error }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project.store'

const router = useRouter()
const projectStore = useProjectStore()

onMounted(async () => {
  // Initialize store
  projectStore.initialize()
  
  // Try to auto-open last project
  if (projectStore.autoOpenLastProject) {
    const opened = await projectStore.tryAutoOpenLastProject()
    if (opened) {
      router.push('/workspace')
    }
  }
})

async function openProject() {
  const success = await projectStore.openProject()
  if (success) {
    router.push('/workspace')
  }
}

async function openRecentProject(project: any) {
  const success = await projectStore.openRecentProject(project)
  if (success) {
    router.push('/workspace')
  }
}

function removeProject(path: string) {
  projectStore.removeRecent(path)
}

function formatDate(dateString: string): string {
  try {
    const date = new Date(dateString)
    const now = new Date()
    const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60)
    
    if (diffInHours < 1) {
      return 'Just now'
    } else if (diffInHours < 24) {
      const hours = Math.floor(diffInHours)
      return `${hours} hour${hours > 1 ? 's' : ''} ago`
    } else if (diffInHours < 24 * 7) {
      const days = Math.floor(diffInHours / 24)
      return `${days} day${days > 1 ? 's' : ''} ago`
    } else {
      return date.toLocaleDateString()
    }
  } catch {
    return 'Unknown'
  }
}
</script>