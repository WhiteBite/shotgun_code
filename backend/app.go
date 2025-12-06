package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/application"
	"shotgun_code/cmd/app"
	"shotgun_code/domain"
	"shotgun_code/handlers"
	"shotgun_code/infrastructure/git"
	"shotgun_code/infrastructure/wailsbridge"
	contextservice "shotgun_code/internal/context"
	projectservice "shotgun_code/internal/project"
	"strings"
	"time"
)

type contextAnalysisService interface {
	domain.ContextAnalyzer
	AnalyzeTaskAndCollectContext(ctx context.Context, task string, allFiles []*domain.FileNode, rootDir string) (*domain.ContextAnalysisResult, error)
}

type App struct {
	ctx       context.Context
	log       domain.Logger
	bridge    *wailsbridge.Bridge
	container *app.AppContainer

	// Handlers (new architecture) - primary delegation targets
	projectHandler  *handlers.ProjectHandler
	contextHandler  *handlers.ContextHandler
	aiHandler       *handlers.AIHandler
	analysisHandler *handlers.AnalysisHandler
	settingsHandler *handlers.SettingsHandler
	taskflowHandler *handlers.TaskflowHandler

	// Services (kept for methods not yet migrated to handlers)
	projectService        *projectservice.Service
	contextService        *contextservice.Service
	aiService             *application.AIService
	settingsService       *application.SettingsService
	contextAnalysis       contextAnalysisService
	symbolGraph           *application.SymbolGraphService
	testService           domain.ITestService
	staticAnalyzerService domain.IStaticAnalyzerService
	sbomService           *application.SBOMService
	repairService         domain.RepairService
	guardrailService      domain.GuardrailService
	taskflowService       domain.TaskflowService
	uxMetricsService      domain.UXMetricsService
	applyService          *application.ApplyService
	diffService           *application.DiffService
	buildService          domain.IBuildService
	fileWatcher           domain.FileSystemWatcher
	gitRepo               domain.GitRepository
	exportService         *application.ExportService
	fileReader            domain.FileContentReader
	reportService         *application.ReportService

	// Task Protocol Services
	taskProtocolService         domain.TaskProtocolService
	taskProtocolConfigService   *application.TaskProtocolConfigService
	taskflowProtocolIntegration *application.TaskflowProtocolIntegration

	// Qwen Services
	qwenHandler *handlers.QwenHandler
}

func (a *App) startup(ctx context.Context, container *app.AppContainer) {
	a.ctx = ctx
	a.log = container.Log
	a.bridge = container.Bridge
	a.container = container

	// Initialize handlers from container
	a.projectHandler = container.ProjectHandler
	a.contextHandler = container.ContextHandler
	a.aiHandler = container.AIHandler
	a.analysisHandler = container.AnalysisHandler
	a.settingsHandler = container.SettingsHandler
	a.taskflowHandler = container.TaskflowHandler

	// Services (for methods not yet migrated)
	a.projectService = container.ProjectService
	a.contextService = container.ContextService
	a.aiService = container.AIService
	a.settingsService = container.SettingsService
	if analyzer, ok := container.ContextAnalysis.(contextAnalysisService); ok {
		a.contextAnalysis = analyzer
	} else {
		a.contextAnalysis = nil
		if container.ContextAnalysis != nil {
			a.log.Warning("context analysis service does not implement required AnalyzeTaskAndCollectContext method")
		}
	}
	a.symbolGraph = container.SymbolGraph
	a.testService = container.TestService
	a.staticAnalyzerService = container.StaticAnalyzerService
	a.sbomService = container.SBOMService
	a.repairService = container.RepairService
	a.guardrailService = container.GuardrailService
	a.taskflowService = container.TaskflowService
	a.uxMetricsService = container.UXMetricsService
	a.applyService = container.ApplyService
	a.diffService = container.DiffService
	a.buildService = container.BuildService
	a.fileWatcher = container.Watcher
	a.gitRepo = container.GitRepo
	a.exportService = container.ExportService
	a.fileReader = container.FileReader
	a.reportService = container.ReportService

	// Task Protocol Services
	a.taskProtocolService = container.TaskProtocolService
	a.taskProtocolConfigService = container.TaskProtocolConfigService
	a.taskflowProtocolIntegration = container.TaskflowProtocolIntegration

	// Qwen Handler
	a.qwenHandler = container.QwenHandler
}

