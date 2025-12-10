/**
 * Command Service for CommandPalette
 * Provides context-aware commands with real actions
 */

import { useFileStore } from '@/features/files/model/file.store'
import { useProjectStore } from '@/stores/project.store'
import { useUIStore } from '@/stores/ui.store'
import {
  ArrowPathIcon,
  CodeBracketIcon,
  CogIcon,
  DocumentTextIcon,
  FolderOpenIcon,
  HomeIcon,
  MagnifyingGlassIcon,
  QuestionMarkCircleIcon,
  Squares2X2Icon
} from '@heroicons/vue/24/outline'

export interface Command {
  id: string
  name: string
  description: string
  shortcut?: string
  icon: any
  action: () => void | Promise<void>
  category?: string
  when?: () => boolean // Conditional visibility
}

export class CommandService {
  /**
   * Get all available commands based on current context
   */
  getCommands(): Command[] {
    const projectStore = useProjectStore()
    const fileStore = useFileStore()
    const uiStore = useUIStore()

    const commands: Command[] = [
      // Project commands
      {
        id: 'open-project',
        name: 'Open Project',
        description: 'Open a project folder',
        shortcut: 'Ctrl+O',
        icon: FolderOpenIcon,
        category: 'Project',
        action: () => {
          projectStore.clearProject()
          uiStore.addToast('Select a project folder', 'info')
        }
      },
      {
        id: 'change-project',
        name: 'Change Project',
        description: 'Switch to a different project',
        shortcut: 'Ctrl+Shift+P',
        icon: ArrowPathIcon,
        category: 'Project',
        when: () => projectStore.hasProject,
        action: () => {
          projectStore.clearProject()
          uiStore.addToast('Select a new project', 'info')
        }
      },
      {
        id: 'close-project',
        name: 'Close Project',
        description: 'Close current project',
        shortcut: 'Ctrl+Shift+W',
        icon: HomeIcon,
        category: 'Project',
        when: () => projectStore.hasProject,
        action: () => {
          projectStore.clearProject()
          uiStore.addToast('Project closed', 'success')
        }
      },

      // File commands
      {
        id: 'search-files',
        name: 'Search Files',
        description: 'Search for files in current project',
        shortcut: 'Ctrl+P',
        icon: MagnifyingGlassIcon,
        category: 'Files',
        when: () => projectStore.hasProject,
        action: () => {
          // TODO: Implement file search modal
          uiStore.addToast('File search coming soon!', 'info')
        }
      },
      {
        id: 'refresh-files',
        name: 'Refresh File Tree',
        description: 'Reload files from disk',
        shortcut: 'Ctrl+R',
        icon: ArrowPathIcon,
        category: 'Files',
        when: () => projectStore.hasProject,
        action: async () => {
          try {
            await fileStore.refreshFileTree()
            uiStore.addToast('File tree refreshed', 'success')
          } catch (err) {
            uiStore.addToast('Failed to refresh files', 'error')
          }
        }
      },

      // Code actions
      {
        id: 'analyze-code',
        name: 'Analyze Code',
        description: 'Run code analysis on project',
        shortcut: 'Ctrl+Shift+A',
        icon: CodeBracketIcon,
        category: 'Code',
        when: () => projectStore.hasProject,
        action: () => {
          // TODO: Trigger code analysis
          uiStore.addToast('Code analysis started...', 'info')
        }
      },
      {
        id: 'generate-code',
        name: 'Generate Code',
        description: 'Generate code from task description',
        shortcut: 'Ctrl+G',
        icon: Squares2X2Icon,
        category: 'Code',
        when: () => projectStore.hasProject,
        action: () => {
          // TODO: Open code generation dialog
          uiStore.addToast('Code generation coming soon!', 'info')
        }
      },

      // App commands
      {
        id: 'reload-window',
        name: 'Reload Window',
        description: 'Reload the application',
        shortcut: 'Ctrl+Shift+R',
        icon: ArrowPathIcon,
        category: 'App',
        action: () => {
          window.location.reload()
        }
      },
      {
        id: 'settings',
        name: 'Settings',
        description: 'Open application settings',
        shortcut: 'Ctrl+,',
        icon: CogIcon,
        category: 'App',
        action: () => {
          // TODO: Open settings modal
          uiStore.addToast('Settings coming soon!', 'info')
        }
      },
      {
        id: 'keyboard-shortcuts',
        name: 'Keyboard Shortcuts',
        description: 'Show keyboard shortcuts guide',
        shortcut: 'Ctrl+/',
        icon: QuestionMarkCircleIcon,
        category: 'Help',
        action: () => {
          // TODO: Trigger KeyboardShortcutsModal
          uiStore.addToast('Press Ctrl+/ to see shortcuts', 'info')
        }
      },
      {
        id: 'new-document',
        name: 'New Document',
        description: 'Create a new document',
        shortcut: 'Ctrl+N',
        icon: DocumentTextIcon,
        category: 'Files',
        action: () => {
          // TODO: Create new document
          uiStore.addToast('New document feature coming soon!', 'info')
        }
      }
    ]

    // Filter commands based on `when` condition
    return commands.filter(cmd => !cmd.when || cmd.when())
  }

  /**
   * Get commands grouped by category
   */
  getCommandsByCategory(): Record<string, Command[]> {
    const commands = this.getCommands()
    const grouped: Record<string, Command[]> = {}

    for (const command of commands) {
      const category = command.category || 'Other'
      if (!grouped[category]) {
        grouped[category] = []
      }
      grouped[category].push(command)
    }

    return grouped
  }

  /**
   * Execute a command by ID
   */
  async executeCommand(commandId: string): Promise<boolean> {
    const commands = this.getCommands()
    const command = commands.find(cmd => cmd.id === commandId)

    if (!command) {
      console.warn(`Command not found: ${commandId}`)
      return false
    }

    try {
      await command.action()
      return true
    } catch (err) {
      console.error(`Failed to execute command ${commandId}:`, err)
      return false
    }
  }
}

// Singleton instance
export const commandService = new CommandService()
