<template>
  <div class="flex h-full">
    <ContextBuilderPanel />

    <div class="flex-1 flex flex-col p-4 gap-4 overflow-y-auto min-w-0">
      <TaskComposer />
      <ContextPreview />
    </div>

    <ActionBar />

    <CommitHistoryModal
        :is-visible="uiStore.isCommitModalVisible"
        @close="uiStore.closeCommitModal"
        @apply="handleApplyCommits"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useProjectStore } from '@/stores/projectStore';
import { useContextStore } from '@/stores/contextStore';
import { useGitStore } from '@/stores/gitStore';
import { useUiStore } from '@/stores/uiStore';
import ContextBuilderPanel from '../components/layout/ContextBuilderPanel.vue';
import TaskComposer from '../components/workspace/TaskComposer.vue';
import ContextPreview from '../components/workspace/ContextPreview.vue';
import ActionBar from '../components/global/ActionBar.vue';
import CommitHistoryModal from '../components/modals/CommitHistoryModal.vue';

const router = useRouter();
const projectStore = useProjectStore();
const contextStore = useContextStore();
const gitStore = useGitStore();
const uiStore = useUiStore();

const loadDataForProject = () => {
  if (!projectStore.isProjectLoaded || !projectStore.currentProject) {
    router.push('/');
    return;
  }
  contextStore.clearProjectData(); // Clear old data before loading new
  contextStore.fetchFileTree(projectStore.currentProject.path).then(() => {
    gitStore.checkAvailabilityAndStatus(); // Check git after tree is loaded
  });
};

onMounted(loadDataForProject);
// Watch for project changes (e.g., user selects a different recent project)
watch(() => projectStore.currentProject?.path, (newPath, oldPath) => {
  if (newPath && newPath !== oldPath) {
    loadDataForProject();
  }
});

const handleApplyCommits = (commitHashes: string[]) => {
  gitStore.selectFilesFromCommits(commitHashes);
  uiStore.closeCommitModal();
};
</script>