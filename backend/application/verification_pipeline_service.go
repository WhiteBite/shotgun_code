package application

import (
	"context"
	"fmt"
	"path/filepath"
	"shotgun_code/domain"
	"time"
)

// VerificationPipelineService предоставляет высокоуровневый API для verification pipeline
type VerificationPipelineService struct {
	log              domain.Logger
	buildService     domain.IBuildService
	testService      domain.ITestService
	staticAnalyzer   domain.IStaticAnalyzerService
	formatterService *FormatterService
	reportWriter     domain.FileSystemWriter
	taskProtocol     domain.TaskProtocolService // NEW: Task Protocol integration
}

// NewVerificationPipelineService создает новый сервис verification pipeline
func NewVerificationPipelineService(
	log domain.Logger,
	buildService domain.IBuildService,
	testService domain.ITestService,
	staticAnalyzer domain.IStaticAnalyzerService,
	formatterService *FormatterService,
	reportWriter domain.FileSystemWriter,
	taskProtocol domain.TaskProtocolService, // NEW: Task Protocol parameter
) *VerificationPipelineService {
	return &VerificationPipelineService{
		log:              log,
		buildService:     buildService,
		testService:      testService,
		staticAnalyzer:   staticAnalyzer,
		formatterService: formatterService,
		reportWriter:     reportWriter,
		taskProtocol:     taskProtocol, // NEW: Initialize task protocol
	}
}

// RunVerificationPipeline выполняет полный verification pipeline
func (s *VerificationPipelineService) RunVerificationPipeline(ctx context.Context, config *domain.VerificationConfig) (*domain.VerificationResult, error) {
	s.log.Info(fmt.Sprintf("Starting verification pipeline for project: %s", config.ProjectPath))

	result := &domain.VerificationResult{
		ProjectPath: config.ProjectPath,
		Languages:   config.Languages,
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
		Steps:       make([]*domain.VerificationStep, 0),
	}

	// Шаг 1: Форматирование
	formatStep := &domain.VerificationStep{
		Name:      "format",
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}

	formatResult, err := s.runFormatStep(ctx, config)
	formatStep.Result = formatResult
	formatStep.Success = err == nil
	formatStep.Error = err
	formatStep.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	result.Steps = append(result.Steps, formatStep)

	if err != nil {
		s.log.Warning(fmt.Sprintf("Format step failed: %v", err))
		// Продолжаем выполнение, но логируем ошибку
	}

	// Шаг 2: Build и Type Check
	buildStep := &domain.VerificationStep{
		Name:      "build-typecheck",
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}

	buildResult, err := s.runBuildTypeCheckStep(ctx, config)
	buildStep.Result = buildResult
	buildStep.Success = err == nil
	buildStep.Error = err
	buildStep.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	result.Steps = append(result.Steps, buildStep)

	if err != nil {
		s.log.Error(fmt.Sprintf("Build/TypeCheck step failed: %v", err))
		result.Success = false
		result.CompletedAt = time.Now().UTC().Format(time.RFC3339)
		return result, fmt.Errorf("build/typecheck failed: %w", err)
	}

	// Шаг 3: Smoke Tests
	testStep := &domain.VerificationStep{
		Name:      "smoke-tests",
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}

	testResult, err := s.runSmokeTestsStep(ctx, config)
	testStep.Result = testResult
	testStep.Success = err == nil
	testStep.Error = err
	testStep.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	result.Steps = append(result.Steps, testStep)

	if err != nil {
		s.log.Error(fmt.Sprintf("Smoke tests step failed: %v", err))
		result.Success = false
		result.CompletedAt = time.Now().UTC().Format(time.RFC3339)
		return result, fmt.Errorf("smoke tests failed: %w", err)
	}

	// Шаг 4: Static Analysis
	staticStep := &domain.VerificationStep{
		Name:      "static-analysis",
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}

	staticResult, err := s.runStaticAnalysisStep(ctx, config)
	staticStep.Result = staticResult
	staticStep.Success = err == nil
	staticStep.Error = err
	staticStep.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	result.Steps = append(result.Steps, staticStep)

	if err != nil {
		s.log.Warning(fmt.Sprintf("Static analysis step failed: %v", err))
		// Не прерываем выполнение, но логируем ошибку
	}

	// Определяем общий успех
	result.Success = true
	for _, step := range result.Steps {
		if !step.Success && step.Name != "format" && step.Name != "static-analysis" {
			result.Success = false
			break
		}
	}

	result.CompletedAt = time.Now().UTC().Format(time.RFC3339)

	// Сохраняем отчет
	if err := s.saveVerificationReport(ctx, result); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to save verification report: %v", err))
	}

	s.log.Info(fmt.Sprintf("Verification pipeline completed with success: %t", result.Success))
	return result, nil
}

