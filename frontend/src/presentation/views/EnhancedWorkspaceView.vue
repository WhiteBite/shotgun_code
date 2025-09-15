<template>
  <main class="enhanced-workspace" :class="workspaceClasses">
    <!-- Command Bar -->
    <CommandBar class="workspace-command-bar" />
    
    <!-- Main Content Area -->
    <div class="workspace-content">
      <!-- File Panel -->
      <FilePanelModern
        :is-collapsed="uiStore.panels.file.collapsed"
        @toggle="handlePanelToggle('file', $event)"
        @resize="handlePanelResize('file', $event)"
        class="workspace-file-panel"
      />
      
      <!-- Context Panel -->
      <ContextPanelModern
        :is-collapsed="uiStore.panels.context.collapsed"
        @toggle="handlePanelToggle('context', $event)"
        @resize="handlePanelResize('context', $event)"
        class="workspace-context-panel"
      />
      
      <!-- Main Content Area -->
      <div class="workspace-main-content">
        <!-- Manual Mode -->
        <div v-if="uiStore.workspaceMode === 'manual'" class="manual-mode-content">
          <ResponsiveLayoutGrid
            :layout="mainContentLayout"
            :gap="16"
            class="main-grid"
          >
            <template #task-composer>
              <BasePanel
                title="Task Composer"
                :icon="EditIcon"
                :collapsible="true"
                :scrollable="true"
                variant="secondary"
                size="lg"
                class="task-composer-panel"
              >
                <TaskComposer />
              </BasePanel>
            </template>
            
            <template #context-preview>
              <BasePanel
                title="Context Preview"
                :icon="EyeIcon"
                :collapsible="true"
                :scrollable="false"
                variant="elevated"
                size="lg"
                class="context-preview-panel"
              >
                <ContextPreview />
              </BasePanel>
            </template>
          </ResponsiveLayoutGrid>
        </div>

        <!-- Autonomous Mode -->
        <div v-else-if="uiStore.workspaceMode === 'autonomous'" class="autonomous-mode-content">
          <BasePanel
            title="Autonomous Control"
            :icon="CpuIcon"
            :collapsible="true"
            :scrollable="true"
            variant="elevated"
            size="xl"
            class="autonomous-panel"
          >
            <AutonomousControlPanel />
          </BasePanel>
        </div>
        
        <!-- Reports Mode -->
        <div v-else-if="uiStore.workspaceMode === 'reports'" class="reports-mode-content">
          <BasePanel
            title="Reports & Analytics"
            :icon="BarChartIcon"
            :collapsible="true"
            :scrollable="true"
            variant="primary"
            size="xl"
            class="reports-panel"
          >
            <ReportsPanel />
          </BasePanel>
        </div>
      </div>
    </div>

    <!-- Action Bar -->
    <ActionBar class="workspace-action-bar" />
    
    <!-- Console (if visible) -->
    <Transition name="console-slide">
      <BottomConsole 
        v-if="uiStore.isConsoleVisible"
        class="workspace-console"
      />
    </Transition>

    <!-- Context Menu -->
    <ContextMenu />

    <!-- Drawers -->
    <Transition name="drawer-slide-right">
      <IgnoreDrawer v-if="uiStore.drawers.ignore" />
    </Transition>
    
    <Transition name="drawer-slide-right">
      <PromptsDrawer v-if="uiStore.drawers.prompts" />
    </Transition>
    
    <Transition name="drawer-slide-right">
      <SettingsDrawer 
        v-if="uiStore.drawers.settings"
        @close="uiStore.closeDrawer('settings')"
      />
    </Transition>

    <!-- Modals -->
    <Transition name="modal-fade">
      <ExportModal v-if="uiStore.modals.export" />
    </Transition>
    
    <Transition name="modal-fade">
      <CommitHistoryModal v-if="uiStore.modals.commitHistory" />
    </Transition>
    
    <Transition name="modal-fade">
      <ReportViewerModal v-if="uiStore.modals.reportViewer" />
    </Transition>

    <!-- Toast Notifications -->
    <ToastNotifications class="workspace-notifications" />
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { 
  EditIcon, 
  EyeIcon, 
  CpuIcon, 
  BarChartIcon 
} from 'lucide-vue-next'

