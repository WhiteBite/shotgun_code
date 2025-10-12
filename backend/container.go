package main

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"time"

	// Internal service packages - gradually migrating
	contextservice "shotgun_code/internal/context"
	projectservice "shotgun_code/internal/project"

	// Legacy services (to be migrated)
	"shotgun_code/application"

	// Infrastructure
	"shotgun_code/infrastructure/ai"
	"shotgun_code/infrastructure/filereader"
	"shotgun_code/infrastructure/fsscanner"
	"shotgun_code/infrastructure/fswatcher"
	"shotgun_code/infrastructure/git"
	"shotgun_code/infrastructure/settingsfs"
	"shotgun_code/infrastructure/textutils"
	"shotgun_code/infrastructure/wailsbridge"
)

const openRouterHost = "https://openrouter.ai/api/v1"

// Container holds all the services and repositories organized by bounded contexts
type Container struct {
	// Infrastructure Layer
	Log             domain.Logger
	Bus             domain.EventBus
	SettingsRepo    domain.SettingsRepository
	FileReader      domain.FileContentReader
	GitRepo         domain.GitRepository
	TreeBuilder     domain.TreeBuilder
	ContextSplitter domain.ContextSplitter
	Watcher         domain.FileSystemWatcher
	Bridge          *wailsbridge.Bridge

	// Bounded Context Services (New Internal Architecture)
	ProjectService *projectservice.Service
	ContextService *contextservice.Service

	// Legacy Services (To be migrated)
	SettingsService       *application.SettingsService
	AIService             *application.AIService
	ContextAnalysis       domain.ContextAnalyzer
	SymbolGraph           *application.SymbolGraphService
	TestService           *application.TestService
	StaticAnalyzerService *application.StaticAnalyzerService
	SBOMService           *application.SBOMService
	RepairService         domain.RepairService
	GuardrailService      domain.GuardrailService
	TaskflowService       domain.TaskflowService
	UXMetricsService      domain.UXMetricsService
	ApplyService          *application.ApplyService
	DiffService           *application.DiffService
	BuildService          *application.BuildService
	ExportService         *application.ExportService
	ReportService         *application.ReportService

	// Task Protocol Services (New)
	TaskProtocolService         domain.TaskProtocolService
	ErrorAnalyzer               domain.ErrorAnalyzer
	CorrectionEngine            domain.CorrectionEngine
	TaskProtocolConfigService   *application.TaskProtocolConfigService
	VerificationPipelineService *application.VerificationPipelineService
}

// NewContainer creates and wires up all the application dependencies with new architecture
func NewContainer(ctx context.Context, embeddedIgnoreGlob, defaultCustomPrompt string) (*Container, error) {
	c := &Container{}

	// Initialize infrastructure layer
	if err := c.initializeInfrastructure(ctx, embeddedIgnoreGlob, defaultCustomPrompt); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize bounded context services
	if err := c.initializeBoundedContexts(); err != nil {
		return nil, fmt.Errorf("failed to initialize bounded contexts: %w", err)
	}

	return c, nil
}

// initializeInfrastructure sets up all infrastructure dependencies
func (c *Container) initializeInfrastructure(ctx context.Context, embeddedIgnoreGlob, defaultCustomPrompt string) error {
	// Bridge for Wails (Logger and EventBus)
	bridge := wailsbridge.New(ctx)
	c.Bridge = bridge
	c.Log = bridge
	c.Bus = bridge

	var err error

	// Repositories and Infrastructure
	c.SettingsRepo, err = settingsfs.New(c.Log, embeddedIgnoreGlob, defaultCustomPrompt)
	if err != nil {
		return err
	}

	c.FileReader = filereader.NewSecureFileReader(c.Log)
	c.GitRepo = git.New(c.Log)
	c.TreeBuilder = fsscanner.New(c.SettingsRepo, c.Log)
	c.ContextSplitter = textutils.NewContextSplitter(c.Log)

	c.Watcher, err = fswatcher.New(ctx, c.Bus)
	if err != nil {
		return err
	}

	return nil
}

