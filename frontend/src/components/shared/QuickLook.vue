<!-- QuickLook.vue (исправлено позиционирование по центру) -->
<template>
  <teleport to="body">
    <div
      v-if="quickLookState.isActive"
      class="fixed inset-0 z-[30] flex items-center justify-center"
      :class="overlayClasses"
      role="dialog"
      aria-modal="true"
      @click.self="close"
    >
      <div
        ref="panelRef"
        class="quicklook-panel bg-gray-800/95 backdrop-blur-sm shadow-2xl rounded-lg border border-gray-600 overflow-hidden pointer-events-auto"
        :class="panelClasses"
        :style="panelStyle"
        @mouseleave="handlePanelMouseLeave"
      >
        <!-- Header (drag handle) -->
        <div
          class="quicklook-header p-2 border-b border-gray-700 bg-gray-800/80 flex justify-between items-center cursor-move"
          @mousedown="startDrag"
        >
          <span class="text-xs text-gray-400 font-mono truncate">{{ quickLookState.path }}</span>
          <div class="flex items-center gap-1">
            <button
              @click.stop="togglePin"
              class="p-1 rounded hover:bg-gray-700 transition-colors"
              :title="quickLookState.isPinned ? 'Unpin' : 'Pin'"
              aria-label="Pin"
            >
              {{ quickLookState.isPinned ? '📍' : '📌' }}
            </button>
            <button
              @click.stop="close"
              class="p-1 rounded hover:bg-gray-700 transition-colors"
              title="Close (Esc)"
              aria-label="Close"
            >
              ✖
            </button>
          </div>
        </div>

        <!-- Content -->
        <div class="quicklook-content p-3 overflow-auto" style="max-height: 70vh; max-width: 70vw;">
          <div v-if="quickLookState.truncated" class="mb-2 text-xs text-yellow-300 bg-yellow-900/20 border border-yellow-700 rounded px-2 py-1">
            Preview truncated. Open the file to see full content.
          </div>
          <div v-if="quickLookState.error" class="text-red-400 text-sm">
            Error: {{ quickLookState.error }}
          </div>
          <pre v-else class="text-sm"><code
            class="hljs"
            :class="`language-${quickLookState.language}`"
            v-html="quickLookState.content"
          ></code></pre>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useUiStore } from '@/stores/ui.store'

const uiStore = useUiStore()
const quickLookState = computed(() => uiStore.quickLook)

const panelRef = ref<HTMLElement>()
const isDragging = ref(false)
const dragOffset = ref({ x: 0, y: 0 })

const overlayClasses = computed(() => ({
  'bg-black/50': quickLookState.value.isPinned,
  'bg-transparent': !quickLookState.value.isPinned,
}))

const panelClasses = computed(() => ({
  'shadow-2xl': quickLookState.value.isPinned,
  'shadow-lg': !quickLookState.value.isPinned,
}))

const panelStyle = computed(() => {
  if (quickLookState.value.isPinned && quickLookState.value.position) {
    return {
      position: 'absolute' as const,
      left: `${quickLookState.value.position.x}px`,
      top: `${quickLookState.value.position.y}px`,
      transform: 'none',
    }
  }
  // Всегда центрируем непинованное окно
  return {
    position: 'relative' as const,
    maxWidth: '600px',
    width: '90vw',
    margin: '0 auto',
  }
})

function handlePanelMouseLeave() {
  if (!quickLookState.value.isPinned) {
    uiStore.hideQuickLook()
  }
}

function togglePin() {
  uiStore.togglePin()
}

function close() {
  uiStore.hideQuickLook()
}

// Drag только от header
function startDrag(event: MouseEvent) {
  if (!quickLookState.value.isPinned || !panelRef.value) return
  isDragging.value = true
  const rect = panelRef.value.getBoundingClientRect()
  dragOffset.value = {
    x: event.clientX - rect.left,
    y: event.clientY - rect.top,
  }
  document.addEventListener('mousemove', handleDrag)
  document.addEventListener('mouseup', stopDrag)
}

function handleDrag(event: MouseEvent) {
  if (!isDragging.value || !panelRef.value) return
  const rect = panelRef.value.getBoundingClientRect()
  const newX = Math.max(0, Math.min(window.innerWidth - rect.width, event.clientX - dragOffset.value.x))
  const newY = Math.max(0, Math.min(window.innerHeight - rect.height, event.clientY - dragOffset.value.y))
  uiStore.setPosition({ x: newX, y: newY })
}

function stopDrag() {
  isDragging.value = false
  document.removeEventListener('mousemove', handleDrag)
  document.removeEventListener('mouseup', stopDrag)
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    close()
  }
}

onMounted(() => { document.addEventListener('keydown', handleKeydown) })
onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.removeEventListener('mousemove', handleDrag)
  document.removeEventListener('mouseup', stopDrag)
})
</script>

<style scoped>
.quicklook-panel { min-width: 400px; min-height: 200px; max-width: 70vw; max-height: 80vh; }
.quicklook-header { user-select: none; }
.quicklook-content { scrollbar-width: thin; scrollbar-color: #4b5563 rgba(31, 41, 55, 0.5); }
</style>
