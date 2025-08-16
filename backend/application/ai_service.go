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
	settingsService *SettingsService
	log             domain.Logger
	providerFactory ProviderFactory
}

// NewAIService creates a new AIService.
func NewAIService(
	settingsService *SettingsService,
	log domain.Logger,
	providerFactory ProviderFactory,
) *AIService {
	return &AIService{
		settingsService: settingsService,
		log:             log,
		providerFactory: providerFactory,
	}
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
	}

	tctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := provider.Generate(tctx, req)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	return resp.Content, nil
}
