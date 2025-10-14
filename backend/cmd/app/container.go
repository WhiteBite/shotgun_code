package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"shotgun_code/application"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/ai"
	"shotgun_code/infrastructure/applyengine"
	execinfra "shotgun_code/infrastructure/exec"
	"shotgun_code/infrastructure/filereader"
	"shotgun_code/infrastructure/filesystem"
	"shotgun_code/infrastructure/formatters"
	"shotgun_code/infrastructure/fsscanner"
	"shotgun_code/infrastructure/fswatcher"
	"shotgun_code/infrastructure/git"
	"shotgun_code/infrastructure/reportfs"
	"shotgun_code/infrastructure/sbomlicensing"
	"shotgun_code/infrastructure/settingsfs"
	"shotgun_code/infrastructure/staticanalyzer"
	"shotgun_code/infrastructure/taskflowrepo"
	"shotgun_code/infrastructure/testengine"
	"shotgun_code/infrastructure/textutils"
	"shotgun_code/infrastructure/uxreports"
	"shotgun_code/infrastructure/wailsbridge"
	"time"

	// new wiring
	archiverinfra "shotgun_code/infrastructure/archiver"
	"shotgun_code/infrastructure/buildpipeline"
	"shotgun_code/infrastructure/diffengine"
	"shotgun_code/infrastructure/pdfgen"
	"shotgun_code/infrastructure/policy"
	"shotgun_code/infrastructure/symbolgraph"
)

const openRouterHost = "https://openrouter.ai/api/v1"

// AppContainer holds all the services and repositories for the application.
type AppContainer struct {
	Log                   domain.Logger
	Bus                   domain.EventBus
	SettingsRepo          domain.SettingsRepository
	FileReader            domain.FileContentReader
	GitRepo               domain.GitRepository
	TreeBuilder           domain.TreeBuilder
	ContextSplitter       domain.ContextSplitter
	Watcher               domain.FileSystemWatcher
	CommandRunner         domain.CommandRunner
	SettingsService       *application.SettingsService
	ProjectService        *application.ProjectService
	AIService             *application.AIService
	ContextAnalysis       domain.ContextAnalyzer
	SymbolGraph           *application.SymbolGraphService
	TestService           domain.ITestService
	StaticAnalyzerService domain.IStaticAnalyzerService
	SBOMService           *application.SBOMService
	RepairService         domain.RepairService
	GuardrailService      domain.GuardrailService
	TaskflowService       domain.TaskflowService
	UXMetricsService      domain.UXMetricsService
	ApplyService          *application.ApplyService
	DiffService           *application.DiffService
	BuildService          domain.IBuildService
	ExportService         *application.ExportService
	// NEW: Separate context services following SRP
	ContextBuilder    domain.ContextBuilder
	ContextAnalyzer   domain.ContextAnalyzer
	ContextRepository domain.ContextRepository
	ContextGenerator  *application.ContextGenerator // NEW: For async context generation
	ReportService     *application.ReportService
	RouterLLMService  *application.RouterLLMService
	Bridge            *wailsbridge.Bridge
	GitService        domain.GitService

	// Task Protocol Services
	TaskProtocolService         domain.TaskProtocolService
	ErrorAnalyzer               domain.ErrorAnalyzer
	CorrectionEngine            domain.CorrectionEngine
	TaskProtocolConfigService   *application.TaskProtocolConfigService
	TaskflowProtocolIntegration *application.TaskflowProtocolIntegration
	VerificationPipelineService *application.VerificationPipelineService
}

