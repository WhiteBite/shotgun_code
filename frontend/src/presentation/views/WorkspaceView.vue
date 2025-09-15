<template>
  <main 
    class="workspace-main h-screen flex flex-col bg-gray-900 text-white"
    @keydown.prevent="handleKeydown"
    tabindex="-1"
  >
    <!-- Header Bar -->
    <div v-if="store.layout.headerBarVisible" class="header-bar flex-shrink-0 p-3 bg-gray-800 border-b border-gray-700">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <h1 class="text-lg font-semibold text-white">Shotgun Code</h1>
          <div class="flex items-center gap-2 text-sm text-gray-400">
            <span class="capitalize">{{ store.currentMode }}</span>
            <span>•</span>
            <span>{{ store.activePanels.length }} panels</span>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="toggleMode"
            :class="[
              'px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
              store.currentMode === 'autonomous' 
                ? 'bg-blue-600 hover:bg-blue-500 text-white' 
                : 'bg-gray-600 hover:bg-gray-500 text-white'
            ]"
          >
            {{ store.currentMode === 'manual' ? 'Switch to Autonomous' : 'Switch to Manual' }}
          </button>
          <button
            @click="toggleConsole"
            :class="[
              'p-2 rounded-lg transition-colors duration-200',
              uiStore.isConsoleVisible 
                ? 'text-blue-400 bg-gray-700' 
                : 'text-gray-400 hover:text-white hover:bg-gray-700'
            ]"
            title="Toggle Console"
          >
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
          </button>
          <button
            @click="openSettings"
            class="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors duration-200"
            title="Settings"
          >
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
              />
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>
    
    <!-- Main Workspace Area -->
    <div class="workspace-body flex-1 flex overflow-hidden relative">
      <!-- Left Sidebar - File Explorer -->
      <div 
        v-if="store.layout.panelVisibility.contextArea"
        class="file-explorer-panel flex-shrink-0 border-r border-gray-700 bg-gray-800 flex flex-col"
        :style="{ width: store.layout.contextPanelWidth + 'px' }"
      >
        <!-- File Explorer Header -->
        <div class="panel-header p-4 border-b border-gray-700 bg-gradient-to-r from-gray-800 to-gray-750">
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-sm font-bold text-white flex items-center gap-2">
              <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5a2 2 0 012-2h4a2 2 0 012 2v6H8V5z" />
              </svg>
              Project Files
            </h3>
            <div class="flex items-center gap-2">
              <div class="px-2 py-1 bg-blue-600 text-white text-xs rounded-full font-medium">
                {{ selectedFilesCount }}
              </div>
            </div>
          </div>
          <div class="text-xs text-gray-400">
            Select files to build context
          </div>
        </div>
        
        <!-- File Tree -->
        <div class="flex-1 overflow-hidden bg-gray-850 relative">
          <!-- Loading State -->
          <div v-if="fileTreeStore.isLoading" class="absolute inset-0 flex items-center justify-center bg-gray-800/50 backdrop-blur-sm">
            <div class="text-center">
              <svg class="animate-spin h-8 w-8 text-blue-500 mx-auto mb-2" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              <p class="text-sm text-gray-400">Loading files...</p>
            </div>
          </div>
          
          <!-- Error State -->
          <div v-else-if="fileTreeStore.error" class="p-4 text-center">
            <div class="text-red-400 mb-2">
              <svg class="w-8 h-8 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
              <p class="text-sm">Error loading files</p>
              <p class="text-xs text-gray-500 mt-1">{{ fileTreeStore.error }}</p>
            </div>
            <button 
              @click="initializeProject" 
              class="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white text-xs rounded transition-colors"
            >
              Retry
            </button>
          </div>
          
          <!-- Empty State -->
          <div v-else-if="!fileTreeStore.hasFiles" class="p-4 text-center">
            <div class="text-gray-400 mb-4">
              <svg class="w-12 h-12 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5a2 2 0 012-2h4a2 2 0 012 2v6H8V5z" />
              </svg>
              <p class="text-sm font-medium mb-1">No project loaded</p>
              <p class="text-xs text-gray-500 mb-4">Open a project to browse files</p>
            </div>
            <button 
              @click="initializeProject" 
              class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg transition-colors flex items-center gap-2 mx-auto"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              Open Project
            </button>
          </div>
          
          <!-- File Tree Content -->
          <div v-else class="h-full overflow-auto">
            <FileTree />
          </div>
        </div>
        
        <!-- Context Builder Controls -->
        <div class="panel-footer p-4 border-t border-gray-700 bg-gradient-to-r from-gray-800 to-gray-750">
          <div class="mb-3">
            <div class="flex items-center justify-between mb-2">
              <div class="text-sm font-medium text-gray-200">Context Builder</div>
              <div class="text-xs text-gray-400">{{ selectedFilesCount }} files selected</div>
            </div>
            <div class="w-full bg-gray-700 rounded-full h-1.5">
              <div 
                class="bg-gradient-to-r from-blue-500 to-purple-500 h-1.5 rounded-full transition-all duration-300"
                :style="{ width: Math.min((selectedFilesCount / 10) * 100, 100) + '%' }"
              ></div>
            </div>
          </div>
          <div class="flex gap-2 mb-3">
            <SplitSettingsPopover
              v-model="contextActions.splitSettings.value"
              :preview="contextActions.splitPreview.value"
              :disabled="selectedFilesCount === 0"
              @apply="applySplitSettings"
              @refresh="refreshSplitPreview"
            />
          </div>
          <button
            @click="buildContext"
            :disabled="selectedFilesCount === 0 || contextBuilderStore.isBuilding"
            class="w-full px-4 py-3 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-500 hover:to-blue-500 disabled:from-gray-600 disabled:to-gray-600 disabled:cursor-not-allowed text-white text-sm font-semibold rounded-lg transition-all duration-200 flex items-center justify-center gap-2 shadow-lg hover:shadow-xl transform hover:scale-105 disabled:transform-none"
          >
            <svg v-if="contextBuilderStore.isBuilding" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
            </svg>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            {{ contextBuilderStore.isBuilding ? 'Building Context...' : 'Build Context' }}
            <kbd v-if="!contextBuilderStore.isBuilding" class="ml-auto px-1.5 py-0.5 text-xs bg-white/20 rounded">Ctrl+Enter</kbd>
          </button>
        </div>
      </div>
      
      <!-- Resize handle for left panel -->
      <div 
        v-if="store.layout.panelVisibility.contextArea"
        class="resize-handle"
        :style="{ left: store.layout.contextPanelWidth + 'px' }"
        @mousedown="startResize('left', $event)"
      ></div>
      
      <!-- Center Panel - Main Content -->
      <div class="central-area flex-1 flex flex-col overflow-hidden bg-gray-900">
        <!-- Mode-specific main panel -->
        <div class="flex-1 overflow-hidden">
          <ManualModePanel v-if="store.currentMode === 'manual'" />
          <AutonomousControlPanel v-else-if="store.currentMode === 'autonomous'" />
          
          <!-- Fallback content -->
          <div v-else class="h-full flex items-center justify-center">
            <div class="text-center text-gray-400">
              <div class="mb-4">
                <svg class="w-16 h-16 mx-auto text-gray-500 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
                <h3 class="text-lg font-semibold mb-2">Workspace Ready</h3>
                <p class="text-sm">Select a workspace mode to get started</p>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Bottom Console (if visible) -->
        <div v-if="store.layout.panelVisibility.console" class="border-t border-gray-700 bg-gray-800">
          <BottomConsole />
        </div>
      </div>
      
      <!-- Resize handle for right panel -->
      <div 
        v-if="store.layout.panelVisibility.resultsArea"
        class="resize-handle"
        :style="{ right: store.layout.resultsPanelWidth + 'px' }"
        @mousedown="startResize('right', $event)"
      ></div>
      
      <!-- Right Panel - Results Area -->
      <div 
        v-if="store.layout.panelVisibility.resultsArea"
        class="results-area flex-shrink-0 border-l border-gray-700 bg-gray-800"
        :style="{ width: store.layout.resultsPanelWidth + 'px' }"
      >
        <ResultsArea />
      </div>
    </div>
    
    <!-- Status Bar -->
    <div v-if="store.layout.statusBarVisible" class="status-bar flex-shrink-0 px-3 py-1.5 bg-gray-900 border-t border-gray-700">
      <div class="flex justify-between text-xs text-gray-500">
        <div class="flex items-center gap-4">
          <span class="flex items-center gap-1">
            <div class="w-2 h-2 bg-green-500 rounded-full"></div>
            Ready
          </span>
          <span class="capitalize">{{ store.currentMode }} mode</span>
        </div>
        <div class="flex items-center gap-4">
          <span>{{ store.activePanels.length }} panels active</span>
          <span>{{ store.availableWorkspaceArea.width }}×{{ store.availableWorkspaceArea.height }}</span>
        </div>
      </div>
    </div>
    
    <!-- Settings Drawer -->
    <Transition name="drawer-slide-right">
      <SettingsDrawer 
        v-if="uiStore.drawers.settings"
        :isOpen="uiStore.drawers.settings"
        @close="uiStore.closeDrawer('settings')"
      />
    </Transition>
    
    <!-- QuickLook -->
    <QuickLook />
    
    <!-- Toast Notifications -->
    <ToastNotifications />
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, defineExpose, ref, watch } from "vue";
import { useWorkspaceStore } from "@/stores/workspace.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useProjectStore } from "@/stores/project.store";
import { useUiStore } from "@/stores/ui.store";
import FileTree from "@/presentation/components/workspace/FileTree.vue";
import ManualModePanel from "@/presentation/components/panels/ManualModePanel.vue";
import AutonomousControlPanel from "@/presentation/components/panels/AutonomousControlPanel.vue";
import ResultsArea from "@/presentation/components/panels/ResultsArea.vue";
import BottomConsole from "@/presentation/components/workspace/BottomConsole.vue";
import SettingsDrawer from "@/presentation/components/drawers/SettingsDrawer.vue";
import QuickLook from "@/presentation/components/shared/QuickLook.vue";
import ToastNotifications from "@/presentation/components/shared/ToastNotifications.vue";
import SplitSettingsPopover from "@/presentation/components/workspace/ContextActions/SplitSettingsPopover.vue";
import { useContextActions } from "@/composables/useContextActions";
import { apiService } from '@/services/api.service';