func (a *App) domReady(ctx context.Context) {
	a.ctx = ctx
	a.bridge.SetWailsContext(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.ctx = ctx
	
	// Shutdown container services first
	if a.container != nil {
		if err := a.container.Shutdown(ctx); err != nil {
			a.log.Warning("Container shutdown error: " + err.Error())
		}
	}
	
	// Stop file watcher
	a.fileWatcher.Stop()
}

func (a *App) SelectDirectory() (string, error) {
	return a.bridge.OpenDirectoryDialog()
}

// GetCurrentDirectory возвращает текущую рабочую директорию
func (a *App) GetCurrentDirectory() (string, error) {
	return os.Getwd()
}

func (a *App) ListFiles(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*domain.FileNode, error) {
	return a.projectHandler.ListFiles(dirPath, useGitignore, useCustomIgnore)
}

func (a *App) RequestShotgunContextGeneration(rootDir string, includedPaths []string) {
	a.projectHandler.GenerateContext(a.ctx, rootDir, includedPaths)
}

func (a *App) GenerateCode(systemPrompt, userPrompt string) (string, error) {
	return a.aiHandler.GenerateCode(a.ctx, systemPrompt, userPrompt)
}

// GenerateCodeStream generates code with streaming response via Wails events
func (a *App) GenerateCodeStream(systemPrompt, userPrompt string) {
	go func() {
		a.aiHandler.GenerateCodeStream(a.ctx, systemPrompt, userPrompt, func(chunk domain.StreamChunk) {
			a.bridge.Emit("ai:stream:chunk", chunk)
		})
	}()
}

// BuildSymbolGraph строит граф символов для проекта
func (a *App) BuildSymbolGraph(projectRoot, language string) (*domain.SymbolGraph, error) {
	return a.analysisHandler.BuildSymbolGraph(a.ctx, projectRoot, language)
}

// GetSymbolSuggestions возвращает предложения символов
func (a *App) GetSymbolSuggestions(query, language string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	return a.analysisHandler.GetSymbolSuggestions(a.ctx, query, language, graph)
}

// GetSymbolDependencies возвращает зависимости символа
func (a *App) GetSymbolDependencies(symbolID, language string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	return a.analysisHandler.GetSymbolDependencies(a.ctx, symbolID, language, graph)
}

// GetSymbolDependents возвращает символы, зависящие от указанного
func (a *App) GetSymbolDependents(symbolID, language string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	return a.analysisHandler.GetSymbolDependents(a.ctx, symbolID, language, graph)
}

// ApplyEdits применяет правки из Edits JSON
func (a *App) ApplyEdits(edits *domain.EditsJSON) ([]*domain.ApplyResult, error) {
	return a.applyService.ApplyEdits(a.ctx, edits)
}

// ApplySingleEdit применяет одну правку
func (a *App) ApplySingleEdit(edit *domain.Edit) (*domain.ApplyResult, error) {
	return a.applyService.ApplySingleEdit(a.ctx, edit)
}

// ValidateEdits проверяет корректность правок
func (a *App) ValidateEdits(edits *domain.EditsJSON) error {
	return a.applyService.ValidateEdits(a.ctx, edits)
}

// RollbackEdits откатывает правки
func (a *App) RollbackEdits(results []*domain.ApplyResult) error {
	return a.applyService.RollbackEdits(a.ctx, results)
}

// GenerateDiff генерирует diff между двумя состояниями
func (a *App) GenerateDiff(beforePath, afterPath string, format domain.DiffFormat) (*domain.DiffResult, error) {
	return a.diffService.GenerateDiff(a.ctx, beforePath, afterPath, format)
}

// GenerateDiffFromResults генерирует diff из результатов применения правок
func (a *App) GenerateDiffFromResults(results []*domain.ApplyResult, format domain.DiffFormat) (*domain.DiffResult, error) {
	return a.diffService.GenerateDiffFromResults(a.ctx, results, format)
}

// GenerateDiffFromEdits генерирует diff из Edits JSON
func (a *App) GenerateDiffFromEdits(edits *domain.EditsJSON, format domain.DiffFormat) (*domain.DiffResult, error) {
	return a.diffService.GenerateDiffFromEdits(a.ctx, edits, format)
}

// PublishDiff публикует diff
func (a *App) PublishDiff(diff *domain.DiffResult) error {
	return a.diffService.PublishDiff(a.ctx, diff)
}

// GenerateAndPublishDiff генерирует и публикует diff
func (a *App) GenerateAndPublishDiff(beforePath, afterPath string, format domain.DiffFormat) (*domain.DiffResult, error) {
	return a.diffService.GenerateAndPublishDiff(a.ctx, beforePath, afterPath, format)
}

// Build выполняет сборку проекта
func (a *App) Build(projectPath, language string) (*domain.BuildResult, error) {
	return a.analysisHandler.Build(a.ctx, projectPath, language)
}

// TypeCheck выполняет проверку типов
func (a *App) TypeCheck(projectPath, language string) (*domain.TypeCheckResult, error) {
	return a.analysisHandler.TypeCheck(a.ctx, projectPath, language)
}

// BuildAndTypeCheck выполняет сборку и проверку типов
func (a *App) BuildAndTypeCheck(projectPath, language string) (*domain.BuildResult, *domain.TypeCheckResult, error) {
	return a.analysisHandler.BuildAndTypeCheck(a.ctx, projectPath, language)
}

// ValidateProject выполняет полную валидацию проекта
func (a *App) ValidateProject(projectPath string, languages []string) (*domain.ProjectValidationResult, error) {
	return a.analysisHandler.ValidateProject(a.ctx, projectPath, languages)
}

// DetectLanguages определяет языки в проекте
func (a *App) DetectLanguages(projectPath string) ([]string, error) {
	return a.analysisHandler.DetectLanguages(a.ctx, projectPath)
}

// RunTests выполняет тесты согласно конфигурации
func (a *App) RunTests(config *domain.TestConfig) ([]*domain.TestResult, error) {
	return a.analysisHandler.RunTests(a.ctx, config)
}

// RunTargetedTests выполняет целевые тесты для затронутых файлов
func (a *App) RunTargetedTests(config *domain.TestConfig, changedFiles []string) ([]*domain.TestResult, error) {
	return a.analysisHandler.RunTargetedTests(a.ctx, config, changedFiles)
}

// DiscoverTests обнаруживает тесты в проекте
func (a *App) DiscoverTests(projectPath, language string) (*domain.TestSuite, error) {
	return a.analysisHandler.DiscoverTests(a.ctx, projectPath, language)
}

// BuildAffectedGraph строит граф затронутых файлов
func (a *App) BuildAffectedGraph(changedFiles []string, projectPath string) (*domain.AffectedGraph, error) {
	return a.analysisHandler.BuildAffectedGraph(a.ctx, changedFiles, projectPath)
}

// RunSmokeTests выполняет только smoke тесты
func (a *App) RunSmokeTests(projectPath, language string) ([]*domain.TestResult, error) {
	return a.analysisHandler.RunSmokeTests(a.ctx, projectPath, language)
}

// RunUnitTests выполняет только unit тесты
func (a *App) RunUnitTests(projectPath, language string) ([]*domain.TestResult, error) {
	return a.analysisHandler.RunUnitTests(a.ctx, projectPath, language)
}

// RunIntegrationTests выполняет только integration тесты
func (a *App) RunIntegrationTests(projectPath, language string) ([]*domain.TestResult, error) {
	return a.analysisHandler.RunIntegrationTests(a.ctx, projectPath, language)
}

// ValidateTestResults валидирует результаты тестов
func (a *App) ValidateTestResults(results []*domain.TestResult) *domain.TestValidationResult {
	return a.analysisHandler.ValidateTestResults(results)
}

// AnalyzeProject выполняет статический анализ проекта
func (a *App) AnalyzeProject(projectPath string, languages []string) (*domain.StaticAnalysisReport, error) {
	return a.analysisHandler.AnalyzeProject(a.ctx, projectPath, languages)
}

// AnalyzeFile выполняет статический анализ одного файла
func (a *App) AnalyzeFile(filePath, language string) (*domain.StaticAnalysisResult, error) {
	return a.analysisHandler.AnalyzeFile(a.ctx, filePath, language)
}

// AnalyzeGoProject выполняет статический анализ Go проекта
func (a *App) AnalyzeGoProject(projectPath string) (*domain.StaticAnalysisResult, error) {
	return a.analysisHandler.AnalyzeGoProject(a.ctx, projectPath)
}

// AnalyzeTypeScriptProject выполняет статический анализ TypeScript проекта
func (a *App) AnalyzeTypeScriptProject(projectPath string) (*domain.StaticAnalysisResult, error) {
	return a.analysisHandler.AnalyzeTypeScriptProject(a.ctx, projectPath)
}

// AnalyzeJavaScriptProject выполняет статический анализ JavaScript проекта
func (a *App) AnalyzeJavaScriptProject(projectPath string) (*domain.StaticAnalysisResult, error) {
	return a.analysisHandler.AnalyzeJavaScriptProject(a.ctx, projectPath)
}

// GetSupportedAnalyzers возвращает поддерживаемые анализаторы
func (a *App) GetSupportedAnalyzers() []domain.StaticAnalyzerType {
	return a.analysisHandler.GetSupportedAnalyzers()
}

// ValidateAnalysisResults валидирует результаты статического анализа
func (a *App) ValidateAnalysisResults(results map[string]*domain.StaticAnalysisResult) *domain.StaticAnalysisValidationResult {
	return a.analysisHandler.ValidateAnalysisResults(results)
}

// GenerateSBOM генерирует SBOM для проекта
func (a *App) GenerateSBOM(projectPath string, format domain.SBOMFormat) (*domain.SBOMResult, error) {
	return a.analysisHandler.GenerateSBOM(a.ctx, projectPath, format)
}

// ScanVulnerabilities сканирует уязвимости в проекте
func (a *App) ScanVulnerabilities(projectPath string) (*domain.VulnerabilityScanResult, error) {
	return a.analysisHandler.ScanVulnerabilities(a.ctx, projectPath)
}

// ScanLicenses сканирует лицензии в проекте
func (a *App) ScanLicenses(projectPath string) (*domain.LicenseScanResult, error) {
	return a.analysisHandler.ScanLicenses(a.ctx, projectPath)
}

// GenerateComplianceReport генерирует отчет о соответствии
func (a *App) GenerateComplianceReport(projectPath string, requirements *domain.ComplianceRequirements) (*domain.ComplianceReport, error) {
	return a.analysisHandler.GenerateComplianceReport(a.ctx, projectPath, requirements)
}

// GetSupportedSBOMFormats возвращает поддерживаемые форматы SBOM
func (a *App) GetSupportedSBOMFormats() []domain.SBOMFormat {
	return a.analysisHandler.GetSupportedSBOMFormats()
}

// ValidateSBOM валидирует SBOM
func (a *App) ValidateSBOM(sbomPath string, format domain.SBOMFormat) error {
	return a.analysisHandler.ValidateSBOM(a.ctx, sbomPath, format)
}

// ExecuteRepair выполняет repair цикл
func (a *App) ExecuteRepair(projectPath, errorOutput, language string, maxAttempts int) (*domain.RepairResult, error) {
	req := domain.RepairRequest{
		ProjectPath: projectPath,
		ErrorOutput: errorOutput,
		Language:    language,
		MaxAttempts: maxAttempts,
	}
	return a.repairService.ExecuteRepair(a.ctx, req)
}

// GetAvailableRepairRules возвращает доступные правила для языка
func (a *App) GetAvailableRepairRules(language string) ([]domain.RepairRule, error) {
	return a.repairService.GetAvailableRules(language)
}

// AddRepairRule добавляет новое правило
func (a *App) AddRepairRule(rule domain.RepairRule) error {
	return a.repairService.AddRule(rule)
}

// RemoveRepairRule удаляет правило
func (a *App) RemoveRepairRule(ruleID string) error {
	return a.repairService.RemoveRule(ruleID)
}

// ValidateRepairRule проверяет корректность правила
func (a *App) ValidateRepairRule(rule domain.RepairRule) error {
	return a.repairService.ValidateRule(rule)
}

// ValidatePath проверяет путь на соответствие политикам
func (a *App) ValidatePath(path string) ([]domain.GuardrailViolation, error) {
	return a.guardrailService.ValidatePath(path)
}

// ValidateBudget проверяет бюджетные ограничения
func (a *App) ValidateBudget(budgetType domain.BudgetType, current int64) ([]domain.BudgetViolation, error) {
	return a.guardrailService.ValidateBudget(budgetType, current)
}

// GetGuardrailPolicies возвращает все политики
func (a *App) GetGuardrailPolicies() ([]domain.GuardrailPolicy, error) {
	return a.guardrailService.GetPolicies()
}

// GetBudgetPolicies возвращает бюджетные политики
func (a *App) GetBudgetPolicies() ([]domain.BudgetPolicy, error) {
	return a.guardrailService.GetBudgetPolicies()
}

// AddGuardrailPolicy добавляет новую политику
func (a *App) AddGuardrailPolicy(policy domain.GuardrailPolicy) error {
	return a.guardrailService.AddPolicy(policy)
}

// RemoveGuardrailPolicy удаляет политику
func (a *App) RemoveGuardrailPolicy(policyID string) error {
	return a.guardrailService.RemovePolicy(policyID)
}

// UpdateGuardrailPolicy обновляет политику
func (a *App) UpdateGuardrailPolicy(policy domain.GuardrailPolicy) error {
	return a.guardrailService.UpdatePolicy(policy)
}

// AddBudgetPolicy добавляет бюджетную политику
func (a *App) AddBudgetPolicy(policy domain.BudgetPolicy) error {
	return a.guardrailService.AddBudgetPolicy(policy)
}

// RemoveBudgetPolicy удаляет бюджетную политику
func (a *App) RemoveBudgetPolicy(policyID string) error {
	return a.guardrailService.RemoveBudgetPolicy(policyID)
}

// UpdateBudgetPolicy обновляет бюджетную политику
func (a *App) UpdateBudgetPolicy(policy domain.BudgetPolicy) error {
	return a.guardrailService.UpdateBudgetPolicy(policy)
}

// LoadTasks загружает задачи из plan.yaml
func (a *App) LoadTasks() ([]domain.Task, error) {
	return a.taskflowService.LoadTasks()
}

// GetTaskStatus возвращает статус задачи
func (a *App) GetTaskStatus(taskID string) (*domain.TaskStatus, error) {
	return a.taskflowService.GetTaskStatus(taskID)
}

// UpdateTaskStatus обновляет статус задачи
func (a *App) UpdateTaskStatus(taskID string, state domain.TaskState, message string) error {
	return a.taskflowService.UpdateTaskStatus(taskID, state, message)
}

// ExecuteTask выполняет задачу
func (a *App) ExecuteTask(taskID string) error {
	return a.taskflowService.ExecuteTask(taskID)
}

// ExecuteTaskflow выполняет весь taskflow
func (a *App) ExecuteTaskflow() error {
	return a.taskflowService.ExecuteTaskflow()
}

// GetReadyTasks возвращает готовые к выполнению задачи
func (a *App) GetReadyTasks() ([]domain.Task, error) {
	return a.taskflowService.GetReadyTasks()
}

// GetTaskDependencies возвращает зависимости задачи
func (a *App) GetTaskDependencies(taskID string) ([]domain.Task, error) {
	return a.taskflowService.GetTaskDependencies(taskID)
}

// ValidateTaskflow проверяет корректность taskflow
func (a *App) ValidateTaskflow() error {
	return a.taskflowService.ValidateTaskflow()
}

// GetTaskflowProgress возвращает прогресс выполнения
func (a *App) GetTaskflowProgress() (float64, error) {
	return a.taskflowService.GetTaskflowProgress()
}

// ResetTaskflow сбрасывает taskflow
func (a *App) ResetTaskflow() error {
	return a.taskflowService.ResetTaskflow()
}

// GenerateIntelligentCode выполняет интеллектуальную генерацию кода
func (a *App) GenerateIntelligentCode(task, context, optionsJson string) (string, error) {
	return a.aiHandler.GenerateIntelligentCode(a.ctx, task, context, optionsJson)
}

// GenerateCodeWithOptions генерирует код с дополнительными опциями
func (a *App) GenerateCodeWithOptions(systemPrompt, userPrompt, optionsJson string) (string, error) {
	return a.aiHandler.GenerateCodeWithOptions(a.ctx, systemPrompt, userPrompt, optionsJson)
}

// GetProviderInfo возвращает информацию о текущем провайдере
func (a *App) GetProviderInfo() (string, error) {
	return a.aiHandler.GetProviderInfo(a.ctx)
}

// ListAvailableModels возвращает список доступных моделей
func (a *App) ListAvailableModels() ([]string, error) {
	return a.aiHandler.ListAvailableModels(a.ctx)
}

func (a *App) SuggestContextFiles(task string, allFiles []*domain.FileNode) ([]string, error) {
	return a.aiHandler.SuggestContextFiles(a.ctx, task, allFiles)
}

// AnalyzeTaskAndCollectContext анализирует задачу и собирает релевантный контекст
func (a *App) AnalyzeTaskAndCollectContext(task string, allFilesJson string, rootDir string) (string, error) {
	return a.aiHandler.AnalyzeTaskAndCollectContext(a.ctx, task, allFilesJson, rootDir)
}

// TestBackend простой тест для проверки работы backend
func (a *App) TestBackend(allFilesJson string, rootDir string) (string, error) {
	var allFiles []*domain.FileNode
	if err := json.Unmarshal([]byte(allFilesJson), &allFiles); err != nil {
		return "", fmt.Errorf("failed to parse files JSON: %w", err)
	}

	// Простой тест - возвращаем информацию о файлах
	testResult := map[string]interface{}{
		"status":     "success",
		"filesCount": len(allFiles),
		"rootDir":    rootDir,
		"timestamp":  time.Now().Unix(),
		"message":    "Backend работает корректно",
	}

	if len(allFiles) > 0 {
		testResult["sampleFile"] = allFiles[0].RelPath
	}

	resultJson, err := json.Marshal(testResult)
	if err != nil {
		return "", fmt.Errorf("failed to marshal test result: %w", err)
	}

	return string(resultJson), nil
}

func (a *App) GetSettings() (domain.SettingsDTO, error) {
	return a.settingsHandler.GetSettings()
}

func (a *App) SaveSettings(settingsJson string) error {
	return a.settingsHandler.SaveSettings(settingsJson)
}

func (a *App) RefreshAIModels(provider, apiKey string) error {
	return a.settingsHandler.RefreshAIModels(provider, apiKey)
}

func (a *App) StartFileWatcher(rootDirPath string) error {
	return a.projectHandler.StartFileWatcher(rootDirPath)
}

func (a *App) StopFileWatcher() {
	a.projectHandler.StopFileWatcher()
}

func (a *App) IsGitAvailable() bool {
	return a.projectHandler.IsGitAvailable()
}

func (a *App) GetUncommittedFiles(projectRoot string) ([]domain.FileStatus, error) {
	return a.projectHandler.GetUncommittedFiles(projectRoot)
}

func (a *App) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	return a.projectHandler.GetRichCommitHistory(projectRoot, branchName, limit)
}

