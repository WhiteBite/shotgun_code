<template>
  <div 
    class="flex items-center justify-center h-screen bg-gray-900 text-white"
    @drop.prevent="handleDrop"
    @dragover.prevent="isDragging = true"
    @dragleave.prevent="isDragging = false"
  >
    <!-- Drag & Drop Overlay -->
    <Transition name="fade">
      <div
        v-if="isDragging"
        class="fixed inset-0 bg-blue-600/20 border-4 border-dashed border-blue-500 rounded-xl flex items-center justify-center z-50 m-8 pointer-events-none"
      >
        <div class="text-center">
          <svg class="h-20 w-20 text-blue-400 mx-auto mb-4 animate-bounce" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <p class="text-3xl font-bold text-white drop-shadow-lg">Drop folder here</p>
          <p class="text-lg text-blue-200 mt-2">To open as project</p>
        </div>
      </div>
    </Transition>

    <div class="w-full max-w-2xl px-4">
      <div class="text-center">
      <h1 class="text-4xl font-bold mb-4">Shotgun Code</h1>
      <p class="text-gray-400 mb-2">Select a project to get started</p>
      <p class="text-sm text-gray-500 mb-8">or drag & drop a folder here</p>
      
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
import { computed, ref, onMounted } from 'vue'
import { useProjectStore } from '../stores/project.store'
import { useUIStore } from '../stores/ui.store'

const emit = defineEmits<{
  (e: 'opened', path: string): void
}>()

const projectStore = useProjectStore()
const uiStore = useUIStore()
const isDragging = ref(false)

const recentProjects = computed(() => projectStore.recentProjects)

// Force reload recent projects when component mounts
onMounted(() => {
  // Recent projects should already be loaded by store initialization
  // But if not, we can trigger a re-render by accessing the computed
  console.log('Recent projects loaded:', recentProjects.value.length)
})

function onToggleAutoOpen(e: Event) {
  const target = e.target as HTMLInputElement
  projectStore.setAutoOpenLast(target.checked)
}

// Drag & Drop handlers
async function handleDrop(e: DragEvent) {
  isDragging.value = false
  
  const items = e.dataTransfer?.items
  if (!items) return
  
  // Try to find a directory in the dragged items
  for (let i = 0; i < items.length; i++) {
    const item = items[i]
    if (item.kind === 'file') {
      // Get file entry
      const entry = item.webkitGetAsEntry?.()
      if (entry?.isDirectory) {
        // @ts-ignore - fullPath exists on FileSystemEntry
        const path = entry.fullPath
        console.log('Dropped folder:', path)
        
        // Open the project
        const success = await projectStore.openProjectByPath(path)
        if (success) {
          emit('opened', path)
        }
        return
      }
    }
  }
  
  // If no directory found, show error
  uiStore.addToast('Please drop a folder, not a file', 'warning')
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
