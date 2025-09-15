<template>
  <header class="bg-gray-800 border-b border-gray-700 px-4 py-3 flex items-center justify-between min-h-[64px]">
    <!-- Project Info Card -->
    <div class="flex items-center space-x-4">
      <div 
        class="flex items-center space-x-3 cursor-pointer hover:bg-gray-700/50 rounded-lg px-2 py-1 transition-colors"
        v-tooltip="{
          content: getProjectTooltipContent,
          position: 'bottom',
          maxWidth: 300,
          allowHTML: true
        }"
        @click="openProjectSettings"
      >
        <!-- Project Icon -->
        <div 
          class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center"
          v-tooltip="'Project root directory'"
        >
          <FolderIcon class="w-5 h-5 text-white" />
        </div>
        
        <!-- Project Details -->
        <div class="flex flex-col">
          <h1 class="text-lg font-semibold text-gray-100 truncate max-w-xs">
            {{ projectName }}
          </h1>
          <div class="flex items-center space-x-2 text-xs text-gray-400">
            <span 
              class="truncate max-w-64 hover:text-gray-200 transition-colors" 
              :title="projectPath"
              v-tooltip="{
                content: `Full path: ${projectPath}`,
                position: 'bottom'
              }"
            >
              {{ truncatedPath }}
            </span>
            <div 
              class="flex items-center space-x-1"
              v-tooltip="{
                content: getStatusTooltipContent,
                position: 'bottom',
                allowHTML: true
              }"
            >
              <StatusIndicator :status="projectStatus" />
              <span class="capitalize">{{ projectStatus }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Mode Toggle -->
    <div class="flex items-center space-x-6">
      <div class="flex items-center space-x-3">
        <span 
          class="text-sm text-gray-300 font-medium"
          v-tooltip="'Switch between manual control and autonomous AI mode'"
        >
          Mode:
        </span>
        <div class="relative">
          <div 
            class="flex items-center bg-gray-700 rounded-lg p-1 cursor-pointer transition-all duration-300"
            :class="[
              workspaceStore.isTransitioning ? 'opacity-75 pointer-events-none' : '',
              workspaceStore.layoutClasses
            ]"
            @click="handleModeToggle"
            v-tooltip="{
              content: getModeToggleTooltip,
              position: 'bottom',
              maxWidth: 250,
              allowHTML: true,
              interactive: true
            }"
          >
            <!-- Manual Mode Button -->
            <button
              class="px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 min-w-[80px]"
              :class="[
                workspaceStore.isManualMode 
                  ? 'bg-blue-600 text-white shadow-lg' 
                  : 'text-gray-300 hover:text-white hover:bg-gray-600'
              ]"
              :disabled="workspaceStore.isTransitioning"
              @click.stop="workspaceStore.setMode('manual')"
              v-tooltip="{
                content: getManualModeTooltip,
                position: 'bottom',
                maxWidth: 200,
                allowHTML: true
              }"
            >
              <div class="flex items-center justify-center space-x-1">
                <CodeBracketIcon class="w-4 h-4" />
                <span>Manual</span>
              </div>
            </button>
            
            <!-- Autonomous Mode Button -->
            <button
              class="px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 min-w-[80px]"
              :class="[
                workspaceStore.isAutonomousMode 
                  ? 'bg-purple-600 text-white shadow-lg' 
                  : 'text-gray-300 hover:text-white hover:bg-gray-600'
              ]"
              :disabled="workspaceStore.isTransitioning"
              @click.stop="workspaceStore.setMode('autonomous')"
              v-tooltip="{
                content: getAutonomousModeTooltip,
                position: 'bottom',
                maxWidth: 200,
                allowHTML: true
              }"
            >
              <div class="flex items-center justify-center space-x-1">
                <CpuChipIcon class="w-4 h-4" />
                <span>Auto</span>
              </div>
            </button>
          </div>
          
          <!-- Loading indicator during transition -->
          <div 
            v-if="workspaceStore.isTransitioning"
            class="absolute inset-0 flex items-center justify-center bg-gray-700 bg-opacity-50 rounded-lg"
          >
            <LoadingSpinner class="w-4 h-4" />
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="flex items-center space-x-2">
      <!-- Language Switcher -->
      <div class="relative">
        <IconButton
          icon="LanguageIcon"
          v-tooltip="{
            content: 'Switch Language (Переключить язык)',
            position: 'bottom',
            allowHTML: true
          }"
          @click="toggleLanguageMenu"
        />
        
        <!-- Language Dropdown Menu -->
        <div 
          v-if="isLanguageMenuOpen"
          class="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-gray-800 border border-gray-700 z-50"
        >
          <div class="py-1">
            <button 
              class="w-full px-4 py-2 text-sm text-left text-gray-300 hover:bg-gray-700 flex items-center space-x-2"
              @click="changeLanguage('en')"
            >
              <span class="opacity-70">🇬🇧</span>
              <span>English</span>
              <span v-if="currentLanguage === 'en'" class="ml-auto text-blue-400">✓</span>
            </button>
            <button 
              class="w-full px-4 py-2 text-sm text-left text-gray-300 hover:bg-gray-700 flex items-center space-x-2"
              @click="changeLanguage('ru')"
            >
              <span class="opacity-70">🇷🇺</span>
              <span>Русский</span>
              <span v-if="currentLanguage === 'ru'" class="ml-auto text-blue-400">✓</span>
            </button>
          </div>
        </div>
      </div>
      
      <!-- Context Build Status -->
      <div 
        v-if="contextStore.buildStatus !== 'idle'" 
        class="flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-gray-700 cursor-pointer hover:bg-gray-600 transition-colors"
        v-tooltip="{
          content: getContextStatusTooltip,
          position: 'bottom',
          maxWidth: 250,
          allowHTML: true
        }"
        @click="openContextDetails"
      >
        <LoadingSpinner 
          v-if="contextStore.buildStatus === 'building'" 
          class="w-4 h-4 text-blue-400" 
        />
        <CheckCircleIcon 
          v-else-if="contextStore.buildStatus === 'complete'" 
          class="w-4 h-4 text-green-400" 
        />
        <ExclamationCircleIcon 
          v-else-if="contextStore.buildStatus === 'error'" 
          class="w-4 h-4 text-red-400" 
        />
        <span class="text-xs text-gray-300">
          {{ getContextStatusText() }}
        </span>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center space-x-1">
        <!-- Settings -->
        <IconButton
          icon="Cog6ToothIcon"
          v-tooltip="{
            content: 'Open Settings<br><kbd>Ctrl+,</kbd>',
            position: 'bottom',
            allowHTML: true
          }"
          @click="openSettings"
        />
        
        <!-- Export -->
        <IconButton
          icon="ArrowUpTrayIcon"
          v-tooltip="{
            content: contextStore.hasSelectedFiles ? 
              'Export current context<br><kbd>Ctrl+E</kbd>' : 
              'No files selected for export',
            position: 'bottom',
            allowHTML: true
          }"
          :disabled="!contextStore.hasSelectedFiles"
          @click="openExport"
        />
        
        <!-- Help -->
        <IconButton
          icon="QuestionMarkCircleIcon"
          v-tooltip="{
            content: 'Help & Keyboard Shortcuts<br><kbd>F1</kbd> or <kbd>?</kbd>',
            position: 'bottom',
            allowHTML: true
          }"
          @click="showHelp"
        />
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useWorkspaceStore } from '@/stores/workspace.store'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useProjectStore } from '@/stores/project.store'
import { useUiStore } from '@/stores/ui.store'

