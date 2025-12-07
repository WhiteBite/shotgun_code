package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"shotgun_code/domain"
	"sync"
	"sync/atomic"
	"time"
)

// AIService orchestrates AI-related operations with caching and connection pooling.
type AIService struct {
	settingsService    *SettingsService
	log                domain.Logger
	providerRegistry   map[string]domain.AIProviderFactory
	intelligentService *IntelligentAIService

	// Provider cache - reuse providers instead of creating new ones
	providerCache   map[string]domain.AIProvider
	providerCacheMu sync.RWMutex

	// Response cache for identical requests (LRU-like)
	responseCache   map[string]*cachedAIResponse
	responseCacheMu sync.RWMutex

	// Metrics
	totalRequests   int64
	cacheHits       int64
	cacheMisses     int64
	totalTokensUsed int64

	// Graceful shutdown
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

type cachedAIResponse struct {
	content   string
	timestamp time.Time
	tokens    int
}

// AI Service configuration constants
const (
	// Cache settings
	maxResponseCacheSize = 100
	responseCacheTTL     = 30 * time.Minute
	cacheCleanupInterval = 5 * time.Minute

	// Default generation parameters
	defaultTemperature   = 0.7
	defaultMaxTokens     = 4000
	defaultTopP          = 1.0
	defaultTimeout       = 60 * time.Second
	defaultStreamTimeout = 120 * time.Second

	// Cache behavior thresholds
	deterministicTempThreshold = 0.1 // Disable cache above this temperature
)

// NewAIService creates a new AIService.
func NewAIService(
	settingsService *SettingsService,
	log domain.Logger,
	providerRegistry map[string]domain.AIProviderFactory,
	intelligentService *IntelligentAIService,
) *AIService {
	service := &AIService{
		settingsService:    settingsService,
		log:                log,
		providerRegistry:   providerRegistry,
		intelligentService: intelligentService,
		providerCache:      make(map[string]domain.AIProvider),
		responseCache:      make(map[string]*cachedAIResponse),
		stopCh:             make(chan struct{}),
	}

	// Start cache cleanup goroutine
	service.wg.Add(1)
	go service.cleanupResponseCache()

	return service
}

// Shutdown gracefully stops the AI service
func (s *AIService) Shutdown(ctx context.Context) error {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("AIService shutdown complete")
	case <-ctx.Done():
		s.log.Warning("AIService shutdown timed out")
		return ctx.Err()
	}

	// Clear caches
	s.responseCacheMu.Lock()
	s.responseCache = make(map[string]*cachedAIResponse)
	s.responseCacheMu.Unlock()

	s.providerCacheMu.Lock()
	s.providerCache = make(map[string]domain.AIProvider)
	s.providerCacheMu.Unlock()

	return nil
}

// cleanupResponseCache periodically removes expired cache entries
func (s *AIService) cleanupResponseCache() {
	defer s.wg.Done()

	ticker := time.NewTicker(cacheCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.performCacheCleanup()
		}
	}
}

// performCacheCleanup removes expired and excess cache entries
func (s *AIService) performCacheCleanup() {
	s.responseCacheMu.Lock()
	defer s.responseCacheMu.Unlock()

	s.removeExpiredEntries()
	s.enforceMaxCacheSize()
}

// removeExpiredEntries removes entries older than TTL
func (s *AIService) removeExpiredEntries() {
	now := time.Now()
	for key, entry := range s.responseCache {
		if now.Sub(entry.timestamp) > responseCacheTTL {
			delete(s.responseCache, key)
		}
	}
}

// enforceMaxCacheSize removes oldest entries to stay within size limit
func (s *AIService) enforceMaxCacheSize() {
	for len(s.responseCache) > maxResponseCacheSize {
		oldestKey := s.findOldestCacheEntry()
		if oldestKey == "" {
			break
		}
		delete(s.responseCache, oldestKey)
	}
}

// findOldestCacheEntry finds the key of the oldest cache entry
func (s *AIService) findOldestCacheEntry() string {
	oldest := time.Now()
	oldestKey := ""
	for key, entry := range s.responseCache {
		if entry.timestamp.Before(oldest) {
			oldest = entry.timestamp
			oldestKey = key
		}
	}
	return oldestKey
}