func (a *App) GetFileContentAtCommit(projectRoot, filePath, commitHash string) (string, error) {
	return a.projectHandler.GetFileContentAtCommit(projectRoot, filePath, commitHash)
}

func (a *App) GetGitignoreContent(projectRoot string) (string, error) {
	return a.projectHandler.GetGitignoreContent(projectRoot)
}

func (a *App) ReadFileContent(rootDir, relPath string) (string, error) {
	return a.projectHandler.ReadFileContent(a.ctx, rootDir, relPath)
}

func (a *App) ExportContext(settingsJson string) (domain.ExportResult, error) {
	var settings domain.ExportSettings
	if err := json.Unmarshal([]byte(settingsJson), &settings); err != nil {
		validationErr := domain.NewValidationError("failed to parse export settings", map[string]interface{}{
			"originalError": err.Error(),
			"settingsJson":  settingsJson,
		})
		return domain.ExportResult{}, a.transformError(validationErr)
	}

	result, err := a.exportService.Export(a.ctx, settings)
	if err != nil {
		return domain.ExportResult{}, a.transformError(err)
	}

	// Если результат содержит FilePath, то это большой файл - нужно переместить в Downloads
	if result.FilePath != "" {
		// Для больших файлов возвращаем путь как есть, фронтенд сам решит что делать
		// В будущем можно добавить логику перемещения в Downloads
		return result, nil
	}

	return result, nil
}

