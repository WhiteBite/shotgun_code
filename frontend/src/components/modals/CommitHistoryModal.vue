<template>
  <transition name="modal-fade">
    <div v-if="isVisible" class="fixed inset-0 bg-black/60 z-40 flex items-center justify-center" @click.self="close">
      <div class="bg-gray-800 rounded-lg shadow-2xl w-full max-w-3xl max-h-[85vh] flex flex-col border border-gray-700">
        <header class="p-4 border-b border-gray-700 flex-shrink-0">
          <h2 class="text-xl font-semibold text-white">История коммитов</h2>
        </header>

        <main class="p-4 overflow-y-auto flex-grow">
          <!-- Commit list will be implemented here -->
          <div class="text-center text-gray-400">Список коммитов...</div>
        </main>

        <footer class="p-4 border-t border-gray-700 flex justify-end space-x-4 flex-shrink-0">
          <button @click="close" class="px-4 py-2 bg-gray-600 hover:bg-gray-500 rounded-md">Отмена</button>
          <button @click="apply" class="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-md">
            Применить
          </button>
        </footer>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
defineProps<{ isVisible: boolean }>();
const emit = defineEmits<{
  (_e: 'close'): void;
  (_e: 'apply', _hashes: string[]): void;
}>();

const close = () => emit('close');
const apply = () => emit('apply', []);
</script>

<style scoped>
.modal-fade-enter-active, .modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from, .modal-fade-leave-to {
  opacity: 0;
}
</style>