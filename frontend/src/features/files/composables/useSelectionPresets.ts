import { useProjectStore } from '@/stores/project.store'
import { computed, ref } from 'vue'
import { useFileStore } from '../model/file.store'

export interface SelectionPreset {
    id: string
    name: string
    description?: string
    paths: string[]
    createdAt: number
    projectPath: string
}

const STORAGE_KEY = 'file-selection-presets'

// Reactive state for presets
const presetsRef = ref<SelectionPreset[]>([])

// Initialize presets from localStorage
function initializePresets() {
    try {
        const saved = localStorage.getItem(STORAGE_KEY)
        presetsRef.value = saved ? JSON.parse(saved) : []
    } catch (err) {
        console.warn('Failed to load presets:', err)
        presetsRef.value = []
    }
}

// Sync reactive state with localStorage
function syncToStorage() {
    try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(presetsRef.value))
    } catch (err) {
        console.error('Failed to save presets:', err)
    }
}

// Initialize on module load
initializePresets()

export function useSelectionPresets() {
    const fileStore = useFileStore()
    const projectStore = useProjectStore()

    // Computed property for project-specific presets
    const projectPresets = computed(() => {
        if (!projectStore.currentPath) return []
        return presetsRef.value.filter(p => p.projectPath === projectStore.currentPath)
    })

    function refreshPresets() {
        initializePresets()
    }

    function getAllPresets(): SelectionPreset[] {
        return presetsRef.value
    }

    function getProjectPresets(): SelectionPreset[] {
        return projectPresets.value
    }

    function getPresetById(id: string): SelectionPreset | undefined {
        return presetsRef.value.find(p => p.id === id)
    }

    function validatePresetPaths(preset: SelectionPreset): { valid: string[]; invalid: string[] } {
        const existingPaths = new Set(
            fileStore.flattenedNodes
                .filter(n => !n.isDir)
                .map(n => n.path)
        )
        const valid: string[] = []
        const invalid: string[] = []

        for (const path of preset.paths) {
            if (existingPaths.has(path)) {
                valid.push(path)
            } else {
                invalid.push(path)
            }
        }

        return { valid, invalid }
    }

    function savePreset(name: string, description?: string): SelectionPreset | null {
        if (!projectStore.currentPath || !fileStore.hasSelectedFiles) return null

        const preset: SelectionPreset = {
            id: `preset-${Date.now()}`,
            name,
            description,
            paths: fileStore.selectedFilesList,
            createdAt: Date.now(),
            projectPath: projectStore.currentPath
        }

        presetsRef.value.push(preset)
        syncToStorage()

        return preset
    }

    function loadPreset(presetId: string) {
        const preset = presetsRef.value.find(p => p.id === presetId)

        if (preset) {
            fileStore.clearSelection()
            fileStore.selectMultiple(preset.paths)
            return true
        }
        return false
    }

    function deletePreset(presetId: string) {
        presetsRef.value = presetsRef.value.filter(p => p.id !== presetId)
        syncToStorage()
    }

    function updatePreset(presetId: string, updates: Partial<SelectionPreset>) {
        const index = presetsRef.value.findIndex(p => p.id === presetId)

        if (index !== -1) {
            presetsRef.value[index] = { ...presetsRef.value[index], ...updates }
            syncToStorage()
            return true
        }
        return false
    }

    return {
        presets: presetsRef,
        projectPresets,
        getAllPresets,
        getProjectPresets,
        getPresetById,
        refreshPresets,
        savePreset,
        loadPreset,
        deletePreset,
        updatePreset,
        validatePresetPaths
    }
}
