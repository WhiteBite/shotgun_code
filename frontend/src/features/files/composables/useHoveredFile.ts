/**
 * Composable for managing hovered file state using provide/inject pattern.
 * Replaces global window.__hoveredFilePath and window.__hoveredFileIsDir.
 */

import { inject, provide, ref, type InjectionKey, type Ref } from 'vue'

export interface HoveredFileState {
    path: Ref<string | null>
    isDir: Ref<boolean | null>
    setHovered: (path: string | null, isDir: boolean | null) => void
    clearHovered: (currentPath: string) => void
}

export const HoveredFileKey: InjectionKey<HoveredFileState> = Symbol('HoveredFile')

/**
 * Provides hovered file state to child components.
 * Should be called in the root FileExplorer component.
 */
export function provideHoveredFile(): HoveredFileState {
    const path = ref<string | null>(null)
    const isDir = ref<boolean | null>(null)

    const setHovered = (newPath: string | null, newIsDir: boolean | null) => {
        path.value = newPath
        isDir.value = newIsDir
    }

    const clearHovered = (currentPath: string) => {
        if (path.value === currentPath) {
            path.value = null
            isDir.value = null
        }
    }

    const state: HoveredFileState = {
        path,
        isDir,
        setHovered,
        clearHovered,
    }

    provide(HoveredFileKey, state)
    return state
}

/**
 * Creates a fallback hovered file state for components not wrapped in provider.
 */
function createFallbackState(): HoveredFileState {
    const path = ref<string | null>(null)
    const isDir = ref<boolean | null>(null)

    return {
        path,
        isDir,
        setHovered: (newPath, newIsDir) => {
            path.value = newPath
            isDir.value = newIsDir
        },
        clearHovered: (currentPath) => {
            if (path.value === currentPath) {
                path.value = null
                isDir.value = null
            }
        },
    }
}

/**
 * Injects hovered file state from parent component.
 * Returns a fallback state if not provided (for backwards compatibility).
 */
export function useHoveredFile(): HoveredFileState {
    // Use inject with default value to avoid Vue warning
    return inject(HoveredFileKey, createFallbackState())
}
