<template>
  <div class="flex h-full">
    <!-- Левая панель: Список файлов -->
    <FileListPanel />

    <!-- Центральная область: Просмотр дифа -->
    <div class="flex-1 p-4">
      <DiffViewer />
    </div>

    <!-- Правая панель: Разделение и экспорт -->
    <SplitExportPanel />

    <!-- Нижняя липкая панель с действиями -->
    <div class="absolute bottom-0 left-0 right-0 h-16 bg-gray-900/80 backdrop-blur-sm border-t border-gray-700 flex items-center justify-between px-4 z-10">
      <button @click="$router.push('/workspace')" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-md font-semibold flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline></svg>
        Назад к задаче
      </button>
      <div class="flex items-center gap-3">
        <button class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-md font-semibold">Перегенерировать</button>
        <button class="px-5 py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-bold text-white">Применить выбранное</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import FileListPanel from '../components/review/FileListPanel.vue';
import DiffViewer from '../components/review/DiffViewer.vue';
import SplitExportPanel from '../components/review/SplitExportPanel.vue';
</script>