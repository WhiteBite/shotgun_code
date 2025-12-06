<template>
  <div class="h-full flex flex-col">
    <!-- Model Limits Indicator -->
    <div class="p-3 pb-0">
      <ModelLimitsIndicator
        v-if="isConnected"
        :used-tokens="totalUsedTokens"
        :model-name="currentModel"
        :provider-name="providerName"
      />
    </div>

    <!-- Messages -->
    <div ref="messagesContainer" class="flex-1 overflow-y-auto p-3 space-y-3">
      <!-- Welcome Message -->
      <div v-if="messages.length === 0" class="text-center py-8">
        <div class="empty-state-icon !w-14 !h-14 mx-auto mb-4 bg-gradient-to-br from-purple-500/20 to-indigo-500/20 border-purple-500/30">
          <svg class="!w-7 !h-7 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
        </div>
        <h3 class="text-sm font-semibold text-white mb-1">{{ t('chat.title') }}</h3>
        <p class="text-xs text-gray-500 mb-4">{{ t('chat.welcome') }}</p>
        
        <!-- Provider Status -->
        <div class="mb-4 px-3">
          <div class="list-item !p-2 text-left">
            <div class="flex items-center gap-2">
              <div :class="isConnected ? 'w-2 h-2 rounded-full bg-emerald-400' : 'w-2 h-2 rounded-full bg-red-400'"></div>
              <span class="text-xs text-gray-300">{{ providerName }}</span>
              <span v-if="isConnected" class="text-[10px] text-gray-500">{{ currentModel }}</span>
            </div>
          </div>
        </div>
        
        <!-- Quick Actions -->
        <div class="space-y-2">
          <button
            v-for="action in quickActions"
            :key="action.id"
            @click="sendQuickAction(action)"
            class="w-full list-item !p-2 text-left"
          >
            <div class="flex items-center gap-2">
              <span class="text-lg">{{ action.icon }}</span>
              <span class="text-xs text-gray-300">{{ action.label }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Chat Messages -->
      <div
        v-for="(msg, index) in messages"
        :key="index"
        class="group"
        :class="[
          'rounded-xl p-3 max-w-[95%]',
          msg.role === 'user' 
            ? 'ml-auto bg-indigo-500/20 border border-indigo-500/30' 
            : 'bg-gray-800/50 border border-gray-700/30'
        ]"
      >
        <div class="flex items-start gap-2">
          <div v-if="msg.role === 'assistant'" class="w-6 h-6 rounded-full bg-purple-500/20 flex items-center justify-center flex-shrink-0">
            <svg class="w-3.5 h-3.5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-xs text-gray-300 whitespace-pre-wrap break-words">{{ msg.content }}</p>
            <div v-if="msg.contextAttached" class="mt-2 flex items-center gap-1 text-[10px] text-indigo-400">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
              {{ t('chat.contextAttached') }} ({{ msg.tokenCount || 0 }} tokens)
            </div>
            <div v-if="msg.error" class="mt-2 text-[10px] text-red-400">
              {{ msg.error }}
            </div>
          </div>
          <button
            v-if="msg.role === 'assistant'"
            @click="copyMessage(msg.content)"
            class="icon-btn-sm opacity-0 group-hover:opacity-100 transition-opacity"
            :title="t('chat.copy')"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Thinking Indicator -->
      <div v-if="isThinking" class="bg-gray-800/50 border border-gray-700/30 rounded-xl p-3">
        <div class="flex items-center gap-2">
          <div class="flex gap-1">
            <span class="w-2 h-2 bg-purple-400 rounded-full animate-bounce" style="animation-delay: 0ms"></span>
            <span class="w-2 h-2 bg-purple-400 rounded-full animate-bounce" style="animation-delay: 150ms"></span>
            <span class="w-2 h-2 bg-purple-400 rounded-full animate-bounce" style="animation-delay: 300ms"></span>
          </div>
          <span class="text-xs text-gray-500">{{ t('chat.thinking') }}</span>
        </div>
      </div>
    </div>

    <!-- Input Area -->
    <div class="p-3 border-t border-gray-700/30">
      <!-- Context Toggle -->
      <div class="flex items-center justify-between mb-2">
        <label class="flex items-center gap-2 text-xs text-gray-400 cursor-pointer">
          <input
            type="checkbox"
            v-model="includeContext"
            :disabled="!contextStore.hasContext"
            class="w-3.5 h-3.5 rounded border-gray-600 bg-gray-800 text-indigo-500 focus:ring-indigo-500 focus:ring-offset-0"
          />
          <span :class="{ 'text-gray-600': !contextStore.hasContext }">
            {{ t('chat.useContext') }}
          </span>
        </label>
        <div class="flex items-center gap-2">
          <span v-if="contextStore.hasContext" class="text-[10px] text-indigo-400">
            {{ contextStore.fileCount }} {{ t('context.files') }}
          </span>
          <button
            v-if="messages.length > 0"
            @click="clearChat"
            class="text-[10px] text-gray-500 hover:text-gray-300"
          >
            {{ t('chat.clear') }}
          </button>
        </div>
      </div>

      <!-- Input -->
      <div class="flex gap-2">
        <textarea
          v-model="inputMessage"
          @keydown.enter.exact="sendMessage"
          @keydown.shift.enter.prevent="inputMessage += '\n'"
          class="input text-xs !py-2 flex-1 resize-none"
          :placeholder="t('chat.placeholder')"
          :disabled="isThinking"
          rows="1"
          style="min-height: 36px; max-height: 120px;"
          @input="autoResize"
          ref="inputRef"
        ></textarea>
        <button
          @click="sendMessage"
          :disabled="!inputMessage.trim() || isThinking"
          class="btn btn-primary !px-3 !py-2 self-end"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { EventsOff, EventsOn } from '#wailsjs/runtime/runtime'
