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
	"shotgun_code/infrastructure/wailsbridge"
	"strings"
	"time"
)

type App struct {
	ctx                   context.Context
	projectService        *application.ProjectService
	aiService             *application.AIService
	settingsService       *application.SettingsService
	contextAnalysis       domain.ContextAnalyzer
	symbolGraph           *application.SymbolGraphService
	testService           *application.TestService
	staticAnalyzerService *application.StaticAnalyzerService
	sbomService           *application.SBOMService
	repairService         domain.RepairService
	guardrailService      domain.GuardrailService
	taskflowService       domain.TaskflowService
	uxMetricsService      domain.UXMetricsService
	applyService          *application.ApplyService
	diffService           *application.DiffService
	buildService          *application.BuildService
	fileWatcher           domain.FileSystemWatcher
	gitRepo               domain.GitRepository
	exportService         *application.ExportService
	fileReader            domain.FileContentReader
	contextService        *application.ContextService
	reportService         *application.ReportService
	bridge                *wailsbridge.Bridge
	log                   domain.Logger
	// Task Protocol Services
	taskProtocolService         domain.TaskProtocolService
	taskProtocolConfigService   *application.TaskProtocolConfigService
	taskflowProtocolIntegration *application.TaskflowProtocolIntegration
}

func (a *App) startup(ctx context.Context, container *app.AppContainer) {
	a.ctx = ctx
	a.projectService = container.ProjectService
	a.aiService = container.AIService
	a.settingsService = container.SettingsService
	a.contextAnalysis = container.ContextAnalysis
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
	a.contextService = container.ContextService
	a.reportService = container.ReportService
	a.bridge = container.Bridge
	a.log = container.Log
	// Task Protocol Services
	a.taskProtocolService = container.TaskProtocolService
	a.taskProtocolConfigService = container.TaskProtocolConfigService
	a.taskflowProtocolIntegration = container.TaskflowProtocolIntegration
}

