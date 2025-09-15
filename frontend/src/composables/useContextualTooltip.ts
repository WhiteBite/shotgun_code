import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { useWorkspaceStore } from '@/stores/workspace.store'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useProjectStore } from '@/stores/project.store'
import { useTheme } from '@/composables/useTheme'

export interface ContextualTooltipConfig {
  baseContent: string | (() => string)
  contextFactors: string[]
  adaptToMode?: boolean
  adaptToProject?: boolean
  adaptToFiles?: boolean
  adaptToTheme?: boolean
  showShortcuts?: boolean
  showStats?: boolean
  showTips?: boolean
  maxLength?: number
}

export interface TooltipContext {
  workspaceMode: 'manual' | 'autonomous'
  hasProject: boolean
  hasFiles: boolean
  fileCount: number
  tokenCount: number
  projectStatus: string
  theme: string
  isTransitioning: boolean
}

export const useContextualTooltipStore = defineStore('contextualTooltip', () => {
  // Stores
  const workspaceStore = useWorkspaceStore()
  const contextStore = useContextBuilderStore()
  const projectStore = useProjectStore()
  const themeStore = useTheme()
  
  // State
  const tooltipHistory = ref<Map<string, string[]>>(new Map())
  const userPreferences = ref({
    showAdvancedTips: true,
    showKeyboardShortcuts: true,
    showContextStats: true,
    adaptiveContent: true,
    verboseMode: false
  })
  
  // Current context
  const currentContext = computed((): TooltipContext => ({
    workspaceMode: workspaceStore.currentMode,
    hasProject: !!projectStore.currentProject,
    hasFiles: contextStore.hasSelectedFiles,
    fileCount: contextStore.selectedFilesCount,
    tokenCount: contextStore.contextMetrics.tokenCount,
    projectStatus: projectStore.currentProject?.status || 'none',
    theme: themeStore.settings.colorScheme,
    isTransitioning: workspaceStore.isTransitioning
  }))
  
  // Context-aware content generators
  const generateProjectAwareContent = (baseContent: string, elementType: string) => {
    const context = currentContext.value
    let enhancedContent = baseContent
    
    if (!context.hasProject) {
      enhancedContent += '<br><div class=\"text-yellow-400 text-xs mt-1\">💡 Load a project to unlock more features</div>'
    } else if (!context.hasFiles) {
      enhancedContent += '<br><div class=\"text-blue-400 text-xs mt-1\">📁 Select files to build context</div>'
    } else if (context.tokenCount > 10000) {
      enhancedContent += '<br><div class=\"text-orange-400 text-xs mt-1\">⚠️ Large context - consider reducing files</div>'
    }
    
    return enhancedContent
  }
  
  const generateModeAwareContent = (baseContent: string, elementType: string) => {
    const context = currentContext.value
    let enhancedContent = baseContent
    
    if (context.workspaceMode === 'manual') {
      if (elementType === 'button' || elementType === 'action') {
        enhancedContent += '<br><div class=\"text-blue-400 text-xs mt-1\">🔧 Manual mode - You control each step</div>'
      }
    } else if (context.workspaceMode === 'autonomous') {
      if (elementType === 'button' || elementType === 'action') {
        enhancedContent += '<br><div class=\"text-purple-400 text-xs mt-1\">🤖 Autonomous mode - AI will handle automatically</div>'
      }
    }
    
    if (context.isTransitioning) {
      enhancedContent += '<br><div class=\"text-gray-400 text-xs mt-1\">⏳ Mode switching in progress...</div>'
    }
    
    return enhancedContent
  }
  
  const generateThemeAwareContent = (baseContent: string) => {
    const context = currentContext.value
    let enhancedContent = baseContent
    
    if (context.theme === 'dark') {
      // Add dark theme specific tips
      if (Math.random() < 0.3) { // 30% chance
        enhancedContent += '<br><div class=\"text-gray-400 text-xs mt-1\">🌙 Dark theme active</div>'
      }
    }
    
    return enhancedContent
  }
  
  const generateKeyboardShortcuts = (elementType: string, action?: string) => {
    const shortcuts: Record<string, string> = {
      'settings': 'Ctrl+,',
      'export': 'Ctrl+E',
      'help': 'F1',
      'save': 'Ctrl+S',
      'generate': 'Ctrl+Enter',
      'toggle-mode': 'Ctrl+M',
      'new-file': 'Ctrl+N',
      'open': 'Ctrl+O',
      'search': 'Ctrl+F',
      'panel-context': 'Ctrl+Shift+P',
      'panel-results': 'Ctrl+Shift+R',
      'panel-console': 'Ctrl+Shift+C'
    }
    
    const shortcut = action ? shortcuts[action] : shortcuts[elementType]
    return shortcut ? `<kbd class=\"bg-gray-700 px-1 rounded text-xs\">${shortcut}</kbd>` : ''
  }
  
  const generateSmartTips = (elementType: string, context: TooltipContext) => {
    const tips = []
    
    // Context-based tips
    if (!context.hasProject && elementType === 'file-related') {
      tips.push('💡 Load a project first to work with files')
    }
    
    if (context.fileCount > 20) {
      tips.push('🚀 Try using file filters to manage large projects')
    }
    
    if (context.tokenCount > 8000) {
      tips.push('⚡ Large contexts may take longer to process')
    }
    
    // Mode-specific tips
    if (context.workspaceMode === 'manual') {
      tips.push('🔧 Manual mode gives you full control over each step')
    } else {
      tips.push('🤖 Autonomous mode will handle routine tasks automatically')
    }
    
    // Random helpful tips (20% chance)
    if (Math.random() < 0.2) {
      const generalTips = [
        '💡 Use Ctrl+? to see all keyboard shortcuts',
        '🎨 Customize the interface in Settings',
        '📊 Check the status bar for quick project info',
        '🔍 Use the search function to find files quickly',
        '📋 Templates can speed up common tasks'
      ]
      tips.push(generalTips[Math.floor(Math.random() * generalTips.length)])
    }
    
    return tips.slice(0, 2) // Max 2 tips to avoid overwhelming
  }
  
  // Main contextual tooltip generator
  const generateContextualTooltip = (
    config: ContextualTooltipConfig,
    elementType: string = 'generic',
    action?: string
  ): string => {
    const context = currentContext.value
    let content = typeof config.baseContent === 'function' 
      ? config.baseContent() 
      : config.baseContent
    
    // Apply contextual enhancements
    if (config.adaptToProject) {
      content = generateProjectAwareContent(content, elementType)
    }
    
    if (config.adaptToMode) {
      content = generateModeAwareContent(content, elementType)
    }
    
    if (config.adaptToTheme) {
      content = generateThemeAwareContent(content)
    }
    
    // Add keyboard shortcuts if enabled
    if (config.showShortcuts && userPreferences.value.showKeyboardShortcuts) {
      const shortcut = generateKeyboardShortcuts(elementType, action)
      if (shortcut) {
        content += `<br><div class=\"text-xs text-gray-400 mt-1\">${shortcut}</div>`
      }
    }
    
    // Add contextual statistics
    if (config.showStats && userPreferences.value.showContextStats) {
      if (context.hasFiles) {
        content += `<br><div class=\"text-xs text-gray-400 mt-1\">📊 ${context.fileCount} files, ${context.tokenCount.toLocaleString()} tokens</div>`
      }
    }
    
    // Add smart tips
    if (config.showTips && userPreferences.value.showAdvancedTips) {
      const tips = generateSmartTips(elementType, context)
      if (tips.length > 0) {
        content += '<br><div class=\"border-t border-gray-600 mt-2 pt-1\">'
        tips.forEach(tip => {
          content += `<div class=\"text-xs text-gray-400\">${tip}</div>`
        })
        content += '</div>'
      }
    }
    
    // Apply length limit
    if (config.maxLength && content.length > config.maxLength) {
      content = content.substring(0, config.maxLength) + '...'
    }
    
    // Track tooltip usage
    trackTooltipUsage(elementType, content)
    
    return content
  }
  
  // Usage tracking for analytics and improvements
  const trackTooltipUsage = (elementType: string, content: string) => {
    const history = tooltipHistory.value.get(elementType) || []
    history.push(content)
    
    // Keep only last 10 entries per element type
    if (history.length > 10) {
      history.shift()
    }
    
    tooltipHistory.value.set(elementType, history)
  }
  
  // Predefined contextual tooltip configurations
  const tooltipConfigs = {
    fileAction: {
      adaptToProject: true,
      adaptToFiles: true,
      showShortcuts: true,
      showStats: true,
      contextFactors: ['project', 'files']
    },
    modeAction: {
      adaptToMode: true,
      showShortcuts: true,
      showTips: true,
      contextFactors: ['mode', 'transition']
    },
    projectAction: {
      adaptToProject: true,
      adaptToMode: true,
      showStats: true,
      showTips: true,
      contextFactors: ['project', 'mode', 'files']
    },
    uiElement: {
      adaptToTheme: true,
      showShortcuts: true,
      contextFactors: ['theme']
    }
  }
  
  // Quick tooltip generators for common use cases
  const createFileTooltip = (baseContent: string, action?: string) => {
    return generateContextualTooltip({
      baseContent,
      ...tooltipConfigs.fileAction
    }, 'file-action', action)
  }
  
  const createModeTooltip = (baseContent: string, action?: string) => {
    return generateContextualTooltip({
      baseContent,
      ...tooltipConfigs.modeAction
    }, 'mode-action', action)
  }
  
  const createProjectTooltip = (baseContent: string, action?: string) => {
    return generateContextualTooltip({
      baseContent,
      ...tooltipConfigs.projectAction
    }, 'project-action', action)
  }
  
  const createUITooltip = (baseContent: string, action?: string) => {
    return generateContextualTooltip({
      baseContent,
      ...tooltipConfigs.uiElement
    }, 'ui-element', action)
  }
  
  // Settings management
  const updatePreferences = (newPrefs: Partial<typeof userPreferences.value>) => {
    userPreferences.value = { ...userPreferences.value, ...newPrefs }
    localStorage.setItem('tooltip-preferences', JSON.stringify(userPreferences.value))
  }
  
  const loadPreferences = () => {
    const saved = localStorage.getItem('tooltip-preferences')
    if (saved) {
      try {
        userPreferences.value = { ...userPreferences.value, ...JSON.parse(saved) }
      } catch (error) {
        console.warn('Failed to load tooltip preferences:', error)
      }
    }
  }
  
  // Initialize
  loadPreferences()
  
  return {
    // State
    currentContext,
    userPreferences,
    tooltipHistory,
    
    // Main API
    generateContextualTooltip,
    
    // Quick generators
    createFileTooltip,
    createModeTooltip,
    createProjectTooltip,
    createUITooltip,
    
    // Utilities
    updatePreferences,
    trackTooltipUsage
  }
})

// Composable hook
export function useContextualTooltip() {
  return useContextualTooltipStore()
}