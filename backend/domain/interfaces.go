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
	SetCustomIgnoreRules(rules string)
	GetCustomPromptRules() string
	SetCustomPromptRules(rules string)

	GetUseGitignore() bool
	SetUseGitignore(enabled bool)
	GetUseCustomIgnore() bool
	SetUseCustomIgnore(enabled bool)

	GetOpenAIKey() string
	SetOpenAIKey(key string)
	GetGeminiKey() string
	SetGeminiKey(key string)
	GetOpenRouterKey() string
	SetOpenRouterKey(key string)
	GetLocalAIKey() string
	SetLocalAIKey(key string)
	GetLocalAIHost() string
	SetLocalAIHost(host string)
	GetLocalAIModelName() string
	SetLocalAIModelName(name string)

	GetSelectedAIProvider() string
	SetSelectedAIProvider(provider string)
	GetSelectedModel(provider string) string
	SetSelectedModel(provider string, model string)
	GetModels(provider string) []string
	SetModels(provider string, models []string)

	Save() error
}

// GitRepository определяет интерфейс для взаимодействия с Git.
type GitRepository interface {
	CheckAvailability() (bool, error)
	GetUncommittedFiles(projectRoot string) ([]FileStatus, error)
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