func (a *App) domReady(ctx context.Context) {
	a.ctx = ctx
	a.bridge.SetWailsContext(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.ctx = ctx
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
	return a.projectService.ListFiles(dirPath, useGitignore, useCustomIgnore)
}

func (a *App) RequestShotgunContextGeneration(rootDir string, includedPaths []string) {
	a.projectService.GenerateContext(a.ctx, rootDir, includedPaths)
}

func (a *App) GenerateCode(systemPrompt, userPrompt string) (string, error) {
	return a.aiService.GenerateCode(a.ctx, systemPrompt, userPrompt)
}

// BuildSymbolGraph строит граф символов для проекта
func (a *App) BuildSymbolGraph(projectRoot, language string) (*domain.SymbolGraph, error) {
	return a.symbolGraph.BuildSymbolGraph(a.ctx, projectRoot, language)
}

// GetSymbolSuggestions возвращает предложения символов
func (a *App) GetSymbolSuggestions(query, language string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	return a.symbolGraph.GetSuggestions(a.ctx, query, language, graph)
}

// GetSymbolDependencies возвращает зависимости символа
func (a *App) GetSymbolDependencies(symbolID, language string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	return a.symbolGraph.GetDependencies(a.ctx, symbolID, language, graph)
}

// GetSymbolDependents возвращает символы, зависящие от указанного
func (a *App) GetSymbolDependents(symbolID, language string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	return a.symbolGraph.GetDependents(a.ctx, symbolID, language, graph)
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
	return a.buildService.Build(a.ctx, projectPath, language)
}

// TypeCheck выполняет проверку типов
func (a *App) TypeCheck(projectPath, language string) (*domain.TypeCheckResult, error) {
	return a.buildService.TypeCheck(a.ctx, projectPath, language)
}

// BuildAndTypeCheck выполняет сборку и проверку типов
func (a *App) BuildAndTypeCheck(projectPath, language string) (*domain.BuildResult, *domain.TypeCheckResult, error) {
	return a.buildService.BuildAndTypeCheck(a.ctx, projectPath, language)
}

// ValidateProject выполняет полную валидацию проекта
func (a *App) ValidateProject(projectPath string, languages []string) (*domain.ProjectValidationResult, error) {
	return a.buildService.ValidateProject(a.ctx, projectPath, languages)
}

// DetectLanguages определяет языки в проекте
func (a *App) DetectLanguages(projectPath string) ([]string, error) {
	return a.buildService.DetectLanguages(a.ctx, projectPath)
}

// RunTests выполняет тесты согласно конфигурации
func (a *App) RunTests(config *domain.TestConfig) ([]*domain.TestResult, error) {
	return a.testService.RunTests(a.ctx, config)
}

// RunTargetedTests выполняет целевые тесты для затронутых файлов
func (a *App) RunTargetedTests(config *domain.TestConfig, changedFiles []string) ([]*domain.TestResult, error) {
	return a.testService.RunTargetedTests(a.ctx, config, changedFiles)
}

// DiscoverTests обнаруживает тесты в проекте
func (a *App) DiscoverTests(projectPath, language string) (*domain.TestSuite, error) {
	return a.testService.DiscoverTests(a.ctx, projectPath, language)
}

// BuildAffectedGraph строит граф затронутых файлов
func (a *App) BuildAffectedGraph(changedFiles []string, projectPath string) (*domain.AffectedGraph, error) {
	return a.testService.BuildAffectedGraph(a.ctx, changedFiles, projectPath)
}

// RunSmokeTests выполняет только smoke тесты
func (a *App) RunSmokeTests(projectPath, language string) ([]*domain.TestResult, error) {
	return a.testService.RunSmokeTests(a.ctx, projectPath, language)
}

// RunUnitTests выполняет только unit тесты
func (a *App) RunUnitTests(projectPath, language string) ([]*domain.TestResult, error) {
	return a.testService.RunUnitTests(a.ctx, projectPath, language)
}

// RunIntegrationTests выполняет только integration тесты
func (a *App) RunIntegrationTests(projectPath, language string) ([]*domain.TestResult, error) {
	return a.testService.RunIntegrationTests(a.ctx, projectPath, language)
}

// ValidateTestResults валидирует результаты тестов
func (a *App) ValidateTestResults(results []*domain.TestResult) *domain.TestValidationResult {
	return a.testService.ValidateTestResults(results)
}

// AnalyzeProject выполняет статический анализ проекта
func (a *App) AnalyzeProject(projectPath string, languages []string) (*domain.StaticAnalysisReport, error) {
	return a.staticAnalyzerService.AnalyzeProject(a.ctx, projectPath, languages)
}

// AnalyzeFile выполняет статический анализ одного файла
func (a *App) AnalyzeFile(filePath, language string) (*domain.StaticAnalysisResult, error) {
	return a.staticAnalyzerService.AnalyzeFile(a.ctx, filePath, language)
}

// AnalyzeGoProject выполняет статический анализ Go проекта
func (a *App) AnalyzeGoProject(projectPath string) (*domain.StaticAnalysisResult, error) {
	return a.staticAnalyzerService.AnalyzeGoProject(a.ctx, projectPath)
}

// AnalyzeTypeScriptProject выполняет статический анализ TypeScript проекта
func (a *App) AnalyzeTypeScriptProject(projectPath string) (*domain.StaticAnalysisResult, error) {
	return a.staticAnalyzerService.AnalyzeTypeScriptProject(a.ctx, projectPath)
}

// AnalyzeJavaScriptProject выполняет статический анализ JavaScript проекта
func (a *App) AnalyzeJavaScriptProject(projectPath string) (*domain.StaticAnalysisResult, error) {
	return a.staticAnalyzerService.AnalyzeJavaScriptProject(a.ctx, projectPath)
}

// GetSupportedAnalyzers возвращает поддерживаемые анализаторы
func (a *App) GetSupportedAnalyzers() []domain.StaticAnalyzerType {
	return a.staticAnalyzerService.GetSupportedAnalyzers()
}

// ValidateAnalysisResults валидирует результаты статического анализа
func (a *App) ValidateAnalysisResults(results map[string]*domain.StaticAnalysisResult) *domain.StaticAnalysisValidationResult {
	return a.staticAnalyzerService.ValidateAnalysisResults(results)
}

// GenerateSBOM генерирует SBOM для проекта
func (a *App) GenerateSBOM(projectPath string, format domain.SBOMFormat) (*domain.SBOMResult, error) {
	return a.sbomService.GenerateSBOM(a.ctx, projectPath, format)
}

// ScanVulnerabilities сканирует уязвимости в проекте
func (a *App) ScanVulnerabilities(projectPath string) (*domain.VulnerabilityScanResult, error) {
	return a.sbomService.ScanVulnerabilities(a.ctx, projectPath)
}

// ScanLicenses сканирует лицензии в проекте
func (a *App) ScanLicenses(projectPath string) (*domain.LicenseScanResult, error) {
	return a.sbomService.ScanLicenses(a.ctx, projectPath)
}

// GenerateComplianceReport генерирует отчет о соответствии
func (a *App) GenerateComplianceReport(projectPath string, requirements *domain.ComplianceRequirements) (*domain.ComplianceReport, error) {
	return a.sbomService.GenerateComplianceReport(a.ctx, projectPath, requirements)
}

// GetSupportedSBOMFormats возвращает поддерживаемые форматы SBOM
func (a *App) GetSupportedSBOMFormats() []domain.SBOMFormat {
	return a.sbomService.GetSupportedSBOMFormats()
}

// ValidateSBOM валидирует SBOM
func (a *App) ValidateSBOM(sbomPath string, format domain.SBOMFormat) error {
	return a.sbomService.ValidateSBOM(a.ctx, sbomPath, format)
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
	var options application.IntelligentGenerationOptions
	if err := json.Unmarshal([]byte(optionsJson), &options); err != nil {
		validationErr := domain.NewValidationError("failed to parse options JSON", map[string]interface{}{
			"originalError": err.Error(),
			"optionsJson":   optionsJson,
		})
		return "", a.transformError(validationErr)
	}

	result, err := a.aiService.GenerateIntelligentCode(a.ctx, task, context, options)
	if err != nil {
		return "", a.transformError(err)
	}

	// Возвращаем результат как JSON
	resultJson, err := json.Marshal(result)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal result", err)
		return "", a.transformError(marshalErr)
	}

	return string(resultJson), nil
}

