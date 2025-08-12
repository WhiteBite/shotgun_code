<template>
  <RecycleScroller
      v-if="nodes.length > 0"
      class="h-full"
      :items="nodes"
      :item-size="28"
      key-field="path"
      v-slot="{ item }"
  >
    <FileTreeItem :node="item" />
  </RecycleScroller>
  <div v-else class="p-4 text-center text-gray-500 text-sm">
    No files or folders to show.
  </div>
</template>

<script setup lang="ts">
import { RecycleScroller } from 'vue-virtual-scroller';
import FileTreeItem from './FileTreeItem.vue';
import type { FileNode } from '@/types/dto';

defineProps<{ nodes: FileNode[] }>();
</script>

<style>
.vue-recycle-scroller__item-view {
  box-sizing: border-box;
}
.vue-recycle-scroller.ready {
  width: 100%;
}
</style>