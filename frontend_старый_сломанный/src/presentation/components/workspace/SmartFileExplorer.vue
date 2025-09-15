<template>
  <div class="h-full flex flex-col bg-gray-800">
    <!-- Smart Filters Header -->
    <div class="p-3 border-b border-gray-700">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-sm font-medium text-gray-200">File Explorer</h3>
        <div class="flex items-center space-x-1">
          <IconButton
            icon="AdjustmentsHorizontalIcon"
            size="sm"
            tooltip="Filter Options"
            @click="showFilterPanel = !showFilterPanel"
          />
          <IconButton
            icon="ViewColumnsIcon"
            size="sm"
            tooltip="View Options"
            @click="toggleViewMode"
          />
        </div>
      </div>
      
      <!-- Search Bar -->
      <div class="relative mb-2">
        <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search files..."
          class="w-full pl-10 pr-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-sm text-gray-200 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          @keydown.escape="searchQuery = ''"
        />
        <button
          v-if="searchQuery"
          class="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-200"
          @click="searchQuery = ''"
        >
          <XMarkIcon class="w-4 h-4" />
        </button>
      </div>
      
      <!-- Quick Filter Buttons -->
      <div class="flex flex-wrap gap-1">
        <button
          v-for="filter in quickFilters"
          :key="filter.id"
          class="px-2 py-1 text-xs rounded-md border transition-colors"
          :class="[
            activeFilters.includes(filter.id)
              ? 'bg-blue-600 border-blue-500 text-white'
              : 'bg-gray-700 border-gray-600 text-gray-300 hover:bg-gray-600'
          ]"
          @click="toggleFilter(filter.id)"
        >
          <component :is="filter.icon" class="w-3 h-3 inline mr-1" />
          {{ filter.label }}
          <span v-if="filter.count > 0" class="ml-1 opacity-75">({{ filter.count }})</span>
        </button>
      </div>
    </div>
    
    <!-- Advanced Filter Panel (Collapsible) -->
    <div 
      v-if="showFilterPanel"
      class="p-3 bg-gray-800 border-b border-gray-700 text-sm animate-slide-down"
    >
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="block text-xs font-medium text-gray-300 mb-1">File Types</label>
          <div class="space-y-1">
            <label v-for="type in fileTypes" :key="type.ext" class="flex items-center">
              <input
                v-model="selectedFileTypes"
                :value="type.ext"
                type="checkbox"
                class="w-3 h-3 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
              />
              <span class="ml-2 text-gray-300">.{{ type.ext }}</span>
              <span class="ml-auto text-xs text-gray-500">({{ type.count }})</span>
            </label>
          </div>
        </div>
        
        <div>
          <label class="block text-xs font-medium text-gray-300 mb-1">Size Range</label>
          <select 
            v-model="sizeFilter"
            class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-1 text-gray-200 text-xs"
          >
            <option value="">All sizes</option>
            <option value="small">Small (&lt; 10KB)</option>
            <option value="medium">Medium (10KB - 100KB)</option>
            <option value="large">Large (&gt; 100KB)</option>
          </select>
          
          <label class="block text-xs font-medium text-gray-300 mb-1 mt-2">Modified</label>
          <select 
            v-model="modifiedFilter"
            class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-1 text-gray-200 text-xs"
          >
            <option value="">Any time</option>
            <option value="today">Today</option>
            <option value="week">This week</option>
            <option value="month">This month</option>
          </select>
        </div>
      </div>
    </div>
    
    <!-- Batch Selection Tools -->
    <div 
      v-if="selectedFiles.length > 0"
      class="p-2 bg-blue-900 bg-opacity-20 border-b border-blue-700 flex items-center justify-between text-sm"
    >
      <div class="flex items-center space-x-2 text-blue-300">
        <CheckCircleIcon class="w-4 h-4" />
        <span>{{ selectedFiles.length }} files selected</span>
      </div>
      
      <div class="flex items-center space-x-1">
        <button
          class="px-2 py-1 bg-blue-600 text-white rounded text-xs hover:bg-blue-700 transition-colors"
          @click="addSelectedToContext"
        >
          Add to Context
        </button>
        <button
          class="px-2 py-1 bg-gray-600 text-gray-200 rounded text-xs hover:bg-gray-500 transition-colors"
          @click="clearSelection"
        >
          Clear
        </button>
      </div>
    </div>
    
    <!-- File Tree -->
    <div class="flex-1 overflow-auto">
      <div class="p-2">
        <!-- Statistics Bar -->
        <div class="flex items-center justify-between text-xs text-gray-400 mb-2 px-1">
          <span>{{ filteredFiles.length }} of {{ totalFiles }} files</span>
          <div class="flex items-center space-x-3">
            <span>{{ selectedFiles.length }} selected</span>
            <button
              v-if="filteredFiles.length > 0"
              class="text-blue-400 hover:text-blue-300"
              @click="selectAllVisible"
            >
              Select All
            </button>
          </div>
        </div>
        
        <!-- File List -->
        <div class="space-y-0.5">
          <div
            v-for="file in paginatedFiles"
            :key="file.path"
            class="flex items-center p-2 rounded-lg hover:bg-gray-700 cursor-pointer transition-colors group"
            :class="{
              'bg-blue-900 bg-opacity-30 border border-blue-700': isSelected(file.path),
              'bg-gray-700': !isSelected(file.path) && isHighlighted(file.path)
            }"
            @click="toggleFileSelection(file.path, $event)"
            @dblclick="openFilePreview(file)"
          >
            <!-- Selection Checkbox -->
            <div class="flex-shrink-0 mr-2">
              <input
                :checked="isSelected(file.path)"
                type="checkbox"
                class="w-3 h-3 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
                @click.stop
                @change="toggleFileSelection(file.path)"
              />
            </div>
            
            <!-- File Icon -->
            <div class="flex-shrink-0 mr-2">
              <component 
                :is="getFileIcon(file.extension)" 
                class="w-4 h-4"
                :class="getFileIconColor(file.extension)"
              />
            </div>
            
            <!-- File Info -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center space-x-2">
                <span 
                  class="text-sm text-gray-200 truncate"
                  :class="{ 'font-medium': isSelected(file.path) }"
                >
                  {{ file.name }}
                </span>
                
                <!-- File Badges -->
                <div class="flex items-center space-x-1">
                  <span
                    v-if="file.isModified"
                    class="px-1 py-0.5 bg-yellow-600 text-yellow-100 rounded text-xs"
                  >
                    M
                  </span>
                  <span
                    v-if="file.isNew"
                    class="px-1 py-0.5 bg-green-600 text-green-100 rounded text-xs"
                  >
                    N
                  </span>
                  <span
                    v-if="file.hasErrors"
                    class="px-1 py-0.5 bg-red-600 text-red-100 rounded text-xs"
                  >
                    !
                  </span>
                </div>
              </div>
              
              <div class="flex items-center space-x-2 text-xs text-gray-400 mt-0.5">
                <span>{{ file.relativePath }}</span>
                <span>•</span>
                <span>{{ formatFileSize(file.size) }}</span>
                <span v-if="file.lastModified">•</span>
                <span v-if="file.lastModified">{{ formatDate(file.lastModified) }}</span>
              </div>
            </div>
            
            <!-- Quick Actions -->
            <div class="flex-shrink-0 ml-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <div class="flex items-center space-x-1">
                <IconButton
                  icon="EyeIcon"
                  size="sm"
                  tooltip="Quick Preview"
                  @click.stop="openFilePreview(file)"
                />
                <IconButton
                  v-if="isSelected(file.path)"
                  icon="MinusIcon"
                  size="sm"
                  tooltip="Remove from Context"
                  @click.stop="removeFromContext(file.path)"
                />
                <IconButton
                  v-else
                  icon="PlusIcon"
                  size="sm"
                  tooltip="Add to Context"
                  @click.stop="addToContext(file.path)"
                />
              </div>
            </div>
          </div>
        </div>
        
        <!-- Load More / Pagination -->
        <div 
          v-if="filteredFiles.length > paginatedFiles.length"
          class="text-center p-4"
        >
          <button
            class="px-4 py-2 bg-gray-700 text-gray-200 rounded-lg hover:bg-gray-600 transition-colors text-sm"
            @click="loadMore"
          >
            Load More ({{ filteredFiles.length - paginatedFiles.length }} remaining)
          </button>
        </div>
        
        <!-- Empty State -->
        <div 
          v-if="filteredFiles.length === 0 && !isLoading"
          class="text-center py-8 text-gray-500"
        >
          <FolderIcon class="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>No files match your filters</p>
          <button
            class="mt-2 text-blue-400 hover:text-blue-300 text-sm"
            @click="clearAllFilters"
          >
            Clear all filters
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useFileTreeStore } from '@/stores/file-tree.store'
import { useContextBuilderStore } from '@/stores/context-builder.store'
import { useUiStore } from '@/stores/ui.store'

