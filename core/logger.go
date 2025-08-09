package core

// Logger определяет стандартный интерфейс для логирования, абстрагируя
// конкретную реализацию (например, логгер Wails). Это позволяет сервисам
// быть тестируемыми без зависимости от живого экземпляра Wails.
type Logger interface {
	Debug(message string)
	Info(message string)
	Warning(message string)
	Error(message string)
	Fatal(message string)
}
