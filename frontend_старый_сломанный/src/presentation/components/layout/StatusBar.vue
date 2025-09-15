<template>
  <footer class="bg-gray-800 border-t border-gray-700 px-4 py-2 flex items-center justify-between text-xs text-gray-400 min-h-[32px]">
    <!-- Left Section: System Status -->
    <div class="flex items-center space-x-4">
      <!-- AI Provider Status -->
      <div class="flex items-center space-x-2">
        <div class="w-2 h-2 rounded-full" :class="statusBarService.aiProviderStatusColor" />
        <span>{{ statusBarService.aiProviderName }}</span>
      </div>
      
      <!-- Git Status -->
      <div v-if="statusBarService.gitStatus" class="flex items-center space-x-2">
        <component :is="gitStatusIcon" class="w-3 h-3" />
        <span>{{ statusBarService.gitStatus }}</span>
      </div>
      
      <!-- Build Status -->
      <div v-if="statusBarService.buildStatus" class="flex items-center space-x-2">
        <LoadingSpinner v-if="statusBarService.buildStatus.isRunning" size="xs" />
        <component :is="buildStatusIcon" v-else class="w-3 h-3" />
        <span>{{ statusBarService.buildStatus.message }}</span>
      </div>
    </div>

    <!-- Center Section: Progress Indicators -->
    <div class="flex items-center space-x-6">
      <!-- Context Building Progress -->
      <div v-if="statusBarService.contextProgress.visible" class="flex items-center space-x-2">
        <div class="flex items-center space-x-1">
          <LoadingSpinner size="xs" />
          <span>{{ statusBarService.contextProgress.message }}</span>
        </div>
        <div class="w-24 h-1 bg-gray-700 rounded-full overflow-hidden">
          <div 
            class="h-full bg-gradient-to-r from-blue-500 to-purple-500 transition-all duration-300"
            :style="{ width: `${statusBarService.contextProgress.percentage}%` }"
          />
        </div>
      </div>
      
      <!-- Generation Progress -->
      <div v-if="statusBarService.generationProgress.visible" class="flex items-center space-x-2">
        <div class="flex items-center space-x-1">
          <LoadingSpinner size="xs" color="success" />
          <span>{{ statusBarService.generationProgress.message }}</span>
        </div>
        <div class="w-32 h-1 bg-gray-700 rounded-full overflow-hidden">
          <div 
            class="h-full bg-gradient-to-r from-green-500 to-blue-500 transition-all duration-500"
            :style="{ width: `${statusBarService.generationProgress.percentage}%` }"
          />
        </div>
      </div>
    </div>

    <!-- Right Section: Quick Stats & Notifications -->
    <div class="flex items-center space-x-4">
      <!-- Selected Files Count -->
      <div class="flex items-center space-x-1">
        <DocumentIcon class="w-3 h-3" />
        <span>{{ statusBarService.selectedFilesText }}</span>
      </div>
      
      <!-- Token Count -->
      <div v-if="statusBarService.tokenCount > 0" class="flex items-center space-x-1">
        <CpuChipIcon class="w-3 h-3" />
        <span>{{ statusBarService.formattedTokenCount }}</span>
      </div>
      
      <!-- Estimated Cost -->
      <div v-if="statusBarService.estimatedCost > 0" class="flex items-center space-x-1">
        <CurrencyDollarIcon class="w-3 h-3" />
        <span>${{ statusBarService.formattedCost }}</span>
      </div>
      
      <!-- Last Action -->
      <div v-if="statusBarService.lastAction" class="flex items-center space-x-1 text-gray-500">
        <ClockIcon class="w-3 h-3" />
        <span>{{ statusBarService.lastAction.message }}</span>
        <span>{{ statusBarService.lastAction.timeAgo }}</span>
      </div>
      
      <!-- Notifications -->
      <div v-if="statusBarService.notificationCount > 0" class="flex items-center space-x-1">
        <div class="relative">
          <BellIcon class="w-3 h-3" />
          <div class="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full" />
        </div>
        <span>{{ statusBarService.notificationCount }}</span>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useStatusBarService } from '@/composables/useStatusBarService'

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

// Services and stores
const { statusBarService, contextStore, settingsStore, uiStore } = useStatusBarService()

// Computed Properties that still need direct access to specific store properties
const gitStatusIcon = computed(() => {
  return CodeBracketIcon
})

const buildStatusIcon = computed(() => {
  const iconType = statusBarService.buildStatusIconType
  if (iconType === 'check') return CheckCircleIcon
  if (iconType === 'error') return XCircleIcon
  return ExclamationCircleIcon
})
</script>

<style scoped>
/* Custom scrollbar for long text */
.truncate-with-tooltip {
  @apply truncate cursor-help;
}
</style>