// CleanupTempFiles - утилита для очистки временных файлов экспорта
func (a *App) CleanupTempFiles(filePath string) error {
	if filePath == "" {
		return nil
	}

	// Проверяем что это действительно temp файл
	if !strings.Contains(filePath, "shotgun-export-") {
		return fmt.Errorf("not a temp export file")
	}

	// Удаляем весь temp каталог
	tempDir := filepath.Dir(filePath)
	return os.RemoveAll(tempDir)
}

// GenerateWhyViewReport генерирует отчёт "почему эти файлы"
func (a *App) GenerateWhyViewReport(taskID string, files []string) (*domain.WhyViewReport, error) {
	return a.uxMetricsService.GenerateWhyViewReport(taskID, files, nil)
}

// GenerateTimeToGreenMetrics генерирует метрики time_to_green
func (a *App) GenerateTimeToGreenMetrics(taskID string) (*domain.TimeToGreenMetrics, error) {
	return a.uxMetricsService.GenerateTimeToGreenMetrics(taskID)
}

// GenerateDerivedDiffReport генерирует отчёт о derived diff
func (a *App) GenerateDerivedDiffReport(taskID string, originalDiff, derivedDiff string) (*domain.DerivedDiffReport, error) {
	return a.uxMetricsService.GenerateDerivedDiffReport(taskID, originalDiff, derivedDiff)
}

// GeneratePerformanceMetrics генерирует метрики производительности
func (a *App) GeneratePerformanceMetrics(taskID string) (*domain.PerformanceMetrics, error) {
	return a.uxMetricsService.GeneratePerformanceMetrics(taskID)
}

// GetUXReport возвращает UX отчёт
func (a *App) GetUXReport(reportID string) (*domain.UXReport, error) {
	return a.uxMetricsService.GetUXReport(reportID)
}

// SaveUXReport сохраняет UX отчёт
func (a *App) SaveUXReport(report *domain.UXReport) error {
	return a.uxMetricsService.SaveUXReport(report)
}

// GetUXReports возвращает все UX отчёты
func (a *App) GetUXReports(reportType domain.UXReportType) ([]*domain.UXReport, error) {
	return a.uxMetricsService.GetUXReports(reportType)
}

// DeleteUXReport удаляет UX отчёт
func (a *App) DeleteUXReport(reportID string) error {
	return a.uxMetricsService.DeleteUXReport(reportID)
}

// GetMetricsSummary возвращает сводку метрик
func (a *App) GetMetricsSummary() (map[string]interface{}, error) {
	return a.uxMetricsService.GetMetricsSummary()
}

// StartAutonomousTask запускает автономную задачу
func (a *App) StartAutonomousTask(requestJson string) (string, error) {
	var request domain.AutonomousTaskRequest
	if err := json.Unmarshal([]byte(requestJson), &request); err != nil {
		validationErr := domain.NewValidationError("Invalid JSON request format", map[string]interface{}{
			"originalError": err.Error(),
			"requestJson":   requestJson,
		})
		return "", a.transformDomainError(validationErr)
	}

	result, err := a.taskflowService.StartAutonomousTask(a.ctx, request)
	if err != nil {
		// Transform domain errors for frontend consumption
		if domainErr, ok := err.(*domain.DomainError); ok {
			return "", a.transformDomainError(domainErr)
		}
		return "", domain.NewInternalError("Unexpected error occurred", err)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		configErr := domain.NewConfigurationError("Failed to serialize response", err)
		return "", a.transformDomainError(configErr)
	}

	return string(resultJson), nil
}

// transformError transforms any error into a structured domain error
// If the error is already a domain error, it's returned as-is
// Otherwise, it's wrapped in a generic internal error
func (a *App) transformError(err error) error {
	if err == nil {
		return nil
	}

	// If it's already a domain error, return as-is
	if domainErr, ok := err.(*domain.DomainError); ok {
		return a.transformDomainError(domainErr)
	}

	// Wrap generic errors as internal errors
	return a.transformDomainError(domain.NewInternalError("An unexpected error occurred", err))
}

// transformDomainError transforms domain errors into frontend-friendly format
func (a *App) transformDomainError(domainErr *domain.DomainError) error {
	// Create frontend-friendly error structure
	frontendError := map[string]interface{}{
		"code":        domainErr.Code,
		"message":     domainErr.Message,
		"recoverable": domainErr.Recoverable,
		"context":     domainErr.Context,
	}

	if domainErr.Cause != nil {
		frontendError["cause"] = domainErr.Cause.Error()
	}

	errorJson, _ := json.Marshal(frontendError)
	return fmt.Errorf("domain_error:%s", string(errorJson))
}

// CancelAutonomousTask отменяет автономную задачу
func (a *App) CancelAutonomousTask(taskId string) error {
	return a.taskflowService.CancelAutonomousTask(a.ctx, taskId)
}

// GetAutonomousTaskStatus получает статус автономной задачи
func (a *App) GetAutonomousTaskStatus(taskId string) (string, error) {
	status, err := a.taskflowService.GetAutonomousTaskStatus(a.ctx, taskId)
	if err != nil {
		return "", a.transformError(err)
	}

	statusJson, err := json.Marshal(status)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal status", err)
		return "", a.transformError(marshalErr)
	}

	return string(statusJson), nil
}

// ListReports получает список отчетов
func (a *App) ListReports(reportType string) (string, error) {
	reports, err := a.uxMetricsService.ListReports(a.ctx, reportType)
	if err != nil {
		return "", a.transformError(err)
	}

	reportsJson, err := json.Marshal(reports)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal reports", err)
		return "", a.transformError(marshalErr)
	}

	return string(reportsJson), nil
}

// GetReport получает конкретный отчет
func (a *App) GetReport(reportId string) (string, error) {
	report, err := a.uxMetricsService.GetReport(a.ctx, reportId)
	if err != nil {
		return "", a.transformError(err)
	}

	reportJson, err := json.Marshal(report)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal report", err)
		return "", a.transformError(marshalErr)
	}

	return string(reportJson), nil
}

// CRITICAL OOM FIX: BuildContext now returns ContextSummary instead of full context
// This prevents OOM issues by not storing large text content in memory
func (a *App) BuildContext(projectPath string, includedPaths []string, optionsJson string) (string, error) {
	if a.contextService == nil {
		return "", a.transformError(domain.NewConfigurationError("context service not available", nil))
	}

	var options domain.ContextBuildOptions
	if strings.TrimSpace(optionsJson) != "" {
		if err := json.Unmarshal([]byte(optionsJson), &options); err != nil {
			validationErr := domain.NewValidationError("failed to parse options JSON", map[string]interface{}{
				"originalError": err.Error(),
				"optionsJson":   optionsJson,
			})
			return "", a.transformError(validationErr)
		}
	}

	// Allow empty includedPaths for backward compatibility - use all files from projectPath
	if len(includedPaths) == 0 {
		a.log.Warning("BuildContext called with empty includedPaths - this may include all project files")
		// In future, we could auto-discover files here, but for now just warn
	}

	summary, err := a.contextService.BuildContextSummary(a.ctx, projectPath, includedPaths, &options)
	if err != nil {
		return "", a.transformError(err)
	}

	contextJson, err := json.Marshal(summary)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal context summary", err)
		return "", a.transformError(marshalErr)
	}

	return string(contextJson), nil
}

