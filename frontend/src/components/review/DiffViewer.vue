<template>
  <div class="flex flex-col h-full bg-gray-900 rounded-lg border border-gray-700">
    <div class="flex items-center justify-between p-2 border-b border-gray-700 bg-gray-800/50 rounded-t-lg flex-shrink-0">
      <div class="flex items-center gap-2">
        <span class="text-sm font-semibold text-gray-300">Вид:</span>
        <button @click="reviewStore.viewMode = 'split'" :class="['px-2 py-1 text-xs rounded-md', reviewStore.viewMode === 'split' ? 'bg-blue-600 text-white' : 'bg-gray-700 hover:bg-gray-600']">Параллельно</button>
        <button @click="reviewStore.viewMode = 'unified'" :class="['px-2 py-1 text-xs rounded-md', reviewStore.viewMode === 'unified' ? 'bg-blue-600 text-white' : 'bg-gray-700 hover:bg-gray-600']">Общий</button>
      </div>
      <div class="flex items-center gap-3 text-sm">
        <label class="flex items-center gap-1.5 cursor-pointer"><input type="checkbox" class="form-checkbox" v-model="reviewStore.showLineNumbers"> Номера строк</label>
        <label class="flex items-center gap-1.5 cursor-pointer"><input type="checkbox" class="form-checkbox" v-model="reviewStore.ignoreWhitespace"> Игнорировать пробелы</label>
      </div>
    </div>

    <div class="flex-grow p-4 overflow-auto">
      <div v-for="file in reviewStore.parsedDiff" :key="file.filePath" class="mb-6 font-mono text-sm">
        <div class="p-2 bg-gray-800 border border-gray-700 rounded-t-md flex justify-between items-center">
          <div class="flex items-center gap-2">
            <input type="checkbox" :checked="reviewStore.selectedFiles.has(file.filePath)" @change="reviewStore.toggleFileSelection(file.filePath)" class="form-checkbox">
            <span class="font-bold text-white">{{ file.filePath }}</span>
          </div>
          <div class="text-xs">
            <span class="font-bold text-green-400">+{{ file.stats.added }}</span>
            <span class="font-bold text-red-400 ml-1">-{{ file.stats.removed }}</span>
          </div>
        </div>
        <div class="bg-gray-900/50 border-x border-b border-gray-700 rounded-b-md">
          <div v-for="(hunk, index) in file.hunks" :key="index" class="border-t border-gray-700/50">
            <div class="flex">
              <div class="w-10 flex-shrink-0 bg-gray-800/30 text-right p-1">
                <input type="checkbox" checked class="form-checkbox">
              </div>
              <div class="bg-gray-800/30 p-1 text-cyan-400 font-semibold flex-grow">
                {{ hunk.header }}
              </div>
            </div>
            <div v-for="(line, lineIndex) in hunk.lines" :key="lineIndex"
                 :class="['flex', getLineClass(line)]">
              <div class="w-10 flex-shrink-0 text-right px-2 text-gray-600 select-none">
                <!-- Line numbers placeholder -->
              </div>
              <div class="flex-grow pr-4">
                <span class="mr-2 select-none">{{ line.charAt(0) }}</span>
                <span class="whitespace-pre-wrap">{{ line.substring(1) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useReviewStore } from '@/stores/reviewStore';
const reviewStore = useReviewStore();

const getLineClass = (line: string) => {
  if (line.startsWith('+') && !line.startsWith('+++')) return 'bg-green-900/20 text-green-300';
  if (line.startsWith('-') && !line.startsWith('---')) return 'bg-red-900/20 text-red-300';
  return 'text-gray-400';
};
</script>