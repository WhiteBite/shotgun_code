<template>
  <ul class="space-y-0.5">
    <li v-for="node in nodes" :key="node.path">
      <div
          @click.stop="toggleExpansion(node)"
          :class="['flex items-center gap-1.5 p-1 rounded-md cursor-pointer', {'bg-gray-700/50': isSelected(node)}]"
          :style="{ 'padding-left': `${depth * 16 + 4}px` }"
      >
        <!-- Checkbox -->
        <input
            type="checkbox"
            :checked="node.selected === 'on'"
            :indeterminate="node.selected === 'partial'"
            @click.stop
            @change.stop="contextStore.toggleNodeSelection(node)"
            class="form-checkbox w-4 h-4 bg-transparent border-gray-500 rounded text-blue-500 focus:ring-blue-500/50 focus:ring-offset-0"
        >

        <!-- Toggler Icon -->
        <div class="w-4 h-4 flex-shrink-0 flex items-center justify-center">
          <svg v-if="node.isDir" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-gray-500 transition-transform" :class="{'rotate-90': node.expanded}">
            <polyline points="9 18 15 12 9 6"></polyline>
          </svg>
        </div>

        <!-- File/Folder Icon -->
        <div class="w-5 h-5 flex-shrink-0 flex items-center justify-center text-gray-400">
          <svg v-if="node.isDir && node.expanded" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-blue-400"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"></path></svg>
          <svg v-else-if="node.isDir && !node.expanded" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"></path><polyline points="14 2 14 8 20 8"></polyline></svg>
        </div>

        <!-- Name -->
        <span class="text-sm truncate select-none" :class="{'text-white': node.selected !== 'off', 'text-gray-400': node.selected === 'off'}">{{ node.name }}</span>
      </div>

      <!-- Recursive render for children -->
      <FileTree v-if="node.isDir && node.expanded && node.children.length > 0" :nodes="node.children" :depth="depth + 1" />
    </li>
  </ul>
</template>

<script lang="ts">
export default {
  name: 'FileTree'
}
</script>

<script setup lang="ts">
import { ref } from 'vue';
import { useContextStore } from '@/stores/context.store';
import type { FileNode } from '@/types/dto';

withDefaults(defineProps<{
  nodes: FileNode[];
  depth?: number;
}>(), {
  depth: 0,
});

const contextStore = useContextStore();
const selectedNodePath = ref<string | null>(null);

const toggleExpansion = (node: FileNode) => {
  if (node.isDir) {
    node.expanded = !node.expanded;
  }
  selectedNodePath.value = node.path;
};

const isSelected = (node: FileNode) => {
  return selectedNodePath.value === node.path;
};
</script>