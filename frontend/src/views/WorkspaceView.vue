<template>
  <div class="flex flex-col h-full overflow-hidden">
    <div class="flex flex-1 min-h-0">
      <FilePanel />
      <MainPanel />
    </div>
    <ActionsPanel />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useProjectStore } from '@/stores/project.store';
import { useContextStore } from '@/stores/context.store';
import { useRouter } from 'vue-router';
import FilePanel from '@/components/panels/FilePanel.vue';
import MainPanel from '@/components/panels/MainPanel.vue';
import ActionsPanel from '@/components/panels/ActionsPanel.vue';

const projectStore = useProjectStore();
const contextStore = useContextStore();
const router = useRouter();

onMounted(() => {
  if (!projectStore.isProjectLoaded || !projectStore.currentProject) {
    router.push('/');
    return;
  }
  contextStore.fetchFileTree();
});
</script>