// GetContextContent returns paginated context content for memory-safe viewing
// This allows accessing context content without loading it all into memory
func (a *App) GetContextContent(contextID string, startLine int, lineCount int) (string, error) {
	if a.contextService == nil {
		return "", a.transformError(domain.NewConfigurationError("context service not available", nil))
	}

	if lineCount <= 0 {
		lineCount = 1000
	}

	chunk, err := a.contextService.ReadContextChunk(a.ctx, contextID, startLine, lineCount)
	if err != nil {
		return "", a.transformError(err)
	}

	chunkJson, err := json.Marshal(chunk)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal context chunk", err)
		return "", a.transformError(marshalErr)
	}

	return string(chunkJson), nil
}

// GetFullContextContent returns the full context content as a string
// This method is needed for export functionality when the entire context text is required at once
func (a *App) GetFullContextContent(contextID string) (string, error) {
	if a.contextService == nil {
		return "", a.transformError(domain.NewConfigurationError("context service not available", nil))
	}

	content, err := a.contextService.ReadContextContent(a.ctx, contextID)
	if err != nil {
		return "", a.transformError(err)
	}

	return content, nil
}

// Legacy BuildContextLegacy for backward compatibility - DEPRECATED
// The disk-based context architecture replaces this functionality
func (a *App) BuildContextLegacy() (string, error) {
	return "", a.transformError(domain.NewConfigurationError("legacy context building is no longer supported", nil))
}

// GetContext retrieves context metadata by ID without loading full content into memory
func (a *App) GetContext(contextID string) (string, error) {
	if a.contextService == nil {
		return "", a.transformError(domain.NewConfigurationError("context service not available", nil))
	}

	summary, err := a.contextService.GetContextSummary(a.ctx, contextID)
	if err != nil {
		return "", a.transformError(err)
	}

	contextJson, err := json.Marshal(summary)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal context summary", err)
		return "", a.transformError(marshalErr)
	}

	return string(contextJson), nil
}

// GetProjectContexts lists all stored context summaries for a project path
func (a *App) GetProjectContexts(projectPath string) (string, error) {
	if a.contextService == nil {
		return "", a.transformError(domain.NewConfigurationError("context service not available", nil))
	}

	summaries, err := a.contextService.GetProjectContextSummaries(a.ctx, projectPath)
	if err != nil {
		return "", a.transformError(err)
	}

	contextsJson, err := json.Marshal(summaries)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal context summaries", err)
		return "", a.transformError(marshalErr)
	}

	return string(contextsJson), nil
}

// DeleteContext removes context metadata and associated content from disk
func (a *App) DeleteContext(contextID string) error {
	if a.contextService == nil {
		return a.transformError(domain.NewConfigurationError("context service not available", nil))
	}

	if err := a.contextService.DeleteContext(a.ctx, contextID); err != nil {
		return a.transformError(err)
	}

	return nil
}

// BuildContextFromRequest builds context using provided options and returns ContextSummary
func (a *App) BuildContextFromRequest(projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextSummary, error) {
	if a.contextService == nil {
		return nil, a.transformError(domain.NewConfigurationError("context service not available", nil))
	}

	// Allow empty includedPaths for backward compatibility
	if len(includedPaths) == 0 {
		a.log.Warning("BuildContextFromRequest called with empty includedPaths")
	}

	if options == nil {
		options = &domain.ContextBuildOptions{}
	}

	// Note: We allow files from any location, not just within projectPath
	// This gives users flexibility to include files from multiple projects
	summary, err := a.contextService.BuildContextSummary(a.ctx, projectPath, includedPaths, options)
	if err != nil {
		return nil, a.transformError(err)
	}

	return summary, nil
}

// GetContextLines returns a chunk of context content between startLine and endLine inclusive
func (a *App) GetContextLines(contextID string, startLine, endLine int64) (string, error) {
	if a.contextService == nil {
		return "", a.transformError(domain.NewConfigurationError("context service not available", nil))
	}

	if endLine < startLine {
		return "", a.transformError(domain.NewValidationError("invalid line range", map[string]interface{}{
			"startLine": startLine,
			"endLine":   endLine,
		}))
	}

	lineCount := int(endLine-startLine) + 1
	chunk, err := a.contextService.ReadContextChunk(a.ctx, contextID, int(startLine), lineCount)
	if err != nil {
		return "", a.transformError(err)
	}

	chunkJson, err := json.Marshal(chunk)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal context chunk", err)
		return "", a.transformError(marshalErr)
	}

	return string(chunkJson), nil
}

// CreateStreamingContext delegates to BuildContext to create a disk-backed context summary
func (a *App) CreateStreamingContext(projectPath string, includedPaths []string, optionsJson string) (string, error) {
	return a.BuildContext(projectPath, includedPaths, optionsJson)
}

// GetStreamingContext returns context summary metadata for streaming compatibility
func (a *App) GetStreamingContext(contextID string) (string, error) {
	return a.GetContext(contextID)
}

// CloseStreamingContext removes a streaming context and associated resources
func (a *App) CloseStreamingContext(contextID string) error {
	return a.DeleteContext(contextID)
}

// SetSLAPolicy устанавливает SLA политику
func (a *App) SetSLAPolicy(policyJson string) error {
	var policy domain.SLAPolicy
	if err := json.Unmarshal([]byte(policyJson), &policy); err != nil {
		return fmt.Errorf("failed to parse SLA policy JSON: %w", err)
	}

	// Save SLA policy to settings or a dedicated SLA service
	// For now, we'll just log it since there's no dedicated SLA service yet
	a.log.Info(fmt.Sprintf("SLA Policy set: %s", policy.Name))
	return nil
}

// GetSLAPolicy получает текущую SLA политику
func (a *App) GetSLAPolicy() (string, error) {
	// Return a default SLA policy for now
	defaultPolicy := domain.SLAPolicy{
		Name:        "standard",
		Description: "Standard SLA policy",
		MaxTokens:   10000,
		MaxFiles:    50,
		MaxTime:     300,               // 5 minutes
		MaxMemory:   1024 * 1024 * 100, // 100MB
		MaxRetries:  3,
		Timeout:     30, // 30 seconds
	}

	policyJson, err := json.Marshal(defaultPolicy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SLA policy: %w", err)
	}

	return string(policyJson), nil
}

// === MISSING API ENDPOINTS FOR REPOSITORY IMPLEMENTATIONS ===

// GetFileStats returns file statistics
func (a *App) GetFileStats(filePath string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get file stats: %w", err)
	}

	stats := map[string]interface{}{
		"name":    fileInfo.Name(),
		"size":    fileInfo.Size(),
		"modTime": fileInfo.ModTime().Unix(),
		"isDir":   fileInfo.IsDir(),
		"mode":    fileInfo.Mode().String(),
	}

	statsJson, err := json.Marshal(stats)
	if err != nil {
		return "", fmt.Errorf("failed to marshal file stats: %w", err)
	}

	return string(statsJson), nil
}

// GetBranches returns all git branches
func (a *App) GetBranches(projectRoot string) (string, error) {
	branches, err := a.gitRepo.GetBranches(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get branches: %w", err)
	}

	branchesJson, err := json.Marshal(branches)
	if err != nil {
		return "", fmt.Errorf("failed to marshal branches: %w", err)
	}

	return string(branchesJson), nil
}

