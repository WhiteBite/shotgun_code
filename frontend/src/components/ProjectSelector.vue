<template>
  <div class="flex items-center justify-center h-screen bg-gray-900 text-white">
    <div class="text-center max-w-md">
      <h1 class="text-4xl font-bold mb-4">Shotgun Code</h1>
      <p class="text-gray-400 mb-8">Select a project to get started</p>
      
      <button
        @click="selectProject"
        class="px-6 py-3 bg-blue-600 hover:bg-blue-700 rounded-lg font-semibold transition-colors"
      >
        Open Project Directory
      </button>

      <!-- Recent Projects -->
      <div v-if="recentProjects.length > 0" class="mt-8">
        <h2 class="text-lg font-semibold mb-4 text-gray-300">Recent Projects</h2>
        <div class="space-y-2">
          <button
            v-for="project in recentProjects"
            :key="project.path"
            @click="openRecentProject(project.path)"
            class="w-full px-4 py-2 bg-gray-800 hover:bg-gray-700 rounded text-left transition-colors"
          >
            <div class="font-medium">{{ project.name }}</div>
            <div class="text-sm text-gray-400 truncate">{{ project.path }}</div>
          </button>
        </div>
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
