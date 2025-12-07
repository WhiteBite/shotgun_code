<template>
    <div class="call-graph-container">
        <div class="call-graph-header">
            <div class="call-graph-title">
                <div class="panel-icon panel-icon-indigo">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                </div>
                <span>{{ t('callGraph.title') }}</span>
            </div>

            <div class="call-graph-actions">
                <button class="btn btn-ghost btn-sm" @click="zoomIn" :title="t('callGraph.zoomIn')">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7" />
                    </svg>
                </button>
                <button class="btn btn-ghost btn-sm" @click="zoomOut" :title="t('callGraph.zoomOut')">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM13 10H7" />
                    </svg>
                </button>
                <button class="btn btn-ghost btn-sm" @click="resetZoom" :title="t('callGraph.reset')">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                </button>
            </div>
        </div>

        <div class="call-graph-content" ref="graphContainer">
            <div v-if="loading" class="call-graph-loading">
                <svg class="w-8 h-8 animate-spin text-indigo-500" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                </svg>
                <span>{{ t('callGraph.building') }}</span>
            </div>

            <div v-else-if="!mermaidCode" class="call-graph-empty">
                <svg class="w-12 h-12 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                <p>{{ t('callGraph.noCallers') }}</p>
            </div>

            <div v-else ref="mermaidContainer" class="call-graph-diagram" :style="{ transform: `scale(${zoom})` }"
                v-html="renderedDiagram"></div>
        </div>

        <div v-if="selectedFunction" class="call-graph-info">
            <div class="call-graph-info-title">{{ selectedFunction }}</div>
            <div class="call-graph-info-stats">
                <span v-if="callerCount > 0">
                    <strong>{{ callerCount }}</strong> {{ t('callGraph.callers') }}
                </span>
                <span v-if="calleeCount > 0">
                    <strong>{{ calleeCount }}</strong> {{ t('callGraph.callees') }}
                </span>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { onMounted, ref, watch } from 'vue'

const { t } = useI18n()

interface Props {
    mermaidCode?: string
    loading?: boolean
    selectedFunction?: string
    callerCount?: number
    calleeCount?: number
}

const props = withDefaults(defineProps<Props>(), {
    loading: false,
    callerCount: 0,
    calleeCount: 0,
})

const renderedDiagram = ref('')
const zoom = ref(1)

async function renderMermaid() {
    if (!props.mermaidCode) {
        renderedDiagram.value = ''
        return
    }

    // Convert mermaid code to simple HTML representation
    // Full mermaid support requires: npm install mermaid
    try {
        const lines = props.mermaidCode.split('\n').filter(l => l.trim() && !l.startsWith('graph'))
        const html = lines.map(line => {
            const match = line.match(/(\w+)\s*-->?\s*(\w+)/)
            if (match) {
                return `<div class="call-edge">${match[1]} → ${match[2]}</div>`
            }
            return `<div class="call-node">${line.trim()}</div>`
        }).join('')
        renderedDiagram.value = `<div class="call-graph-text">${html}</div>`
    } catch (error) {
        console.error('Failed to render diagram:', error)
        renderedDiagram.value = `<pre class="text-xs text-gray-400">${props.mermaidCode}</pre>`
    }
}

function zoomIn() {
    zoom.value = Math.min(zoom.value + 0.2, 3)
}

function zoomOut() {
    zoom.value = Math.max(zoom.value - 0.2, 0.3)
}

function resetZoom() {
    zoom.value = 1
}

watch(() => props.mermaidCode, renderMermaid)
onMounted(renderMermaid)
</script>

<style scoped>
.call-graph-container {
    @apply flex flex-col h-full bg-gray-900 rounded-lg border border-gray-700/50;
}

.call-graph-header {
    @apply flex items-center justify-between px-4 py-3 border-b border-gray-700/50;
}

.call-graph-title {
    @apply flex items-center gap-2 text-sm font-medium text-gray-200;
}

.call-graph-actions {
    @apply flex items-center gap-1;
}

.call-graph-content {
    @apply flex-1 overflow-auto p-4 flex items-center justify-center;
}

.call-graph-loading {
    @apply flex flex-col items-center gap-3 text-gray-400;
}

.call-graph-empty {
    @apply flex flex-col items-center gap-3 text-gray-500;
}

.call-graph-diagram {
    @apply transition-transform origin-center;
}

.call-graph-diagram :deep(svg) {
    @apply max-w-full h-auto;
}

.call-graph-info {
    @apply px-4 py-3 border-t border-gray-700/50 bg-gray-800/50;
}

.call-graph-info-title {
    @apply text-sm font-medium text-gray-200 mb-1;
}

.call-graph-info-stats {
    @apply flex gap-4 text-xs text-gray-400;
}

.call-graph-text {
    @apply text-sm font-mono space-y-1;
}

.call-edge {
    @apply text-indigo-400 py-1;
}

.call-node {
    @apply text-gray-400 py-1;
}
</style>
