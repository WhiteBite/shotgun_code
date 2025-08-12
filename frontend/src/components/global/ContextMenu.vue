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
        :style="{ top: `${uiStore.contextMenu.y}px`, left: `${uiStore.contextMenu.x}px` }"
        role="menu"
    >
      <div class="py-1" role="none">
        <a href="#" @click.prevent="handleToggleSelection" class="text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700" role="menuitem">
          {{ isTargetSelected ? 'Убрать из контекста' : 'Добавить в контекст' }}
        </a>
        <a href="#" @click.prevent="handleCopyPath" class="text-gray-200 block px-4 py-2 text-sm hover:bg-gray-700" role="menuitem">
          Копировать относительный путь
        </a>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue';
import { useUiStore } from '@/stores/ui.store';
import { useContextStore } from '@/stores/context.store';
import { ClipboardWriteText } from '@wails/runtime';

const uiStore = useUiStore();
const contextStore = useContextStore();

const targetNode = computed(() => {
  if (!uiStore.contextMenu.targetPath) return null;
  return contextStore.nodesMap.get(uiStore.contextMenu.targetPath);
});

const isTargetSelected = computed(() => {
  return targetNode.value?.selected === 'on' || targetNode.value?.selected === 'partial';
});

function handleToggleSelection() {
  if (uiStore.contextMenu.targetPath) {
    // contextStore.toggleNodeSelection(uiStore.contextMenu.targetPath);
  }
  uiStore.closeContextMenu();
}

async function handleCopyPath() {
  if (targetNode.value) {
    try {
      await ClipboardWriteText(targetNode.value.relPath);
      uiStore.addToast('Путь скопирован!', 'success');
    } catch (err) {
      uiStore.addToast('Не удалось скопировать путь', 'error');
    }
  }
  uiStore.closeContextMenu();
}

const handleClickOutside = () => {
  if (uiStore.contextMenu.visible) {
    uiStore.closeContextMenu();
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
  document.addEventListener('contextmenu', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
  document.removeEventListener('contextmenu', handleClickOutside);
});
</script>