package project

import (
	"context"
	"fmt"
	"path/filepath"
	"shotgun_code/application"
	"shotgun_code/domain"
	"time"
)

// VerificationPipelineService provides high-level API for verification pipeline
type VerificationPipelineService struct {
	log              domain.Logger
	buildService     *BuildService
	testService      *TestService
	staticAnalyzer   *application.StaticAnalyzerService
	formatterService *FormatterService
	reportWriter     domain.FileSystemWriter
}

// NewVerificationPipelineService creates a new verification pipeline service
func NewVerificationPipelineService(
	log domain.Logger,
	buildService *BuildService,
	testService *TestService,
	staticAnalyzer *application.StaticAnalyzerService,
	formatterService *FormatterService,
	reportWriter domain.FileSystemWriter,
) *VerificationPipelineService {
	return &VerificationPipelineService{
		log:              log,
		buildService:     buildService,
		testService:      testService,
		staticAnalyzer:   staticAnalyzer,
		formatterService: formatterService,
		reportWriter:     reportWriter,
	}
}

// RunVerificationPipeline executes complete verification pipeline
func (s *VerificationPipelineService) RunVerificationPipeline(ctx context.Context, config *domain.VerificationConfig) (*domain.VerificationResult, error) {
	s.log.Info(fmt.Sprintf("Starting verification pipeline for project: %s", config.ProjectPath))

	result := &domain.VerificationResult{
		ProjectPath: config.ProjectPath,
		Languages:   config.Languages,
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
		Steps:       make([]*domain.VerificationStep, 0),
	}

	// Step 1: Formatting
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
	}

	// Step 2: Build and Type Check
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

	// Step 3: Smoke Tests
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

	// Determine overall success
	result.Success = true
	for _, step := range result.Steps {
		if !step.Success && step.Name != "format" {
			result.Success = false
			break
		}
	}

	result.CompletedAt = time.Now().UTC().Format(time.RFC3339)

	// Save report
	if err := s.saveVerificationReport(ctx, result); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to save verification report: %v", err))
	}

	s.log.Info(fmt.Sprintf("Verification pipeline completed with success: %t", result.Success))
	return result, nil
}

// runFormatStep executes formatting step
func (s *VerificationPipelineService) runFormatStep(ctx context.Context, config *domain.VerificationConfig) (interface{}, error) {
	s.log.Info("Running format step")

	var results []*domain.FormatResult
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

// runBuildTypeCheckStep executes build and type check step
func (s *VerificationPipelineService) runBuildTypeCheckStep(ctx context.Context, config *domain.VerificationConfig) (interface{}, error) {
	s.log.Info("Running build/typecheck step")

	validation, err := s.buildService.ValidateProject(ctx, config.ProjectPath, config.Languages)
	if err != nil {
		return nil, fmt.Errorf("project validation failed: %w", err)
	}

	if !validation.Success {
		return nil, fmt.Errorf("project validation failed for some languages")
	}

	return validation, nil
}

// runSmokeTestsStep executes smoke tests step
func (s *VerificationPipelineService) runSmokeTestsStep(ctx context.Context, config *domain.VerificationConfig) (interface{}, error) {
	s.log.Info("Running smoke tests step")

	var allResults []*domain.TestResult
	for _, language := range config.Languages {
		results, err := s.testService.RunSmokeTests(ctx, config.ProjectPath, language)
		if err != nil {
			return nil, fmt.Errorf("smoke tests failed for %s: %w", language, err)
		}

		validation := s.testService.ValidateTestResults(results)
		if !validation.Success {
			return nil, fmt.Errorf("smoke tests failed for %s: %d tests failed", language, validation.FailedTests)
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// saveVerificationReport saves verification report
func (s *VerificationPipelineService) saveVerificationReport(ctx context.Context, result *domain.VerificationResult) error {
	// Create reports directory
	reportsDir := filepath.Join("reports", "verification")
	if err := s.reportWriter.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Save report (simplified implementation)
	return nil // Simplified for migration
}
