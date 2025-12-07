package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"shotgun_code/application"
	"shotgun_code/domain"
	"shotgun_code/domain/analysis"
	"shotgun_code/handlers"
	"shotgun_code/infrastructure/ai"
	"shotgun_code/infrastructure/analyzers"
	"shotgun_code/infrastructure/applyengine"
	"shotgun_code/infrastructure/embeddings"
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
	"sync"
	"time"

	// new wiring
	archiverinfra "shotgun_code/infrastructure/archiver"
	"shotgun_code/infrastructure/buildpipeline"
	"shotgun_code/infrastructure/diffengine"
	"shotgun_code/infrastructure/pdfgen"
	"shotgun_code/infrastructure/policy"
	"shotgun_code/infrastructure/symbolgraph"
	"shotgun_code/internal/initmanager"

	// Internal services (unified architecture)
	contextservice "shotgun_code/internal/context"
	projectservice "shotgun_code/internal/project"
)

// Use domain constants for default hosts
const openRouterHost = domain.OpenRouterDefaultHost

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

	// Unified internal services (new architecture)
	ContextService *contextservice.Service
	ProjectService *projectservice.Service

	// Context interfaces (implemented by ContextService)
	ContextBuilder    domain.ContextBuilder
	ContextAnalyzer   domain.ContextAnalyzer
	ContextRepository domain.ContextRepository

	ReportService    *application.ReportService
	RouterLLMService *application.RouterLLMService
	Bridge           *wailsbridge.Bridge
	GitService       domain.GitService

	// Task Protocol Services
	TaskProtocolService         domain.TaskProtocolService
	ErrorAnalyzer               domain.ErrorAnalyzer
	CorrectionEngine            domain.CorrectionEngine
	TaskProtocolConfigService   *application.TaskProtocolConfigService
	TaskflowProtocolIntegration *application.TaskflowProtocolIntegration
	VerificationPipelineService *application.VerificationPipelineService

	// Qwen Task Services
	SmartContextService *application.SmartContextService
	QwenTaskService     *application.QwenTaskService

	// Semantic Search Services
	SymbolIndex       analysis.SymbolIndex
	EmbeddingProvider domain.EmbeddingProvider
	VectorStore       domain.VectorStore
	SemanticSearch    domain.SemanticSearchService
	RAGService        domain.RAGService
	SemanticHandler   *handlers.SemanticHandler

	// Handlers (new architecture)
	ProjectHandler  *handlers.ProjectHandler
	ContextHandler  *handlers.ContextHandler
	QwenHandler     *handlers.QwenHandler
	AIHandler       *handlers.AIHandler
	AnalysisHandler *handlers.AnalysisHandler
	SettingsHandler *handlers.SettingsHandler
	TaskflowHandler *handlers.TaskflowHandler

	// Lazy initialization support
	lazyInitOnce              sync.Once
	testServiceOnce           sync.Once
	staticAnalyzerServiceOnce sync.Once
	sbomServiceOnce           sync.Once
	symbolGraphOnce           sync.Once

	// Lazy service manager for coordinated lifecycle management
	lazyManager *initmanager.LazyServiceManager

	// Cleanup goroutine control
	cleanupStopCh chan struct{}
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

	// Set provider getter in IntelligentAIService (uses interface to break circular dependency)
	intelligentService.SetProviderGetter(c.AIService)

	// Connect SettingsService to AIService for cache invalidation on settings change
	c.SettingsService.SetAICacheInvalidator(c.AIService)

	// Create OPA service
	opaService := policy.NewOPAService(c.Log)

	// Create path provider (using standard filepath implementation)
	pathProvider := filesystem.NewFilePathProvider()

	// Create file system writer (using standard os implementation)
	fileSystemWriter := &OSFileSystemWriter{}

	// Get context directory
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return nil, fmt.Errorf("failed to determine user home directory: %w", homeErr)
	}
	contextDir := filepath.Join(homeDir, ".shotgun-code", "contexts")
	if mkErr := os.MkdirAll(contextDir, 0o755); mkErr != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", mkErr)
	}

	// Create comment stripper for code preprocessing
	commentStripper := textutils.NewCommentStripper(c.Log)
	_ = commentStripper  // Will be used by ContextService internally
	_ = opaService       // Will be used by ContextService internally
	_ = pathProvider     // Will be used by ProjectService internally
	_ = fileSystemWriter // Will be used by ContextService internally

	// Create unified ContextService (replaces ContextBuilder, ContextGenerator, ContextRepository)
	c.ContextService, err = contextservice.NewService(
		c.FileReader,
		&SimpleTokenCounter{},
		c.Bus,
		c.Log,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create context service: %w", err)
	}

	// ContextService implements ContextRepository interface
	c.ContextRepository = c.ContextService

	// Create ContextBuilder adapter
	c.ContextBuilder = contextservice.NewContextBuilderAdapter(c.ContextService)

	// Create analyzer responsible for task-driven context suggestions
	contextAnalyzer := application.NewContextAnalyzer(c.Log, c.AIService)
	c.ContextAnalyzer = contextAnalyzer

	// Create unified ProjectService
	c.ProjectService = projectservice.NewService(
		c.Log,
		c.Bus,
		c.TreeBuilder,
		c.GitRepo,
		c.ContextService, // Pass context service interface
	)

	// Initialize remaining services
	c.ContextAnalysis = contextAnalyzer

	// Create symbol graph builders
	goSymbolGraphBuilder := symbolgraph.NewGoSymbolGraphBuilder(c.Log)
	symbolGraphBuilders := make(map[string]domain.SymbolGraphBuilder)
	symbolGraphBuilders["go"] = goSymbolGraphBuilder

	// Create import graph builders (currently no implementation, using nil map)
	importGraphBuilders := make(map[string]domain.ImportGraphBuilder)

	c.SymbolGraph = application.NewSymbolGraphService(c.Log, symbolGraphBuilders, importGraphBuilders)

	// Create CallStack Analyzer and Smart Context Service for Qwen integration
	callStackAnalyzer := symbolgraph.NewCallStackAnalyzerAdapter(c.Log)
	c.SmartContextService = application.NewSmartContextService(
		c.Log,
		c.FileReader,
		c.SymbolGraph,
		callStackAnalyzer,
	)

	// Create TestService with lazy initialization
	c.testServiceOnce.Do(func() {
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
	})

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
	if taskTypeProvider, ok := c.TaskflowService.(domain.TaskTypeProvider); ok {
		c.GuardrailService.SetTaskTypeProvider(taskTypeProvider)
	} else {
		c.Log.Warning("TaskflowService does not implement TaskTypeProvider; guardrails may be limited")
	}

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
	reportRepo, err := reportfs.NewReportFileSystemRepository(c.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize report repository: %w", err)
	}
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

	// Initialize lazy service manager for memory optimization
	c.lazyManager = initmanager.NewLazyServiceManager()

	// Note: Services are currently eagerly initialized for compatibility
	// Future enhancement: wrap heavy services in LazyService[T] and register with manager
	// Example: c.lazyManager.Register("symbolgraph", lazySymbolGraphService)

	// Start periodic cleanup of unused services (runs every 5 minutes)
	// Note: This goroutine will be stopped when lazyManager is shutdown
	c.cleanupStopCh = make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-c.cleanupStopCh:
				return
			case <-ticker.C:
				// Unload services idle for more than 10 minutes
				unloaded := c.lazyManager.UnloadUnusedServices(10 * time.Minute)
				if unloaded > 0 {
					c.Log.Info(fmt.Sprintf("Unloaded %d idle services to free memory", unloaded))
				}
			}
		}
	}()

	// Initialize Semantic Search Services
	if err := c.initializeSemanticSearch(); err != nil {
		c.Log.Warning(fmt.Sprintf("Semantic search initialization failed (non-critical): %v", err))
		// Non-critical - continue without semantic search
	}

	// Initialize handlers (new architecture)
	if err := c.initializeHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	return c, nil
}