// Enhanced Components
import BasePanel from '@/presentation/components/BasePanel.vue'
import FilePanelModern from '@/presentation/components/panels/FilePanelModern.vue'
import ContextPanelModern from '@/presentation/components/panels/ContextPanelModern.vue'
import ResponsiveLayoutGrid from '@/presentation/components/ResponsiveLayoutGrid.vue'

// Legacy Components (to be gradually replaced)
import CommandBar from '@/presentation/components/bars/CommandBar.vue'
import ActionBar from '@/presentation/components/bars/ActionBar.vue'
import BottomConsole from '@/presentation/components/workspace/BottomConsole.vue'
import ContextMenu from '@/presentation/components/shared/ContextMenu.vue'
import TaskComposer from '@/presentation/components/workspace/TaskComposer.vue'
import ContextPreview from '@/presentation/components/workspace/ContextPreview.vue'
import AutonomousControlPanel from '@/presentation/components/panels/AutonomousControlPanel.vue'
import ReportsPanel from '@/presentation/components/panels/ReportsPanel.vue'

// Drawers
import IgnoreDrawer from '@/presentation/components/drawers/IgnoreDrawer.vue'
import PromptsDrawer from '@/presentation/components/drawers/PromptsDrawer.vue'
import SettingsDrawer from '@/presentation/components/drawers/SettingsDrawer.vue'

// Modals
import ExportModal from '@/presentation/components/modals/ExportModal.vue'
import CommitHistoryModal from '@/presentation/components/modals/CommitHistoryModal.vue'
import ReportViewerModal from '@/presentation/components/modals/ReportViewerModal.vue'

// Notifications
import ToastNotifications from '@/presentation/components/shared/ToastNotifications.vue'

// Stores and Composables
import { useUiStore } from '@/stores/ui.store'
import { useProjectStore } from '@/stores/project.store'
import { attachShortcuts, detachShortcuts } from '@/composables/useKeyboardShortcuts'
import { setupProjectSession } from '@/composables/useProjectSession'

// Stores
const uiStore = useUiStore()
const projectStore = useProjectStore()

// Computed
const workspaceClasses = computed(() => [
  `workspace-mode-${uiStore.workspaceMode}`,
  {
    'workspace-console-visible': uiStore.isConsoleVisible,
    'workspace-panels-collapsed': allPanelsCollapsed.value,
    'workspace-mobile': isMobile.value
  }
])

const allPanelsCollapsed = computed(() => {
  return Object.values(uiStore.panels).every(panel => panel.collapsed)
})

const isMobile = computed(() => {
  return window.innerWidth < 768
})

const mainContentLayout = computed(() => {
  if (isMobile.value) {
    return {
      gridTemplateRows: 'auto auto',
      gridTemplateColumns: '1fr',
      gridTemplateAreas: `
        "task-composer"
        "context-preview"
      `
    }
  }
  
  return {
    gridTemplateRows: '400px 1fr',
    gridTemplateColumns: '1fr',
    gridTemplateAreas: `
      "task-composer"
      "context-preview"
    `
  }
})

// Methods
function handlePanelToggle(panelName: string, collapsed: boolean) {
  uiStore.setPanelCollapsed(panelName, collapsed)
}

function handlePanelResize(panelName: string, width: number) {
  uiStore.setPanelWidth(panelName, width)
}

function handleKeyboardShortcut(shortcut: string) {
  switch (shortcut) {
    case 'toggle-file-panel':
      uiStore.togglePanel('file')
      break
    case 'toggle-context-panel':
      uiStore.togglePanel('context')
      break
    case 'toggle-console':
      uiStore.toggleConsole()
      break
    case 'switch-to-manual':
      uiStore.setWorkspaceMode('manual')
      break
    case 'switch-to-autonomous':
      uiStore.setWorkspaceMode('autonomous')
      break
    case 'switch-to-reports':
      uiStore.setWorkspaceMode('reports')
      break
  }
}

// Lifecycle
onMounted(() => {
  attachShortcuts()
  setupProjectSession()
  
  // Listen for custom keyboard shortcuts
  window.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  detachShortcuts()
  window.removeEventListener('keydown', handleGlobalKeydown)
})

