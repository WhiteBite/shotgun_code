package ai

import (
	"context"
	"errors"
	"fmt"
	"shotgun_code/domain"
	"sort"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GeminiProviderImpl struct {
	client *genai.GenerativeModel
	log    domain.Logger
	apiKey string
}

func NewGemini(apiKey, host string, log domain.Logger) (domain.AIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini API key is required")
	}
	// The host parameter is ignored for the official Gemini client, but kept for signature compatibility.
	return &GeminiProviderImpl{
		log:    log,
		apiKey: apiKey,
	}, nil
}

func (p *GeminiProviderImpl) ListModels(ctx context.Context) ([]string, error) {
	p.log.Info("Requesting model list from Gemini API...")
	client, err := genai.NewClient(ctx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client for model listing: %w", err)
	}
	defer client.Close()

	var models []string
	iter := client.ListModels(ctx)
	for {
		m, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			if strings.Contains(err.Error(), "API_KEY_INVALID") {
				return nil, domain.ErrInvalidAPIKey
			}
			return nil, fmt.Errorf("failed to iterate models: %w", err)
		}

		isSupported := false
		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				isSupported = true
				break
			}
		}
		if isSupported {
			models = append(models, m.Name)
		}
	}

	sort.Strings(models)
	p.log.Info(fmt.Sprintf("Received %d supported models from Gemini.", len(models)))
	return models, nil
}

func (p *GeminiProviderImpl) Generate(ctx context.Context, req domain.AIRequest) (domain.AIResponse, error) {
	p.log.Info(fmt.Sprintf("Sending request to Gemini API with model: %s", req.Model))

	client, err := genai.NewClient(ctx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return domain.AIResponse{}, fmt.Errorf("failed to create gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(req.Model)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(req.SystemPrompt)},
	}

	resp, err := model.GenerateContent(ctx, genai.Text(req.UserPrompt))
	if err != nil {
		p.log.Error(fmt.Sprintf("Gemini API request failed: %v", err))
		return domain.AIResponse{}, err
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return domain.AIResponse{}, fmt.Errorf("no content returned from Gemini API")
	}

	firstPart := resp.Candidates[0].Content.Parts[0]
	if text, ok := firstPart.(genai.Text); ok {
		return domain.AIResponse{Content: string(text)}, nil
	}

	return domain.AIResponse{}, fmt.Errorf("unsupported content type returned from Gemini: %T", firstPart)
}
