<template>
  <div
      class="relative group flex items-center gap-1.5 h-[28px] pr-2 rounded cursor-pointer"
      :class="{'bg-gray-800/60': isSelected}"
      :style="{ paddingLeft: `${node.depth * 16 + 4}px` }"
      @click.stop="handleClick"
  >
    <!-- Indent guides -->
    <div
        v-for="i in node.depth"
        :key="i"
        class="absolute top-0 w-px bg-gray-700/40 group-hover:bg-gray-600/50 transition-colors"
        :style="{ left: `${(i - 1) * 16 + 18}px`, height: '28px' }"
    ></div>

    <!-- Toggler Icon -->
    <div class="w-4 h-4 flex-shrink-0 flex items-center justify-center z-10" @click.stop="tree.toggleExpansion(node.path)">
      <svg v-if="node.isDir" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-gray-500 transition-transform" :class="{'rotate-90': node.expanded}">
        <polyline points="9 18 15 12 9 6"></polyline>
      </svg>
    </div>

    <!-- Checkbox is now part of the item, logic inside store -->
    <input
        type="checkbox"
        :checked="node.selected === 'on'"
        :indeterminate="node.selected === 'partial'"
        @click.stop
        @change.stop="tree.toggleSelection(node.path)"
        class="form-checkbox w-4 h-4 bg-gray-800 border-gray-500 rounded text-blue-500 focus:ring-blue-500/50 focus:ring-offset-0 transition-opacity z-10"
        :class="{
        'opacity-0 group-hover:opacity-100': contextStore.treeMode === TreeMode.Navigation && node.selected === 'off',
        'opacity-100': contextStore.treeMode === TreeMode.Selection || node.selected !== 'off'
      }"
    />

    <!-- Icon -->
    <div class="w-5 h-5 flex-shrink-0 flex items-center justify-center text-gray-400 -ml-1.5 z-10">
      <svg v-if="node.isDir && node.expanded" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-blue-400"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"></path></svg>
      <svg v-else-if="node.isDir && !node.expanded" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
      <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"></path><polyline points="14 2 14 8 20 8"></polyline></svg>
    </div>

    <!-- Name and Badges -->
    <div class="flex-grow flex items-center justify-between overflow-hidden">
      <div class="truncate">
        <span class="text-sm truncate select-none" :class="{'text-white': isSelected, 'text-gray-300': !isSelected}">{{ node.isCompact ? node.name.split('/').pop() : node.name }}</span>
        <span v-if="node.isCompact" class="text-xs text-gray-500 ml-1 truncate">{{ compactPath(node.name) }}</span>
      </div>
      <div class="flex-shrink-0 ml-2">
        <span v-if="gitStatus" class="px-1.5 py-0.5 text-xs font-semibold rounded-full" :class="gitStatus.classes">{{ gitStatus.label }}</span>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { FileNode } from '@/types/dto';
import { useContextStore } from '@/stores/contextStore';
import { useFileTree } from '@/composables/useFileTree';
import { TreeMode, GitStatus } from '@/types/enums';

const props = defineProps<{
  node: FileNode & { isCompact?: boolean };
}>();

const contextStore = useContextStore();
const tree = useFileTree();
const isSelected = computed(() => contextStore.activeNodePath === props.node.path);

const gitStatus = computed(() => {
  if (props.node.isDir) return null;
  switch(props.node.gitStatus) {
    case GitStatus.Modified: return { label: 'M', classes: 'bg-amber-500/20 text-amber-400' };
    case GitStatus.Untracked: return { label: 'U', classes: 'bg-gray-500/20 text-gray-400' };
    case GitStatus.Conflict: return { label: 'C', classes: 'bg-red-500/20 text-red-400' };
    default: return null;
  }
});


function handleClick() {
  tree.setActiveNode(props.node.path);
  if (contextStore.treeMode === TreeMode.Selection) {
    tree.toggleSelection(props.node.path);
  } else if (props.node.isDir) {
    tree.toggleExpansion(props.node.path);
  }
}

function compactPath(path: string) {
  const parts = path.split('/');
  parts.pop();
  return parts.join('/') + '/';
}
</script>