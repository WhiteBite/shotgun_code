import type { domain } from '#wailsjs/go/models'
import { useLogger } from '@/composables/useLogger'
import { useProjectStore } from '@/stores/project.store'
import { useSettingsStore } from '@/stores/settings.store'
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { contextApi } from '../api/context.api'

const logger = useLogger('ContextStore')

// Генерация умного названия контекста на основе файлов
function generateSmartName(files: string[]): string {
    if (!files || files.length === 0) return 'Пустой контекст'

    // Извлекаем расширения и папки
    const extensions = new Map<string, number>()
    const folders = new Map<string, number>()
    const fileNames: string[] = []

    for (const file of files) {
        const parts = file.split('/')
        const fileName = parts[parts.length - 1]
        fileNames.push(fileName)

        // Считаем расширения
        const ext = fileName.includes('.') ? fileName.split('.').pop()?.toLowerCase() : ''
        if (ext) {
            extensions.set(ext, (extensions.get(ext) || 0) + 1)
        }

        // Считаем папки верхнего уровня
        if (parts.length > 1) {
            const topFolder = parts[0]
            folders.set(topFolder, (folders.get(topFolder) || 0) + 1)
        }
    }

    // Определяем доминирующий тип
    const sortedExts = [...extensions.entries()].sort((a, b) => b[1] - a[1])
    const sortedFolders = [...folders.entries()].sort((a, b) => b[1] - a[1])

    const typeLabels: Record<string, string> = {
        'vue': 'Vue компоненты',
        'ts': 'TypeScript',
        'tsx': 'React TSX',
        'js': 'JavaScript',
        'jsx': 'React JSX',
        'go': 'Go код',
        'py': 'Python',
        'java': 'Java',
        'css': 'Стили',
        'scss': 'SCSS стили',
        'json': 'Конфигурация',
        'yaml': 'YAML конфиг',
        'yml': 'YAML конфиг',
        'md': 'Документация',
        'sql': 'SQL запросы',
        'html': 'HTML шаблоны',
        'test': 'Тесты',
        'spec': 'Тесты'
    }

    // Проверяем на тесты
    const hasTests = fileNames.some(f => f.includes('test') || f.includes('spec') || f.includes('Test'))

    let name = ''

    // Если один файл - используем его имя
    if (files.length === 1) {
        const fileName = fileNames[0]
        return fileName.length > 30 ? fileName.substring(0, 27) + '...' : fileName
    }

    // Если все файлы из одной папки
    if (sortedFolders.length === 1 && sortedFolders[0][1] === files.length) {
        name = sortedFolders[0][0]
    }
    // Если доминирует один тип файлов (>60%)
    else if (sortedExts.length > 0 && sortedExts[0][1] / files.length > 0.6) {
        const ext = sortedExts[0][0]
        name = typeLabels[ext] || `.${ext} файлы`
    }
    // Если есть явная папка-лидер
    else if (sortedFolders.length > 0 && sortedFolders[0][1] / files.length > 0.5) {
        name = sortedFolders[0][0]
    }
    // Смешанный контекст
    else {
        const topTypes = sortedExts.slice(0, 2).map(([ext]) => typeLabels[ext] || ext)
        name = topTypes.join(' + ') || 'Смешанный'
    }

    // Добавляем метку тестов
    if (hasTests && !name.toLowerCase().includes('тест')) {
        name += ' (с тестами)'
    }

    // Добавляем количество файлов
    name += ` [${files.length}]`

    return name
}

export interface ContextSummary {
    id: string
    name: string              // Умное название контекста
    fileCount: number
    totalSize: number
    lineCount: number
    tokenCount?: number
    createdAt?: string
    projectPath?: string
    isFavorite?: boolean      // Закреплённый контекст
    files?: string[]          // Список файлов для генерации названия
    metadata?: {
        warnings?: string[]
        skippedFiles?: string[]
        skippedReasons?: Record<string, string>
        projectPath?: string
    }
}

export interface ContextChunk {
    lines: string[]
    startLine: number
    endLine: number
    hasMore: boolean
}

export const MAX_CONTEXT_SIZE = 20 * 1024 * 1024 // 20 MB (reduced from 50MB for memory safety)

