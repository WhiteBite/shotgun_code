<template>
  <div class="bg-gray-800 border border-gray-700 rounded-lg overflow-hidden">
    <!-- Header -->
    <div class="p-4 border-b border-gray-700 bg-gray-800">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-medium text-gray-200">Context Summary</h3>
        <div class="flex items-center space-x-2">
          <!-- Context Health Indicator -->
          <div class="flex items-center space-x-2">
            <div 
              class="w-2 h-2 rounded-full"
              :class="healthIndicatorColor"
            />
            <span class="text-xs text-gray-400">{{ healthStatus }}</span>
          </div>
          
          <!-- Quick Actions -->
          <IconButton
            icon="ArrowPathIcon"
            size="sm"
            tooltip="Refresh Metrics"
            @click="refreshMetrics"
          />
          <IconButton
            icon="ShareIcon"
            size="sm"
            tooltip="Export Summary"
            @click="exportSummary"
          />
        </div>
      </div>
    </div>
    
    <!-- Main Metrics Grid -->
    <div class="p-4">
      <div class="grid grid-cols-2 gap-4 mb-4">
        <!-- Files Count -->
        <div class="bg-gray-700 rounded-lg p-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-2xl font-bold text-gray-100">
                {{ contextStore.contextMetrics.fileCount }}
              </div>
              <div class="text-sm text-gray-400">Files Selected</div>
            </div>
            <DocumentIcon class="w-8 h-8 text-blue-400" />
          </div>
          <div class="mt-2 text-xs text-gray-500">
            {{ percentageOfTotal }}% of project
          </div>
        </div>
        
        <!-- Token Count -->
        <div class="bg-gray-700 rounded-lg p-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-2xl font-bold text-gray-100">
                {{ formattedTokenCount }}
              </div>
              <div class="text-sm text-gray-400">Total Tokens</div>
            </div>
            <CpuChipIcon class="w-8 h-8 text-green-400" />
          </div>
          <div class="mt-2">
            <div class="flex justify-between text-xs">
              <span class="text-gray-500">{{ tokenUsagePercentage }}% of limit</span>
              <span class="text-gray-500">{{ contextStore.maxTokenLimit }} max</span>
            </div>
            <div class="w-full h-1 bg-gray-600 rounded-full mt-1">
              <div 
                class="h-full rounded-full transition-all duration-300"
                :class="tokenUsageColor"
                :style="{ width: `${Math.min(tokenUsagePercentage, 100)}%` }"
              />
            </div>
          </div>
        </div>
        
        <!-- Build Time -->
        <div class="bg-gray-700 rounded-lg p-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-2xl font-bold text-gray-100">
                {{ formattedBuildTime }}
              </div>
              <div class="text-sm text-gray-400">Build Time</div>
            </div>
            <ClockIcon class="w-8 h-8 text-yellow-400" />
          </div>
          <div class="mt-2 text-xs text-gray-500">
            {{ buildPerformanceText }}
          </div>
        </div>
        
        <!-- Estimated Cost -->
        <div class="bg-gray-700 rounded-lg p-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-2xl font-bold text-gray-100">
                ${{ formattedCost }}
              </div>
              <div class="text-sm text-gray-400">Est. Cost</div>
            </div>
            <CurrencyDollarIcon class="w-8 h-8 text-purple-400" />
          </div>
          <div class="mt-2 text-xs text-gray-500">
            Per API call
          </div>
        </div>
      </div>
      
      <!-- Detailed Breakdown -->
      <div class="space-y-4">
        <!-- File Type Distribution -->
        <div>
          <h4 class="text-sm font-medium text-gray-300 mb-2">File Type Distribution</h4>
          <div class="space-y-2">
            <div
              v-for="type in fileTypeDistribution"
              :key="type.extension"
              class="flex items-center justify-between"
            >
              <div class="flex items-center space-x-2">
                <div 
                  class="w-3 h-3 rounded-full"
                  :style="{ backgroundColor: type.color }"
                />
                <span class="text-sm text-gray-300">.{{ type.extension }}</span>
              </div>
              <div class="flex items-center space-x-2">
                <span class="text-sm text-gray-400">{{ type.count }}</span>
                <div class="w-16 h-2 bg-gray-600 rounded-full">
                  <div 
                    class="h-full rounded-full"
                    :style="{ 
                      width: `${type.percentage}%`, 
                      backgroundColor: type.color 
                    }"
                  />
                </div>
                <span class="text-xs text-gray-500 w-8">{{ type.percentage }}%</span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Context Quality Metrics -->
        <div>
          <h4 class="text-sm font-medium text-gray-300 mb-2">Context Quality</h4>
          <div class="grid grid-cols-3 gap-3">
            <div class="text-center">
              <div class="text-lg font-bold" :class="getScoreColor(diversityScore)">
                {{ diversityScore }}
              </div>
              <div class="text-xs text-gray-400">Diversity</div>
            </div>
            <div class="text-center">
              <div class="text-lg font-bold" :class="getScoreColor(completenessScore)">
                {{ completenessScore }}
              </div>
              <div class="text-xs text-gray-400">Completeness</div>
            </div>
            <div class="text-center">
              <div class="text-lg font-bold" :class="getScoreColor(contextStore.contextHealth)">
                {{ contextStore.contextHealth }}
              </div>
              <div class="text-xs text-gray-400">Overall</div>
            </div>
          </div>
        </div>
        
        <!-- Warnings and Recommendations -->
        <div v-if="warnings.length > 0 || recommendations.length > 0">
          <!-- Warnings -->
          <div v-if="warnings.length > 0" class="mb-3">
            <h4 class="text-sm font-medium text-yellow-400 mb-2 flex items-center">
              <ExclamationTriangleIcon class="w-4 h-4 mr-1" />
              Warnings
            </h4>
            <div class="space-y-1">
              <div
                v-for="warning in warnings"
                :key="warning.id"
                class="flex items-start space-x-2 p-2 bg-yellow-900 bg-opacity-20 border border-yellow-700 rounded text-sm"
              >
                <ExclamationTriangleIcon class="w-4 h-4 text-yellow-400 flex-shrink-0 mt-0.5" />
                <span class="text-yellow-200">{{ warning.message }}</span>
              </div>
            </div>
          </div>
          
          <!-- Recommendations -->
          <div v-if="recommendations.length > 0">
            <h4 class="text-sm font-medium text-blue-400 mb-2 flex items-center">
              <LightBulbIcon class="w-4 h-4 mr-1" />
              Recommendations
            </h4>
            <div class="space-y-1">
              <div
                v-for="recommendation in recommendations"
                :key="recommendation.id"
                class="flex items-start justify-between p-2 bg-blue-900 bg-opacity-20 border border-blue-700 rounded text-sm"
              >
                <div class="flex items-start space-x-2">
                  <LightBulbIcon class="w-4 h-4 text-blue-400 flex-shrink-0 mt-0.5" />
                  <span class="text-blue-200">{{ recommendation.message }}</span>
                </div>
                <button
                  v-if="recommendation.action"
                  class="ml-2 px-2 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700 transition-colors flex-shrink-0"
                  @click="applyRecommendation(recommendation)"
                >
                  {{ recommendation.actionLabel }}
                </button>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Last Build Info -->
        <div v-if="contextStore.lastContextGeneration" class="pt-3 border-t border-gray-700">
          <div class="flex items-center justify-between text-sm">
            <div class="flex items-center space-x-2 text-gray-400">
              <CheckCircleIcon class="w-4 h-4" />
              <span>Last built {{ formatTimestamp(contextStore.lastContextGeneration) }}</span>
            </div>
            <div class="text-xs text-gray-500">
              Status: {{ contextStore.buildStatus }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useFileTreeStore } from '@/stores/file-tree.store'

