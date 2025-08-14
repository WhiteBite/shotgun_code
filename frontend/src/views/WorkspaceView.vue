<template>
  <div class="flex flex-col h-full overflow-hidden">
    <div class="flex flex-1 min-h-0">
      <FilePanel />
      <div class="flex-1 flex flex-col p-4 gap-4 overflow-y-auto min-w-0">
        <TaskComposer />
        <FilePreview />
      </div>
    </div>
    <ActionsPanel />
    <QuickLook />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { useProjectStore } from "@/stores/project.store";
import { useContextStore } from "@/stores/context.store";
import { useRouter } from "vue-router";
import { useProjectSession } from "@/composables/useProjectSession";
import FilePanel from "@/components/panels/FilePanel.vue";
import TaskComposer from "@/components/workspace/TaskComposer.vue";
import FilePreview from "@/components/workspace/FilePreview.vue";
import ActionsPanel from "@/components/panels/ActionsPanel.vue";
import QuickLook from "@/components/shared/QuickLook.vue";
const projectStore = useProjectStore();
const contextStore = useContextStore();
const router = useRouter();
const { loadSession } = useProjectSession();

onMounted(async () => {
  if (!projectStore.isProjectLoaded || !projectStore.currentProject) {
    router.push("/");
    return;
  }
  await contextStore.fetchFileTree();
  loadSession();
});
</script>