// runFormatStep выполняет шаг форматирования
func (s *VerificationPipelineService) runFormatStep(ctx context.Context, config *domain.VerificationConfig) (interface{}, error) {
	s.log.Info("Running format step")

	results := make([]*domain.FormatResult, 0, len(config.Languages))

	for _, language := range config.Languages {
		result, err := s.formatterService.FormatProject(ctx, config.ProjectPath, language)
		if err != nil {
			s.log.Warning(fmt.Sprintf("Failed to format %s: %v", language, err))
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// runBuildTypeCheckStep выполняет шаг сборки и проверки типов
func (s *VerificationPipelineService) runBuildTypeCheckStep(ctx context.Context, config *domain.VerificationConfig) (interface{}, error) {
	s.log.Info("Running build/typecheck step")

	// Выполняем валидацию проекта
	validation, err := s.buildService.ValidateProject(ctx, config.ProjectPath, config.Languages)
	if err != nil {
		return nil, fmt.Errorf("project validation failed: %w", err)
	}

	if !validation.Success {
		return nil, fmt.Errorf("project validation failed for some languages")
	}

	return validation, nil
}

// runSmokeTestsStep выполняет шаг smoke тестов
func (s *VerificationPipelineService) runSmokeTestsStep(ctx context.Context, config *domain.VerificationConfig) (interface{}, error) {
	s.log.Info("Running smoke tests step")

	var allResults []*domain.TestResult

	for _, language := range config.Languages {
		results, err := s.testService.RunSmokeTests(ctx, config.ProjectPath, language)
		if err != nil {
			return nil, fmt.Errorf("smoke tests failed for %s: %w", language, err)
		}

		// Валидируем результаты тестов
		validation := s.testService.ValidateTestResults(results)
		if !validation.Success {
			return nil, fmt.Errorf("smoke tests failed for %s: %d tests failed", language, validation.FailedTests)
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// runStaticAnalysisStep выполняет шаг статического анализа
func (s *VerificationPipelineService) runStaticAnalysisStep(ctx context.Context, config *domain.VerificationConfig) (interface{}, error) {
	s.log.Info("Running static analysis step")

	report, err := s.staticAnalyzer.AnalyzeProject(ctx, config.ProjectPath, config.Languages)
	if err != nil {
		return nil, fmt.Errorf("static analysis failed: %w", err)
	}

	return report, nil
}

// saveVerificationReport сохраняет отчет о verification pipeline
func (s *VerificationPipelineService) saveVerificationReport(ctx context.Context, result *domain.VerificationResult) error {
	// Создаем директорию tasks/reports если не существует
	reportsDir := "tasks/reports"
	if err := s.reportWriter.MkdirAll(reportsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Генерируем имя файла с временной меткой
	timestamp := time.Now().UTC().Format("20060102_150405")
	filename := fmt.Sprintf("verification_%s.json", timestamp)
	filepath := filepath.Join(reportsDir, filename)

	// Сериализуем результат в JSON
	jsonData, err := s.serializeVerificationResult(result)
	if err != nil {
		return fmt.Errorf("failed to serialize verification result: %w", err)
	}

	// Сохраняем файл
	if err := s.reportWriter.WriteFile(filepath, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write verification report: %w", err)
	}

	s.log.Info(fmt.Sprintf("Saved verification report to %s", filepath))
	return nil
}

// serializeVerificationResult сериализует результат verification pipeline
func (s *VerificationPipelineService) serializeVerificationResult(result *domain.VerificationResult) ([]byte, error) {
	// Простая сериализация - в реальной реализации нужно использовать encoding/json
	// Здесь возвращаем простую структуру для демонстрации
	content := fmt.Sprintf(`{
  "projectPath": "%s",
  "languages": %v,
  "success": %t,
  "startedAt": "%s",
  "completedAt": "%s",
  "steps": %d
}`, result.ProjectPath, result.Languages, result.Success, result.StartedAt, result.CompletedAt, len(result.Steps))

	return []byte(content), nil
}

// GetSupportedLanguages возвращает поддерживаемые языки
func (s *VerificationPipelineService) GetSupportedLanguages() []string {
	return s.buildService.GetSupportedLanguages()
}

// DetectLanguages определяет языки в проекте
func (s *VerificationPipelineService) DetectLanguages(ctx context.Context, projectPath string) ([]string, error) {
	return s.buildService.DetectLanguages(ctx, projectPath)
}

// RunTaskProtocolVerification executes the Task Protocol verification for a task
func (s *VerificationPipelineService) RunTaskProtocolVerification(ctx context.Context, config *domain.TaskProtocolConfig) (*domain.TaskProtocolResult, error) {
	s.log.Info(fmt.Sprintf("Starting Task Protocol verification for project: %s", config.ProjectPath))

	if s.taskProtocol == nil {
		return nil, fmt.Errorf("task protocol service not available")
	}

	// Execute the task protocol verification
	result, err := s.taskProtocol.ExecuteProtocol(ctx, config)
	if err != nil {
		s.log.Error(fmt.Sprintf("Task Protocol verification failed: %v", err))
		return result, err
	}

	// Save the protocol report
	if err := s.saveTaskProtocolReport(ctx, result); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to save task protocol report: %v", err))
	}

	s.log.Info(fmt.Sprintf("Task Protocol verification completed with success: %t", result.Success))
	return result, nil
}

// RunTaskProtocolStage executes a single Task Protocol verification stage
func (s *VerificationPipelineService) RunTaskProtocolStage(ctx context.Context, stage domain.ProtocolStage, config *domain.TaskProtocolConfig) (*domain.ProtocolStageResult, error) {
	s.log.Info(fmt.Sprintf("Running Task Protocol stage: %s", stage))

	if s.taskProtocol == nil {
		return nil, fmt.Errorf("task protocol service not available")
	}

	return s.taskProtocol.ValidateStage(ctx, stage, config)
}

// CreateTaskProtocolConfig creates a TaskProtocolConfig from VerificationConfig
func (s *VerificationPipelineService) CreateTaskProtocolConfig(verifyConfig *domain.VerificationConfig) *domain.TaskProtocolConfig {
	return &domain.TaskProtocolConfig{
		ProjectPath: verifyConfig.ProjectPath,
		Languages:   verifyConfig.Languages,
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
}

// saveTaskProtocolReport saves the task protocol report
func (s *VerificationPipelineService) saveTaskProtocolReport(ctx context.Context, result *domain.TaskProtocolResult) error {
	// Create directory for task protocol reports
	reportsDir := "tasks/reports/protocols"
	if err := s.reportWriter.MkdirAll(reportsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create protocol reports directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().UTC().Format("20060102_150405")
	filename := fmt.Sprintf("task_protocol_%s_%s.json", result.TaskID, timestamp)
	filepath := filepath.Join(reportsDir, filename)

	// Serialize result to JSON
	jsonData, err := s.serializeTaskProtocolResult(result)
	if err != nil {
		return fmt.Errorf("failed to serialize task protocol result: %w", err)
	}

	// Save file
	if err := s.reportWriter.WriteFile(filepath, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write task protocol report: %w", err)
	}

	s.log.Info(fmt.Sprintf("Saved task protocol report to %s", filepath))
	return nil
}

// serializeTaskProtocolResult serializes the task protocol result
func (s *VerificationPipelineService) serializeTaskProtocolResult(result *domain.TaskProtocolResult) ([]byte, error) {
	// Simplified serialization - in real implementation, use encoding/json
	content := fmt.Sprintf(`{
  "taskId": "%s",
  "success": %t,
  "startedAt": "%s",
  "completedAt": "%s",
  "correctionCycles": %d,
  "stages": %d,
  "finalError": "%s"
}`,
		result.TaskID,
		result.Success,
		result.StartedAt.Format(time.RFC3339),
		result.CompletedAt.Format(time.RFC3339),
		result.CorrectionCycles,
		len(result.Stages),
		result.FinalError)

	return []byte(content), nil
}
