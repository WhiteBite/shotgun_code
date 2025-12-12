/**
 * Application constants - centralized magic numbers and configuration values
 */

// File Tree Configuration
export const FILE_TREE = {
    MAX_SEARCH_RESULTS: 100,
    MAX_FLATTEN_DEPTH: 5,
    DEBOUNCE_MS: 300,
    MAX_SAVED_SELECTIONS: 100,
    DEFAULT_EXPANDED_DEPTH: 2,
    VIRTUAL_ITEM_HEIGHT: 28,
    OVERSCAN_COUNT: 5,
} as const

// Cache Configuration
export const CACHE = {
    TTL_MS: 5 * 60 * 1000, // 5 minutes
    MAX_ENTRIES: 1000,
    STALE_TIME_MS: 30 * 1000, // 30 seconds
} as const

// UI Configuration
export const UI = {
    TOAST_DURATION_MS: 3000,
    TOAST_ERROR_DURATION_MS: 5000,
    ANIMATION_DURATION_MS: 200,
    COPY_FEEDBACK_DURATION_MS: 2000,
    MODAL_BACKDROP_OPACITY: 0.5,
} as const

// API Configuration
export const API = {
    RETRY_COUNT: 3,
    RETRY_DELAY_MS: 1000,
    TIMEOUT_MS: 30000,
    STREAM_CHUNK_SIZE: 1024,
} as const

// Context Configuration
export const CONTEXT = {
    MAX_FILES: 500,
    MAX_TOKENS: 100000,
    MAX_FILE_SIZE_BYTES: 1024 * 1024, // 1MB
    CHUNK_SIZE: 50000,
} as const

// Git Configuration
export const GIT = {
    MAX_COMMITS: 100,
    MAX_DIFF_LINES: 1000,
    MAX_CHANGED_FILES: 50,
} as const

// Undo/Redo Configuration
export const UNDO_REDO = {
    MAX_HISTORY_SIZE: 50,
} as const

// Keyboard Shortcuts
export const SHORTCUTS = {
    UNDO: 'Ctrl+Z',
    REDO: 'Ctrl+Shift+Z',
    SEARCH: 'Ctrl+P',
    COMMAND_PALETTE: 'Ctrl+Shift+P',
    COPY: 'Ctrl+C',
    SELECT_ALL: 'Ctrl+A',
} as const
