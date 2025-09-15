<template>
  <div class="scrollable-content h-full flex flex-col min-h-0 relative">
    <div
      ref="contentRef"
      class="flex-1 overflow-auto p-4 scroll-smooth min-h-0"
      @scroll="handleScroll"
      @mousemove="handleMouseMove"
      @click="handleClick"
    >
      <template v-if="!highlight && virtualize && lines.length > 100">
        <RecycleScroller
          :items="lines"
          :item-size="lineHeight"
          key-field="i"
          class="h-full overflow-auto"
        >
          <template #default="{ item }">
            <div class="flex">
              <div
                class="w-12 text-right pr-3 text-gray-500 text-xs select-none"
              >
                {{ item.i + 1 }}
              </div>
              <pre
                class="flex-1 text-sm font-mono text-gray-300 whitespace-pre-wrap select-text m-0 leading-relaxed"
              >
                <code>{{ item.t }}</code>
              </pre>
            </div>
          </template>
        </RecycleScroller>
      </template>

      <template v-else>
        <div v-if="!highlight && content" class="code-content">
          <div
            v-for="(line, index) in lines"
            :key="index"
            class="flex hover:bg-gray-800/50 transition-colors"
            :class="{ 'bg-blue-900/20': isSearchResult(index) }"
          >
            <div
              class="w-12 text-right pr-3 text-gray-500 text-xs select-none border-r border-gray-700"
            >
              {{ index + 1 }}
            </div>
            <pre
              class="flex-1 text-sm font-mono text-gray-300 whitespace-pre-wrap select-text m-0 leading-relaxed pl-3"
            >
              <code>{{ line.t }}</code>
            </pre>
          </div>
        </div>
        <pre
          v-else-if="highlight && content"
          class="text-sm font-mono whitespace-pre-wrap select-text m-0 leading-relaxed"
        ><code
            class="hljs"
            :class="language ? `language-${language}` : ''"
            v-html="highlightedContent"
        ></code></pre>
        <div
          v-else
          class="flex items-center justify-center h-full text-gray-500"
        >
          <div class="text-center">
            <svg
              class="w-8 h-8 mx-auto mb-2 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <p class="text-sm">No content to display</p>
          </div>
        </div>
      </template>
    </div>

    <!-- Навигационные кнопки -->
    <div
      v-if="showScrollTop"
      class="absolute top-4 right-4 bg-gray-800/90 text-white px-3 py-2 rounded-lg opacity-75 hover:opacity-100 cursor-pointer transition-all z-10 shadow-lg"
      @click="scrollToTop"
    >
      <div class="flex items-center gap-2">
        <svg
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 10l7-7m0 0l7 7m-7-7v18"
          />
        </svg>
        <span class="text-xs">Top</span>
      </div>
    </div>
    <div
      v-if="showScrollBottom"
      class="absolute bottom-4 right-4 bg-gray-800/90 text-white px-3 py-2 rounded-lg opacity-75 hover:opacity-100 cursor-pointer transition-all z-10 shadow-lg"
      @click="scrollToBottom"
    >
      <div class="flex items-center gap-2">
        <svg
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 14l-7 7m0 0l-7-7m7 7V3"
          />
        </svg>
        <span class="text-xs">Bottom</span>
      </div>
    </div>

    <!-- Индикатор прокрутки -->
    <div
      v-if="showScrollIndicator"
      class="absolute right-2 top-1/2 transform -translate-y-1/2 bg-gray-800/90 rounded-lg p-2 z-10"
    >
      <div class="w-2 h-32 bg-gray-600 rounded-full relative">
        <div
          class="absolute w-full bg-blue-500 rounded-full transition-all duration-200"
          :style="{
            top: `${scrollPercentage}%`,
            height: `${scrollThumbHeight}%`,
          }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, watch } from "vue";
import hljs from "highlight.js";
import { RecycleScroller } from "vue-virtual-scroller";
import { APP_CONFIG } from '@/config/app-config';

