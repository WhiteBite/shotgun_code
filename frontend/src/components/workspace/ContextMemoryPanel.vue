<template>
    <div class="memory-panel">
        <div class="panel-header">
            <div class="panel-title">
                <div class="panel-icon panel-icon-emerald">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                    </svg>
                </div>
                <span>{{ t('memory.title') }}</span>
            </div>
        </div>

        <!-- Tabs -->
        <div class="memory-tabs">
            <button class="tab-btn" :class="activeTab === 'contexts' ? 'tab-btn-active-emerald' : 'tab-btn-inactive'"
                @click="activeTab = 'contexts'">
                {{ t('memory.savedContexts') }}
            </button>
            <button class="tab-btn" :class="activeTab === 'preferences' ? 'tab-btn-active-emerald' : 'tab-btn-inactive'"
                @click="activeTab = 'preferences'">
                {{ t('memory.preferences') }}
            </button>
        </div>

        <!-- Contexts Tab -->
        <div v-if="activeTab === 'contexts'" class="memory-content">
            <div v-if="loading" class="memory-loading">
                <svg class="w-6 h-6 animate-spin text-emerald-500" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                </svg>
            </div>

            <div v-else-if="contexts.length === 0" class="memory-empty">
                <svg class="w-10 h-10 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
                <p>{{ t('memory.noContexts') }}</p>
            </div>

            <div v-else class="memory-list">
                <div v-for="ctx in contexts" :key="ctx.id" class="memory-item" @click="$emit('select-context', ctx)">
                    <div class="memory-item-header">
                        <span class="memory-item-topic">{{ ctx.topic }}</span>
                        <span class="memory-item-date">{{ formatDate(ctx.lastAccessed) }}</span>
                    </div>
                    <div v-if="ctx.summary" class="memory-item-summary">
                        {{ truncate(ctx.summary, 100) }}
                    </div>
                    <div class="memory-item-meta">
                        <span>{{ ctx.files?.length || 0 }} {{ t('memory.files') }}</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Preferences Tab -->
        <div v-if="activeTab === 'preferences'" class="memory-content">
            <div v-if="Object.keys(preferences).length === 0" class="memory-empty">
                <svg class="w-10 h-10 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <p>{{ t('memory.noPreferences') }}</p>
            </div>

            <div v-else class="preferences-list">
                <div v-for="(value, key) in preferences" :key="key" class="preference-item">
                    <span class="preference-key">{{ formatPreferenceKey(key) }}</span>
                    <span class="preference-value">{{ value }}</span>
                </div>
            </div>

            <!-- Quick preference toggles -->
            <div class="preferences-quick">
                <label class="preference-toggle">
                    <input type="checkbox" :checked="preferences.exclude_tests === 'true'"
                        @change="togglePreference('exclude_tests', $event)" />
                    <span>{{ t('memory.excludeTests') }}</span>
                </label>
                <label class="preference-toggle">
                    <input type="checkbox" :checked="preferences.exclude_vendor === 'true'"
                        @change="togglePreference('exclude_vendor', $event)" />
                    <span>{{ t('memory.excludeVendor') }}</span>
                </label>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { ref } from 'vue'

const { t } = useI18n()

interface SavedContext {
    id: string
    topic: string
    summary?: string
    files?: string[]
    lastAccessed: string
}

interface Props {
    contexts?: SavedContext[]
    preferences?: Record<string, string>
    loading?: boolean
}

withDefaults(defineProps<Props>(), {
    contexts: () => [],
    preferences: () => ({}),
    loading: false,
})

const emit = defineEmits<{
    'select-context': [ctx: SavedContext]
    'set-preference': [key: string, value: string]
}>()

const activeTab = ref<'contexts' | 'preferences'>('contexts')

function formatDate(dateStr: string): string {
    const date = new Date(dateStr)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))

    if (days === 0) return t('memory.today')
    if (days === 1) return t('memory.yesterday')
    if (days < 7) return `${days} ${t('memory.daysAgo')}`
    return date.toLocaleDateString()
}

function truncate(text: string, maxLength: number): string {
    if (text.length <= maxLength) return text
    return text.slice(0, maxLength) + '...'
}

function formatPreferenceKey(key: string): string {
    return key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

function togglePreference(key: string, event: Event) {
    const target = event.target as HTMLInputElement
    emit('set-preference', key, target.checked ? 'true' : 'false')
}
</script>

<style scoped>
.memory-panel {
    @apply flex flex-col h-full bg-transparent;
}

.memory-tabs {
    @apply flex gap-1 px-3 py-2 border-b border-gray-700/30;
}

.memory-content {
    @apply flex-1 overflow-auto p-3;
}

.memory-loading {
    @apply flex items-center justify-center h-32;
}

.memory-empty {
    @apply flex flex-col items-center justify-center h-32 gap-2 text-gray-500 text-sm;
}

.memory-list {
    @apply space-y-2;
}

.memory-item {
    @apply p-3 bg-gray-800/50 rounded-lg border border-gray-700/30 cursor-pointer hover:bg-gray-700/50 transition-colors;
}

.memory-item-header {
    @apply flex items-center justify-between mb-1;
}

.memory-item-topic {
    @apply text-sm font-medium text-gray-200;
}

.memory-item-date {
    @apply text-xs text-gray-500;
}

.memory-item-summary {
    @apply text-xs text-gray-400 mb-2;
}

.memory-item-meta {
    @apply text-xs text-gray-500;
}

.preferences-list {
    @apply space-y-2 mb-4;
}

.preference-item {
    @apply flex items-center justify-between p-2 bg-gray-800/50 rounded border border-gray-700/30;
}

.preference-key {
    @apply text-sm text-gray-300;
}

.preference-value {
    @apply text-sm text-gray-400 font-mono;
}

.preferences-quick {
    @apply space-y-2 pt-4 border-t border-gray-700/30;
}

.preference-toggle {
    @apply flex items-center gap-2 text-sm text-gray-300 cursor-pointer;
}

.preference-toggle input {
    @apply w-4 h-4 rounded border-gray-600 bg-gray-700 text-emerald-500 focus:ring-emerald-500 focus:ring-offset-gray-900;
}
</style>
