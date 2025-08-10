package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"shotgun_code/domain"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProviderImpl является конкретной реализацией интерфейса domain.AIProvider для OpenAI.
type OpenAIProviderImpl struct {
	client *openai.Client
	log    domain.Logger
}

// NewOpenAI создает новый экземпляр провайдера OpenAI.
// Принимает опциональный hostURL для работы с локальными OpenAI-совместимыми серверами.
func NewOpenAI(apiKey, hostURL string, log domain.Logger) (domain.AIProvider, error) {
	config := openai.DefaultConfig(apiKey)
	if hostURL != "" {
		config.BaseURL = hostURL
		log.Info("OpenAI клиент настроен на использование кастомного хоста: " + hostURL)
	}

	client := openai.NewClientWithConfig(config)

	return &OpenAIProviderImpl{
		client: client,
		log:    log,
	}, nil
}

// Generate выполняет запрос к API OpenAI, адаптируясь к новой доменной модели.
func (p *OpenAIProviderImpl) Generate(ctx context.Context, req domain.AIRequest) (domain.AIResponse, error) {
	p.log.Info(fmt.Sprintf("Отправка запроса к OpenAI-совместимому API с моделью: %s", req.Model))

	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       req.Model,
			Temperature: req.Temperature,
			Messages:    messages,
		},
	)

	if err != nil {
		p.log.Error(fmt.Sprintf("Ошибка от API OpenAI: %v", err))
		var apiErr *openai.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.HTTPStatusCode {
			case http.StatusUnauthorized:
				return domain.AIResponse{}, domain.ErrInvalidAPIKey
			case http.StatusNotFound:
				return domain.AIResponse{}, domain.ErrModelNotFound
			case http.StatusTooManyRequests:
				return domain.AIResponse{}, domain.ErrRateLimitExceeded
			}
		}
		return domain.AIResponse{}, err
	}

	if len(resp.Choices) == 0 {
		p.log.Warning("Ответ от OpenAI не содержит вариантов (choices).")
		return domain.AIResponse{}, fmt.Errorf("пустой ответ от OpenAI")
	}

	content := resp.Choices[0].Message.Content
	p.log.Info("Успешно получен ответ от OpenAI-совместимого API.")

	return domain.AIResponse{Content: content}, nil
}
