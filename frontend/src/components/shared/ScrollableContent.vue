<template>
  <div class="scrollable-content h-full flex flex-col min-h-0 relative">
    <div
        ref="contentRef"
        class="flex-1 overflow-auto p-3 scroll-smooth min-h-0"
        @scroll="handleScroll"
    >
      <template v-if="!highlight && virtualize">
        <RecycleScroller
            :items="lines"
            :item-size="lineHeight"
            key-field="i"
            class="h-full"
        >
          <template #default="{ item }">
            <pre class="text-xs font-mono text-gray-300 whitespace-pre-wrap select-text m-0">
              <code>{{ item.t }}</code>
            </pre>
          </template>
        </RecycleScroller>
      </template>

      <template v-else>
        <pre
            v-if="!highlight"
            class="text-xs font-mono text-gray-300 whitespace-pre-wrap select-text m-0"
        ><code>{{ content }}</code></pre>
        <pre
            v-else
            class="text-xs font-mono whitespace-pre-wrap select-text m-0"
        ><code
            class="hljs"
            :class="language ? `language-${language}` : ''"
            v-html="highlightedContent"
        ></code></pre>
      </template>
    </div>

    <div
        v-if="showScrollTop"
        class="absolute top-2 right-2 bg-blue-600 text-white px-2 py-1 rounded text-xs opacity-75 hover:opacity-100 cursor-pointer transition-opacity z-10"
        @click="scrollToTop"
    >
      ↑ Top
    </div>
    <div
        v-if="showScrollBottom"
        class="absolute bottom-2 right-2 bg-blue-600 text-white px-2 py-1 rounded text-xs opacity-75 hover:opacity-100 cursor-pointer transition-opacity z-10"
        @click="scrollToBottom"
    >
      ↓ Bottom
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import hljs from 'highlight.js'
import { RecycleScroller } from 'vue-virtual-scroller'

interface Props {
  content: string
  language?: string
  highlight?: boolean
  virtualize?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  language: 'plaintext',
  highlight: false,
  virtualize: false,
})

const contentRef = ref<HTMLElement>()
const scrollTop = ref(0)
const scrollHeight = ref(0)
const clientHeight = ref(0)
const lineHeight = 18

const highlightedContent = computed(() => {
  if (!props.highlight || !props.content) return props.content
  try {
    if (props.language && hljs.getLanguage(props.language)) {
      return hljs.highlight(props.content, { language: props.language }).value
    } else {
      return hljs.highlightAuto(props.content).value
    }
  } catch {
    return props.content.replace(/[&<>]/g, (s) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[s] || s))
  }
})

const lines = computed(() => props.content.split(/\r?\n/).map((t, i) => ({ i, t })))

const showScrollTop = computed(() => scrollTop.value > 100)
const showScrollBottom = computed(() => scrollTop.value + clientHeight.value < scrollHeight.value - 100)

function handleScroll() {
  if (!contentRef.value) return
  scrollTop.value = contentRef.value.scrollTop
  scrollHeight.value = contentRef.value.scrollHeight
  clientHeight.value = contentRef.value.clientHeight
}

function scrollToTop() {
  contentRef.value?.scrollTo({ top: 0, behavior: 'smooth' })
}

function scrollToBottom() {
  if (!contentRef.value) return
  contentRef.value.scrollTo({ top: contentRef.value.scrollHeight, behavior: 'smooth' })
}

onMounted(() => nextTick(() => handleScroll()))
</script>

<style scoped>
.scrollable-content ::-webkit-scrollbar { width: 8px; }
.scrollable-content ::-webkit-scrollbar-track { background: rgba(31, 41, 55, 0.5); border-radius: 4px; }
.scrollable-content ::-webkit-scrollbar-thumb { background: #4b5563; border-radius: 4px; }
.scrollable-content ::-webkit-scrollbar-thumb:hover { background: #6b7280; }
.scrollable-content pre { margin: 0; }
</style>