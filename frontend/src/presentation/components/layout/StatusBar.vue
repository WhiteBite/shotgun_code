<template>
  <footer class="bg-gray-800 border-t border-gray-700 px-4 py-2 flex items-center justify-between text-xs text-gray-400 min-h-[32px]">
    <!-- Left Section: System Status -->
    <div class="flex items-center space-x-4">
      <!-- AI Provider Status -->
      <div class="flex items-center space-x-2">
        <div class="w-2 h-2 rounded-full" :class="aiProviderStatusColor" />
        <span>{{ aiProviderName }}</span>
      </div>
      
      <!-- Git Status -->
      <div v-if="gitStatus" class="flex items-center space-x-2">
        <component :is="gitStatusIcon" class="w-3 h-3" />
        <span>{{ gitStatus }}</span>
      </div>
      
      <!-- Build Status -->
      <div v-if="buildStatus" class="flex items-center space-x-2">
        <LoadingSpinner v-if="buildStatus.isRunning" size="xs" />
        <component v-else :is="buildStatusIcon" class="w-3 h-3" />
        <span>{{ buildStatus.message }}</span>
      </div>
    </div>

    <!-- Center Section: Progress Indicators -->
    <div class="flex items-center space-x-6">
      <!-- Context Building Progress -->
      <div v-if="contextProgress.visible" class="flex items-center space-x-2">
        <div class="flex items-center space-x-1">
          <LoadingSpinner size="xs" />
          <span>{{ contextProgress.message }}</span>
        </div>
        <div class="w-24 h-1 bg-gray-700 rounded-full overflow-hidden">
          <div 
            class="h-full bg-gradient-to-r from-blue-500 to-purple-500 transition-all duration-300"
            :style="{ width: `${contextProgress.percentage}%` }"
          />
        </div>
      </div>
      
      <!-- Generation Progress -->
      <div v-if="generationProgress.visible" class="flex items-center space-x-2">
        <div class="flex items-center space-x-1">
          <LoadingSpinner size="xs" color="success" />
          <span>{{ generationProgress.message }}</span>
        </div>
        <div class="w-32 h-1 bg-gray-700 rounded-full overflow-hidden">
          <div 
            class="h-full bg-gradient-to-r from-green-500 to-blue-500 transition-all duration-500"
            :style="{ width: `${generationProgress.percentage}%` }"
          />
        </div>
      </div>
    </div>

    <!-- Right Section: Quick Stats & Notifications -->
    <div class="flex items-center space-x-4">
      <!-- Selected Files Count -->
      <div class="flex items-center space-x-1">
        <DocumentIcon class="w-3 h-3" />
        <span>{{ selectedFilesText }}</span>
      </div>
      
      <!-- Token Count -->
      <div v-if="tokenCount > 0" class="flex items-center space-x-1">
        <CpuChipIcon class="w-3 h-3" />
        <span>{{ formattedTokenCount }}</span>
      </div>
      
      <!-- Estimated Cost -->
      <div v-if="estimatedCost > 0" class="flex items-center space-x-1">
        <CurrencyDollarIcon class="w-3 h-3" />
        <span>${{ formattedCost }}</span>
      </div>
      
      <!-- Last Action -->
      <div v-if="lastAction" class="flex items-center space-x-1 text-gray-500">
        <ClockIcon class="w-3 h-3" />
        <span>{{ lastAction.message }}</span>
        <span>{{ lastAction.timeAgo }}</span>
      </div>
      
      <!-- Notifications -->
      <div v-if="notificationCount > 0" class="flex items-center space-x-1">
        <div class="relative">
          <BellIcon class="w-3 h-3" />
          <div class="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full" />
        </div>
        <span>{{ notificationCount }}</span>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useWorkspaceStore } from '@/stores/workspace.store'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useSettingsStore } from '@/stores/settings.store'
import { useUiStore } from '@/stores/ui.store'

// Icons
import {
  DocumentIcon,
  CpuChipIcon,
  CurrencyDollarIcon,
  ClockIcon,
  BellIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
  XCircleIcon
} from '@heroicons/vue/24/outline'
import {
  CodeBracketIcon,
  CommandLineIcon
} from '@heroicons/vue/24/solid'

import LoadingSpinner from '@/presentation/components/shared/LoadingSpinner.vue'

// Stores
const workspaceStore = useWorkspaceStore()
const contextStore = useContextBuilderStore()
const settingsStore = useSettingsStore()
const uiStore = useUiStore()

// Computed Properties
const aiProviderName = computed(() => {
  // Get current AI provider from settings
  return settingsStore.settings?.aiProvider || 'No Provider'
})

const aiProviderStatusColor = computed(() => {
  // This would check actual AI provider connectivity
  return 'bg-green-400' // Connected
  // return 'bg-yellow-400' // Warning
  // return 'bg-red-400' // Disconnected
})

const gitStatus = computed(() => {
  // This would show current git status
  return null // 'main • 2 ahead'
})

const gitStatusIcon = computed(() => {
  return CodeBracketIcon
})

const buildStatus = computed(() => {
  if (contextStore.buildStatus === 'building') {
    return {
      isRunning: true,
      message: 'Building context...'
    }
  }
  
  if (contextStore.buildStatus === 'error') {
    return {
      isRunning: false,
      message: 'Build failed'
    }
  }
  
  return null
})

const buildStatusIcon = computed(() => {
  if (contextStore.buildStatus === 'complete') return CheckCircleIcon
  if (contextStore.buildStatus === 'error') return XCircleIcon
  return ExclamationCircleIcon
})

const contextProgress = computed(() => {
  if (contextStore.buildStatus === 'building') {
    return {
      visible: true,
      message: 'Building context',
      percentage: 45 // This would be calculated based on actual progress
    }
  }
  
  return { visible: false, message: '', percentage: 0 }
})

const generationProgress = computed(() => {
  // This would track AI generation progress
  return { visible: false, message: '', percentage: 0 }
})

const selectedFilesText = computed(() => {
  const count = contextStore.selectedFilesCount
  if (count === 0) return '0 files'
  return `${count} file${count === 1 ? '' : 's'}`
})

const tokenCount = computed(() => {
  return contextStore.contextMetrics.tokenCount
})

const formattedTokenCount = computed(() => {
  const count = tokenCount.value
  if (count >= 1000) {
    return `${(count / 1000).toFixed(1)}k tokens`
  }
  return `${count} tokens`
})

const estimatedCost = computed(() => {
  return contextStore.contextMetrics.estimatedCost
})

const formattedCost = computed(() => {
  return estimatedCost.value.toFixed(4)
})

const lastAction = computed(() => {
  if (contextStore.lastContextGeneration) {
    const timeAgo = getTimeAgo(contextStore.lastContextGeneration)
    return {
      message: 'Context built',
      timeAgo
    }
  }
  return null
})

const notificationCount = computed(() => {
  return uiStore.toasts.length
})

// Helper functions
const getTimeAgo = (date: Date): string => {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  
  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes}m ago`
  
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}
</script>

<style scoped>
/* Custom scrollbar for long text */
.truncate-with-tooltip {
  @apply truncate cursor-help;
}
</style>