// Stores
const store = useWorkspaceStore();
const fileTreeStore = useFileTreeStore();
const contextBuilderStore = useContextBuilderStore();
const projectStore = useProjectStore();
const uiStore = useUiStore();

// Context actions composable
const contextActions = useContextActions();

// Computed
const selectedFilesCount = computed(() => fileTreeStore.selectedFiles.length);

// Watch for project changes and load file tree accordingly
watch(
  () => projectStore.currentProject,
  async (newProject, oldProject) => {
    console.log('Project changed:', { newProject, oldProject });
    if (newProject && newProject.path !== (oldProject?.path || '')) {
      try {
        console.log('Loading file tree for project:', newProject.path);
        await fileTreeStore.loadProject(newProject.path);
        console.log('File tree loaded for project');
      } catch (error) {
        console.error('Failed to load file tree for project:', error);
      }
    } else if (!newProject) {
      // Clear file tree when no project is loaded
      console.log('Clearing file tree');
      fileTreeStore.clearProject();
    }
  },
  { immediate: true }
);

// Initialize the workspace store
onMounted(async () => {
  store.initializeStore();
  
  // Initialize project and file tree if needed
  if (!projectStore.currentProject && !fileTreeStore.hasFiles) {
    try {
      // Try to load a default project or prompt user
      await initializeProject();
    } catch (error) {
      console.warn('No project loaded:', error);
    }
  }
});

