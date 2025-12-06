<template>
  <div class="select-none">
    <!-- Use virtual scroller for large lists -->
    <RecycleScroller
      v-if="visibleNodes.length > 100"
      :items="visibleNodes"
      :item-size="32"
      key-field="path"
      class="h-full"
      v-slot="{ item: node }"
    >
      <div
        :class="['tree-item', isSelected(node.path) ? 'tree-item-selected' : '']"
        :style="{ paddingLeft: `${node.depth * 12 + 4}px` }"
        @click="handleClick(node)"
      >
        <TreeNodeContent
          :node="node"
          :is-selected="isSelected(node.path)"
          :is-expanded="isExpanded(node.path)"
          :has-some-selected="node.isDir && hasSomeSelected(node.path)"
          :has-all-selected="node.isDir && hasAllSelected(node.path)"
          :files-count="node.isDir ? getFilesInDir(node.path).length : 0"
          @toggle-select="toggleSelect(node)"
        />
      </div>
    </RecycleScroller>

    <!-- Regular rendering for small lists -->
    <div v-else class="p-2">
      <template v-for="node in visibleNodes" :key="node.path">
        <div
          :class="['tree-item', isSelected(node.path) ? 'tree-item-selected' : '']"
          :style="{ paddingLeft: `${node.depth * 12 + 4}px` }"
          @click="handleClick(node)"
          @dblclick="handleDoubleClick(node)"
        >
          <!-- Expand/Collapse Icon (for directories) -->
          <div v-if="node.isDir" class="tree-expand-icon">
            <svg
              :class="['tree-expand-arrow', isExpanded(node.path) ? 'tree-expand-arrow-open' : '']"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
            </svg>
          </div>
          <div v-else class="w-4"></div>

          <!-- Checkbox -->
          <div
            :class="[
              'tree-checkbox',
              isSelected(node.path) || (node.isDir && hasSomeSelected(node.path)) ? 'tree-checkbox-checked' : ''
            ]"
            @click.stop="toggleSelect(node)"
          >
            <svg v-if="isSelected(node.path) || (node.isDir && hasAllSelected(node.path))" class="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
            <div v-else-if="node.isDir && hasSomeSelected(node.path)" class="w-2 h-0.5 bg-white"></div>
          </div>

          <!-- File/Folder Icon -->
          <div class="tree-icon">
            <svg v-if="node.isDir" class="tree-folder-icon" fill="currentColor" viewBox="0 0 20 20">
              <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
            </svg>
            <span v-else class="tree-file-icon">{{ getFileIcon(node.name) }}</span>
          </div>

          <!-- Name -->
          <span class="tree-name">{{ node.name }}</span>

          <!-- Preview button for files -->
          <button
            v-if="!node.isDir"
            @click.stop="emit('preview-file', node.path)"
            class="tree-preview-btn opacity-0 group-hover:opacity-100"
            :title="'Preview'"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
          </button>

          <!-- File count for dirs -->
          <span v-if="node.isDir" class="tree-count">
            {{ getFilesInDir(node.path).length }}
          </span>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getFileIcon } from '@/utils/fileIcons'
import { computed, defineComponent, h, ref } from 'vue'
import { RecycleScroller } from 'vue-virtual-scroller'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'

interface TreeNode {
  path: string
  name: string
  isDir: boolean
  depth: number
}

const props = defineProps<{
  files: string[]
  selectedPaths: Set<string>
}>()

const emit = defineEmits<{
  (e: 'toggle-select', path: string): void
  (e: 'select-folder', paths: string[]): void
  (e: 'preview-file', path: string): void
}>()

const expandedPaths = ref<Set<string>>(new Set())

