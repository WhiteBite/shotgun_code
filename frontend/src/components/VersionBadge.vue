<template>
    <div class="version-badge-container">
        <!-- Compact badge -->
        <button @click="openChangelog" class="version-badge" :class="{ 'has-update': hasUpdate }"
            :title="t('version.changelog')">
            <!-- Version icon -->
            <svg class="badge-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
            </svg>
            <span class="version-text">{{ displayVersion }}</span>
            <span v-if="hasUpdate" class="update-indicator">
                <span class="update-dot"></span>
            </span>
        </button>

        <!-- Changelog Modal -->
        <Teleport to="body">
            <Transition name="modal">
                <div v-if="isOpen" class="modal-overlay" @click.self="close">
                    <div class="changelog-modal">
                        <!-- Header with gradient -->
                        <div class="modal-header">
                            <div class="header-bg"></div>
                            <div class="header-content">
                                <div class="header-icon">
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                    </svg>
                                </div>
                                <div class="header-text">
                                    <h2 class="modal-title">{{ t('version.changelog') }}</h2>
                                    <div class="version-info">
                                        <span class="current-version">{{ displayVersion }}</span>
                                        <span v-if="hasUpdate" class="update-badge">
                                            <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                                                <path fill-rule="evenodd"
                                                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 0l-3 3a1 1 0 001.414 1.414L9 9.414V13a1 1 0 102 0V9.414l1.293 1.293a1 1 0 001.414-1.414z"
                                                    clip-rule="evenodd" />
                                            </svg>
                                            {{ latestVersion }}
                                        </span>
                                    </div>
                                </div>
                            </div>
                            <button @click="close" class="close-btn">
                                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                        d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>

                        <!-- Content -->
                        <div class="modal-content">
                            <!-- Loading -->
                            <div v-if="isLoading" class="state-container">
                                <div class="spinner"></div>
                                <span class="state-text">{{ t('version.checkingUpdates') }}</span>
                            </div>

                            <!-- Error -->
                            <div v-else-if="error" class="state-container state-error">
                                <div class="state-icon state-icon-error">
                                    <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                    </svg>
                                </div>
                                <span class="state-text">{{ t('version.errorLoading') }}</span>
                                <p class="state-subtext">{{ error }}</p>
                                <button @click="loadReleases" class="retry-btn">
                                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                                    </svg>
                                    {{ t('action.retry') || 'Повторить' }}
                                </button>
                            </div>

                            <!-- Empty state - no releases yet -->
                            <div v-else-if="releases.length === 0" class="state-container state-empty">
                                <div class="state-icon state-icon-empty">
                                    <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                            d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                                    </svg>
                                </div>
                                <span class="state-text">{{ t('version.noReleases') }}</span>
                                <p class="state-subtext">Релизы появятся после публикации на GitHub</p>
                                <a href="https://github.com/WhiteBite/shotgun_code/releases" target="_blank"
                                    rel="noopener noreferrer" class="github-btn">
                                    <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                                        <path
                                            d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                                    </svg>
                                    {{ t('version.viewOnGithub') }}
                                </a>
                            </div>

                            <!-- Releases list -->
                            <div v-else class="releases-list">
                                <div v-for="(release, index) in releases" :key="release.tag_name" class="release-item"
                                    :class="{
                                        'is-latest': index === 0,
                                        'is-prerelease': release.prerelease
                                    }">
                                    <div class="release-header">
                                        <div class="release-info">
                                            <span class="release-version">{{ release.tag_name }}</span>
                                            <span v-if="release.prerelease"
                                                class="tag tag-prerelease">Pre-release</span>
                                            <span v-else-if="index === 0" class="tag tag-latest">Latest</span>
                                        </div>
                                        <div class="release-actions">
                                            <span class="release-date">{{ formatDate(release.published_at) }}</span>
                                            <a :href="release.html_url" target="_blank" rel="noopener noreferrer"
                                                class="release-link">
                                                <svg class="w-4 h-4" fill="none" stroke="currentColor"
                                                    viewBox="0 0 24 24">
                                                    <path stroke-linecap="round" stroke-linejoin="round"
                                                        stroke-width="2"
                                                        d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                                                </svg>
                                            </a>
                                        </div>
                                    </div>
                                    <div v-if="release.body" class="release-body" v-html="renderMarkdown(release.body)">
                                    </div>
                                    <div v-else class="release-empty">Нет описания</div>
                                </div>
                            </div>
                        </div>

                        <!-- Footer -->
                        <div class="modal-footer">
                            <a href="https://github.com/WhiteBite/shotgun_code/releases" target="_blank"
                                rel="noopener noreferrer" class="footer-link">
                                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                                    <path
                                        d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                                </svg>
                                Все релизы на GitHub
                            </a>
                        </div>
                    </div>
                </div>
            </Transition>
        </Teleport>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { computed, onMounted, ref } from 'vue'
import { GetReleases, GetVersionInfo } from '../../wailsjs/go/main/App'

interface Release {
    tag_name: string
    name: string
    body: string
    published_at: string
    html_url: string
    prerelease: boolean
}