function handleGlobalKeydown(event: KeyboardEvent) {
  // Handle global shortcuts
  if (event.ctrlKey || event.metaKey) {
    switch (event.key) {
      case '1':
        event.preventDefault()
        uiStore.setWorkspaceMode('manual')
        break
      case '2':
        event.preventDefault()
        uiStore.setWorkspaceMode('autonomous')
        break
      case '3':
        event.preventDefault()
        uiStore.setWorkspaceMode('reports')
        break
      case 'b':
        event.preventDefault()
        uiStore.togglePanel('file')
        break
      case 'k':
        event.preventDefault()
        uiStore.togglePanel('context')
        break
      case '`':
        event.preventDefault()
        uiStore.toggleConsole()
        break
    }
  }
}

// Watch for workspace mode changes to update layout
watch(
  () => uiStore.workspaceMode,
  (newMode) => {
    // Apply mode-specific optimizations
    if (newMode === 'autonomous') {
      // Collapse context panel in autonomous mode for more space
      uiStore.setPanelCollapsed('context', true)
    } else if (newMode === 'manual') {
      // Ensure context panel is expanded in manual mode
      uiStore.setPanelCollapsed('context', false)
    }
  }
)
</script>

<style scoped>
.enhanced-workspace {
  /* Main layout */
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  
  /* Enhanced background with glass morphism */
  background: 
    radial-gradient(circle at 20% 50%, rgba(168, 85, 247, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 80% 50%, rgba(59, 130, 246, 0.1) 0%, transparent 50%),
    linear-gradient(180deg, rgb(2, 6, 23) 0%, rgb(15, 23, 42) 100%);
  
  /* Typography */
  color: rgb(248, 250, 252);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  
  /* Overflow management */
  overflow: hidden;
}

.workspace-command-bar {
  flex-shrink: 0;
  z-index: 100;
}

.workspace-content {
  display: flex;
  flex: 1;
  min-height: 0;
  gap: 12px;
  padding: 12px;
}

.workspace-file-panel,
.workspace-context-panel {
  flex-shrink: 0;
  transition: all var(--transition-normal);
}

.workspace-main-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.manual-mode-content,
.autonomous-mode-content,
.reports-mode-content {
  flex: 1;
  min-height: 0;
}

.main-grid {
  height: 100%;
}

.task-composer-panel {
  grid-area: task-composer;
}

.context-preview-panel {
  grid-area: context-preview;
}

.autonomous-panel,
.reports-panel {
  height: 100%;
}

.workspace-action-bar {
  flex-shrink: 0;
  z-index: 100;
}

.workspace-console {
  flex-shrink: 0;
  z-index: 50;
}

.workspace-notifications {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 1000;
  pointer-events: none;
}

/* Responsive Design */
@media (max-width: 768px) {
  .workspace-content {
    flex-direction: column;
    gap: 8px;
    padding: 8px;
  }
  
  .workspace-file-panel,
  .workspace-context-panel {
    order: -1;
  }
  
  .workspace-main-content {
    order: 0;
  }
}

@media (max-width: 640px) {
  .enhanced-workspace {
    font-size: 0.875rem;
  }
  
  .workspace-content {
    padding: 4px;
    gap: 4px;
  }
}

/* Mode-specific styles */
.workspace-mode-manual {
  --workspace-accent: rgb(59, 130, 246);
}

.workspace-mode-autonomous {
  --workspace-accent: rgb(168, 85, 247);
}

.workspace-mode-reports {
  --workspace-accent: rgb(34, 197, 94);
}

/* State-based styles */
.workspace-console-visible {
  .workspace-content {
    padding-bottom: 8px;
  }
}

.workspace-panels-collapsed {
  .workspace-main-content {
    margin-left: 0;
    margin-right: 0;
  }
}

/* Transitions */
.console-slide-enter-active,
.console-slide-leave-active {
  transition: transform var(--transition-normal), opacity var(--transition-normal);
}

.console-slide-enter-from,
.console-slide-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.drawer-slide-right-enter-active,
.drawer-slide-right-leave-active {
  transition: transform var(--transition-normal), opacity var(--transition-normal);
}

.drawer-slide-right-enter-from,
.drawer-slide-right-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity var(--transition-normal);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

/* Focus management */
.enhanced-workspace:focus-within {
  outline: none;
}

/* Print styles */
@media print {
  .enhanced-workspace {
    background: white;
    color: black;
  }
  
  .workspace-command-bar,
  .workspace-action-bar,
  .workspace-console,
  .workspace-notifications {
    display: none;
  }
}

/* High contrast mode */
@media (prefers-contrast: high) {
  .enhanced-workspace {
    background: black;
    color: white;
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>