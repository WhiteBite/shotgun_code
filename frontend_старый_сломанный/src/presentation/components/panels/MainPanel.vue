<template>
  <div class="flex-1 flex flex-col p-4 gap-4 overflow-y-auto min-w-0">
    <TaskComposer />
    <StreamingContextViewer :context-id="currentContext?.id" />
  </div>
</template>

<script setup lang="ts">
import TaskComposer from "@/presentation/components/workspace/TaskComposer.vue";
import StreamingContextViewer from "@/presentation/components/shared/StreamingContextViewer.vue";
import { useMainPanelService } from '@/composables/useMainPanelService';

const { 
  mainPanelService,
  contextBuilderStore,
  // Store refs
  currentContext
} = useMainPanelService();
</script>