<template>
  <div class="select-none">
    <div :class="[
      'tree-row group',
      isSelected ? 'tree-row-selected' : '',
      { 'opacity-50': node.isIgnored },
      animating ? 'tree-stagger' : ''
    ]" :style="{ paddingLeft: `${depth * 16 + 8}px` }" @click="handleClick" @contextmenu.prevent="handleContextMenu"
      :title="node.isIgnored ? `${node.path} (ignored)` : node.path" tabindex="0" @keydown="handleKeydown"
      @mouseenter="handleMouseEnter" @mouseleave="handleMouseLeave"
      :data-path="node.path" :data-is-dir="node.isDir">
      <!-- Tree guide lines - continuous vertical + L/T connectors -->
      <svg v-if="depth > 0" class="tree-guides" :width="depth * 16 + 16" height="32" :style="{ left: '0' }">
        <!-- Vertical continuation lines for ancestors that have more siblings -->
        <template v-for="(hasMore, idx) in ancestorHasMoreSiblings" :key="'v-' + idx">
          <line v-if="hasMore" 
            :x1="idx * 16 + 16" y1="0" 
            :x2="idx * 16 + 16" y2="32"
            class="tree-guide-line" :class="`tree-guide-${Math.min(idx + 1, 5)}`" />
        </template>
        <!-- Current level: vertical part (L or └ shape) -->
        <line :x1="(depth - 1) * 16 + 16" y1="0" 
              :x2="(depth - 1) * 16 + 16" :y2="isLast ? 16 : 32"
              class="tree-guide-line" :class="`tree-guide-${Math.min(depth, 5)}`" />
        <!-- Current level: horizontal connector -->
        <line :x1="(depth - 1) * 16 + 16" y1="16" 
              :x2="depth * 16 + 12" y2="16"
              class="tree-guide-line" :class="`tree-guide-${Math.min(depth, 5)}`" />
      </svg>

      <!-- Expand/Collapse Icon -->
      <div v-if="node.isDir" :class="['tree-expand', compactInfo.lastNode.isExpanded ? 'tree-expand-open' : '']"
        @click.stop="handleClick">
        <svg class="tree-expand-icon" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd"
            d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
            clip-rule="evenodd" />
        </svg>
      </div>
      <div v-else class="w-5"></div>

      <!-- Checkbox -->
      <div :class="[
        'tree-cb',
        (isSelected && !node.isDir) || (node.isDir && selectionState !== 'none') ? 'tree-cb-checked' : '',
        node.isDir && selectionState === 'partial' ? 'tree-cb-partial' : '',
        ripple ? 'tree-cb-ripple' : ''
      ]" @click.stop="handleToggleSelect">
        <svg v-if="isSelected && !node.isDir" class="tree-cb-icon" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd"
            d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
            clip-rule="evenodd" />
        </svg>
        <svg v-else-if="node.isDir && selectionState === 'full'" class="tree-cb-icon" fill="currentColor"
          viewBox="0 0 20 20">
          <path fill-rule="evenodd"
            d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
            clip-rule="evenodd" />
        </svg>
        <div v-else-if="node.isDir && selectionState === 'partial'" class="w-2 h-0.5 bg-white rounded-full"></div>
      </div>

      <!-- File/Folder Icon -->
      <div class="tree-icon">
        <svg v-if="node.isDir" class="tree-folder-icon" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
        </svg>
        <span v-else class="tree-file-icon">{{ getFileIcon(node.name) }}</span>
      </div>

      <!-- Ignored Icon -->
      <div v-if="node.isIgnored" class="relative group/ignored">
        <svg class="w-4 h-4 text-gray-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
        </svg>
        <div class="tree-tooltip hidden group-hover/ignored:block -top-8 left-0">
          Ignored by rules
        </div>
      </div>

      <!-- Name -->
      <span class="tree-name">{{ displayName }}</span>

      <!-- File count badge -->
      <span v-if="node.isDir && node.children" :class="['tree-count', countBounce ? 'tree-count-bounce' : '']"
        :title="`${fileCount} files`">
        {{ fileCount }}
      </span>

      <!-- File size -->
      <span v-else-if="node.size" class="tree-size">
        {{ formatSize(node.size) }}
      </span>

      <!-- QuickLook button (files only) -->
      <button v-if="!node.isDir" class="tree-preview" @click.stop="handleQuickLook" :title="t('files.quickLook')">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
        </svg>
      </button>
    </div>

    <!-- Children (recursive) with animation -->
    <Transition name="tree-children">
      <div v-if="node.isDir && compactInfo.lastNode.isExpanded && displayChildren.length > 0" class="tree-children">
        <FileTreeNode v-for="(child, index) in displayChildren" :key="child.path" :node="child" :depth="depth + 1"
          :compact-mode="compactMode" :is-last="index === displayChildren.length - 1"
          :ancestor-has-more-siblings="childAncestorHasMoreSiblings" @toggle-select="$emit('toggle-select', $event)"
          @toggle-expand="$emit('toggle-expand', $event)"
          @contextmenu="(childNode, event) => $emit('contextmenu', childNode, event)"
          @quicklook="$emit('quicklook', $event)" />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { getFileIcon } from '@/utils/fileIcons'
