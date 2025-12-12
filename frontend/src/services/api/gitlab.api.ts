/**
 * GitLab API (no clone required)
 * Handles GitLab repository operations via API
 */

import * as wails from '#wailsjs/go/main/App'
import type { GitLabBranch, GitLabCommit } from '../types'
import { apiCall, apiCallWithDefault, parseJsonResponse } from './base'

export const gitlabApi = {
    isGitLabURL: (url: string): Promise<boolean> =>
        apiCallWithDefault(() => wails.IsGitLabURL(url), false, 'gitlab'),

    getDefaultBranch: (repoURL: string): Promise<string> =>
        apiCall(() => wails.GitLabGetDefaultBranch(repoURL), 'Failed to get default branch.', 'gitlab'),

    getBranches: async (repoURL: string): Promise<GitLabBranch[]> => {
        const result = await apiCall(
            () => wails.GitLabGetBranches(repoURL),
            'Failed to get GitLab branches.',
            'gitlab'
        )
        return parseJsonResponse(result, 'Failed to parse GitLab branches.')
    },

    getCommits: async (repoURL: string, branch: string, limit = 50): Promise<GitLabCommit[]> => {
        const result = await apiCall(
            () => wails.GitLabGetCommits(repoURL, branch, limit),
            'Failed to get GitLab commits.',
            'gitlab'
        )
        return parseJsonResponse(result, 'Failed to parse GitLab commits.')
    },

    listFiles: async (repoURL: string, ref: string): Promise<string[]> => {
        const result = await apiCall(
            () => wails.GitLabListFiles(repoURL, ref),
            'Failed to list GitLab files.',
            'gitlab'
        )
        const parsed = parseJsonResponse<unknown>(result, 'Failed to parse GitLab files.')
        return Array.isArray(parsed) ? parsed : []
    },

    getFileContent: (repoURL: string, filePath: string, ref: string): Promise<string> =>
        apiCall(
            () => wails.GitLabGetFileContent(repoURL, filePath, ref),
            'Failed to get GitLab file content.',
            'gitlab'
        ),

    buildContext: (repoURL: string, files: string[], ref: string): Promise<string> =>
        apiCall(
            () => wails.GitLabBuildContext(repoURL, files, ref),
            'Failed to build GitLab context.',
            'gitlab'
        ),
}
