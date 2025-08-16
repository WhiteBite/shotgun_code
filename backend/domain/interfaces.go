package domain

import "context"

// Logger определяет интерфейс для логирования
type Logger interface {
	Debug(message string)
	Info(message string)
	Warning(message string)
	Error(message string)
	Fatal(message string)
}

// NoopLogger provides a dummy implementation of domain.Logger for cases where a logger
// is required but full logging infrastructure might not be available or desired.
type NoopLogger struct{}

func (l *NoopLogger) Debug(message string)   {}
func (l *NoopLogger) Info(message string)    {}
func (l *NoopLogger) Warning(message string) {}
func (l *NoopLogger) Error(message string)   {}
func (l *NoopLogger) Fatal(message string)   {}

// EventBus определяет интерфейс для событийной шины
type EventBus interface {
	Emit(eventName string, data ...interface{})
}

// TreeBuilder определяет интерфейс для построения дерева файлов
type TreeBuilder interface {
	BuildTree(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*FileNode, error)
}

// FileContentReader определяет интерфейс для чтения содержимого файлов
type FileContentReader interface {
	ReadContents(
		ctx context.Context,
		filePaths []string,
		rootDir string,
		progress func(current, total int64),
	) (map[string]string, error)
}

// GitRepository определяет интерфейс для работы с Git
type GitRepository interface {
	GetUncommittedFiles(projectRoot string) ([]FileStatus, error)
	GetRichCommitHistory(projectRoot, branchName string, limit int) ([]CommitWithFiles, error)
	GetFileContentAtCommit(projectRoot, filePath, commitHash string) (string, error)
	GetGitignoreContent(projectRoot string) (string, error)
	IsGitAvailable() bool
}

// SettingsRepository определяет интерфейс для работы с настройками
type SettingsRepository interface {
	GetCustomIgnoreRules() string
	SetCustomIgnoreRules(rules string)
	GetCustomPromptRules() string
	SetCustomPromptRules(rules string)
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
	SetSelectedModel(provider, model string)
	GetModels(provider string) []string
	SetModels(provider string, models []string)
	GetUseGitignore() bool
	SetUseGitignore(use bool)
	GetUseCustomIgnore() bool
	SetUseCustomIgnore(use bool)
	Save() error
	GetSettingsDTO() (SettingsDTO, error) // Added as per compilation error
}

// FileSystemWatcher определяет интерфейс для отслеживания изменений файловой системы
type FileSystemWatcher interface {
	Start(rootPath string) error
	Stop()
	RefreshAndRescan() error
}

// ContextSplitter определяет интерфейс для разбиения большого контекста на части.
type ContextSplitter interface {
	// SplitContext разбивает текст контекста на части, учитывая лимиты токенов и стратегии.
	SplitContext(ctxText string, settings SplitSettings) ([]string, error)
}
