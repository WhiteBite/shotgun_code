package domain

import "errors"

// Определяем стандартные доменные ошибки для AI-операций.
// Это позволяет нам обрабатывать их единообразно на верхних уровнях приложения.
var (
	ErrInvalidAPIKey         = errors.New("неверный или недействительный API ключ")
	ErrModelNotFound         = errors.New("указанная модель не найдена или недоступна")
	ErrRateLimitExceeded     = errors.New("превышен лимит запросов к API")
	ErrProviderNotConfigured = errors.New("провайдер не настроен")
)