// GetCurrentBranch returns the current git branch
func (a *App) GetCurrentBranch(projectRoot string) (string, error) {
	branch, err := a.gitRepo.GetCurrentBranch(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return branch, nil
}

// IsGitRepository checks if the given path is a git repository
func (a *App) IsGitRepository(projectPath string) bool {
	return a.gitRepo.IsGitRepository(projectPath)
}

// CloneRepository clones a remote git repository
func (a *App) CloneRepository(url string) (string, error) {
	// Create temp directory for clone
	tempDir, err := os.MkdirTemp("", "shotgun-git-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Shallow clone for speed
	if err := a.gitRepo.CloneRepository(url, tempDir, 1); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}

// CheckoutBranch switches to a specific branch in a git repository
func (a *App) CheckoutBranch(projectPath, branch string) error {
	return a.gitRepo.CheckoutBranch(projectPath, branch)
}

// CheckoutCommit switches to a specific commit in a git repository
func (a *App) CheckoutCommit(projectPath, commitHash string) error {
	return a.gitRepo.CheckoutCommit(projectPath, commitHash)
}

// GetCommitHistory returns recent commits for selection
func (a *App) GetCommitHistory(projectPath string, limit int) (string, error) {
	commits, err := a.gitRepo.GetCommitHistory(projectPath, limit)
	if err != nil {
		return "", err
	}

	commitsJson, err := json.Marshal(commits)
	if err != nil {
		return "", fmt.Errorf("failed to marshal commits: %w", err)
	}

	return string(commitsJson), nil
}

// GetRemoteBranches returns all remote branches
func (a *App) GetRemoteBranches(projectPath string) (string, error) {
	branches, err := a.gitRepo.FetchRemoteBranches(projectPath)
	if err != nil {
		return "", err
	}

	branchesJson, err := json.Marshal(branches)
	if err != nil {
		return "", fmt.Errorf("failed to marshal branches: %w", err)
	}

	return string(branchesJson), nil
}

// CleanupTempRepository removes a temporary cloned repository
func (a *App) CleanupTempRepository(path string) error {
	// Safety check - only remove paths in temp directory
	tempDir := os.TempDir()
	if !strings.HasPrefix(path, tempDir) && !strings.Contains(path, "shotgun-git-") {
		return fmt.Errorf("refusing to remove non-temp path: %s", path)
	}
	return os.RemoveAll(path)
}

// ListFilesAtRef returns list of files at a specific branch/commit without checkout
func (a *App) ListFilesAtRef(projectPath, ref string) (string, error) {
	files, err := a.gitRepo.ListFilesAtRef(projectPath, ref)
	if err != nil {
		return "", err
	}
	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// GetFileAtRef returns file content at a specific branch/commit without checkout
func (a *App) GetFileAtRef(projectPath, filePath, ref string) (string, error) {
	return a.gitRepo.GetFileAtRef(projectPath, filePath, ref)
}

// BuildContextAtRef builds context from files at a specific git ref without checkout
func (a *App) BuildContextAtRef(projectPath string, files []string, ref string, optionsJson string) (string, error) {
	var contents []string

	for _, file := range files {
		content, err := a.gitRepo.GetFileAtRef(projectPath, file, ref)
		if err != nil {
			a.log.Info(fmt.Sprintf("Warning: Failed to read file %s at ref %s: %v", file, ref, err))
			continue
		}
		contents = append(contents, fmt.Sprintf("// File: %s (ref: %s)\n%s", file, ref, content))
	}

	result := strings.Join(contents, "\n\n---\n\n")
	return result, nil
}

// ============================================
// GitHub API Methods (no clone required)
// ============================================

// GitHubGetBranches returns branches for a GitHub repository via API
func (a *App) GitHubGetBranches(repoURL string) (string, error) {
	api := git.NewGitHubAPI()

	repo, err := git.ParseGitHubURL(repoURL)
	if err != nil {
		return "", err
	}

	branches, err := api.GetBranches(repo.Owner, repo.Name)
	if err != nil {
		return "", err
	}

	result, err := json.Marshal(branches)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// GitHubGetCommits returns commits for a GitHub repository branch via API
func (a *App) GitHubGetCommits(repoURL, branch string, limit int) (string, error) {
	api := git.NewGitHubAPI()

	repo, err := git.ParseGitHubURL(repoURL)
	if err != nil {
		return "", err
	}

	commits, err := api.GetCommits(repo.Owner, repo.Name, branch, limit)
	if err != nil {
		return "", err
	}

	result, err := json.Marshal(commits)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// GitHubListFiles returns file list for a GitHub repository at specific ref via API
func (a *App) GitHubListFiles(repoURL, ref string) (string, error) {
	api := git.NewGitHubAPI()

	repo, err := git.ParseGitHubURL(repoURL)
	if err != nil {
		return "", err
	}

	files, err := api.ListFiles(repo.Owner, repo.Name, ref)
	if err != nil {
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// GitHubGetFileContent returns file content from GitHub repository via API
func (a *App) GitHubGetFileContent(repoURL, filePath, ref string) (string, error) {
	api := git.NewGitHubAPI()

	repo, err := git.ParseGitHubURL(repoURL)
	if err != nil {
		return "", err
	}

	// Use raw content for better performance
	content, err := api.GetRawFileContent(repo.Owner, repo.Name, filePath, ref)
	if err != nil {
		return "", err
	}

	return content, nil
}

// GitHubBuildContext builds context from GitHub files via API (no clone)
func (a *App) GitHubBuildContext(repoURL string, files []string, ref string) (string, error) {
	api := git.NewGitHubAPI()

	repo, err := git.ParseGitHubURL(repoURL)
	if err != nil {
		return "", err
	}

	var contents []string
	for _, file := range files {
		content, err := api.GetRawFileContent(repo.Owner, repo.Name, file, ref)
		if err != nil {
			a.log.Info(fmt.Sprintf("Warning: Failed to read GitHub file %s: %v", file, err))
			continue
		}
		contents = append(contents, fmt.Sprintf("// File: %s (GitHub: %s/%s@%s)\n%s", file, repo.Owner, repo.Name, ref, content))
	}

	result := strings.Join(contents, "\n\n---\n\n")
	return result, nil
}

// GitHubGetDefaultBranch returns the default branch for a GitHub repository
func (a *App) GitHubGetDefaultBranch(repoURL string) (string, error) {
	api := git.NewGitHubAPI()

	repo, err := git.ParseGitHubURL(repoURL)
	if err != nil {
		return "", err
	}

	return api.GetDefaultBranch(repo.Owner, repo.Name)
}

// IsGitHubURL checks if URL is a GitHub repository
func (a *App) IsGitHubURL(url string) bool {
	return git.IsGitHubURL(url)
}

// ============ GITLAB API ENDPOINTS ============

// IsGitLabURL checks if URL is a GitLab repository
func (a *App) IsGitLabURL(url string) bool {
	return git.IsGitLabURL(url)
}

// GitLabGetBranches returns branches for a GitLab repository
func (a *App) GitLabGetBranches(repoURL string) (string, error) {
	repo, err := git.ParseGitLabURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse GitLab URL: %w", err)
	}

	api := git.NewGitLabAPI()
	branches, err := api.GetBranches(repo.Host, repo.Namespace, repo.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get branches: %w", err)
	}

	branchesJson, err := json.Marshal(branches)
	if err != nil {
		return "", fmt.Errorf("failed to marshal branches: %w", err)
	}

	return string(branchesJson), nil
}

// GitLabGetDefaultBranch returns the default branch for a GitLab repository
func (a *App) GitLabGetDefaultBranch(repoURL string) (string, error) {
	repo, err := git.ParseGitLabURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse GitLab URL: %w", err)
	}

	api := git.NewGitLabAPI()
	return api.GetDefaultBranch(repo.Host, repo.Namespace, repo.Name)
}

// GitLabGetCommits returns commits for a GitLab repository branch
func (a *App) GitLabGetCommits(repoURL, branch string, limit int) (string, error) {
	repo, err := git.ParseGitLabURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse GitLab URL: %w", err)
	}

	api := git.NewGitLabAPI()
	commits, err := api.GetCommits(repo.Host, repo.Namespace, repo.Name, branch, limit)
	if err != nil {
		return "", fmt.Errorf("failed to get commits: %w", err)
	}

	commitsJson, err := json.Marshal(commits)
	if err != nil {
		return "", fmt.Errorf("failed to marshal commits: %w", err)
	}

	return string(commitsJson), nil
}

// GitLabListFiles returns list of files in a GitLab repository at a specific ref
func (a *App) GitLabListFiles(repoURL, ref string) (string, error) {
	repo, err := git.ParseGitLabURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse GitLab URL: %w", err)
	}

	api := git.NewGitLabAPI()
	files, err := api.ListFiles(repo.Host, repo.Namespace, repo.Name, ref)
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	filesJson, err := json.Marshal(files)
	if err != nil {
		return "", fmt.Errorf("failed to marshal files: %w", err)
	}

	return string(filesJson), nil
}

// GitLabGetFileContent returns content of a file from GitLab repository
func (a *App) GitLabGetFileContent(repoURL, filePath, ref string) (string, error) {
	repo, err := git.ParseGitLabURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse GitLab URL: %w", err)
	}

	api := git.NewGitLabAPI()
	return api.GetFileContent(repo.Host, repo.Namespace, repo.Name, filePath, ref)
}

// GitLabBuildContext builds context from GitLab repository files
func (a *App) GitLabBuildContext(repoURL string, files []string, ref string) (string, error) {
	repo, err := git.ParseGitLabURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse GitLab URL: %w", err)
	}

	api := git.NewGitLabAPI()
	var contextBuilder strings.Builder

	for _, filePath := range files {
		content, err := api.GetFileContent(repo.Host, repo.Namespace, repo.Name, filePath, ref)
		if err != nil {
			a.log.Warning(fmt.Sprintf("Failed to get file %s: %v", filePath, err))
			continue
		}

		contextBuilder.WriteString(fmt.Sprintf("// File: %s\n", filePath))
		contextBuilder.WriteString(content)
		contextBuilder.WriteString("\n\n")
	}

	return contextBuilder.String(), nil
}

// GenerateReport generates a new report
func (a *App) GenerateReport(reportType string, parametersJson string) (string, error) {
	var parameters map[string]interface{}
	if parametersJson != "" {
		if err := json.Unmarshal([]byte(parametersJson), &parameters); err != nil {
			return "", fmt.Errorf("failed to parse parameters JSON: %w", err)
		}
	}

	report, err := a.reportService.GenerateReport(a.ctx, reportType, parameters)
	if err != nil {
		return "", fmt.Errorf("failed to generate report: %w", err)
	}

	reportJson, err := json.Marshal(report)
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	return string(reportJson), nil
}

// DeleteReport deletes a report by ID
func (a *App) DeleteReport(reportId string) error {
	return a.reportService.DeleteReport(a.ctx, reportId)
}

// ExportReport exports a report in the specified format
func (a *App) ExportReport(reportId string, format string) (string, error) {
	result, err := a.reportService.ExportReport(a.ctx, reportId, format)
	if err != nil {
		return "", fmt.Errorf("failed to export report: %w", err)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal export result: %w", err)
	}

	return string(resultJson), nil
}

// ListAutonomousTasks lists all autonomous tasks
func (a *App) ListAutonomousTasks(projectPath string) (string, error) {
	tasks, err := a.taskflowService.ListAutonomousTasks(a.ctx, projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to list autonomous tasks: %w", err)
	}

	tasksJson, err := json.Marshal(tasks)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tasks: %w", err)
	}

	return string(tasksJson), nil
}

