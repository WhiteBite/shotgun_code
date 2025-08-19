<!-- SplitPaneManager.vue -->
<template>
  <div class="split-pane-manager h-full flex flex-col">
    <div class="split-toolbar flex items-center justify-between p-2 border-b border-gray-700">
      <div class="flex items-center gap-2">
        <button
            @click="addVerticalSplit"
            class="flex items-center gap-1 px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-xs"
            title="Add vertical split"
        >
          ⫾ Split Vertical
        </button>
        <button
            @click="addHorizontalSplit"
            class="flex items-center gap-1 px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-xs"
            title="Add horizontal split"
        >
          ⫿ Split Horizontal
        </button>
        <button
            @click="closeSplit"
            v-if="layout !== 'single'"
            class="flex items-center gap-1 px-2 py-1 bg-red-700 hover:bg-red-600 rounded text-xs"
            title="Close split"
        >
          ✕ Close Split
        </button>
      </div>
      <div class="flex items-center gap-2">
        <select v-model="layout" class="bg-gray-700 border-gray-600 rounded px-2 py-1 text-xs">
          <option value="single">Single Pane</option>
          <option value="vertical">Two Vertical</option>
          <option value="horizontal">Two Horizontal</option>
          <option value="grid">Four Grid</option>
        </select>
      </div>
    </div>

    <div class="split-content flex-grow">
      <component :is="layoutComponent" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import SinglePane from './layouts/SinglePane.vue'
import VerticalSplit from './layouts/VerticalSplit.vue'
import HorizontalSplit from './layouts/HorizontalSplit.vue'
import GridLayout from './layouts/GridLayout.vue'

const layout = ref('single')

const layoutComponent = computed(() => {
  switch (layout.value) {
    case 'vertical': return VerticalSplit
    case 'horizontal': return HorizontalSplit
    case 'grid': return GridLayout
    default: return SinglePane
  }
})

function addVerticalSplit() {
  layout.value = 'vertical'
}

function addHorizontalSplit() {
  layout.value = 'horizontal'
}

function closeSplit() {
  layout.value = 'single'
}
</script>