package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"shotgun_code/domain"
	"sort"

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

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    req.Model,
			Messages: messages,
		},
	)

	if err != nil {
		p.log.Error(fmt.Sprintf("OpenAI API request failed: %v", err))
		return domain.AIResponse{}, err
	}

	if len(resp.Choices) == 0 {
		return domain.AIResponse{}, fmt.Errorf("no choices returned from OpenAI API")
	}

	return domain.AIResponse{Content: resp.Choices[0].Message.Content}, nil
}