// Icons
import {
  MagnifyingGlassIcon,
  XMarkIcon,
  CheckCircleIcon,
  FolderIcon
} from '@heroicons/vue/24/outline'

import {
  DocumentIcon,
  CodeBracketIcon,
  PhotoIcon,
  FilmIcon,
  MusicalNoteIcon
} from '@heroicons/vue/24/solid'

import IconButton from '@/presentation/components/shared/IconButton.vue'

// Stores
const fileTreeStore = useFileTreeStore()
const contextStore = useContextBuilderStore()
const uiStore = useUiStore()

// Component State
const searchQuery = ref('')
const selectedFiles = ref<string[]>([])
const showFilterPanel = ref(false)
const viewMode = ref<'list' | 'grid'>('list')
const currentPage = ref(1)
const pageSize = ref(50)
const isLoading = ref(false)

// Filters
const activeFilters = ref<string[]>([])
const selectedFileTypes = ref<string[]>([])
const sizeFilter = ref('')
const modifiedFilter = ref('')

// Quick Filters Configuration
const quickFilters = ref([
  { id: 'recent', label: 'Recent', icon: 'ClockIcon', count: 0 },
  { id: 'modified', label: 'Modified', icon: 'PencilIcon', count: 0 },
  { id: 'large', label: 'Large Files', icon: 'DocumentIcon', count: 0 },
  { id: 'code', label: 'Code', icon: 'CodeBracketIcon', count: 0 },
  { id: 'config', label: 'Config', icon: 'Cog6ToothIcon', count: 0 }
])

