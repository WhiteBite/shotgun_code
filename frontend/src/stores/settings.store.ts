import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type OutputFormat = 'markdown' | 'xml' | 'json' | 'plain'

export interface ContextSettings {
    maxTokens: number
    stripComments: boolean
    includeTests: boolean
    splitStrategy: 'semantic' | 'fixed' | 'adaptive'
    outputFormat: OutputFormat
}

export interface QuickFilterConfig {
    id: string
    label: string
    extensions: string[]
    patterns: string[]  // glob patterns like '**/Test/**'
    enabled: boolean
}

export interface FileExplorerSettings {
    useGitignore: boolean
    useCustomIgnore: boolean
    autoSaveSelection: boolean
    compactNestedFolders: boolean
    showIgnoredFiles: boolean
    customIgnoreRules: string
    quickFilters: QuickFilterConfig[]
}

export interface ContextStorageSettings {
    maxContexts: number          // Максимальное количество сохранённых контекстов
    maxStorageMB: number         // Максимальный размер хранилища в MB
    autoCleanupDays: number      // Автоудаление контекстов старше N дней (0 = отключено)
    autoCleanupOnLimit: boolean  // Автоматически удалять старые при превышении лимита
}

export interface AppSettings {
    context: ContextSettings
    contextStorage: ContextStorageSettings
    fileExplorer: FileExplorerSettings
    aiModel: string
    theme: 'dark' | 'light' | 'auto'
}

const DEFAULT_SETTINGS: AppSettings = {
    context: {
        maxTokens: 100000, // 100K tokens by default (reasonable for most projects)
        stripComments: false,
        includeTests: true,
        splitStrategy: 'semantic',
        outputFormat: 'markdown' // Default format
    },
    contextStorage: {
        maxContexts: 20,
        maxStorageMB: 100,
        autoCleanupDays: 30,
        autoCleanupOnLimit: true
    },
    fileExplorer: {
        useGitignore: true,
        useCustomIgnore: false,
        autoSaveSelection: true,
        compactNestedFolders: true,
        showIgnoredFiles: true,
        customIgnoreRules: '',
        quickFilters: [
            { id: 'source', label: 'Исходный код', extensions: ['.ts', '.js', '.tsx', '.jsx', '.vue', '.go', '.py', '.java', '.cpp', '.c', '.rs'], patterns: [], enabled: true },
            { id: 'tests', label: 'Тесты', extensions: [], patterns: ['**/*.test.*', '**/*.spec.*', '**/test/**', '**/tests/**', '**/Test/**'], enabled: true },
            { id: 'config', label: 'Конфигурация', extensions: ['.json', '.yaml', '.yml', '.toml', '.ini', '.env'], patterns: [], enabled: true },
            { id: 'docs', label: 'Документация', extensions: ['.md', '.txt', '.rst', '.adoc'], patterns: [], enabled: true },
            { id: 'styles', label: 'Стили', extensions: ['.css', '.scss', '.sass', '.less'], patterns: [], enabled: true }
        ]
    },
    aiModel: 'gpt-4',
    theme: 'dark'
}

export const useSettingsStore = defineStore('settings', () => {
    const settings = ref<AppSettings>(loadSettings())

    // Watch for changes and save to localStorage
    watch(
        settings,
        (newSettings) => {
            saveSettings(newSettings)
        },
        { deep: true }
    )

    function loadSettings(): AppSettings {
        try {
            const saved = localStorage.getItem('app-settings')
            if (saved) {
                const parsed = JSON.parse(saved)
                // Merge with defaults to handle new settings
                return {
                    ...DEFAULT_SETTINGS,
                    ...parsed,
                    context: {
                        ...DEFAULT_SETTINGS.context,
                        ...parsed.context
                    },
                    contextStorage: {
                        ...DEFAULT_SETTINGS.contextStorage,
                        ...parsed.contextStorage
                    },
                    fileExplorer: {
                        ...DEFAULT_SETTINGS.fileExplorer,
                        ...parsed.fileExplorer
                    }
                }
            }
        } catch (err) {
            console.warn('Failed to load settings:', err)
        }
        return DEFAULT_SETTINGS
    }

    function saveSettings(settings: AppSettings) {
        try {
            localStorage.setItem('app-settings', JSON.stringify(settings))
        } catch (err) {
            console.warn('Failed to save settings:', err)
        }
    }

    function resetToDefaults() {
        settings.value = { ...DEFAULT_SETTINGS }
    }

    function updateContextSettings(updates: Partial<ContextSettings>) {
        settings.value.context = {
            ...settings.value.context,
            ...updates
        }
    }

    function updateAIModel(model: string) {
        settings.value.aiModel = model
    }

    function updateTheme(theme: 'dark' | 'light' | 'auto') {
        settings.value.theme = theme
    }

    function updateFileExplorerSettings(updates: Partial<FileExplorerSettings>) {
        settings.value.fileExplorer = {
            ...settings.value.fileExplorer,
            ...updates
        }
    }

    function getCustomIgnoreRules(): string {
        return settings.value.fileExplorer.customIgnoreRules
    }

    function setCustomIgnoreRules(rules: string) {
        settings.value.fileExplorer.customIgnoreRules = rules
    }

    function updateContextStorageSettings(updates: Partial<ContextStorageSettings>) {
        settings.value.contextStorage = {
            ...settings.value.contextStorage,
            ...updates
        }
    }

    return {
        settings,
        resetToDefaults,
        updateContextSettings,
        updateContextStorageSettings,
        updateAIModel,
        updateTheme,
        updateFileExplorerSettings,
        getCustomIgnoreRules,
        setCustomIgnoreRules
    }
})
