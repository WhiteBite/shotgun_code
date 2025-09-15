<template>
  <div class="h-full flex flex-col bg-gray-800">
    <!-- Header -->
    <div class="p-4 border-b border-gray-700">
      <div class="flex items-center justify-between mb-2">
        <h3 class="text-lg font-medium text-gray-200">Context Builder</h3>
        <div class="flex items-center space-x-2">
          <div class="flex items-center space-x-1 text-sm text-gray-400">
            <span>{{ contextStore.selectedFilesList?.length || 0 }} files</span>
            <span>•</span>
            <span>{{ contextStore.contextMetrics.tokenCount }} tokens</span>
          </div>
          <IconButton
            icon="Cog6ToothIcon"
            size="sm"
            tooltip="Context Settings"
            @click="showSettings = !showSettings"
          />
        </div>
      </div>
      
      <!-- Context Name and Description -->
      <div class="space-y-2">
        <input
          v-model="contextName"
          type="text"
          placeholder="Context name (optional)"
          class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-sm text-gray-200 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <textarea
          v-model="contextDescription"
          placeholder="Describe what this context is for..."
          rows="2"
          class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-sm text-gray-200 placeholder-gray-400 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>
    </div>
    
    <!-- Settings Panel (Collapsible) -->
    <div 
      v-if="showSettings"
      class="p-4 bg-gray-800 border-b border-gray-700 space-y-3 animate-slide-down"
    >
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="flex items-center space-x-2">
            <input
              v-model="settings.includeComments"
              type="checkbox"
              class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
            />
            <span class="text-sm text-gray-300">Include comments</span>
          </label>
        </div>
        
        <div>
          <label class="flex items-center space-x-2">
            <input
              v-model="settings.includeImports"
              type="checkbox"
              class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
            />
            <span class="text-sm text-gray-300">Include imports</span>
          </label>
        </div>
        
        <div>
          <label class="flex items-center space-x-2">
            <input
              v-model="settings.includeLineNumbers"
              type="checkbox"
              class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
            />
            <span class="text-sm text-gray-300">Include line numbers</span>
          </label>
        </div>
        
        <div>
          <label class="flex items-center space-x-2">
            <input
              v-model="settings.minifyWhitespace"
              type="checkbox"
              class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
            />
            <span class="text-sm text-gray-300">Minify whitespace</span>
          </label>
        </div>
      </div>
      
      <div>
        <label class="block text-sm text-gray-300 mb-1">Max tokens per file</label>
        <input
          v-model.number="settings.maxTokensPerFile"
          type="number"
          min="100"
          max="5000"
          class="w-full px-3 py-1 bg-gray-700 border border-gray-600 rounded text-sm text-gray-200"
        />
      </div>
    </div>
    
    <!-- Drop Zone -->
    <div 
      class="flex-1 flex flex-col min-h-0"
      @dragover.prevent
      @drop="handleDrop"
      @dragenter.prevent="isDragOver = true"
      @dragleave.prevent="isDragOver = false"
    >
      <!-- Selected Files List -->
      <div class="flex-1 overflow-auto">
        <div 
          v-if="!contextStore.selectedFilesList || contextStore.selectedFilesList.length === 0"
          class="h-full flex items-center justify-center"
          :class="{
            'bg-blue-900 bg-opacity-20 border-2 border-dashed border-blue-500': isDragOver,
            'border-2 border-dashed border-gray-600': !isDragOver
          }"
        >
          <div class="text-center p-8">
            <FolderPlusIcon class="w-16 h-16 mx-auto mb-4 text-gray-500" />
            <h4 class="text-lg font-medium text-gray-300 mb-2">No files selected</h4>
            <p class="text-sm text-gray-500 mb-4">
              Drag files here or select them from the file explorer
            </p>
            <div class="space-y-2 text-xs text-gray-600">
              <p>• Supports multiple file selection</p>
              <p>• Drag & drop from file explorer</p>
              <p>• Auto-preview with syntax highlighting</p>
            </div>
          </div>
        </div>
        
        <div v-else class="p-4 space-y-2">
          <!-- File Items with Drag and Drop -->
          <draggable
            :model-value="contextStore.selectedFilesList"
            item-key="path"
            group="context-files"
            :animation="200"
            class="space-y-2"
            @update:model-value="handleFilesUpdate"
            @start="isDragging = true"
            @end="isDragging = false"
          >
            <template #item="{ element: filePath }">
              <div
                class="bg-gray-700 rounded-lg border border-gray-600 transition-all duration-200"
                :class="{
                  'border-blue-500 shadow-lg': selectedPreviewFile === filePath,
                  'opacity-50': isDragging,
                  'transform scale-105': isDragging && selectedPreviewFile === filePath
                }"
              >
                <!-- File Header -->
                <div class="flex items-center p-3 cursor-move">
                  <!-- Drag Handle -->
                  <div class="mr-3 text-gray-400">
                    <Bars3Icon class="w-4 h-4" />
                  </div>
                  
                  <!-- File Icon and Name -->
                  <div class="flex items-center flex-1 min-w-0">
                    <component 
                      :is="getFileIcon(getFileExtension(filePath))" 
                      class="w-4 h-4 mr-2 flex-shrink-0"
                      :class="getFileIconColor(getFileExtension(filePath))"
                    />
                    <span class="text-sm font-medium text-gray-200 truncate">
                      {{ getFileName(filePath) }}
                    </span>
                    <span class="text-xs text-gray-400 ml-2 flex-shrink-0">
                      {{ getFileSize(filePath) }}
                    </span>
                  </div>
                  
                  <!-- Actions -->
                  <div class="flex items-center space-x-1 ml-3">
                    <IconButton
                      :icon="selectedPreviewFile === filePath ? 'EyeSlashIcon' : 'EyeIcon'"
                      size="sm"
                      :tooltip="selectedPreviewFile === filePath ? 'Hide Preview' : 'Show Preview'"
                      @click="togglePreview(filePath)"
                    />
                    <IconButton
                      icon="XMarkIcon"
                      size="sm"
                      tooltip="Remove from Context"
                      @click="removeFile(filePath)"
                    />
                  </div>
                </div>
                
                <!-- File Preview (Collapsible) -->
                <div 
                  v-if="selectedPreviewFile === filePath"
                  class="border-t border-gray-600 bg-gray-800"
                >
                  <div class="p-3">
                    <div class="flex items-center justify-between mb-2">
                      <div class="flex items-center space-x-2 text-xs text-gray-400">
                        <span>{{ getRelativePath(filePath) }}</span>
                        <span>•</span>
                        <span>{{ getTokenCount(filePath) }} tokens</span>
                        <span>•</span>
                        <span>Lines: {{ getLineCount(filePath) }}</span>
                      </div>
                      
                      <div class="flex items-center space-x-1">
                        <button
                          class="px-2 py-1 text-xs bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors"
                          @click="copyFileContent(filePath)"
                        >
                          Copy
                        </button>
                        <button
                          class="px-2 py-1 text-xs bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors"
                          @click="openInEditor(filePath)"
                        >
                          Edit
                        </button>
                      </div>
                    </div>
                    
                    <!-- Syntax Highlighted Preview -->
                    <div class="bg-black bg-opacity-30 rounded border border-gray-600 overflow-hidden">
                      <div class="max-h-64 overflow-auto">
                        <pre class="p-3 text-xs"><code class="language-typescript text-gray-300">{{ getFilePreview(filePath) }}</code></pre>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </template>
          </draggable>
        </div>
      </div>
      
      <!-- Build Actions -->
      <div class="p-4 border-t border-gray-700 bg-gray-800">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4 text-sm text-gray-400">
            <div class="flex items-center space-x-1">
              <DocumentIcon class="w-4 h-4" />
              <span>{{ contextStore.selectedFilesList?.length || 0 }} files</span>
            </div>
            <div class="flex items-center space-x-1">
              <CpuChipIcon class="w-4 h-4" />
              <span>~{{ estimatedTokens }} tokens</span>
            </div>
          </div>
          
          <div class="flex items-center space-x-2">
            <button
              class="px-3 py-1 text-sm bg-gray-700 text-gray-300 rounded-lg hover:bg-gray-600 transition-colors"
              @click="previewContext"
            >
              Preview
            </button>
            <button
              class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
              :disabled="!contextStore.canBuildContext"
              @click="buildContext"
            >
              {{ contextStore.buildStatus === 'building' ? 'Building...' : 'Build Context' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useProjectStore } from '@/stores/project.store'
import draggable from 'vuedraggable'

// Icons
import {
  FolderPlusIcon,
  Bars3Icon,
  DocumentIcon,
  CpuChipIcon
} from '@heroicons/vue/24/outline'

import {
  DocumentIcon as DocumentSolidIcon,
  CodeBracketIcon,
  PhotoIcon
} from '@heroicons/vue/24/solid'

import IconButton from '@/presentation/components/shared/IconButton.vue'
import { FileProcessingService } from '@/domain/services/FileProcessingService'
import { UIFormattingService } from '@/domain/services/UIFormattingService'

// Services
const fileProcessingService = new FileProcessingService()
const uiFormattingService = new UIFormattingService()

// Stores
const contextStore = useContextBuilderStore()
const projectStore = useProjectStore()

// Component State
const isDragOver = ref(false)
const isDragging = ref(false)
const selectedPreviewFile = ref<string | null>(null)
const showSettings = ref(false)
const contextName = ref('')
const contextDescription = ref('')

// Settings
const settings = ref({
  includeComments: true,
  includeImports: true,
  includeLineNumbers: false,
  minifyWhitespace: false,
  maxTokensPerFile: 2000
})

// Add this new method to handle updates from draggable
const handleFilesUpdate = (newFiles: string[]) => {
  contextStore.setSelectedFiles(newFiles)
}

const estimatedTokens = computed(() => {
  return contextStore.contextMetrics.tokenCount || 
         (contextStore.selectedFilesList?.length || 0) * 500 // Rough estimate
})

// Methods
const handleDrop = (event: DragEvent) => {
  event.preventDefault()
  isDragOver.value = false
  
  const files = event.dataTransfer?.files
  if (files) {
    Array.from(files).forEach(file => {
      contextStore.addSelectedFile(file.name)
    })
  }
}

const togglePreview = (filePath: string) => {
  selectedPreviewFile.value = selectedPreviewFile.value === filePath ? null : filePath
}

const removeFile = (filePath: string) => {
  contextStore.removeSelectedFile(filePath)
  if (selectedPreviewFile.value === filePath) {
    selectedPreviewFile.value = null
  }
}

const buildContext = async () => {
  if (!projectStore.currentProject?.path) return
  
  try {
    await contextStore.buildContextFromSelection(projectStore.currentProject.path)
  } catch (error) {
    console.error('Failed to build context:', error)
  }
}

const previewContext = () => {
  console.log('Preview context')
}

const copyFileContent = (filePath: string) => {
  const content = getFilePreview(filePath)
  navigator.clipboard.writeText(content)
}

const openInEditor = (filePath: string) => {
  console.log('Open in editor:', filePath)
}

// Helper functions using the new services
const getFileName = (filePath: string) => {
  return filePath.split('/').pop() || filePath
}

const getFileExtension = (filePath: string) => {
  return fileProcessingService.getFileExtension(filePath)
}

const getFileIcon = (extension: string) => {
  switch (extension) {
    case 'ts':
    case 'js':
    case 'jsx':
    case 'tsx':
    case 'vue':
      return CodeBracketIcon
    case 'png':
    case 'jpg':
    case 'jpeg':
    case 'gif':
    case 'svg':
      return PhotoIcon
    default:
      return DocumentSolidIcon
  }
}

const getFileIconColor = (extension: string) => {
  switch (extension) {
    case 'ts':
    case 'tsx':
      return 'text-blue-400'
    case 'js':
    case 'jsx':
      return 'text-yellow-400'
    case 'vue':
      return 'text-green-400'
    case 'json':
      return 'text-orange-400'
    default:
      return 'text-gray-400'
  }
}

const getRelativePath = (filePath: string) => {
  if (!projectStore.currentProject?.path) return filePath
  return fileProcessingService.getRelativePath(filePath, projectStore.currentProject.path)
}

const getFileSize = (filePath: string) => {
  // In a real implementation, this would return the actual file size
  // For now, we're using a placeholder to acknowledge the parameter
  return uiFormattingService.formatFileSize(filePath.length)
}

const getTokenCount = (filePath: string) => {
  // In a real implementation, this would calculate the actual token count
  // For now, we're using a placeholder to acknowledge the parameter
  return fileProcessingService.estimateTokenCount(filePath)
}

const getLineCount = (filePath: string) => {
  // In a real implementation, this would calculate the actual line count
  // For now, we're using a placeholder to acknowledge the parameter
  return fileProcessingService.countLines(filePath)
}

const getFilePreview = (filePath: string) => {
  // In a real implementation, this would return the actual file content
  // For now, we're using a placeholder to acknowledge the parameter
  const fileName = getFileName(filePath)
  return fileProcessingService.getFilePreview(`// ${fileName}
import { ref, computed } from 'vue'

export default {
  name: '${fileProcessingService.getFileNameWithoutExtension(fileName)}',
  setup() {
    const count = ref(0)
    
    const doubleCount = computed(() => count.value * 2)
    
    function increment() {
      count.value++
    }
    
    return {
      count,
      doubleCount,
      increment
    }
  }
}`)
}
</script>

<style scoped>
@keyframes slide-down {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-slide-down {
  animation: slide-down 0.2s ease-out;
}

.sortable-ghost {
  opacity: 0.5;
}

.sortable-chosen {
  transform: scale(1.02);
}

::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #4a5568;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #718096;
}
</style>