export const useContextStore = defineStore('context', () => {
    // State
    const contextId = ref<string | null>(null)
    const summary = ref<ContextSummary | null>(null)
    const currentChunk = ref<ContextChunk | null>(null)
    const isLoading = ref(false)
    const isBuilding = ref(false)
    const buildProgress = ref(0)
    const error = ref<string | null>(null)
    const contextList = ref<ContextSummary[]>([])
    const warnings = ref<string[]>([])
    const skippedFiles = ref<string[]>([])

    // Cache for full context content to avoid reloading on pagination
    const fullContentCache = ref<string | null>(null)
    const cachedContextId = ref<string | null>(null)

    // Computed
    const hasContext = computed(() => contextId.value !== null)
    const totalSize = computed(() => summary.value?.totalSize || 0)
    const fileCount = computed(() => summary.value?.fileCount || 0)
    const lineCount = computed(() => summary.value?.lineCount || 0)
    const tokenCount = computed(() => summary.value?.tokenCount || Math.round(lineCount.value * 2.5))
    const estimatedCost = computed(() => (tokenCount.value / 1000) * 0.002)

    // Actions
    async function buildContext(filePaths: string[], options?: Partial<domain.ContextBuildOptions>) {
        if (filePaths.length === 0) {
            error.value = 'No files selected'
            return
        }

        isBuilding.value = true
        isLoading.value = true
        buildProgress.value = 0
        error.value = null

        try {
            const projectStore = useProjectStore()
            if (!projectStore.currentPath) {
                error.value = 'No project selected'
                throw new Error('No project selected')
            }
            const projectPath = projectStore.currentPath

            const buildOptions = {
                maxTokens: options?.maxTokens || 100000, // Default 100K tokens (reasonable for most projects)
                maxMemoryMB: options?.maxMemoryMB || 50,
                stripComments: options?.stripComments ?? false,
                includeManifest: options?.includeManifest ?? false,
                includeTests: options?.includeTests ?? true,
                splitStrategy: options?.splitStrategy || 'smart',
                forceStream: true,
                enableProgressEvents: true,
                outputFormat: options?.outputFormat || 'xml',
                // Content optimization options
                excludeTests: options?.excludeTests ?? false,
                collapseEmptyLines: options?.collapseEmptyLines ?? false,
                stripLicense: options?.stripLicense ?? false,
                compactDataFiles: options?.compactDataFiles ?? false,
                trimWhitespace: options?.trimWhitespace ?? false
            } as domain.ContextBuildOptions

            logger.debug('Building context with options:', {
                outputFormat: buildOptions.outputFormat,
                stripComments: buildOptions.stripComments,
                maxTokens: buildOptions.maxTokens
            })

            buildProgress.value = 30

            const result = await contextApi.buildContext(
                projectPath,
                filePaths,
                buildOptions
            )

            buildProgress.value = 70

            contextId.value = result.id || `ctx-${Date.now()}`
            const smartName = generateSmartName(filePaths)
            summary.value = {
                id: contextId.value,
                name: smartName,
                fileCount: result.fileCount || filePaths.length,
                totalSize: result.totalSize || 0,
                lineCount: result.lineCount || 0,
                tokenCount: result.tokenCount,
                createdAt: new Date().toISOString(),
                files: filePaths,
                isFavorite: false
            }

            // Check size limit
            if (summary.value.totalSize > MAX_CONTEXT_SIZE) {
                error.value = `Context size (${Math.round(summary.value.totalSize / (1024 * 1024))}MB) exceeds maximum allowed size (${Math.round(MAX_CONTEXT_SIZE / (1024 * 1024))}MB)`
                clearContext()
                throw new Error(error.value)
            }

            buildProgress.value = 100

            // Store warnings and skipped files from metadata
            warnings.value = result.metadata?.warnings || []
            skippedFiles.value = result.metadata?.skippedFiles || []

            // Check if context was built with empty content (path traversal issue)
            // Backend may return fileCount > 0 but with 0 size/lines if all files were filtered
            if (summary.value.fileCount === 0 || summary.value.totalSize === 0 || summary.value.lineCount === 0) {
                const projectStore = useProjectStore()
                const currentProject = projectStore.currentPath || 'unknown'
                const skippedInfo = skippedFiles.value.length > 0
                    ? `\n\nSkipped files (${skippedFiles.value.length}):\n${skippedFiles.value.slice(0, 5).join('\n')}${skippedFiles.value.length > 5 ? '\n...' : ''}`
                    : ''

                console.warn('[ContextStore] Context built with empty content. This usually indicates path traversal issues or files outside project directory.', {
                    contextId: contextId.value,
                    summary: summary.value,
                    requestedFiles: filePaths.length,
                    currentProject,
                    skippedFiles: skippedFiles.value
                })
                error.value = `Context built with empty content (${summary.value.fileCount} files, ${summary.value.totalSize} bytes, ${summary.value.lineCount} lines).\n\nCurrent project: ${currentProject}\n\nFiles may be outside the project directory. Please ensure you have opened the correct project.${skippedInfo}`
                // Keep the context data so UI can display the warning with context info
                // Don't call clearContext() - let the error state persist
                throw new Error(error.value)
            }

            // Auto-load preview after successful build (only if there are files)
            if (contextId.value && summary.value.fileCount > 0) {
                await loadContextContent(contextId.value, 0, 100)
            }
        } catch (err) {
            // Handle token limit exceeded error specially
            if (err instanceof Error && err.message === 'TOKEN_LIMIT_EXCEEDED') {
                const { hasTokenInfo } = await import('@/types/errors')
                const settingsStore = useSettingsStore()
                const currentLimit = settingsStore.settings.context.maxTokens

                if (hasTokenInfo(err)) {
                    error.value = `TOKEN_LIMIT_EXCEEDED:${err.tokenInfo.actual}:${err.tokenInfo.limit}`
                    console.error('[ContextStore] Token limit exceeded:', {
                        actual: err.tokenInfo.actual,
                        limit: err.tokenInfo.limit,
                        actualK: Math.round(err.tokenInfo.actual / 1000),
                        limitK: Math.round(err.tokenInfo.limit / 1000),
                        currentSetting: currentLimit
                    })
                } else {
                    error.value = `TOKEN_LIMIT_EXCEEDED:0:${currentLimit}`
                }
                throw err
            }

            // Enhance error message to include hint about path issues
            let errorMessage = err instanceof Error ? err.message : 'Failed to build context'

            // Add specific tips for different error types
            if (errorMessage.includes('path traversal') || errorMessage.includes('invalid path')) {
                errorMessage += ' Make sure all selected files are within the project directory.'
            }

            console.error('[ContextStore] Error building context:', err)
            error.value = errorMessage

            // Throw new error with enhanced message so callers receive the tips
            throw new Error(errorMessage)
        } finally {
            isBuilding.value = false
            isLoading.value = false
            buildProgress.value = 0
        }
    }

    async function loadContextContent(ctxId: string, startLine: number = 0, lineCount: number = 100) {
        if (!ctxId) {
            console.error('[ContextStore] No context ID provided')
            error.value = 'No context ID provided'
            return
        }

        isLoading.value = true
        error.value = null

        try {
            // Set contextId if loading new context
            if (contextId.value !== ctxId) {
                contextId.value = ctxId

                // Clear cache when switching contexts
                fullContentCache.value = null
                cachedContextId.value = null

                // Try to get summary from local list first
                const existingSummary = contextList.value.find(c => c.id === ctxId)
                if (existingSummary) {
                    summary.value = existingSummary
                } else {
                    // If not in list, try to fetch from backend
                    logger.debug('Summary not in local list, fetching from backend...')
                    try {
                        const contexts = await contextApi.getProjectContexts(useProjectStore().currentPath || '')
                        const fetchedSummary = contexts.find((c: { id: string }) => c.id === ctxId)
                        if (fetchedSummary) {
                            summary.value = {
                                id: fetchedSummary.id,
                                name: fetchedSummary.name || generateSmartName(fetchedSummary.files || []),
                                fileCount: fetchedSummary.fileCount || 0,
                                totalSize: fetchedSummary.totalSize || 0,
                                lineCount: fetchedSummary.lineCount || 0,
                                tokenCount: fetchedSummary.tokenCount,
                                createdAt: fetchedSummary.createdAt,
                                isFavorite: false
                            }
                            logger.debug('Summary fetched from backend:', summary.value)
                        } else {
                            console.warn('[ContextStore] Context summary not found on backend')
                        }
                    } catch (err) {
                        console.error('[ContextStore] Failed to fetch summary from backend:', err)
                    }
                }
            }

            // Use cached content if available for same context
            let content: string
            if (cachedContextId.value === ctxId && fullContentCache.value !== null) {
                logger.debug('Using cached content for pagination')
                content = fullContentCache.value
            } else {
                // Load full content from API (backend handles streaming internally)
                logger.debug('Loading full content from API...')
                content = await contextApi.getContextContent(ctxId)

                // Check size limit before accepting
                if (content.length > MAX_CONTEXT_SIZE) {
                    const errorMsg = `Context content (${Math.round(content.length / (1024 * 1024))}MB) exceeds maximum allowed size (${Math.round(MAX_CONTEXT_SIZE / (1024 * 1024))}MB)`
                    console.error('[ContextStore]', errorMsg)
                    error.value = errorMsg
                    throw new Error(errorMsg)
                }

                // Cache the content for pagination
                fullContentCache.value = content
                cachedContextId.value = ctxId
                logger.debug('Content cached, size:', Math.round(content.length / 1024), 'KB')
            }

            // Extract requested chunk from full content
            const lines = content.split('\n')
            const endLine = Math.min(startLine + lineCount, lines.length)

            currentChunk.value = {
                lines: lines.slice(startLine, endLine),
                startLine,
                endLine,
                hasMore: endLine < lines.length
            }

            logger.debug('Loaded chunk:', startLine, '-', endLine, 'of', lines.length, 'lines')
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to load context'
            console.error('[ContextStore] Error loading context:', err)
            error.value = errorMsg
            throw err
        } finally {
            isLoading.value = false
        }
    }

    async function getFullContextContent(): Promise<string> {
        if (!contextId.value) {
            console.error('[ContextStore] No context ID available for getFullContextContent')
            throw new Error('No context ID available')
        }

        try {
            // Use cache if available
            if (cachedContextId.value === contextId.value && fullContentCache.value !== null) {
                logger.debug('Returning cached full content')
                return fullContentCache.value
            }

            logger.debug('Fetching full content for copy/export...')
            const content = await contextApi.getContextContent(contextId.value)

            // Update cache
            fullContentCache.value = content
            cachedContextId.value = contextId.value

            return content
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to get full context content'
            console.error('[ContextStore] Error getting full context:', err)
            error.value = errorMsg
            throw err
        }
    }

    async function deleteContext(ctxId: string) {
        try {
            logger.debug('Deleting context:', ctxId)
            await contextApi.deleteContext(ctxId)

            // Clear current context if we're deleting the active one
            if (contextId.value === ctxId) {
                logger.debug('Deleted active context, clearing state')
                clearContext()
            }

            // Remove from list
            contextList.value = contextList.value.filter(c => c.id !== ctxId)
            logger.debug('Context deleted successfully')
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to delete context'
            console.error('[ContextStore] Error deleting context:', err)
            error.value = errorMsg
            throw err
        }
    }

    async function exportContext(ctxId: string) {
        if (!ctxId) {
            console.error('[ContextStore] No context ID provided for export')
            error.value = 'No context ID provided'
            throw new Error('No context ID provided')
        }

        try {
            logger.debug('Exporting context:', ctxId)

            // Create export settings object
            const exportSettings = {
                contextId: ctxId,
                format: 'markdown',
                includeMetadata: true
            }

            const result = await contextApi.exportContext(exportSettings)
            logger.debug('Context exported successfully:', result)
            return result
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to export context'
            console.error('[ContextStore] Error exporting context:', err)
            error.value = errorMsg
            throw err
        }
    }

    async function listProjectContexts() {
        isLoading.value = true
        error.value = null

        try {
            const projectStore = useProjectStore()
            if (!projectStore.currentPath) {
                console.error('[ContextStore] No project selected for listing contexts')
                error.value = 'No project selected'
                throw new Error('No project selected')
            }
            const projectPath = projectStore.currentPath

            logger.debug('Listing contexts for project:', projectPath)
            const contexts = await contextApi.getProjectContexts(projectPath)

            // Ensure contexts is an array (defensive check)
            const contextArray = Array.isArray(contexts) ? contexts : []
            contextList.value = contextArray.map((ctx: { id: string; name?: string; files?: string[]; totalSize?: number; totalTokens?: number; createdAt?: string; fileCount?: number; lineCount?: number; tokenCount?: number }) => ({
                id: ctx.id,
                name: ctx.name || generateSmartName(ctx.files || []),
                fileCount: ctx.fileCount || 0,
                totalSize: ctx.totalSize || 0,
                lineCount: ctx.lineCount || 0,
                tokenCount: ctx.tokenCount,
                createdAt: ctx.createdAt,
                files: ctx.files,
                isFavorite: false
            }))

            // Загружаем сохранённые метаданные (имена, избранное)
            loadContextMetadata()

            logger.debug('Found', contextList.value.length, 'contexts')

            // Запускаем автоочистку
            await autoCleanup()
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to list contexts'
            console.error('[ContextStore] Error listing contexts:', err)
            error.value = errorMsg
            throw err
        } finally {
            isLoading.value = false
        }
    }

    function clearContext() {
        contextId.value = null
        summary.value = null
        currentChunk.value = null
        error.value = null

        // Clear content cache
        fullContentCache.value = null
        cachedContextId.value = null

        // Force garbage collection hint (window.gc typed in performance.d.ts)
        if (typeof window !== 'undefined' && window.gc) {
            try {
                window.gc()
            } catch {
                // Ignore if gc is not available
            }
        }
    }

    // Set raw context content directly (for git ref builds)
    function setRawContext(content: string, fileCount: number) {
        const id = `git-ref-${Date.now()}`
        const lines = content.split('\n')

        contextId.value = id
        fullContentCache.value = content
        cachedContextId.value = id

        summary.value = {
            id,
            name: `Git ref [${fileCount}]`,
            fileCount,
            totalSize: content.length,
            lineCount: lines.length,
            tokenCount: Math.round(content.length / 4),
            createdAt: new Date().toISOString(),
            isFavorite: false
        }

        currentChunk.value = {
            lines: lines.slice(0, 100),
            startLine: 0,
            endLine: Math.min(100, lines.length),
            hasMore: lines.length > 100
        }
    }

    function getMemoryUsage(): number {
        // OPTIMIZED: Use mathematical estimation instead of JSON.stringify
        let size = 0

        // Summary: fixed size for metadata object (~1KB)
        if (summary.value) {
            size += 1024
        }

        // Current chunk: estimate based on line count
        if (currentChunk.value && currentChunk.value.lines) {
            // Average 80 bytes per line * 2 (UTF-16)
            size += currentChunk.value.lines.length * 80 * 2
        }

        // Full content cache: direct string length * 2 (UTF-16)
        if (fullContentCache.value) {
            size += fullContentCache.value.length * 2
        }

        return size
    }

    // Переименование контекста
    function renameContext(ctxId: string, newName: string) {
        const ctx = contextList.value.find(c => c.id === ctxId)
        if (ctx) {
            ctx.name = newName
            saveContextMetadata()
        }
        if (summary.value?.id === ctxId) {
            summary.value.name = newName
        }
    }

    // Переключение избранного
    function toggleFavorite(ctxId: string) {
        const ctx = contextList.value.find(c => c.id === ctxId)
        if (ctx) {
            ctx.isFavorite = !ctx.isFavorite
            saveContextMetadata()
        }
        if (summary.value?.id === ctxId) {
            summary.value.isFavorite = !summary.value.isFavorite
        }
    }

    // Сохранение метаданных контекстов в localStorage
    function saveContextMetadata() {
        try {
            const metadata = contextList.value.map(c => ({
                id: c.id,
                name: c.name,
                isFavorite: c.isFavorite
            }))
            localStorage.setItem('context-metadata', JSON.stringify(metadata))
        } catch (err) {
            console.warn('[ContextStore] Failed to save context metadata:', err)
        }
    }

    // Загрузка метаданных контекстов из localStorage
    function loadContextMetadata() {
        try {
            const saved = localStorage.getItem('context-metadata')
            if (saved) {
                const metadata = JSON.parse(saved) as Array<{ id: string; name: string; isFavorite?: boolean }>
                for (const ctx of contextList.value) {
                    const meta = metadata.find(m => m.id === ctx.id)
                    if (meta) {
                        ctx.name = meta.name
                        ctx.isFavorite = meta.isFavorite
                    }
                }
            }
        } catch (err) {
            console.warn('[ContextStore] Failed to load context metadata:', err)
        }
    }

    // Автоочистка старых контекстов
    async function autoCleanup() {
        const settingsStore = useSettingsStore()
        const storage = settingsStore.settings.contextStorage

        if (!storage.autoCleanupOnLimit) return

        // Фильтруем не-избранные контексты
        const nonFavorites = contextList.value
            .filter(c => !c.isFavorite)
            .sort((a, b) => {
                const dateA = new Date(a.createdAt || 0).getTime()
                const dateB = new Date(b.createdAt || 0).getTime()
                return dateA - dateB // Старые первые
            })

        // Удаляем по лимиту количества
        const totalCount = contextList.value.length
        if (totalCount > storage.maxContexts) {
            const toDelete = totalCount - storage.maxContexts
            const contextsToDelete = nonFavorites.slice(0, toDelete)

            for (const ctx of contextsToDelete) {
                try {
                    await deleteContext(ctx.id)
                    logger.debug('Auto-deleted old context:', ctx.id)
                } catch (err) {
                    console.error('[ContextStore] Failed to auto-delete context:', err)
                }
            }
        }

        // Удаляем по возрасту
        if (storage.autoCleanupDays > 0) {
            const cutoffDate = Date.now() - storage.autoCleanupDays * 24 * 60 * 60 * 1000

            for (const ctx of nonFavorites) {
                const ctxDate = new Date(ctx.createdAt || 0).getTime()
                if (ctxDate < cutoffDate) {
                    try {
                        await deleteContext(ctx.id)
                        logger.debug('Auto-deleted expired context:', ctx.id)
                    } catch (err) {
                        console.error('[ContextStore] Failed to auto-delete expired context:', err)
                    }
                }
            }
        }
    }

    // Дублирование контекста
    async function duplicateContext(ctxId: string): Promise<string | null> {
        try {
            const content = await contextApi.getContextContent(ctxId)
            const original = contextList.value.find(c => c.id === ctxId)

            if (!original) return null

            const newId = `ctx-copy-${Date.now()}`
            const newName = `${original.name} (копия)`

            // Создаём новый контекст с тем же содержимым
            setRawContext(content, original.fileCount)

            if (summary.value) {
                summary.value.id = newId
                summary.value.name = newName
            }

            // Добавляем в список
            contextList.value.push({
                id: newId,
                name: newName,
                fileCount: original.fileCount,
                totalSize: original.totalSize,
                lineCount: original.lineCount,
                tokenCount: original.tokenCount,
                createdAt: new Date().toISOString(),
                isFavorite: false
            })

            saveContextMetadata()
            return newId
        } catch (err) {
            console.error('[ContextStore] Failed to duplicate context:', err)
            return null
        }
    }

    // Объединение нескольких контекстов в один
    async function mergeContexts(contextIds: string[]): Promise<string | null> {
        if (contextIds.length < 2) return null

        try {
            const contents: string[] = []
            let totalFiles = 0
            const names: string[] = []

            for (const ctxId of contextIds) {
                // Load each context
                if (contextId.value !== ctxId) {
                    await loadContextContent(ctxId, 0, 100)
                }
                const content = await getFullContextContent()
                contents.push(content)

                const ctx = contextList.value.find(c => c.id === ctxId)
                if (ctx) {
                    totalFiles += ctx.fileCount
                    names.push(ctx.name || ctx.id)
                }
            }

            // Merge contents with separator
            const mergedContent = contents.join('\n\n' + '='.repeat(80) + '\n\n')
            const newId = `merged-${Date.now()}`
            const newName = `Merged: ${names.slice(0, 2).join(' + ')}${names.length > 2 ? ` +${names.length - 2}` : ''}`

            // Set as current context
            setRawContext(mergedContent, totalFiles)

            if (summary.value) {
                summary.value.id = newId
                summary.value.name = newName
            }

            // Add to list
            contextList.value.push({
                id: newId,
                name: newName,
                fileCount: totalFiles,
                totalSize: mergedContent.length,
                lineCount: mergedContent.split('\n').length,
                tokenCount: Math.round(mergedContent.length / 4),
                createdAt: new Date().toISOString(),
                isFavorite: false
            })

            saveContextMetadata()
            return newId
        } catch (err) {
            console.error('[ContextStore] Failed to merge contexts:', err)
            return null
        }
    }

    // Изменение порядка контекстов (для drag & drop)
    function reorderContexts(fromIndex: number, toIndex: number) {
        const list = [...contextList.value]
        const [removed] = list.splice(fromIndex, 1)
        list.splice(toIndex, 0, removed)
        contextList.value = list
        saveContextMetadata()
    }

    return {
        // State
        contextId,
        summary,
        currentChunk,
        isLoading,
        isBuilding,
        buildProgress,
        error,
        contextList,
        warnings,
        skippedFiles,
        // Computed
        hasContext,
        totalSize,
        fileCount,
        lineCount,
        tokenCount,
        estimatedCost,
        // Actions
        buildContext,
        loadContextContent,
        deleteContext,
        exportContext,
        listProjectContexts,
        clearContext,
        setRawContext,
        getFullContextContent,
        getMemoryUsage,
        // New actions
        renameContext,
        toggleFavorite,
        duplicateContext,
        autoCleanup,
        loadContextMetadata,
        saveContextMetadata,
        generateSmartName,
        mergeContexts,
        reorderContexts
    }
})

// Export for use in store
export { generateSmartName }