import { computed, ref, watch } from 'vue'
import { useFileStore, type FileNode } from '../model/file.store'

const { t } = useI18n()

interface Props {
  node: FileNode
  depth?: number
  compactMode?: boolean
  isLast?: boolean
  ancestorHasMoreSiblings?: boolean[] // Array of booleans for each ancestor level
}

const props = withDefaults(defineProps<Props>(), {
  depth: 0,
  compactMode: false,
  isLast: false,
  ancestorHasMoreSiblings: () => []
})

const emit = defineEmits<{
  (e: 'toggle-select', path: string): void
  (e: 'toggle-expand', path: string): void
  (e: 'contextmenu', node: FileNode, event: MouseEvent): void
  (e: 'quicklook', path: string): void
}>()

const fileStore = useFileStore()
const ripple = ref(false)
const countBounce = ref(false)
const animating = ref(false)

const isSelected = computed(() => fileStore.selectedPaths.has(props.node.path))

// Compact mode: merge single-child folder chains (e.g., src/main/java → "src/main/java")
const compactInfo = computed(() => {
  if (!props.compactMode || !props.node.isDir || !props.node.children) {
    return { name: props.node.name, skipChildren: null as FileNode[] | null, lastNode: props.node }
  }
  
  // Check if we should compact: single folder child only
  const children = props.node.children
  if (children.length === 1 && children[0].isDir) {
    const path: string[] = [props.node.name]
    let current = props.node
    
    // Traverse down while there's only one folder child
    while (current.children && current.children.length === 1 && current.children[0].isDir) {
      const child = current.children[0]
      path.push(child.name)
      current = child
    }
    
    // Only compact if we found multiple levels
    if (path.length > 1) {
      return { 
        name: path.join('/'), 
        skipChildren: current.children || null,
        lastNode: current
      }
    }
  }
  
  return { name: props.node.name, skipChildren: null, lastNode: props.node }
})

const displayName = computed(() => compactInfo.value.name)

// Children to display (may skip intermediate folders in compact mode)
const displayChildren = computed(() => {
  if (compactInfo.value.skipChildren) {
    return compactInfo.value.skipChildren
  }
  return props.node.children || []
})

const selectionState = computed(() => {
  if (!props.node.isDir) {
    return isSelected.value ? 'full' : 'none'
  }

  // Early exit if no files selected globally
  if (fileStore.selectedCount === 0) {
    return 'none'
  }

  const allFiles = fileStore.getAllFilesInNode(props.node)
  if (allFiles.length === 0) {
    return isSelected.value ? 'full' : 'none'
  }

  // Optimized: count selected instead of filtering entire array
  let selectedCount = 0
  const selectedSet = fileStore.selectedPaths
  for (const filePath of allFiles) {
    if (selectedSet.has(filePath)) {
      selectedCount++
    }
  }

  if (selectedCount === 0) return 'none'
  if (selectedCount === allFiles.length) return 'full'
  return 'partial'
})

const fileCount = computed(() => {
  if (!props.node.isDir) return 0
  return fileStore.getAllFilesInNode(props.node).length
})

// Build ancestor array for children
// In compact mode, we need to account for skipped levels
const childAncestorHasMoreSiblings = computed(() => {
  // Current node's contribution: does it have more siblings?
  const currentHasMore = !props.isLast
  
  // In compact mode, if we're showing a merged path like "java/com/example",
  // we only add ONE entry (for the visual row), not multiple
  return [...props.ancestorHasMoreSiblings, currentHasMore]
})

function handleClick() {
  if (props.node.isDir) {
    // In compact mode, toggle the last node in the chain
    const targetPath = compactInfo.value.lastNode.path
    emit('toggle-expand', targetPath)
  }
}

function handleToggleSelect() {
  ripple.value = true
  setTimeout(() => { ripple.value = false }, 400)

  if (props.node.isDir) {
    countBounce.value = true
    setTimeout(() => { countBounce.value = false }, 300)
  }

  emit('toggle-select', props.node.path)
}

function handleContextMenu(event: MouseEvent) {
  emit('contextmenu', props.node, event)
}

function handleQuickLook() {
  emit('quicklook', props.node.path)
}

function handleMouseEnter() {
  // Store hovered node path in a global variable for Space key handling
  window.__hoveredFilePath = props.node.path
  window.__hoveredFileIsDir = props.node.isDir
}

function handleMouseLeave() {
  // Clear only if this node was the hovered one
  if (window.__hoveredFilePath === props.node.path) {
    window.__hoveredFilePath = null
    window.__hoveredFileIsDir = null
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter') {
    event.preventDefault()
    handleToggleSelect()
  } else if (event.key === 'ArrowRight' && props.node.isDir && !props.node.isExpanded) {
    emit('toggle-expand', props.node.path)
  } else if (event.key === 'ArrowLeft' && props.node.isDir && props.node.isExpanded) {
    emit('toggle-expand', props.node.path)
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return Math.round(bytes / 1024) + ' KB'
  return Math.round(bytes / (1024 * 1024)) + ' MB'
}

// Trigger animation when node becomes visible
watch(() => props.node.isExpanded, (expanded) => {
  if (expanded) {
    animating.value = true
    setTimeout(() => { animating.value = false }, 300)
  }
}, { immediate: false })
</script>