import { useI18n } from '@/composables/useI18n'
import { useContextStore } from '@/features/context'
import { apiService } from '@/services/api.service'
import { useAIStore } from '@/stores/ai.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import ModelLimitsIndicator from './ModelLimitsIndicator.vue'

interface StreamChunk {
  content: string
  done: boolean
  error?: string
  tokensUsed?: number
  finishReason?: string
}

const { t } = useI18n()
const contextStore = useContextStore()
const uiStore = useUIStore()
const aiStore = useAIStore()

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  contextAttached?: boolean
  tokenCount?: number
  error?: string
}

const messages = ref<ChatMessage[]>([])
const inputMessage = ref('')
const isThinking = ref(false)
const includeContext = ref(true)
const messagesContainer = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLTextAreaElement | null>(null)

const quickActions = [
  { id: 'explain', icon: '💡', label: 'Объясни этот код', prompt: 'Объясни подробно, что делает этот код:' },
  { id: 'review', icon: '🔍', label: 'Code Review', prompt: 'Сделай code review этого кода. Укажи на проблемы, баги и возможные улучшения:' },
  { id: 'improve', icon: '✨', label: 'Как улучшить?', prompt: 'Как можно улучшить этот код? Предложи рефакторинг:' },
  { id: 'bugs', icon: '🐛', label: 'Найди баги', prompt: 'Найди потенциальные баги и проблемы в этом коде:' },
]

// Computed properties for model limits (using AI store)
const currentModel = computed(() => aiStore.currentModel)
const providerName = computed(() => aiStore.providerInfo.name)
const isConnected = computed(() => aiStore.isConnected)

const totalUsedTokens = computed(() => {
  const contextTokens = includeContext.value && contextStore.hasContext ? contextStore.tokenCount : 0
  const conversationTokens = messages.value.reduce((sum, msg) => {
    return sum + Math.ceil(msg.content.length / 4)
  }, 0)
  return contextTokens + conversationTokens
})

// Streaming state
const streamingMessageIndex = ref<number | null>(null)

// Load provider info on mount
onMounted(async () => {
  await aiStore.loadProviderInfo()
  loadChatHistory()
  
  // Subscribe to streaming events
  EventsOn('ai:stream:chunk', handleStreamChunk)
})

