<template>
  <transition
    enter-active-class="transition ease-out duration-100"
    enter-from-class="transform opacity-0 scale-95"
    enter-to-class="transform opacity-100 scale-100"
    leave-active-class="transition ease-in duration-75"
    leave-from-class="transform opacity-100 scale-100"
    leave-to-class="transform opacity-0 scale-95"
  >
    <div
      v-if="uiStore.contextMenu.visible"
      class="fixed z-50 w-56 rounded-md shadow-lg bg-gray-800 ring-1 ring-black ring-opacity-5 border border-gray-600 focus:outline-none"
      :style="{
        top: `${uiStore.contextMenu.y}px`,
        left: `${uiStore.contextMenu.x}px`,
      }"
      role="menu"
    >
      <div class="py-1" role="none">
        <a
          href="#"
          @click.prevent="handleToggleSelection"
          class="text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700"
          role="menuitem"
        >
          {{ isTargetSelected ? "Убрать из контекста" : "Добавить в контекст" }}
        </a>
        <a
          href="#"
          @click.prevent="handleQuickLook"
          class="text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700"
          role="menuitem"
          v-if="targetNode && !targetNode.isDir && !targetNode.isIgnored"
        >
          Быстрый предпросмотр
        </a>
        <a
          href="#"
          @click.prevent="handleCopyPath"
          class="text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700"
          role="menuitem"
        >
          Копировать относительный путь
        </a>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from "vue";
import { useUiStore } from "@/stores/ui.store";
import { useContextStore } from "@/stores/context.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useProjectStore } from "@/stores/project.store";

const uiStore = useUiStore();
const contextStore = useContextStore();
const treeStateStore = useTreeStateStore();
const projectStore = useProjectStore();

const targetNode = computed(() => {
  if (!uiStore.contextMenu.targetPath) return null;
  return contextStore.nodesMap.get(uiStore.contextMenu.targetPath);
});

const isTargetSelected = computed(() => {
  const path = uiStore.contextMenu.targetPath;
  if (!path) return false;
  return treeStateStore.selectedPaths.has(path);
});

function handleToggleSelection() {
  if (uiStore.contextMenu.targetPath) {
    treeStateStore.toggleSelection(uiStore.contextMenu.targetPath);
  }
  uiStore.closeContextMenu();
}

function handleQuickLook() {
  const node = targetNode.value;
  const rootDir = projectStore.currentProject?.path || "";
  if (!node || node.isDir || node.isIgnored || !rootDir) {
    uiStore.closeContextMenu();
    return;
  }
  const fakeEvent = new MouseEvent("click", {
    clientX: uiStore.contextMenu.x,
    clientY: uiStore.contextMenu.y,
  });
  uiStore.showQuickLook({
    rootDir,
    path: node.relPath,
    type: "fs",
    event: fakeEvent,
    isPinned: true,
  });
  uiStore.closeContextMenu();
}

async function handleCopyPath() {
  if (!targetNode.value) {
    uiStore.closeContextMenu();
    return;
  }
  const text = targetNode.value.relPath;
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
    } else {
      const ta = document.createElement("textarea");
      ta.value = text;
      ta.style.position = "fixed";
      ta.style.left = "-9999px";
      document.body.appendChild(ta);
      ta.select();
      document.execCommand("copy");
      document.body.removeChild(ta);
    }
    uiStore.addToast("Путь скопирован!", "success");
  } catch {
    uiStore.addToast("Не удалось скопировать путь", "error");
  }
  uiStore.closeContextMenu();
}

const handleClickOutside = () => {
  if (uiStore.contextMenu.visible) {
    uiStore.closeContextMenu();
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
  document.addEventListener("contextmenu", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
  document.removeEventListener("contextmenu", handleClickOutside);
});
</script>