// NewContainer creates and wires up all the application dependencies.
func NewContainer(ctx context.Context, embeddedIgnoreGlob, defaultCustomPrompt string) (*AppContainer, error) {
	c := &AppContainer{}
	var err error

	// Bridge for Wails (Logger and EventBus)
	bridge := wailsbridge.New(ctx)
	c.Bridge = bridge
	c.Log = bridge
	c.Bus = bridge

	// Repositories and Infrastructure
	c.SettingsRepo, err = settingsfs.New(c.Log, embeddedIgnoreGlob, defaultCustomPrompt)
	if err != nil {
		return nil, err
	}
	c.FileReader = filereader.NewSecureFileReader(c.Log)
	c.GitRepo = git.New(c.Log)
	c.TreeBuilder = fsscanner.New(c.SettingsRepo, c.Log)
	c.ContextSplitter = textutils.NewContextSplitter(c.Log)
	c.Watcher, err = fswatcher.New(ctx, c.Bus)
	if err != nil {
		return nil, err
	}
	c.CommandRunner = execinfra.NewCommandRunnerImpl(c.Log)

	// Application Services
	modelFetchers := createModelFetchers(ctx, c.Log, c.SettingsRepo)
	c.SettingsService, err = application.NewSettingsService(c.Log, c.Bus, c.SettingsRepo, modelFetchers)
	if err != nil {
		return nil, err
	}
	// Connect watcher to settings changes
	c.SettingsService.OnIgnoreRulesChanged(c.Watcher.RefreshAndRescan)

	// AI Service needs to be created before context service
	providerRegistry := createProviderRegistry(c.Log, c.SettingsService)

	// Create rate limiter and metrics collector
	rateLimiter := application.NewRateLimiter()
	metrics := application.NewMetricsCollector()

	// Create intelligent service with dependencies
	intelligentService := application.NewIntelligentAIService(c.SettingsService, c.Log, providerRegistry, rateLimiter, metrics)

	// Create AI service with intelligent service
	c.AIService = application.NewAIService(c.SettingsService, c.Log, providerRegistry, intelligentService)

	// NEW: Create separate context services following SRP
	tokenCounter := application.SimpleTokenCounter

	// Create OPA service
	opaService := policy.NewOPAService(c.Log)

	// Create path provider (using standard filepath implementation)
	pathProvider := filesystem.NewFilePathProvider()

	// Create file system writer (using standard os implementation)
	fileSystemWriter := &OSFileSystemWriter{}

	// Get context directory
	homeDir, _ := os.UserHomeDir()
	contextDir := filepath.Join(homeDir, ".shotgun-code", "contexts")
	os.MkdirAll(contextDir, 0755)

	// Create comment stripper for code preprocessing
	commentStripper := textutils.NewCommentStripper(c.Log)

	// Create context repository for persistent, memory-safe storage
	contextRepository := application.NewContextRepository(c.Log, contextDir)
	c.ContextRepository = contextRepository

	// Create separate context services
	c.ContextBuilder = application.NewContextBuilder(
		c.FileReader,
		tokenCounter,
		c.Log,
		c.SettingsService,
		c.Bus,
		opaService,
		pathProvider,
		fileSystemWriter,
		commentStripper,
		contextRepository,
		contextDir,
	)

	// Create ContextGenerator for async operations
	c.ContextGenerator = application.NewContextGenerator(
		c.FileReader,
		c.Log,
		c.Bus,
		contextDir,
	)

	// Create analyzer responsible for task-driven context suggestions
	contextAnalyzer := application.NewContextAnalyzer(c.Log, c.AIService)
	c.ContextAnalyzer = contextAnalyzer

	// Create ProjectService with the new context builder and generator
	c.ProjectService = application.NewProjectService(c.Log, c.Bus, c.TreeBuilder, c.GitRepo, c.ContextBuilder, c.ContextGenerator, pathProvider, &OSFileStatProvider{})

	// Initialize remaining services
	c.ContextAnalysis = contextAnalyzer

	// Create symbol graph builders
	goSymbolGraphBuilder := symbolgraph.NewGoSymbolGraphBuilder(c.Log)
	symbolGraphBuilders := make(map[string]domain.SymbolGraphBuilder)
	symbolGraphBuilders["go"] = goSymbolGraphBuilder

	// Create import graph builders (currently no implementation, using nil map)
	importGraphBuilders := make(map[string]domain.ImportGraphBuilder)

	c.SymbolGraph = application.NewSymbolGraphService(c.Log, symbolGraphBuilders, importGraphBuilders)

	// Create TestEngine and infrastructure components
	testEngine := testengine.NewTestEngine(c.Log, goSymbolGraphBuilder)
	testEngine.RegisterTestRunner("go", testengine.NewGoTestRunner(c.Log))
	// testEngine.RegisterTestRunner("typescript", testengine.NewTypeScriptTestRunner(c.Log))
	// testEngine.RegisterTestRunner("java", testengine.NewJavaTestRunner(c.Log))

	// Register test analyzers for supported languages
	testEngine.RegisterTestAnalyzer("go", testengine.NewGoTestAnalyzer(c.Log))
	// testEngine.RegisterTestAnalyzer("typescript", testengine.NewTypeScriptTestAnalyzer(c.Log))
	// testEngine.RegisterTestAnalyzer("java", testengine.NewJavaTestAnalyzer(c.Log))

	// Create TestService with the TestEngine
	c.TestService = application.NewTestService(c.Log, testEngine)

	// Create Static Analyzer Engine and infrastructure components
	staticAnalyzerEngine := staticanalyzer.NewStaticAnalyzerEngine(c.Log)
	staticAnalyzerEngine.RegisterAnalyzer(staticanalyzer.NewStaticcheckAnalyzer(c.Log))
	staticAnalyzerEngine.RegisterAnalyzer(staticanalyzer.NewESLintAnalyzer(c.Log))
	staticAnalyzerEngine.RegisterAnalyzer(staticanalyzer.NewErrorProneAnalyzer(c.Log))
	staticAnalyzerEngine.RegisterAnalyzer(staticanalyzer.NewRuffAnalyzer(c.Log))
	staticAnalyzerEngine.RegisterAnalyzer(staticanalyzer.NewClangTidyAnalyzer(c.Log))
	c.StaticAnalyzerService = application.NewStaticAnalyzerService(c.Log, staticAnalyzerEngine)

	// Create SBOM infrastructure components
	sbomGenerator := sbomlicensing.NewSyftGenerator(c.Log)
	vulnScanner := sbomlicensing.NewGrypeScanner(c.Log)
	licenseScanner := sbomlicensing.NewLicenseScanner(c.Log)
	sbomFileStatProvider := &OSFileStatProvider{}
	c.SBOMService = application.NewSBOMService(c.Log, sbomGenerator, vulnScanner, licenseScanner, sbomFileStatProvider)

	c.RepairService = application.NewRepairService(c.Log, c.CommandRunner)

	// Create TaskflowRepository
	taskflowRepo := taskflowrepo.NewFileSystemTaskflowRepository("tasks/status.json")

	// Create RouterPlannerService
	planner := application.NewRouterPlannerService(c.Log, c.BuildService, c.TestService, c.StaticAnalyzerService, c.RepairService)

	// Create OPA service and file stat provider for GuardrailService
	guardrailOPAService := policy.NewOPAService(c.Log)
	guardrailFileStatProvider := &OSFileStatProvider{}
	c.GuardrailService = application.NewGuardrailService(c.Log, guardrailOPAService, guardrailFileStatProvider)

	// Create TaskflowService with injected dependencies
	c.TaskflowService = application.NewTaskflowService(c.Log, planner, c.RouterLLMService, c.GuardrailService, taskflowRepo, c.GitRepo)

	// ⚠️ CRITICAL: Update GuardrailService with TaskTypeProvider to resolve circular dependency
	// This MUST be called AFTER TaskflowService is created
	// Order matters: TaskflowService → GuardrailService.SetTaskTypeProvider
	c.GuardrailService.(domain.GuardrailService).SetTaskTypeProvider(c.TaskflowService.(domain.TaskTypeProvider))

	// Create UXReportRepository
	uxReportRepo := uxreports.NewFileSystemUXReportRepository("reports/ux")
	c.UXMetricsService = application.NewUXMetricsService(c.Log, uxReportRepo)

	// Создаем конфигурацию для движка применения
	applyConfig := &domain.ApplyEngineConfig{
		AutoFormat:     true,
		AutoFixImports: true,
		BackupFiles:    true,
		ValidateAfter:  true,
		Languages:      []string{"go", "typescript", "ts"},
	}

	// Создаем движок применения
	applyEngine := applyengine.NewApplyEngine(c.Log, applyConfig)

	// Создаем форматтеры
	formatterMap := map[string]domain.Formatter{
		"go":         formatters.NewGoFormatter(c.Log),
		"typescript": formatters.NewTypeScriptFormatter(c.Log),
		"ts":         formatters.NewTypeScriptFormatter(c.Log),
	}

	// Создаем исправители импортов
	// TEMPORARY: GoFormatter and TypeScriptFormatter implement both Formatter and ImportFixer interfaces
	// This is acceptable as they correctly handle both formatting and import fixing,
	// but in the future we should create dedicated GoImportFixer and TypeScriptImportFixer
	importFixerMap := map[string]domain.ImportFixer{
		"go":         formatters.NewGoFormatter(c.Log),         // Temporary: same as formatter
		"typescript": formatters.NewTypeScriptFormatter(c.Log), // Temporary: same as formatter
		"ts":         formatters.NewTypeScriptFormatter(c.Log), // Temporary: same as formatter
	}

	c.ApplyService = application.NewApplyService(c.Log, applyConfig, applyEngine, formatterMap, importFixerMap)

	// Создаем движок diff
	diffEngine := diffengine.NewDiffEngine(c.Log)
	c.DiffService = application.NewDiffService(c.Log, diffEngine)

	// Создаем build pipeline
	buildPipeline := buildpipeline.NewBuildPipeline(c.Log)
	c.BuildService = application.NewBuildService(c.Log, buildPipeline)

	// new: wire PDF and ZIP implementations
	pdfGen := pdfgen.NewGofpdfGenerator(c.Log)
	arch := archiverinfra.NewZipArchiver(c.Log)
	tempFileProvider := &OSTempFileProvider{}
	exportFileStatProvider := &OSFileStatProvider{}
	// Create path provider and file system writer for ExportService
	exportPathProvider := &FilePathProvider{}
	exportFileSystemWriter := &OSFileSystemWriter{}
	c.ExportService = application.NewExportService(c.Log, c.ContextSplitter, pdfGen, arch, tempFileProvider, exportPathProvider, exportFileSystemWriter, exportFileStatProvider)

	// Initialize new services
	reportRepo := reportfs.NewReportFileSystemRepository(c.Log)
	c.ReportService = application.NewReportService(c.Log, reportRepo)

	// Initialize RouterLLMService
	routerLLMConfig := application.RouterLLMConfig{
		Enabled: false, // Disabled by default
		LLMConfig: domain.LLMConfig{
			BaseURL:       "http://localhost:8080",
			Timeout:       30 * time.Second,
			MaxTokens:     2048,
			Temperature:   0.7,
			TopP:          0.9,
			TopK:          40,
			RepeatPenalty: 1.1,
		},
		FallbackToHeuristic: true,
		MaxRetries:          3,
		Timeout:             30 * time.Second,
	}

	// Create LLM client
	llamaClient := ai.NewLlamaCppClient(ai.LlamaCppConfig{
		BaseURL:       routerLLMConfig.LLMConfig.BaseURL,
		Timeout:       routerLLMConfig.LLMConfig.Timeout,
		MaxTokens:     routerLLMConfig.LLMConfig.MaxTokens,
		Temperature:   routerLLMConfig.LLMConfig.Temperature,
		TopP:          routerLLMConfig.LLMConfig.TopP,
		TopK:          routerLLMConfig.LLMConfig.TopK,
		RepeatPenalty: routerLLMConfig.LLMConfig.RepeatPenalty,
	}, c.Log)

	// Create adapter
	llmClient := ai.NewLlamaCppClientAdapter(llamaClient)

	// Create file reader
	fileReader := filereader.NewFileReader()

	c.RouterLLMService = application.NewRouterLLMServiceWithClient(routerLLMConfig, c.Log, llmClient, fileReader)

	// Initialize Task Protocol Services
	if err := initializeTaskProtocolServices(c); err != nil {
		return nil, fmt.Errorf("failed to initialize task protocol services: %w", err)
	}

	return c, nil
}