// getCacheKey generates a hash key for caching including generation parameters
func (s *AIService) getCacheKey(systemPrompt, userPrompt, model string, temperature float64, maxTokens int, topP float64) string {
	h := sha256.New()
	h.Write([]byte(systemPrompt))
	h.Write([]byte(userPrompt))
	h.Write([]byte(model))
	h.Write([]byte(fmt.Sprintf("%.2f:%d:%.2f", temperature, maxTokens, topP)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// GetMetrics returns AI service metrics
func (s *AIService) GetMetrics() map[string]interface{} {
	s.responseCacheMu.RLock()
	cacheSize := len(s.responseCache)
	s.responseCacheMu.RUnlock()

	s.providerCacheMu.RLock()
	providerCount := len(s.providerCache)
	s.providerCacheMu.RUnlock()

	return map[string]interface{}{
		"total_requests":      atomic.LoadInt64(&s.totalRequests),
		"cache_hits":          atomic.LoadInt64(&s.cacheHits),
		"cache_misses":        atomic.LoadInt64(&s.cacheMisses),
		"total_tokens_used":   atomic.LoadInt64(&s.totalTokensUsed),
		"response_cache_size": cacheSize,
		"cached_providers":    providerCount,
	}
}

// generationParams holds parameters for code generation
type generationParams struct {
	model       string
	temperature float64
	maxTokens   int
	topP        float64
	timeout     time.Duration
	priority    domain.RequestPriority
	useCache    bool
}

// applyOptions applies options to generation params
func applyOptions(params *generationParams, options *CodeGenerationOptions) {
	if options == nil {
		return
	}
	if options.Model != "" {
		params.model = options.Model
	}
	if options.Temperature != 0 {
		params.temperature = options.Temperature
		if params.temperature > deterministicTempThreshold {
			params.useCache = false
		}
	}
	if options.MaxTokens != 0 {
		params.maxTokens = options.MaxTokens
	}
	if options.TopP != 0 {
		params.topP = options.TopP
	}
	if options.Timeout != 0 {
		params.timeout = options.Timeout
	}
	if options.Priority != domain.PriorityLow {
		params.priority = options.Priority
	}
}

// checkCache checks response cache and returns cached content if valid
func (s *AIService) checkCache(cacheKey string, useCache bool) (string, bool) {
	if !useCache {
		return "", false
	}
	s.responseCacheMu.RLock()
	defer s.responseCacheMu.RUnlock()
	if cached, ok := s.responseCache[cacheKey]; ok && time.Since(cached.timestamp) < responseCacheTTL {
		atomic.AddInt64(&s.cacheHits, 1)
		return cached.content, true
	}
	return "", false
}

// generateCodeInternal is a helper method that encapsulates the common logic for code generation
func (s *AIService) generateCodeInternal(ctx context.Context, systemPrompt, userPrompt string, options *CodeGenerationOptions) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	atomic.AddInt64(&s.totalRequests, 1)

	provider, model, err := s.getProvider(ctx)
	if err != nil {
		return "", err
	}

	params := &generationParams{
		model: model, temperature: defaultTemperature, maxTokens: defaultMaxTokens,
		topP: defaultTopP, timeout: defaultTimeout, priority: domain.PriorityNormal, useCache: true,
	}
	applyOptions(params, options)

	cacheKey := s.getCacheKey(systemPrompt, userPrompt, params.model, params.temperature, params.maxTokens, params.topP)
	if content, found := s.checkCache(cacheKey, params.useCache); found {
		return content, nil
	}
	atomic.AddInt64(&s.cacheMisses, 1)

	req := domain.AIRequest{
		Model:        params.model,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  params.temperature,
		MaxTokens:    params.maxTokens,
		TopP:         params.topP,
		RequestID:    fmt.Sprintf("req_%d", time.Now().UnixNano()),
		Priority:     params.priority,
		Timeout:      params.timeout,
	}

	tctx, cancel := context.WithTimeout(ctx, params.timeout)
	defer cancel()

	resp, err := provider.Generate(tctx, req)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	// Cache the response for deterministic requests
	if params.useCache && resp.Content != "" {
		s.responseCacheMu.Lock()
		s.responseCache[cacheKey] = &cachedAIResponse{
			content:   resp.Content,
			timestamp: time.Now(),
			tokens:    resp.TokensUsed,
		}
		s.responseCacheMu.Unlock()
	}

	atomic.AddInt64(&s.totalTokensUsed, int64(resp.TokensUsed))

	return resp.Content, nil
}

// GenerateCode selects the appropriate AI provider and generates code.
func (s *AIService) GenerateCode(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return s.generateCodeInternal(ctx, systemPrompt, userPrompt, nil)
}

// GenerateIntelligentCode использует интеллектуальную систему для генерации кода
func (s *AIService) GenerateIntelligentCode(
	ctx context.Context,
	task string,
	context string,
	options IntelligentGenerationOptions,
) (*IntelligentGenerationResult, error) {
	return s.intelligentService.GenerateIntelligentCode(ctx, task, context, options)
}

// GenerateCodeWithOptions генерирует код с дополнительными опциями
func (s *AIService) GenerateCodeWithOptions(
	ctx context.Context,
	systemPrompt, userPrompt string,
	options CodeGenerationOptions,
) (string, error) {
	return s.generateCodeInternal(ctx, systemPrompt, userPrompt, &options)
}

// GetProviderInfo возвращает информацию о текущем провайдере
func (s *AIService) GetProviderInfo(ctx context.Context) (*domain.ProviderInfo, error) {
	provider, _, err := s.getProvider(ctx)
	if err != nil {
		return nil, err
	}

	info := provider.GetProviderInfo()
	return &info, nil
}

// ListAvailableModels возвращает список доступных моделей
func (s *AIService) ListAvailableModels(ctx context.Context) ([]string, error) {
	provider, _, err := s.getProvider(ctx)
	if err != nil {
		return nil, err
	}

	return provider.ListModels(ctx)
}

// CodeGenerationOptions опции для генерации кода
type CodeGenerationOptions struct {
	Model       string
	Temperature float64
	MaxTokens   int
	TopP        float64
	Priority    domain.RequestPriority
	Timeout     time.Duration
}

// getProvider инкапсулирует логику получения AI-провайдера и модели
// Использует кэширование провайдеров для повторного использования соединений
func (s *AIService) getProvider(ctx context.Context) (domain.AIProvider, string, error) {
	dto, err := s.settingsService.GetSettingsDTO()
	if err != nil {
		return nil, "", fmt.Errorf("could not get settings: %w", err)
	}

	providerType := dto.SelectedProvider
	if providerType == "" {
		return nil, "", fmt.Errorf("no AI provider selected")
	}

	// Get API key based on provider type
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
	case "qwen":
		apiKey = dto.QwenAPIKey
	case "qwen-cli":
		// Qwen CLI doesn't require API key - uses local CLI
		apiKey = ""
	default:
		return nil, "", fmt.Errorf("unsupported AI provider: %s", providerType)
	}

	// LocalAI and qwen-cli don't require a key
	if apiKey == "" && providerType != "localai" && providerType != "qwen-cli" {
		return nil, "", fmt.Errorf("API key for %s is not set", providerType)
	}

	// Create cache key based on provider type and API key hash
	cacheKey := fmt.Sprintf("%s:%s", providerType, s.hashKey(apiKey))

	// Check provider cache first
	s.providerCacheMu.RLock()
	if cachedProvider, ok := s.providerCache[cacheKey]; ok {
		s.providerCacheMu.RUnlock()
		model := s.getModelForProvider(dto, providerType)
		return cachedProvider, model, nil
	}
	s.providerCacheMu.RUnlock()

	// Create new provider
	factory, exists := s.providerRegistry[providerType]
	if !exists {
		return nil, "", fmt.Errorf("no factory registered for provider %s", providerType)
	}

	provider, err := factory(providerType, apiKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create provider %s: %w", providerType, err)
	}

	// Cache the provider
	s.providerCacheMu.Lock()
	s.providerCache[cacheKey] = provider
	s.providerCacheMu.Unlock()

	model := s.getModelForProvider(dto, providerType)
	return provider, model, nil
}

// hashKey creates a short hash of the API key for cache key
func (s *AIService) hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])[:8]
}

