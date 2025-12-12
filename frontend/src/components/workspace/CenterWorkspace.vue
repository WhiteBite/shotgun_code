<template>
  <div class="center-container">
    <!-- Task Panel (conditionally rendered) -->
    <div v-if="showTaskPanel" class="task-panel">
      <TaskPanel />
    </div>
    
    <!-- Context Panel -->
    <div :class="showTaskPanel ? 'context-panel-with-task' : 'context-panel-full'" data-tour="context-preview">
      <ContextPanel />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ContextPanel } from '@/features/context'
import { TaskPanel } from '@/features/task'
import { ref, watch } from 'vue'

const showTaskPanel = ref(false)

// Persist task panel visibility
watch(showTaskPanel, (visible) => {
  try {
    localStorage.setItem('task-panel-visible', visible.toString())
  } catch (err) {
    console.warn('Failed to save task panel visibility:', err)
  }
})

// Restore task panel visibility
try {
  const savedVisibility = localStorage.getItem('task-panel-visible')
  if (savedVisibility) {
    showTaskPanel.value = savedVisibility === 'true'
  } else {
    showTaskPanel.value = false
  }
} catch (err) {
  console.warn('Failed to load task panel visibility:', err)
}
</script>

<style scoped>
.center-container {
  @apply h-full flex flex-col;
}

.task-panel {
  @apply h-64 flex-shrink-0;
  border-bottom: 1px solid var(--border-subtle);
}

.context-panel-with-task {
  @apply flex-1;
}

.context-panel-full {
  @apply h-full;
}
</style>