// GetTaskLogs returns logs for a specific task
func (a *App) GetTaskLogs(taskId string) (string, error) {
	logs, err := a.taskflowService.GetTaskLogs(a.ctx, taskId)
	if err != nil {
		return "", fmt.Errorf("failed to get task logs: %w", err)
	}

	logsJson, err := json.Marshal(logs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal logs: %w", err)
	}

	return string(logsJson), nil
}

// PauseTask pauses an autonomous task
func (a *App) PauseTask(taskId string) error {
	return a.taskflowService.PauseTask(a.ctx, taskId)
}

// ResumeTask resumes a paused autonomous task
func (a *App) ResumeTask(taskId string) error {
	return a.taskflowService.ResumeTask(a.ctx, taskId)
}

// ExportProject exports an entire project
func (a *App) ExportProject(projectPath string, format string, optionsJson string) (string, error) {
	var options map[string]interface{}
	if optionsJson != "" {
		if err := json.Unmarshal([]byte(optionsJson), &options); err != nil {
			return "", fmt.Errorf("failed to parse options JSON: %w", err)
		}
	}

	// Create export settings for project export
	exportSettings := domain.ExportSettings{
		ProjectPath: projectPath,
		Format:      format,
		Options:     options,
	}

	result, err := a.exportService.Export(a.ctx, exportSettings)
	if err != nil {
		return "", fmt.Errorf("failed to export project: %w", err)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal export result: %w", err)
	}

	return string(resultJson), nil
}

// GetExportHistory returns export history
func (a *App) GetExportHistory(projectPath string) (string, error) {
	history, err := a.exportService.GetExportHistory(a.ctx, projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to get export history: %w", err)
	}

	historyJson, err := json.Marshal(history)
	if err != nil {
		return "", fmt.Errorf("failed to marshal export history: %w", err)
	}

	return string(historyJson), nil
}

// ============ TASK PROTOCOL API ENDPOINTS ============

// ExecuteTaskProtocol executes the full Task Protocol verification for a project
func (a *App) ExecuteTaskProtocol(configJson string) (string, error) {
	var config domain.TaskProtocolConfig
	if err := json.Unmarshal([]byte(configJson), &config); err != nil {
		return "", fmt.Errorf("failed to parse protocol config JSON: %w", err)
	}

	result, err := a.taskProtocolService.ExecuteProtocol(a.ctx, &config)
	if err != nil {
		return "", fmt.Errorf("protocol execution failed: %w", err)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal protocol result: %w", err)
	}

	return string(resultJson), nil
}

// ValidateTaskProtocolStage executes a single Task Protocol verification stage
func (a *App) ValidateTaskProtocolStage(stage string, configJson string) (string, error) {
	var config domain.TaskProtocolConfig
	if err := json.Unmarshal([]byte(configJson), &config); err != nil {
		return "", fmt.Errorf("failed to parse protocol config JSON: %w", err)
	}

	// Parse stage
	var protocolStage domain.ProtocolStage
	switch stage {
	case "linting":
		protocolStage = domain.StageLinting
	case "building":
		protocolStage = domain.StageBuilding
	case "testing":
		protocolStage = domain.StageTesting
	case "guardrails":
		protocolStage = domain.StageGuardrails
	default:
		return "", fmt.Errorf("unsupported protocol stage: %s", stage)
	}

	result, err := a.taskProtocolService.ValidateStage(a.ctx, protocolStage, &config)
	if err != nil {
		return "", fmt.Errorf("stage validation failed: %w", err)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal stage result: %w", err)
	}

	return string(resultJson), nil
}

// ValidateAIGeneratedCode validates AI-generated code using Task Protocol
func (a *App) ValidateAIGeneratedCode(requestJson string) (string, error) {
	var request struct {
		ProjectPath    string   `json:"projectPath"`
		Context        string   `json:"context"`
		Languages      []string `json:"languages"`
		GeneratedFiles []struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		} `json:"generatedFiles"`
	}

	if err := json.Unmarshal([]byte(requestJson), &request); err != nil {
		return "", fmt.Errorf("failed to parse validation request JSON: %w", err)
	}

	// Convert to internal request format
	validationRequest := &application.AICodeValidationRequest{
		ProjectPath: request.ProjectPath,
		Context:     request.Context,
		Languages:   request.Languages,
	}

	result, err := a.taskflowProtocolIntegration.ValidateAIGeneratedCode(a.ctx, validationRequest)
	if err != nil {
		return "", fmt.Errorf("AI code validation failed: %w", err)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal validation result: %w", err)
	}

	return string(resultJson), nil
}

// GetTaskProtocolConfiguration loads the current Task Protocol configuration
func (a *App) GetTaskProtocolConfiguration(projectPath string, languages []string) (string, error) {
	// Create default configuration for the project
	config := &domain.TaskProtocolConfig{
		ProjectPath: projectPath,
		Languages:   languages,
		EnabledStages: []domain.ProtocolStage{
			domain.StageLinting,
			domain.StageBuilding,
			domain.StageTesting,
			domain.StageGuardrails,
		},
		MaxRetries: 3,
		FailFast:   false,
		SelfCorrection: domain.SelfCorrectionConfig{
			Enabled:      true,
			MaxAttempts:  5,
			AIAssistance: true,
		},
		Timeouts: map[string]time.Duration{
			"linting":    5 * time.Minute,
			"building":   10 * time.Minute,
			"testing":    15 * time.Minute,
			"guardrails": 2 * time.Minute,
		},
	}

	// Try to load project-specific configuration
	configPath := filepath.Join(projectPath, "task_protocol.yaml")
	if loadedConfig, err := a.taskProtocolConfigService.LoadConfiguration(configPath); err == nil {
		config = loadedConfig
	}

	configJson, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal protocol configuration: %w", err)
	}

	return string(configJson), nil
}