// initializeSemanticSearch initializes semantic search services
func (c *AppContainer) initializeSemanticSearch() error {
	// Get data directory for embeddings storage
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	dataDir := filepath.Join(homeDir, ".shotgun-code", "embeddings")

	// Create analyzer registry for symbol extraction
	analyzerRegistry := analyzers.NewAnalyzerRegistry()

	// Create symbol index with SQLite caching for incremental indexing
	symbolCacheDir := filepath.Join(dataDir, "symbol_cache")
	cachedSymbolIndex, err := analyzers.NewCachedSymbolIndex(analyzerRegistry, symbolCacheDir)
	if err != nil {
		c.Log.Warning(fmt.Sprintf("Failed to create cached symbol index, falling back to in-memory: %v", err))
		c.SymbolIndex = analyzers.NewSymbolIndex(analyzerRegistry)
	} else {
		c.SymbolIndex = cachedSymbolIndex
	}

	// Create vector store (SQLite-based)
	vectorStore, err := embeddings.NewSQLiteVectorStore(dataDir, c.Log)
	if err != nil {
		return fmt.Errorf("failed to create vector store: %w", err)
	}
	c.VectorStore = vectorStore

	// Create embedding provider (OpenAI by default)
	// Get API key from settings
	settings, err := c.SettingsService.GetSettingsDTO()
	if err != nil {
		c.Log.Warning("Failed to get settings for embedding provider: " + err.Error())
	}

	apiKey := ""
	if settings.OpenAIAPIKey != "" {
		apiKey = settings.OpenAIAPIKey
	}

	if apiKey != "" {
		embeddingProvider, err := embeddings.NewOpenAIEmbeddingProvider(
			apiKey,
			domain.EmbeddingModelOpenAI3S, // Use small model by default
			c.Log,
		)
		if err != nil {
			c.Log.Warning("Failed to create embedding provider: " + err.Error())
		} else {
			c.EmbeddingProvider = embeddingProvider
		}
	}

	// Create semantic search service (only if embedding provider is available)
	if c.EmbeddingProvider != nil {
		c.SemanticSearch = application.NewSemanticSearchService(
			c.EmbeddingProvider,
			c.VectorStore,
			c.SymbolIndex,
			c.Log,
		)

		// Create RAG service
		c.RAGService = application.NewRAGService(
			c.SemanticSearch,
			c.EmbeddingProvider,
			c.Log,
		)

		// Create semantic handler
		c.SemanticHandler = handlers.NewSemanticHandler(
			c.SemanticSearch,
			c.RAGService,
			c.Log,
		)

		c.Log.Info("Semantic search services initialized successfully")
	} else {
		c.Log.Warning("Semantic search disabled: no embedding provider configured (set OpenAI API key)")
	}

	return nil
}

