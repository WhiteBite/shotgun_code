package domain

const (
	RoleSystem = "system"
	RoleUser   = "user"
)

// Message представляет собой одно сообщение в диалоге с AI.
// Роль может быть "system" или "user".
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIRequest представляет собой стандартизированный запрос к AI-провайдеру.
type AIRequest struct {
	Messages    []Message
	Model       string
	Temperature float32
	Options     map[string]any // Для специфичных настроек провайдера, например, SafetySettings
}

// AIResponse представляет собой стандартизированный ответ от AI-провайдера.
type AIResponse struct {
	Content string
}
