import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface Message {
    id: string
    role: 'user' | 'assistant'
    content: string
    timestamp: string
    contextId?: string
}

export interface ChatHistory {
    id: string
    title: string
    messages: Message[]
    createdAt: string
    updatedAt: string
}

// TODO: AI Chat module - prepared for future implementation
export const useChatStore = defineStore('chat', () => {
    // State
    const messages = ref<Message[]>([])
    const isStreaming = ref(false)
    const currentModel = ref<string>('gpt-4')
    const chatHistory = ref<ChatHistory[]>([])
    const currentChatId = ref<string | null>(null)

    // Computed
    const hasMessages = computed(() => messages.value.length > 0)
    const lastMessage = computed(() => messages.value[messages.value.length - 1])

    // Actions - Stubs for future implementation
    async function sendMessage(content: string, contextId?: string): Promise<void> {
        // TODO: Implement message sending
        console.log('[ChatStore] sendMessage stub called:', { content, contextId })

        const userMessage: Message = {
            id: `msg-${Date.now()}`,
            role: 'user',
            content,
            timestamp: new Date().toISOString(),
            contextId
        }

        messages.value.push(userMessage)

        // TODO: Call backend API to get AI response
        // TODO: Handle streaming response
    }

    async function streamMessage(content: string, contextId?: string): Promise<void> {
        // TODO: Implement streaming message
        console.log('[ChatStore] streamMessage stub called:', { content, contextId })
        isStreaming.value = true

        try {
            // TODO: Implement streaming logic
            await sendMessage(content, contextId)
        } finally {
            isStreaming.value = false
        }
    }

    function clearChat(): void {
        messages.value = []
        currentChatId.value = null
    }

    async function loadHistory(): Promise<void> {
        // TODO: Load chat history from backend or localStorage
        console.log('[ChatStore] loadHistory stub called')
    }

    async function saveChat(): Promise<void> {
        // TODO: Save current chat to history
        console.log('[ChatStore] saveChat stub called')
    }

    async function loadChat(chatId: string): Promise<void> {
        // TODO: Load specific chat from history
        console.log('[ChatStore] loadChat stub called:', chatId)
    }

    function deleteMessage(messageId: string): void {
        messages.value = messages.value.filter(m => m.id !== messageId)
    }

    function editMessage(messageId: string, newContent: string): void {
        const message = messages.value.find(m => m.id === messageId)
        if (message) {
            message.content = newContent
        }
    }

    return {
        // State
        messages,
        isStreaming,
        currentModel,
        chatHistory,
        currentChatId,
        // Computed
        hasMessages,
        lastMessage,
        // Actions
        sendMessage,
        streamMessage,
        clearChat,
        loadHistory,
        saveChat,
        loadChat,
        deleteMessage,
        editMessage
    }
})
