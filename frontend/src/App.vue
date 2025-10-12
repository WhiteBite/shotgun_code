﻿<template>
  <div id="app" class="h-screen bg-gray-900 text-white overflow-hidden">
    <!-- Global Error Handler -->
    <div v-if="globalError" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div class="bg-red-900 border border-red-700 rounded-lg p-6 max-w-md mx-4">
        <h3 class="text-lg font-semibold text-red-200 mb-2">Critical Error</h3>
        <p class="text-red-300 text-sm mb-4">{{ globalError }}</p>
        <button
          @click="clearGlobalError"
          class="px-4 py-2 bg-red-700 hover:bg-red-600 text-white rounded transition-colors"
        >
          Dismiss
        </button>
      </div>
    </div>

    <!-- Global Loading Overlay -->
    <div v-if="projectStore.isLoading" class="fixed inset-0 z-40 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-gray-800 border border-gray-600 rounded-lg p-8 max-w-sm mx-4 text-center">
        <div class="flex items-center justify-center mb-4">
          <svg
            class="animate-spin h-8 w-8 text-blue-500"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
            ></path>
          </svg>
        </div>
        <h3 class="text-lg font-semibold text-white mb-2">Loading Project</h3>
        <p class="text-gray-300 text-sm">Please wait...</p>
      </div>
    </div>

    <!-- Main Content: Show ProjectSelector if no project, otherwise show workspace -->
    <ProjectSelector v-if="!projectStore.hasProject" @opened="onProjectOpened" />
    <MainWorkspace v-else />

    <!-- Toast Notifications -->
    <div class="fixed top-4 right-4 z-50 space-y-2">
      <div
        v-for="toast in uiStore.toasts"
        :key="toast.id"
        :class="[
          'px-4 py-3 rounded-lg shadow-lg max-w-sm',
          toast.type === 'error' ? 'bg-red-900 border border-red-700 text-red-100' :
          toast.type === 'success' ? 'bg-green-900 border border-green-700 text-green-100' :
          toast.type === 'warning' ? 'bg-yellow-900 border border-yellow-700 text-yellow-100' :
          'bg-blue-900 border border-blue-700 text-blue-100'
        ]"
      >
        {{ toast.message }}
      </div>
    </div>

    <!-- Global components (lazy loaded) -->
    <CommandPalette v-if="projectStore.hasProject" />
    <KeyboardShortcutsModal v-if="projectStore.hasProject" />
    
    <!-- Theme toggle (always available) -->
    <div class="fixed bottom-4 right-4 z-40">
      <ThemeToggle />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineAsyncComponent } from 'vue'
import ProjectSelector from '@/components/ProjectSelector.vue'
import MainWorkspace from '@/components/workspace/MainWorkspace.vue'
import { useProjectStore } from '@/stores/project.store'
import { useUIStore } from '@/stores/ui.store'
import { useTheme } from '@/composables/useTheme'

// Lazy load heavy components
const CommandPalette = defineAsyncComponent(() => import('@/components/CommandPalette.vue'))
const KeyboardShortcutsModal = defineAsyncComponent(() => import('@/components/KeyboardShortcutsModal.vue'))
const ThemeToggle = defineAsyncComponent(() => import('@/components/ThemeToggle.vue'))

const projectStore = useProjectStore()
const uiStore = useUIStore()
const globalError = ref<string | null>(null)

// Initialize theme early to prevent white flash on load
useTheme()

function clearGlobalError() {
  globalError.value = null
}

function onProjectOpened(path: string) {
  console.log('Project opened:', path)
  uiStore.addToast('Project loaded successfully', 'success')
}

onMounted(() => {
  // Try to auto-open last project if setting is enabled
  projectStore.maybeAutoOpenLastProject()
})
</script>

<style>
#app {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  font-feature-settings: 'cv02', 'cv03', 'cv04', 'cv11';
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Global scrollbar styles */
* {
  scrollbar-width: thin;
  scrollbar-color: #6b7280 rgba(31, 41, 55, 0.4);
}

*::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

*::-webkit-scrollbar-track {
  background: rgba(31, 41, 55, 0.4);
  border-radius: 4px;
}

*::-webkit-scrollbar-thumb {
  background: #6b7280;
  border-radius: 4px;
  border: 1px solid rgba(31, 41, 55, 0.6);
}

*::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

*::-webkit-scrollbar-corner {
  background: rgba(31, 41, 55, 0.4);
}

/* Focus styles */
*:focus {
  outline: none;
}

/* Scope focus outline to interactive controls to avoid artifacts during resize */
:where(
  a,
  button,
  input,
  select,
  textarea,
  [role="button"],
  [role="checkbox"],
  [role="menuitem"],
  [role="tab"],
  [role="switch"],
  [contenteditable],
  [tabindex]:not([tabindex="-1"]) 
):focus-visible {
  outline: 2px solid #3b82f6;
  outline-offset: 2px;
}

/* Never show focus ring on layout/resizer elements */
[role="separator"],
[data-resize-handle],
.splitpanes__splitter,
.vue-resizable-handle {
  outline: none !important;
}

/* Transition defaults */
* {
  transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, backdrop-filter;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

/* Selection styling */
::selection {
  background-color: rgba(59, 130, 246, 0.3);
  color: #ffffff;
}

::-moz-selection {
  background-color: rgba(59, 130, 246, 0.3);
  color: #ffffff;
}

/* Animations */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.fade-in {
  animation: fadeIn 0.3s ease-out;
}

.slide-in {
  animation: slideIn 0.3s ease-out;
}

/* Improved button focus and hover states */
button:focus-visible {
  outline: 2px solid #3b82f6;
  outline-offset: 2px;
}

/* Better form input styles */
input, textarea, select {
  transition: all 0.2s ease-in-out;
}

input:focus, textarea:focus, select:focus {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* Loading animation improvements */
.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>