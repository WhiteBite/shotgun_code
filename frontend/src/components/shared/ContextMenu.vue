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
      v-if="uiStore.contextMenu?.isVisible"
      class="fixed z-50 w-56 rounded-md shadow-lg bg-gray-800 ring-1 ring-black ring-opacity-5 border border-gray-600 focus:outline-none"
      :style="menuStyle"
      role="menu"
      tabindex="0"
      @keydown.stop.prevent="onKeydown"
      @contextmenu.stop
      @click.stop
    >
      <div class="py-1" role="none">
        <button type="button" @click="handleToggleSelection" class="w-full text-left text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700" role="menuitem" ref="itemRefs">
          {{ isTargetSelected ? 'Убрать из контекста' : 'Добавить в контекст' }}
        </button>
        <button type="button" @click="handleCopyPath" class="w-full text-left text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700" role="menuitem" ref="itemRefs">
          Копировать относительный путь
        </button>
        <button
          v-if="targetNode?.isDir"
          type="button"
          @click="addFolderRecursive"
          class="w-full text-left text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700"
          role="menuitem"
          ref="itemRefs"
        >
          Добавить папку рекурсивно в контекст
        </button>
        <button type="button" @click="handleToggleIgnore" class="w-full text-left text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700" role="menuitem" ref="itemRefs">
          {{ ruleExists ? 'Убрать из игнора' : 'Игнорировать путь' }}
        </button>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, nextTick } from "vue";
import { useUiStore } from "@/stores/ui.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useIgnoreRules } from "@/composables/useIgnoreRules";

const uiStore = useUiStore();
const fileTreeStore = useFileTreeStore();
const treeStateStore = useTreeStateStore();
const { normalizeRule, hasRule, toggleIgnore } = useIgnoreRules();

const targetNode = computed(() => {
  if (!uiStore.contextMenu?.nodePath) return null;
  return fileTreeStore.nodesMap.get(uiStore.contextMenu.nodePath);
});
const isTargetSelected = computed(() => {
  const path = uiStore.contextMenu?.nodePath;
  if (!path) return false;
  return treeStateStore.selectedPaths.has(path);
});
const ruleExists = computed(() => {
  const node: any = targetNode.value;
  if (!node) return false;
  const rule = normalizeRule(node.relPath, node.isDir);
  if (!rule) return false;
  return hasRule(rule);
});

// Clamp position to viewport
const menuStyle = computed(() => {
  const cm = uiStore.contextMenu;
  if (!cm) return {};
  const margin = 8;
  const width = 224;
  const height = 180;
  const left = Math.min(Math.max(cm.x, margin), window.innerWidth - width - margin);
  const top = Math.min(Math.max(cm.y, margin), window.innerHeight - height - margin);
  return { left: left + "px", top: top + "px" };
});

function handleToggleSelection() {
  const path = uiStore.contextMenu?.nodePath;
  if (path) {
    treeStateStore.toggleNodeSelection(path, fileTreeStore.nodesMap);
  }
  uiStore.closeContextMenu();
}

async function handleCopyPath() {
  const node: any = targetNode.value;
  if (!node) { uiStore.closeContextMenu(); return; }
  await navigator.clipboard.writeText(node.relPath);
  uiStore.addToast("Путь скопирован!", "success");
  uiStore.closeContextMenu();
}

async function handleToggleIgnore() {
  const node: any = targetNode.value;
  if (!node) return;
  await toggleIgnore(node.relPath, node.isDir);
  uiStore.closeContextMenu();
}

function addFolderRecursive() {
  const node: any = targetNode.value;
  if (!node || !node.isDir) return;
  treeStateStore.toggleNodeSelection(node.path, fileTreeStore.nodesMap);
  treeStateStore.toggleExpansionRecursive(node.path, fileTreeStore.nodesMap, true);
  uiStore.closeContextMenu();
}

function closeOnScroll() {
  if (uiStore.contextMenu?.isVisible) uiStore.closeContextMenu();
}

// Keyboard navigation
const itemRefs = ref<HTMLElement[] | null>(null);
let idx = 0;
function onKeydown(e: KeyboardEvent) {
  const container = e.currentTarget as HTMLElement;
  const items = Array.from(container.querySelectorAll('[role="menuitem"]')) as HTMLElement[];
  if (!items.length) return;
  if (e.key === "Escape") { uiStore.closeContextMenu(); return; }
  if (e.key === "ArrowDown") {
    idx = (idx + 1) % items.length;
    items[idx].focus();
  } else if (e.key === "ArrowUp") {
    idx = (idx - 1 + items.length) % items.length;
    items[idx].focus();
  } else if (e.key === "Enter") {
    items[idx].click();
  }
}

onMounted(() => {
  window.addEventListener("scroll", closeOnScroll, true);
  window.addEventListener("wheel", closeOnScroll, true);
});
onUnmounted(() => {
  window.removeEventListener("scroll", closeOnScroll, true);
  window.removeEventListener("wheel", closeOnScroll, true);
});
</script>
