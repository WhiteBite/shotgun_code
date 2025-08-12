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
        @mouseenter="quickLook.handleMouseEnter($event, node)"
        @mouseleave="quickLook.handleMouseLeave"
    >
      <div class="w-4 h-4 flex-shrink-0 flex items-center justify-center z-10" @click.stop="handleClickExpansion">
        <svg v-if="node.isDir" xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="text-gray-400 hover:text-blue-400 transition-all" :class="{'rotate-90 text-blue-400': isExpanded}">
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </div>

      <input
          type="checkbox"
          :checked="selectionState === 'on'"
          :indeterminate="selectionState === 'partial'"
          :disabled="node.isIgnored"
          @click.stop
          @change.stop="treeActions.toggleNodeSelection(node.path)"
          class="form-checkbox w-4 h-4 z-10 disabled:opacity-50 disabled:cursor-not-allowed hover:border-blue-400 transition-colors"
      />

      <div class="w-5 h-5 flex-shrink-0 flex items-center justify-center z-10 -ml-1 text-base">
        <span v-if="node.isDir && isExpanded" class="text-blue-400">📂</span>
        <span v-else-if="node.isDir && !isExpanded" class="text-gray-400">📁</span>
        <span v-else>{{ getFileIcon(node.name) }}</span>
      </div>

      <div class="flex-grow flex items-center justify-between overflow-hidden">
        <span class="text-sm truncate select-none transition-colors" :class="getTextColor()">
          {{ node.name }}
        </span>
        <div class="flex-shrink-0 ml-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button v-if="!node.isDir && !node.isIgnored" @click.stop="quickLook.showPinnedQuickLook($event, node)" class="p-1 rounded hover:bg-blue-500/30" title="Quick Look (Space)">
            👁️
          </button>
        </div>
      </div>
    </div>
  </Tooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { FileNode } from '@/types/dto';
import { useTreeStateStore } from '@/stores/tree-state.store';
import { useContextStore } from '@/stores/context.store';
import { useTreeActions } from '@/composables/useTreeActions';
import { useQuickLook } from '@/composables/useQuickLook';
import Tooltip from '@/components/shared/Tooltip.vue';

const props = defineProps<{ node: FileNode }>();

const treeStateStore = useTreeStateStore();
const contextStore = useContextStore();
const treeActions = useTreeActions();
const quickLook = useQuickLook();

const isActive = computed(() => treeStateStore.activeNodePath === props.node.path);
const isExpanded = computed(() => treeStateStore.expandedPaths.has(props.node.path));

const selectionState = computed<'on' | 'off' | 'partial'>(() => {
  if (treeStateStore.selectedPaths.has(props.node.path)) {
    return 'on';
  }
  if (!props.node.isDir || !props.node.children) {
    return 'off';
  }

  const children = props.node.children.map(c => contextStore.nodesMap.get(c.path)).filter(Boolean) as FileNode[];
  const selectedChildren = children.filter(c => treeStateStore.selectedPaths.has(c.path));

  if (selectedChildren.length === 0) return 'off';
  if (selectedChildren.length === children.length) {
    // This is not fully correct without recursive check, but visually sufficient
    return 'on';
  }
  return 'partial';
});


const tooltipText = computed(() => {
  if (props.node.isIgnored) {
    if (props.node.isGitignored && props.node.isCustomIgnored) return 'Ignored by .gitignore & Custom Rules';
    if (props.node.isGitignored) return 'Ignored by .gitignore';
    if (props.node.isCustomIgnored) return 'Ignored by Custom Rules';
  }
  return props.node.relPath;
});

function getFileIcon(fileName: string): string {
  const ext = fileName.split('.').pop()?.toLowerCase() || '';
  const iconMap: Record<string, string> = {
    'js': '🟨', 'ts': '🔷', 'json': '📋', 'md': '📝', 'vue': '💚',
    'go': '💧', 'py': '🐍', 'html': '🌐', 'css': '🎨', 'scss': '🎨',
    'dockerfile': '🐳', 'yml': '📋', 'yaml': '📋', 'gitignore': '🚫',
    'ps1': '📜', 'sh': '📜',
  };
  return iconMap[ext] || '📄';
}

function getTextColor(): string {
  if (props.node.isIgnored) return 'text-gray-500';
  if (isActive.value) return 'text-white font-medium';
  if (selectionState.value !== 'off') return 'text-blue-200';
  return 'text-gray-300';
}

function handleClick(event: MouseEvent) {
  if (props.node.isIgnored) return;
  treeActions.setActiveNode(props.node.path);
  if (props.node.isDir) {
    treeActions.toggleNodeExpansion(props.node.path, event.altKey);
  }
}

function handleClickExpansion(event: MouseEvent) {
  if (props.node.isDir) {
    treeActions.toggleNodeExpansion(props.node.path, event.altKey);
  }
}
</script>