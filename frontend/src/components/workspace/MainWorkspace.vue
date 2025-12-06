<template>
  <div class="h-full flex flex-col relative bg-gradient-to-br from-gray-900 via-gray-850 to-gray-900">
    <!-- Background decoration -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -right-40 w-96 h-96 bg-indigo-500/5 rounded-full blur-3xl"></div>
      <div class="absolute -bottom-40 -left-40 w-96 h-96 bg-purple-500/5 rounded-full blur-3xl"></div>
    </div>

    <!-- Main Workspace: 3-column layout with resizable panels -->
    <div class="flex-1 flex overflow-hidden relative z-10">
      <!-- Left Sidebar: File Explorer (Resizable) -->
      <div 
        ref="leftPanelRef"
        :style="{ width: `${leftWidth}px` }"
        class="flex-shrink-0 border-r border-gray-700/30 bg-gray-900/50 backdrop-blur-sm"
      >
        <LeftSidebar @preview-file="handlePreviewFile" />
      </div>

      <!-- Left Resize Handle -->
      <div 
        class="workspace-separator"
        @mousedown="(e) => leftResize.onMouseDown(e)"
        :class="{ 'is-resizing': leftResize.isResizing.value }"
      ></div>

      <!-- Center: Task + Context (Flexible) -->
      <div class="flex-1 min-w-0 bg-gray-900/30">
        <CenterWorkspace />
      </div>

      <!-- Right Sidebar Toggle Button -->
      <button
        @click="toggleRightSidebar"
        class="absolute right-0 top-1/2 -translate-y-1/2 z-20 p-1.5 bg-gray-800/80 border border-gray-700/50 rounded-l-lg text-gray-400 hover:text-white hover:bg-gray-700/80 transition-all"
        :class="{ 'right-sidebar-open': showRightSidebar }"
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
        class="workspace-separator"
        @mousedown="(e) => rightResize.onMouseDown(e)"
        :class="{ 'is-resizing': rightResize.isResizing.value }"
      ></div>

      <!-- Right Sidebar: Results/Output (Resizable) -->
      <Transition name="slide-right">
        <div 
          v-if="showRightSidebar"
          ref="rightPanelRef"
          :style="{ width: `${rightWidth}px` }"
          class="flex-shrink-0 border-l border-gray-700/30 bg-gray-900/50 backdrop-blur-sm"
        >
          <RightSidebar @open-export="handleOpenExport" />
        </div>
      </Transition>
    </div>

    <!-- Action Bar (Footer) -->
    <ActionBar
      class="relative z-10 bg-gray-900/80 backdrop-blur-md border-t border-gray-700/30" 
      @build-context="handleBuildContext"
      @open-export="handleOpenExport"
      @generate-solution="handleGenerateSolution"
    />

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
    return saved !== 'false' // Default to true
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

// Left panel resizing
const leftResize = useResizablePanel({
  minWidth: 280,
  maxWidth: 700,
  defaultWidth: 380,
  storageKey: 'workspace-left-width'
})
const leftPanelRef = leftResize.panelRef
const leftWidth = leftResize.width

// Right panel resizing
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
  // TODO: Implement file preview in right sidebar or modal
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

// Register and cleanup global keyboard shortcuts
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

  if (contextStore.isBuilding) {
    return
  }

  try {
    const filePaths = Array.from(fileStore.selectedPaths)
    
    // Use settings from settings store
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

function handleGenerateSolution() {
  uiStore.addToast('Генерация решения - скоро будет доступна', 'info')
  // TODO: Implement AI solution generation
}
</script>

<style scoped>
.workspace-separator {
  width: 4px;
  background: transparent;
  cursor: col-resize;
  position: relative;
  transition: background 0.2s ease;
}

/* Серая полоска по умолчанию - едва заметная */
.workspace-separator::before {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 2px;
  height: 60px;
  background: rgba(107, 114, 128, 0.2);
  border-radius: 1px;
  transition: all 0.2s ease;
}

/* При hover - становится голубой и более заметной */
.workspace-separator:hover {
  background: linear-gradient(90deg, 
    rgba(59, 130, 246, 0.05) 0%, 
    rgba(59, 130, 246, 0.1) 50%,
    rgba(59, 130, 246, 0.05) 100%);
}

.workspace-separator:hover::before {
  background: rgba(59, 130, 246, 0.6);
  height: 80px;
  box-shadow: 0 0 8px rgba(59, 130, 246, 0.3);
}

/* При активном ресайзе - более яркая */
.workspace-separator:active {
  background: linear-gradient(90deg, 
    rgba(59, 130, 246, 0.1) 0%, 
    rgba(59, 130, 246, 0.2) 50%,
    rgba(59, 130, 246, 0.1) 100%);
}

.workspace-separator:active::before {
  background: rgb(59, 130, 246);
  height: 100px;
  box-shadow: 0 0 12px rgba(59, 130, 246, 0.5),
              0 0 24px rgba(59, 130, 246, 0.3);
}

/* Right sidebar slide animation */
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

/* Toggle button positioning */
.right-sidebar-open {
  border-radius: 0.5rem 0 0 0.5rem;
}
</style>