// getModelForProvider returns the selected model for a provider
func (s *AIService) getModelForProvider(dto domain.SettingsDTO, providerType string) string {
	model, ok := dto.SelectedModels[providerType]
	if !ok || model == "" {
		// Try to use first available model as fallback
		models, ok := dto.AvailableModels[providerType]
		if ok && len(models) > 0 {
			model = models[0]
			s.log.Warning(fmt.Sprintf("No model selected for %s, falling back to: %s", providerType, model))
		}
	}
	return model
}

// InvalidateProviderCache clears the provider cache (useful when settings change)
func (s *AIService) InvalidateProviderCache() {
	s.providerCacheMu.Lock()
	s.providerCache = make(map[string]domain.AIProvider)
	s.providerCacheMu.Unlock()
	s.log.Info("Provider cache invalidated")
}

// GetIntelligentService returns the intelligent AI service for advanced operations
func (s *AIService) GetIntelligentService() *IntelligentAIService {
	return s.intelligentService
}

// GetProvider возвращает текущий провайдер и модель (публичный метод для IntelligentAIService)
func (s *AIService) GetProvider(ctx context.Context) (domain.AIProvider, string, error) {
	return s.getProvider(ctx)
}

// GenerateCodeStream generates code with streaming response
func (s *AIService) GenerateCodeStream(ctx context.Context, systemPrompt, userPrompt string, onChunk func(chunk domain.StreamChunk)) error {
	if ctx == nil {
		ctx = context.Background()
	}

	atomic.AddInt64(&s.totalRequests, 1)

	provider, model, err := s.getProvider(ctx)
	if err != nil {
		onChunk(domain.StreamChunk{Done: true, Error: err.Error()})
		return err
	}

	req := domain.AIRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  defaultTemperature,
		MaxTokens:    defaultMaxTokens,
		TopP:         defaultTopP,
		RequestID:    fmt.Sprintf("stream_%d", time.Now().UnixNano()),
		Priority:     domain.PriorityNormal,
		Timeout:      defaultStreamTimeout,
	}

	tctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	return provider.GenerateStream(tctx, req, onChunk)
}
