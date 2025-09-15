<template>
  <transition
    enter-active-class="transition ease-out duration-100"
    enter-from-class="transform opacity-0 scale-95"
    enter-to-class="transform opacity-100 scale-100"
    leave-active-class="transition ease-in duration-75"
    leave-from-class="transform opacity-100 scale-100"
    leave-to-class="opacity-0 scale-95"
  >
    <div
      v-if="contextMenuService.getContextMenuState()?.isOpen"
      class="fixed z-50 w-56 rounded-md shadow-lg bg-gray-800 ring-1 ring-black ring-opacity-5 border border-gray-600 focus:outline-none"
      :style="menuStyle"
      role="menu"
      tabindex="0"
      @keydown.stop.prevent="onKeydown"
      @contextmenu.stop
      @click.stop
    >
      <div class="py-1" role="none">
        <button
          ref="itemRefs"
          type="button"
          class="w-full text-left text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700"
          role="menuitem"
          @click="handleToggleSelection"
        >
          {{ isTargetSelected ? "Убрать из контекста" : "Добавить в контекст" }}
        </button>
        <button
          ref="itemRefs"
          type="button"
          class="w-full text-left text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700"
          role="menuitem"
          @click="handleCopyPath"
        >
          Копировать относительный путь
        </button>
        <button
          v-if="targetNode?.isDir"
          ref="itemRefs"
          type="button"
          class="w-full text-left text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700"
          role="menuitem"
          @click="addFolderRecursive"
        >
          Добавить папку рекурсивно в контекст
        </button>
        <button
          ref="itemRefs"
          type="button"
          class="w-full text-left text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700"
          role="menuitem"
          @click="handleToggleIgnore"
        >
          {{ ruleExists ? "Убрать из игнора" : "Игнорировать путь" }}
        </button>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue"
import { useContextMenuService } from "@/composables/useContextMenuService"
import { useUIStateService } from "@/composables/useUIStateService"
import { useFileTreeStateService } from "@/composables/useFileTreeStateService"
import { useIgnoreRules } from "@/composables/useIgnoreRules"

const { contextMenuService } = useContextMenuService()
const { uiStateService } = useUIStateService()
const { fileTreeStateService } = useFileTreeStateService()
const { normalizeRule, hasRule, toggleIgnore } = useIgnoreRules()

// For now, we'll keep some direct store access since we need to fully understand
// the data structures and interactions before completely removing store dependencies
import { useUiStore } from "@/stores/ui.store"
import { useFileTreeStore } from "@/stores/file-tree.store"
import { useTreeStateStore } from "@/stores/tree-state.store"

const uiStore = useUiStore()
const fileTreeStore = useFileTreeStore()
const treeStateStore = useTreeStateStore()

const targetNode = computed(() => {
  const contextMenuState = contextMenuService.getContextMenuState()
  if (!contextMenuState?.targetPath) return null
  return fileTreeStore.nodesMap.get(contextMenuState.targetPath as string)
})

const isTargetSelected = computed(() => {
  const contextMenuState = contextMenuService.getContextMenuState()
  const path = contextMenuState?.targetPath
  if (!path) return false
  return treeStateStore.selectedPaths.has(path)
})

const ruleExists = computed(() => {
  const node = targetNode.value as any
  if (!node) return false
  const rule = normalizeRule(node.relPath, node.isDir)
  if (!rule) return false
  return hasRule(rule)
})

// Clamp position to viewport
const menuStyle = computed(() => {
  return contextMenuService.getMenuStyle()
})

function handleToggleSelection() {
  const contextMenuState = contextMenuService.getContextMenuState()
  const path = contextMenuState?.targetPath
  if (path) {
    treeStateStore.toggleNodeSelection(path, fileTreeStore.nodesMap as any)
  }
  contextMenuService.closeContextMenu()
  uiStore.closeContextMenu() // For backward compatibility
}

async function handleCopyPath() {
  const node = targetNode.value as any
  if (!node) {
    contextMenuService.closeContextMenu()
    uiStore.closeContextMenu() // For backward compatibility
    return
  }
  await navigator.clipboard.writeText(node.relPath)
  uiStateService.addToast("Путь скопирован!", "success")
  uiStore.addToast("Путь скопирован!", "success") // For backward compatibility
  contextMenuService.closeContextMenu()
  uiStore.closeContextMenu() // For backward compatibility
}

async function handleToggleIgnore() {
  const node = targetNode.value as any
  if (!node) return
  await toggleIgnore(node.relPath, node.isDir)
  contextMenuService.closeContextMenu()
  uiStore.closeContextMenu() // For backward compatibility
}

function addFolderRecursive() {
  const node = targetNode.value as any
  if (!node || !node.isDir) return
  treeStateStore.toggleNodeSelection(node.path, fileTreeStore.nodesMap as any)
  treeStateStore.toggleExpansionRecursive(
    node.path,
    fileTreeStore.nodesMap as any,
    true,
  )
  contextMenuService.closeContextMenu()
  uiStore.closeContextMenu() // For backward compatibility
}

function closeOnScroll() {
  if (contextMenuService.getContextMenuState()?.isOpen) {
    contextMenuService.closeContextMenu()
  }
  if (uiStore.contextMenu?.isOpen) {
    uiStore.closeContextMenu()
  }
}

// Keyboard navigation
const itemRefs = ref<HTMLElement[] | null>(null)
let idx = 0
function onKeydown(e: KeyboardEvent) {
  const container = e.currentTarget as HTMLElement
  const items = Array.from(
    container.querySelectorAll('[role="menuitem"]'),
  ) as HTMLElement[]
  if (!items.length) return
  if (e.key === "Escape") {
    contextMenuService.closeContextMenu()
    uiStore.closeContextMenu() // For backward compatibility
    return
  }
  if (e.key === "ArrowDown") {
    idx = (idx + 1) % items.length
    items[idx].focus()
  } else if (e.key === "ArrowUp") {
    idx = (idx - 1 + items.length) % items.length
    items[idx].focus()
  } else if (e.key === "Enter") {
    items[idx].click()
  }
}

onMounted(() => {
  window.addEventListener("scroll", closeOnScroll, true)
  window.addEventListener("wheel", closeOnScroll, true)
})
onUnmounted(() => {
  window.removeEventListener("scroll", closeOnScroll, true)
  window.removeEventListener("wheel", closeOnScroll, true)
})
</script>