interface Props {
  content: string;
  language?: string;
  highlight?: boolean;
  virtualize?: boolean;
  searchResults?: Array<{ line: number; text: string; index: number }>;
  currentSearchIndex?: number;
}

const props = withDefaults(defineProps<Props>(), {
  language: "plaintext",
  highlight: false,
  virtualize: false,
  searchResults: () => [],
  currentSearchIndex: 0,
});

const emit = defineEmits<{
  (e: "scroll", position: number): void;
  (e: "line-change", line: number): void;
  (e: "click", event: MouseEvent): void;
}>();

const contentRef = ref<HTMLElement>();
const scrollTop = ref(0);
const scrollHeight = ref(0);
const clientHeight = ref(0);
const lineHeight = APP_CONFIG.ui.virtualScroll.LINE_HEIGHT;
const showScrollIndicator = ref(false);
let _hideTimer: number | null = null;

const highlightedContent = computed(() => {
  if (!props.highlight || !props.content) return props.content;
  try {
    if (props.language && hljs.getLanguage(props.language)) {
      return hljs.highlight(props.content, { language: props.language }).value;
    } else {
      return hljs.highlightAuto(props.content).value;
    }
  } catch {
    return props.content.replace(
      /[&<>]/g,
      (s) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;" })[s] || s,
    );
  }
});

const lines = computed(() =>
  props.content.split(/\r?\n/).map((t, i) => ({ i, t })),
);

const showScrollTop = computed(() => scrollTop.value > 200);
const showScrollBottom = computed(
  () => scrollTop.value + clientHeight.value < scrollHeight.value - 200,
);
const scrollPercentage = computed(() => {
  if (scrollHeight.value <= clientHeight.value) return 0;
  return (scrollTop.value / (scrollHeight.value - clientHeight.value)) * 100;
});
const scrollThumbHeight = computed(() => {
  if (scrollHeight.value <= clientHeight.value) return 100;
  return Math.max(10, (clientHeight.value / scrollHeight.value) * 100);
});

function handleScroll() {
  if (!contentRef.value) return;
  scrollTop.value = contentRef.value.scrollTop;
  scrollHeight.value = contentRef.value.scrollHeight;
  clientHeight.value = contentRef.value.clientHeight;

  // Вычисляем текущую строку
  const currentLine = Math.floor(scrollTop.value / lineHeight) + 1;
  emit("line-change", currentLine);

  // Вычисляем позицию прокрутки в процентах
  const position =
    scrollHeight.value > clientHeight.value
      ? (scrollTop.value / (scrollHeight.value - clientHeight.value)) * 100
      : 0;
  emit("scroll", position);
}

function handleMouseMove(_event: MouseEvent) {
  // Показываем индикатор прокрутки при наведении мыши
  showScrollIndicator.value = true;
  if (_hideTimer) {
    window.clearTimeout(_hideTimer);
  }
  _hideTimer = window.setTimeout(() => {
    showScrollIndicator.value = false;
    _hideTimer = null;
  }, APP_CONFIG.ui.quicklook.SCROLL_INDICATOR_HIDE_MS);
}

function handleClick(event: MouseEvent) {
  emit("click", event);
}

function scrollToTop() {
  contentRef.value?.scrollTo({ top: 0, behavior: "smooth" });
}

function scrollToBottom() {
  if (!contentRef.value) return;
  contentRef.value.scrollTo({
    top: contentRef.value.scrollHeight,
    behavior: "smooth",
  });
}

function scrollToLine(lineNumber: number) {
  if (!contentRef.value) return;
  const targetTop = (lineNumber - 1) * lineHeight;
  contentRef.value.scrollTo({ top: targetTop, behavior: "smooth" });
}

function isSearchResult(lineIndex: number): boolean {
  return props.searchResults.some((result) => result.line === lineIndex + 1);
}

// Экспортируем методы для использования в родительском компоненте
defineExpose({
  scrollToLine,
  scrollToTop,
  scrollToBottom,
  getScrollPosition: () => scrollPercentage.value,
  getCurrentLine: () => Math.floor(scrollTop.value / lineHeight) + 1,
});

