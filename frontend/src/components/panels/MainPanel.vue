<template>
  <div class="flex-1 flex flex-col p-4 gap-4 overflow-y-auto min-w-0">
    <TaskComposer />
    <FilePreview />
  </div>
</template>

<script setup lang="ts">
import TaskComposer from '@/components/workspace/TaskComposer.vue';
import FilePreview from '@/components/workspace/FilePreview.vue';
</script>