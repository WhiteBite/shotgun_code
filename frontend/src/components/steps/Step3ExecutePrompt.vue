
<template>
  <div class="p-6 flex flex-col h-full">
    <!-- ... -->
    <textarea v-model="diff.gitDiffInput"></textarea>
    <input type="number" v-model.number="diff.splitLineLimit" />
    <button @click="diff.processAndSplitDiff">Split Diff & Proceed</button>
  </div>
</template>

<script setup>
import { useDiffStore } from '../../stores/diff';
const diff = useDiffStore();
</script>