// initializeBoundedContexts creates the new internal services with proper dependency injection
func (c *Container) initializeBoundedContexts() error {
	// Initialize application services that were previously in initializeLegacyServices
	modelFetchers := c.createModelFetchers(context.Background())
	var err error

	c.SettingsService, err = application.NewSettingsService(c.Log, c.Bus, c.SettingsRepo, modelFetchers)
	if err != nil {
		return err
	}

	// Connect watcher to settings changes
	c.SettingsService.OnIgnoreRulesChanged(c.Watcher.RefreshAndRescan)

	// AI Service needs to be created before context service
	providerRegistry := c.createProviderRegistry()

	// Create rate limiter and metrics collector
	rateLimiter := application.NewRateLimiter()
	metrics := application.NewMetricsCollector()

	// Create intelligent service with dependencies
	intelligentService := application.NewIntelligentAIService(c.SettingsService, c.Log, providerRegistry, rateLimiter, metrics)

	// Create AI service with intelligent service
	c.AIService = application.NewAIService(c.SettingsService, c.Log, providerRegistry, intelligentService)

	// Context Analysis service
	c.ContextAnalysis = application.NewKeywordAnalyzer(c.Log)

	// Context Service (Unified service replacing three context services)
	tokenCounter := &SimpleTokenCounter{}
	c.ContextService = contextservice.NewService(
		c.FileReader,
		tokenCounter,
		c.Bus,
		c.Log,
	)

	// Project Service (with interface to context service)
	c.ProjectService = projectservice.NewService(
		c.Log,
		c.Bus,
		c.TreeBuilder,
		c.GitRepo,
		c.ContextService, // Pass context service interface
	)

	// Initialize Task Protocol Services
	if err := c.initializeTaskProtocolServices(); err != nil {
		return fmt.Errorf("failed to initialize task protocol services: %w", err)
	}

	return nil
}

func (c *Container) createModelFetchers(ctx context.Context) domain.ModelFetcherRegistry {
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
				host = c.SettingsRepo.GetLocalAIHost()
			}

			models, err := config.ModelFetcher(ctx, apiKey, host, c.Log)
			if err != nil {
				c.Log.Warning(fmt.Sprintf("Failed to create %s client for model listing: %s", providerType, err.Error()))
				return nil, err
			}
			return models, nil
		}
	}

	return fetchers
}

func (c *Container) createProviderRegistry() map[string]domain.AIProviderFactory {
	resolveHost := func(providerType string) (string, error) {
		switch providerType {
		case "openrouter":
			return openRouterHost, nil
		case "localai":
			dto, err := c.SettingsService.GetSettingsDTO()
			if err != nil {
				return "", err
			}
			return dto.LocalAIHost, nil
		default:
			return "", nil
		}
	}

	return ai.NewAIProviderFactoryRegistry(c.Log, openRouterHost, resolveHost)
}

// SimpleTokenCounter provides basic token estimation for context services
type SimpleTokenCounter struct{}

func (s *SimpleTokenCounter) CountTokens(text string) int {
	// Simple approximation: 1 token ≈ 4 characters
	return len(text) / 4
}

// GetProjectService returns the new project service
func (c *Container) GetProjectService() *projectservice.Service {
	return c.ProjectService
}

// GetContextService returns the new unified context service
func (c *Container) GetContextService() *contextservice.Service {
	return c.ContextService
}

// initializeTaskProtocolServices initializes all Task Protocol related services
func (c *Container) initializeTaskProtocolServices() error {
	// Initialize Error Analyzer
	c.ErrorAnalyzer = application.NewErrorAnalyzer(c.Log)

	// Initialize Correction Engine
	// We need a FileSystemProvider - for now we'll create a simple wrapper
	fileSystemProvider := &FileSystemProviderImpl{}
	c.CorrectionEngine = application.NewCorrectionEngine(c.Log, fileSystemProvider)

	// Initialize Task Protocol Config Service
	c.TaskProtocolConfigService = application.NewTaskProtocolConfigService(c.Log, fileSystemProvider)

	// We need to initialize some basic services that Task Protocol depends on
	if err := c.initializeBasicServices(); err != nil {
		return err
	}

	// Initialize Task Protocol Service
	c.TaskProtocolService = application.NewTaskProtocolService(
		c.Log,
		c.VerificationPipelineService,
		c.StaticAnalyzerService,
		c.TestService,
		c.BuildService,
		c.GuardrailService,
		c.AIService.GetIntelligentService(), // Get the intelligent service
		c.ErrorAnalyzer,
		c.CorrectionEngine,
	)

	// Update VerificationPipelineService with Task Protocol
	// Since we can't easily modify the existing constructor, we'll create a new instance
	fileSystemWriter := &FileSystemWriterImpl{}
	formatterService := application.NewFormatterService(c.Log, &CommandRunnerImpl{})

	c.VerificationPipelineService = application.NewVerificationPipelineService(
		c.Log,
		c.BuildService,
		c.TestService,
		c.StaticAnalyzerService,
		formatterService,
		fileSystemWriter,
		c.TaskProtocolService, // Pass the task protocol service
	)

	return nil
}

// initializeBasicServices initializes basic services needed for Task Protocol
func (c *Container) initializeBasicServices() error {
	// For now, we'll use the existing services if they exist
	// In a production environment, these would be properly initialized

	// Create minimal stub implementations if services don't exist
	// These should be replaced with proper service initialization

	return nil
}

// Infrastructure implementations for Task Protocol