// Обновляем скролл при изменении контента
watch(
  () => props.content,
  () => {
    nextTick(() => handleScroll());
  },
);

// Следим за изменениями поиска
watch(
  () => props.currentSearchIndex,
  (newIndex) => {
    if (props.searchResults[newIndex]) {
      const result = props.searchResults[newIndex];
      scrollToLine(result.line);
    }
  },
);

onMounted(() => nextTick(() => handleScroll()));
import { onUnmounted } from "vue";
onUnmounted(() => {
  if (_hideTimer) window.clearTimeout(_hideTimer);
});
</script>

<style scoped>
.scrollable-content ::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
.scrollable-content ::-webkit-scrollbar-track {
  background: rgba(31, 41, 55, 0.4);
  border-radius: 4px;
}
.scrollable-content ::-webkit-scrollbar-thumb {
  background: #6b7280;
  border-radius: 4px;
  border: 1px solid rgba(31, 41, 55, 0.6);
}
.scrollable-content ::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}
.scrollable-content ::-webkit-scrollbar-corner {
  background: rgba(31, 41, 55, 0.4);
}

.scrollable-content pre {
  margin: 0;
  line-height: 1.6;
  word-wrap: break-word;
  white-space: pre-wrap;
}

.code-content {
  font-family: "JetBrains Mono", "Fira Code", "SF Mono", "Consolas", monospace;
  font-size: 13px;
  line-height: 1.5;
}

/* Better context display */
.context-content .hljs {
  background: transparent !important;
  color: #e5e7eb;
  font-size: 13px;
  line-height: 1.6;
  padding: 1rem;
}

.context-content pre {
  background: transparent;
  border: none;
  margin: 0;
  padding: 0;
}

/* Improved readability for long context */
.context-content .hljs {
  max-width: none;
  overflow-wrap: break-word;
  word-break: break-word;
}

/* Анимации для поиска */
.code-content .flex {
  transition: background-color 0.2s ease-in-out;
}

/* Подсветка текущего результата поиска */
.code-content .flex.bg-blue-900\/20 {
  animation: searchHighlight 2s ease-in-out;
}

@keyframes searchHighlight {
  0%,
  100% {
    background-color: rgba(59, 130, 246, 0.1);
  }
  50% {
    background-color: rgba(59, 130, 246, 0.3);
  }
}

/* Улучшенная типографика для контекста */
.context-content {
  font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.context-content h1, .context-content h2, .context-content h3 {
  color: #f3f4f6;
  font-weight: 600;
  margin-top: 1.5rem;
  margin-bottom: 0.75rem;
}

.context-content h1 {
  font-size: 1.5rem;
}

.context-content h2 {
  font-size: 1.25rem;
}

.context-content h3 {
  font-size: 1.125rem;
}

.context-content p {
  color: #d1d5db;
  margin-bottom: 0.75rem;
  line-height: 1.7;
}

.context-content code {
  background: rgba(75, 85, 99, 0.3);
  border-radius: 4px;
  padding: 2px 4px;
  font-size: 0.875em;
  color: #fbbf24;
  font-family: "JetBrains Mono", "Fira Code", "SF Mono", "Consolas", monospace;
}

.context-content pre code {
  background: transparent;
  color: inherit;
  padding: 0;
}

.context-content blockquote {
  border-left: 3px solid #6b7280;
  padding-left: 1rem;
  margin: 1rem 0;
  color: #9ca3af;
  font-style: italic;
}

.context-content ul, .context-content ol {
  padding-left: 1.5rem;
  margin-bottom: 0.75rem;
}

.context-content li {
  margin-bottom: 0.25rem;
  color: #d1d5db;
}

.context-content table {
  border-collapse: collapse;
  width: 100%;
  margin: 1rem 0;
}

.context-content th, .context-content td {
  border: 1px solid #4b5563;
  padding: 0.5rem;
  text-align: left;
}

.context-content th {
  background: rgba(75, 85, 99, 0.3);
  font-weight: 600;
  color: #f3f4f6;
}

.context-content td {
  color: #d1d5db;
}
</style>
