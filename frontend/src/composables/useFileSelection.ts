/**
 * useFileSelection - File selection state and operations
 * Manages selected files, recursive selection, and selection stats
 */

import type { FileNode } from '@/types/domain'
import { computed, shallowRef, triggerRef } from 'vue'

export interface UseFileSelectionOptions {
    findNode: (path: string) => FileNode | null
    getAllFilesInNode: (node: FileNode) => string[]
}

export function useFileSelection(options: UseFileSelectionOptions) {
    const { findNode, getAllFilesInNode } = options

    // State - Use shallowRef to avoid deep reactivity on large sets
    const selectedPaths = shallowRef<Set<string>>(new Set())

    // Computed
    const hasSelectedFiles = computed(() => selectedPaths.value.size > 0)
    const selectedCount = computed(() => selectedPaths.value.size)
    const selectedFilesList = computed(() => Array.from(selectedPaths.value))

    // Actions
    function toggleSelect(path: string) {
        const node = findNode(path)
        if (!node) return

        if (node.isDir) {
            toggleSelectRecursive(path)
        } else {
            if (selectedPaths.value.has(path)) {
                selectedPaths.value.delete(path)
            } else {
                selectedPaths.value.add(path)
            }
        }
        triggerRef(selectedPaths)
    }

    function selectPath(path: string) {
        selectedPaths.value.add(path)
        triggerRef(selectedPaths)
    }

    function deselectPath(path: string) {
        selectedPaths.value.delete(path)
        triggerRef(selectedPaths)
    }

    function selectMultiple(paths: string[]) {
        paths.forEach((p) => selectedPaths.value.add(p))
        triggerRef(selectedPaths)
    }

    function clearSelection() {
        selectedPaths.value.clear()
        triggerRef(selectedPaths)
    }

    function selectRecursive(path: string) {
        const node = findNode(path)
        if (!node) return

        const filePaths = getAllFilesInNode(node)
        filePaths.forEach((p) => selectedPaths.value.add(p))
        triggerRef(selectedPaths)
    }

    function deselectRecursive(path: string) {
        const node = findNode(path)
        if (!node) return

        const filePaths = getAllFilesInNode(node)
        filePaths.forEach((p) => selectedPaths.value.delete(p))
        triggerRef(selectedPaths)
    }

    function toggleSelectRecursive(path: string) {
        const node = findNode(path)
        if (!node || !node.isDir) return

        const childFilePaths = getAllFilesInNode(node)
        const anySelected = childFilePaths.some((filePath) =>
            selectedPaths.value.has(filePath)
        )

        if (anySelected) {
            childFilePaths.forEach((filePath) => selectedPaths.value.delete(filePath))
        } else {
            childFilePaths.forEach((filePath) => selectedPaths.value.add(filePath))
        }
        triggerRef(selectedPaths)
    }

    function selectByExtension(
        extension: string,
        nodes: FileNode[],
        walkTree: (tree: FileNode[], fn: (node: FileNode) => void) => void
    ) {
        walkTree(nodes, (node) => {
            if (!node.isDir && node.name.endsWith(extension)) {
                selectedPaths.value.add(node.path)
            }
        })
        triggerRef(selectedPaths)
    }

    function getSelectedFileCountInNode(node: FileNode): number {
        const allFiles = getAllFilesInNode(node)
        return allFiles.filter((filePath) => selectedPaths.value.has(filePath)).length
    }

    function isSelected(path: string): boolean {
        return selectedPaths.value.has(path)
    }

    function getSelectionState(
        node: FileNode
    ): 'none' | 'partial' | 'full' {
        if (!node.isDir) {
            return selectedPaths.value.has(node.path) ? 'full' : 'none'
        }

        if (selectedCount.value === 0) {
            return 'none'
        }

        const allFiles = getAllFilesInNode(node)
        if (allFiles.length === 0) {
            return selectedPaths.value.has(node.path) ? 'full' : 'none'
        }

        let selectedFileCount = 0
        for (const filePath of allFiles) {
            if (selectedPaths.value.has(filePath)) {
                selectedFileCount++
            }
        }

        if (selectedFileCount === 0) return 'none'
        if (selectedFileCount === allFiles.length) return 'full'
        return 'partial'
    }

    return {
        // State
        selectedPaths,
        // Computed
        hasSelectedFiles,
        selectedCount,
        selectedFilesList,
        // Actions
        toggleSelect,
        selectPath,
        deselectPath,
        selectMultiple,
        clearSelection,
        selectRecursive,
        deselectRecursive,
        toggleSelectRecursive,
        selectByExtension,
        getSelectedFileCountInNode,
        isSelected,
        getSelectionState,
    }
}
