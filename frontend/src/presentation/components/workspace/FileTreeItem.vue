<template>
  <div
    :style="{ paddingLeft: depth * 10 + 8 + 'px' }"
    :class="{
      'bg-gray-700/50': isActive,
      'text-gray-400': node.isGitignored || node.isCustomIgnored,
      'hover:bg-gray-800': !(node.isGitignored || node.isCustomIgnored),
      'font-semibold': node.isDir,
      'text-sm flex items-center h-[24px] cursor-pointer select-none': true,
    }"
    @click.stop="handleRowClick($event)"
    @contextmenu.prevent="handleContextMenu"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
    @mousemove="handleMouseMove"
  >
    <div class="flex items-center gap-1.5 w-full">
      <button
        v-show="node.isDir"
        class="flex-shrink-0 w-4 h-4 text-gray-500 hover:text-white"
        aria-label="Toggle directory"
        @click.stop="toggleDir($event)"
      >
        <svg
          v-show="isExpanded"
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
        <svg
          v-show="!isExpanded"
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </button>
      <div v-show="!node.isDir" class="w-4 h-4 flex-shrink-0"></div>
      <span class="w-4 h-4 flex-shrink-0">{{ icon }}</span>
      <span class="flex-grow min-w-0 truncate" :title="titleText">{{
        node.name
      }}</span>
      <input
        ref="cbRef"
        type="checkbox"
        :checked="isSelected"
        :disabled="node.isGitignored || node.isCustomIgnored"
        class="form-checkbox h-3 w-3"
        @click.stop
        @change="onToggleSelection($event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { FileNode } from "@/types/dto";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useProjectStore } from "@/stores/project.store";
import { useUiStore } from "@/stores/ui.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useKeyboardState } from "@/composables/useKeyboardState";
import { getFileIcon } from "@/utils/fileIcons";
import { useTriStateSelection } from "@/composables/useTriStateSelection";

// CRITICAL: Add debouncing and processing state to prevent rapid clicks
const isProcessing = ref(false);
const lastClickTime = ref(0);
const CLICK_DEBOUNCE_MS = 300; // Minimum 300ms between clicks

const props = defineProps<{
  node: FileNode;
  depth: number;
}>();

const emit = defineEmits<{
  select: [node: FileNode, event: MouseEvent];
  expand: [node: FileNode];
}>();

const fileTreeStore = useFileTreeStore();
const projectStore = useProjectStore();
const uiStore = useUiStore();
const treeStateStore = useTreeStateStore();
const { isAltPressed, isCtrlPressed } = useKeyboardState();

const cbRef = ref<HTMLInputElement | null>(null);

const icon = computed(() => getFileIcon(props.node.name));
const isSelected = computed(() => {
  // Проверяем состояние выделения в treeStateStore
  return treeStateStore.selectedPaths.has(props.node.path);
});

const isPartiallySelected = computed(() => {
  if (!props.node.isDir) return false;
  // Используем computeSelection из tri-state
  const { computeSelection } = useTriStateSelection(
    fileTreeStore.nodesMap as unknown as Map<string, FileNode>,
    treeStateStore.selectedPaths as Set<string>,
  );
  return computeSelection(props.node) === "partial";
});
const isExpanded = computed(() => {
  // Check if this node is expanded in the tree state
  return treeStateStore.expandedPaths.has(props.node.path);
});
const isActive = computed(
  () => treeStateStore.activeNodePath === props.node.path,
);

const titleText = computed(() => {
  if (props.node.isDir) return props.node.relPath;
  return `${props.node.relPath} • ${props.node.size.toLocaleString()} B`;
});

watch(
  [isSelected, isPartiallySelected],
  ([selected, partiallySelected]) => {
    if (cbRef.value) {
      cbRef.value.indeterminate = partiallySelected;
      cbRef.value.checked = selected;
    }
  },
  { immediate: true },
);

function toggleDir(_ev: MouseEvent) {
  // CRITICAL: Debounce rapid clicks
  const now = Date.now();
  if (isProcessing.value || (now - lastClickTime.value) < CLICK_DEBOUNCE_MS) {
    console.log('Ignoring rapid click on directory toggle');
    return;
  }
  
  isProcessing.value = true;
  lastClickTime.value = now;
  
  try {
    // Если нажат Alt, делаем рекурсивное разворачивание
    if (isAltPressed.value) {
      treeStateStore.toggleExpansionRecursive(
        props.node.path,
        fileTreeStore.nodesMap as unknown as Map<string, FileNode>,
        !isExpanded.value,
      );
    } else {
      treeStateStore.toggleExpansion(props.node.path);
    }
    emit("expand", props.node);
  } catch (error) {
    console.error('Error during directory toggle:', error);
  } finally {
    // Release processing lock after delay
    setTimeout(() => {
      isProcessing.value = false;
    }, CLICK_DEBOUNCE_MS);
  }
}