// UpdateTaskProtocolConfiguration saves an updated Task Protocol configuration
func (a *App) UpdateTaskProtocolConfiguration(configJson string) error {
	var config domain.TaskProtocolConfig
	if err := json.Unmarshal([]byte(configJson), &config); err != nil {
		return fmt.Errorf("failed to parse protocol config JSON: %w", err)
	}

	// Validate the configuration
	if err := a.taskProtocolConfigService.ValidateConfiguration(&config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Save to project-specific location
	configPath := filepath.Join(config.ProjectPath, "task_protocol.yaml")
	if err := a.taskProtocolConfigService.SaveConfiguration(&config, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	a.log.Info(fmt.Sprintf("Task Protocol configuration updated for project: %s", config.ProjectPath))
	return nil
}

// RequestTaskProtocolCorrectionGuidance requests AI guidance for correcting errors
func (a *App) RequestTaskProtocolCorrectionGuidance(errorDetailsJson string, contextJson string) (string, error) {
	var errorDetails domain.ErrorDetails
	if err := json.Unmarshal([]byte(errorDetailsJson), &errorDetails); err != nil {
		return "", fmt.Errorf("failed to parse error details JSON: %w", err)
	}

	var taskContext domain.TaskContext
	if err := json.Unmarshal([]byte(contextJson), &taskContext); err != nil {
		return "", fmt.Errorf("failed to parse task context JSON: %w", err)
	}

	guidance, err := a.taskProtocolService.RequestCorrectionGuidance(a.ctx, &errorDetails, &taskContext)
	if err != nil {
		return "", fmt.Errorf("correction guidance request failed: %w", err)
	}

	guidanceJson, err := json.Marshal(guidance)
	if err != nil {
		return "", fmt.Errorf("failed to marshal correction guidance: %w", err)
	}

	return string(guidanceJson), nil
}

// CreateTaskProtocolForProject creates a default Task Protocol configuration for a project
func (a *App) CreateTaskProtocolForProject(projectPath string, languages []string) (string, error) {
	// Auto-detect languages if not provided
	if len(languages) == 0 {
		detectedLanguages, err := a.buildService.DetectLanguages(a.ctx, projectPath)
		if err != nil {
			a.log.Warning(fmt.Sprintf("Failed to detect languages for %s: %v", projectPath, err))
			languages = []string{"go"} // Default fallback
		} else {
			languages = detectedLanguages
		}
	}

	// Create configuration
	config := &domain.TaskProtocolConfig{
		ProjectPath: projectPath,
		Languages:   languages,
		EnabledStages: []domain.ProtocolStage{
			domain.StageLinting,
			domain.StageBuilding,
			domain.StageTesting,
			domain.StageGuardrails,
		},
		MaxRetries: 3,
		FailFast:   false,
		SelfCorrection: domain.SelfCorrectionConfig{
			Enabled:      true,
			MaxAttempts:  5,
			AIAssistance: true,
		},
		Timeouts: map[string]time.Duration{
			"linting":    5 * time.Minute,
			"building":   10 * time.Minute,
			"testing":    15 * time.Minute,
			"guardrails": 2 * time.Minute,
		},
	}

	// Save configuration to the project
	configPath := filepath.Join(projectPath, "task_protocol.yaml")
	if err := a.taskProtocolConfigService.SaveConfiguration(config, configPath); err != nil {
		return "", fmt.Errorf("failed to save protocol configuration: %w", err)
	}

	configJson, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal protocol configuration: %w", err)
	}

	a.log.Info(fmt.Sprintf("Task Protocol configuration created for project: %s", projectPath))
	return string(configJson), nil
}

// GetRecentProjects returns the list of recently opened projects
func (a *App) GetRecentProjects() (string, error) {
	projects := a.settingsService.GetRecentProjects()
	result, err := json.Marshal(projects)
	if err != nil {
		return "", fmt.Errorf("failed to marshal recent projects: %w", err)
	}
	return string(result), nil
}

// AddRecentProject adds a project to the recent list and saves settings
func (a *App) AddRecentProject(path, name string) error {
	a.settingsService.AddRecentProject(path, name)
	return a.settingsService.Save()
}

// RemoveRecentProject removes a project from the recent list and saves settings
func (a *App) RemoveRecentProject(path string) error {
	a.settingsService.RemoveRecentProject(path)
	return a.settingsService.Save()
}

// ============ IGNORE RULES API ENDPOINTS ============

// GetGitignoreContentForProject returns .gitignore content for a project
func (a *App) GetGitignoreContentForProject(projectPath string) (string, error) {
	content, err := a.gitRepo.GetGitignoreContent(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to get .gitignore content: %w", err)
	}
	return content, nil
}

// GetCustomIgnoreRules returns custom ignore rules from settings
func (a *App) GetCustomIgnoreRules() (string, error) {
	dto, err := a.settingsService.GetSettingsDTO()
	if err != nil {
		return "", fmt.Errorf("failed to get settings: %w", err)
	}
	return dto.CustomIgnoreRules, nil
}

// UpdateCustomIgnoreRules updates custom ignore rules in settings
func (a *App) UpdateCustomIgnoreRules(rules string) error {
	dto, err := a.settingsService.GetSettingsDTO()
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}
	dto.CustomIgnoreRules = rules
	return a.settingsService.SaveSettingsDTO(dto)
}

// TestIgnoreRules tests ignore rules against project files
func (a *App) TestIgnoreRules(projectPath string, rules string) ([]string, error) {
	// Get all files in project
	files, err := a.projectService.ListFiles(projectPath, false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Test which files would be ignored
	var ignoredFiles []string
	for _, file := range files {
		if !file.IsDir && a.matchesIgnorePattern(file.RelPath, rules) {
			ignoredFiles = append(ignoredFiles, file.RelPath)
		}
	}

	return ignoredFiles, nil
}

// matchesIgnorePattern checks if a path matches any ignore pattern
func (a *App) matchesIgnorePattern(path string, rules string) bool {
	if rules == "" {
		return false
	}

	lines := strings.Split(rules, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Simple pattern matching (can be enhanced with proper gitignore library)
		if strings.Contains(path, line) || strings.HasSuffix(path, line) {
			return true
		}

		// Handle wildcards
		if strings.Contains(line, "*") {
			pattern := strings.ReplaceAll(line, "*", ".*")
			if matched, _ := filepath.Match(pattern, path); matched {
				return true
			}
		}
	}

	return false
}

// AddToGitignore adds a pattern to .gitignore file
func (a *App) AddToGitignore(projectPath string, pattern string) error {
	gitignorePath := filepath.Join(projectPath, ".gitignore")
	
	// Read existing content
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	// Append new pattern
	newContent := string(content)
	if !strings.HasSuffix(newContent, "\n") && newContent != "" {
		newContent += "\n"
	}
	newContent += pattern + "\n"

	// Write back
	if err := os.WriteFile(gitignorePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}


// ==================== Qwen Task Methods ====================

// QwenExecuteTask executes a task using Qwen with smart context collection
func (a *App) QwenExecuteTask(requestJson string) (string, error) {
	var req handlers.ExecuteTaskRequest
	if err := json.Unmarshal([]byte(requestJson), &req); err != nil {
		return "", fmt.Errorf("failed to parse request: %w", err)
	}

	result := a.qwenHandler.ExecuteTask(req)

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJson), nil
}

// QwenPreviewContext returns a preview of the context that would be collected
func (a *App) QwenPreviewContext(requestJson string) (string, error) {
	var req handlers.ExecuteTaskRequest
	if err := json.Unmarshal([]byte(requestJson), &req); err != nil {
		return "", fmt.Errorf("failed to parse request: %w", err)
	}

	result := a.qwenHandler.PreviewContext(req)

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJson), nil
}

// QwenGetAvailableModels returns available Qwen models
func (a *App) QwenGetAvailableModels() (string, error) {
	models := a.qwenHandler.GetAvailableModels()

	resultJson, err := json.Marshal(models)
	if err != nil {
		return "", fmt.Errorf("failed to marshal models: %w", err)
	}

	return string(resultJson), nil
}
