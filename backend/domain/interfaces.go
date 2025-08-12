package domain

import "context"

type Logger interface {
	Debug(message string)
	Info(message string)
	Warning(message string)
	Error(message string)
	Fatal(message string)
}
type EventBus interface {
	Emit(eventName string, data ...interface{})
}
type FileContentReader interface {
	ReadContents(
		ctx context.Context,
		filePaths []string,
		rootDir string,
		progress func(current, total int64),
	) (map[string]string, error)
}
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
type GitRepository interface {
	CheckAvailability() (bool, error)
	GetUncommittedFiles(projectRoot string) ([]FileStatus, error)
	GetRichCommitHistory(projectRoot, branchName string, limit int) ([]CommitWithFiles, error)
	GetFileContentAtCommit(projectRoot, filePath, commitHash string) (string, error)
	GetGitignoreContent(projectRoot string) (string, error)
}
type FileTreeBuilder interface {
	BuildTree(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*FileNode, error)
}
type FileSystemWatcher interface {
	Start(path string) error
	Stop()
	RefreshAndRescan() error
}
