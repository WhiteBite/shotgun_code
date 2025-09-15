package application

import (
	"context"
	"fmt"
	"math/rand"
	"shotgun_code/domain"
	"strings"
	"time"
)

// IntelligentAIService предоставляет интеллектуальные возможности для работы с ИИ
type IntelligentAIService struct {
	settingsService *SettingsService
	log             domain.Logger
	providerFactory domain.AIProviderFactory
	providers       map[string]domain.AIProvider
	rateLimiter     *RateLimiter
	metrics         *MetricsCollector
}

// NewIntelligentAIService создает новый интеллектуальный сервис ИИ
func NewIntelligentAIService(
	settingsService *SettingsService,
	log domain.Logger,
	providerFactory domain.AIProviderFactory,
	rateLimiter *RateLimiter,
	metrics *MetricsCollector,
) *IntelligentAIService {
	return &IntelligentAIService{
		settingsService: settingsService,
		log:             log,
		providerFactory: providerFactory,
		providers:       make(map[string]domain.AIProvider),
		rateLimiter:     rateLimiter,
		metrics:         metrics,
	}
}

// GenerateIntelligentCode выполняет интеллектуальную генерацию кода
func (s *IntelligentAIService) GenerateIntelligentCode(
	ctx context.Context,
	task string,
	context string,
	options IntelligentGenerationOptions,
) (*IntelligentGenerationResult, error) {
	startTime := time.Now()

	// Получаем настройки
	dto, err := s.settingsService.GetSettingsDTO()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Создаем интеллектуальный запрос
	intelligentReq := s.buildIntelligentRequest(task, context, dto, options)

	// Выбираем оптимальный провайдер и модель
	provider, model, err := s.selectOptimalProvider(ctx, intelligentReq, dto)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}

	// Оптимизируем промпт
	optimizedPrompt, err := s.optimizePrompt(ctx, task, context, options)
	if err != nil {
		s.log.Warning(fmt.Sprintf("Failed to optimize prompt, using original: %v", err))
		optimizedPrompt = task
	}

	// Создаем базовый запрос
	req := domain.AIRequest{
		Model:        model,
		SystemPrompt: s.buildSystemPrompt(options),
		UserPrompt:   optimizedPrompt,
		Temperature:  options.Temperature,
		MaxTokens:    options.MaxTokens,
		TopP:         options.TopP,
		RequestID:    generateRequestID(),
		Priority:     options.Priority,
		Timeout:      options.Timeout,
	}

	// Проверяем rate limit
	if err := s.rateLimiter.CheckLimit(provider.GetProviderInfo().Name); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Выполняем запрос с retry логикой
	var response domain.AIResponse
	var lastErr error

	for attempt := 0; attempt <= options.MaxRetries; attempt++ {
		if attempt > 0 {
			s.log.Info(fmt.Sprintf("Retry attempt %d for request %s", attempt, req.RequestID))
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		response, lastErr = provider.Generate(ctx, req)
		if lastErr == nil {
			break
		}

		// Если это последняя попытка, пробуем fallback
		if attempt == options.MaxRetries && options.EnableFallback {
			response, lastErr = s.tryFallback(ctx, req, dto)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all generation attempts failed: %w", lastErr)
	}

	// Анализируем ответ
	analysis, err := s.analyzeResponse(ctx, response, req)
	if err != nil {
		s.log.Warning(fmt.Sprintf("Failed to analyze response: %v", err))
	}

	// Собираем метрики
	s.metrics.RecordGeneration(provider.GetProviderInfo().Name, model, time.Since(startTime), response.TokensUsed)

	result := &IntelligentGenerationResult{
		Content:        response.Content,
		ModelUsed:      response.ModelUsed,
		TokensUsed:     response.TokensUsed,
		ProcessingTime: response.ProcessingTime,
		QualityScore:   analysis.QualityScore,
		Suggestions:    analysis.Suggestions,
		Warnings:       response.Warnings,
		RequestID:      req.RequestID,
		Provider:       provider.GetProviderInfo().Name,
	}

	return result, nil
}

// buildIntelligentRequest создает интеллектуальный запрос
func (s *IntelligentAIService) buildIntelligentRequest(
	task string,
	context string,
	dto domain.SettingsDTO,
	options IntelligentGenerationOptions,
) domain.IntelligentRequest {
	return domain.IntelligentRequest{
		BaseRequest: domain.AIRequest{
			UserPrompt: task,
			Priority:   options.Priority,
			Timeout:    options.Timeout,
		},
		Optimization: domain.OptimizationConfig{
			AutoOptimizePrompt: options.AutoOptimizePrompt,
			ContextCompression: options.ContextCompression,
			TokenOptimization:  options.TokenOptimization,
			ModelSelection:     options.ModelSelectionStrategy,
		},
		FallbackConfig: domain.FallbackConfig{
			EnableFallback:      options.EnableFallback,
			FallbackModels:      options.FallbackModels,
			FallbackProviders:   options.FallbackProviders,
			MaxFallbackAttempts: options.MaxFallbackAttempts,
		},
		Monitoring: domain.MonitoringConfig{
			EnableMetrics:        true,
			EnableLogging:        true,
			PerformanceThreshold: options.PerformanceThreshold,
		},
	}
}

// selectOptimalProvider выбирает оптимальный провайдер и модель
func (s *IntelligentAIService) selectOptimalProvider(
	ctx context.Context,
	req domain.IntelligentRequest,
	dto domain.SettingsDTO,
) (domain.AIProvider, string, error) {
	// Получаем или создаем провайдер
	providerType := dto.SelectedProvider
	if providerType == "" {
		return nil, "", fmt.Errorf("no AI provider selected")
	}

	provider, exists := s.providers[providerType]
	if !exists {
		var err error
		provider, err = s.createProvider(providerType, dto)
		if err != nil {
			return nil, "", err
		}
		s.providers[providerType] = provider
	}

	// Выбираем модель
	model := dto.SelectedModels[providerType]
	if model == "" {
		// Пытаемся выбрать оптимальную модель
		if intelligentProvider, ok := provider.(domain.IntelligentAIProvider); ok {
			model, err := intelligentProvider.SelectOptimalModel(ctx, req.BaseRequest, req.Optimization.ModelSelection)
			if err != nil {
				return nil, "", fmt.Errorf("failed to select optimal model: %w", err)
			}
			return provider, model, nil
		}
		return nil, "", fmt.Errorf("no model selected for provider %s", providerType)
	}

	return provider, model, nil
}

// createProvider создает провайдер
func (s *IntelligentAIService) createProvider(providerType string, dto domain.SettingsDTO) (domain.AIProvider, error) {
	var apiKey string
	switch providerType {
	case "openai":
		apiKey = dto.OpenAIAPIKey
	case "gemini":
		apiKey = dto.GeminiAPIKey
	case "openrouter":
		apiKey = dto.OpenRouterAPIKey
	case "localai":
		apiKey = dto.LocalAIAPIKey
	}

	if apiKey == "" && providerType != "localai" {
		return nil, fmt.Errorf("API key for %s is not set", providerType)
	}

	return s.providerFactory(providerType, apiKey)
}

// optimizePrompt оптимизирует промпт
func (s *IntelligentAIService) optimizePrompt(
	ctx context.Context,
	task string,
	context string,
	options IntelligentGenerationOptions,
) (string, error) {
	if !options.AutoOptimizePrompt {
		return task, nil
	}

	// Простая оптимизация промпта
	optimized := strings.TrimSpace(task)

	// Добавляем контекст если он есть
	if context != "" {
		optimized = fmt.Sprintf("Context:\n%s\n\nTask:\n%s", context, optimized)
	}

	// Добавляем специфичные инструкции
	if options.ProjectType != "" {
		optimized = fmt.Sprintf("%s\n\nPlease provide solution optimized for %s project.", optimized, options.ProjectType)
	}

	return optimized, nil
}

// buildSystemPrompt создает системный промпт
func (s *IntelligentAIService) buildSystemPrompt(options IntelligentGenerationOptions) string {
	basePrompt := "You are an expert software developer. Your task is to implement the user's request by providing the necessary code changes in the form of a standard git diff. Do not include any explanations, comments, or apologies outside of the `git diff` block."

	if options.ProjectType != "" {
		basePrompt += fmt.Sprintf(" Focus on %s best practices.", options.ProjectType)
	}

	if options.CodeStyle != "" {
		basePrompt += fmt.Sprintf(" Follow %s coding style.", options.CodeStyle)
	}

	return basePrompt
}

// tryFallback пытается использовать fallback провайдеры
func (s *IntelligentAIService) tryFallback(
	ctx context.Context,
	req domain.AIRequest,
	dto domain.SettingsDTO,
) (domain.AIResponse, error) {
	fallbackProviders := []string{"openai", "gemini", "openrouter"}

	for _, providerType := range fallbackProviders {
		if providerType == dto.SelectedProvider {
			continue // Пропускаем основной провайдер
		}

		provider, err := s.createProvider(providerType, dto)
		if err != nil {
			s.log.Warning(fmt.Sprintf("Failed to create fallback provider %s: %v", providerType, err))
			continue
		}

		// Пробуем с fallback моделью
		fallbackModel := dto.SelectedModels[providerType]
		if fallbackModel == "" {
			continue
		}

		req.Model = fallbackModel
		response, err := provider.Generate(ctx, req)
		if err == nil {
			s.log.Info(fmt.Sprintf("Successfully used fallback provider %s", providerType))
			return response, nil
		}

		s.log.Warning(fmt.Sprintf("Fallback provider %s failed: %v", providerType, err))
	}

	return domain.AIResponse{}, fmt.Errorf("all fallback providers failed")
}

// analyzeResponse анализирует ответ
func (s *IntelligentAIService) analyzeResponse(
	ctx context.Context,
	response domain.AIResponse,
	req domain.AIRequest,
) (domain.ResponseAnalysis, error) {
	// Простой анализ ответа
	analysis := domain.ResponseAnalysis{
		QualityScore:      0.8, // Базовая оценка
		RelevanceScore:    0.8,
		CompletenessScore: 0.8,
		Confidence:        0.8,
	}

	// Проверяем наличие git diff
	if strings.Contains(response.Content, "diff --git") {
		analysis.QualityScore += 0.1
		analysis.CompletenessScore += 0.1
	}

	// Проверяем длину ответа
	if len(response.Content) > 100 {
		analysis.CompletenessScore += 0.1
	}

	// Добавляем базовые предложения
	analysis.Suggestions = []string{
		"Review the generated code before applying",
		"Test the changes in a development environment",
	}

	return analysis, nil
}

// IntelligentGenerationOptions опции для интеллектуальной генерации
type IntelligentGenerationOptions struct {
	Temperature            float64
	MaxTokens              int
	TopP                   float64
	Priority               domain.RequestPriority
	Timeout                time.Duration
	MaxRetries             int
	AutoOptimizePrompt     bool
	ContextCompression     bool
	TokenOptimization      bool
	ModelSelectionStrategy domain.ModelSelectionStrategy
	EnableFallback         bool
	FallbackModels         []string
	FallbackProviders      []string
	MaxFallbackAttempts    int
	PerformanceThreshold   time.Duration
	ProjectType            string
	CodeStyle              string
}

// IntelligentGenerationResult результат интеллектуальной генерации
type IntelligentGenerationResult struct {
	Content        string
	ModelUsed      string
	TokensUsed     int
	ProcessingTime time.Duration
	QualityScore   float64
	Suggestions    []string
	Warnings       []string
	RequestID      string
	Provider       string
}

// Вспомогательные функции
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().Unix(), rand.Intn(1000))
}

// RateLimiter простой rate limiter
type RateLimiter struct {
	limits map[string]*time.Ticker
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*time.Ticker),
	}
}

func (r *RateLimiter) CheckLimit(provider string) error {
	// Простая реализация - в реальном проекте нужно использовать более сложную логику
	return nil
}

// MetricsCollector сборщик метрик
type MetricsCollector struct {
	generations []GenerationMetric
}

type GenerationMetric struct {
	Provider   string
	Model      string
	Duration   time.Duration
	TokensUsed int
	Timestamp  time.Time
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		generations: make([]GenerationMetric, 0),
	}
}

func (m *MetricsCollector) RecordGeneration(provider, model string, duration time.Duration, tokens int) {
	metric := GenerationMetric{
		Provider:   provider,
		Model:      model,
		Duration:   duration,
		TokensUsed: tokens,
		Timestamp:  time.Now(),
	}
	m.generations = append(m.generations, metric)
}