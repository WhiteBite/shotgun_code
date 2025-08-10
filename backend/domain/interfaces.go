package domain

import "context"

// Logger определяет стандартный интерфейс для логирования.
type Logger interface {
	Debug(message string)
	Info(message string)
	Warning(message string)
	Error(message string)
	Fatal(message string)
}

// EventBus определяет интерфейс для отправки событий во внешний мир (например, на фронтенд).
type EventBus interface {
	Emit(eventName string, data ...interface{})
}

// FileContentReader provides an abstraction for reading file contents.
type FileContentReader interface {
	// ReadContents reads multiple files and returns their content.
	// It also sends progress updates via the provided callback.
	ReadContents(
		ctx context.Context,
		filePaths []string,
		rootDir string,
		progress func(current, total int64),
	) (map[string]string, error)
}

// SettingsRepository определяет интерфейс для работы с хранилищем настроек.
type SettingsRepository interface {
	GetCustomIgnoreRules() string
	GetCustomPromptRules() string
	SetCustomIgnoreRules(rules string) error
	SetCustomPromptRules(rules string) error

	// OpenAI
	GetOpenAIKey() string
	SetOpenAIKey(key string) error

	// Gemini
	GetGeminiKey() string
	SetGeminiKey(key string) error

	// LocalAI
	GetLocalAIKey() string
	SetLocalAIKey(key string) error
	GetLocalAIHost() string
	SetLocalAIHost(host string) error
	GetLocalAIModelName() string
	SetLocalAIModelName(name string) error

	// General AI Settings
	GetSelectedAIProvider() string
	SetSelectedAIProvider(provider string) error
	GetSelectedModel(provider string) string
	SetSelectedModel(provider string, model string) error
	GetModels(provider string) []string
	SetModels(provider string, models []string) error
}

// GitRepository определяет интерфейс для взаимодействия с Git.
type GitRepository interface {
	CheckAvailability() (bool, error)
	GetUncommittedFiles(projectRoot string) ([]string, error)
	GetRichCommitHistory(projectRoot, branchName string, limit int) ([]CommitWithFiles, error)
}

// FileTreeBuilder определяет интерфейс для построения дерева файлов.
type FileTreeBuilder interface {
	BuildTree(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*FileNode, error)
}

// FileSystemWatcher определяет интерфейс для наблюдения за изменениями в файловой системе.
type FileSystemWatcher interface {
	Start(path string) error
	Stop()
	RefreshAndRescan() error
}