interface ReleasesResponse {
    currentVersion: string
    latestVersion: string
    hasUpdate: boolean
    releases: Release[]
    error?: string
}

const { t, locale } = useI18n()

const isOpen = ref(false)
const isLoading = ref(false)
const error = ref<string | null>(null)
const currentVersion = ref('dev')
const latestVersion = ref('')
const hasUpdate = ref(false)
const releases = ref<Release[]>([])

const displayVersion = computed(() => {
    const v = currentVersion.value
    return v.startsWith('v') ? v : `v${v}`
})

onMounted(async () => {
    try {
        const info = await GetVersionInfo()
        currentVersion.value = info.version
    } catch (e) {
        console.warn('Failed to get version info:', e)
    }
})

async function loadReleases() {
    if (isLoading.value) return

    isLoading.value = true
    error.value = null

    try {
        const response: ReleasesResponse = await GetReleases()
        currentVersion.value = response.currentVersion
        latestVersion.value = response.latestVersion
        hasUpdate.value = response.hasUpdate
        releases.value = response.releases || []

        if (response.error) {
            error.value = response.error
        }
    } catch (e) {
        console.error('Failed to load releases:', e)
        error.value = e instanceof Error ? e.message : 'Unknown error'
    } finally {
        isLoading.value = false
    }
}

function openChangelog() {
    isOpen.value = true
    if (releases.value.length === 0 && !error.value) {
        loadReleases()
    }
}

function close() {
    isOpen.value = false
}

