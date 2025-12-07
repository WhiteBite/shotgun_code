<template>
    <div class="tool-call-container">
        <div class="tool-call-header" @click="expanded = !expanded">
            <div class="tool-call-icon" :class="statusClass">
                <svg v-if="status === 'executing'" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                    </path>
                </svg>
                <svg v-else-if="status === 'completed'" class="w-4 h-4" fill="none" stroke="currentColor"
                    viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </div>

            <div class="tool-call-info">
                <span class="tool-call-name">{{ toolName }}</span>
                <span class="tool-call-status">{{ t(`toolCalls.${status}`) }}</span>
            </div>

            <div class="tool-call-duration" v-if="duration">
                {{ formatDuration(duration) }}
            </div>

            <svg class="w-4 h-4 text-gray-500 transition-transform" :class="{ 'rotate-180': expanded }" fill="none"
                stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
        </div>

        <div v-if="expanded" class="tool-call-details">
            <div v-if="Object.keys(args).length > 0" class="tool-call-section">
                <div class="tool-call-section-title">{{ t('toolCalls.arguments') }}</div>
                <pre class="tool-call-code">{{ formatArgs(args) }}</pre>
            </div>

            <div v-if="result" class="tool-call-section">
                <div class="tool-call-section-title">{{ t('toolCalls.result') }}</div>
                <pre class="tool-call-code tool-call-result">{{ truncateResult(result) }}</pre>
            </div>

            <div v-if="error" class="tool-call-section">
                <div class="tool-call-section-title text-red-400">{{ t('error.generic') }}</div>
                <pre class="tool-call-code tool-call-error">{{ error }}</pre>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { computed, ref } from 'vue'

const { t } = useI18n()

interface Props {
    toolName: string
    args: Record<string, unknown>
    status: 'executing' | 'completed' | 'failed'
    result?: string
    error?: string
    duration?: number
}

const props = defineProps<Props>()
const expanded = ref(false)

const statusClass = computed(() => ({
    'tool-call-icon-executing': props.status === 'executing',
    'tool-call-icon-completed': props.status === 'completed',
    'tool-call-icon-failed': props.status === 'failed',
}))

function formatArgs(args: Record<string, unknown>): string {
    return JSON.stringify(args, null, 2)
}

function formatDuration(ms: number): string {
    if (ms < 1000) return `${ms}ms`
    return `${(ms / 1000).toFixed(1)}s`
}

function truncateResult(result: string): string {
    const maxLength = 500
    if (result.length <= maxLength) return result
    return result.slice(0, maxLength) + '\n... (truncated)'
}
</script>

<style scoped>
.tool-call-container {
    @apply bg-gray-800/50 border border-gray-700/50 rounded-lg overflow-hidden my-2;
}

.tool-call-header {
    @apply flex items-center gap-3 px-3 py-2 cursor-pointer hover:bg-gray-700/30 transition-colors;
}

.tool-call-icon {
    @apply w-6 h-6 rounded flex items-center justify-center flex-shrink-0;
}

.tool-call-icon-executing {
    @apply bg-blue-500/20 text-blue-400;
}

.tool-call-icon-completed {
    @apply bg-emerald-500/20 text-emerald-400;
}

.tool-call-icon-failed {
    @apply bg-red-500/20 text-red-400;
}

.tool-call-info {
    @apply flex-1 flex items-center gap-2;
}

.tool-call-name {
    @apply text-sm font-medium text-gray-200;
}

.tool-call-status {
    @apply text-xs text-gray-500;
}

.tool-call-duration {
    @apply text-xs text-gray-500 font-mono;
}

.tool-call-details {
    @apply border-t border-gray-700/50 p-3 space-y-3;
}

.tool-call-section {
    @apply space-y-1;
}

.tool-call-section-title {
    @apply text-xs font-medium text-gray-400 uppercase tracking-wide;
}

.tool-call-code {
    @apply text-xs font-mono bg-gray-900/50 rounded p-2 overflow-x-auto text-gray-300 max-h-48 overflow-y-auto;
}

.tool-call-result {
    @apply text-emerald-300/80;
}

.tool-call-error {
    @apply text-red-300/80;
}
</style>
