<template>
  <div class="workspace-container">
    <!-- Background decoration -->
    <div class="workspace-bg">
      <div class="workspace-glow workspace-glow-1"></div>
      <div class="workspace-glow workspace-glow-2"></div>
    </div>

    <!-- Main 3-column layout -->
    <div class="workspace-layout">
      <!-- Left Sidebar -->
      <div 
        ref="leftPanelRef"
        :style="{ width: `${leftWidth}px` }"
        class="workspace-panel workspace-panel-left"
      >
        <LeftSidebar @preview-file="handlePreviewFile" @build-context="handleBuildContext" />
      </div>

      <!-- Left Resize Handle -->
      <div 
        class="resize-handle"
        @mousedown="(e) => leftResize.onMouseDown(e)"
        :class="{ 'is-resizing': leftResize.isResizing.value }"
      >
        <div class="resize-handle-line"></div>
      </div>

      <!-- Center Panel -->
      <div class="workspace-center">
        <CenterWorkspace />
      </div>

      <!-- Right Sidebar Toggle -->
      <button
        @click="toggleRightSidebar"
        class="sidebar-toggle"
        :class="{ 'sidebar-toggle-open': showRightSidebar }"
        :style="showRightSidebar ? { right: `${rightWidth}px` } : {}"
        :title="t('sidebar.toggle')"
      >
        <svg class="w-4 h-4 transition-transform" :class="{ 'rotate-180': !showRightSidebar }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>

      <!-- Right Resize Handle -->
      <div 
        v-if="showRightSidebar"
        class="resize-handle"
        @mousedown="(e) => rightResize.onMouseDown(e)"
        :class="{ 'is-resizing': rightResize.isResizing.value }"
      >
        <div class="resize-handle-line"></div>
      </div>

      <!-- Right Sidebar -->
      <Transition name="slide-right">
        <div 
          v-if="showRightSidebar"
          ref="rightPanelRef"
          :style="{ width: `${rightWidth}px` }"
          class="workspace-panel workspace-panel-right"
        >
          <RightSidebar @open-export="handleOpenExport" />
        </div>
      </Transition>
    </div>

    <!-- Action Bar -->
    <ActionBar class="workspace-actionbar" @open-export="handleOpenExport" />

    <!-- Export Modal -->
    <ExportModal ref="exportModalRef" />
  </div>
</template>


<script setup lang="ts">
import ExportModal from '@/components/ExportModal.vue'
import { useI18n } from '@/composables/useI18n'
import { useResizablePanel } from '@/composables/useResizablePanel'
import { useContextStore } from '@/features/context'
import { useFileStore } from '@/features/files'
import { useSettingsStore } from '@/stores/settings.store'
import { useUIStore } from '@/stores/ui.store'
import { onMounted, onUnmounted, ref } from 'vue'
import ActionBar from './ActionBar.vue'
import CenterWorkspace from './CenterWorkspace.vue'
import LeftSidebar from './LeftSidebar.vue'
import RightSidebar from './RightSidebar.vue'

const { t } = useI18n()
const contextStore = useContextStore()
const fileStore = useFileStore()
const uiStore = useUIStore()
const settingsStore = useSettingsStore()
const exportModalRef = ref<InstanceType<typeof ExportModal> | null>(null)

// Right sidebar visibility
const showRightSidebar = ref(loadSidebarState())

function loadSidebarState(): boolean {
  try {
    const saved = localStorage.getItem('right-sidebar-visible')
    return saved !== 'false'
  } catch {
    return true
  }
}

function toggleRightSidebar() {
  showRightSidebar.value = !showRightSidebar.value
  try {
    localStorage.setItem('right-sidebar-visible', String(showRightSidebar.value))
  } catch (e) {
    console.warn('Failed to save sidebar state:', e)
  }
}

// Panel resizing
const leftResize = useResizablePanel({
  minWidth: 280,
  maxWidth: 700,
  defaultWidth: 380,
  storageKey: 'workspace-left-width'
})
const leftPanelRef = leftResize.panelRef
const leftWidth = leftResize.width

const rightResize = useResizablePanel({
  minWidth: 320,
  maxWidth: 800,
  defaultWidth: 420,
  storageKey: 'workspace-right-width'
})
const rightPanelRef = rightResize.panelRef
const rightWidth = rightResize.width

function handlePreviewFile(filePath: string) {
  console.log('Preview file:', filePath)
}

// Global keyboard shortcut handlers
const handleGlobalBuildContext = () => handleBuildContext()
const handleGlobalOpenExport = () => handleOpenExport()
const handleGlobalCopyContext = async () => {
  if (contextStore.hasContext && contextStore.contextId) {
    try {
      const content = await contextStore.getFullContextContent()
      await navigator.clipboard.writeText(content)
      uiStore.addToast('Контекст скопирован в буфер обмена', 'success')
    } catch (error) {
      console.error('Failed to copy context:', error)
      uiStore.addToast('Ошибка при копировании контекста', 'error')
    }
  }
}