// GenerateCodeWithOptions генерирует код с дополнительными опциями
func (a *App) GenerateCodeWithOptions(systemPrompt, userPrompt, optionsJson string) (string, error) {
	var options application.CodeGenerationOptions
	if err := json.Unmarshal([]byte(optionsJson), &options); err != nil {
		return "", fmt.Errorf("failed to parse options JSON: %w", err)
	}

	return a.aiService.GenerateCodeWithOptions(a.ctx, systemPrompt, userPrompt, options)
}

// GetProviderInfo возвращает информацию о текущем провайдере
func (a *App) GetProviderInfo() (string, error) {
	info, err := a.aiService.GetProviderInfo(a.ctx)
	if err != nil {
		return "", a.transformError(err)
	}

	infoJson, err := json.Marshal(info)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal provider info", err)
		return "", a.transformError(marshalErr)
	}

	return string(infoJson), nil
}

// ListAvailableModels возвращает список доступных моделей
func (a *App) ListAvailableModels() ([]string, error) {
	return a.aiService.ListAvailableModels(a.ctx)
}

func (a *App) SuggestContextFiles(task string, allFiles []*domain.FileNode) ([]string, error) {
	return a.contextAnalysis.SuggestFiles(a.ctx, task, allFiles)
}

// AnalyzeTaskAndCollectContext анализирует задачу и собирает релевантный контекст
func (a *App) AnalyzeTaskAndCollectContext(task string, allFilesJson string, rootDir string) (string, error) {
	var allFiles []*domain.FileNode
	if err := json.Unmarshal([]byte(allFilesJson), &allFiles); err != nil {
		return "", domain.NewValidationError("invalid allFiles JSON", map[string]interface{}{
			"error": err.Error(),
		})
	}

	result, err := a.contextService.AnalyzeTaskAndCollectContext(a.ctx, task, allFiles, rootDir)
	if err != nil {
		return "", a.transformError(err)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal analysis result", err)
		return "", a.transformError(marshalErr)
	}

	return string(resultJson), nil
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
	return a.settingsService.GetSettingsDTO()
}

