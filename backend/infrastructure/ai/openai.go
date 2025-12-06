package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"shotgun_code/domain"
	"sort"
	"time"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProviderImpl struct {
	client *openai.Client
	log    domain.Logger
}

func NewOpenAI(apiKey, host string, log domain.Logger) (domain.AIProvider, error) {
	config := openai.DefaultConfig(apiKey)
	if host != "" {
		config.BaseURL = host
	}
	client := openai.NewClientWithConfig(config)
	return &OpenAIProviderImpl{
		client: client,
		log:    log,
	}, nil
}

func (p *OpenAIProviderImpl) ListModels(ctx context.Context) ([]string, error) {
	p.log.Info("Requesting model list from OpenAI compatible API...")
	resp, err := p.client.ListModels(ctx)
	if err != nil {
		p.log.Error(fmt.Sprintf("Error getting model list: %v", err))
		var apiErr *openai.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.HTTPStatusCode {
			case http.StatusUnauthorized:
				return nil, domain.ErrInvalidAPIKey
			case http.StatusTooManyRequests:
				return nil, domain.ErrRateLimitExceeded
			}
		}
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var models []string
	for _, model := range resp.Models {
		models = append(models, model.ID)
	}
	sort.Strings(models)
	p.log.Info(fmt.Sprintf("Received %d models.", len(models)))
	return models, nil
}

func (p *OpenAIProviderImpl) Generate(ctx context.Context, req domain.AIRequest) (domain.AIResponse, error) {
	startTime := time.Now()
	p.log.Info(fmt.Sprintf("Sending request to OpenAI compatible API with model: %s", req.Model))

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: req.UserPrompt,
		},
	}

	// Применяем параметры запроса
	completionReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		TopP:        float32(req.TopP),
	}

	// Добавляем frequency_penalty и presence_penalty если они установлены
	if req.FrequencyPenalty != 0 {
		completionReq.FrequencyPenalty = float32(req.FrequencyPenalty)
	}
	if req.PresencePenalty != 0 {
		completionReq.PresencePenalty = float32(req.PresencePenalty)
	}

	resp, err := p.client.CreateChatCompletion(ctx, completionReq)

	if err != nil {
		p.log.Error(fmt.Sprintf("OpenAI API request failed: %v", err))
		return domain.AIResponse{}, err
	}

	if len(resp.Choices) == 0 {
		return domain.AIResponse{}, fmt.Errorf("no choices returned from OpenAI API")
	}

	processingTime := time.Since(startTime)

	// Подсчитываем токены (примерная оценка)
	tokensUsed := resp.Usage.TotalTokens

	return domain.AIResponse{
		Content:        resp.Choices[0].Message.Content,
		TokensUsed:     tokensUsed,
		ModelUsed:      req.Model,
		ProcessingTime: processingTime,
		FinishReason:   string(resp.Choices[0].FinishReason),
		Confidence:     0.9, // Базовая оценка
	}, nil
}

func (p *OpenAIProviderImpl) GetProviderInfo() domain.ProviderInfo {
	return domain.ProviderInfo{
		Name:            "OpenAI",
		Version:         "1.0",
		Capabilities:    []string{"chat", "completion", "embeddings"},
		Limitations:     []string{"rate_limited", "token_limited"},
		SupportedModels: []string{"gpt-4", "gpt-3.5-turbo", "gpt-4-turbo"},
	}
}

func (p *OpenAIProviderImpl) ValidateRequest(req domain.AIRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	if req.UserPrompt == "" {
		return fmt.Errorf("user prompt is required")
	}
	if req.Temperature < 0 || req.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	if req.MaxTokens < 1 {
		return fmt.Errorf("max tokens must be greater than 0")
	}
	return nil
}

func (p *OpenAIProviderImpl) EstimateTokens(req domain.AIRequest) (int, error) {
	// Простая оценка токенов (примерно 4 символа на токен)
	totalChars := len(req.SystemPrompt) + len(req.UserPrompt)
	estimatedTokens := totalChars / 4

	// Добавляем буфер для безопасности
	return estimatedTokens + 100, nil
}

func (p *OpenAIProviderImpl) GetPricing(model string) domain.PricingInfo {
	// Базовая информация о стоимости OpenAI
	pricing := domain.PricingInfo{
		Model:    model,
		Currency: "USD",
	}

	switch model {
	case "gpt-4":
		pricing.InputTokensPer1K = 0.03
		pricing.OutputTokensPer1K = 0.06
	case "gpt-4-turbo":
		pricing.InputTokensPer1K = 0.01
		pricing.OutputTokensPer1K = 0.03
	case "gpt-3.5-turbo":
		pricing.InputTokensPer1K = 0.0015
		pricing.OutputTokensPer1K = 0.002
	default:
		pricing.InputTokensPer1K = 0.01
		pricing.OutputTokensPer1K = 0.02
	}

	return pricing
}

// GenerateStream sends a streaming request to OpenAI API
func (p *OpenAIProviderImpl) GenerateStream(ctx context.Context, req domain.AIRequest, onChunk func(chunk domain.StreamChunk)) error {
	p.log.Info(fmt.Sprintf("Starting streaming request to OpenAI API with model: %s", req.Model))

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: req.UserPrompt,
		},
	}

	completionReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		TopP:        float32(req.TopP),
		Stream:      true,
	}

	if req.FrequencyPenalty != 0 {
		completionReq.FrequencyPenalty = float32(req.FrequencyPenalty)
	}
	if req.PresencePenalty != 0 {
		completionReq.PresencePenalty = float32(req.PresencePenalty)
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, completionReq)
	if err != nil {
		p.log.Error(fmt.Sprintf("OpenAI API stream request failed: %v", err))
		onChunk(domain.StreamChunk{Done: true, Error: err.Error()})
		return err
	}
	defer stream.Close()

	totalTokens := 0
	for {
		response, err := stream.Recv()
		if errors.Is(err, context.Canceled) {
			onChunk(domain.StreamChunk{Done: true, Error: "Request cancelled"})
			return err
		}
		if err != nil {
			if err.Error() == "EOF" {
				onChunk(domain.StreamChunk{Done: true, TokensUsed: totalTokens, FinishReason: "stop"})
				return nil
			}
			p.log.Error(fmt.Sprintf("Stream error: %v", err))
			onChunk(domain.StreamChunk{Done: true, Error: err.Error()})
			return err
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			finishReason := string(response.Choices[0].FinishReason)

			if content != "" {
				totalTokens += len(content) / 4
				onChunk(domain.StreamChunk{
					Content: content,
					Done:    false,
				})
			}

			if finishReason == "stop" || finishReason == "length" {
				onChunk(domain.StreamChunk{
					Done:         true,
					TokensUsed:   totalTokens,
					FinishReason: finishReason,
				})
				return nil
			}
		}
	}
}
