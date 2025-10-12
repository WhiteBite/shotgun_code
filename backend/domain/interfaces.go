package domain

import (
	"context"
	"io/fs"
	"time"
)

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
	GetBranches(projectRoot string) ([]string, error)
	GetCurrentBranch(projectRoot string) (string, error)
	GetAllFiles(projectPath string) ([]string, error)
	GenerateDiff(projectPath string) (string, error)
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
	GetRecentProjects() []RecentProjectInfo
	AddRecentProject(path, name string)
	RemoveRecentProject(path string)

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

// CommentStripper определяет интерфейс для удаления комментариев из исходного кода
type CommentStripper interface {
	// Strip удаляет комментарии из содержимого файла, опираясь на расширение filePath
	Strip(content string, filePath string) string
}

// AIProviderFactory is a function type that creates an AIProvider.
type AIProviderFactory func(providerType, apiKey string) (AIProvider, error)

// ModelFetcher is a function type that fetches available models for a provider.
type ModelFetcher func(apiKey string) ([]string, error)

// ModelFetcherRegistry is a map of provider types to their model fetchers.
type ModelFetcherRegistry map[string]ModelFetcher

// PathProvider определяет интерфейс для работы с путями файловой системы
type PathProvider interface {
	// Join соединяет элементы пути
	Join(elem ...string) string

	// Base возвращает последний элемент пути
	Base(path string) string

	// Dir возвращает все элементы пути кроме последнего
	Dir(path string) string

	// IsAbs проверяет, является ли путь абсолютным
	IsAbs(path string) bool

	// Clean возвращает очищенную версию пути
	Clean(path string) string

	// Getwd возвращает текущую рабочую директорию
	Getwd() (string, error)
}

// FileSystemWriter определяет интерфейс для записи файлов
type FileSystemWriter interface {
	// WriteFile записывает данные в файл
	WriteFile(filename string, data []byte, perm int) error

	// MkdirAll создает директорию вместе со всеми родительскими директориями
	MkdirAll(path string, perm int) error

	// Remove удаляет файл или директорию
	Remove(name string) error

	// RemoveAll удаляет файл или директорию и все её содержимое
	RemoveAll(path string) error
}

// OPAService определяет интерфейс для работы с OPA политиками
type OPAService interface {
	// ValidatePath проверяет путь через OPA политики
	ValidatePath(path string) (*OPAValidationResult, error)

	// ValidateBudget проверяет бюджет через OPA политики
	ValidateBudget(budgetType string, current, limit int64) (*OPAValidationResult, error)

	// ValidateTask проверяет задачу через OPA политики
	ValidateTask(taskID string, files []string, linesChanged int64, ephemeralMode bool) (*OPAValidationResult, error)

	// ValidateConfig проверяет конфигурацию через OPA политики
	ValidateConfig(config GuardrailConfig) (*OPAValidationResult, error)
}

// FileStatProvider определяет интерфейс для получения информации о файлах
type FileStatProvider interface {
	// Stat возвращает информацию о файле
	Stat(name string) (FileInfo, error)
}

// FileInfo описывает файл и возвращается Stat
type FileInfo interface {
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() FileMode     // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() interface{}   // underlying data source (can return nil)
}

// FileMode represents a file's mode and permission bits.
type FileMode = fs.FileMode

// TempFileProvider определяет интерфейс для работы с временными файлами
type TempFileProvider interface {
	// MkdirTemp создает временную директорию
	MkdirTemp(dir, pattern string) (string, error)
}

// CommandRunner определяет интерфейс для выполнения команд
type CommandRunner interface {
	// RunCommand выполняет команду с заданным контекстом и аргументами
	RunCommand(ctx context.Context, name string, args ...string) ([]byte, error)

	// RunCommandInDir выполняет команду в указанной директории
	RunCommandInDir(ctx context.Context, dir, name string, args ...string) ([]byte, error)
}

// Task Protocol Verification System Interfaces

// TaskProtocolService defines the interface for task protocol verification
type TaskProtocolService interface {
	// ExecuteProtocol executes the full verification protocol for a task
	ExecuteProtocol(ctx context.Context, config *TaskProtocolConfig) (*TaskProtocolResult, error)

	// ValidateStage executes a single verification stage
	ValidateStage(ctx context.Context, stage ProtocolStage, config *TaskProtocolConfig) (*ProtocolStageResult, error)

	// RequestCorrectionGuidance requests AI-generated correction guidance for errors
	RequestCorrectionGuidance(ctx context.Context, error *ErrorDetails, context *TaskContext) (*CorrectionGuidance, error)
}

// ErrorAnalyzer defines the interface for analyzing verification errors
type ErrorAnalyzer interface {
	// AnalyzeError analyzes error output and provides detailed error information
	AnalyzeError(errorOutput string, stage ProtocolStage) (*ErrorDetails, error)

	// SuggestCorrections provides correction suggestions for a given error
	SuggestCorrections(error *ErrorDetails) ([]*CorrectionStep, error)

	// ClassifyErrorType determines the type of error from output
	ClassifyErrorType(errorOutput string) ErrorType
}

