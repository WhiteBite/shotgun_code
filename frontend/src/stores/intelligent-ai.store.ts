import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { apiService } from "@/infrastructure/api/api.service";
import { useNotificationsStore } from "./notifications.store";
import { useUiStore } from "./ui.store";

export interface IntelligentGenerationOptions {
  temperature: number;
  maxTokens: number;
  topP: number;
  priority: "low" | "normal" | "high" | "critical";
  timeout: number;
  maxRetries: number;
  autoOptimizePrompt: boolean;
  contextCompression: boolean;
  tokenOptimization: boolean;
  modelSelectionStrategy: "fastest" | "cheapest" | "best" | "balanced";
  enableFallback: boolean;
  fallbackModels: string[];
  fallbackProviders: string[];
  maxFallbackAttempts: number;
  performanceThreshold: number;
  projectType: string;
  codeStyle: string;
}

export interface IntelligentGenerationResult {
  content: string;
  modelUsed: string;
  tokensUsed: number;
  processingTime: number;
  qualityScore: number;
  suggestions: string[];
  warnings: string[];
  requestId: string;
  provider: string;
}

export interface ProviderInfo {
  name: string;
  version: string;
  capabilities: string[];
  limitations: string[];
  supportedModels: string[];
}

export const useIntelligentAIStore = defineStore("intelligentAI", () => {
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const lastResult = ref<IntelligentGenerationResult | null>(null);
  const providerInfo = ref<ProviderInfo | null>(null);
  const availableModels = ref<string[]>([]);

  const notifications = useNotificationsStore();
  const uiStore = useUiStore();

  // Состояние настроек
  const defaultOptions = ref<IntelligentGenerationOptions>({
    temperature: 0.7,
    maxTokens: 4000,
    topP: 1.0,
    priority: "normal",
    timeout: 60000,
    maxRetries: 3,
    autoOptimizePrompt: true,
    contextCompression: true,
    tokenOptimization: true,
    modelSelectionStrategy: "balanced",
    enableFallback: true,
    fallbackModels: [],
    fallbackProviders: ["openai", "gemini"],
    maxFallbackAttempts: 2,
    performanceThreshold: 30000,
    projectType: "",
    codeStyle: "",
  });

  const currentOptions = ref<IntelligentGenerationOptions>({
    ...defaultOptions.value,
  });

  // Computed properties
  const hasResult = computed(() => !!lastResult.value);
  const canGenerate = computed(() => !isLoading.value);

  // Actions
  async function generateIntelligentCode(
    task: string,
    context: string,
    options?: Partial<IntelligentGenerationOptions>,
  ) {
    if (!task.trim()) {
      notifications.addLog("Задача не может быть пустой", "error");
      return;
    }

    isLoading.value = true;
    error.value = null;
    lastResult.value = null;

    try {
      // Объединяем опции
      const finalOptions = { ...currentOptions.value, ...options };

      uiStore.setProgress({
        isActive: true,
        message: "Выполняется интеллектуальная генерация...",
        value: 0,
      });

      const resultJson = await apiService.generateIntelligentCode(
        task,
        context,
        JSON.stringify(finalOptions),
      );
      const result = JSON.parse(resultJson) as IntelligentGenerationResult;

      lastResult.value = result;

      notifications.addLog(
        `Интеллектуальная генерация завершена. Модель: ${result.modelUsed}, Токены: ${result.tokensUsed}`,
        "success",
      );

      // Показываем качество результата
      if (result.qualityScore < 0.7) {
        notifications.addLog(
          "Качество результата ниже ожидаемого. Рекомендуется проверить код.",
          "warning",
        );
      }

      // Показываем предложения если есть
      if (result.suggestions.length > 0) {
        result.suggestions.forEach((suggestion) => {
          notifications.addLog(`Предложение: ${suggestion}`, "info");
        });
      }
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : "An unknown error occurred";
      error.value = `Ошибка интеллектуальной генерации: ${errorMessage}`;
      notifications.addLog(error.value, "error");
    } finally {
      isLoading.value = false;
      uiStore.clearProgress();
    }
  }

  async function generateCodeWithOptions(
    systemPrompt: string,
    userPrompt: string,
    options?: Partial<IntelligentGenerationOptions>,
  ) {
    if (!userPrompt.trim()) {
      notifications.addLog("Промпт не может быть пустым", "error");
      return;
    }

    isLoading.value = true;
    error.value = null;

    try {
      const finalOptions = { ...currentOptions.value, ...options };

      uiStore.setProgress({
        isActive: true,
        message: "Генерация кода с опциями...",
        value: 0,
      });

      const result = await apiService.generateCodeWithOptions(
        systemPrompt,
        userPrompt,
        JSON.stringify(finalOptions),
      );

      lastResult.value = {
        content: result,
        modelUsed: "unknown",
        tokensUsed: 0,
        processingTime: 0,
        qualityScore: 0.8,
        suggestions: [],
        warnings: [],
        requestId: "",
        provider: "unknown",
      };

      notifications.addLog("Генерация кода завершена успешно", "success");
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : "An unknown error occurred";
      error.value = `Ошибка генерации: ${errorMessage}`;
      notifications.addLog(error.value, "error");
    } finally {
      isLoading.value = false;
      uiStore.clearProgress();
    }
  }

  async function loadProviderInfo() {
    try {
      const infoJson = await apiService.getProviderInfo();
      providerInfo.value = JSON.parse(infoJson) as ProviderInfo;
      notifications.addLog(
        `Загружена информация о провайдере: ${providerInfo.value.name}`,
        "info",
      );
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : "An unknown error occurred";
      notifications.addLog(
        `Ошибка загрузки информации о провайдере: ${errorMessage}`,
        "error",
      );
    }
  }

  async function loadAvailableModels() {
    try {
      availableModels.value = await apiService.listAvailableModels();
      notifications.addLog(
        `Загружено ${availableModels.value.length} доступных моделей`,
        "info",
      );
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : "An unknown error occurred";
      notifications.addLog(`Ошибка загрузки моделей: ${errorMessage}`, "error");
    }
  }

  function updateOptions(options: Partial<IntelligentGenerationOptions>) {
    currentOptions.value = { ...currentOptions.value, ...options };
  }

  function resetOptions() {
    currentOptions.value = { ...defaultOptions.value };
  }

  function clearResult() {
    lastResult.value = null;
    error.value = null;
  }

  function getQualityColor(score: number): string {
    if (score >= 0.8) return "text-green-600";
    if (score >= 0.6) return "text-yellow-600";
    return "text-red-600";
  }

  function getQualityText(score: number): string {
    if (score >= 0.8) return "Отличное";
    if (score >= 0.6) return "Хорошее";
    return "Требует внимания";
  }

  return {
    // State
    isLoading,
    error,
    lastResult,
    providerInfo,
    availableModels,
    currentOptions,
    defaultOptions,

    // Computed
    hasResult,
    canGenerate,

    // Actions
    generateIntelligentCode,
    generateCodeWithOptions,
    loadProviderInfo,
    loadAvailableModels,
    updateOptions,
    resetOptions,
    clearResult,
    getQualityColor,
    getQualityText,
  };
});