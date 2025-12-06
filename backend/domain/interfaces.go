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
	// New methods for remote/branch context building
	IsGitRepository(projectPath string) bool
	CloneRepository(url, targetPath string, depth int) error
	CheckoutBranch(projectPath, branch string) error
	CheckoutCommit(projectPath, commitHash string) error
	GetCommitHistory(projectPath string, limit int) ([]CommitInfo, error)
	FetchRemoteBranches(projectPath string) ([]string, error)
	// Read files at specific ref without checkout
	ListFilesAtRef(projectPath, ref string) ([]string, error)
	GetFileAtRef(projectPath, filePath, ref string) (string, error)
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
	GetQwenKey() string
	SetQwenKey(key string)
	GetQwenHost() string
	SetQwenHost(host string)
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

// KeyResolver defines the interface for resolving API keys for different providers
type KeyResolver interface {
	// GetKey retrieves the API key for a given provider type from settings
	GetKey(providerType string, settings SettingsDTO) (string, error)
}

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
}

// ContextRepository определяет интерфейс для хранения контекста
type ContextRepository interface {
	// SaveContextSummary persists lightweight context metadata on disk
	SaveContextSummary(summary *ContextSummary) error

	// GetContextSummary retrieves persisted context metadata by ID
	GetContextSummary(ctx context.Context, contextID string) (*ContextSummary, error)

	// GetProjectContextSummaries lists context metadata for a project path
	GetProjectContextSummaries(ctx context.Context, projectPath string) ([]*ContextSummary, error)

	// DeleteContext removes context metadata and content from disk
	DeleteContext(ctx context.Context, contextID string) error

	// ReadContextChunk returns a memory-safe chunk of context content
	ReadContextChunk(ctx context.Context, contextID string, startLine int, lineCount int) (*ContextChunk, error)

	// ReadContextContent returns the full context content as a string
	ReadContextContent(ctx context.Context, contextID string) (string, error)
}

// IBuildService defines the interface for build service operations
type IBuildService interface {
	// Build executes project build
	Build(ctx context.Context, projectPath, language string) (*BuildResult, error)

	// TypeCheck performs type checking
	TypeCheck(ctx context.Context, projectPath, language string) (*TypeCheckResult, error)

	// BuildAndTypeCheck executes build and type checking
	BuildAndTypeCheck(ctx context.Context, projectPath, language string) (*BuildResult, *TypeCheckResult, error)

	// BuildMultiLanguage executes build for multiple languages
	BuildMultiLanguage(ctx context.Context, projectPath string, languages []string) (map[string]*BuildResult, error)

	// TypeCheckMultiLanguage executes type checking for multiple languages
	TypeCheckMultiLanguage(ctx context.Context, projectPath string, languages []string) (map[string]*TypeCheckResult, error)

	// ValidateProject performs full project validation
	ValidateProject(ctx context.Context, projectPath string, languages []string) (*ProjectValidationResult, error)

	// GetSupportedLanguages returns supported languages
	GetSupportedLanguages() []string

	// DetectLanguages detects languages in project
	DetectLanguages(ctx context.Context, projectPath string) ([]string, error)
}

// ITestService defines the interface for test service operations
type ITestService interface {
	// RunTests executes tests according to configuration
	RunTests(ctx context.Context, config *TestConfig) ([]*TestResult, error)

	// RunTargetedTests executes targeted tests for affected files
	RunTargetedTests(ctx context.Context, config *TestConfig, changedFiles []string) ([]*TestResult, error)

	// DiscoverTests discovers tests in project
	DiscoverTests(ctx context.Context, projectPath, language string) (*TestSuite, error)

	// BuildAffectedGraph builds affected files graph
	BuildAffectedGraph(ctx context.Context, changedFiles []string, projectPath string) (*AffectedGraph, error)

	// GetTestCoverage gets test coverage
	GetTestCoverage(ctx context.Context, testPath string) (*TestCoverage, error)

	// GetSupportedLanguages returns supported languages
	GetSupportedLanguages() []string

	// RunSmokeTests executes only smoke tests
	RunSmokeTests(ctx context.Context, projectPath, language string) ([]*TestResult, error)

	// RunUnitTests executes only unit tests
	RunUnitTests(ctx context.Context, projectPath, language string) ([]*TestResult, error)

	// RunIntegrationTests executes only integration tests
	RunIntegrationTests(ctx context.Context, projectPath, language string) ([]*TestResult, error)

	// ValidateTestResults validates test results
	ValidateTestResults(results []*TestResult) *TestValidationResult
}

// IStaticAnalyzerService defines the interface for static analyzer service operations
type IStaticAnalyzerService interface {
	// AnalyzeProject analyzes project
	AnalyzeProject(ctx context.Context, projectPath string, languages []string) (*StaticAnalysisReport, error)

	// AnalyzeFile analyzes single file
	AnalyzeFile(ctx context.Context, filePath, language string) (*StaticAnalysisResult, error)

	// GetSupportedAnalyzers returns supported analyzers
	GetSupportedAnalyzers() []StaticAnalyzerType

	// GetAnalyzerForLanguage returns analyzer for language
	GetAnalyzerForLanguage(language string) (StaticAnalyzer, error)

	// ValidateAnalysisResults validates analysis results
	ValidateAnalysisResults(results map[string]*StaticAnalysisResult) *StaticAnalysisValidationResult

	// AnalyzeGoProject analyzes Go project
	AnalyzeGoProject(ctx context.Context, projectPath string) (*StaticAnalysisResult, error)

	// AnalyzeTypeScriptProject analyzes TypeScript project
	AnalyzeTypeScriptProject(ctx context.Context, projectPath string) (*StaticAnalysisResult, error)

	// AnalyzeJavaScriptProject analyzes JavaScript project
	AnalyzeJavaScriptProject(ctx context.Context, projectPath string) (*StaticAnalysisResult, error)

	// AnalyzeJavaProject analyzes Java project
	AnalyzeJavaProject(ctx context.Context, projectPath string) (*StaticAnalysisResult, error)

	// AnalyzePythonProject analyzes Python project
	AnalyzePythonProject(ctx context.Context, projectPath string) (*StaticAnalysisResult, error)

	// AnalyzeCppProject analyzes C/C++ project
	AnalyzeCppProject(ctx context.Context, projectPath string) (*StaticAnalysisResult, error)
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
