<template>
  <div class="intelligent-ai-panel bg-white rounded-lg shadow-sm border border-gray-200">
    <!-- Заголовок -->
    <div class="p-4 border-b border-gray-200">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-900 flex items-center gap-2">
          <div class="w-6 h-6 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
          </div>
          Интеллектуальная система ИИ
        </h3>
        <div class="flex items-center gap-2">
          <button
            @click="loadProviderInfo"
            class="px-3 py-1 text-sm bg-blue-50 text-blue-600 rounded-md hover:bg-blue-100 transition-colors"
          >
            Информация о провайдере
          </button>
          <button
            @click="loadAvailableModels"
            class="px-3 py-1 text-sm bg-green-50 text-green-600 rounded-md hover:bg-green-100 transition-colors"
          >
            Загрузить модели
          </button>
        </div>
      </div>
    </div>

    <!-- Основной контент -->
    <div class="p-4 space-y-4">
      <!-- Задача -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">
          Задача
        </label>
        <textarea
          v-model="task"
          rows="3"
          class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          placeholder="Опишите задачу для ИИ..."
        ></textarea>
      </div>

      <!-- Контекст -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">
          Контекст проекта
        </label>
        <textarea
          v-model="context"
          rows="4"
          class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          placeholder="Контекст проекта (автоматически заполняется из выбранных файлов)..."
          readonly
        ></textarea>
      </div>

      <!-- Настройки -->
      <div class="border-t pt-4">
        <div class="flex items-center justify-between mb-3">
          <h4 class="text-sm font-medium text-gray-700">Настройки генерации</h4>
          <button
            @click="showAdvancedSettings = !showAdvancedSettings"
            class="text-sm text-blue-600 hover:text-blue-800"
          >
            {{ showAdvancedSettings ? 'Скрыть' : 'Показать' }} расширенные
          </button>
        </div>

        <!-- Базовые настройки -->
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Температура</label>
            <input
              v-model.number="currentOptions.temperature"
              type="range"
              min="0"
              max="2"
              step="0.1"
              class="w-full"
            />
            <div class="text-xs text-gray-500 mt-1">{{ currentOptions.temperature }}</div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Макс. токенов</label>
            <input
              v-model.number="currentOptions.maxTokens"
              type="number"
              min="100"
              max="8000"
              class="w-full px-2 py-1 border border-gray-300 rounded text-sm"
            />
          </div>
        </div>

        <!-- Расширенные настройки -->
        <div v-if="showAdvancedSettings" class="mt-4 space-y-3">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Приоритет</label>
              <select
                v-model="currentOptions.priority"
                class="w-full px-2 py-1 border border-gray-300 rounded text-sm"
              >
                <option value="low">Низкий</option>
                <option value="normal">Обычный</option>
                <option value="high">Высокий</option>
                <option value="critical">Критический</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Стратегия выбора модели</label>
              <select
                v-model="currentOptions.modelSelectionStrategy"
                class="w-full px-2 py-1 border border-gray-300 rounded text-sm"
              >
                <option value="fastest">Быстрее всего</option>
                <option value="cheapest">Дешевле всего</option>
                <option value="best">Лучшее качество</option>
                <option value="balanced">Сбалансированно</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Тип проекта</label>
              <input
                v-model="currentOptions.projectType"
                type="text"
                class="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                placeholder="например: React, Go, Python..."
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Стиль кода</label>
              <input
                v-model="currentOptions.codeStyle"
                type="text"
                class="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                placeholder="например: functional, OOP, clean code..."
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="flex items-center">
              <input
                v-model="currentOptions.autoOptimizePrompt"
                type="checkbox"
                class="mr-2"
              />
              <span class="text-sm text-gray-700">Автоматическая оптимизация промпта</span>
            </label>
            <label class="flex items-center">
              <input
                v-model="currentOptions.contextCompression"
                type="checkbox"
                class="mr-2"
              />
              <span class="text-sm text-gray-700">Сжатие контекста</span>
            </label>
            <label class="flex items-center">
              <input
                v-model="currentOptions.enableFallback"
                type="checkbox"
                class="mr-2"
              />
              <span class="text-sm text-gray-700">Включить fallback провайдеры</span>
            </label>
          </div>
        </div>
      </div>

      <!-- Кнопки действий -->
      <div class="flex items-center gap-3 pt-4 border-t">
        <button
          @click="generateIntelligentCode"
          :disabled="!canGenerate || !task.trim()"
          class="flex-1 px-4 py-2 bg-gradient-to-r from-blue-500 to-purple-600 text-white rounded-md hover:from-blue-600 hover:to-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
        >
          <svg v-if="isLoading" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ isLoading ? 'Генерация...' : 'Интеллектуальная генерация' }}
        </button>
        <button
          @click="resetOptions"
          class="px-4 py-2 text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
        >
          Сбросить
        </button>
      </div>
    </div>

    <!-- Результат -->
    <div v-if="hasResult && lastResult" class="border-t border-gray-200">
      <div class="p-4">
        <div class="flex items-center justify-between mb-3">
          <h4 class="text-sm font-medium text-gray-700">Результат генерации</h4>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500">
              {{ lastResult.provider }} / {{ lastResult.modelUsed }}
            </span>
            <span class="text-xs text-gray-500">
              {{ lastResult.tokensUsed }} токенов
            </span>
            <span class="text-xs text-gray-500">
              {{ (lastResult.processingTime / 1000).toFixed(1) }}с
            </span>
          </div>
        </div>

        <!-- Качество результата -->
        <div class="mb-3 p-3 bg-gray-50 rounded-md">
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium text-gray-700">Качество результата:</span>
            <span :class="getQualityColor(lastResult.qualityScore)" class="text-sm font-medium">
              {{ getQualityText(lastResult.qualityScore) }} ({{ (lastResult.qualityScore * 100).toFixed(0) }}%)
            </span>
          </div>
        </div>

        <!-- Предложения -->
        <div v-if="lastResult.suggestions.length > 0" class="mb-3">
          <h5 class="text-sm font-medium text-gray-700 mb-2">Предложения:</h5>
          <ul class="space-y-1">
            <li
              v-for="suggestion in lastResult.suggestions"
              :key="suggestion"
              class="text-sm text-blue-600 bg-blue-50 p-2 rounded"
            >
              {{ suggestion }}
            </li>
          </ul>
        </div>

        <!-- Предупреждения -->
        <div v-if="lastResult.warnings.length > 0" class="mb-3">
          <h5 class="text-sm font-medium text-gray-700 mb-2">Предупреждения:</h5>
          <ul class="space-y-1">
            <li
              v-for="warning in lastResult.warnings"
              :key="warning"
              class="text-sm text-yellow-600 bg-yellow-50 p-2 rounded"
            >
              {{ warning }}
            </li>
          </ul>
        </div>

        <!-- Сгенерированный код -->
        <div>
          <h5 class="text-sm font-medium text-gray-700 mb-2">Сгенерированный код:</h5>
          <pre class="bg-gray-900 text-green-400 p-3 rounded-md text-sm overflow-x-auto">{{ lastResult.content }}</pre>
        </div>

        <!-- Кнопки действий с результатом -->
        <div class="flex items-center gap-2 mt-3">
          <button
            @click="copyResult"
            class="px-3 py-1 text-sm bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition-colors"
          >
            Копировать
          </button>
          <button
            @click="clearResult"
            class="px-3 py-1 text-sm bg-red-50 text-red-600 rounded hover:bg-red-100 transition-colors"
          >
            Очистить
          </button>
        </div>
      </div>
    </div>

    <!-- Информация о провайдере -->
    <div v-if="providerInfo" class="border-t border-gray-200">
      <div class="p-4">
        <h4 class="text-sm font-medium text-gray-700 mb-2">Информация о провайдере</h4>
        <div class="space-y-2 text-sm">
          <div><strong>Название:</strong> {{ providerInfo.name }}</div>
          <div><strong>Версия:</strong> {{ providerInfo.version }}</div>
          <div><strong>Возможности:</strong> {{ providerInfo.capabilities.join(', ') }}</div>
          <div><strong>Ограничения:</strong> {{ providerInfo.limitations.join(', ') }}</div>
        </div>
      </div>
    </div>

    <!-- Доступные модели -->
    <div v-if="availableModels.length > 0" class="border-t border-gray-200">
      <div class="p-4">
        <h4 class="text-sm font-medium text-gray-700 mb-2">Доступные модели ({{ availableModels.length }})</h4>
        <div class="grid grid-cols-2 gap-2 text-sm">
          <div
            v-for="model in availableModels.slice(0, 10)"
            :key="model"
            class="px-2 py-1 bg-gray-50 rounded text-gray-600"
          >
            {{ model }}
          </div>
        </div>
        <div v-if="availableModels.length > 10" class="text-xs text-gray-500 mt-2">
          И еще {{ availableModels.length - 10 }} моделей...
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useIntelligentAIStore } from '@/stores/intelligent-ai.store';
import { useContextBuilderStore } from '@/stores/context-builder.store';

