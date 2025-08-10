package domain

import "context"

// AIProvider определяет контракт для любого провайдера LLM.
// Это позволяет легко заменять OpenAI на Gemini, Claude или локальные модели.
type AIProvider interface {
	// Generate выполняет основной запрос к LLM для генерации контента.
	Generate(ctx context.Context, req AIRequest) (AIResponse, error)
}
