// Git Feature - Public API
export { gitApi } from './api/git.api'
export { getGitCache, useGitCache } from './composables/useGitCache'
export { useFileSelection, useGitFilters } from './composables/useGitFilters'
export { useGitStore } from './model/git.store'

// Types
export type {
    FileFilter, GitBranch,
    GitCommit, GitDiffFile, GitFileNode, GitProvider, RecentRepo
} from './model/types'

export { FILE_TYPE_FILTERS } from './model/types'

