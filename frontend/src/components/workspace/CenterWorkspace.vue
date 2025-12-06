<template>
  <div class="h-full flex flex-col">
    <!-- Context Building Mode Only - AI Chat removed from here (now in RightSidebar) -->
    <div class="h-full flex flex-col">
      <!-- Task Panel (Top) - conditionally rendered -->
      <div v-if="showTaskPanel" class="h-64 border-b border-gray-700 flex-shrink-0">
        <TaskPanel />
      </div>
      <!-- Context Panel (Bottom) - takes full height when TaskPanel is hidden -->
      <div :class="showTaskPanel ? 'flex-1' : 'h-full'">
        <ContextPanel />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { TaskPanel } from '@/features/task'
import { ContextPanel } from '@/features/context'

const showTaskPanel = ref(false)

// Save task panel visibility to localStorage
watch(showTaskPanel, (visible) => {
  try {
    localStorage.setItem('task-panel-visible', visible.toString())
  } catch (err) {
    console.warn('Failed to save task panel visibility:', err)
  }
})

// Restore task panel visibility from localStorage
try {
  const savedVisibility = localStorage.getItem('task-panel-visible')
  if (savedVisibility) {
    showTaskPanel.value = savedVisibility === 'true'
  } else {
    // Set default to false as per the plan
    showTaskPanel.value = false
  }
} catch (err) {
  console.warn('Failed to load task panel visibility:', err)
}
</script>