<template>
  <div class="h-full flex flex-col">
    <!-- Action Bar -->
    <ActionBar />

    <!-- Main Workspace: 3-column layout with visual separators -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Left Sidebar: File Explorer -->
      <div class="w-80 flex-shrink-0 border-r border-gray-700/50">
        <LeftSidebar @preview-file="handlePreviewFile" />
      </div>

      <!-- Separator -->
      <div class="workspace-separator"></div>

      <!-- Center: Task + Context -->
      <div class="flex-1">
        <CenterWorkspace />
      </div>

      <!-- Separator -->
      <div class="workspace-separator"></div>

      <!-- Right Sidebar: Results/Output -->
      <div class="w-96 flex-shrink-0 border-l border-gray-700/50">
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
