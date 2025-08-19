<template>
  <div class="intelligent-context-panel bg-white rounded-lg shadow-sm border border-gray-200">
    <!-- Заголовок -->
    <div class="p-4 border-b border-gray-200">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-900 flex items-center gap-2">
          <div class="w-6 h-6 bg-gradient-to-r from-green-500 to-blue-600 rounded-lg flex items-center justify-center">
            <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          Интеллектуальный анализ контекста
        </h3>
        <div class="flex items-center gap-2">
          <button
            @click="clearAnalysis"
            class="px-3 py-1 text-sm bg-gray-50 text-gray-600 rounded-md hover:bg-gray-100 transition-colors"
          >
            Очистить
          </button>
        </div>
      </div>
    </div>

    <!-- Основной контент -->
    <div class="p-4 space-y-4">
      <!-- Поле ввода задачи -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">
          Опишите задачу
        </label>
        <textarea
          v-model="task"
          rows="3"
          class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-transparent"
          placeholder="Например: 'Добавить валидацию формы регистрации' или 'Исправить ошибку в функции calculateTotal'"
          @keydown.enter.prevent="analyzeTask"
        ></textarea>
      </div>

      <!-- Кнопка анализа -->
      <div class="flex items-center gap-3">
        <button
          @click="analyzeTask"
          :disabled="!canAnalyze"
          class="flex-1 px-4 py-2 bg-gradient-to-r from-green-500 to-blue-600 text-white rounded-md hover:from-green-600 hover:to-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
        >
          <svg v-if="isAnalyzing" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ isAnalyzing ? 'Анализируем...' : 'Анализировать и собрать контекст' }}
        </button>
        <button
          @click="testBackend"
          :disabled="isAnalyzing"
          class="px-3 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition-colors"
          title="Тест backend"
        >
          Тест
        </button>
      </div>

      <!-- Сообщение об ошибке -->
      <div v-if="errorMessage" class="p-3 bg-red-50 border border-red-200 rounded-md">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-sm text-red-700">{{ errorMessage }}</span>
        </div>
      </div>
    </div>

    <!-- Результат анализа -->
    <div v-if="analysisResult" class="border-t border-gray-200">
      <div class="p-4 space-y-4">
        <!-- Информация о задаче -->
        <div class="bg-blue-50 p-3 rounded-md">
          <div class="flex items-center justify-between mb-2">
            <h4 class="text-sm font-medium text-blue-900">Анализ задачи</h4>
            <div class="flex items-center gap-2">
              <span class="text-xs px-2 py-1 bg-blue-100 text-blue-800 rounded">
                {{ analysisResult.taskType }}
              </span>
              <span class="text-xs px-2 py-1 bg-green-100 text-green-800 rounded">
                {{ analysisResult.priority }}
              </span>
            </div>
          </div>
          <div class="text-sm text-blue-800">
            <div><strong>Задача:</strong> {{ analysisResult.task }}</div>
            <div><strong>Время анализа:</strong> {{ formatDuration(analysisResult.analysisTime) }}</div>
            <div><strong>Уверенность:</strong> {{ (analysisResult.confidence * 100).toFixed(0) }}%</div>
          </div>
        </div>

        <!-- Выбранные файлы -->
        <div>
          <h4 class="text-sm font-medium text-gray-700 mb-2">
            Выбранные файлы ({{ analysisResult.selectedFiles?.length || 0 }})
          </h4>
          <div class="space-y-1 max-h-40 overflow-y-auto">
            <div
              v-for="file in analysisResult.selectedFiles || []"
              :key="file.relPath"
              class="flex items-center justify-between p-2 bg-gray-50 rounded text-sm"
            >
              <span class="text-gray-700">{{ file.relPath }}</span>
              <span class="text-xs text-gray-500">{{ formatFileSize(file.size) }}</span>
            </div>
          </div>
        </div>

        <!-- Зависимости -->
        <div v-if="analysisResult.dependencyFiles?.length > 0">
          <h4 class="text-sm font-medium text-gray-700 mb-2">
            Зависимости ({{ analysisResult.dependencyFiles.length }})
          </h4>
          <div class="space-y-1 max-h-32 overflow-y-auto">
            <div
              v-for="file in analysisResult.dependencyFiles"
              :key="file.relPath"
              class="flex items-center justify-between p-2 bg-yellow-50 rounded text-sm"
            >
              <span class="text-gray-700">{{ file.relPath }}</span>
              <span class="text-xs text-gray-500">{{ formatFileSize(file.size) }}</span>
            </div>
          </div>
        </div>

        <!-- Рекомендации -->
        <div v-if="analysisResult.recommendations?.length > 0">
          <h4 class="text-sm font-medium text-gray-700 mb-2">Рекомендации</h4>
          <ul class="space-y-1">
            <li
              v-for="recommendation in analysisResult.recommendations"
              :key="recommendation"
              class="text-sm text-blue-600 bg-blue-50 p-2 rounded"
            >
              {{ recommendation }}
            </li>
          </ul>
        </div>

        <!-- Статистика -->
        <div class="bg-gray-50 p-3 rounded-md">
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span class="text-gray-600">Токенов:</span>
              <span class="font-medium">{{ analysisResult.estimatedTokens || 0 }}</span>
            </div>
            <div>
              <span class="text-gray-600">Всего файлов:</span>
              <span class="font-medium">{{ (analysisResult.selectedFiles?.length || 0) + (analysisResult.dependencyFiles?.length || 0) }}</span>
            </div>
          </div>
        </div>

        <!-- Кнопки действий -->
        <div class="flex items-center gap-2 pt-2 border-t">
          <button
            @click="applyToContext"
            class="flex-1 px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors"
          >
            Применить к контексту
          </button>
          <button
            @click="copyContext"
            class="px-3 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition-colors"
          >
            Копировать контекст
          </button>
        </div>
      </div>
    </div>

    <!-- Предварительный просмотр контекста -->
    <div v-if="analysisResult?.context" class="border-t border-gray-200">
      <div class="p-4">
        <h4 class="text-sm font-medium text-gray-700 mb-2">Предварительный просмотр контекста</h4>
        <div class="bg-gray-900 text-green-400 p-3 rounded-md text-sm max-h-60 overflow-y-auto">
          <pre>{{ analysisResult.context }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { apiService } from '@/services/api.service';
import { useNotificationsStore } from '@/stores/notifications.store';
import { useFileTreeStore } from '@/stores/file-tree.store';
import { useContextBuilderStore } from '@/stores/context-builder.store';

const notifications = useNotificationsStore();
const fileTreeStore = useFileTreeStore();
const contextBuilderStore = useContextBuilderStore();

// Состояние
const task = ref('');
const isAnalyzing = ref(false);
const analysisResult = ref<any>(null);
const errorMessage = ref('');

// Computed
const canAnalyze = computed(() => task.value.trim().length > 0 && !isAnalyzing.value);

// Методы
async function analyzeTask() {
  if (!task.value.trim()) {
    notifications.addLog('Задача не может быть пустой', 'error');
    return;
  }

  isAnalyzing.value = true;
  analysisResult.value = null;
  errorMessage.value = '';

  try {
    console.log('=== НАЧАЛО АНАЛИЗА ===');
    
    // Проверяем настройки AI
    console.log('Проверяем настройки AI...');
    const settings = await apiService.getSettings();
    console.log('Настройки получены:', settings);
    
    if (!settings.selectedProvider) {
      throw new Error('AI провайдер не выбран. Перейдите в настройки и выберите провайдера.');
    }
    
    const apiKey = settings[`${settings.selectedProvider}APIKey` as keyof typeof settings] as string;
    if (!apiKey && settings.selectedProvider !== 'localai') {
      throw new Error(`API ключ для ${settings.selectedProvider} не настроен. Перейдите в настройки.`);
    }
    
    console.log('AI настройки корректны');
    
    // Получаем все файлы из дерева
    const allFiles = fileTreeStore.getAllFiles();
    console.log('Получены файлы:', allFiles.length, 'файлов');
    
    if (allFiles.length === 0) {
      throw new Error('Нет доступных файлов для анализа');
    }
    
    // Конвертируем в формат для backend
    const allFilesJson = JSON.stringify(allFiles);
    console.log('JSON файлов создан, размер:', allFilesJson.length, 'символов');
    
    console.log('Отправляем анализ:', {
      task: task.value,
      filesCount: allFiles.length,
      rootDir: fileTreeStore.rootDir
    });
    
    // Проверяем, что API сервис доступен
    console.log('API сервис доступен:', !!apiService);
    console.log('Метод analyzeTaskAndCollectContext доступен:', !!apiService.analyzeTaskAndCollectContext);
    
    // Выполняем анализ
    console.log('Вызываем API...');
    const resultJson = await apiService.analyzeTaskAndCollectContext(
      task.value,
      allFilesJson,
      fileTreeStore.rootDir || ''
    );
    
    console.log('Получен результат:', resultJson);
    console.log('Тип результата:', typeof resultJson);
    console.log('Длина результата:', resultJson?.length || 0);
    
    if (!resultJson || resultJson.trim() === '') {
      throw new Error('Получен пустой ответ от сервера');
    }
    
    console.log('Парсим JSON...');
    analysisResult.value = JSON.parse(resultJson);
    console.log('Результат распарсен:', analysisResult.value);
    
    if (!analysisResult.value) {
      throw new Error('Не удалось разобрать результат анализа');
    }
    
    notifications.addLog(
      `Анализ завершен! Выбрано ${analysisResult.value.selectedFiles?.length || 0} файлов`, 
      'success'
    );

  } catch (err: any) {
    console.error('=== ОШИБКА АНАЛИЗА ===');
    console.error('Тип ошибки:', typeof err);
    console.error('Сообщение ошибки:', err.message);
    console.error('Стек ошибки:', err.stack);
    console.error('Полная ошибка:', err);
    
    const errorMsg = err.message || err.toString() || 'Неизвестная ошибка';
    errorMessage.value = `Ошибка анализа: ${errorMsg}`;
    notifications.addLog(`Ошибка анализа: ${errorMsg}`, 'error');
  } finally {
    console.log('=== КОНЕЦ АНАЛИЗА ===');
    isAnalyzing.value = false;
  }
}

async function testBackend() {
  try {
    isAnalyzing.value = true;
    errorMessage.value = '';
    console.log('=== НАЧАЛО ТЕСТА БАКЕНДА ===');

    // Проверяем настройки AI
    console.log('Проверяем настройки AI...');
    const settings = await apiService.getSettings();
    console.log('Настройки получены:', settings);
    
    if (!settings.selectedProvider) {
      throw new Error('AI провайдер не выбран. Перейдите в настройки и выберите провайдера.');
    }
    
    const apiKey = settings[`${settings.selectedProvider}APIKey` as keyof typeof settings] as string;
    if (!apiKey && settings.selectedProvider !== 'localai') {
      throw new Error(`API ключ для ${settings.selectedProvider} не настроен. Перейдите в настройки.`);
    }
    
    console.log('AI настройки корректны');

    // Получаем все файлы из дерева
    const allFiles = fileTreeStore.getAllFiles();
    console.log('Получены файлы:', allFiles.length, 'файлов');
    
    if (allFiles.length === 0) {
      throw new Error('Нет доступных файлов для теста');
    }
    
    // Конвертируем в формат для backend
    const allFilesJson = JSON.stringify(allFiles);
    console.log('JSON файлов создан, размер:', allFilesJson.length, 'символов');
    
    console.log('Отправляем тест:', {
      filesCount: allFiles.length,
      rootDir: fileTreeStore.rootDir
    });
    
    // Проверяем, что API сервис доступен
    console.log('API сервис доступен:', !!apiService);
    console.log('Метод testBackend доступен:', !!apiService.testBackend);
    
    // Выполняем тест
    console.log('Вызываем API...');
    const resultJson = await apiService.testBackend(
      allFilesJson,
      fileTreeStore.rootDir || ''
    );
    
    console.log('Получен результат:', resultJson);
    console.log('Тип результата:', typeof resultJson);
    console.log('Длина результата:', resultJson?.length || 0);
    
    if (!resultJson || resultJson.trim() === '') {
      throw new Error('Получен пустой ответ от сервера');
    }
    
    console.log('Парсим JSON...');
    const result = JSON.parse(resultJson);
    console.log('Результат распарсен:', result);
    
    if (!result) {
      throw new Error('Не удалось разобрать результат теста');
    }
    
    notifications.addLog(
      `Тест завершен! Статус: ${result.status}`, 
      'success'
    );

  } catch (err: any) {
    console.error('=== ОШИБКА ТЕСТА БАКЕНДА ===');
    console.error('Тип ошибки:', typeof err);
    console.error('Сообщение ошибки:', err.message);
    console.error('Стек ошибки:', err.stack);
    console.error('Полная ошибка:', err);
    
    const errorMsg = err.message || err.toString() || 'Неизвестная ошибка';
    errorMessage.value = `Ошибка теста: ${errorMsg}`;
    notifications.addLog(`Ошибка теста: ${errorMsg}`, 'error');
  } finally {
    console.log('=== КОНЕЦ ТЕСТА БАКЕНДА ===');
    isAnalyzing.value = false;
  }
}

function applyToContext() {
  if (!analysisResult.value) return;
  
  // Добавляем выбранные файлы в контекст
  const selectedPaths = (analysisResult.value.selectedFiles || []).map((file: any) => file.relPath);
  const dependencyPaths = (analysisResult.value.dependencyFiles || []).map((file: any) => file.relPath);
  
  // Объединяем все пути
  const allPaths = [...selectedPaths, ...dependencyPaths];
  
  if (allPaths.length === 0) {
    notifications.addLog('Нет файлов для добавления в контекст', 'warn');
    return;
  }
  
  // Применяем к контексту
  contextBuilderStore.setSelectedFiles(allPaths);
  
  notifications.addLog(
    `Добавлено ${allPaths.length} файлов в контекст`, 
    'success'
  );
}

function copyContext() {
  if (!analysisResult.value?.context) {
    notifications.addLog('Нет контекста для копирования', 'warn');
    return;
  }
  
  navigator.clipboard.writeText(analysisResult.value.context);
  notifications.addLog('Контекст скопирован в буфер обмена', 'success');
}

function clearAnalysis() {
  task.value = '';
  analysisResult.value = null;
  errorMessage.value = '';
}

function formatDuration(durationMs: number): string {
  if (!durationMs) return '< 1с';
  const seconds = Math.round(durationMs / 1000);
  if (seconds < 1) return '< 1с';
  return `${seconds}с`;
}

function formatFileSize(bytes: number): string {
  if (!bytes) return '0 B';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// Инициализация
onMounted(() => {
  // Можно добавить загрузку сохраненных задач или других данных
});
</script>

<style scoped>
.intelligent-context-panel {
  max-height: 80vh;
  overflow-y: auto;
}
</style>
