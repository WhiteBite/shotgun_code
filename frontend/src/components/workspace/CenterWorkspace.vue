<template>
  <div class="h-full flex flex-col">
    <!-- Task Panel (Top) -->
    <div class="h-64 border-b border-gray-700 flex-shrink-0">
      <TaskPanel />
    </div>

    <!-- Context Panel (Bottom) -->
    <div class="flex-1 overflow-hidden">
      <ContextPanel />
    </div>
  </div>
</template>

<script setup lang="ts">
import TaskPanel from '@/components/TaskPanel.vue'
import ContextPanel from '@/components/ContextPanel.vue'
</script>
