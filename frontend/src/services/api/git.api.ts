/**
 * Local Git operations API
 * Handles git status, branches, commits for local repositories
 */

import * as wails from '#wailsjs/go/main/App'
import type { domain } from '#wailsjs/go/models'
import type { CommitInfo } from '../types'
import { apiCall, apiCallWithDefault, parseJsonResponse } from './base'

export const gitApi = {
    getUncommittedFiles: (repoPath: string): Promise<domain.FileStatus[]> =>
        apiCall(() => wails.GetUncommittedFiles(repoPath), 'Failed to get git status.', 'git'),

    getBranches: (repoPath: string): Promise<string> =>
        apiCall(() => wails.GetBranches(repoPath), 'Failed to get git branches.', 'git'),

    getCurrentBranch: (repoPath: string): Promise<string> =>
        apiCall(() => wails.GetCurrentBranch(repoPath), 'Failed to get current branch.', 'git'),

    getRichCommitHistory: (
        repoPath: string,
        filePath: string,
        limit: number
    ): Promise<domain.CommitWithFiles[]> =>
        apiCall(
            () => wails.GetRichCommitHistory(repoPath, filePath, limit),
            'Failed to get commit history.',
            'git'
        ),

    isGitAvailable: (): Promise<boolean> =>
        apiCallWithDefault(() => wails.IsGitAvailable(), false, 'git'),

    isGitRepository: (projectPath: string): Promise<boolean> =>
        apiCallWithDefault(() => wails.IsGitRepository(projectPath), false, 'git'),

    cloneRepository: (url: string): Promise<string> =>
        apiCall(() => wails.CloneRepository(url), 'Failed to clone repository.', 'git'),

    checkoutBranch: (projectPath: string, branch: string): Promise<void> =>
        apiCall(() => wails.CheckoutBranch(projectPath, branch), 'Failed to checkout branch.', 'git'),

    checkoutCommit: (projectPath: string, commitHash: string): Promise<void> =>
        apiCall(
            () => wails.CheckoutCommit(projectPath, commitHash),
            'Failed to checkout commit.',
            'git'
        ),

    getCommitHistory: async (projectPath: string, limit = 50): Promise<CommitInfo[]> => {
        const result = await apiCall(
            () => wails.GetCommitHistory(projectPath, limit),
            'Failed to get commit history.',
            'git'
        )
        return parseJsonResponse(result, 'Failed to parse commit history.')
    },

    getRemoteBranches: async (projectPath: string): Promise<string[]> => {
        const result = await apiCall(
            () => wails.GetRemoteBranches(projectPath),
            'Failed to get remote branches.',
            'git'
        )
        return parseJsonResponse(result, 'Failed to parse remote branches.')
    },

    cleanupTempRepository: async (path: string): Promise<void> => {
        try {
            await wails.CleanupTempRepository(path)
        } catch (error) {
            console.error('[API:git] Error cleaning up temp repository:', error)
        }
    },

    listFilesAtRef: async (projectPath: string, ref: string): Promise<string[]> => {
        const result = await apiCall(
            () => wails.ListFilesAtRef(projectPath, ref),
            'Failed to list files at ref.',
            'git'
        )
        const parsed = parseJsonResponse<unknown>(result, 'Failed to parse files at ref.')
        return Array.isArray(parsed) ? parsed : []
    },

    getFileAtRef: (projectPath: string, filePath: string, ref: string): Promise<string> =>
        apiCall(
            () => wails.GetFileAtRef(projectPath, filePath, ref),
            'Failed to get file at ref.',
            'git'
        ),

    buildContextAtRef: (projectPath: string, files: string[], ref: string): Promise<string> =>
        apiCall(
            () => wails.BuildContextAtRef(projectPath, files, ref, '{}'),
            'Failed to build context at ref.',
            'git'
        ),
}