// Icons
import { 
  FolderIcon, 
  CodeBracketIcon, 
  CpuChipIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
  Cog6ToothIcon,
  ArrowUpTrayIcon,
  QuestionMarkCircleIcon,
  LanguageIcon
} from '@heroicons/vue/24/outline'

// Components
import IconButton from '@/presentation/components/shared/IconButton.vue'
import StatusIndicator from '@/presentation/components/shared/StatusIndicator.vue'
import LoadingSpinner from '@/presentation/components/shared/LoadingSpinner.vue'

// Stores
const workspaceStore = useWorkspaceStore()
const contextStore = useContextBuilderStore()
const projectStore = useProjectStore()
const uiStore = useUiStore()

// Computed Properties
const projectName = computed(() => {
  const name = projectStore.currentProject?.name
  return name || 'No Project'
})

const projectPath = computed(() => {
  return projectStore.currentProject?.path || ''
})

const projectStatus = computed(() => {
  if (!projectStore.currentProject) return 'none'
  
  // Determine project health based on various factors
  const hasFiles = contextStore.selectedFilesCount > 0
  const hasContext = contextStore.buildStatus === 'complete'
  const hasErrors = contextStore.buildStatus === 'error'
  
  if (hasErrors) return 'error'
  if (hasContext && hasFiles) return 'healthy'
  if (hasFiles) return 'ready'
  return 'idle'
})

const truncatedPath = computed(() => {
  const path = projectPath.value
  if (!path) return ''
  
  const maxLength = 40
  if (path.length <= maxLength) return path
  
  const parts = path.split('/')
  if (parts.length > 2) {
    return `.../${parts.slice(-2).join('/')}`
  }
  
  return `...${path.slice(-maxLength)}`
})

