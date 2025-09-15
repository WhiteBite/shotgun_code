<template>
  <div class="file-tree overflow-auto h-full">
    <VirtualFileTree ref="virtualFileTree" />
  </div>
</template>

<script setup lang="ts">
import VirtualFileTree from "./VirtualFileTree.vue";
import { ref } from "vue";

const virtualFileTree = ref<InstanceType<typeof VirtualFileTree> | null>(null);
</script>

<style scoped>
.file-tree {
  height: 100%;
}
</style>
