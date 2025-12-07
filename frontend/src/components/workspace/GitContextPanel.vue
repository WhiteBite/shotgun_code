<template>
    <div class="git-context-panel">
        <div class="panel-header">
            <div class="panel-title">
                <div class="panel-icon panel-icon-orange">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </div>
                <span>{{ t('gitContext.title') }}</span>
            </div>
        </div>

        <!-- Tabs -->
        <div class="git-context-tabs">
            <button class="tab-btn" :class="activeTab === 'recent' ? 'tab-btn-active-orange' : 'tab-btn-inactive'"
                @click="activeTab = 'recent'">
                {{ t('gitContext.recentChanges') }}
            </button>
            <button class="tab-btn" :class="activeTab === 'coChanged' ? 'tab-btn-active-orange' : 'tab-btn-inactive'"
                @click="activeTab = 'coChanged'">
                {{ t('gitContext.coChanged') }}
            </button>
            <button class="tab-btn" :class="activeTab === 'suggested' ? 'tab-btn-active-orange' : 'tab-btn-inactive'"
                @click="activeTab = 'suggested'">
                {{ t('gitContext.suggestContext') }}
            </button>
        </div>

        <!-- Recent Changes Tab -->
        <div v-if="activeTab === 'recent'" class="git-context-content">
            <div v-if="loading" class="git-context-loading">
                <svg class="w-6 h-6 animate-spin text-orange-500" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                </svg>
            </div>

            <div v-else-if="recentChanges.length === 0" class="git-context-empty">
                <p>{{ t('gitContext.noChanges') }}</p>
            </div>

            <div v-else class="file-list">
                <div v-for="file in recentChanges" :key="file.path" class="file-item"
                    @click="$emit('select-file', file.path)">
                    <div class="file-item-main">
                        <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                        <span class="file-item-name">{{ getFileName(file.path) }}</span>
                    </div>
                    <div class="file-item-meta">
                        <span class="file-item-changes">{{ file.changeCount }} {{ t('gitContext.changeCount') }}</span>
                        <span class="file-item-date">{{ formatDate(file.lastChanged) }}</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Co-Changed Tab -->
        <div v-if="activeTab === 'coChanged'" class="git-context-content">
            <div v-if="!selectedFile" class="git-context-hint">
                <p>{{ t('gitContext.selectFileHint') }}</p>
            </div>

            <div v-else-if="coChangedFiles.length === 0" class="git-context-empty">
                <p>{{ t('gitContext.noCoChanged') }}</p>
            </div>

            <div v-else class="file-list">
                <div v-for="(file, index) in coChangedFiles" :key="file" class="file-item"
                    @click="$emit('select-file', file)">
                    <div class="file-item-main">
                        <span class="file-item-rank">{{ index + 1 }}</span>
                        <span class="file-item-name">{{ getFileName(file) }}</span>
                    </div>
                    <button class="btn btn-ghost btn-xs" @click.stop="$emit('add-to-context', file)">
                        Add
                    </button>
                </div>
            </div>
        </div>

        <!-- Suggested Tab -->
        <div v-if="activeTab === 'suggested'" class="git-context-content">
            <div class="suggested-input">
                <input v-model="taskDescription" type="text" :placeholder="t('gitContext.taskPlaceholder')"
                    class="input" @keyup.enter="handleSuggest" />
                <button class="btn btn-primary btn-sm" :disabled="!taskDescription.trim()" @click="handleSuggest">
                    {{ t('gitContext.suggestContext') }}
                </button>
            </div>

            <div v-if="suggestedFiles.length === 0" class="git-context-empty">
                <p>{{ t('gitContext.noSuggestions') }}</p>
            </div>

            <div v-else class="file-list">
                <div v-for="file in suggestedFiles" :key="file" class="file-item" @click="$emit('select-file', file)">
                    <div class="file-item-main">
                        <svg class="w-4 h-4 text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                        </svg>
                        <span class="file-item-name">{{ getFileName(file) }}</span>
                    </div>
                    <button class="btn btn-ghost btn-xs" @click.stop="$emit('add-to-context', file)">
                        Add
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { ref } from 'vue'

const { t } = useI18n()

interface RecentChange {
    path: string
    changeCount: number
    lastChanged: string
    authors?: string[]
}

interface Props {
    recentChanges?: RecentChange[]
    coChangedFiles?: string[]
    suggestedFiles?: string[]
    selectedFile?: string
    loading?: boolean
}

withDefaults(defineProps<Props>(), {
    recentChanges: () => [],
    coChangedFiles: () => [],
    suggestedFiles: () => [],
    loading: false,
})

const activeTab = ref<'recent' | 'coChanged' | 'suggested'>('recent')
const taskDescription = ref('')
let suggestTimeout: ReturnType<typeof setTimeout> | null = null

const emit = defineEmits<{
    'select-file': [path: string]
    'add-to-context': [path: string]
    'suggest-context': [task: string]
}>()

function handleSuggest() {
    if (!taskDescription.value.trim()) return
    if (suggestTimeout) clearTimeout(suggestTimeout)
    suggestTimeout = setTimeout(() => {
        emit('suggest-context', taskDescription.value)
    }, 300)
}

function getFileName(path: string): string {
    return path.split('/').pop() || path
}

function formatDate(dateStr: string): string {
    const date = new Date(dateStr)
    return date.toLocaleDateString()
}
</script>

<style scoped>
.git-context-panel {
    @apply flex flex-col h-full bg-transparent;
}

.git-context-tabs {
    @apply flex gap-1 px-3 py-2 border-b border-gray-700/30;
}

.git-context-content {
    @apply flex-1 overflow-auto p-3;
}

.git-context-loading {
    @apply flex items-center justify-center h-32;
}

.git-context-empty {
    @apply flex items-center justify-center h-32 text-gray-500 text-sm;
}

.git-context-hint {
    @apply flex items-center justify-center h-32 text-gray-500 text-sm;
}

.file-list {
    @apply space-y-1;
}

.file-item {
    @apply flex items-center justify-between p-2 rounded hover:bg-gray-700/30 cursor-pointer transition-colors;
}

.file-item-main {
    @apply flex items-center gap-2 flex-1 min-w-0;
}

.file-item-rank {
    @apply w-5 h-5 rounded bg-orange-500/20 text-orange-400 text-xs flex items-center justify-center flex-shrink-0;
}

.file-item-name {
    @apply text-sm text-gray-300 truncate;
}

.file-item-meta {
    @apply flex items-center gap-2 text-xs text-gray-500;
}

.file-item-changes {
    @apply text-orange-400;
}

.suggested-input {
    @apply flex gap-2 mb-3;
}
</style>