onMounted(() => {
  window.addEventListener('global-build-context', handleGlobalBuildContext)
  window.addEventListener('global-open-export', handleGlobalOpenExport)
  window.addEventListener('global-copy-context', handleGlobalCopyContext)
})

onUnmounted(() => {
  window.removeEventListener('global-build-context', handleGlobalBuildContext)
  window.removeEventListener('global-open-export', handleGlobalOpenExport)
  window.removeEventListener('global-copy-context', handleGlobalCopyContext)
})

async function handleBuildContext() {
  if (fileStore.selectedPaths.size === 0) {
    uiStore.addToast('Выберите файлы для построения контекста', 'warning')
    return
  }

  if (contextStore.isBuilding) return

  try {
    const filePaths = Array.from(fileStore.selectedPaths)
    const options = {
      maxTokens: settingsStore.settings.context.maxTokens,
      stripComments: settingsStore.settings.context.stripComments,
      includeTests: settingsStore.settings.context.includeTests,
      splitStrategy: settingsStore.settings.context.splitStrategy
    }
    
    await contextStore.buildContext(filePaths, options)
    uiStore.addToast('Контекст успешно построен', 'success')
  } catch (error) {
    console.error('Failed to build context:', error)
    const errorMsg = error instanceof Error ? error.message : 'Unknown error'
    uiStore.addToast(`Ошибка при построении контекста: ${errorMsg}`, 'error')
  }
}

function handleOpenExport() {
  if (!contextStore.hasContext) {
    uiStore.addToast('Сначала постройте контекст', 'warning')
    return
  }
  exportModalRef.value?.open()
}
</script>


<style scoped>
.workspace-container {
  @apply h-full flex flex-col relative;
  background: var(--bg-app);
}

/* Background */
.workspace-bg {
  @apply absolute inset-0 overflow-hidden pointer-events-none;
}

.workspace-glow {
  @apply absolute rounded-full blur-3xl opacity-30;
}

.workspace-glow-1 {
  @apply w-[600px] h-[600px] -top-64 -right-64;
  background: radial-gradient(circle, rgba(139, 92, 246, 0.08) 0%, transparent 70%);
}

.workspace-glow-2 {
  @apply w-[500px] h-[500px] -bottom-48 -left-48;
  background: radial-gradient(circle, rgba(236, 72, 153, 0.06) 0%, transparent 70%);
}

/* Layout */
.workspace-layout {
  @apply flex-1 flex overflow-hidden relative z-10;
}

.workspace-panel {
  @apply flex-shrink-0 backdrop-blur-sm;
  background: rgba(11, 16, 32, 0.6);
}

.workspace-panel-left {
  border-right: 1px solid var(--border-subtle);
}

.workspace-panel-right {
  border-left: 1px solid var(--border-subtle);
}

.workspace-center {
  @apply flex-1 min-w-0;
  background: rgba(5, 8, 22, 0.4);
}

/* Resize Handle */
.resize-handle {
  @apply w-1 cursor-col-resize relative;
  @apply flex items-center justify-center;
  transition: all 150ms ease-out;
}

.resize-handle:hover,
.resize-handle.is-resizing {
  background: var(--accent-purple-bg);
}

.resize-handle-line {
  @apply w-0.5 h-16 rounded-full;
  background: var(--border-strong);
  transition: all 150ms ease-out;
}

.resize-handle:hover .resize-handle-line,
.resize-handle.is-resizing .resize-handle-line {
  @apply h-24;
  background: var(--accent-purple-border);
  box-shadow: 0 0 8px rgba(139, 92, 246, 0.4);
}

/* Sidebar Toggle */
.sidebar-toggle {
  @apply absolute top-1/2 -translate-y-1/2 z-20;
  @apply p-1.5 rounded-l-lg;
  background: var(--bg-1);
  border: 1px solid var(--border-default);
  border-right: 0;
  color: var(--text-muted);
  transition: all 200ms ease-out;
  right: 0;
}

.sidebar-toggle:hover {
  color: var(--text-primary);
  background: var(--bg-2);
}

.sidebar-toggle-open {
  border-radius: 0.5rem 0 0 0.5rem;
}

/* Action Bar */
.workspace-actionbar {
  @apply relative z-10;
  border-top: 1px solid var(--border-default);
  background: var(--bg-1);
  backdrop-filter: blur(12px);
}

/* Slide Animation */
.slide-right-enter-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide-right-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.slide-right-enter-from,
.slide-right-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
