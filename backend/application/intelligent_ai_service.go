package application

import (
	"context"
	"fmt"
	"math/rand"
	"shotgun_code/domain"
	"strings"
	"sync"
	"time"
)

// IntelligentAIService предоставляет интеллектуальные возможности для работы с ИИ
type IntelligentAIService struct {
	settingsService *SettingsService
	log             domain.Logger
	providerGetter  domain.AIProviderGetter // Используем интерфейс для разрыва циклической зависимости
	rateLimiter     *RateLimiter
	metrics         *MetricsCollector
}

// NewIntelligentAIService создает новый интеллектуальный сервис ИИ
func NewIntelligentAIService(
	settingsService *SettingsService,
	log domain.Logger,
	providerRegistry map[string]domain.AIProviderFactory,
	rateLimiter *RateLimiter,
	metrics *MetricsCollector,
) *IntelligentAIService {
	return &IntelligentAIService{
		settingsService: settingsService,
		log:             log,
		providerGetter:  nil, // Will be set via SetProviderGetter after creation
		rateLimiter:     rateLimiter,
		metrics:         metrics,
	}
}

// SetProviderGetter устанавливает источник провайдеров для доступа к AI
func (s *IntelligentAIService) SetProviderGetter(getter domain.AIProviderGetter) {
	s.providerGetter = getter
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
	if s.providerGetter == nil {
		return nil, "", fmt.Errorf("provider getter not initialized")
	}

	// Используем интерфейс для получения провайдера
	provider, model, err := s.providerGetter.GetProvider(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Если модель не указана, пытаемся выбрать оптимальную
	if model == "" {
		if intelligentProvider, ok := provider.(domain.IntelligentAIProvider); ok {
			model, err = intelligentProvider.SelectOptimalModel(ctx, req.BaseRequest, req.Optimization.ModelSelection)
			if err != nil {
				return nil, "", fmt.Errorf("failed to select optimal model: %w", err)
			}
		} else {
			return nil, "", fmt.Errorf("no model selected for provider")
		}
	}

	return provider, model, nil
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
	// Fallback не поддерживается в упрощенной версии
	// В будущем можно добавить через AIService
	s.log.Warning("Fallback providers not supported in current implementation")
	return domain.AIResponse{}, fmt.Errorf("fallback not supported")
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
	return fmt.Sprintf("req_%d_%d", time.Now().Unix(), rand.Intn(1000)) //nolint:gosec // Not used for security
}

// RateLimiter implements token bucket rate limiting per provider
type RateLimiter struct {
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
	config  map[string]rateLimitConfig
}

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

type rateLimitConfig struct {
	tokensPerSecond float64
	maxTokens       float64
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*tokenBucket),
		config: map[string]rateLimitConfig{
			"openai":     {tokensPerSecond: 10, maxTokens: 60},
			"gemini":     {tokensPerSecond: 10, maxTokens: 60},
			"openrouter": {tokensPerSecond: 5, maxTokens: 30},
			"localai":    {tokensPerSecond: 100, maxTokens: 100}, // Local has no real limit
		},
	}
}

func (r *RateLimiter) CheckLimit(provider string) error {
	r.mu.RLock()
	bucket, exists := r.buckets[provider]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()
		// Double-check after acquiring write lock
		if bucket, exists = r.buckets[provider]; !exists {
			config := r.config[provider]
			if config.maxTokens == 0 {
				config = rateLimitConfig{tokensPerSecond: 10, maxTokens: 60}
			}
			bucket = &tokenBucket{
				tokens:     config.maxTokens,
				lastRefill: time.Now(),
			}
			r.buckets[provider] = bucket
		}
		r.mu.Unlock()
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens based on time elapsed
	config := r.config[provider]
	if config.maxTokens == 0 {
		config = rateLimitConfig{tokensPerSecond: 10, maxTokens: 60}
	}

	elapsed := time.Since(bucket.lastRefill).Seconds()
	bucket.tokens += elapsed * config.tokensPerSecond
	if bucket.tokens > config.maxTokens {
		bucket.tokens = config.maxTokens
	}
	bucket.lastRefill = time.Now()

	// Check if we have a token
	if bucket.tokens < 1 {
		return fmt.Errorf("rate limit exceeded for provider %s, please wait", provider)
	}

	bucket.tokens--
	return nil
}

// MetricsCollector сборщик метрик (thread-safe)
type MetricsCollector struct {
	generations []GenerationMetric
	mu          sync.RWMutex
	maxSize     int
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
		generations: make([]GenerationMetric, 0, 1000),
		maxSize:     1000,
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

	m.mu.Lock()
	defer m.mu.Unlock()

	// Limit size to prevent unbounded growth
	if len(m.generations) >= m.maxSize {
		// Remove oldest 10%
		m.generations = m.generations[m.maxSize/10:]
	}
	m.generations = append(m.generations, metric)
}

// GetMetrics returns aggregated metrics
func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalTokens := 0
	totalDuration := time.Duration(0)
	providerCounts := make(map[string]int)

	for _, gen := range m.generations {
		totalTokens += gen.TokensUsed
		totalDuration += gen.Duration
		providerCounts[gen.Provider]++
	}

	return map[string]interface{}{
		"total_generations": len(m.generations),
		"total_tokens":      totalTokens,
		"total_duration_ms": totalDuration.Milliseconds(),
		"by_provider":       providerCounts,
	}
}
