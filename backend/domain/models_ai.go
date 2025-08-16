package domain

// AIRequest представляет унифицированный запрос к AI провайдеру.
type AIRequest struct {
	Model        string
	SystemPrompt string
	UserPrompt   string
	// Дополнительные параметры, такие как температура, могут быть добавлены здесь.
}

// AIResponse представляет унифицированный ответ от AI провайдера.
type AIResponse struct {
	Content string
	// Дополнительные метаданные, такие как количество токенов, могут быть добавлены здесь.
}
