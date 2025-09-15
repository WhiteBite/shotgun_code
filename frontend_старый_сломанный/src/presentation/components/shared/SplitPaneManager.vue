<!-- SplitPaneManager.vue -->
<template>
  <div class="split-pane-manager h-full flex flex-col">
    <div
      class="split-toolbar flex items-center justify-between p-2 border-b border-gray-700"
    >
      <div class="flex items-center gap-2">
        <button
          class="flex items-center gap-1 px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-xs"
          title="Add vertical split"
          @click="addVerticalSplit"
        >
          ⫾ Split Vertical
        </button>
        <button
          class="flex items-center gap-1 px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-xs"
          title="Add horizontal split"
          @click="addHorizontalSplit"
        >
          ⫿ Split Horizontal
        </button>
        <button
          v-if="layout !== 'single'"
          class="flex items-center gap-1 px-2 py-1 bg-red-700 hover:bg-red-600 rounded text-xs"
          title="Close split"
          @click="closeSplit"
        >
          ✕ Close Split
        </button>
      </div>
      <div class="flex items-center gap-2">
        <select
          v-model="layout"
          class="bg-gray-700 border-gray-600 rounded px-2 py-1 text-xs"
        >
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
import { ref, computed } from "vue";
import SplitPane from "./SplitPane.vue";

// Define proper components for each layout
const SinglePane = {
  name: "SinglePane",
  template: `
    <div class="single-pane h-full">
      <slot name="default" />
    </div>
  `,
};

const VerticalSplit = {
  name: "VerticalSplit",
  template: `
    <SplitPane 
      direction="horizontal" 
      :initial-ratio="0.5"
      class="h-full"
    >
      <template #left>
        <div class="panel-left h-full bg-gray-800 p-4">
          <slot name="left" />
        </div>
      </template>
      <template #right>
        <div class="panel-right h-full bg-gray-800 p-4">
          <slot name="right" />
        </div>
      </template>
    </SplitPane>
  `,
  components: { SplitPane }
};

const HorizontalSplit = {
  name: "HorizontalSplit",
  template: `
    <SplitPane 
      direction="vertical" 
      :initial-ratio="0.5"
      class="h-full"
    >
      <template #left>
        <div class="panel-top h-full bg-gray-800 p-4">
          <slot name="top" />
        </div>
      </template>
      <template #right>
        <div class="panel-bottom h-full bg-gray-800 p-4">
          <slot name="bottom" />
        </div>
      </template>
    </SplitPane>
  `,
  components: { SplitPane }
};

const GridLayout = {
  name: "GridLayout",
  template: `
    <div class="grid-layout h-full grid grid-cols-2 grid-rows-2 gap-1">
      <div class="panel-top-left bg-gray-800 p-4">
        <slot name="top-left" />
      </div>
      <div class="panel-top-right bg-gray-800 p-4">
        <slot name="top-right" />
      </div>
      <div class="panel-bottom-left bg-gray-800 p-4">
        <slot name="bottom-left" />
      </div>
      <div class="panel-bottom-right bg-gray-800 p-4">
        <slot name="bottom-right" />
      </div>
    </div>
  `,
};

const layout = ref("single");

const layoutComponent = computed(() => {
  switch (layout.value) {
    case "vertical":
      return VerticalSplit;
    case "horizontal":
      return HorizontalSplit;
    case "grid":
      return GridLayout;
    default:
      return SinglePane;
  }
});

function addVerticalSplit() {
  layout.value = "vertical";
}

function addHorizontalSplit() {
  layout.value = "horizontal";
}

function closeSplit() {
  layout.value = "single";
}
</script>