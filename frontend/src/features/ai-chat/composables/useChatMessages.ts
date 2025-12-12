import { useContextStore } from '@/features/context'
import { apiService } from '@/services/api.service'
import { useProjectStore } from '@/stores/project.store'
import { useSettingsStore } from '@/stores/settings.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, ref, watch } from 'vue'

export interface Message {
    id: string
    role: 'user' | 'assistant'
    content: string
    timestamp: string
    toolCalls?: ToolCallLog[]
    contextAttached?: boolean
    tokenCount?: number
    iterations?: number
    error?: string
}

export interface ToolCallLog {
    tool: string
    args?: Record<string, unknown>
    arguments?: string
    result?: string
    error?: string
    duration?: number
}

export interface SmartContextPreview {
    files: string[]
    totalTokens: number
    query: string
}

type ChatMode = 'manual' | 'smart' | 'agentic'
type TranslateFunc = (key: string, params?: Record<string, string | number>) => string

export function useChatMessages() {
    const uiStore = useUIStore()
    const projectStore = useProjectStore()
    const settingsStore = useSettingsStore()
    const contextStore = useContextStore()

    // State
    const messages = ref<Message[]>([])
    const inputMessage = ref('')
    const isThinking = ref(false)
    const isAnalyzing = ref(false)
    const includeContext = ref(true)
    const chatMode = ref<ChatMode>('manual')
    const smartContextPreview = ref<SmartContextPreview | null>(null)
    const expandedToolCalls = ref<Set<number>>(new Set())
    const abortController = ref<AbortController | null>(null)

    // Computed
    const currentModel = computed(() => settingsStore.settings.aiModel || 'gpt-4')
    const providerName = computed(() => 'OpenAI') // Provider is determined by backend
    const isConnected = computed(() => true) // Connection status managed by backend
    const totalUsedTokens = computed(() => {
        return messages.value.reduce((acc, msg) => acc + (msg.content?.length || 0) / 4, 0)
    })

    const chatModeIndex = computed(() => {
        const modes: ChatMode[] = ['manual', 'smart', 'agentic']
        return modes.indexOf(chatMode.value)
    })

    const chatModeIndicatorClass = computed(() => ({
        'chat-mode-indicator-indigo': chatMode.value === 'manual',
        'chat-mode-indicator-purple': chatMode.value === 'smart',
        'chat-mode-indicator-emerald': chatMode.value === 'agentic',
    }))

    // Actions
    async function sendMessage(scrollToBottom: () => void) {
        if (!inputMessage.value.trim() || isThinking.value) return

        const content = inputMessage.value.trim()
        inputMessage.value = ''

        const userMessage: Message = {
            id: `msg-${Date.now()}`,
            role: 'user',
            content,
            timestamp: new Date().toISOString(),
        }
        messages.value.push(userMessage)
        scrollToBottom()

        isThinking.value = true
        abortController.value = new AbortController()

        try {
            const projectRoot = projectStore.currentPath || ''
            // Get context content if available (for future use)
            if (includeContext.value && contextStore.hasContext) {
                try {
                    await contextStore.getFullContextContent()
                } catch {
                    // Ignore context loading errors
                }
            }

            const response = await apiService.agenticChat(content, projectRoot)

            const assistantMessage: Message = {
                id: `msg-${Date.now()}-ai`,
                role: 'assistant',
                content: response.response,
                timestamp: new Date().toISOString(),
                toolCalls: response.toolCalls,
            }
            messages.value.push(assistantMessage)
            scrollToBottom()
        } catch (error) {
            if ((error as Error).name !== 'AbortError') {
                messages.value.push({
                    id: `msg-${Date.now()}-error`,
                    role: 'assistant',
                    content: 'Error: Failed to get response',
                    timestamp: new Date().toISOString(),
                })
            }
        } finally {
            isThinking.value = false
            abortController.value = null
        }
    }

    async function sendAgenticMessage(scrollToBottom: () => void) {
        if (!inputMessage.value.trim() || isThinking.value) return

        const content = inputMessage.value.trim()
        inputMessage.value = ''

        const userMessage: Message = {
            id: `msg-${Date.now()}`,
            role: 'user',
            content,
            timestamp: new Date().toISOString(),
        }
        messages.value.push(userMessage)
        scrollToBottom()

        isThinking.value = true

        try {
            const projectRoot = projectStore.currentPath || ''
            const response = await apiService.agenticChat(content, projectRoot)

            const assistantMessage: Message = {
                id: `msg-${Date.now()}-ai`,
                role: 'assistant',
                content: response.response,
                timestamp: new Date().toISOString(),
                toolCalls: response.toolCalls,
            }
            messages.value.push(assistantMessage)
            scrollToBottom()
        } catch {
            messages.value.push({
                id: `msg-${Date.now()}-error`,
                role: 'assistant',
                content: 'Error: Agentic chat failed',
                timestamp: new Date().toISOString(),
            })
        } finally {
            isThinking.value = false
        }
    }

    async function analyzeAndPreview(t: TranslateFunc) {
        if (!inputMessage.value.trim()) return

        isAnalyzing.value = true
        try {
            // Smart context analysis - use semantic search to find relevant files
            const result = await apiService.semanticSearch({
                query: inputMessage.value,
                projectRoot: projectStore.currentPath || '',
                topK: 10
            })

            if (result.results && result.results.length > 0) {
                const files = result.results.map((r) => r.chunk.filePath)
                smartContextPreview.value = {
                    files,
                    totalTokens: files.length * 500, // Rough estimate
                    query: inputMessage.value,
                }
            } else {
                uiStore.addToast(t('chat.noFilesFound'), 'warning')
            }
        } catch {
            uiStore.addToast(t('chat.analysisFailed'), 'error')
        } finally {
            isAnalyzing.value = false
        }
    }

    async function confirmSmartContext(t: TranslateFunc, scrollToBottom: () => void) {
        if (!smartContextPreview.value) return

        const preview = smartContextPreview.value
        smartContextPreview.value = null

        // Build context from selected files
        try {
            await contextStore.buildContext(preview.files)

            // Now send the message with context
            inputMessage.value = preview.query
            includeContext.value = true
            await sendMessage(scrollToBottom)
        } catch {
            uiStore.addToast(t('chat.contextBuildFailed'), 'error')
        }
    }

    function cancelSmartContext() {
        smartContextPreview.value = null
    }

    function clearChat() {
        messages.value = []
        expandedToolCalls.value.clear()
    }

    function toggleToolCalls(index: number) {
        if (expandedToolCalls.value.has(index)) {
            expandedToolCalls.value.delete(index)
        } else {
            expandedToolCalls.value.add(index)
        }
        expandedToolCalls.value = new Set(expandedToolCalls.value)
    }

    function stopGeneration() {
        if (abortController.value) {
            abortController.value.abort()
            abortController.value = null
        }
        isThinking.value = false
    }

    function copyMessage(content: string, t: TranslateFunc) {
        navigator.clipboard.writeText(content)
        uiStore.addToast(t('common.copied'), 'success')
    }

    function initialize() {
        // Load saved chat mode
        const savedMode = localStorage.getItem('chat-mode') as ChatMode | null
        if (savedMode && ['manual', 'smart', 'agentic'].includes(savedMode)) {
            chatMode.value = savedMode
        }
    }

    function cleanup() {
        stopGeneration()
    }

    // Watch chat mode changes
    watch(chatMode, (newMode) => {
        localStorage.setItem('chat-mode', newMode)
    })

    return {
        // State
        messages,
        inputMessage,
        isThinking,
        isAnalyzing,
        includeContext,
        chatMode,
        smartContextPreview,
        expandedToolCalls,
        // Computed
        currentModel,
        providerName,
        isConnected,
        totalUsedTokens,
        chatModeIndex,
        chatModeIndicatorClass,
        // Actions
        sendMessage,
        sendAgenticMessage,
        analyzeAndPreview,
        confirmSmartContext,
        cancelSmartContext,
        clearChat,
        toggleToolCalls,
        stopGeneration,
        copyMessage,
        initialize,
        cleanup,
    }
}
