<template>
  <RecycleScroller
      v-if="contextStore.visibleNodes.length > 0"
      class="h-full"
      :items="contextStore.visibleNodes"
      :item-size="28"
      key-field="path"
      v-slot="{ item }"
  >
    <FileTreeItem :node="item" />
  </RecycleScroller>
  <div v-else class="p-4 text-center text-gray-500 text-sm">
    Загрузка дерева...
  </div>
</template>

<script setup lang="ts">
import { RecycleScroller } from 'vue-virtual-scroller';
import { useContextStore } from '@/stores/contextStore';
import FileTreeItem from './FileTreeItem.vue';

const contextStore = useContextStore();
</script>

<style>
.vue-recycle-scroller__item-view {
  box-sizing: border-box;
}
</style>