// Initialize project
const initializeProject = async () => {
  try {
    // Try to get the current directory from the backend
    const currentDir = await apiService.getCurrentDirectory();
    
    if (currentDir) {
      // Use the actual current directory
      await projectStore.loadProject(currentDir);
      console.log('Project initialized with current directory:', currentDir);
    } else {
      // Fallback to using the frontend directory
      const fallbackPath = '..'; // Use parent directory (frontend's parent)
      await projectStore.loadProject(fallbackPath);
      console.log('Project initialized with fallback path:', fallbackPath);
    }
  } catch (error) {
    console.error('Failed to initialize project:', error);
  }
};

// Methods
const toggleMode = async () => {
  await store.toggleMode();
};

const toggleConsole = () => {
  uiStore.toggleConsole();
};

const buildContext = () => {
  if (projectStore.currentProject?.path) {
    contextBuilderStore.buildContextFromSelection(projectStore.currentProject.path);
  }
};

const openSettings = () => {
  uiStore.openDrawer('settings');
};

const applySplitSettings = () => {
  // Apply split settings through context actions
  // The settings are already reactive through the composable
  console.log('Split settings applied:', contextActions.splitSettings.value);
};

const refreshSplitPreview = async () => {
  // Refresh the split preview through context actions
  contextActions.refreshPreview();
};