// CorrectionEngine defines the interface for applying corrections
type CorrectionEngine interface {
	// ApplyCorrection applies a single correction step
	ApplyCorrection(ctx context.Context, step *CorrectionStep, projectPath string) (*CorrectionResult, error)

	// ApplyCorrections applies multiple correction steps
	ApplyCorrections(ctx context.Context, steps []*CorrectionStep, projectPath string) (*CorrectionResult, error)

	// CanHandle checks if the engine can handle a specific error type
	CanHandle(error *ErrorDetails) bool
}

// LanguageErrorAnalyzer defines language-specific error analysis
type LanguageErrorAnalyzer interface {
	// AnalyzeError analyzes error output for a specific language
	AnalyzeError(errorOutput string) (*ErrorDetails, error)

	// SuggestCorrections provides language-specific correction suggestions
	SuggestCorrections(error *ErrorDetails) ([]*CorrectionStep, error)

	// ClassifyErrorType determines the error type for the language
	ClassifyErrorType(errorOutput string) ErrorType

	// GetLanguage returns the supported language
	GetLanguage() string
}

// CorrectionRule defines rules for applying corrections
type CorrectionRule interface {
	// CanHandle checks if this rule can handle the given error
	CanHandle(error *ErrorDetails) bool

	// ApplyCorrection applies the correction rule
	ApplyCorrection(error *ErrorDetails, projectPath string) (*CorrectionResult, error)

	// GetPriority returns the priority of this rule (higher = more important)
	GetPriority() int

	// GetErrorTypes returns the error types this rule can handle
	GetErrorTypes() []ErrorType
}

// FileSystemProvider defines file system operations for task protocol
type FileSystemProvider interface {
	// ReadFile reads the contents of a file
	ReadFile(filename string) ([]byte, error)

	// WriteFile writes data to a file
	WriteFile(filename string, data []byte, perm int) error

	// MkdirAll creates directories
	MkdirAll(path string, perm int) error
}

// ContextBuilder определяет интерфейс для построения контекста
type ContextBuilder interface {
	// BuildContext builds a context from project files and returns a ContextSummary to prevent OOM issues
	BuildContext(ctx context.Context, projectPath string, includedPaths []string, options *ContextBuildOptions) (*ContextSummary, error)

	// BuildContextLegacy builds context with legacy format (DEPRECATED - can cause OOM)
	BuildContextLegacy(ctx context.Context, projectPath string, includedPaths []string, options ContextBuildOptions) (*Context, error)

	// GenerateContext builds a context asynchronously with progress tracking
	GenerateContext(ctx context.Context, rootDir string, includedPaths []string)
}

// ContextStreamer определяет интерфейс для стриминга контекста
type ContextStreamer interface {
	// CreateStreamingContext creates a streaming context from project files
	CreateStreamingContext(ctx context.Context, projectPath string, includedPaths []string, options *ContextBuildOptions) (*ContextStream, error)

	// GetContextLines retrieves a range of lines from a streaming context
	GetContextLines(ctx context.Context, contextID string, startLine, endLine int64) (*ContextLineRange, error)

	// GetContextContent returns paginated context content for memory-safe viewing
	GetContextContent(ctx context.Context, contextID string, startLine int, lineCount int) (interface{}, error)
}

// ContextRepository определяет интерфейс для хранения контекста
type ContextRepository interface {
	// GetContext retrieves a context by ID
	GetContext(ctx context.Context, contextID string) (*Context, error)

	// GetProjectContexts lists all contexts for a project
	GetProjectContexts(ctx context.Context, projectPath string) ([]*Context, error)

	// DeleteContext deletes a context by ID
	DeleteContext(ctx context.Context, contextID string) error

	// SaveContext saves a context to disk
	SaveContext(context *Context) error

	// SaveContextSummary saves a context summary to disk
	SaveContextSummary(contextSummary *ContextSummary) error
}

// ContextStream represents a streaming context
type ContextStream struct {
	ID          string
	Name        string
	Description string
	Files       []string
	ProjectPath string
	TotalLines  int64
	TotalChars  int64
	CreatedAt   string
	UpdatedAt   string
	TokenCount  int
	contextPath string
}

// ContextLineRange represents a range of lines from a context
type ContextLineRange struct {
	StartLine int64    `json:"startLine"`
	EndLine   int64    `json:"endLine"`
	Lines     []string `json:"lines"`
}

// TaskAnalysis contains analysis results for a task
type TaskAnalysis struct {
	Type         string
	Priority     string
	Technologies []string
	FileTypes    []string
	Keywords     []string
	Reasoning    string
}

// ContextAnalysisResult contains the results of context analysis
type ContextAnalysisResult struct {
	Task            string
	TaskType        string
	Priority        string
	SelectedFiles   []*FileNode
	DependencyFiles []*FileNode
	Context         string
	AnalysisTime    time.Duration
	Recommendations []string
	EstimatedTokens int
	Confidence      float64
}