func createModelFetchers(ctx context.Context, log domain.Logger, repo domain.SettingsRepository) domain.ModelFetcherRegistry {
	registry := ai.GetProviderRegistry(openRouterHost)
	fetchers := make(domain.ModelFetcherRegistry)

	for providerType, config := range registry {
		// Capture variables for closure
		providerType := providerType
		config := config

		fetchers[providerType] = func(apiKey string) ([]string, error) {
			// For the model fetchers, we need to provide the host and logger
			// We'll use the container's logger and get the host based on provider type
			host := ""
			if providerType == "openrouter" {
				host = openRouterHost
			} else if providerType == "localai" {
				host = repo.GetLocalAIHost()
			}

			models, err := config.ModelFetcher(ctx, apiKey, host, log)
			if err != nil {
				log.Warning("Failed to create " + providerType + " client for model listing: " + err.Error())
				return nil, err
			}
			return models, nil
		}
	}

	return fetchers
}

func createProviderRegistry(log domain.Logger, settingsService *application.SettingsService) map[string]domain.AIProviderFactory {
	resolveHost := func(providerType string) (string, error) {
		switch providerType {
		case "openrouter":
			return openRouterHost, nil
		case "localai":
			dto, err := settingsService.GetSettingsDTO()
			if err != nil {
				return "", err
			}
			return dto.LocalAIHost, nil
		default:
			return "", nil
		}
	}

	return ai.NewAIProviderFactoryRegistry(log, openRouterHost, resolveHost)
}

