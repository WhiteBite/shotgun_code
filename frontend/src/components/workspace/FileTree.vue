<template>
  <div class="relative h-full overflow-hidden">
    <RecycleScroller
        class="h-full"
        :items="nodes"
        :item-size="24"
        key-field="path"
        v-slot="{ item }"
    >
      <FileTreeItem :node="item" @select-folder="handleSelectFolder" />
    </RecycleScroller>
  </div>
</template>

<script setup lang="ts">
// eslint-disable-next-line @typescript-eslint/no-unused-vars
import { computed } from "vue"; // 'computed' is no longer directly used in <script setup>, only in the template, so it's fine.
import { RecycleScroller } from "vue-virtual-scroller";
import "vue-virtual-scroller/dist/vue-virtual-scroller.css";
import FileTreeItem from "./FileTreeItem.vue";
import type { FileNode } from "@/types/dto";

const props = defineProps<{
  nodes: FileNode[];
}>();

// Emits for folder selection
const emit = defineEmits(["select-folder"]); // eslint-disable-line @typescript-eslint/no-unused-vars

function handleSelectFolder(folderPath: string) {
  emit("select-folder", folderPath);
}
</script>

<style>
/* Adjust vue-virtual-scroller styles if needed */
.vue-recycle-scroller__item-wrapper {
  overflow: visible !important; /* Allow tooltips and context menus to overflow */
}
</style>