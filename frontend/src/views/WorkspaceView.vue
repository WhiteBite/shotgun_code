<template>
  <main class="h-screen w-screen bg-gray-900 text-gray-200 flex flex-col">
    <CommandBar />
    <div class="flex-grow flex min-h-0">
      <FilePanel />
      <div class="flex-grow flex flex-col min-h-0 min-w-0">
        <SplitPaneVertical :initialTop="360" :storageKey="splitPaneKey" class="flex-grow">
          <template #top>
            <div class="pl-0 p-4 pb-2 h-full">
              <TaskComposer />
            </div>
          </template>
          <template #bottom>
            <div class="pl-0 p-4 pt-2 h-full">
              <FilePreview />
            </div>
          </template>
        </SplitPaneVertical>
      </div>
    </div>
    <ActionBar />
    <BottomConsole v-if="uiStore.isConsoleVisible"/>
    <ToastNotifications />
    <ContextMenu />
    <QuickLook />
    <!-- Drawers -->
    <IgnoreDrawer />
    <PromptsDrawer />
    <SettingsDrawer />
    <!-- Modals -->
    <ExportModal />
    <CommitHistoryModal />
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from "vue";
import { useUiStore } from "@/stores/ui.store";
import { attachShortcuts, detachShortcuts } from "@/composables/useKeyboardShortcuts";
import { setupProjectSession } from "@/composables/useProjectSession";
import { useProjectStore } from "@/stores/project.store";
import { hashStringFNV1a } from "@/utils/hash";

import FilePanel from "@/components/panels/FilePanel.vue";
import FilePreview from "@/components/workspace/FilePreview.vue";
import TaskComposer from "@/components/workspace/TaskComposer.vue";
import CommandBar from "@/components/bars/CommandBar.vue";
import ActionBar from "@/components/bars/ActionBar.vue";
import BottomConsole from "@/components/workspace/BottomConsole.vue";
import ToastNotifications from "@/components/shared/ToastNotifications.vue";
import ContextMenu from "@/components/shared/ContextMenu.vue";
import QuickLook from "@/components/shared/QuickLook.vue";
import ExportModal from "@/components/modals/ExportModal.vue";
import CommitHistoryModal from "@/components/modals/CommitHistoryModal.vue";
import SplitPaneVertical from "@/components/shared/SplitPaneVertical.vue";

import IgnoreDrawer from "@/components/drawers/IgnoreDrawer.vue";
import PromptsDrawer from "@/components/drawers/PromptsDrawer.vue";
import SettingsDrawer from "@/components/drawers/SettingsDrawer.vue";

const uiStore = useUiStore();
const projectStore = useProjectStore();

const splitPaneKey = computed(() => {
  const p = projectStore.currentProject?.path ?? "no-project";
  return "split:workspace:main:" + hashStringFNV1a(p) + ":v1";
});

onMounted(() => {
  attachShortcuts();
  setupProjectSession();
});
onUnmounted(() => {
  detachShortcuts();
});
</script>