<template>
  <div class="h-full flex flex-col">
    <!-- Action Bar -->
    <ActionBar />

    <!-- Main Workspace: 3-column layout with resizable panels -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Left Sidebar: File Explorer (Resizable) -->
      <div 
        ref="leftPanelRef"
        :style="{ width: `${leftWidth}px` }"
        class="flex-shrink-0 border-r border-gray-700/50"
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
      <div class="flex-1 min-w-0">
        <CenterWorkspace />
      </div>

      <!-- Right Resize Handle -->
      <div 
        class="workspace-separator"
        @mousedown="(e) => rightResize.onMouseDown(e)"
        :class="{ 'is-resizing': rightResize.isResizing.value }"
      ></div>

      <!-- Right Sidebar: Results/Output (Resizable) -->
      <div 
        ref="rightPanelRef"
        :style="{ width: `${rightWidth}px` }"
        class="flex-shrink-0 border-l border-gray-700/50"
      >
        <RightSidebar />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import ActionBar from './ActionBar.vue'
import LeftSidebar from './LeftSidebar.vue'
import CenterWorkspace from './CenterWorkspace.vue'
import RightSidebar from './RightSidebar.vue'
import { useResizablePanel } from '@/composables/useResizablePanel'

// Left panel resizing
const leftResize = useResizablePanel({
  minWidth: 200,
  maxWidth: 600,
  defaultWidth: 320,
  storageKey: 'workspace-left-width'
})
const leftPanelRef = leftResize.panelRef
const leftWidth = leftResize.width

// Right panel resizing
const rightResize = useResizablePanel({
  minWidth: 250,
  maxWidth: 700,
  defaultWidth: 384,
  storageKey: 'workspace-right-width'
})
const rightPanelRef = rightResize.panelRef
const rightWidth = rightResize.width

function handlePreviewFile(filePath: string) {
  console.log('Preview file:', filePath)
  // TODO: Implement file preview in right sidebar or modal
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
</style>