// Resize handling
const isResizing = ref(false);
const resizeType = ref<'left' | 'right' | null>(null);
const resizeStartX = ref(0);
const initialWidth = ref(0);

const startResize = (type: 'left' | 'right', e: MouseEvent) => {
  e.preventDefault();
  isResizing.value = true;
  resizeType.value = type;
  resizeStartX.value = e.clientX;
  
  if (type === 'left') {
    initialWidth.value = store.layout.contextPanelWidth;
  } else if (type === 'right') {
    initialWidth.value = store.layout.resultsPanelWidth;
  }
  
  document.body.style.cursor = 'col-resize';
  document.body.classList.add('select-none'); // Prevent text selection during resize
  
  const handleMouseMove = (e: MouseEvent) => {
    if (!isResizing.value) return;
    
    const container = document.querySelector('.workspace-body');
    if (!container) return;
    
    const containerRect = container.getBoundingClientRect();
    const delta = e.clientX - resizeStartX.value;
    const minWidth = 200;
    const maxWidth = containerRect.width * 0.5;
    
    if (resizeType.value === 'left') {
      // Calculate new width for left panel
      const newWidth = Math.min(
        maxWidth, 
        Math.max(minWidth, initialWidth.value + delta)
      );
      
      // Use the new setLayout method instead of updateLayout
      store.setLayout({
        contextPanelWidth: newWidth
      });
    } else if (resizeType.value === 'right') {
      // Calculate new width for right panel
      const newWidth = Math.min(
        maxWidth, 
        Math.max(minWidth, initialWidth.value - delta)
      );
      
      // Use the new setLayout method instead of updateLayout
      store.setLayout({
        resultsPanelWidth: newWidth
      });
    }
  };
  
  const handleMouseUp = () => {
    isResizing.value = false;
    resizeType.value = null;
    document.body.style.cursor = '';
    document.body.classList.remove('select-none');
    document.removeEventListener('mousemove', handleMouseMove);
    document.removeEventListener('mouseup', handleMouseUp);
  };
  
  document.addEventListener('mousemove', handleMouseMove);
  document.addEventListener('mouseup', handleMouseUp);
};

// Keyboard shortcuts
const handleKeydown = (event: KeyboardEvent) => {
  // F11 - Toggle fullscreen
  if (event.key === 'F11') {
    event.preventDefault();
    store.toggleFullscreen();
  }
  
  // F12 - Toggle focus mode
  if (event.key === 'F12') {
    event.preventDefault();
    store.toggleFocusMode();
  }
  
  // Ctrl+Shift+M - Toggle mode
  if (event.ctrlKey && event.shiftKey && event.key === 'M') {
    event.preventDefault();
    toggleMode();
  }
  
  // Ctrl+Enter - Build context
  if (event.ctrlKey && event.key === 'Enter') {
    event.preventDefault();
    if (selectedFilesCount.value > 0) {
      buildContext();
    }
  }
};

