package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"time"
)

// AIService orchestrates AI-related operations.
type AIService struct {
	settingsService    *SettingsService
	log                domain.Logger
	providerFactory    domain.AIProviderFactory
	intelligentService *IntelligentAIService
}

// NewAIService creates a new AIService.
func NewAIService(
	settingsService *SettingsService,
	log domain.Logger,
	providerFactory domain.AIProviderFactory,
	intelligentService *IntelligentAIService,
) *AIService {
	service := &AIService{
		settingsService:    settingsService,
		log:                log,
		providerFactory:    providerFactory,
		intelligentService: intelligentService,
	}

	return service
}

// generateCodeInternal is a helper method that encapsulates the common logic for code generation
func (s *AIService) generateCodeInternal(ctx context.Context, systemPrompt, userPrompt string, options *CodeGenerationOptions) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	provider, model, err := s.getProvider(ctx)
	if err != nil {
		return "", err
	}

	// Set default values
	temperature := 0.7
	maxTokens := 4000
	topP := 1.0
	timeout := 60 * time.Second
	priority := domain.PriorityNormal

	// Apply options if provided
	if options != nil {
		if options.Model != "" {
			model = options.Model
		}
		if options.Temperature != 0 {
			temperature = options.Temperature
		}
		if options.MaxTokens != 0 {
			maxTokens = options.MaxTokens
		}
		if options.TopP != 0 {
			topP = options.TopP
		}
		if options.Timeout != 0 {
			timeout = options.Timeout
		}
		if options.Priority != domain.PriorityLow { // Check if priority is not the zero value
			priority = options.Priority
		}
	}

	req := domain.AIRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  temperature,
		MaxTokens:    maxTokens,
		TopP:         topP,
		RequestID:    fmt.Sprintf("req_%d", time.Now().Unix()),
		Priority:     priority,
		Timeout:      timeout,
	}

	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := provider.Generate(tctx, req)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

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
// Устраняет дублирование в GenerateCode, GenerateCodeWithOptions, GetProviderInfo и ListAvailableModels
func (s *AIService) getProvider(ctx context.Context) (domain.AIProvider, string, error) {
	dto, err := s.settingsService.GetSettingsDTO()
	if err != nil {
		return nil, "", fmt.Errorf("could not get settings: %w", err)
	}

	providerType := dto.SelectedProvider
	if providerType == "" {
		return nil, "", fmt.Errorf("no AI provider selected")
	}

	// Get API key for the selected provider
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

	// LocalAI might not require a key, so we don't check for empty key for it
	if apiKey == "" && providerType != "localai" {
		return nil, "", fmt.Errorf("API key for %s is not set", providerType)
	}

	// Create provider instance
	provider, err := s.providerFactory(providerType, apiKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create provider %s: %w", providerType, err)
	}

	// Get model from settings
	model := dto.SelectedModels[providerType]
	if model == "" {
		// Try to use first available model as fallback
		models := dto.AvailableModels[providerType]
		if len(models) > 0 {
			model = models[0]
		}
	}

	if model == "" {
		return nil, "", fmt.Errorf("no model selected for provider %s", providerType)
	}

	return provider, model, nil
}

// GetIntelligentService returns the intelligent AI service for advanced operations
func (s *AIService) GetIntelligentService() *IntelligentAIService {
	return s.intelligentService
}

