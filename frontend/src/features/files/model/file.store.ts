import type { domain } from '#wailsjs/go/models'
import { useSettingsStore } from '@/stores/settings.store'
import Fuse from 'fuse.js'
import { defineStore } from 'pinia'
import { computed, ref, shallowRef, triggerRef } from 'vue'
import { filesApi } from '../api/files.api'

export interface FileNode {
    name: string
    path: string
    isDir: boolean
    isExpanded?: boolean
    isSelected?: boolean
    children?: FileNode[]
    size?: number
    isIgnored?: boolean
}

export const useFileStore = defineStore('file', () => {
    // State
    const nodes = ref<FileNode[]>([])
    const selectedPaths = shallowRef<Set<string>>(new Set()) // Use shallowRef to avoid deep reactivity
    const isLoading = ref(false)
    const error = ref<string | null>(null)
    const searchQuery = ref('')
    const filterExtensions = ref<string[]>([])
    const currentDirectory = ref<string>('')
    const directoryHistory = ref<string[]>([])
    const rootPath = ref<string>('')

    // Debounce timers
    let saveExpandedStateTimer: ReturnType<typeof setTimeout> | null = null
    let saveSelectionTimer: ReturnType<typeof setTimeout> | null = null

    // Cache for getAllFilesInNode - NOT REACTIVE to avoid Reactivity Storm
    // Using plain Map instead of ref to prevent Vue from creating proxies for 5000+ entries
    const allFilesCache = new Map<string, string[]>()

    // Computed
    const hasSelectedFiles = computed(() => selectedPaths.value.size > 0)
    const selectedCount = computed(() => selectedPaths.value.size)
    const selectedFilesList = computed(() => Array.from(selectedPaths.value))

    // Estimated token count for selected files (1 token ≈ 4 bytes)
    const estimatedTokenCount = computed(() => {
        let totalSize = 0
        selectedPaths.value.forEach(path => {
            const node = findNode(nodes.value, path)
            if (node && !node.isDir && node.size) {
                totalSize += node.size
            }
        })
        return Math.round(totalSize / 4)
    })

    // Estimated context size in MB
    const estimatedContextSize = computed(() => {
        let totalSize = 0
        selectedPaths.value.forEach(path => {
            const node = findNode(nodes.value, path)
            if (node && !node.isDir && node.size) {
                totalSize += node.size
            }
        })
        return totalSize / (1024 * 1024)
    })

    // Get total size of selected files in bytes
    function getSelectedFilesSize(): number {
        let totalSize = 0
        selectedPaths.value.forEach(path => {
            const node = findNode(nodes.value, path)
            if (node && !node.isDir && node.size) {
                totalSize += node.size
            }
        })
        return totalSize
    }

    const breadcrumbs = computed(() => {
        if (!currentDirectory.value || !rootPath.value) return []

        const relative = currentDirectory.value.replace(rootPath.value, '').replace(/^[\\/]+/, '')
        if (!relative) return [projectName.value]

        const segments = relative.split(/[\\/]/)
        return [projectName.value, ...segments]
    })

    const projectName = computed(() => {
        if (!rootPath.value) return 'Project'
        return rootPath.value.split(/[\\/]/).pop() || 'Project'
    })



    // Lazy flattened tree cache (only compute when needed)
    const flattenedNodesCache = ref<FileNode[] | null>(null)
    const flattenedNodes = computed(() => {
        if (nodes.value.length === 0) return []
        if (!flattenedNodesCache.value) {
            flattenedNodesCache.value = generateFlattened(nodes.value, 5) // Max depth 5
        }
        return flattenedNodesCache.value
    })

    // Search (optimized with lazy flattening and reduced limits)
    const searchResults = computed(() => {
        if (!searchQuery.value) return []

        const allFiles = flattenedNodes.value

        // For large trees, limit Fuse.js usage
        if (allFiles.length > 2000) {
            // Simple string matching for large trees
            return allFiles.filter(file =>
                file.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                file.path.toLowerCase().includes(searchQuery.value.toLowerCase())
            ).slice(0, 100) // Limit results
        }

        const fuse = new Fuse(allFiles, {
            keys: ['name', 'path'],
            threshold: 0.3
        })

        return fuse.search(searchQuery.value).map(result => result.item).slice(0, 100)
    })

    // Filtered nodes
    const filteredNodes = computed(() => {
        if (filterExtensions.value.length === 0) return nodes.value
        return filterTreeByExtensions(nodes.value, filterExtensions.value)
    })

    // Actions
    function setFileTree(tree: FileNode[] | domain.FileNode[]) {
        nodes.value = convertDomainNodes(tree)
        // Clear cache when tree changes (direct access, no .value)
        allFilesCache.clear()
        flattenedNodesCache.value = null
    }

    async function loadFileTree(projectPath: string, directory?: string) {
        isLoading.value = true
        error.value = null

        try {
            const targetPath = directory || projectPath
            const files = await filesApi.listFiles(targetPath, false, true)
            setFileTree(files)

            // Set root path on first load
            if (!rootPath.value) {
                rootPath.value = projectPath
                currentDirectory.value = projectPath

                // Load expanded state or auto-expand root folders
                loadExpandedState()

                // If no saved state, auto-expand root folders
                const hasExpandedState = nodes.value.some(n => n.isDir && n.isExpanded)
                if (!hasExpandedState) {
                    nodes.value.forEach(node => {
                        if (node.isDir) {
                            node.isExpanded = true
                        }
                    })
                }

                // Load saved selection for this project
                loadSelectionFromStorage(projectPath)
            } else if (directory) {
                currentDirectory.value = directory
            }
        } catch (err) {
            error.value = err instanceof Error ? err.message : 'Failed to load files'
            throw err
        } finally {
            isLoading.value = false
        }
    }

    function toggleSelect(path: string) {
        const node = findNode(nodes.value, path)
        if (!node) return

        if (node.isDir) {
            toggleSelectRecursive(path)
            // Trigger reactivity manually for shallowRef
            triggerRef(selectedPaths)
        } else {
            if (selectedPaths.value.has(path)) {
                selectedPaths.value.delete(path)
            } else {
                selectedPaths.value.add(path)
            }
            // Trigger reactivity manually for shallowRef
            triggerRef(selectedPaths)
        }

        // Auto-save selection with debounce
        if (autoSaveSelection.value) {
            debouncedSaveSelection()
        }

        // Note: We don't clear allFilesCache here because selection doesn't change the tree structure
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
        paths.forEach(p => selectedPaths.value.add(p))
        triggerRef(selectedPaths)
    }

    function clearSelection() {
        selectedPaths.value.clear()
        // Trigger reactivity manually for shallowRef
        triggerRef(selectedPaths)

        // Auto-save empty selection with debounce
        if (autoSaveSelection.value) {
            debouncedSaveSelection()
        }
    }

    function toggleExpand(path: string) {
        const node = findNode(nodes.value, path)
        if (node && node.isDir) {
            node.isExpanded = !node.isExpanded
            // Save expanded state with debounce
            debouncedSaveExpandedState()
        }
    }

    function expandPath(path: string) {
        const node = findNode(nodes.value, path)
        if (node && node.isDir) {
            node.isExpanded = true
        }
    }

    function collapsePath(path: string) {
        const node = findNode(nodes.value, path)
        if (node && node.isDir) {
            node.isExpanded = false
        }
    }

    function expandRecursive(path: string) {
        const node = findNode(nodes.value, path)
        if (!node || !node.isDir) return

        const expandNode = (n: FileNode) => {
            if (n.isDir) {
                n.isExpanded = true
                if (n.children) {
                    n.children.forEach(expandNode)
                }
            }
        }
        expandNode(node)
        debouncedSaveExpandedState()
    }

    function collapseRecursive(path: string) {
        const node = findNode(nodes.value, path)
        if (!node || !node.isDir) return

        const collapseNode = (n: FileNode) => {
            if (n.isDir) {
                n.isExpanded = false
                if (n.children) {
                    n.children.forEach(collapseNode)
                }
            }
        }
        collapseNode(node)
        debouncedSaveExpandedState()
    }

    function expandAll() {
        walkTree(nodes.value, node => {
            if (node.isDir) {
                node.isExpanded = true
            }
        })
        debouncedSaveExpandedState()
    }

    function collapseAll() {
        walkTree(nodes.value, node => {
            if (node.isDir) {
                node.isExpanded = false
            }
        })
        debouncedSaveExpandedState()
    }

    // Get all currently expanded paths
    function getExpandedPaths(): string[] {
        const expanded: string[] = []
        walkTree(nodes.value, node => {
            if (node.isDir && node.isExpanded) {
                expanded.push(node.path)
            }
        })
        return expanded
    }

    // Restore expanded state from array of paths
    function restoreExpandedPaths(paths: string[]) {
        const pathSet = new Set(paths)
        walkTree(nodes.value, node => {
            if (node.isDir) {
                node.isExpanded = pathSet.has(node.path)
            }
        })
    }

    // Check if a node exists in the tree
    function nodeExists(path: string): boolean {
        return findNode(nodes.value, path) !== null
    }

    function selectRecursive(path: string) {
        const node = findNode(nodes.value, path)
        if (!node) return

        // Use cached getAllFilesInNode for efficiency
        const filePaths = getAllFilesInNode(node)
        filePaths.forEach(p => selectedPaths.value.add(p))
        triggerRef(selectedPaths)
    }

    function deselectRecursive(path: string) {
        const node = findNode(nodes.value, path)
        if (!node) return

        // Use cached getAllFilesInNode for efficiency
        const filePaths = getAllFilesInNode(node)
        filePaths.forEach(p => selectedPaths.value.delete(p))
        triggerRef(selectedPaths)
    }

    function toggleSelectRecursive(path: string) {
        const node = findNode(nodes.value, path)
        if (!node || !node.isDir) return

        // Check if any child files are currently selected
        const childFilePaths = getAllFilesInNode(node)
        const anySelected = childFilePaths.some(filePath => selectedPaths.value.has(filePath))

        if (anySelected) {
            // Deselect all child files recursively
            childFilePaths.forEach(filePath => selectedPaths.value.delete(filePath))
        } else {
            // Select all child files recursively
            childFilePaths.forEach(filePath => selectedPaths.value.add(filePath))
        }
        // Trigger reactivity manually for shallowRef
        triggerRef(selectedPaths)
    }

    function getRecursiveFileCount(node: FileNode): number {
        if (!node.isDir) return 0

        let count = 0
        if (node.children) {
            for (const child of node.children) {
                if (!child.isDir) {
                    count++
                } else {
                    count += getRecursiveFileCount(child)
                }
            }
        }
        return count
    }

    function getAllFilesInNode(node: FileNode): string[] {
        // Check cache first (direct access, no .value)
        const cached = allFilesCache.get(node.path)
        if (cached) {
            return cached
        }

        const files: string[] = []

        if (node.children) {
            for (const child of node.children) {
                if (!child.isDir) {
                    files.push(child.path)
                } else {
                    files.push(...getAllFilesInNode(child))
                }
            }
        }

        // Cache the result (direct access, no .value)
        allFilesCache.set(node.path, files)
        return files
    }

    function getSelectedFileCountInNode(node: FileNode): number {
        const allFiles = getAllFilesInNode(node)
        return allFiles.filter(filePath => selectedPaths.value.has(filePath)).length
    }

    function isDirectory(path: string): boolean {
        const node = findNode(nodes.value, path)
        return node ? node.isDir : false
    }

    function getNodesByPaths(paths: string[]): Map<string, FileNode> {
        const result = new Map<string, FileNode>()
        const pathSet = new Set(paths)

        function walk(nodes: FileNode[]) {
            for (const node of nodes) {
                if (pathSet.has(node.path)) {
                    result.set(node.path, node)
                    if (result.size === paths.length) return // Early exit
                }
                if (node.children) {
                    walk(node.children)
                }
            }
        }

        walk(nodes.value)
        return result
    }

    function selectByExtension(extension: string) {
        walkTree(nodes.value, node => {
            if (!node.isDir && node.name.endsWith(extension)) {
                selectedPaths.value.add(node.path)
            }
        })
    }

    function setSearchQuery(query: string) {
        searchQuery.value = query
    }

    function setFilterExtensions(extensions: string[]) {
        filterExtensions.value = extensions
    }

    async function refreshFileTree(): Promise<void> {
        filesApi.clearCache()
    }

    function setRootPath(path: string) {
        rootPath.value = path
        currentDirectory.value = path
        directoryHistory.value = []
    }

    function getAvailableExtensions(): string[] {
        const extensions = new Set<string>()

        function collectExtensions(nodes: FileNode[]) {
            for (const node of nodes) {
                if (!node.isDir && node.name.includes('.')) {
                    const ext = '.' + node.name.split('.').pop()
                    extensions.add(ext)
                }
                if (node.children) {
                    collectExtensions(node.children)
                }
            }
        }

        collectExtensions(nodes.value)
        return Array.from(extensions).sort()
    }

    function resetStore() {
        nodes.value = []
        selectedPaths.value.clear()
        rootPath.value = ''
        currentDirectory.value = ''
        directoryHistory.value = []
        searchQuery.value = ''
        filterExtensions.value = []
        error.value = null
        isLoading.value = false
        flattenedNodesCache.value = null

        // Force garbage collection
        if (typeof window !== 'undefined' && (window as any).gc) {
            try {
                (window as any).gc()
            } catch (e) {
                // Ignore
            }
        }
    }

    function getMemoryUsage(): number {
        // OPTIMIZED: Use mathematical estimation instead of JSON.stringify
        let size = 0

        // Count all nodes in tree (not just top level)
        let nodeCount = 0
        const countNodes = (nodeList: FileNode[]) => {
            nodeCount += nodeList.length
            nodeList.forEach(node => {
                if (node.children) countNodes(node.children)
            })
        }
        countNodes(nodes.value)

        // Estimate: ~250 bytes per node (path, name, metadata)
        size += nodeCount * 250

        // Selected paths: ~100 bytes per path
        size += selectedPaths.value.size * 100

        // Flattened cache: ~200 bytes per cached node
        size += (flattenedNodesCache.value?.length || 0) * 200

        // allFilesCache: estimate based on cache size (direct access, no .value)
        size += allFilesCache.size * 150
        return size
    }

    function pruneUnusedBranches() {
        // Remove collapsed branches that haven't been accessed recently
        // This is a placeholder for future optimization
        console.log('[FileStore] Pruning unused branches...')
    }

    // Helper functions
    function findNode(tree: FileNode[], path: string): FileNode | null {
        for (const node of tree) {
            if (node.path === path) return node
            if (node.children) {
                const found = findNode(node.children, path)
                if (found) return found
            }
        }
        return null
    }

    function walkTree(tree: FileNode[], fn: (node: FileNode) => void) {
        for (const node of tree) {
            fn(node)
            if (node.children) {
                walkTree(node.children, fn)
            }
        }
    }

    function generateFlattened(tree: FileNode[], maxDepth: number, currentDepth: number = 0): FileNode[] {
        if (currentDepth >= maxDepth) return []

        const result: FileNode[] = []
        for (const node of tree) {
            result.push(node)
            if (node.children && currentDepth < maxDepth - 1) {
                result.push(...generateFlattened(node.children, maxDepth, currentDepth + 1))
            }
        }
        return result
    }

    function convertDomainNodes(domainNodes: any[]): FileNode[] {
        return domainNodes.map(node => ({
            name: node.name,
            path: node.path,
            isDir: node.isDir,
            isExpanded: false,
            isSelected: false,
            children: node.children ? convertDomainNodes(node.children) : undefined,
            size: node.size,
            isIgnored: node.isIgnored || node.isGitignored || node.isCustomIgnored
        }))
    }

    function filterTreeByExtensions(tree: FileNode[], extensions: string[]): FileNode[] {
        return tree.filter(node => {
            if (node.isDir) {
                const filteredChildren = node.children ? filterTreeByExtensions(node.children, extensions) : []
                return filteredChildren.length > 0
            }
            return extensions.some(ext => node.name.endsWith(ext))
        }).map(node => ({
            ...node,
            children: node.children ? filterTreeByExtensions(node.children, extensions) : undefined
        }))
    }

    // Selection persistence
    // Import settings store for autoSaveSelection
    const settingsStore = useSettingsStore()
    const autoSaveSelection = computed(() => settingsStore.settings.fileExplorer.autoSaveSelection)

    function saveSelectionToStorage() {
        if (!rootPath.value || !autoSaveSelection.value) return

        try {
            const key = `file-selection-${rootPath.value}`
            const selection = Array.from(selectedPaths.value).slice(0, 100) // Limit to 100 files
            localStorage.setItem(key, JSON.stringify(selection))
            console.log(`[FileStore] Saved selection: ${selection.length} files for ${rootPath.value}`)
        } catch (err) {
            console.warn('[FileStore] Failed to save selection:', err)
        }
    }

    function debouncedSaveSelection() {
        // Clear existing timer
        if (saveSelectionTimer) {
            clearTimeout(saveSelectionTimer)
        }

        // Set new timer - save after 300ms of inactivity
        saveSelectionTimer = setTimeout(() => {
            saveSelectionToStorage()
            saveSelectionTimer = null
        }, 300)
    }

    function loadSelectionFromStorage(projectPath: string) {
        try {
            const key = `file-selection-${projectPath}`
            const saved = localStorage.getItem(key)
            if (saved) {
                const selection = JSON.parse(saved) as string[]
                selectedPaths.value = new Set(selection)
                console.log(`[FileStore] Loaded selection: ${selection.length} files for ${projectPath}`)
            }
        } catch (err) {
            console.warn('[FileStore] Failed to load selection:', err)
        }
    }

    function clearSelectionHistory(projectPath?: string) {
        try {
            if (projectPath) {
                const key = `file-selection-${projectPath}`
                localStorage.removeItem(key)
            } else {
                // Clear all selection history
                const keys = Object.keys(localStorage).filter(k => k.startsWith('file-selection-'))
                keys.forEach(k => localStorage.removeItem(k))
            }
        } catch (err) {
            console.warn('[FileStore] Failed to clear selection history:', err)
        }
    }

    function getSelectionStats() {
        const stats: Record<string, number> = {}
        try {
            const keys = Object.keys(localStorage).filter(k => k.startsWith('file-selection-'))
            keys.forEach(key => {
                const projectPath = key.replace('file-selection-', '')
                const saved = localStorage.getItem(key)
                if (saved) {
                    const selection = JSON.parse(saved) as string[]
                    stats[projectPath] = selection.length
                }
            })
        } catch (err) {
            console.warn('[FileStore] Failed to get selection stats:', err)
        }
        return stats
    }

    // Expanded state persistence
    function saveExpandedState() {
        if (!rootPath.value) return

        try {
            const expandedPaths: string[] = []
            walkTree(nodes.value, node => {
                if (node.isDir && node.isExpanded) {
                    expandedPaths.push(node.path)
                }
            })

            const key = `file-expanded-${rootPath.value}`
            localStorage.setItem(key, JSON.stringify(expandedPaths))
        } catch (err) {
            console.warn('[FileStore] Failed to save expanded state:', err)
        }
    }

    function debouncedSaveExpandedState() {
        // Clear existing timer
        if (saveExpandedStateTimer) {
            clearTimeout(saveExpandedStateTimer)
        }

        // Set new timer - save after 500ms of inactivity
        saveExpandedStateTimer = setTimeout(() => {
            saveExpandedState()
            saveExpandedStateTimer = null
        }, 500)
    }

    function loadExpandedState() {
        if (!rootPath.value) return

        try {
            const key = `file-expanded-${rootPath.value}`
            const saved = localStorage.getItem(key)
            if (saved) {
                const expandedPaths = JSON.parse(saved) as string[]
                expandedPaths.forEach(path => {
                    const node = findNode(nodes.value, path)
                    if (node && node.isDir) {
                        node.isExpanded = true
                    }
                })
            }
        } catch (err) {
            console.warn('[FileStore] Failed to load expanded state:', err)
        }
    }

    return {
        // State
        nodes,
        selectedPaths,
        isLoading,
        error,
        searchQuery,
        filterExtensions,
        currentDirectory,
        directoryHistory,
        rootPath,
        // Computed
        hasSelectedFiles,
        selectedCount,
        selectedFilesList,
        estimatedTokenCount,
        estimatedContextSize,
        searchResults,
        filteredNodes,
        breadcrumbs,
        projectName,
        // Actions
        setFileTree,
        loadFileTree,
        toggleSelect,
        selectPath,
        deselectPath,
        selectMultiple,
        clearSelection,
        toggleExpand,
        expandPath,
        collapsePath,
        expandRecursive,
        collapseRecursive,
        expandAll,
        collapseAll,
        selectRecursive,
        deselectRecursive,
        selectByExtension,
        setSearchQuery,
        setFilterExtensions,
        refreshFileTree,
        setRootPath,
        getAvailableExtensions,
        getSelectedFilesSize,
        resetStore,
        getMemoryUsage,
        pruneUnusedBranches,
        // Public utility methods for UI components
        getRecursiveFileCount,
        getAllFilesInNode,
        getSelectedFileCountInNode,
        isDirectory,
        getNodesByPaths,
        // Selection persistence
        autoSaveSelection,
        saveSelectionToStorage,
        loadSelectionFromStorage,
        clearSelectionHistory,
        getSelectionStats,
        // Expanded state persistence
        saveExpandedState,
        loadExpandedState,
        getExpandedPaths,
        restoreExpandedPaths,
        nodeExists
    }
})