function handleRowClick(event: MouseEvent) {
  // CRITICAL: Debounce rapid clicks to prevent memory issues
  const now = Date.now();
  if (isProcessing.value || (now - lastClickTime.value) < CLICK_DEBOUNCE_MS) {
    console.log('Ignoring rapid click on row');
    return;
  }
  
  isProcessing.value = true;
  lastClickTime.value = now;
  
  try {
    // Устанавливаем активный узел
    treeStateStore.activeNodePath = props.node.path;

    if (!props.node.isDir) {
      // Для файлов - переключаем выделение
      treeStateStore.toggleNodeSelection(
        props.node.path,
        fileTreeStore.nodesMap as unknown as Map<string, FileNode>,
      );
      emit("select", props.node, event);
    } else {
      // Для директорий - разворачиваем/сворачиваем
      if (isAltPressed.value) {
        // Рекурсивное разворачивание при Alt
        treeStateStore.toggleExpansionRecursive(
          props.node.path,
          fileTreeStore.nodesMap as unknown as Map<string, FileNode>,
          !isExpanded.value,
        );
      } else {
        treeStateStore.toggleExpansion(props.node.path);
      }
      emit("expand", props.node);
    }
  } catch (error) {
    console.error('Error during row click:', error);
  } finally {
    // Release processing lock after delay
    setTimeout(() => {
      isProcessing.value = false;
    }, CLICK_DEBOUNCE_MS);
  }
}

function handleContextMenu(event: MouseEvent) {
  uiStore.openContextMenu(event.clientX, event.clientY, props.node.path);
}

function onToggleSelection(_event: Event) {
  // CRITICAL: Debounce checkbox clicks
  const now = Date.now();
  if (isProcessing.value || (now - lastClickTime.value) < CLICK_DEBOUNCE_MS) {
    console.log('Ignoring rapid checkbox click');
    return;
  }
  
  isProcessing.value = true;
  lastClickTime.value = now;
  
  try {
    // Переключаем выделение напрямую через treeStateStore
    treeStateStore.toggleNodeSelection(
      props.node.path,
      fileTreeStore.nodesMap as unknown as Map<string, FileNode>,
    );
    // НЕ эмитим событие select, так как это может вызвать двойное выделение
  } catch (error) {
    console.error('Error during checkbox toggle:', error);
  } finally {
    // Release processing lock after delay
    setTimeout(() => {
      isProcessing.value = false;
    }, CLICK_DEBOUNCE_MS);
  }
}

function handleMouseEnter(ev: MouseEvent) {
  // Показываем QuickLook только при зажатом Ctrl
  if (
    isCtrlPressed.value &&
    !props.node.isDir &&
    !props.node.isGitignored &&
    !props.node.isCustomIgnored
  ) {
    const rootDir = projectStore.currentProject?.path;
    if (rootDir) {
      uiStore.showQuickLook({
        rootDir,
        path: props.node.relPath,
        type: "fs",
        position: { x: ev.clientX, y: ev.clientY },
        isPinned: false,
      });
    }
  }
}

let _qlMoveTimer: number | null = null;
function handleMouseMove(ev: MouseEvent) {
  // Дебаунс обновления позиции QuickLook
  if (
    isCtrlPressed.value &&
    !props.node.isDir &&
    !props.node.isGitignored &&
    !props.node.isCustomIgnored &&
    uiStore.quickLook.isActive
  ) {
    if (_qlMoveTimer) window.clearTimeout(_qlMoveTimer);
    _qlMoveTimer = window.setTimeout(() => {
      uiStore.setPosition({ x: ev.clientX + 10, y: ev.clientY + 10 });
      _qlMoveTimer = null;
    }, 24);
  }
}

function handleMouseLeave() {
  // Скрываем QuickLook если он не закреплен
  if (!uiStore.quickLook.isPinned) {
    uiStore.hideQuickLook();
  }
}
</script>