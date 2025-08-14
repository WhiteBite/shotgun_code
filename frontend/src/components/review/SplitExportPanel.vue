<template>
  <aside
    class="w-80 bg-gray-800/60 p-3 border-l border-gray-700 flex flex-col flex-shrink-0"
  >
    <h3 class="font-semibold text-sm mb-3 text-gray-200">
      Разделение и экспорт
    </h3>

    <!-- Настройка стратегии -->
    <div class="p-3 bg-gray-900/50 rounded-md border border-gray-700">
      <label class="block text-xs text-gray-400 mb-1">Стратегия:</label>
      <select
        v-model="reviewStore.splitStrategy"
        class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm focus:outline-none"
      >
        <option value="file">По файлам</option>
        <option value="hunk">По фрагментам (hunks)</option>
        <option value="lines">~N строк</option>
      </select>

      <div v-if="reviewStore.splitStrategy === 'lines'" class="mt-2">
        <label class="block text-xs text-gray-400 mb-1">Лимит строк:</label>
        <input
          type="number"
          v-model="reviewStore.splitLineLimit"
          class="w-full bg-gray-700 border-gray-600 rounded-md px-2 py-1 text-sm"
        />
      </div>

      <button
        @click="reviewStore.generatePatches()"
        class="w-full mt-3 py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold text-sm"
      >
        Сгенерировать части
      </button>
    </div>

    <!-- Список частей -->
    <div class="flex-grow min-h-0 mt-4 overflow-y-auto">
      <p
        class="text-xs text-gray-500 mb-2"
        v-if="reviewStore.generatedPatches.length === 0"
      >
        Части появятся здесь после генерации.
      </p>
      <div v-else class="space-y-2">
        <div
          v-for="patch in reviewStore.generatedPatches"
          :key="patch.id"
          class="p-2 bg-gray-700/60 rounded-md"
        >
          <div class="flex justify-between items-center">
            <span class="font-semibold text-sm text-white"
              >Часть #{{ patch.id }}</span
            >
            <div class="flex gap-1">
              <button
                class="px-2 py-0.5 text-xs bg-gray-600 hover:bg-gray-500 rounded-md"
                title="Копировать"
              >
                Copy
              </button>
              <button
                class="px-2 py-0.5 text-xs bg-gray-600 hover:bg-gray-500 rounded-md"
                title="Сохранить"
              >
                Save
              </button>
              <button
                class="px-2 py-0.5 text-xs bg-green-600 hover:bg-green-500 rounded-md"
                title="Применить"
              >
                Apply
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useReviewStore } from "@/stores/review.store";
const reviewStore = useReviewStore();
</script>