// Enhanced tooltip content
const getProjectTooltipContent = computed(() => {
  const project = projectStore.currentProject
  if (!project) return 'No project loaded'
  
  return `
    <div class="space-y-2">
      <div class="font-medium">${project.name}</div>
      <div class="text-sm text-gray-400">Path: ${project.path}</div>
      <div class="text-sm text-gray-400">Files: ${contextStore.selectedFilesCount}</div>
      <div class="text-xs text-gray-500 mt-2">Click to open project settings</div>
    </div>
  `
})

const getStatusTooltipContent = computed(() => {
  const status = projectStatus.value
  const statusInfo = {
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
})

const getModeToggleTooltip = computed(() => {
  return `
    <div class="space-y-2">
      <div class="font-medium">Workspace Mode</div>
      <div class="text-sm text-gray-300">
        Current: <span class="font-medium ${workspaceStore.isManualMode ? 'text-blue-400' : 'text-purple-400'}">
          ${workspaceStore.isManualMode ? 'Manual' : 'Autonomous'}
        </span>
      </div>
      <div class="text-xs text-gray-400">
        Click to toggle between modes<br>
        <kbd class="bg-gray-700 px-1 rounded">Ctrl+M</kbd> to switch
      </div>
    </div>
  `
})

const getManualModeTooltip = computed(() => {
  return `
    <div class="space-y-1">
      <div class="font-medium text-blue-400">Manual Mode</div>
      <div class="text-sm text-gray-300">Direct control over all operations</div>
      <div class="text-xs text-gray-400">• Step-by-step execution</div>
      <div class="text-xs text-gray-400">• Full user control</div>
      <div class="text-xs text-gray-400">• Review each action</div>
    </div>
  `
})

const getAutonomousModeTooltip = computed(() => {
  return `
    <div class="space-y-1">
      <div class="font-medium text-purple-400">Autonomous Mode</div>
      <div class="text-sm text-gray-300">AI handles tasks automatically</div>
      <div class="text-xs text-gray-400">• Automated execution</div>
      <div class="text-xs text-gray-400">• Smart decision making</div>
      <div class="text-xs text-gray-400">• Minimal user intervention</div>
    </div>
  `
})

const getContextStatusTooltip = computed(() => {
  const status = contextStore.buildStatus
  const metrics = contextStore.contextMetrics
  
  const statusInfo = {
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
})

// Language switcher state
const isLanguageMenuOpen = ref(false)
const currentLanguage = ref(localStorage.getItem('app_language') || 'en')

// Click outside handler for language menu
const closeLanguageMenu = (event: MouseEvent) => {
  if (isLanguageMenuOpen.value && event.target) {
    isLanguageMenuOpen.value = false
  }
}

// Add event listener when component is mounted
onMounted(() => {
  document.addEventListener('click', closeLanguageMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeLanguageMenu)
})

// Methods
const toggleLanguageMenu = (event: MouseEvent) => {
  // Stop propagation to prevent immediate closing
  event.stopPropagation()
  isLanguageMenuOpen.value = !isLanguageMenuOpen.value
}

const changeLanguage = (lang: string) => {
  currentLanguage.value = lang
  localStorage.setItem('app_language', lang)
  isLanguageMenuOpen.value = false
  
  // Here you would typically also update your i18n instance
  // i18n.global.locale.value = lang
  
  // For demonstration, we'll show a toast message
  uiStore.addToast(
    lang === 'en' ? 'Language changed to English' : 'Язык изменен на Русский',
    'success'
  )
}
const handleModeToggle = () => {
  if (workspaceStore.isTransitioning) return
  workspaceStore.toggleMode()
}

const getContextStatusText = () => {
  switch (contextStore.buildStatus) {
    case 'building':
      return 'Building context...'
    case 'complete':
      return `${contextStore.contextMetrics.tokenCount} tokens`
    case 'error':
      return 'Build failed'
    case 'validating':
      return 'Validating...'
    default:
      return ''
  }
}

const openSettings = () => {
  uiStore.openDrawer('settings')
}

const openExport = () => {
  // This would open export modal or drawer
  console.log('Open export functionality')
}

const showHelp = () => {
  // This would show help modal with keyboard shortcuts
  console.log('Show help modal')
}

const openProjectSettings = () => {
  // Open project-specific settings
  console.log('Open project settings')
}

const openContextDetails = () => {
  // Open detailed context information
  console.log('Open context details')
}
</script>

<style scoped>
.workspace-manual {
  @apply border-l-4 border-blue-500;
}

.workspace-autonomous {
  @apply border-l-4 border-purple-500;
}

.workspace-transitioning {
  @apply animate-pulse;
}

.workspace-animated * {
  @apply transition-all duration-300;
}
</style>