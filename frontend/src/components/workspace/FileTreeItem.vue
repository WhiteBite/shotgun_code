<!-- frontend/src/components/workspace/FileTreeItem.vue -->
<template>
  <Tooltip :text="tooltipText" position="top-start">
    <div
      class="relative group flex items-center gap-1.5 h-[28px] pr-2 rounded"
      :class="{
        'bg-blue-600/30': isActive,
        'opacity-60': node.isIgnored,
        'hover:bg-blue-500/20': !node.isIgnored && !isActive,
        'cursor-pointer': !node.isIgnored,
        'cursor-not-allowed': node.isIgnored,
      }"
      :style="{ paddingLeft: `${node.depth * 20}px` }"
      @click.stop="handleClick"
      @contextmenu.prevent.stop="openContextMenu"
      @mouseenter="handleMouseEnter($event)"
      @mouseleave="handleMouseLeave"
    >
      <div
        class="w-4 h-4 flex-shrink-0 flex items-center justify-center z-10"
        @click.stop="handleClickExpansion"
      >
        <svg
          v-if="node.isDir"
          xmlns="http://www.w3.org/2000/svg"
          width="12"
          height="12"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.5"
          class="text-gray-400 hover:text-blue-400 transition-all"
          :class="{ 'rotate-90 text-blue-400': isExpanded }"
        >
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </div>

      <input
        type="checkbox"
        ref="cbRef"
        :checked="selectionState === 'on'"
        :indeterminate="selectionState === 'partial'"
        :disabled="node.isIgnored"
        @click.stop
        @change.stop="treeActions.toggleNodeSelection(node.path)"
        class="form-checkbox w-4 h-4 z-10 disabled:opacity-50 disabled:cursor-not-allowed hover:border-blue-400 transition-colors"
      />

      <div
        class="w-5 h-5 flex-shrink-0 flex items-center justify-center z-10 -ml-1 text-base"
      >
        <span v-if="node.isDir && isExpanded" class="text-blue-400">📂</span>
        <span v-else-if="node.isDir && !isExpanded" class="text-gray-400"
          >📁</span
        >
        <span v-else>{{ getFileIcon(node.name) }}</span>
      </div>

      <div class="flex-grow flex items-center justify-between overflow-hidden">
        <span
          class="text-sm truncate select-none transition-colors"
          :class="getTextColor()"
          >{{ node.name }}</span
        >
        <div
          class="flex-shrink-0 ml-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
        >
          <button
            v-if="!node.isDir && !node.isIgnored"
            @click.stop="showPinnedQuickLook($event)"
            class="p-1 rounded hover:bg-blue-500/30"
            title="Quick Look (Space)"
          >
            👁️
          </button>
        </div>
      </div>
    </div>
  </Tooltip>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { FileNode } from "@/types/dto";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useContextStore } from "@/stores/context.store";
import { useTreeActions } from "@/composables/useTreeActions";
import { useUiStore } from "@/stores/ui.store";
import { useKeyboardState } from "@/composables/useKeyboardState";
import { useProjectStore } from "@/stores/project.store";
import Tooltip from "@/components/shared/Tooltip.vue";
import { getFileIcon } from "@/utils/fileIcons";

const props = defineProps<{ node: FileNode }>();

const treeStateStore = useTreeStateStore();
const contextStore = useContextStore();
const treeActions = useTreeActions();
const uiStore = useUiStore();
const { isCtrlPressed } = useKeyboardState();
const projectStore = useProjectStore();

const isActive = computed(
  () => treeStateStore.activeNodePath === props.node.path,
);
const isExpanded = computed(() =>
  treeStateStore.expandedPaths.has(props.node.path),
);

const tooltipText = computed(() => {
  if (props.node.isIgnored) return `${props.node.name} (ignored)`;
  return props.node.relPath;
});

function computeSelection(node: FileNode): "on" | "off" | "partial" {
  if (!node.isDir)
    return treeStateStore.selectedPaths.has(node.path) ? "on" : "off";
  const children = (node.children || [])
    .map((c) => contextStore.nodesMap.get(c.path))
    .filter(Boolean) as FileNode[];
  if (children.length === 0)
    return treeStateStore.selectedPaths.has(node.path) ? "on" : "off";
  let hasOn = false,
    hasOff = false;
  for (const child of children) {
    if (child.isIgnored) continue;
    const st = computeSelection(child);
    if (st === "partial") return "partial";
    if (st === "on") hasOn = true;
    if (st === "off") hasOff = true;
    if (hasOn && hasOff) return "partial";
  }
  if (hasOn && !hasOff) return "on";
  if (!hasOn && hasOff) return "off";
  return "off";
}
const selectionState = computed(() => computeSelection(props.node));

const cbRef = ref<HTMLInputElement | null>(null);
watch(
  selectionState,
  (val) => {
    if (cbRef.value) cbRef.value.indeterminate = val === "partial";
  },
  { immediate: true },
);

function getTextColor() {
  if (props.node.isIgnored) return "text-gray-500";
  if (isActive.value) return "text-white font-medium";
  if (selectionState.value !== "off") return "text-blue-200";
  return "text-gray-300";
}

function handleClick(e: MouseEvent) {
  if (props.node.isIgnored) return;
  treeActions.setActiveNode(props.node.path);
  if (props.node.isDir)
    treeActions.toggleNodeExpansion(props.node.path, e.altKey);
}

function handleClickExpansion(e: MouseEvent) {
  if (props.node.isDir)
    treeActions.toggleNodeExpansion(props.node.path, e.altKey);
}

function handleMouseEnter(event: MouseEvent) {
  const rootDir = projectStore.currentProject?.path || "";
  if (!rootDir) return;
  if (isCtrlPressed.value && !props.node.isDir && !props.node.isIgnored) {
    uiStore.showQuickLook({
      rootDir,
      path: props.node.relPath,
      type: "fs",
      event,
      isPinned: false,
    });
  }
}

function handleMouseLeave() {
  if (!uiStore.quickLook.isPinned) {
    uiStore.hideQuickLook();
  }
}

function showPinnedQuickLook(event: MouseEvent) {
  const rootDir = projectStore.currentProject?.path || "";
  if (!rootDir || props.node.isDir || props.node.isIgnored) return;
  uiStore.showQuickLook({
    rootDir,
    path: props.node.relPath,
    type: "fs",
    event,
    isPinned: true,
  });
}

function openContextMenu(event: MouseEvent) {
  if (props.node.isIgnored) return;
  uiStore.openContextMenu(event.clientX, event.clientY, props.node.path);
}
</script>