// FilePathProvider implements domain.PathProvider using standard filepath functions
type FilePathProvider struct{}

func (p *FilePathProvider) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (p *FilePathProvider) Base(path string) string {
	return filepath.Base(path)
}

func (p *FilePathProvider) Dir(path string) string {
	return filepath.Dir(path)
}

func (p *FilePathProvider) IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

func (p *FilePathProvider) Clean(path string) string {
	return filepath.Clean(path)
}

func (p *FilePathProvider) Getwd() (string, error) {
	return os.Getwd()
}

// OSFileSystemWriter implements domain.FileSystemWriter using standard os functions
type OSFileSystemWriter struct{}

func (w *OSFileSystemWriter) WriteFile(filename string, data []byte, perm int) error {
	return os.WriteFile(filename, data, os.FileMode(perm))
}

func (w *OSFileSystemWriter) MkdirAll(path string, perm int) error {
	return os.MkdirAll(path, os.FileMode(perm))
}

func (w *OSFileSystemWriter) Remove(name string) error {
	return os.Remove(name)
}

func (w *OSFileSystemWriter) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// OSTempFileProvider implements domain.TempFileProvider using standard os functions
type OSTempFileProvider struct{}

func (p *OSTempFileProvider) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

// OSFileStatProvider implements domain.FileStatProvider using standard os functions
type OSFileStatProvider struct{}