// File Types
const fileTypes = ref([
  { ext: 'ts', count: 0 },
  { ext: 'js', count: 0 },
  { ext: 'vue', count: 0 },
  { ext: 'json', count: 0 },
  { ext: 'md', count: 0 },
  { ext: 'css', count: 0 },
  { ext: 'html', count: 0 }
])

// Computed Properties
const totalFiles = computed(() => fileTreeStore.totalFiles)

const filteredFiles = computed(() => {
  // Mock data for now - in real implementation, this would filter fileTreeStore.allFiles
  const mockFiles = Array.from({ length: 150 }, (_, i) => ({
    path: `/src/components/file${i}.ts`,
    name: `file${i}.ts`,
    relativePath: `src/components/file${i}.ts`,
    extension: ['ts', 'js', 'vue', 'json', 'md'][i % 5],
    size: Math.floor(Math.random() * 100000),
    lastModified: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000),
    isModified: Math.random() > 0.8,
    isNew: Math.random() > 0.9,
    hasErrors: Math.random() > 0.95
  }))
  
  return mockFiles.filter(file => {
    // Search filter
    if (searchQuery.value && !file.name.toLowerCase().includes(searchQuery.value.toLowerCase())) {
      return false
    }
    
    // File type filter
    if (selectedFileTypes.value.length > 0 && !selectedFileTypes.value.includes(file.extension)) {
      return false
    }
    
    // Size filter
    if (sizeFilter.value) {
      const size = file.size
      switch (sizeFilter.value) {
        case 'small':
          if (size >= 10000) return false
          break
        case 'medium':
          if (size < 10000 || size > 100000) return false
          break
        case 'large':
          if (size <= 100000) return false
          break
      }
    }
    
    // Quick filters
    if (activeFilters.value.includes('recent')) {
      const dayAgo = new Date(Date.now() - 24 * 60 * 60 * 1000)
      if (file.lastModified < dayAgo) return false
    }
    
    if (activeFilters.value.includes('modified') && !file.isModified) {
      return false
    }
    
    if (activeFilters.value.includes('large') && file.size < 50000) {
      return false
    }
    
    if (activeFilters.value.includes('code')) {
      const codeExts = ['ts', 'js', 'vue', 'jsx', 'tsx']
      if (!codeExts.includes(file.extension)) return false
    }
    
    if (activeFilters.value.includes('config')) {
      const configExts = ['json', 'yaml', 'yml', 'toml', 'ini']
      if (!configExts.includes(file.extension)) return false
    }
    
    return true
  })
})