// Explicitly expose variables for template
defineExpose({
  store,
  handleKeydown,
  toggleMode,
  buildContext,
  selectedFilesCount
});
</script>

<style scoped>
.workspace-main {
  @apply h-screen flex flex-col bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 text-white;
  height: 100vh; /* Force 100% viewport height */
  overflow: hidden; /* Prevent scrolling on the main container */
}

.header-bar {
  @apply flex-shrink-0 bg-gradient-to-r from-gray-800 via-gray-700 to-gray-800 border-b border-gray-600 shadow-lg;
  min-height: 64px;
  backdrop-filter: blur(10px);
  z-index: 10; /* Ensure header stays on top */
}

.workspace-body {
  @apply flex-1 flex overflow-hidden;
  min-height: 0; /* Allow proper flexbox sizing */
}

.file-explorer-panel {
  @apply bg-gradient-to-b from-gray-800 to-gray-900 border-gray-600 flex-shrink-0 shadow-xl;
  min-width: 200px;
  max-width: 500px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border-right: 1px solid rgba(75, 85, 99, 0.5);
  z-index: 5; /* Higher than central area */
}

.central-area {
  @apply flex-1 flex flex-col overflow-hidden bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900;
  min-width: 300px;
  position: relative;
  z-index: 1;
}

.results-area {
  @apply bg-gradient-to-b from-gray-800 to-gray-900 border-gray-600 flex-shrink-0 shadow-xl;
  min-width: 200px;
  max-width: 500px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border-left: 1px solid rgba(75, 85, 99, 0.5);
  z-index: 5; /* Higher than central area */
}

.status-bar {
  @apply flex-shrink-0 bg-gradient-to-r from-gray-900 via-gray-800 to-gray-900 border-t border-gray-600;
  min-height: 36px;
  backdrop-filter: blur(8px);
  z-index: 10; /* Ensure status bar stays on top */
}

/* Resize handle */
.resize-handle {
  transition: background-color 0.2s ease;
  position: absolute;
  top: 0;
  bottom: 0;
  width: 8px; /* Increased width for easier grabbing */
  background-color: rgba(75, 85, 99, 0.4);
  z-index: 20; /* Above panels */
  cursor: col-resize;
  transform: translateX(-50%); /* Center the handle on the border */
}

.resize-handle:hover, 
.resize-handle:active {
  background-color: rgba(59, 130, 246, 0.6); /* Brighter blue when active/hover */
}

/* When actively resizing */
body.select-none .resize-handle {
  background-color: rgba(59, 130, 246, 0.8); /* Even brighter when actually resizing */
}

/* Hide resize handles and adjust for small screens */
@media (max-width: 768px) {
  .resize-handle {
    display: none;
  }
  
  .workspace-body {
    flex-direction: column;
  }
  
  .file-explorer-panel,
  .results-area {
    width: 100% !important;
    max-height: 250px;
    min-width: auto;
    border-width: 0;
    border-bottom-width: 1px;
  }
  
  .central-area {
    order: -1;
  }
}

/* Panel interactions */
.file-explorer-panel:hover,
.results-area:hover {
  @apply shadow-2xl;
  transform: translateY(-1px);
}

/* Improved button styles */
button {
  @apply transition-all duration-300 ease-in-out;
  transform-origin: center;
}

button:hover {
  @apply transform scale-105 shadow-lg;
  filter: brightness(1.1);
}

button:active {
  @apply transform scale-95;
  filter: brightness(0.9);
}

/* Enhanced focus styles */
button:focus {
  @apply outline-none ring-2 ring-blue-500 ring-opacity-50;
}

/* Modern workspace modes */
.workspace-compact .header-bar {
  min-height: 48px;
}

.workspace-compact .status-bar {
  min-height: 28px;
}

.workspace-fullscreen {
  @apply fixed inset-0 z-50;
}

.workspace-focused .header-bar,
.workspace-focused .status-bar {
  @apply hidden;
}

.workspace-transitioning {
  @apply pointer-events-none;
}

/* Professional scrollbars */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  @apply bg-gray-800;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  @apply bg-gray-600;
  border-radius: 4px;
  transition: background-color 0.2s ease;
}