func (a *App) SaveSettings(settingsJson string) error {
	var dto domain.SettingsDTO
	err := json.Unmarshal([]byte(settingsJson), &dto)
	if err != nil {
		return fmt.Errorf("failed to parse settings JSON: %w", err)
	}
	return a.settingsService.SaveSettingsDTO(dto)
}

func (a *App) RefreshAIModels(provider, apiKey string) error {
	return a.settingsService.RefreshModels(provider, apiKey)
}

func (a *App) StartFileWatcher(rootDirPath string) error {
	return a.fileWatcher.Start(rootDirPath)
}

func (a *App) StopFileWatcher() {
	a.fileWatcher.Stop()
}

func (a *App) IsGitAvailable() bool {
	return a.gitRepo.IsGitAvailable()
}

func (a *App) GetUncommittedFiles(projectRoot string) ([]domain.FileStatus, error) {
	return a.gitRepo.GetUncommittedFiles(projectRoot)
}

func (a *App) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	return a.gitRepo.GetRichCommitHistory(projectRoot, branchName, limit)
}

func (a *App) GetFileContentAtCommit(projectRoot, filePath, commitHash string) (string, error) {
	return a.gitRepo.GetFileContentAtCommit(projectRoot, filePath, commitHash)
}

func (a *App) GetGitignoreContent(projectRoot string) (string, error) {
	return a.gitRepo.GetGitignoreContent(projectRoot)
}

func (a *App) ReadFileContent(rootDir, relPath string) (string, error) {
	contents, err := a.fileReader.ReadContents(a.ctx, []string{relPath}, rootDir, nil)
	if err != nil {
		return "", a.transformError(err)
	}
	if content, ok := contents[relPath]; ok {
		return content, nil
	}
	fileErr := domain.NewValidationError("file not found or could not be read", map[string]interface{}{
		"rootDir": rootDir,
		"relPath": relPath,
	})
	return "", a.transformError(fileErr)
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
	var options domain.ContextBuildOptions
	if err := json.Unmarshal([]byte(optionsJson), &options); err != nil {
		validationErr := domain.NewValidationError("failed to parse options JSON", map[string]interface{}{
			"originalError": err.Error(),
			"optionsJson":   optionsJson,
		})
		return "", a.transformError(validationErr)
	}

	// Use new memory-safe BuildContext that returns ContextSummary
	contextSummary, err := a.contextService.BuildContext(a.ctx, projectPath, includedPaths, &options)
	if err != nil {
		return "", a.transformError(err)
	}

	contextJson, err := json.Marshal(contextSummary)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal context summary", err)
		return "", a.transformError(marshalErr)
	}

	return string(contextJson), nil
}

