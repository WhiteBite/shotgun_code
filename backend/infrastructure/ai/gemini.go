package ai

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"sort"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GeminiProviderImpl является конкретной реализацией интерфейса domain.AIProvider для Google Gemini.
type GeminiProviderImpl struct {
	// Используем genai.Client для доступа к методу ListModels
	client *genai.Client
	model  *genai.GenerativeModel
	log    domain.Logger
}

// NewGemini создает новый экземпляр провайдера Gemini.
func NewGemini(apiKey string, modelName string, log domain.Logger) (domain.AIProvider, error) {
	if apiKey == "" {
		log.Warning("API ключ для Gemini не предоставлен. Провайдер создан, но не сможет выполнять запросы.")
		return nil, domain.ErrProviderNotConfigured
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента Gemini: %w", err)
	}

	model := client.GenerativeModel(modelName)

	return &GeminiProviderImpl{
		client: client,
		model:  model,
		log:    log,
	}, nil
}

// ListModels запрашивает у API Gemini список доступных моделей.
func (p *GeminiProviderImpl) ListModels(ctx context.Context) ([]string, error) {
	p.log.Info("Запрос списка моделей от Gemini API...")
	iter := p.client.ListModels(ctx)
	var models []string

	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			p.log.Error(fmt.Sprintf("Ошибка при получении списка моделей Gemini: %v", err))
			return nil, err
		}

		// Корректная проверка: ищем "generateContent" в списке поддерживаемых методов.
		isSupported := false
		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				isSupported = true
				break
			}
		}

		if isSupported {
			// Извлекаем имя модели из полного пути, например, "models/gemini-1.5-pro-latest"
			parts := strings.Split(m.Name, "/")
			if len(parts) == 2 {
				models = append(models, parts[1])
			}
		}
	}

	// Сортируем для консистентности
	sort.Strings(models)
	p.log.Info(fmt.Sprintf("Получено %d моделей от Gemini.", len(models)))
	return models, nil
}

// Generate выполняет запрос к API Gemini, адаптируясь к доменной модели.
func (p *GeminiProviderImpl) Generate(ctx context.Context, req domain.AIRequest) (domain.AIResponse, error) {
	p.log.Info(fmt.Sprintf("Отправка запроса к Gemini с моделью: %s", req.Model))

	var systemInstruction *genai.Content

	var userMessages []string
	for _, msg := range req.Messages {
		if msg.Role == domain.RoleSystem {
			systemInstruction = &genai.Content{
				Parts: []genai.Part{genai.Text(msg.Content)},
			}
		} else if msg.Role == domain.RoleUser {
			userMessages = append(userMessages, msg.Content)
		}
	}

	p.model.SystemInstruction = systemInstruction
	finalPrompt := strings.Join(userMessages, "\n\n")

	resp, err := p.model.GenerateContent(ctx, genai.Text(finalPrompt))
	if err != nil {
		p.log.Error(fmt.Sprintf("Ошибка от API Gemini: %v", err))
		// Проверяем gRPC статус ошибки
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated, codes.PermissionDenied:
				return domain.AIResponse{}, domain.ErrInvalidAPIKey
			case codes.NotFound:
				return domain.AIResponse{}, domain.ErrModelNotFound
			case codes.ResourceExhausted:
				return domain.AIResponse{}, domain.ErrRateLimitExceeded
			}
		}
		return domain.AIResponse{}, err
	}

	// Обработка пустого ответа
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		// Проверим, был ли ответ заблокирован
		if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].FinishReason == genai.FinishReasonSafety {
			p.log.Warning("Ответ от Gemini заблокирован из-за настроек безопасности.")
			return domain.AIResponse{}, fmt.Errorf("ответ заблокирован политикой безопасности Gemini")
		}
		p.log.Warning("Ответ от Gemini не содержит валидного контента.")
		return domain.AIResponse{}, fmt.Errorf("пустой или некорректный ответ от Gemini")
	}

	// Собираем контент из частей
	var contentBuilder strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			contentBuilder.WriteString(string(txt))
		}
	}
	content := contentBuilder.String()

	p.log.Info("Успешно получен ответ от Gemini.")
	return domain.AIResponse{Content: content}, nil
}
