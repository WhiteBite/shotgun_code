package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"time"
)

// ProviderFactory is a function type that creates an AIProvider.
type ProviderFactory func(providerType, apiKey string) (domain.AIProvider, error)

// AIService orchestrates AI-related operations.
type AIService struct {
	settingsService    *SettingsService
	log                domain.Logger
	providerFactory    ProviderFactory
	intelligentService *IntelligentAIService
}

// NewAIService creates a new AIService.
func NewAIService(
	settingsService *SettingsService,
	log domain.Logger,
	providerFactory ProviderFactory,
) *AIService {
	service := &AIService{
		settingsService: settingsService,
		log:             log,
		providerFactory: providerFactory,
	}

	// Создаем интеллектуальный сервис
	service.intelligentService = NewIntelligentAIService(settingsService, log, providerFactory)

	return service
}

// GenerateCode selects the appropriate AI provider and generates code.
func (s *AIService) GenerateCode(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dto, err := s.settingsService.GetSettingsDTO()
	if err != nil {
		return "", fmt.Errorf("could not get settings for AI generation: %w", err)
	}

	providerType := dto.SelectedProvider
	if providerType == "" {
		return "", fmt.Errorf("no AI provider selected")
	}

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

	// LocalAI might not require a key, so we don't check for empty key for it here.
	if apiKey == "" && providerType != "localai" {
		return "", fmt.Errorf("API key for %s is not set", providerType)
	}

	provider, err := s.providerFactory(providerType, apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider '%s': %w", providerType, err)
	}

	model := dto.SelectedModels[providerType]
	if model == "" {
		return "", fmt.Errorf("no model selected for provider %s", providerType)
	}

	req := domain.AIRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    4000,
		TopP:         1.0,
		RequestID:    fmt.Sprintf("req_%d", time.Now().Unix()),
		Priority:     domain.PriorityNormal,
		Timeout:      60 * time.Second,
	}

	tctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := provider.Generate(tctx, req)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	return resp.Content, nil
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
	if ctx == nil {
		ctx = context.Background()
	}

	dto, err := s.settingsService.GetSettingsDTO()
	if err != nil {
		return "", fmt.Errorf("could not get settings for AI generation: %w", err)
	}

	providerType := dto.SelectedProvider
	if providerType == "" {
		return "", fmt.Errorf("no AI provider selected")
	}

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
		return "", fmt.Errorf("API key for %s is not set", providerType)
	}

	provider, err := s.providerFactory(providerType, apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider '%s': %w", providerType, err)
	}

	model := dto.SelectedModels[providerType]
	if model == "" {
		return "", fmt.Errorf("no model selected for provider %s", providerType)
	}

	// Применяем опции
	if options.Model != "" {
		model = options.Model
	}
	if options.Temperature == 0 {
		options.Temperature = 0.7
	}
	if options.MaxTokens == 0 {
		options.MaxTokens = 4000
	}
	if options.TopP == 0 {
		options.TopP = 1.0
	}
	if options.Timeout == 0 {
		options.Timeout = 60 * time.Second
	}

	req := domain.AIRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  options.Temperature,
		MaxTokens:    options.MaxTokens,
		TopP:         options.TopP,
		RequestID:    fmt.Sprintf("req_%d", time.Now().Unix()),
		Priority:     options.Priority,
		Timeout:      options.Timeout,
	}

	tctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	resp, err := provider.Generate(tctx, req)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	return resp.Content, nil
}

// GetProviderInfo возвращает информацию о текущем провайдере
func (s *AIService) GetProviderInfo(ctx context.Context) (*domain.ProviderInfo, error) {
	dto, err := s.settingsService.GetSettingsDTO()
	if err != nil {
		return nil, fmt.Errorf("could not get settings: %w", err)
	}

	providerType := dto.SelectedProvider
	if providerType == "" {
		return nil, fmt.Errorf("no AI provider selected")
	}

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

	provider, err := s.providerFactory(providerType, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider '%s': %w", providerType, err)
	}

	info := provider.GetProviderInfo()
	return &info, nil
}

// ListAvailableModels возвращает список доступных моделей
func (s *AIService) ListAvailableModels(ctx context.Context) ([]string, error) {
	dto, err := s.settingsService.GetSettingsDTO()
	if err != nil {
		return nil, fmt.Errorf("could not get settings: %w", err)
	}

	providerType := dto.SelectedProvider
	if providerType == "" {
		return nil, fmt.Errorf("no AI provider selected")
	}

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

	provider, err := s.providerFactory(providerType, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider '%s': %w", providerType, err)
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