// TreeNodeContent component for virtual scroller
const TreeNodeContent = defineComponent({
  props: {
    node: { type: Object as () => TreeNode, required: true },
    isSelected: Boolean,
    isExpanded: Boolean,
    hasSomeSelected: Boolean,
    hasAllSelected: Boolean,
    filesCount: Number
  },
  emits: ['toggle-select'],
  setup(props, { emit }) {
    return () => h('div', { class: 'flex items-center gap-2 flex-1' }, [
      // Expand icon
      props.node.isDir
        ? h('div', { class: 'flex-shrink-0 w-4 h-4 flex items-center justify-center' }, [
            h('svg', {
              class: ['w-3 h-3 transition-transform', props.isExpanded ? 'rotate-90' : ''].filter(Boolean).join(' '),
              fill: 'currentColor',
              viewBox: '0 0 20 20',
              innerHTML: '<path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />'
            })
          ])
        : h('div', { class: 'w-4' }),
      // Checkbox
      h('div', {
        class: [
          'flex-shrink-0 w-4 h-4 rounded border-2 flex items-center justify-center transition-colors',
          props.isSelected || props.hasSomeSelected
            ? 'bg-indigo-600 border-indigo-600'
            : 'border-gray-600 hover:border-gray-500'
        ].join(' '),
        onClick: (e: Event) => { e.stopPropagation(); emit('toggle-select') }
      }, props.isSelected || props.hasAllSelected
        ? [h('svg', { class: 'w-3 h-3 text-white', fill: 'currentColor', viewBox: '0 0 20 20', innerHTML: '<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />' })]
        : props.hasSomeSelected
          ? [h('div', { class: 'w-2 h-0.5 bg-white' })]
          : []
      ),
      // Icon
      h('div', { class: 'flex-shrink-0' },
        props.node.isDir
          ? [h('svg', { class: 'w-4 h-4 text-blue-400', fill: 'currentColor', viewBox: '0 0 20 20', innerHTML: '<path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />' })]
          : [h('span', { class: 'text-sm' }, getFileIcon(props.node.name))]
      ),
      // Name
      h('span', { class: 'flex-1 truncate text-sm' }, props.node.name),
      // Count
      props.node.isDir
        ? h('span', { class: 'text-xs text-gray-500' }, String(props.filesCount))
        : null
    ])
  }
})

// Safe check for selected
function isSelected(path: string): boolean {
  return props.selectedPaths?.has(path) ?? false
}

function isExpanded(path: string): boolean {
  return expandedPaths.value.has(path)
}

// Get all files in a directory
function getFilesInDir(dirPath: string): string[] {
  if (!props.files || !Array.isArray(props.files)) return []
  return props.files.filter(f => f.startsWith(dirPath + '/'))
}

function hasSomeSelected(dirPath: string): boolean {
  if (!props.selectedPaths) return false
  const filesInDir = getFilesInDir(dirPath)
  return filesInDir.some(f => props.selectedPaths.has(f))
}

function hasAllSelected(dirPath: string): boolean {
  if (!props.selectedPaths) return false
  const filesInDir = getFilesInDir(dirPath)
  return filesInDir.length > 0 && filesInDir.every(f => props.selectedPaths.has(f))
}

// Build tree structure
const visibleNodes = computed(() => {
  if (!props.files || !Array.isArray(props.files) || props.files.length === 0) return []

  const nodes: TreeNode[] = []
  const addedDirs = new Set<string>()

  // Sort files first to ensure proper order
  const sortedFiles = [...props.files].sort()

  for (const file of sortedFiles) {
    const parts = file.split('/')
    
    // Add directory nodes for this file's path
    let currentPath = ''
    let isVisible = true

    for (let i = 0; i < parts.length - 1; i++) {
      const parentPath = currentPath
      currentPath = currentPath ? `${currentPath}/${parts[i]}` : parts[i]

      // Check if parent is expanded (root level is always visible)
      if (parentPath && !expandedPaths.value.has(parentPath)) {
        isVisible = false
        break
      }

      // Add directory if not already added
      if (!addedDirs.has(currentPath)) {
        addedDirs.add(currentPath)
        nodes.push({
          path: currentPath,
          name: parts[i],
          isDir: true,
          depth: i
        })
      }
    }

    // Add file if visible
    if (isVisible) {
      const parentPath = parts.slice(0, -1).join('/')
      if (!parentPath || expandedPaths.value.has(parentPath)) {
        nodes.push({
          path: file,
          name: parts[parts.length - 1],
          isDir: false,
          depth: parts.length - 1
        })
      }
    }
  }

  return nodes
})

function handleClick(node: TreeNode) {
  if (node.isDir) {
    if (expandedPaths.value.has(node.path)) {
      expandedPaths.value.delete(node.path)
    } else {
      expandedPaths.value.add(node.path)
    }
    expandedPaths.value = new Set(expandedPaths.value)
  }
}

function handleDoubleClick(node: TreeNode) {
  if (!node.isDir) {
    emit('preview-file', node.path)
  }
}

function toggleSelect(node: TreeNode) {
  if (node.isDir) {
    const filesInDir = getFilesInDir(node.path)
    emit('select-folder', filesInDir)
  } else {
    emit('toggle-select', node.path)
  }
}
</script>

<style>
.vue-recycle-scroller {
  height: 100%;
}
</style>
