<template>
  <div class="h-full flex flex-col">
    <!-- Mode Selector + Model Limits -->
    <div class="p-3 pb-0 space-y-2">
      <!-- Chat Mode Toggle with Sliding Indicator -->
      <div class="chat-mode-tabs">
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
      <ChatWelcome 
        v-if="messages.length === 0"
        :is-connected="isConnected"
        :provider-name="providerName"
        :current-model="currentModel"
        @quick-action="handleQuickAction"
      />

      <ChatMessageItem
        v-for="(msg, index) in messages"
        :key="index"
        :message="msg"
        :index="index"
        :expanded-tool-calls="expandedToolCalls"
        @copy="copyMessage(msg.content, t)"
        @toggle-tools="toggleToolCalls"
      />

      <ThinkingIndicator v-if="isThinking" @stop="stopGeneration" />
    </div>

    <!-- Smart Context Preview -->
    <SmartContextPreviewPanel
      v-if="smartContextPreview && smartContextPreview.files.length > 0"
      :preview="smartContextPreview"
      @confirm="confirmSmartContext(t, scrollToBottom)"
      @cancel="cancelSmartContext"
    />

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
    <ChatInputArea
      v-model="inputMessage"
      :chat-mode="chatMode"
      :is-thinking="isThinking"
      :is-analyzing="isAnalyzing"
      :include-context="includeContext"
      :has-messages="messages.length > 0"
      @send="handleSend"
      @clear="clearChat"
      @update:include-context="includeContext = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useChatMessages } from '@/features/ai-chat/composables/useChatMessages'
import { nextTick, onMounted, onUnmounted, ref } from 'vue'
import ChatInputArea from './chat/ChatInputArea.vue'
import ChatMessageItem from './chat/ChatMessageItem.vue'
import ChatWelcome from './chat/ChatWelcome.vue'
import SmartContextPreviewPanel from './chat/SmartContextPreviewPanel.vue'
import ThinkingIndicator from './chat/ThinkingIndicator.vue'
import ModelLimitsIndicator from './ModelLimitsIndicator.vue'

const { t } = useI18n()

const {
  messages,
  inputMessage,
  isThinking,
  isAnalyzing,
  includeContext,
  chatMode,
  smartContextPreview,
  expandedToolCalls,
  currentModel,
  providerName,
  isConnected,
  totalUsedTokens,
  chatModeIndex,
  chatModeIndicatorClass,
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
} = useChatMessages()

const messagesContainer = ref<HTMLElement | null>(null)

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

async function handleSend() {
  if (chatMode.value === 'smart') {
    await analyzeAndPreview(t)
  } else if (chatMode.value === 'agentic') {
    await sendAgenticMessage(scrollToBottom)
  } else {
    await sendMessage(scrollToBottom)
  }
}

function handleQuickAction(action: { prompt: string }) {
  inputMessage.value = action.prompt
  handleSend()
}

onMounted(() => {
  initialize()
})

onUnmounted(() => {
  cleanup()
})
</script>
