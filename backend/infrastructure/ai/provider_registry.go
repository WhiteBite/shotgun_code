package ai

import (
	"context"
	"shotgun_code/domain"
)

// ProviderConfig contains configuration for creating AI providers and model fetchers
type ProviderConfig struct {
	FactoryFunc  func(apiKey, host string, log domain.Logger) (domain.AIProvider, error)
	ModelFetcher func(ctx context.Context, apiKey, host string, log domain.Logger) ([]string, error)
}

// GetProviderRegistry returns a unified registry of all AI provider configurations
// This eliminates duplication between createModelFetchers and createProviderFactory
func GetProviderRegistry(openRouterHost string) map[string]ProviderConfig {
	return map[string]ProviderConfig{
		"gemini": {
			FactoryFunc: func(apiKey, host string, log domain.Logger) (domain.AIProvider, error) {
				return NewGemini(apiKey, host, log)
			},
			ModelFetcher: func(ctx context.Context, apiKey, host string, log domain.Logger) ([]string, error) {
				p, err := NewGemini(apiKey, host, log)
				if err != nil {
					return nil, err
				}
				return p.(*GeminiProviderImpl).ListModels(ctx)
			},
		},
		"openai": {
			FactoryFunc: func(apiKey, host string, log domain.Logger) (domain.AIProvider, error) {
				return NewOpenAI(apiKey, host, log)
			},
			ModelFetcher: func(ctx context.Context, apiKey, host string, log domain.Logger) ([]string, error) {
				p, err := NewOpenAI(apiKey, host, log)
				if err != nil {
					return nil, err
				}
				return p.(*OpenAIProviderImpl).ListModels(ctx)
			},
		},
		"openrouter": {
			FactoryFunc: func(apiKey, host string, log domain.Logger) (domain.AIProvider, error) {
				return NewOpenAI(apiKey, openRouterHost, log)
			},
			ModelFetcher: func(ctx context.Context, apiKey, host string, log domain.Logger) ([]string, error) {
				p, err := NewOpenAI(apiKey, openRouterHost, log)
				if err != nil {
					return nil, err
				}
				return p.(*OpenAIProviderImpl).ListModels(ctx)
			},
		},
		"localai": {
			FactoryFunc: func(apiKey, host string, log domain.Logger) (domain.AIProvider, error) {
				return NewLocalAI(apiKey, host, log)
			},
			ModelFetcher: func(ctx context.Context, apiKey, host string, log domain.Logger) ([]string, error) {
				p, err := NewLocalAI(apiKey, host, log)
				if err != nil {
					return nil, err
				}
				return p.(*LocalAIProviderImpl).ListModels(ctx)
			},
		},
	}
}