// GetContextContent returns paginated context content for memory-safe viewing
// This allows accessing context content without loading it all into memory
func (a *App) GetContextContent(contextID string, startLine int, lineCount int) (string, error) {
	chunk, err := a.contextService.GetContextContent(a.ctx, contextID, startLine, lineCount)
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

// Legacy BuildContextLegacy for backward compatibility - DEPRECATED
// This method should be avoided as it can cause OOM issues with large contexts
func (a *App) BuildContextLegacy(projectPath string, includedPaths []string, optionsJson string) (string, error) {
	var options domain.ContextBuildOptions
	if err := json.Unmarshal([]byte(optionsJson), &options); err != nil {
		validationErr := domain.NewValidationError("failed to parse options JSON", map[string]interface{}{
			"originalError": err.Error(),
			"optionsJson":   optionsJson,
		})
		return "", a.transformError(validationErr)
	}

	// Use legacy method that returns full context (can cause OOM)
	context, err := a.contextService.BuildContextLegacy(a.ctx, projectPath, includedPaths, options)
	if err != nil {
		return "", a.transformError(err)
	}

	contextJson, err := json.Marshal(context)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal context", err)
		return "", a.transformError(marshalErr)
	}

	return string(contextJson), nil
}

// GetContext получает контекст по ID
func (a *App) GetContext(contextID string) (string, error) {
	context, err := a.contextService.GetContext(a.ctx, contextID)
	if err != nil {
		return "", a.transformError(err)
	}

	contextJson, err := json.Marshal(context)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal context", err)
		return "", a.transformError(marshalErr)
	}

	return string(contextJson), nil
}

// GetProjectContexts получает все контексты для проекта
func (a *App) GetProjectContexts(projectPath string) (string, error) {
	contexts, err := a.contextService.GetProjectContexts(a.ctx, projectPath)
	if err != nil {
		return "", a.transformError(err)
	}

	contextsJson, err := json.Marshal(contexts)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal contexts", err)
		return "", a.transformError(marshalErr)
	}

	return string(contextsJson), nil
}

// DeleteContext удаляет контекст по ID
func (a *App) DeleteContext(contextID string) error {
	return a.contextService.DeleteContext(a.ctx, contextID)
}

// BuildContextFromRequest builds context with proper JSON handling and returns ContextSummary to prevent OOM
func (a *App) BuildContextFromRequest(projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextSummary, error) {
	return a.contextService.BuildContext(a.ctx, projectPath, includedPaths, options)
}

// GetContextLines retrieves a range of lines from a streaming context
func (a *App) GetContextLines(contextID string, startLine, endLine int64) (string, error) {
	lines, err := a.contextService.GetContextLines(a.ctx, contextID, startLine, endLine)
	if err != nil {
		return "", a.transformError(err)
	}

	linesJson, err := json.Marshal(lines)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal lines", err)
		return "", a.transformError(marshalErr)
	}

	return string(linesJson), nil
}

// CreateStreamingContext creates a streaming context from project files
func (a *App) CreateStreamingContext(projectPath string, includedPaths []string, optionsJson string) (string, error) {
	var options domain.ContextBuildOptions
	if err := json.Unmarshal([]byte(optionsJson), &options); err != nil {
		validationErr := domain.NewValidationError("failed to parse options JSON", map[string]interface{}{
			"originalError": err.Error(),
			"optionsJson":   optionsJson,
		})
		return "", a.transformError(validationErr)
	}

	stream, err := a.contextService.CreateStreamingContext(a.ctx, projectPath, includedPaths, &options)
	if err != nil {
		return "", a.transformError(err)
	}

	streamJson, err := json.Marshal(stream)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal stream", err)
		return "", a.transformError(marshalErr)
	}

	return string(streamJson), nil
}

// GetStreamingContext retrieves a streaming context by ID
func (a *App) GetStreamingContext(contextID string) (string, error) {
	stream, err := a.contextService.GetContextStream(a.ctx, contextID)
	if err != nil {
		return "", a.transformError(err)
	}

	streamJson, err := json.Marshal(stream)
	if err != nil {
		marshalErr := domain.NewInternalError("failed to marshal stream", err)
		return "", a.transformError(marshalErr)
	}

	return string(streamJson), nil
}

// CloseStreamingContext closes a streaming context and cleans up resources
func (a *App) CloseStreamingContext(contextID string) error {
	return a.contextService.CloseContextStream(a.ctx, contextID)
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
