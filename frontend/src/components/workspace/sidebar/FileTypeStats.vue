<template>
  <div class="space-y-4">
    <!-- Summary Stats -->
    <div class="context-stats">
      <p class="text-xs font-semibold text-gray-400 mb-3">{{ t('stats.totalFiles') }}</p>
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-card-value">{{ fileCount }}</div>
          <div class="stat-label">{{ t('tools.files') }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-card-value">{{ formatNumber(lineCount) }}</div>
          <div class="stat-label">{{ t('context.lines') }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-card-value stat-card-value-indigo">{{ formatNumber(tokenCount) }}</div>
          <div class="stat-label">{{ t('action.tokens') }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-card-value stat-card-value-emerald">${{ estimatedCost.toFixed(4) }}</div>
          <div class="stat-label">{{ t('action.cost') }}</div>
        </div>
      </div>
    </div>

    <!-- By File Type -->
    <div class="context-stats" v-if="fileTypeStats.length > 0">
      <p class="text-xs font-semibold text-gray-400 mb-3">{{ t('stats.byType') }}</p>
      <div class="space-y-2">
        <div 
          v-for="stat in fileTypeStats" 
          :key="stat.extension"
          class="flex items-center gap-2"
        >
          <span class="text-lg flex-shrink-0">{{ stat.icon }}</span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between text-xs mb-1">
              <span class="text-gray-300 truncate">.{{ stat.extension }}</span>
              <span class="text-gray-500">{{ stat.count }} {{ t('context.files') }}</span>
            </div>
            <div class="h-1.5 bg-gray-700/50 rounded-full overflow-hidden">
              <div 
                class="h-full rounded-full transition-all duration-500"
                :class="stat.colorClass"
                :style="{ width: `${stat.percentage}%` }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- By Folder -->
    <div class="context-stats" v-if="folderStats.length > 0">
      <p class="text-xs font-semibold text-gray-400 mb-3">{{ t('stats.byFolder') }}</p>
      <div class="space-y-2">
        <div 
          v-for="stat in folderStats.slice(0, 5)" 
          :key="stat.folder"
          class="flex items-center gap-2"
        >
          <svg class="w-4 h-4 text-blue-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between text-xs mb-1">
              <span class="text-gray-300 truncate">{{ stat.folder || '/' }}</span>
              <span class="text-gray-500">{{ stat.count }}</span>
            </div>
            <div class="h-1.5 bg-gray-700/50 rounded-full overflow-hidden">
              <div 
                class="h-full bg-blue-500/60 rounded-full transition-all duration-500"
                :style="{ width: `${stat.percentage}%` }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!hasContext" class="empty-state py-8">
      <div class="empty-state-icon !w-12 !h-12 mb-3">
        <svg class="!w-6 !h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      </div>
      <p class="empty-state-text">{{ t('context.notBuilt') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useContextStore } from '@/features/context'
import { useFileStore } from '@/features/files'
import { computed } from 'vue'

const { t } = useI18n()
const contextStore = useContextStore()
const fileStore = useFileStore()

const hasContext = computed(() => contextStore.hasContext)
const fileCount = computed(() => contextStore.fileCount)
const lineCount = computed(() => contextStore.lineCount)
const tokenCount = computed(() => contextStore.tokenCount)
const estimatedCost = computed(() => contextStore.estimatedCost)

// File type icons and colors
const fileTypeConfig: Record<string, { icon: string; colorClass: string }> = {
  ts: { icon: '📘', colorClass: 'bg-blue-500' },
  tsx: { icon: '⚛️', colorClass: 'bg-cyan-500' },
  js: { icon: '📒', colorClass: 'bg-yellow-500' },
  jsx: { icon: '⚛️', colorClass: 'bg-cyan-500' },
  vue: { icon: '💚', colorClass: 'bg-emerald-500' },
  go: { icon: '🐹', colorClass: 'bg-sky-500' },
  py: { icon: '🐍', colorClass: 'bg-yellow-600' },
  css: { icon: '🎨', colorClass: 'bg-pink-500' },
  scss: { icon: '🎨', colorClass: 'bg-pink-600' },
  html: { icon: '🌐', colorClass: 'bg-orange-500' },
  json: { icon: '📋', colorClass: 'bg-gray-500' },
  md: { icon: '📝', colorClass: 'bg-gray-400' },
  yaml: { icon: '⚙️', colorClass: 'bg-red-400' },
  yml: { icon: '⚙️', colorClass: 'bg-red-400' },
  sql: { icon: '🗃️', colorClass: 'bg-indigo-500' },
  default: { icon: '📄', colorClass: 'bg-gray-500' }
}

const fileTypeStats = computed(() => {
  const selected = Array.from(fileStore.selectedPaths)
  if (selected.length === 0) return []

  const counts = new Map<string, number>()
  
  for (const path of selected) {
    const ext = path.split('.').pop()?.toLowerCase() || 'other'
    counts.set(ext, (counts.get(ext) || 0) + 1)
  }

  const total = selected.length
  return Array.from(counts.entries())
    .map(([extension, count]) => {
      const config = fileTypeConfig[extension] || fileTypeConfig.default
      return {
        extension,
        count,
        percentage: Math.round((count / total) * 100),
        icon: config.icon,
        colorClass: config.colorClass
      }
    })
    .sort((a, b) => b.count - a.count)
    .slice(0, 8)
})

const folderStats = computed(() => {
  const selected = Array.from(fileStore.selectedPaths)
  if (selected.length === 0) return []

  const counts = new Map<string, number>()
  
  for (const path of selected) {
    const parts = path.split('/')
    const folder = parts.length > 1 ? parts[0] : '/'
    counts.set(folder, (counts.get(folder) || 0) + 1)
  }

  const total = selected.length
  return Array.from(counts.entries())
    .map(([folder, count]) => ({
      folder,
      count,
      percentage: Math.round((count / total) * 100)
    }))
    .sort((a, b) => b.count - a.count)
})

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
  return num.toString()
}
</script>