function formatDate(dateStr: string): string {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    return date.toLocaleDateString(locale.value === 'ru' ? 'ru-RU' : 'en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    })
}

function renderMarkdown(text: string): string {
    if (!text) return ''
    return text
        .replace(/^### (.+)$/gm, '<h4 class="md-h4">$1</h4>')
        .replace(/^## (.+)$/gm, '<h3 class="md-h3">$1</h3>')
        .replace(/^# (.+)$/gm, '<h2 class="md-h2">$1</h2>')
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.+?)\*/g, '<em>$1</em>')
        .replace(/`(.+?)`/g, '<code class="md-code">$1</code>')
        .replace(/^- (.+)$/gm, '<li>$1</li>')
        .replace(/(<li>.*<\/li>\n?)+/g, '<ul class="md-list">$&</ul>')
        .replace(/\n\n/g, '<br><br>')
}
</script>


<style scoped>
/* Badge */
.version-badge-container {
    @apply relative;
}

.version-badge {
    @apply flex items-center gap-1.5 px-2 py-1 rounded-lg;
    @apply text-xs font-medium cursor-pointer;
    @apply transition-all duration-200;
    background: rgba(139, 92, 246, 0.1);
    border: 1px solid rgba(139, 92, 246, 0.2);
    color: #c4b5fd;
}

.version-badge:hover {
    background: rgba(139, 92, 246, 0.2);
    border-color: rgba(139, 92, 246, 0.4);
    transform: translateY(-1px);
}

.version-badge.has-update {
    background: rgba(16, 185, 129, 0.1);
    border-color: rgba(16, 185, 129, 0.3);
    color: #6ee7b7;
}

.version-badge.has-update:hover {
    background: rgba(16, 185, 129, 0.2);
    border-color: rgba(16, 185, 129, 0.5);
}

.badge-icon {
    @apply w-3.5 h-3.5 flex-shrink-0;
}

.version-text {
    @apply font-mono text-[11px];
}

.update-indicator {
    @apply flex items-center;
}

.update-dot {
    @apply w-1.5 h-1.5 rounded-full bg-emerald-400;
    animation: pulse-dot 2s ease-in-out infinite;
}

@keyframes pulse-dot {

    0%,
    100% {
        transform: scale(1);
        opacity: 1;
    }

    50% {
        transform: scale(1.2);
        opacity: 0.7;
    }
}

/* Modal Overlay */
.modal-overlay {
    @apply fixed inset-0 z-50 flex items-center justify-center p-6;
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(8px);
}

/* Modal */
.changelog-modal {
    @apply w-full max-w-xl rounded-2xl overflow-hidden flex flex-col;
    max-height: 85vh;
    background: linear-gradient(180deg, #1a1f2e 0%, #0f1219 100%);
    border: 1px solid rgba(139, 92, 246, 0.2);
    box-shadow:
        0 0 0 1px rgba(255, 255, 255, 0.05),
        0 25px 50px -12px rgba(0, 0, 0, 0.8),
        0 0 100px -20px rgba(139, 92, 246, 0.3);
}

/* Header */
.modal-header {
    @apply relative flex items-center justify-between p-5;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.header-bg {
    @apply absolute inset-0;
    background: linear-gradient(135deg, rgba(139, 92, 246, 0.1) 0%, transparent 50%);
}

.header-content {
    @apply relative flex items-center gap-4;
}

.header-icon {
    @apply w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0;
    background: linear-gradient(135deg, #8b5cf6 0%, #a855f7 100%);
    box-shadow: 0 4px 12px rgba(139, 92, 246, 0.4);
}

.header-icon svg {
    @apply w-5 h-5 text-white;
}

.header-text {
    @apply flex flex-col gap-1;
}

.modal-title {
    @apply text-lg font-semibold text-white;
}

.version-info {
    @apply flex items-center gap-2;
}

.current-version {
    @apply text-sm font-mono text-gray-400;
}

.update-badge {
    @apply flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium;
    background: rgba(16, 185, 129, 0.15);
    color: #6ee7b7;
}

.close-btn {
    @apply relative p-2 rounded-lg text-gray-400 hover:text-white;
    @apply transition-colors duration-150;
}

.close-btn:hover {
    background: rgba(255, 255, 255, 0.1);
}

/* Content */
.modal-content {
    @apply flex-1 overflow-y-auto p-5;
    min-height: 200px;
}

/* States */
.state-container {
    @apply flex flex-col items-center justify-center gap-4 py-12 text-center;
}

.state-icon {
    @apply w-16 h-16 rounded-2xl flex items-center justify-center;
}

.state-icon svg {
    @apply w-8 h-8;
}

.state-icon-error {
    background: rgba(239, 68, 68, 0.1);
    color: #f87171;
}

.state-icon-empty {
    background: rgba(139, 92, 246, 0.1);
    color: #a78bfa;
}

.state-text {
    @apply text-base font-medium text-gray-300;
}

.state-subtext {
    @apply text-sm text-gray-500 max-w-xs;
}

.spinner {
    @apply w-10 h-10 rounded-full;
    border: 3px solid rgba(139, 92, 246, 0.2);
    border-top-color: #8b5cf6;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

.retry-btn,
.github-btn {
    @apply flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium;
    @apply transition-all duration-200;
}

.retry-btn {
    background: rgba(239, 68, 68, 0.1);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.2);
}

.retry-btn:hover {
    background: rgba(239, 68, 68, 0.2);
}

.github-btn {
    background: rgba(139, 92, 246, 0.1);
    color: #a78bfa;
    border: 1px solid rgba(139, 92, 246, 0.2);
}

.github-btn:hover {
    background: rgba(139, 92, 246, 0.2);
    transform: translateY(-1px);
}

/* Releases */
.releases-list {
    @apply space-y-3;
}

.release-item {
    @apply p-4 rounded-xl;
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.06);
    transition: all 0.2s ease;
}

.release-item:hover {
    background: rgba(255, 255, 255, 0.04);
    border-color: rgba(255, 255, 255, 0.1);
}

.release-item.is-latest {
    background: rgba(139, 92, 246, 0.05);
    border-color: rgba(139, 92, 246, 0.2);
}

.release-item.is-prerelease {
    border-color: rgba(251, 191, 36, 0.2);
}

.release-header {
    @apply flex items-center justify-between gap-4 mb-2;
}

.release-info {
    @apply flex items-center gap-2;
}

.release-version {
    @apply text-base font-semibold font-mono text-white;
}

.tag {
    @apply px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wide;
}

.tag-latest {
    background: rgba(139, 92, 246, 0.2);
    color: #c4b5fd;
}

.tag-prerelease {
    background: rgba(251, 191, 36, 0.15);
    color: #fcd34d;
}

.release-actions {
    @apply flex items-center gap-3;
}

.release-date {
    @apply text-xs text-gray-500;
}

.release-link {
    @apply p-1.5 rounded-lg text-gray-500 hover:text-white;
    @apply transition-colors duration-150;
}

.release-link:hover {
    background: rgba(255, 255, 255, 0.1);
}

.release-body {
    @apply text-sm text-gray-400 leading-relaxed;
}

.release-body :deep(.md-h2) {
    @apply text-sm font-semibold text-white mt-3 mb-2;
}

.release-body :deep(.md-h3) {
    @apply text-sm font-medium text-gray-300 mt-2 mb-1;
}

.release-body :deep(.md-h4) {
    @apply text-xs font-medium text-gray-400 mt-2 mb-1;
}

.release-body :deep(.md-list) {
    @apply list-disc list-inside space-y-1 my-2 text-gray-400;
}

.release-body :deep(.md-code) {
    @apply px-1.5 py-0.5 rounded text-xs font-mono;
    background: rgba(139, 92, 246, 0.15);
    color: #c4b5fd;
}

.release-empty {
    @apply text-sm text-gray-600 italic;
}

/* Footer */
.modal-footer {
    @apply p-4 flex justify-center;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
    background: rgba(0, 0, 0, 0.2);
}

.footer-link {
    @apply flex items-center gap-2 text-sm text-gray-500 hover:text-gray-300;
    @apply transition-colors duration-150;
}

/* Transitions */
.modal-enter-active,
.modal-leave-active {
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-enter-from,
.modal-leave-to {
    opacity: 0;
}

.modal-enter-from .changelog-modal,
.modal-leave-to .changelog-modal {
    transform: scale(0.95) translateY(20px);
    opacity: 0;
}
</style>
