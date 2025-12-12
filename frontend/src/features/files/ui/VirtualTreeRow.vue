<template>
  <div
    :class="[
      'tree-row group',
      props.isSelected ? 'tree-row-selected' : '',
      { 'opacity-50': item.node.isIgnored }
    ]"
    :style="{ paddingLeft: `${item.depth * 16 + 8}px` }"
    @click="handleClick"
    @contextmenu.prevent="handleContextMenu"
    :title="item.node.isIgnored ? `${item.node.path} (ignored)` : item.node.path"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <!-- Tree guide lines - continuous vertical + L/T connectors -->
    <svg v-if="item.depth > 0" class="tree-guides" :width="item.depth * 16 + 16" height="28" style="left: 0">
      <!-- Vertical continuation lines for ancestors that have more siblings -->
      <template v-for="(hasMore, idx) in item.ancestorHasMoreSiblings" :key="'v-' + idx">
        <line v-if="hasMore" 
          :x1="idx * 16 + 16" y1="0" 
          :x2="idx * 16 + 16" y2="28"
          class="tree-guide-line" :class="`tree-guide-${Math.min(idx + 1, 5)}`" />
      </template>
      <!-- Current level: vertical part (L or └ shape) -->
      <line :x1="(item.depth - 1) * 16 + 16" y1="0" 
            :x2="(item.depth - 1) * 16 + 16" :y2="item.isLast ? 14 : 28"
            class="tree-guide-line" :class="`tree-guide-${Math.min(item.depth, 5)}`" />
      <!-- Current level: horizontal connector -->
      <line :x1="(item.depth - 1) * 16 + 16" y1="14" 
            :x2="item.depth * 16 + 12" y2="14"
            class="tree-guide-line" :class="`tree-guide-${Math.min(item.depth, 5)}`" />
    </svg>

    <!-- Expand/Collapse Icon -->
    <div
      v-if="item.node.isDir"
      :class="['tree-expand', item.node.isExpanded ? 'tree-expand-open' : '']"
      @click.stop="handleExpand"
    >
      <svg class="tree-expand-icon" fill="currentColor" viewBox="0 0 20 20">
        <path
          fill-rule="evenodd"
          d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
          clip-rule="evenodd"
        />
      </svg>
    </div>
    <div v-else class="w-5"></div>

    <!-- Checkbox -->
    <div
      :class="[
        'tree-cb',
        props.checkboxState !== 'none' ? 'tree-cb-checked' : '',
        props.checkboxState === 'partial' ? 'tree-cb-partial' : ''
      ]"
      @click.stop="handleToggleSelect"
    >
      <svg
        v-if="props.checkboxState === 'full'"
        class="tree-cb-icon"
        fill="currentColor"
        viewBox="0 0 20 20"
      >
        <path
          fill-rule="evenodd"
          d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
          clip-rule="evenodd"
        />
      </svg>
      <div
        v-else-if="props.checkboxState === 'partial'"
        class="w-2 h-0.5 bg-white rounded-full"
      ></div>
    </div>

    <!-- File/Folder Icon -->
    <div class="tree-icon">
      <svg
        v-if="item.node.isDir"
        class="tree-folder-icon"
        fill="currentColor"
        viewBox="0 0 20 20"
      >
        <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
      </svg>
      <span v-else class="tree-file-icon">{{ getFileIcon(item.node.name) }}</span>
    </div>

    <!-- Name -->
    <span class="tree-name">{{ item.node.name }}</span>

    <!-- File count badge for folders -->
    <span 
      v-if="item.node.isDir && props.fileCount > 0" 
      class="tree-count"
      :title="t('files.fileCountTooltip', { count: props.fileCount })">
      {{ props.fileCount }}
    </span>

    <!-- File size -->
    <span v-else-if="item.node.size" class="tree-size">
      {{ formatSize(item.node.size) }}
    </span>

    <!-- QuickLook button -->
    <button
      v-if="!item.node.isDir"
      class="tree-preview"
      @click.stop="handleQuickLook"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
        />
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
        />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import type { FlattenedNode } from '@/composables/useVirtualTree'
import { getFileIcon } from '@/utils/fileIcons'
import { useHoveredFile } from '../composables/useHoveredFile'
import type { FileNode } from '../model/file.store'

const { t } = useI18n()

interface Props {
  item: FlattenedNode
  compactMode?: boolean
  isSelected?: boolean
  checkboxState?: 'none' | 'partial' | 'full'
  fileCount?: number
}

const props = withDefaults(defineProps<Props>(), {
  compactMode: false,
  isSelected: false,
  checkboxState: 'none',
  fileCount: 0
})

const emit = defineEmits<{
  (e: 'toggle-select', path: string): void
  (e: 'toggle-expand', path: string): void
  (e: 'contextmenu', node: FileNode, event: MouseEvent): void
  (e: 'quicklook', path: string): void
}>()

// Use provide/inject pattern for hovered file state
const hoveredFile = useHoveredFile()

function handleClick() {
  if (props.item.node.isDir) {
    emit('toggle-expand', props.item.node.path)
  } else {
    // Clicking on a file row toggles its selection
    emit('toggle-select', props.item.node.path)
  }
}

function handleExpand() {
  emit('toggle-expand', props.item.node.path)
}

function handleToggleSelect() {
  emit('toggle-select', props.item.node.path)
}

function handleContextMenu(event: MouseEvent) {
  emit('contextmenu', props.item.node, event)
}

function handleQuickLook() {
  emit('quicklook', props.item.node.path)
}

function handleMouseEnter() {
  hoveredFile.setHovered(props.item.node.path, props.item.node.isDir)
}

function handleMouseLeave() {
  hoveredFile.clearHovered(props.item.node.path)
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return Math.round(bytes / 1024) + ' KB'
  return Math.round(bytes / (1024 * 1024)) + ' MB'
}
</script>