// Icons
import {
  DocumentIcon,
  CpuChipIcon,
  ClockIcon,
  CurrencyDollarIcon,
  ExclamationTriangleIcon,
  LightBulbIcon,
  CheckCircleIcon,
  ArrowPathIcon,
  ShareIcon
} from '@heroicons/vue/24/outline'

import IconButton from '@/presentation/components/shared/IconButton.vue'

// Stores
const contextStore = useContextBuilderStore()
const fileTreeStore = useFileTreeStore()

// Computed Properties
const formattedTokenCount = computed(() => {
  const count = contextStore.contextMetrics.tokenCount
  if (count >= 1000) {
    return `${(count / 1000).toFixed(1)}k`
  }
  return count.toString()
})

const formattedBuildTime = computed(() => {
  const time = contextStore.contextMetrics.buildTime
  if (time < 1000) {
    return `${time}ms`
  }
  return `${(time / 1000).toFixed(1)}s`
})

const formattedCost = computed(() => {
  return contextStore.contextMetrics.estimatedCost.toFixed(4)
})

const percentageOfTotal = computed(() => {
  const total = fileTreeStore.totalFiles
  const selected = contextStore.contextMetrics.fileCount
  return total > 0 ? Math.round((selected / total) * 100) : 0
})

const tokenUsagePercentage = computed(() => {
  const used = contextStore.contextMetrics.tokenCount
  const limit = contextStore.maxTokenLimit
  return Math.round((used / limit) * 100)
})

const tokenUsageColor = computed(() => {
  const percentage = tokenUsagePercentage.value
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-yellow-500'
  return 'bg-green-500'
})

