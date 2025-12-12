/**
 * Project management API
 * Handles projects, directories, paths
 */

import * as wails from '#wailsjs/go/main/App'
import { apiCall, apiCallWithDefault } from './base'

export const projectApi = {
    getRecentProjects: () =>
        apiCall(() => wails.GetRecentProjects(), 'Failed to load recent projects.', 'project'),

    addRecentProject: (path: string, name: string) =>
        apiCall(() => wails.AddRecentProject(path, name), 'Failed to add recent project.', 'project'),

    removeRecentProject: (path: string) =>
        apiCall(() => wails.RemoveRecentProject(path), 'Failed to remove recent project.', 'project'),

    selectDirectory: () =>
        apiCall(() => wails.SelectDirectory(), 'Failed to select directory.', 'project'),

    getCurrentDirectory: () =>
        apiCall(() => wails.GetCurrentDirectory(), 'Failed to get current directory.', 'project'),

    pathExists: (path: string) =>
        apiCallWithDefault(() => wails.PathExists(path), false, 'project'),
}
