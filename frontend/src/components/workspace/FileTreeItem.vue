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
    @click.stop="handleRowClick"
    @contextmenu.prevent="handleContextMenu"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <div class="flex items-center gap-1.5 w-full">
      <button 
        v-if="node.isDir" 
        @click.stop="toggleDir($event)" 
        class="flex-shrink-0 w-4 h-4 text-gray-500 hover:text-white" 
        aria-label="Toggle directory"
      >
        <svg v-if="isExpanded" xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </button>
      <div v-else class="w-4 h-4 flex-shrink-0"></div>
      <span class="w-4 h-4 flex-shrink-0">{{ icon }}</span>
      <span class="flex-grow min-w-0 truncate" :title="titleText">{{ node.name }}</span>
      <input
        ref="cbRef"
        type="checkbox"
        :checked="isSelected"
        :disabled="node.isGitignored || node.isCustomIgnored"
        class="form-checkbox h-3 w-3"
        @click.stop
        @change="onToggleSelection"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { FileNode } from '@/types/api'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useProjectStore } from '@/stores/project.store'
import { useUiStore } from '@/stores/ui.store'
import { useTreeStateStore } from '@/stores/tree-state.store'
import { useKeyboardState } from '@/composables/useKeyboardState'
import { getFileIcon } from '@/utils/fileIcons'

const props = defineProps<{ 
  node: FileNode
  depth: number
}>()

const emit = defineEmits<{
  select: [node: FileNode]
  expand: [node: FileNode]
}>()

const fileTreeStore = useFileTreeStore()
const projectStore = useProjectStore()
const uiStore = useUiStore()
const treeStateStore = useTreeStateStore()
const { isAltPressed, isCtrlPressed } = useKeyboardState()

const cbRef = ref<HTMLInputElement | null>(null)

const icon = computed(() => getFileIcon(props.node.name))
const isSelected = computed(() => {
  // Проверяем состояние выделения в treeStateStore
  return treeStateStore.selectedPaths.has(props.node.path)
})
const isExpanded = computed(() => {
  // Check if this node is expanded in the tree state
  return treeStateStore.expandedPaths.has(props.node.path)
})
const isActive = computed(() => treeStateStore.activeNodePath === props.node.path)

const titleText = computed(() => {
  if (props.node.isDir) return props.node.relPath
  return `${props.node.relPath} • ${props.node.size.toLocaleString()} B`
})

watch(isSelected, (val) => { 
  if (cbRef.value) cbRef.value.indeterminate = false // TODO: Add partial selection logic if needed
}, { immediate: true })

function toggleDir(ev: MouseEvent) {
  // Если нажат Alt, делаем рекурсивное разворачивание
  if (isAltPressed.value) {
    treeStateStore.toggleExpansionRecursive(
      props.node.path, 
      fileTreeStore.nodesMap as Map<string, any>, 
      !isExpanded.value
    )
  } else {
    treeStateStore.toggleExpansion(props.node.path)
  }
  emit('expand', props.node)
}

function handleRowClick() {
  // Устанавливаем активный узел
  treeStateStore.activeNodePath = props.node.path
  
  if (!props.node.isDir) {
    // Для файлов - переключаем выделение
    treeStateStore.toggleNodeSelection(props.node.path, fileTreeStore.nodesMap as Map<string, any>)
    emit('select', props.node)
  } else {
    // Для директорий - разворачиваем/сворачиваем
    if (isAltPressed.value) {
      // Рекурсивное разворачивание при Alt
      treeStateStore.toggleExpansionRecursive(
        props.node.path, 
        fileTreeStore.nodesMap as Map<string, any>, 
        !isExpanded.value
      )
    } else {
      treeStateStore.toggleExpansion(props.node.path)
    }
    emit('expand', props.node)
  }
}

function handleContextMenu(event: MouseEvent) {
  uiStore.openContextMenu(event.clientX, event.clientY, props.node.path)
}

function onToggleSelection() {
  // Переключаем выделение напрямую через treeStateStore
  treeStateStore.toggleNodeSelection(props.node.path, fileTreeStore.nodesMap as Map<string, any>)
  emit('select', props.node)
}

function handleMouseEnter(ev: MouseEvent) {
  // Показываем QuickLook только при зажатом Ctrl
  if (isCtrlPressed.value && !props.node.isDir && !props.node.isGitignored && !props.node.isCustomIgnored) {
    const rootDir = projectStore.currentProject?.path
    if (rootDir) {
      uiStore.showQuickLook({
        rootDir,
        path: props.node.relPath,
        type: "fs",
        event: ev,
        isPinned: false, // Временный QuickLook
      })
    }
  }
}

function handleMouseLeave() {
  // Скрываем QuickLook если он не закреплен
  if (!uiStore.quickLook.isPinned) {
    uiStore.hideQuickLook()
  }
}
</script>