// initializeHandlers creates all handlers with proper dependencies
func (c *AppContainer) initializeHandlers() error {
	// Project Handler - delegates to ProjectService
	c.ProjectHandler = handlers.NewProjectHandler(
		c.Log,
		c.Bus,
		c.ProjectService,
		c.Watcher,
		c.FileReader,
		c.GitRepo,
	)

	// Context Handler - uses unified ContextService
	c.ContextHandler = handlers.NewContextHandler(
		c.Log,
		c.Bus,
		c.ContextService,
	)

	// AI Handler
	c.AIHandler = handlers.NewAIHandler(
		c.Log,
		c.AIService,
		c.ContextAnalysis,
	)

	// Analysis Handler
	c.AnalysisHandler = handlers.NewAnalysisHandler(
		c.Log,
		c.TestService,
		c.StaticAnalyzerService,
		c.BuildService,
		c.SBOMService,
		c.SymbolGraph,
	)

	// Settings Handler
	c.SettingsHandler = handlers.NewSettingsHandler(
		c.Log,
		c.SettingsService,
	)

	// Taskflow Handler
	c.TaskflowHandler = handlers.NewTaskflowHandler(
		c.Log,
		c.TaskflowService,
		c.GuardrailService,
		c.RepairService,
		c.TaskProtocolService,
		c.TaskProtocolConfigService,
		c.TaskflowProtocolIntegration,
		c.BuildService,
	)

	// Qwen Task Service and Handler
	c.QwenTaskService = application.NewQwenTaskService(
		c.Log,
		c.AIService,
		c.SmartContextService,
		c.SettingsService,
	)

	c.QwenHandler = handlers.NewQwenHandler(
		c.Log,
		c.QwenTaskService,
	)

	return nil
}

// GetLazyServiceStats returns statistics about lazy-loaded services
func (c *AppContainer) GetLazyServiceStats() map[string]interface{} {
	if c.lazyManager == nil {
		return map[string]interface{}{
			"enabled": false,
			"message": "Lazy service manager not initialized",
		}
	}

	stats := c.lazyManager.GetInitializationStats()
	stats["enabled"] = true
	return stats
}