const healthIndicatorColor = computed(() => {
  const health = contextStore.contextHealth
  if (health >= 80) return 'bg-green-400'
  if (health >= 60) return 'bg-yellow-400'
  return 'bg-red-400'
})

const healthStatus = computed(() => {
  const health = contextStore.contextHealth
  if (health >= 80) return 'Excellent'
  if (health >= 60) return 'Good'
  if (health >= 40) return 'Fair'
  return 'Poor'
})

const buildPerformanceText = computed(() => {
  const time = contextStore.contextMetrics.buildTime
  if (time < 1000) return 'Very fast'
  if (time < 3000) return 'Fast'
  if (time < 10000) return 'Moderate'
  return 'Slow'
})

const fileTypeDistribution = computed(() => {
  // Mock data - in real implementation, analyze actual files
  const types = [
    { extension: 'ts', count: 12, color: '#3b82f6' },
    { extension: 'vue', count: 8, color: '#10b981' },
    { extension: 'js', count: 5, color: '#f59e0b' },
    { extension: 'json', count: 3, color: '#f97316' },
    { extension: 'md', count: 2, color: '#6b7280' }
  ]
  
  const total = types.reduce((sum, type) => sum + type.count, 0)
  
  return types.map(type => ({
    ...type,
    percentage: Math.round((type.count / total) * 100)
  }))
})

const diversityScore = computed(() => {
  // Calculate based on file type variety
  return Math.min(100, fileTypeDistribution.value.length * 20)
})

const completenessScore = computed(() => {
  // Calculate based on whether all necessary files are included
  const hasTests = contextStore.selectedFiles.some(f => f.includes('.test.') || f.includes('.spec.'))
  const hasConfig = contextStore.selectedFiles.some(f => f.includes('config') || f.includes('.json'))
  const hasComponents = contextStore.selectedFiles.some(f => f.includes('component') || f.includes('.vue'))
  
  let score = 60 // Base score
  if (hasTests) score += 15
  if (hasConfig) score += 10
  if (hasComponents) score += 15
  
  return Math.min(100, score)
})

const warnings = computed(() => {
  const warns = []
  
  if (contextStore.contextMetrics.tokenCount > contextStore.maxTokenLimit * 0.9) {
    warns.push({
      id: 'token-limit',
      message: 'Token count is approaching the limit. Consider removing some files.'
    })
  }
  
  if (contextStore.contextMetrics.fileCount > 50) {
    warns.push({
      id: 'file-count',
      message: 'Large number of files selected. This may impact processing speed.'
    })
  }
  
  if (contextStore.validationErrors.length > 0) {
    warns.push({
      id: 'validation',
      message: `${contextStore.validationErrors.length} validation errors detected.`
    })
  }
  
  return warns
})

const recommendations = computed(() => {
  const recs = []
  
  if (contextStore.contextMetrics.fileCount < 3) {
    recs.push({
      id: 'add-files',
      message: 'Consider adding more files for better context.',
      action: 'suggest-files',
      actionLabel: 'Suggest Files'
    })
  }
  
  if (!contextStore.selectedFiles.some(f => f.includes('.test.') || f.includes('.spec.'))) {
    recs.push({
      id: 'add-tests',
      message: 'Including test files can improve code generation quality.',
      action: 'add-test-files',
      actionLabel: 'Add Tests'
    })
  }
  
  if (contextStore.contextMetrics.tokenCount < contextStore.maxTokenLimit * 0.3) {
    recs.push({
      id: 'more-context',
      message: 'You can safely add more files to provide richer context.',
      action: 'optimize-context',
      actionLabel: 'Optimize'
    })
  }
  
  return recs
})

// Methods
const getScoreColor = (score: number) => {
  if (score >= 80) return 'text-green-400'
  if (score >= 60) return 'text-yellow-400'
  return 'text-red-400'
}

const formatTimestamp = (date: Date) => {
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

const refreshMetrics = () => {
  // Trigger metrics recalculation
  contextStore.updateContextMetrics({})
}

const exportSummary = () => {
  // Export summary as JSON or PDF
  const summary = {
    timestamp: new Date().toISOString(),
    metrics: contextStore.contextMetrics,
    fileCount: contextStore.selectedFiles.length,
    health: contextStore.contextHealth,
    warnings: warnings.value,
    recommendations: recommendations.value
  }
  
  const blob = new Blob([JSON.stringify(summary, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'context-summary.json'
  a.click()
  URL.revokeObjectURL(url)
}

const applyRecommendation = (recommendation: any) => {
  switch (recommendation.action) {
    case 'suggest-files':
      contextStore.generateSmartSuggestions()
      break
    case 'add-test-files':
      // Add test files to selection
      console.log('Add test files')
      break
    case 'optimize-context':
      // Optimize context selection
      console.log('Optimize context')
      break
  }
}
</script>

<style scoped>
/* Add any specific styles here */
</style>