<template>
  <ul class="file-tree">
    <li v-for="node in nodes" :key="node.path" :class="{ 'excluded-node': node.excluded && !isIndeterminate(node) }">
      <div class="node-item" :style="{ 'padding-left': depth * 20 + 'px' }">
        <span v-if="node.isDir" @click="toggleExpand(node)" class="toggler">
          {{ node.expanded ? '▼' : '▶' }}
        </span>
        <span v-else class="item-spacer"></span>

        <input
            type="checkbox"
            :checked="!node.excluded"
            :indeterminate="isIndeterminate(node)"
            @change="emit('toggle-exclude', node)"
            class="exclude-checkbox"
            ref="checkboxes"
        />
        <span @click="node.isDir ? toggleExpand(node) : null" :class="{ 'folder-name': node.isDir }">
          {{ node.name }}
        </span>
      </div>
      <FileTree
          v-if="node.isDir && node.expanded && node.children"
          :nodes="node.children"
          :depth="depth + 1"
          @toggle-exclude="emitToggleExclude"
      />
    </li>
  </ul>
</template>

<script setup>
import { ref, onUpdated } from 'vue';

const props = defineProps({
  nodes: Array,
  depth: { type: Number, default: 0 },
});

const emit = defineEmits(['toggle-exclude']);

const checkboxes = ref([]);

function getChildrenStates(node) {
  if (!node.isDir || !node.children || node.children.length === 0) {
    return { checked: 0, unchecked: 0, total: 0 };
  }
  let checked = 0;
  let unchecked = 0;
  node.children.forEach(child => {
    if (!child.excluded) {
      checked++;
    } else {
      unchecked++;
    }
  });
  return { checked, unchecked, total: node.children.length };
}

const isIndeterminate = (node) => {
  if (!node.isDir) return false;
  const states = getChildrenStates(node);
  return states.total > 0 && states.checked > 0 && states.unchecked > 0;
};

onUpdated(() => {
  // Wails/Vue может не всегда корректно обновлять DOM свойство `indeterminate`,
  // поэтому мы делаем это вручную при обновлении компонента.
  props.nodes.forEach((node, index) => {
    if (checkboxes.value[index]) {
      checkboxes.value[index].indeterminate = isIndeterminate(node);
    }
  });
});

function toggleExpand(node) {
  if (node.isDir) {
    node.expanded = !node.expanded;
  }
}

function emitToggleExclude(node) {
  emit('toggle-exclude', node);
}
</script>

<style scoped>
.file-tree {
  list-style-type: none;
  padding-left: 0;
}
.file-tree li {
  margin: 2px 0;
}
.node-item {
  display: flex;
  align-items: center;
  cursor: default;
  user-select: none;
}
.node-item:hover {
  background-color: #f0f0f0;
}
.toggler {
  cursor: pointer;
  width: 20px;
  display: inline-block;
  text-align: center;
}
.item-spacer {
  width: 20px;
  display: inline-block;
}
.folder-name {
  cursor: pointer;
  font-weight: 500;
}
.exclude-checkbox {
  margin-right: 5px;
  cursor: pointer;
}
/* Зачеркиваем только если полностью исключен (не в промежуточном состоянии) */
.excluded-node > .node-item > span:not(.toggler) {
  text-decoration: line-through;
  color: #999;
}
</style>