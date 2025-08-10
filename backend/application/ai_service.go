package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
)

// ProviderFactory — это тип функции, которая выступает в роли фабрики для создания AI-провайдеров.
type ProviderFactory func(providerType, apiKey, modelName string) (domain.AIProvider, error)

// AIService - это фасад для всех операций, связанных с искусственным интеллектом.
type AIService struct {
	settings        *SettingsService
	log             domain.Logger
	providerFactory ProviderFactory
}

// NewAIService создает новый экземпляр AIService.
func NewAIService(settings *SettingsService, log domain.Logger, factory ProviderFactory) *AIService {
	return &AIService{
		settings:        settings,
		log:             log,
		providerFactory: factory,
	}
}

// GenerateCode выполняет генерацию кода, используя текущего сконфигурированного AI-провайдера.
func (s *AIService) GenerateCode(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	providerName := s.settings.GetSelectedAIProvider()
	if providerName == "" {
		providerName = "openai" // Default fallback
	}

	modelName := s.settings.GetSelectedModel(providerName)
	if modelName == "" {
		return "", fmt.Errorf("не выбрана модель для провайдера %s", providerName)
	}

	var apiKey string
	switch providerName {
	case "openai":
		apiKey = s.settings.GetOpenAIKey()
	case "gemini":
		apiKey = s.settings.GetGeminiKey()
	default:
		return "", fmt.Errorf("неизвестный AI провайдер: %s", providerName)
	}

	if apiKey == "" {
		return "", fmt.Errorf("API ключ для %s не настроен", providerName)
	}

	provider, err := s.providerFactory(providerName, apiKey, modelName)
	if err != nil {
		s.log.Error(fmt.Sprintf("Не удалось создать AI провайдер: %v", err))
		return "", err
	}

	// Формируем запрос с учетом ролей
	messages := []domain.Message{
		{Role: domain.RoleSystem, Content: systemPrompt},
		{Role: domain.RoleUser, Content: userPrompt},
	}

	request := domain.AIRequest{
		Messages:    messages,
		Model:       modelName,
		Temperature: 0.1,
		Options:     make(map[string]any), // Готовим место для специфичных опций
	}

	response, err := provider.Generate(ctx, request)
	if err != nil {
		return "", fmt.Errorf("ошибка при генерации кода: %w", err)
	}

	return response.Content, nil
}
