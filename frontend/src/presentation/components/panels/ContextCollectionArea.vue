<template>
  <div class="h-full flex flex-col">
    <!-- Context Collection Header -->
    <div class="p-4 border-b border-gray-700">
      <div class="flex items-center justify-between mb-2">
        <h2 class="text-lg font-semibold text-gray-100">Context Collection</h2>
        <div class="flex items-center space-x-2">
          <button
            class="px-3 py-1 text-xs rounded-lg transition-colors"
            :class="[
              activeView === 'explorer' 
                ? 'bg-blue-600 text-white' 
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            ]"
            @click="activeView = 'explorer'"
          >
            Explorer
          </button>
          <button
            class="px-3 py-1 text-xs rounded-lg transition-colors"
            :class="[
              activeView === 'builder' 
                ? 'bg-blue-600 text-white' 
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            ]"
            @click="activeView = 'builder'"
          >
            Builder
          </button>
        </div>
      </div>
      <p class="text-sm text-gray-400">Select and organize files to build context for your project</p>
    </div>
    
    <!-- Dynamic Content Area -->
    <div class="flex-1 flex flex-col min-h-0">
      <!-- File Explorer View -->
      <div v-if="activeView === 'explorer'" class="flex-1 flex flex-col min-h-0">
        <SmartFileExplorer />
      </div>
      
      <!-- Context Builder View -->
      <div v-else-if="activeView === 'builder'" class="flex-1 flex flex-col min-h-0">
        <ContextBuilder />
      </div>
    </div>
    
    <!-- Context Summary (Always Visible) -->
    <div class="p-4 border-t border-gray-700">
      <ContextSummary />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import SmartFileExplorer from '@/presentation/components/workspace/SmartFileExplorer.vue'
import ContextBuilder from '@/presentation/components/workspace/ContextBuilder.vue'
import ContextSummary from '@/presentation/components/workspace/ContextSummary.vue'

const contextStore = useContextBuilderStore()
const activeView = ref<'explorer' | 'builder'>('explorer')
</script>