<template>
  <div
      :style="{ paddingLeft: node.depth * 10 + 8 + 'px' }"
      :class="{
      'bg-gray-700/50': treeStateStore.activeNodePath === node.path,
      'text-gray-400': node.isIgnored,
      'hover:bg-gray-800': !node.isIgnored,
      'font-semibold': node.isDir,
      'text-sm flex items-center h-[24px] cursor-pointer select-none': true,
    }"
      @click.stop="handleRowClick"
      @contextmenu.prevent="handleContextMenu"
      @mouseenter="handleMouseEnter"
      @mouseleave="handleMouseLeave"
  >
    <div class="flex items-center gap-1.5 w-full">
      <button v-if="node.isDir" @click.stop="toggleDir($event)" class="flex-shrink-0 w-4 h-4 text-gray-500 hover:text-white" aria-label="Toggle directory">
        <svg v-if="expanded" xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"></polyline></svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"></polyline></svg>
      </button>
      <div v-else class="w-4 h-4 flex-shrink-0"></div>
      <span class="w-4 h-4 flex-shrink-0">{{ icon }}</span>
      <span class="flex-grow min-w-0 truncate" :title="titleText">{{ node.name }}</span>
      <input
          ref="cbRef"
          type="checkbox"
          :checked="selectionState === 'on'"
          :disabled="node.isIgnored"
          class="form-checkbox h-3 w-3"
          :aria-checked="selectionState === 'partial' ? 'mixed' : (selectionState === 'on' ? 'true' : 'false')"
          @click.stop
          @change="onToggleSelection"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { FileNode } from "@/types/dto";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useUiStore } from "@/stores/ui.store";
import { useProjectStore } from "@/stores/project.store";
import { useQuickLook } from "@/composables/useQuickLook";
import { useTriStateSelection, type Sel } from "@/composables/useTriStateSelection";
import { getFileIcon } from "@/utils/fileIcons";

const props = defineProps<{ node: FileNode }>();
const treeStateStore = useTreeStateStore();
const fileTreeStore = useFileTreeStore();
const uiStore = useUiStore();
const projectStore = useProjectStore();
const { handleMouseEnter: hoverEnter, handleMouseLeave: hoverLeave } = useQuickLook();

const expanded = computed(() => treeStateStore.expandedPaths.has(props.node.path));
const icon = computed(() => getFileIcon(props.node.name));
const cbRef = ref<HTMLInputElement | null>(null);

const { computeSelection, toggleCascade } = useTriStateSelection(fileTreeStore.nodesMap, treeStateStore.selectedPaths);
const selectionState = computed<Sel>(() => computeSelection(props.node));

const titleText = computed(() => {
  if (props.node.isDir) return props.node.relPath;
  return `${props.node.relPath} • ${props.node.size.toLocaleString()} B`;
});

watch(selectionState, (val) => { if (cbRef.value) cbRef.value.indeterminate = (val === "partial"); }, { immediate: true });

function toggleDir(ev: MouseEvent) {
  if (ev.altKey) {
    treeStateStore.toggleExpansionRecursive(props.node.path, fileTreeStore.nodesMap, !expanded.value);
  } else {
    treeStateStore.toggleExpansion(props.node.path);
  }
}
function handleRowClick() {
  treeStateStore.activeNodePath = props.node.path;
  if (!props.node.isDir) treeStateStore.toggleSelection(props.node.path);
  else treeStateStore.toggleExpansion(props.node.path);
}
function handleContextMenu(event: MouseEvent) {
  uiStore.openContextMenu(event.clientX, event.clientY, props.node.path);
}
function onToggleSelection() { toggleCascade(props.node); }
function handleMouseEnter(ev: MouseEvent) {
  const rootDir = projectStore.currentProject?.path || ""; if (!rootDir) return;
  hoverEnter(ev, props.node, rootDir);
}
function handleMouseLeave() { hoverLeave(); }
</script>