const intelligentAIStore = useIntelligentAIStore();
const contextBuilderStore = useContextBuilderStore();

// Состояние компонента
const task = ref('');
const context = ref('');
const showAdvancedSettings = ref(false);

// Получаем данные из store
const {
  isLoading,
  error,
  lastResult,
  providerInfo,
  availableModels,
  currentOptions,
  hasResult,
  canGenerate,
  generateIntelligentCode: storeGenerateIntelligentCode,
  loadProviderInfo,
  loadAvailableModels,
  resetOptions,
  clearResult,
  getQualityColor,
  getQualityText,
} = intelligentAIStore;

// Computed
const projectContext = computed(() => contextBuilderStore.shotgunContextText || '');

// Методы
async function generateIntelligentCode() {
  if (!task.value.trim()) return;
  
  await storeGenerateIntelligentCode(
    task.value,
    projectContext.value,
    currentOptions.value
  );
}

function copyResult() {
  if (lastResult.value?.content) {
    navigator.clipboard.writeText(lastResult.value.content);
  }
}

// Инициализация
onMounted(() => {
  // Загружаем информацию о провайдере и модели при монтировании
  loadProviderInfo();
  loadAvailableModels();
  
  // Обновляем контекст при изменении
  context.value = projectContext.value;
});
</script>

<style scoped>
.intelligent-ai-panel {
  max-height: 80vh;
  overflow-y: auto;
}
</style>