// FileSystemProviderImpl implements domain.FileSystemProvider
type FileSystemProviderImpl struct{}

func (f *FileSystemProviderImpl) ReadFile(filename string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *FileSystemProviderImpl) WriteFile(filename string, data []byte, perm int) error {
	return fmt.Errorf("not implemented")
}

func (f *FileSystemProviderImpl) MkdirAll(path string, perm int) error {
	return fmt.Errorf("not implemented")
}

// FileSystemWriterImpl implements domain.FileSystemWriter
type FileSystemWriterImpl struct{}

func (f *FileSystemWriterImpl) WriteFile(filename string, data []byte, perm int) error {
	return fmt.Errorf("not implemented")
}

func (f *FileSystemWriterImpl) MkdirAll(path string, perm int) error {
	return fmt.Errorf("not implemented")
}

func (f *FileSystemWriterImpl) Remove(name string) error {
	return fmt.Errorf("not implemented")
}

func (f *FileSystemWriterImpl) RemoveAll(path string) error {
	return fmt.Errorf("not implemented")
}

// CommandRunnerImpl implements domain.CommandRunner
type CommandRunnerImpl struct{}

func (c *CommandRunnerImpl) RunCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *CommandRunnerImpl) RunCommandInDir(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

// Service stubs for testing - these should be replaced with actual implementations

type TestServiceStub struct {
	log domain.Logger
}

func (t *TestServiceStub) RunSmokeTests(ctx context.Context, projectPath, language string) ([]*domain.TestResult, error) {
	return []*domain.TestResult{}, nil
}

func (t *TestServiceStub) ValidateTestResults(results []*domain.TestResult) *domain.TestValidationResult {
	return &domain.TestValidationResult{Success: true}
}

type StaticAnalyzerServiceStub struct {
	log domain.Logger
}

func (s *StaticAnalyzerServiceStub) AnalyzeProject(ctx context.Context, projectPath string, languages []string) (interface{}, error) {
	return nil, nil
}

type BuildServiceStub struct {
	log domain.Logger
}

func (b *BuildServiceStub) ValidateProject(ctx context.Context, projectPath string, languages []string) (*domain.ProjectValidationResult, error) {
	return &domain.ProjectValidationResult{Success: true}, nil
}

func (b *BuildServiceStub) DetectLanguages(ctx context.Context, projectPath string) ([]string, error) {
	return []string{"go"}, nil
}

func (b *BuildServiceStub) GetSupportedLanguages() []string {
	return []string{"go", "typescript", "javascript"}
}

type GuardrailServiceStub struct {
	log domain.Logger
}

func (g *GuardrailServiceStub) ValidateTask(taskID string, files []string, linesChanged int64) (*domain.TaskValidationResult, error) {
	return &domain.TaskValidationResult{Valid: true}, nil
}

// Implement additional methods required by GuardrailService interface
func (g *GuardrailServiceStub) AddBudgetPolicy(policy interface{}) {}
func (g *GuardrailServiceStub) ValidatePath(path string) ([]domain.GuardrailViolation, error) {
	return []domain.GuardrailViolation{}, nil
}
func (g *GuardrailServiceStub) ValidateBudget(budgetType domain.BudgetType, current int64) ([]domain.BudgetViolation, error) {
	return []domain.BudgetViolation{}, nil
}
func (g *GuardrailServiceStub) EnableEphemeralMode(taskID string, taskType string, duration time.Duration) error {
	return nil
}
func (g *GuardrailServiceStub) DisableEphemeralMode() {}
func (g *GuardrailServiceStub) GetPolicies() ([]domain.GuardrailPolicy, error) {
	return []domain.GuardrailPolicy{}, nil
}
func (g *GuardrailServiceStub) GetBudgetPolicies() ([]domain.BudgetPolicy, error) {
	return []domain.BudgetPolicy{}, nil
}
func (g *GuardrailServiceStub) AddPolicy(policy domain.GuardrailPolicy) error             { return nil }
func (g *GuardrailServiceStub) RemovePolicy(policyID string) error                        { return nil }
func (g *GuardrailServiceStub) UpdatePolicy(policy domain.GuardrailPolicy) error          { return nil }
func (g *GuardrailServiceStub) RemoveBudgetPolicy(policyID string) error                  { return nil }
func (g *GuardrailServiceStub) UpdateBudgetPolicy(policy domain.BudgetPolicy) error       { return nil }
func (g *GuardrailServiceStub) GetConfig() domain.GuardrailConfig                         { return domain.GuardrailConfig{} }
func (g *GuardrailServiceStub) UpdateConfig(config domain.GuardrailConfig) error          { return nil }
func (g *GuardrailServiceStub) SetTaskflowService(taskflowService domain.TaskflowService) {}
