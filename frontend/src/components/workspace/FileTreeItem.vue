<template>
  <Tooltip :text="tooltipText" position="top-start">
    <div
        class="relative group flex items-center gap-1.5 h-[28px] pr-2 rounded"
        :class="{
          'bg-blue-600/30': isSelected,
          'opacity-60': node.isIgnored,
          'hover:bg-blue-500/20': !node.isIgnored && !isSelected,
          'cursor-pointer': !node.isIgnored,
          'cursor-not-allowed': node.isIgnored,
        }"
        @click.stop="handleClick"
        @mouseenter="handleMouseEnter"
        @mouseleave="handleMouseLeave"
    >
      <div
          v-for="i in node.depth"
          :key="i"
          class="absolute top-0 w-px bg-gray-600/40 group-hover:bg-blue-400/50 transition-colors"
          :style="{ left: `${(i - 1) * 16 + 20}px`, height: '28px' }"
      ></div>

      <div class="w-4 h-4 flex-shrink-0 flex items-center justify-center z-10" @click.stop="handleClickExpansion">
        <svg v-if="node.isDir" xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="text-gray-400 hover:text-blue-400 transition-all" :class="{'rotate-90 text-blue-400': node.expanded}">
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </div>

      <input
          type="checkbox"
          :checked="node.selected === 'on'"
          :indeterminate="node.selected === 'partial'"
          :disabled="node.isIgnored"
          @click.stop
          @change.stop="tree.toggleSelection(node.path)"
          class="form-checkbox w-4 h-4 z-10 disabled:opacity-50 disabled:cursor-not-allowed hover:border-blue-400 transition-colors"
      />

      <div class="w-5 h-5 flex-shrink-0 flex items-center justify-center z-10 -ml-1 text-base">
        <span v-if="node.isDir && node.expanded" class="text-blue-400">📂</span>
        <span v-else-if="node.isDir && !node.expanded" class="text-gray-400">📁</span>
        <span v-else>{{ getFileIcon(node.name) }}</span>
      </div>

      <div class="flex-grow flex items-center justify-between overflow-hidden">
        <span class="text-sm truncate select-none transition-colors" :class="getTextColor()">
          {{ node.name }}
        </span>
        <div class="flex-shrink-0 ml-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <span v-if="node.contextOrigin !== ContextOrigin.None" class="px-1.5 py-0.5 text-xs rounded-full" :class="getOriginBadgeClass()" :title="`Added from: ${node.contextOrigin}`">
            {{ getOriginLabel() }}
          </span>
          <span v-if="gitStatus" class="px-1.5 py-0.5 text-xs font-semibold rounded-full" :class="gitStatus.classes">
            {{ gitStatus.label }}
          </span>
        </div>
      </div>
    </div>
  </Tooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { FileNode } from '@/types/dto';
import { useContextStore } from '@/stores/context.store';
import { useUiStore } from '@/stores/ui.store';
import { useFileTree } from '@/composables/useFileTree';
import { GitStatus, ContextOrigin } from '@/types/enums';
import Tooltip from '@/components/shared/Tooltip.vue';

const props = defineProps<{ node: FileNode }>();

const contextStore = useContextStore();
const uiStore = useUiStore();
const tree = useFileTree();
const isSelected = computed(() => contextStore.activeNodePath === props.node.path);

const tooltipText = computed(() => {
  if (props.node.isIgnored) {
    if (props.node.isGitignored && props.node.isCustomIgnored) return 'Ignored by .gitignore & Custom Rules';
    if (props.node.isGitignored) return 'Ignored by .gitignore';
    if (props.node.isCustomIgnored) return 'Ignored by Custom Rules';
  }
  return props.node.relPath;
});

const gitStatus = computed(() => {
  if (props.node.isDir) return null;
  switch(props.node.gitStatus) {
    case GitStatus.Modified: return { label: 'M', classes: 'bg-amber-500/30 text-amber-300 border border-amber-500/50' };
    case GitStatus.Untracked: return { label: 'U', classes: 'bg-green-500/30 text-green-300 border border-green-500/50' };
    case GitStatus.Conflict: return { label: 'C', classes: 'bg-red-500/30 text-red-300 border border-red-500/50' };
    default: return null;
  }
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
  if (isSelected.value) return 'text-white font-medium';
  if (props.node.selected === 'on') return 'text-blue-200';
  return 'text-gray-300';
}

function getOriginBadgeClass(): string {
  switch(props.node.contextOrigin) {
    case ContextOrigin.Manual: return 'bg-blue-500/30 text-blue-200 border border-blue-500/50';
    case ContextOrigin.Git: return 'bg-green-500/30 text-green-200 border border-green-500/50';
    case ContextOrigin.AI: return 'bg-purple-500/30 text-purple-200 border border-purple-500/50';
    default: return '';
  }
}
function getOriginLabel(): string {
  return props.node.contextOrigin.charAt(0).toUpperCase();
}

function handleClick(event: MouseEvent) {
  if (props.node.isIgnored) return;
  tree.setActiveNode(props.node.path);
  if (props.node.isDir) {
    tree.toggleExpansion(props.node.path, event.altKey);
  }
}
function handleClickExpansion(event: MouseEvent) {
  if (props.node.isDir) {
    tree.toggleExpansion(props.node.path, event.altKey);
  }
}

function handleMouseEnter(event: MouseEvent) {
  if ((event.ctrlKey || event.metaKey) && !props.node.isDir && !props.node.isIgnored) {
    uiStore.showQuickLook({ path: props.node.relPath, type: 'fs', event });
  }
}
function handleMouseLeave() {
  uiStore.requestHideQuickLook();
}
</script>