func createModelFetchers(ctx context.Context, log domain.Logger, repo domain.SettingsRepository) domain.ModelFetcherRegistry {
	registry := ai.GetProviderRegistry(openRouterHost)
	fetchers := make(domain.ModelFetcherRegistry)

	for providerType, config := range registry {
		// Capture variables for closure
		providerType := providerType
		config := config

		// Create a cached fetcher
		cachedFetcher := &cachedModelFetcher{
			fetcher: config.ModelFetcher,
			log:     log,
			repo:    repo,
			cache:   make(map[string][]string),
		}

		fetchers[providerType] = func(apiKey string) ([]string, error) {
			// For the model fetchers, we need to provide the host and logger
			// We'll use the container's logger and get the host based on provider type
			host := ""
			if providerType == "openrouter" {
				host = openRouterHost
			} else if providerType == "localai" {
				host = repo.GetLocalAIHost()
			} else if providerType == "qwen" {
				host = repo.GetQwenHost()
			}

			models, err := cachedFetcher.FetchModels(ctx, apiKey, host, log)
			if err != nil {
				log.Warning("Failed to create " + providerType + " client for model listing: " + err.Error())
				return nil, err
			}
			return models, nil
		}
	}

	return fetchers
}

// cachedModelFetcher adds caching to model fetchers
type cachedModelFetcher struct {
	fetcher func(context.Context, string, string, domain.Logger) ([]string, error)
	log     domain.Logger
	repo    domain.SettingsRepository
	cache   map[string][]string
	mu      sync.RWMutex
}

func (c *cachedModelFetcher) FetchModels(ctx context.Context, apiKey, host string, log domain.Logger) ([]string, error) {
	// Create a cache key based on API key and host
	cacheKey := apiKey + "|" + host

	// Check if we have cached models
	c.mu.RLock()
	if models, exists := c.cache[cacheKey]; exists {
		c.mu.RUnlock()
		log.Debug("Using cached models for provider")
		return models, nil
	}
	c.mu.RUnlock()

	// Fetch models and cache them
	models, err := c.fetcher(ctx, apiKey, host, log)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.mu.Lock()
	c.cache[cacheKey] = models
	c.mu.Unlock()

	return models, nil
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
		case "qwen":
			dto, err := settingsService.GetSettingsDTO()
			if err != nil {
				return "", err
			}
			if dto.QwenHost != "" {
				return dto.QwenHost, nil
			}
			return domain.QwenDefaultHost, nil
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

func (c *CommandRunnerImpl) RunCommand(_ context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

func (c *CommandRunnerImpl) RunCommandInDir(_ context.Context, dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Output()
}

// SimpleTokenCounter provides basic token estimation
type SimpleTokenCounter struct{}

func (s *SimpleTokenCounter) CountTokens(text string) int {
	// Simple approximation: 1 token ≈ 4 characters
	return len(text) / 4
}

// Shutdown gracefully shuts down all services in the container
func (c *AppContainer) Shutdown(ctx context.Context) error {
	c.Log.Info("Starting container shutdown...")

	var shutdownErrors []error

	// Stop cleanup goroutine first
	if c.cleanupStopCh != nil {
		close(c.cleanupStopCh)
	}

	// Shutdown handlers that support it
	if c.AIHandler != nil {
		if err := c.AIHandler.Shutdown(ctx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("AIHandler shutdown: %w", err))
		}
	}

	// Shutdown services
	if c.AIService != nil {
		if err := c.AIService.Shutdown(ctx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("AIService shutdown: %w", err))
		}
	}

	// Stop file watcher
	if c.Watcher != nil {
		c.Watcher.Stop()
	}

	// Close cached symbol index if it supports closing
	if c.SymbolIndex != nil {
		if closer, ok := c.SymbolIndex.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				shutdownErrors = append(shutdownErrors, fmt.Errorf("SymbolIndex close: %w", err))
			}
		}
	}

	// Shutdown lazy service manager
	if c.lazyManager != nil {
		// Force unload all services on shutdown
		c.lazyManager.UnloadUnusedServices(0)
	}

	if len(shutdownErrors) > 0 {
		c.Log.Warning(fmt.Sprintf("Container shutdown completed with %d errors", len(shutdownErrors)))
		return fmt.Errorf("shutdown errors: %v", shutdownErrors)
	}

	c.Log.Info("Container shutdown complete")
	return nil
}
