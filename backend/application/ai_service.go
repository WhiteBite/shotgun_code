package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
)

type ProviderFactory func(providerType, apiKey, modelName string) (domain.AIProvider, error)

type AIService struct {
	settingsService *SettingsService
	log             domain.Logger
	providerFactory ProviderFactory
}

func NewAIService(settings *SettingsService, log domain.Logger, factory ProviderFactory) *AIService {
	return &AIService{
		settingsService: settings,
		log:             log,
		providerFactory: factory,
	}
}

func (s *AIService) GenerateCode(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	settingsDTO, err := s.settingsService.GetSettingsDTO()
	if err != nil {
		return "", fmt.Errorf("could not get settings for generation: %w", err)
	}

	providerName := settingsDTO.SelectedProvider
	if providerName == "" {
		return "", fmt.Errorf("no AI provider selected")
	}
	modelName := settingsDTO.SelectedModels[providerName]

	keyMap := map[string]string{
		"openai":     settingsDTO.OpenAIAPIKey,
		"gemini":     settingsDTO.GeminiAPIKey,
		"openrouter": settingsDTO.OpenRouterAPIKey,
		"localai":    settingsDTO.LocalAIAPIKey,
	}
	apiKey := keyMap[providerName]

	if providerName == "localai" && (modelName == "" || modelName == "local-model") {
		modelName = settingsDTO.LocalAIModelName
	}

	provider, err := s.providerFactory(providerName, apiKey, modelName)
	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to create AI provider: %v", err))
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	messages := []domain.Message{
		{Role: domain.RoleSystem, Content: systemPrompt},
		{Role: domain.RoleUser, Content: userPrompt},
	}

	request := domain.AIRequest{
		Messages:    messages,
		Model:       modelName,
		Temperature: 0.1,
	}

	response, err := provider.Generate(ctx, request)
	if err != nil {
		return "", fmt.Errorf("error during code generation: %w", err)
	}

	return response.Content, nil
}