::-webkit-scrollbar-thumb:hover {
  @apply bg-gray-500;
}

/* Drawer transition animations */
.drawer-slide-right-enter-active,
.drawer-slide-right-leave-active {
  transition: all 0.4s cubic-bezier(0.25, 0.8, 0.25, 1);
}

.drawer-slide-right-enter-from,
.drawer-slide-right-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

/* Loading states */
.loading {
  position: relative;
  overflow: hidden;
}

.loading::after {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(59, 130, 246, 0.8), transparent);
  animation: loading-shimmer 2s infinite;
}

@keyframes loading-shimmer {
  0% { left: -100%; }
  100% { left: 100%; }
}

/* Responsive design */
@media (max-width: 1200px) {
  .central-area {
    min-width: 400px;
  }
  
  .file-explorer-panel {
    min-width: 260px;
  }
  
  .results-area {
    min-width: 280px;
  }
}

@media (max-width: 1024px) {
  .workspace-body {
    @apply flex-col;
  }
  
  .file-explorer-panel,
  .results-area {
    @apply relative border-0 border-b border-gray-700;
    width: 100% !important;
    max-height: 250px;
    min-width: auto;
  }
  
  .central-area {
    @apply order-first;
    min-width: auto;
  }
}

@media (max-width: 768px) {
  .header-bar {
    @apply p-2;
  }
  
  .header-bar h1 {
    @apply text-base;
  }
  
  .file-explorer-panel,
  .results-area {
    max-height: 200px;
  }
  
  .status-bar {
    @apply px-2 py-1;
  }
}

/* High contrast mode */
@media (prefers-contrast: high) {
  .workspace-main {
    background-color: black;
    color: white;
  }
  
  .file-explorer-panel,
  .results-area,
  .header-bar,
  .status-bar {
    border-width: 2px;
    border-color: white;
  }
}
</style>

/* Drawer transitions */
.drawer-slide-right-enter-active,
.drawer-slide-right-leave-active {
  transition: transform 0.3s ease-in-out;
}

.drawer-slide-right-enter-from,
.drawer-slide-right-leave-to {
  transform: translateX(100%);
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

/* Focus states */
.workspace-main:focus {
  @apply outline-none;
}

/* Improved scroll bars */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  @apply bg-gray-800;
}

::-webkit-scrollbar-thumb {
  @apply bg-gray-600 rounded-full;
}

::-webkit-scrollbar-thumb:hover {
  @apply bg-gray-500;
}

/* Resize handle */
.resize-handle {
  transition: background-color 0.2s ease;
}

.resize-handle:hover {
  background-color: #3b82f6;
}

/* Improved responsive design */
@media (max-width: 1200px) {
  .central-area {
    min-width: 400px;
  }
  
  .file-explorer-panel {
    min-width: 260px;
  }
  
  .results-area {
    min-width: 280px;
  }
}

@media (max-width: 1024px) {
  .workspace-body {
    flex-direction: column;
  }
  
  .file-explorer-panel,
  .results-area {
    border-width: 0;
    border-bottom-width: 1px;
    width: 100% !important;
    max-height: 250px;
    min-width: auto;
  }
  
  .central-area {
    order: -1;
    min-width: auto;
  }
  
  /* Hide resize handles on mobile */
  .resize-handle {
    display: none;
  }
}

@media (max-width: 768px) {
  .header-bar {
    padding: 0.5rem;
  }
  
  .header-bar h1 {
    font-size: 1.125rem;
  }
  
  .file-explorer-panel,
  .results-area {
    max-height: 200px;
  }
  
  .status-bar {
    padding: 0.25rem 0.5rem;
  }
  
  .panel-header,
  .panel-footer {
    padding: 0.75rem;
  }
}

/* High contrast mode */
@media (prefers-contrast: high) {
  .workspace-main {
    background-color: black;
    color: white;
  }
  
  .file-explorer-panel,
  .results-area,
  .header-bar,
  .status-bar {
    border-width: 2px;
    border-color: white;
  }
}