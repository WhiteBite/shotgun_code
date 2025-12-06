<template>
  <div class="h-full flex flex-col">
    <!-- Mode Selector + Model Limits -->
    <div class="p-3 pb-0 space-y-2">
      <!-- Chat Mode Toggle with Sliding Indicator -->
      <div class="chat-mode-tabs">
        <!-- Sliding Indicator -->
        <div 
          class="chat-mode-indicator"
          :class="chatModeIndicatorClass"
          :style="{ transform: `translateX(${chatModeIndex * 100}%)` }"
        ></div>
        
        <button
          @click="chatMode = 'manual'"
          :class="['chat-mode-tab', chatMode === 'manual' ? 'chat-mode-tab-active text-indigo-300' : '']"
          :title="t('chat.modeManualHint')"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
          </svg>
          {{ t('chat.modeManual') }}
        </button>
        <button
          @click="chatMode = 'smart'"
          :class="['chat-mode-tab', chatMode === 'smart' ? 'chat-mode-tab-active text-purple-300' : '']"
          :title="t('chat.modeSmartHint')"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
          {{ t('chat.modeSmart') }}
        </button>
        <button
          @click="chatMode = 'agentic'"
          :class="['chat-mode-tab', chatMode === 'agentic' ? 'chat-mode-tab-active text-emerald-300' : '']"
          :title="t('chat.modeAgenticHint')"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          {{ t('chat.modeAgentic') }}
        </button>
      </div>

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
        <div class="space-y-2 list-stagger">
          <button
            v-for="action in quickActions"
            :key="action.id"
            @click="sendQuickAction(action)"
            class="quick-action-btn"
          >
            <span class="quick-action-icon">{{ action.icon }}</span>
            <span class="text-xs text-gray-300">{{ action.label }}</span>
          </button>
        </div>
      </div>

      <!-- Chat Messages -->
      <div
        v-for="(msg, index) in messages"
        :key="index"
        class="group"
        :class="[msg.role === 'user' ? 'message-user' : 'message-assistant']"
      >
        <div class="flex items-start gap-2">
          <div v-if="msg.role === 'assistant'" class="message-avatar">
            <svg class="w-3.5 h-3.5 text-purple-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
            <!-- Tool Calls Info -->
            <div v-if="msg.toolCalls && msg.toolCalls.length > 0" class="mt-2">
              <button
                @click="toggleToolCalls(index)"
                class="flex items-center gap-1 text-[10px] text-emerald-400 hover:text-emerald-300"
              >
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                🔧 {{ msg.toolCalls.length }} tools ({{ msg.iterations }} iter)
                <svg :class="['w-3 h-3 transition-transform', expandedToolCalls.has(index) ? 'rotate-180' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              <div v-if="expandedToolCalls.has(index)" class="mt-2 space-y-1 text-[10px]">
                <div
                  v-for="(tc, tcIdx) in msg.toolCalls"
                  :key="tcIdx"
                  class="p-1.5 bg-gray-900/50 rounded border border-gray-700/30"
                >
                  <div class="font-medium text-emerald-400">{{ tc.tool }}</div>
                  <div class="text-gray-500 truncate" :title="tc.arguments">{{ tc.arguments }}</div>
                </div>
              </div>
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

      <!-- Thinking Indicator with Stop button -->
      <div v-if="isThinking" class="thinking-indicator">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div class="flex gap-1">
              <span class="thinking-dot animate-bounce" style="animation-delay: 0ms"></span>
              <span class="thinking-dot animate-bounce" style="animation-delay: 150ms"></span>
              <span class="thinking-dot animate-bounce" style="animation-delay: 300ms"></span>
            </div>
            <span class="text-xs text-purple-300">{{ t('chat.thinking') }}</span>
          </div>
          <button
            @click="stopGeneration"
            class="flex items-center gap-1 px-2 py-1 text-xs text-red-400 hover:text-red-300 hover:bg-red-500/10 rounded-lg transition-all"
            :title="t('chat.stop')"
          >
            <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
              <rect x="6" y="6" width="12" height="12" rx="2" />
            </svg>
            {{ t('chat.stop') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Smart Context Preview -->
    <div v-if="smartContextPreview && smartContextPreview.files.length > 0" class="mx-3 mb-2 p-3 bg-purple-500/10 border border-purple-500/30 rounded-xl">
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <span class="text-xs font-medium text-purple-300">
            {{ t('chat.foundFiles') }}: {{ smartContextPreview.files.length }}
          </span>
        </div>
        <span class="text-[10px] text-gray-500">
          ~{{ smartContextPreview.totalTokens }} tokens
        </span>
      </div>
      
      <!-- File List -->
      <div class="space-y-1 max-h-32 overflow-y-auto mb-3">
        <label
          v-for="file in smartContextPreview.files"
          :key="file.path"
          class="flex items-center gap-2 p-1.5 rounded hover:bg-purple-500/10 cursor-pointer"
        >
          <input
            type="checkbox"
            v-model="file.selected"
            class="w-3 h-3 rounded border-gray-600 bg-gray-800 text-purple-500 focus:ring-purple-500 focus:ring-offset-0"
          />
          <span class="flex-1 text-[11px] text-gray-300 truncate" :title="file.path">
            {{ file.path }}
          </span>
          <span class="text-[10px] text-purple-400/70">
            {{ Math.round(file.relevance * 100) }}%
          </span>
        </label>
      </div>
      
      <!-- Actions -->
      <div class="flex gap-2">
        <button
          @click="confirmSmartContext"
          class="flex-1 btn btn-primary !py-1.5 !text-xs"
        >
          {{ t('chat.useFiles') }}
        </button>
        <button
          @click="cancelSmartContext"
          class="btn btn-ghost !py-1.5 !text-xs"
        >
          {{ t('chat.cancelAnalysis') }}
        </button>
      </div>
    </div>

    <!-- Analyzing Indicator -->
    <div v-if="isAnalyzing" class="mx-3 mb-2 p-3 bg-purple-500/10 border border-purple-500/30 rounded-xl">
      <div class="flex items-center gap-2">
        <svg class="w-4 h-4 text-purple-400 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span class="text-xs text-purple-300">{{ t('chat.analyzing') }}</span>
      </div>
    </div>

    <!-- Input Area -->
    <div class="p-3 border-t border-gray-700/30">
      <!-- Context Toggle (only in manual mode) -->
      <div v-if="chatMode === 'manual'" class="flex items-center justify-between mb-2">
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

      <!-- Smart mode hint -->
      <div v-else class="flex items-center justify-between mb-2">
        <span class="text-[10px] text-purple-400/70">
          {{ t('chat.modeSmartHint') }}
        </span>
        <button
          v-if="messages.length > 0"
          @click="clearChat"
          class="text-[10px] text-gray-500 hover:text-gray-300"
        >
          {{ t('chat.clear') }}
        </button>
      </div>

      <!-- Input -->
      <div class="flex gap-2">
        <textarea
          v-model="inputMessage"
          @keydown.enter.exact="handleSend"
          @keydown.shift.enter.prevent="inputMessage += '\n'"
          class="input-chat text-xs flex-1"
          :placeholder="t('chat.placeholder')"
          :disabled="isThinking || isAnalyzing"
          rows="1"
          style="min-height: 36px; max-height: 120px;"
          @input="autoResize"
          ref="inputRef"
        ></textarea>
        <button
          @click="handleSend"
          :disabled="!inputMessage.trim() || isThinking || isAnalyzing"
          class="btn-cta !px-3 !py-2 self-end"
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
import { useFileStore } from '@/features/files'
import { apiService } from '@/services/api.service'
import { useAIStore } from '@/stores/ai.store'
import { useProjectStore } from '@/stores/project.store'
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

interface SmartContextFile {
  path: string
  relevance: number
  reason: string
  tokenCount: number
  selected: boolean
}

interface SmartContextPreview {
  files: SmartContextFile[]
  totalTokens: number
  reasoning: string
}

const { t } = useI18n()
const contextStore = useContextStore()
const projectStore = useProjectStore()
const fileStore = useFileStore()
const uiStore = useUIStore()
const aiStore = useAIStore()

interface ToolCallInfo {
  tool: string
  arguments: string
  result: string
}

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  contextAttached?: boolean
  tokenCount?: number
  error?: string
  toolCalls?: ToolCallInfo[]
  iterations?: number
}

const messages = ref<ChatMessage[]>([])
const inputMessage = ref('')
const isThinking = ref(false)
const isAnalyzing = ref(false)
const includeContext = ref(true)
const messagesContainer = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLTextAreaElement | null>(null)

// Chat mode: 'manual', 'smart', or 'agentic'
const chatMode = ref<'manual' | 'smart' | 'agentic'>('manual')
const smartContextPreview = ref<SmartContextPreview | null>(null)
const pendingMessage = ref<string>('')
const expandedToolCalls = ref<Set<number>>(new Set())

const quickActions = [
  { id: 'explain', icon: '💡', label: 'Объясни этот код', prompt: 'Объясни подробно, что делает этот код:' },
  { id: 'review', icon: '🔍', label: 'Code Review', prompt: 'Сделай code review этого кода. Укажи на проблемы, баги и возможные улучшения:' },
  { id: 'improve', icon: '✨', label: 'Как улучшить?', prompt: 'Как можно улучшить этот код? Предложи рефакторинг:' },
  { id: 'bugs', icon: '🐛', label: 'Найди баги', prompt: 'Найди потенциальные баги и проблемы в этом коде:' },
]

// Computed properties for model limits
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

// Chat mode indicator computed
const chatModeIndex = computed(() => {
  const modes = ['manual', 'smart', 'agentic']
  return modes.indexOf(chatMode.value)
})

const chatModeIndicatorClass = computed(() => {
  const classes: Record<string, string> = {
    manual: 'chat-mode-indicator-indigo',
    smart: 'chat-mode-indicator-purple',
    agentic: 'chat-mode-indicator-emerald'
  }
  return classes[chatMode.value]
})

// Streaming state
const streamingMessageIndex = ref<number | null>(null)

// Load provider info on mount
onMounted(async () => {
  await aiStore.loadProviderInfo()
  loadChatHistory()
  loadChatMode()
  
  EventsOn('ai:stream:chunk', handleStreamChunk)
})

onUnmounted(() => {
  EventsOff('ai:stream:chunk')
})

function loadChatMode() {
  const saved = localStorage.getItem('ai-chat-mode')
  if (saved === 'smart' || saved === 'manual') {
    chatMode.value = saved
  }
}

// Save mode when changed
function saveChatMode() {
  localStorage.setItem('ai-chat-mode', chatMode.value)
}

// Watch mode changes
import { watch } from 'vue'
watch(chatMode, saveChatMode)

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
        messages.value = parsed.slice(-50)
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

async function handleSend() {
  if (!inputMessage.value.trim() || isThinking.value || isAnalyzing.value) return

  if (chatMode.value === 'smart') {
    await analyzeAndPreview()
  } else if (chatMode.value === 'agentic') {
    await sendAgenticMessage()
  } else {
    await sendMessage()
  }
}

async function analyzeAndPreview() {
  const task = inputMessage.value.trim()
  pendingMessage.value = task
  isAnalyzing.value = true

  try {
    // Call backend to analyze task and suggest files
    console.log('[AIChat] Analyzing task:', task)
    console.log('[AIChat] Files count:', fileStore.nodes.length)
    console.log('[AIChat] Project path:', projectStore.currentPath)
    
    const result = await apiService.analyzeTaskAndCollectContext(
      task,
      JSON.stringify(fileStore.nodes),
      projectStore.currentPath || ''
    )

    console.log('[AIChat] Analysis result:', result)
    const parsed = JSON.parse(result)
    console.log('[AIChat] Parsed result:', parsed)
    
    if (parsed.selectedFiles && parsed.selectedFiles.length > 0) {
      interface ParsedFile {
        relPath?: string
        path?: string
        relevance?: number
        reason?: string
        tokenCount?: number
      }
      smartContextPreview.value = {
        files: parsed.selectedFiles.map((f: ParsedFile | string) => ({
          path: typeof f === 'string' ? f : (f.relPath || f.path || ''),
          relevance: typeof f === 'string' ? 0.8 : (f.relevance || 0.8),
          reason: typeof f === 'string' ? '' : (f.reason || ''),
          tokenCount: typeof f === 'string' ? 0 : (f.tokenCount || 0),
          selected: true
        })),
        totalTokens: parsed.estimatedTokens || 0,
        reasoning: parsed.reasoning || ''
      }
    } else {
      // No files found - show message and switch to manual
      uiStore.addToast(t('chat.noFilesFound'), 'warning')
      smartContextPreview.value = null
      pendingMessage.value = ''
    }
  } catch (e) {
    console.error('[AIChat] Smart context analysis failed:', e)
    uiStore.addToast(t('chat.error'), 'error')
    smartContextPreview.value = null
    pendingMessage.value = ''
  } finally {
    isAnalyzing.value = false
  }
}

async function confirmSmartContext() {
  if (!smartContextPreview.value || !pendingMessage.value) return

  const selectedFiles = smartContextPreview.value.files
    .filter(f => f.selected)
    .map(f => f.path)

  if (selectedFiles.length === 0) {
    uiStore.addToast(t('toast.noFiles'), 'warning')
    return
  }

  // Build context from selected files
  try {
    await contextStore.buildContext(selectedFiles)
    includeContext.value = true
  } catch (e) {
    console.error('[AIChat] Failed to build smart context:', e)
  }

  // Send the message
  inputMessage.value = pendingMessage.value
  smartContextPreview.value = null
  pendingMessage.value = ''
  
  await sendMessage()
}

function cancelSmartContext() {
  smartContextPreview.value = null
  pendingMessage.value = ''
}

async function sendMessage() {
  if (!inputMessage.value.trim() || isThinking.value) return

  const userMessage = inputMessage.value.trim()
  inputMessage.value = ''
  
  if (inputRef.value) {
    inputRef.value.style.height = '36px'
  }

  let contextContent = ''
  let tokenCount = 0
  
  if (includeContext.value && contextStore.hasContext) {
    try {
      contextContent = await contextStore.getFullContextContent()
      tokenCount = contextStore.tokenCount
    } catch (e) {
      console.warn('[AIChat] Failed to get context:', e)
    }
  }

  messages.value.push({
    role: 'user',
    content: userMessage,
    contextAttached: !!contextContent,
    tokenCount: tokenCount
  })

  scrollToBottom()
  isThinking.value = true

  let systemPrompt: string
  if (contextContent) {
    systemPrompt = `Ты - помощник по программированию. Тебе предоставлен код проекта для анализа.

ВАЖНО: Внимательно изучи предоставленный код и отвечай на вопросы ТОЛЬКО на основе этого кода.
Если вопрос касается кода - анализируй именно предоставленный контекст.

=== КОД ПРОЕКТА ===
${contextContent}
=== КОНЕЦ КОДА ===

Отвечай кратко и по существу на русском языке.`
  } else {
    systemPrompt = 'Ты - помощник по программированию. Отвечай кратко и по существу на русском языке.'
  }

  messages.value.push({
    role: 'assistant',
    content: ''
  })
  streamingMessageIndex.value = messages.value.length - 1
  
  apiService.generateCodeStream(systemPrompt, userMessage)
  scrollToBottom()
}

async function sendAgenticMessage() {
  if (!inputMessage.value.trim() || isThinking.value) return

  const userMessage = inputMessage.value.trim()
  inputMessage.value = ''
  
  if (inputRef.value) {
    inputRef.value.style.height = '36px'
  }

  messages.value.push({
    role: 'user',
    content: userMessage
  })

  scrollToBottom()
  isThinking.value = true

  try {
    const response = await apiService.agenticChat(userMessage, projectStore.currentPath || '')

    messages.value.push({
      role: 'assistant',
      content: response.response,
      toolCalls: response.toolCalls,
      iterations: response.iterations
    })
  } catch (e) {
    console.error('[AIChat] Agentic chat failed:', e)
    messages.value.push({
      role: 'assistant',
      content: '',
      error: e instanceof Error ? e.message : 'Ошибка agentic chat'
    })
  } finally {
    isThinking.value = false
    saveChatHistory()
    scrollToBottom()
  }
}

function sendQuickAction(action: { id: string; label: string; prompt: string }) {
  inputMessage.value = action.prompt
  handleSend()
}

async function copyMessage(content: string) {
  await navigator.clipboard.writeText(content)
  uiStore.addToast(t('chat.copy'), 'success')
}

function clearChat() {
  messages.value = []
  expandedToolCalls.value.clear()
  saveChatHistory()
}

function toggleToolCalls(index: number) {
  if (expandedToolCalls.value.has(index)) {
    expandedToolCalls.value.delete(index)
  } else {
    expandedToolCalls.value.add(index)
  }
}

function stopGeneration() {
  if (streamingMessageIndex.value !== null) {
    // Mark current message as stopped
    const msgIndex = streamingMessageIndex.value
    if (messages.value[msgIndex]) {
      messages.value[msgIndex].content += '\n\n[Остановлено]'
    }
    streamingMessageIndex.value = null
  }
  isThinking.value = false
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
