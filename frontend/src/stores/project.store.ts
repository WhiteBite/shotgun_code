/**
 * Simple Project Store - No Clean Architecture dependencies
 * Direct localStorage integration for simplicity
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface RecentProject {
  path: string
  name: string
  lastOpened: number
}

const RECENT_PROJECTS_KEY = 'shotgun_recent_projects'
const MAX_RECENT = 10

export const useProjectStore = defineStore('project', () => {
  // State
  const currentPath = ref<string | null>(null)
  const currentName = ref<string | null>(null)
  const recentProjects = ref<RecentProject[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const hasProject = computed(() => currentPath.value !== null)
  const projectName = computed(() => currentName.value || '')
  const projectPath = computed(() => currentPath.value || '')
  const hasRecentProjects = computed(() => recentProjects.value.length > 0)

  // Actions
  async function openProjectByPath(path: string): Promise<boolean> {
    isLoading.value = true
    error.value = null

    try {
      // Simple: just set the path and name
      currentPath.value = path
      currentName.value = path.split(/[\\/]/).pop() || path

      // Update recent projects
      addToRecent(path)
      saveRecentProjects()

      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load project'
      return false
    } finally {
      isLoading.value = false
    }
  }

  function addToRecent(path: string) {
    const name = path.split(/[\\/]/).pop() || path
    const existing = recentProjects.value.findIndex(p => p.path === path)

    if (existing !== -1) {
      recentProjects.value.splice(existing, 1)
    }

    recentProjects.value.unshift({
      path,
      name,
      lastOpened: Date.now()
    })

    if (recentProjects.value.length > MAX_RECENT) {
      recentProjects.value = recentProjects.value.slice(0, MAX_RECENT)
    }
  }

  function loadRecentProjects() {
    try {
      const stored = localStorage.getItem(RECENT_PROJECTS_KEY)
      if (stored) {
        recentProjects.value = JSON.parse(stored)
      }
    } catch (err) {
      console.warn('Failed to load recent projects:', err)
    }
  }

  function saveRecentProjects() {
    try {
      localStorage.setItem(RECENT_PROJECTS_KEY, JSON.stringify(recentProjects.value))
    } catch (err) {
      console.warn('Failed to save recent projects:', err)
    }
  }

  function removeFromRecent(path: string) {
    recentProjects.value = recentProjects.value.filter(p => p.path !== path)
    saveRecentProjects()
  }

  function clearProject() {
    currentPath.value = null
    currentName.value = null
    error.value = null
  }

  // Initialize
  loadRecentProjects()

  return {
    // State
    currentPath,
    currentName,
    recentProjects,
    isLoading,
    error,
    // Computed
    hasProject,
    projectName,
    projectPath,
    hasRecentProjects,
    // Actions
    openProjectByPath,
    removeFromRecent,
    clearProject
  }
})