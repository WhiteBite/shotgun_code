import { defineStore } from 'pinia'
import { ref, computed, readonly } from 'vue'
import type { Project } from '@/types/api'
import { apiService } from '@/services/api.service'
import { useFileTreeStore } from '@/stores/file-tree.store'

export const useProjectStore = defineStore('project', () => {
  const fileTreeStore = useFileTreeStore()
  
  const currentProject = ref<Project | null>(null)
  const recentProjects = ref<Project[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const autoOpenLastProject = ref(true) // New setting

  const isProjectLoaded = computed(() => currentProject.value !== null)
  const hasRecentProjects = computed(() => recentProjects.value.length > 0)

  // Load settings from localStorage
  function loadSettings() {
    try {
      const settings = localStorage.getItem('project-settings')
      if (settings) {
        const parsed = JSON.parse(settings)
        autoOpenLastProject.value = parsed.autoOpenLastProject ?? true
      }
    } catch (err) {
      console.warn('Failed to load project settings:', err)
    }
  }

  // Save settings to localStorage
  function saveSettings() {
    try {
      const settings = {
        autoOpenLastProject: autoOpenLastProject.value
      }
      localStorage.setItem('project-settings', JSON.stringify(settings))
    } catch (err) {
      console.warn('Failed to save project settings:', err)
    }
  }

  // Load recent projects from localStorage
  function loadRecentProjects() {
    try {
      const stored = localStorage.getItem('recent-projects')
      if (stored) {
        const projects = JSON.parse(stored) as Project[]
        // Validate and filter out invalid projects
        recentProjects.value = projects.filter((project: Project) => 
          project && 
          typeof project.name === 'string' && 
          typeof project.path === 'string' &&
          project.name.trim() !== '' &&
          project.path.trim() !== ''
        )
      }
    } catch (err) {
      console.warn('Failed to load recent projects:', err)
      recentProjects.value = []
    }
  }

  // Save recent projects to localStorage
  function saveRecentProjects() {
    try {
      localStorage.setItem('recent-projects', JSON.stringify(recentProjects.value))
    } catch (err) {
      console.warn('Failed to save recent projects:', err)
    }
  }

  // Add project to recent list
  function addToRecent(project: Project) {
    // Remove if already exists
    recentProjects.value = recentProjects.value.filter((p: Project) => p.path !== project.path)
    
    // Add to beginning
    recentProjects.value.unshift(project)
    
    // Keep only last 10 projects
    if (recentProjects.value.length > 10) {
      recentProjects.value = recentProjects.value.slice(0, 10)
    }
    
    saveRecentProjects()
  }

  // Open project (with auto-open last project support)
  async function openProject(): Promise<boolean> {
    isLoading.value = true
    error.value = null
    
    try {
      const selectedPath = await apiService.selectDirectory()
      if (!selectedPath) {
        return false
      }

      const project: Project = {
        name: selectedPath.split(/[/\\]/).pop() || 'Unknown Project',
        path: selectedPath,
        lastOpened: new Date().toISOString()
      }

      await setCurrentProject(project)
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to open project'
      return false
    } finally {
      isLoading.value = false
    }
  }

  // Set current project
  async function setCurrentProject(project: Project): Promise<boolean> {
    try {
      // Update last opened timestamp
      project.lastOpened = new Date().toISOString()
      
      currentProject.value = project
      addToRecent(project)
      
      // Save to localStorage immediately
      saveRecentProjects()
      
      // Load project files
      await fileTreeStore.loadProject(project.path)
      
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to set current project'
      return false
    }
  }

  // Open recent project by clicking on it (not button)
  async function openRecentProject(project: Project): Promise<boolean> {
    try {
      // Verify project still exists
      const exists = await apiService.verifyProjectPath(project.path)
      if (!exists) {
        removeRecent(project.path)
        error.value = 'Project path no longer exists'
        return false
      }

      await setCurrentProject(project)
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to open recent project'
      return false
    }
  }

  // Remove project from recent list
  function removeRecent(path: string) {
    recentProjects.value = recentProjects.value.filter((p: Project) => p.path !== path)
    saveRecentProjects()
  }

  // Clear current project
  function clearCurrentProject() {
    currentProject.value = null
    fileTreeStore.clearProject()
  }

  // Toggle auto-open last project setting
  function toggleAutoOpenLastProject() {
    autoOpenLastProject.value = !autoOpenLastProject.value
    saveSettings()
  }

  // Auto-open last project if setting is enabled
  async function tryAutoOpenLastProject(): Promise<boolean> {
    if (!autoOpenLastProject.value || recentProjects.value.length === 0) {
      return false
    }

    const lastProject = recentProjects.value[0]
    try {
      const exists = await apiService.verifyProjectPath(lastProject.path)
      if (exists) {
        await setCurrentProject(lastProject)
        return true
      } else {
        // Remove invalid project
        removeRecent(lastProject.path)
        return false
      }
    } catch (err) {
      console.warn('Failed to auto-open last project:', err)
      return false
    }
  }

  // Initialize store
  function initialize() {
    loadSettings()
    loadRecentProjects()
  }

  return {
    // State
    currentProject: readonly(currentProject),
    recentProjects: readonly(recentProjects),
    isLoading: readonly(isLoading),
    error: readonly(error),
    autoOpenLastProject,
    
    // Computed
    isProjectLoaded,
    hasRecentProjects,
    
    // Methods
    openProject,
    setCurrentProject,
    openRecentProject,
    removeRecent,
    clearCurrentProject,
    toggleAutoOpenLastProject,
    tryAutoOpenLastProject,
    initialize
  }
}, {
  persist: {
    key: 'project-store',
    storage: localStorage,
    paths: ['recentProjects', 'autoOpenLastProject']
  }
})