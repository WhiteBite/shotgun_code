<template>
  <div class="flex items-center justify-center h-screen bg-gray-900 text-white">
    <div class="w-full max-w-2xl px-4">
      <div class="text-center">
      <h1 class="text-4xl font-bold mb-4">Shotgun Code</h1>
      <p class="text-gray-400 mb-8">Select a project to get started</p>
      
      <button
        @click="selectProject"
        class="px-6 py-3 bg-blue-600 hover:bg-blue-700 rounded-lg font-semibold transition-colors"
      >
        Open Project Directory
      </button>
      </div>

      <!-- Recent Projects -->
      <div class="mt-10">
        <h2 class="text-lg font-semibold mb-4 text-gray-300">Recent Projects</h2>
        <div class="space-y-2">
          <template v-if="recentProjects.length > 0">
            <button
              v-for="project in recentProjects"
              :key="project.path"
              @click="openRecentProject(project.path)"
              class="w-full px-4 py-2 bg-gray-800 hover:bg-gray-700 rounded text-left transition-colors"
            >
              <div class="font-medium">{{ project.name }}</div>
              <div class="text-sm text-gray-400 truncate">{{ project.path }}</div>
            </button>
          </template>
          <template v-else>
            <div v-for="i in 6" :key="i" class="w-full h-12 bg-gray-800/60 rounded"></div>
          </template>
        </div>
      </div>

      <!-- Settings -->
      <div class="mt-10 bg-gray-800/60 border border-gray-700 rounded-lg p-4">
        <h3 class="font-semibold text-gray-300 mb-3">Settings</h3>
        <label class="flex items-center gap-3 text-sm text-gray-300 select-none">
          <input
            type="checkbox"
            class="w-4 h-4 rounded border-gray-600 bg-gray-800 text-blue-500 focus:ring-0"
            :checked="projectStore.autoOpenLast"
            @change="onToggleAutoOpen"
          />
          Automatically open last project on startup
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useProjectStore } from '../stores/project.store'
import { useUIStore } from '../stores/ui.store'

const emit = defineEmits<{
  (e: 'opened', path: string): void
}>()

const projectStore = useProjectStore()
const uiStore = useUIStore()

const recentProjects = computed(() => projectStore.recentProjects)

function onToggleAutoOpen(e: Event) {
  const target = e.target as HTMLInputElement
  projectStore.setAutoOpenLast(target.checked)
}

async function selectProject() {
  try {
    // Create file input for directory selection
    const input = document.createElement('input')
    input.type = 'file'
    input.webkitdirectory = true
    input.multiple = true
    
    input.onchange = async (e: Event) => {
      const target = e.target as HTMLInputElement
      if (target.files && target.files.length > 0) {
        // Get directory path from first file using webkitRelativePath
        const file = target.files[0]
        // Extract directory name from webkitRelativePath (format: "dirname/filename")
        const relativePath = (file as any).webkitRelativePath as string
        if (relativePath) {
          const dirPath = relativePath.split('/')[0]
          const success = await projectStore.openProjectByPath(dirPath)
          if (success && projectStore.currentPath) {
            emit('opened', projectStore.currentPath)
          }
        } else {
          uiStore.addToast('Failed to detect directory path', 'error')
        }
      }
    }
    
    input.click()
  } catch (error) {
    console.error('Failed to select project:', error)
    uiStore.addToast('Failed to open project', 'error')
  }
}

async function openRecentProject(path: string) {
  try {
    await projectStore.openProjectByPath(path)
    emit('opened', path)
  } catch (error) {
    console.error('Failed to open recent project:', error)
    uiStore.addToast('Failed to open project', 'error')
  }
}
</script>