const paginatedFiles = computed(() => {
  const start = 0
  const end = currentPage.value * pageSize.value
  return filteredFiles.value.slice(start, end)
})

// Methods
const toggleFilter = (filterId: string) => {
  const index = activeFilters.value.indexOf(filterId)
  if (index > -1) {
    activeFilters.value.splice(index, 1)
  } else {
    activeFilters.value.push(filterId)
  }
}

const toggleViewMode = () => {
  viewMode.value = viewMode.value === 'list' ? 'grid' : 'list'
}

const isSelected = (filePath: string) => {
  return selectedFiles.value.includes(filePath) || contextStore.selectedFiles.includes(filePath)
}

const isHighlighted = (_: string) => {
  // Highlight files that match search or are suggested
  return false // Placeholder
}

const toggleFileSelection = (filePath: string, event?: MouseEvent) => {
  if (event?.shiftKey && selectedFiles.value.length > 0) {
    // Handle shift+click for range selection
    const lastSelected = selectedFiles.value[selectedFiles.value.length - 1]
    selectRange(lastSelected, filePath)
  } else if (event?.ctrlKey || event?.metaKey) {
    // Handle ctrl/cmd+click for multi-selection
    const index = selectedFiles.value.indexOf(filePath)
    if (index > -1) {
      selectedFiles.value.splice(index, 1)
    } else {
      selectedFiles.value.push(filePath)
    }
  } else {
    // Single selection
    const index = selectedFiles.value.indexOf(filePath)
    if (index > -1) {
      selectedFiles.value.splice(index, 1)
    } else {
      selectedFiles.value.push(filePath)
    }
  }
}

const selectRange = (start: string, end: string) => {
  const startIndex = filteredFiles.value.findIndex(f => f.path === start)
  const endIndex = filteredFiles.value.findIndex(f => f.path === end)
  
  if (startIndex !== -1 && endIndex !== -1) {
    const min = Math.min(startIndex, endIndex)
    const max = Math.max(startIndex, endIndex)
    
    for (let i = min; i <= max; i++) {
      const filePath = filteredFiles.value[i].path
      if (!selectedFiles.value.includes(filePath)) {
        selectedFiles.value.push(filePath)
      }
    }
  }
}

const selectAllVisible = () => {
  filteredFiles.value.forEach(file => {
    if (!selectedFiles.value.includes(file.path)) {
      selectedFiles.value.push(file.path)
    }
  })
}

const clearSelection = () => {
  selectedFiles.value = []
}

const addSelectedToContext = () => {
  selectedFiles.value.forEach(filePath => {
    contextStore.addSelectedFile(filePath)
  })
  clearSelection()
}

const addToContext = (filePath: string) => {
  contextStore.addSelectedFile(filePath)
}

const removeFromContext = (filePath: string) => {
  contextStore.removeSelectedFile(filePath)
}

interface FileItem {
  path: string;
  name: string;
  relativePath: string;
  extension: string;
  size: number;
  lastModified: Date;
  isModified: boolean;
  isNew: boolean;
  hasErrors: boolean;
}

const openFilePreview = (file: FileItem) => {
  uiStore.showQuickLook({
    rootDir: '',
    path: file.path,
    type: 'file'
  })
}

const loadMore = () => {
  currentPage.value++
}

const clearAllFilters = () => {
  searchQuery.value = ''
  activeFilters.value = []
  selectedFileTypes.value = []
  sizeFilter.value = ''
  modifiedFilter.value = ''
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
    case 'mp4':
    case 'avi':
    case 'mov':
      return FilmIcon
    case 'mp3':
    case 'wav':
    case 'flac':
      return MusicalNoteIcon
    default:
      return DocumentIcon
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
    case 'md':
      return 'text-gray-400'
    case 'css':
      return 'text-purple-400'
    case 'html':
      return 'text-red-400'
    default:
      return 'text-gray-400'
  }
}

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const formatDate = (date: Date) => {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (days === 0) return 'Today'
  if (days === 1) return 'Yesterday'
  if (days < 7) return `${days} days ago`
  if (days < 30) return `${Math.floor(days / 7)} weeks ago`
  return date.toLocaleDateString()
}

// Watch for changes in context store selection
watch(
  () => contextStore.selectedFiles,
  (newSelection) => {
    // Sync with context store
    selectedFiles.value = [...newSelection]
  },
  { immediate: true }
)
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

/* Custom scrollbar */
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