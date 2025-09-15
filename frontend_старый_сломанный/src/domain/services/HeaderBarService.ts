import { computed } from 'vue'
import type { Project } from '@/types/dto'
import { APP_CONFIG } from '@/config/app-config'

export class HeaderBarService {
  constructor(
    private getCurrentProject: () => Project | null,
    private getSelectedFilesCount: () => number,
    private getBuildStatus: () => string,
    private getContextMetrics: () => { tokenCount: number; fileCount: number },
    private getIsManualMode: () => boolean,
    private getIsTransitioning: () => boolean,
    private getLayoutClasses: () => string
  ) {}

  // Project information
  get projectName() {
    const project = this.getCurrentProject()
    return project?.name || 'No Project'
  }

  get projectPath() {
    const project = this.getCurrentProject()
    return project?.path || ''
  }

  get projectStatus() {
    const project = this.getCurrentProject()
    if (!project) return 'none'
    
    const hasFiles = this.getSelectedFilesCount() > 0
    const hasContext = this.getBuildStatus() === 'complete'
    const hasErrors = this.getBuildStatus() === 'error'
    
    if (hasErrors) return 'error'
    if (hasContext && hasFiles) return 'healthy'
    if (hasFiles) return 'ready'
    return 'idle'
  }

  get truncatedPath() {
    const path = this.projectPath
    if (!path) return ''
    
    const maxLength = APP_CONFIG.ui.header.MAX_PATH_LENGTH
    if (path.length <= maxLength) return path
    
    const parts = path.split('/')
    if (parts.length > 2) {
      return `.../${parts.slice(-2).join('/')}`
    }
    
    return `...${path.slice(-maxLength)}`
  }

  // Tooltip content
  getProjectTooltipContent() {
    const project = this.getCurrentProject()
    if (!project) return 'No project loaded'
    
    return `
      <div class="space-y-2">
        <div class="font-medium">${project.name}</div>
        <div class="text-sm text-gray-400">Path: ${project.path}</div>
        <div class="text-sm text-gray-400">Files: ${this.getSelectedFilesCount()}</div>
        <div class="text-xs text-gray-500 mt-2">Click to open project settings</div>
      </div>
    `
  }

  getStatusTooltipContent() {
    const status = this.projectStatus
    const statusInfo: Record<string, { color: string; description: string; icon: string }> = {
      'healthy': {
        color: 'text-green-400',
        description: 'Project is ready with context built',
        icon: '✅'
      },
      'ready': {
        color: 'text-blue-400',
        description: 'Files selected, ready to build context',
        icon: '📁'
      },
      'idle': {
        color: 'text-gray-400',
        description: 'No files selected',
        icon: '⏸️'
      },
      'error': {
        color: 'text-red-400',
        description: 'Context build failed or project has errors',
        icon: '❌'
      },
      'none': {
        color: 'text-gray-500',
        description: 'No project loaded',
        icon: '🚫'
      }
    }
    
    const info = statusInfo[status] || statusInfo['none']
    return `
      <div class="space-y-1">
        <div class="flex items-center space-x-2">
          <span>${info.icon}</span>
          <span class="font-medium ${info.color}">${status.toUpperCase()}</span>
        </div>
        <div class="text-sm text-gray-300">${info.description}</div>
      </div>
    `
  }

  getModeToggleTooltip() {
    const isManualMode = this.getIsManualMode()
    return `
      <div class="space-y-2">
        <div class="font-medium">Workspace Mode</div>
        <div class="text-sm text-gray-300">
          Current: <span class="font-medium ${isManualMode ? 'text-blue-400' : 'text-purple-400'}>
            ${isManualMode ? 'Manual' : 'Autonomous'}
          </span>
        </div>
        <div class="text-xs text-gray-400">
          Click to toggle between modes<br>
          <kbd class="bg-gray-700 px-1 rounded">Ctrl+M</kbd> to switch
        </div>
      </div>
    `
  }

  getManualModeTooltip() {
    return `
      <div class="space-y-1">
        <div class="font-medium text-blue-400">Manual Mode</div>
        <div class="text-sm text-gray-300">Direct control over all operations</div>
        <div class="text-xs text-gray-400">• Step-by-step execution</div>
        <div class="text-xs text-gray-400">• Full user control</div>
        <div class="text-xs text-gray-400">• Review each action</div>
      </div>
    `
  }

  getAutonomousModeTooltip() {
    return `
      <div class="space-y-1">
        <div class="font-medium text-purple-400">Autonomous Mode</div>
        <div class="text-sm text-gray-300">AI handles tasks automatically</div>
        <div class="text-xs text-gray-400">• Automated execution</div>
        <div class="text-xs text-gray-400">• Smart decision making</div>
        <div class="text-xs text-gray-400">• Minimal user intervention</div>
      </div>
    `
  }

  getContextStatusTooltip() {
    const status = this.getBuildStatus()
    const metrics = this.getContextMetrics()
    
    const statusInfo: Record<string, { title: string; description: string; icon: string; color: string }> = {
      'building': {
        title: 'Building Context',
        description: 'Processing selected files...',
        icon: '⚙️',
        color: 'text-blue-400'
      },
      'complete': {
        title: 'Context Ready',
        description: `Built successfully with ${metrics.tokenCount} tokens`,
        icon: '✅',
        color: 'text-green-400'
      },
      'error': {
        title: 'Build Failed',
        description: 'Context build encountered errors',
        icon: '❌',
        color: 'text-red-400'
      },
      'validating': {
        title: 'Validating',
        description: 'Checking context integrity...',
        icon: '🔍',
        color: 'text-yellow-400'
      }
    }
    
    const info = statusInfo[status] || statusInfo['error']
    return `
      <div class="space-y-2">
        <div class="flex items-center space-x-2">
          <span>${info.icon}</span>
          <span class="font-medium ${info.color}">${info.title}</span>
        </div>
        <div class="text-sm text-gray-300">${info.description}</div>
        ${metrics.fileCount ? `<div class="text-xs text-gray-400">Files: ${metrics.fileCount}</div>` : ''}
        ${metrics.tokenCount ? `<div class="text-xs text-gray-400">Tokens: ${metrics.tokenCount.toLocaleString()}</div>` : ''}
        <div class="text-xs text-gray-500 mt-1">Click for details</div>
      </div>
    `
  }

  getContextStatusText() {
    const status = this.getBuildStatus()
    const metrics = this.getContextMetrics()
    
    switch (status) {
      case 'building':
        return 'Building context...'
      case 'complete':
        return `${metrics.tokenCount} tokens`
      case 'error':
        return 'Build failed'
      case 'validating':
        return 'Validating...'
      default:
        return ''
    }
  }

  // Mode handling
  canToggleMode() {
    return !this.getIsTransitioning()
  }

  getModeButtonClasses(isActive: boolean) {
    return {
      'px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 min-w-[80px]': true,
      'bg-blue-600 text-white shadow-lg': isActive && this.getIsManualMode(),
      'bg-purple-600 text-white shadow-lg': isActive && !this.getIsManualMode(),
      'text-gray-300 hover:text-white hover:bg-gray-600': !isActive,
      'opacity-75 pointer-events-none': this.getIsTransitioning()
    }
  }

  getModeToggleClasses() {
    return {
      'flex items-center bg-gray-700 rounded-lg p-1 cursor-pointer transition-all duration-300': true,
      'opacity-75 pointer-events-none': this.getIsTransitioning(),
      [this.getLayoutClasses()]: true
    }
  }
}