func (p *OSFileStatProvider) Stat(name string) (domain.FileInfo, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	return fi, nil
}

// initializeTaskProtocolServices initializes the Task Protocol services
func initializeTaskProtocolServices(c *AppContainer) error {
	// Initialize Error Analyzer
	c.ErrorAnalyzer = application.NewErrorAnalyzer(c.Log)

	// Initialize Correction Engine with file system provider
	fileSystemProvider := &OSFileSystemProvider{}
	c.CorrectionEngine = application.NewCorrectionEngine(c.Log, fileSystemProvider)

	// Initialize Task Protocol Config Service
	c.TaskProtocolConfigService = application.NewTaskProtocolConfigService(c.Log, fileSystemProvider)

	// Initialize Task Protocol Service
	c.TaskProtocolService = application.NewTaskProtocolService(
		c.Log,
		nil, // Will be set after VerificationPipelineService is created
		c.StaticAnalyzerService,
		c.TestService,
		c.BuildService,
		c.GuardrailService,
		c.AIService.GetIntelligentService(),
		c.ErrorAnalyzer,
		c.CorrectionEngine,
	)

	// Create VerificationPipelineService with Task Protocol integration
	formatterService := application.NewFormatterService(c.Log, &CommandRunnerImpl{})
	c.VerificationPipelineService = application.NewVerificationPipelineService(
		c.Log,
		c.BuildService,
		c.TestService,
		c.StaticAnalyzerService,
		formatterService,
		&OSFileSystemWriter{},
		c.TaskProtocolService,
	)

	// Initialize Taskflow Protocol Integration
	c.TaskflowProtocolIntegration = application.NewTaskflowProtocolIntegration(
		c.Log,
		c.TaskflowService,
		c.TaskProtocolService,
		c.TaskProtocolConfigService,
		c.AIService.GetIntelligentService(),
	)

	return nil
}

// OSFileSystemProvider implements domain.FileSystemProvider
type OSFileSystemProvider struct{}

func (o *OSFileSystemProvider) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (o *OSFileSystemProvider) WriteFile(filename string, data []byte, perm int) error {
	return os.WriteFile(filename, data, os.FileMode(perm))
}

func (o *OSFileSystemProvider) MkdirAll(path string, perm int) error {
	return os.MkdirAll(path, os.FileMode(perm))
}

// CommandRunnerImpl implements domain.CommandRunner
type CommandRunnerImpl struct{}

func (c *CommandRunnerImpl) RunCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

func (c *CommandRunnerImpl) RunCommandInDir(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	return cmd.Output()
}