onUnmounted(() => {
  EventsOff('ai:stream:chunk')
})

function handleStreamChunk(chunk: StreamChunk) {
  if (streamingMessageIndex.value === null) return
  
  const msgIndex = streamingMessageIndex.value
  
  if (chunk.error) {
    messages.value[msgIndex].error = chunk.error
    isThinking.value = false
    streamingMessageIndex.value = null
    saveChatHistory()
    return
  }
  
  if (chunk.content) {
    messages.value[msgIndex].content += chunk.content
    scrollToBottom()
  }
  
  if (chunk.done) {
    isThinking.value = false
    streamingMessageIndex.value = null
    saveChatHistory()
    scrollToBottom()
  }
}

function loadChatHistory() {
  try {
    const saved = localStorage.getItem('ai-chat-history')
    if (saved) {
      const parsed = JSON.parse(saved)
      if (Array.isArray(parsed)) {
        messages.value = parsed.slice(-50) // Keep last 50 messages
      }
    }
  } catch (e) {
    console.warn('Failed to load chat history:', e)
  }
}

function saveChatHistory() {
  try {
    localStorage.setItem('ai-chat-history', JSON.stringify(messages.value.slice(-50)))
  } catch (e) {
    console.warn('Failed to save chat history:', e)
  }
}

async function sendMessage() {
  if (!inputMessage.value.trim() || isThinking.value) return

  const userMessage = inputMessage.value.trim()
  inputMessage.value = ''
  
  // Reset textarea height
  if (inputRef.value) {
    inputRef.value.style.height = '36px'
  }

  // Get context if enabled
  let contextContent = ''
  let tokenCount = 0
  
  if (includeContext.value && contextStore.hasContext) {
    try {
      console.log('[AIChat] Getting context content...')
      contextContent = await contextStore.getFullContextContent()
      tokenCount = contextStore.tokenCount
      console.log('[AIChat] Context loaded:', contextContent.length, 'chars,', tokenCount, 'tokens')
    } catch (e) {
      console.warn('[AIChat] Failed to get context:', e)
    }
  } else {
    console.log('[AIChat] Context not included. includeContext:', includeContext.value, 'hasContext:', contextStore.hasContext)
  }

  // Add user message
  messages.value.push({
    role: 'user',
    content: userMessage,
    contextAttached: !!contextContent,
    tokenCount: tokenCount
  })

  scrollToBottom()
  isThinking.value = true

  // Build system prompt
  let systemPrompt: string
  if (contextContent) {
    systemPrompt = `Ты - помощник по программированию. Тебе предоставлен код проекта для анализа.

ВАЖНО: Внимательно изучи предоставленный код и отвечай на вопросы ТОЛЬКО на основе этого кода.
Если вопрос касается кода - анализируй именно предоставленный контекст.

=== КОД ПРОЕКТА ===
${contextContent}
=== КОНЕЦ КОДА ===

Отвечай кратко и по существу на русском языке.`
    console.log('[AIChat] System prompt with context, length:', systemPrompt.length)
  } else {
    systemPrompt = 'Ты - помощник по программированию. Отвечай кратко и по существу на русском языке.'
  }

  // Add empty assistant message for streaming
  messages.value.push({
    role: 'assistant',
    content: ''
  })
  streamingMessageIndex.value = messages.value.length - 1
  
  // Start streaming generation
  apiService.generateCodeStream(systemPrompt, userMessage)
  scrollToBottom()
}

function sendQuickAction(action: { id: string; label: string; prompt: string }) {
  inputMessage.value = action.prompt
  sendMessage()
}

async function copyMessage(content: string) {
  await navigator.clipboard.writeText(content)
  uiStore.addToast(t('chat.copy'), 'success')
}

function clearChat() {
  messages.value = []
  saveChatHistory()
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

function autoResize(event: Event) {
  const textarea = event.target as HTMLTextAreaElement
  textarea.style.height = '36px'
  textarea.style.height = Math.min(textarea.scrollHeight, 120) + 'px'
}
</script>
