package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"sync"
	"time"
)

// TaskProtocolServiceImpl implements the TaskProtocolService interface
type TaskProtocolServiceImpl struct {
	log                  domain.Logger
	verificationPipeline *VerificationPipelineService
	staticAnalyzer       *StaticAnalyzerService
	testService          *TestService
	buildService         *BuildService
	guardrailService     domain.GuardrailService
	aiService            *IntelligentAIService
	errorAnalyzer        domain.ErrorAnalyzer
	correctionEngine     domain.CorrectionEngine
	mu                   sync.RWMutex
}

// NewTaskProtocolService creates a new TaskProtocolService instance
func NewTaskProtocolService(
	log domain.Logger,
	verificationPipeline *VerificationPipelineService,
	staticAnalyzer *StaticAnalyzerService,
	testService *TestService,
	buildService *BuildService,
	guardrailService domain.GuardrailService,
	aiService *IntelligentAIService,
	errorAnalyzer domain.ErrorAnalyzer,
	correctionEngine domain.CorrectionEngine,
) domain.TaskProtocolService {
	return &TaskProtocolServiceImpl{
		log:                  log,
		verificationPipeline: verificationPipeline,
		staticAnalyzer:       staticAnalyzer,
		testService:          testService,
		buildService:         buildService,
		guardrailService:     guardrailService,
		aiService:            aiService,
		errorAnalyzer:        errorAnalyzer,
		correctionEngine:     correctionEngine,
	}
}

// ExecuteProtocol executes the full verification protocol for a task
func (s *TaskProtocolServiceImpl) ExecuteProtocol(ctx context.Context, config *domain.TaskProtocolConfig) (*domain.TaskProtocolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info(fmt.Sprintf("Starting Task Protocol verification for project: %s", config.ProjectPath))

	result := &domain.TaskProtocolResult{
		TaskID:           generateTaskID(),
		StartedAt:        time.Now(),
		Stages:           make([]*domain.ProtocolStageResult, 0),
		CorrectionCycles: 0,
	}

	// Execute each enabled stage
	for _, stage := range config.EnabledStages {
		stageResult, err := s.executeStageWithRetry(ctx, stage, config)
		result.Stages = append(result.Stages, stageResult)

		if !stageResult.Success {
			if config.FailFast {
				result.Success = false
				result.FinalError = fmt.Sprintf("Stage %s failed: %v", stage, err)
				result.CompletedAt = time.Now()
				return result, fmt.Errorf("protocol failed at stage %s: %w", stage, err)
			}
			s.log.Warning(fmt.Sprintf("Stage %s failed but continuing due to FailFast=false", stage))
		}
	}

	// Determine overall success
	result.Success = s.determineOverallSuccess(result.Stages)
	result.CompletedAt = time.Now()

	s.log.Info(fmt.Sprintf("Task Protocol completed with success: %t, correction cycles: %d",
		result.Success, result.CorrectionCycles))

	return result, nil
}

// executeStageWithRetry executes a stage with self-correction retry logic
func (s *TaskProtocolServiceImpl) executeStageWithRetry(ctx context.Context, stage domain.ProtocolStage, config *domain.TaskProtocolConfig) (*domain.ProtocolStageResult, error) {
	stageResult := &domain.ProtocolStageResult{
		Stage:           stage,
		Attempts:        0,
		CorrectionSteps: make([]*domain.CorrectionStep, 0),
	}

	startTime := time.Now()
	maxAttempts := config.MaxRetries + 1

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		stageResult.Attempts = attempt

		s.log.Info(fmt.Sprintf("Executing stage %s, attempt %d/%d", stage, attempt, maxAttempts))

		// Execute the stage
		err := s.executeStage(ctx, stage, config)

		if err == nil {
			// Stage succeeded
			stageResult.Success = true
			stageResult.Duration = time.Since(startTime)
			return stageResult, nil
		}

		// Stage failed - analyze error and attempt correction if enabled
		if config.SelfCorrection.Enabled && attempt < maxAttempts {
			s.log.Info(fmt.Sprintf("Stage %s failed, attempting self-correction", stage))

			if correctionApplied, corrErr := s.attemptSelfCorrection(ctx, err, stage, config); corrErr == nil && correctionApplied {
				s.log.Info(fmt.Sprintf("Self-correction applied for stage %s", stage))
				continue // Retry the stage
			} else if corrErr != nil {
				s.log.Warning(fmt.Sprintf("Self-correction failed for stage %s: %v", stage, corrErr))
			}
		}

		// Final attempt or correction not enabled/failed
		if attempt == maxAttempts {
			stageResult.Success = false
			stageResult.Duration = time.Since(startTime)
			stageResult.ErrorDetails = s.parseErrorDetails(err, stage)
			return stageResult, err
		}
	}

	return stageResult, fmt.Errorf("stage %s failed after %d attempts", stage, maxAttempts)
}

// executeStage executes a single verification stage
func (s *TaskProtocolServiceImpl) executeStage(ctx context.Context, stage domain.ProtocolStage, config *domain.TaskProtocolConfig) error {
	switch stage {
	case domain.StageLinting:
		return s.executeLintingStage(ctx, config)
	case domain.StageBuilding:
		return s.executeBuildingStage(ctx, config)
	case domain.StageTesting:
		return s.executeTestingStage(ctx, config)
	case domain.StageGuardrails:
		return s.executeGuardrailsStage(ctx, config)
	default:
		return fmt.Errorf("unsupported stage: %s", stage)
	}
}

// executeLintingStage executes static analysis and linting
func (s *TaskProtocolServiceImpl) executeLintingStage(ctx context.Context, config *domain.TaskProtocolConfig) error {
	s.log.Info("Running linting stage")

	report, err := s.staticAnalyzer.AnalyzeProject(ctx, config.ProjectPath, config.Languages)
	if err != nil {
		return fmt.Errorf("static analysis failed: %w", err)
	}

	// Check if there are any errors (not just warnings)
	if s.hasLintingErrors(report) {
		return fmt.Errorf("linting errors found in static analysis report")
	}

	return nil
}

// executeBuildingStage executes build and compilation
func (s *TaskProtocolServiceImpl) executeBuildingStage(ctx context.Context, config *domain.TaskProtocolConfig) error {
	s.log.Info("Running building stage")

	validation, err := s.buildService.ValidateProject(ctx, config.ProjectPath, config.Languages)
	if err != nil {
		return fmt.Errorf("build validation failed: %w", err)
	}

	if !validation.Success {
		return fmt.Errorf("project build failed for one or more languages")
	}

	return nil
}

// executeTestingStage executes relevant tests
func (s *TaskProtocolServiceImpl) executeTestingStage(ctx context.Context, config *domain.TaskProtocolConfig) error {
	s.log.Info("Running testing stage")

	for _, language := range config.Languages {
		results, err := s.testService.RunSmokeTests(ctx, config.ProjectPath, language)
		if err != nil {
			return fmt.Errorf("tests failed for %s: %w", language, err)
		}

		validation := s.testService.ValidateTestResults(results)
		if !validation.Success {
			return fmt.Errorf("test validation failed for %s: %d tests failed", language, validation.FailedTests)
		}
	}

	return nil
}

// executeGuardrailsStage executes guardrail validation
func (s *TaskProtocolServiceImpl) executeGuardrailsStage(ctx context.Context, config *domain.TaskProtocolConfig) error {
	s.log.Info("Running guardrails stage")

	// For now, perform basic guardrail validation
	// In a real implementation, this would check security policies, resource limits, etc.
	taskID := generateTaskID()
	files := []string{}      // Would be populated with actual changed files
	linesChanged := int64(0) // Would be calculated from actual changes

	result, err := s.guardrailService.ValidateTask(taskID, files, linesChanged)
	if err != nil {
		return fmt.Errorf("guardrail validation failed: %w", err)
	}

	if !result.Valid {
		return fmt.Errorf("guardrail validation failed: %s", result.Error)
	}

	return nil
}

// attemptSelfCorrection attempts to apply corrections for a failed stage
func (s *TaskProtocolServiceImpl) attemptSelfCorrection(ctx context.Context, err error, stage domain.ProtocolStage, config *domain.TaskProtocolConfig) (bool, error) {
	// Analyze the error
	errorDetails, analyzeErr := s.errorAnalyzer.AnalyzeError(err.Error(), stage)
	if analyzeErr != nil {
		return false, fmt.Errorf("error analysis failed: %w", analyzeErr)
	}

	// Get correction suggestions
	corrections, corrErr := s.errorAnalyzer.SuggestCorrections(errorDetails)
	if corrErr != nil {
		return false, fmt.Errorf("correction suggestion failed: %w", corrErr)
	}

	if len(corrections) == 0 {
		return false, nil // No corrections available
	}

	// Apply corrections
	correctionResult, applyErr := s.correctionEngine.ApplyCorrections(ctx, corrections, config.ProjectPath)
	if applyErr != nil {
		return false, fmt.Errorf("correction application failed: %w", applyErr)
	}

	return correctionResult.Success, nil
}

// ValidateStage executes a single verification stage
func (s *TaskProtocolServiceImpl) ValidateStage(ctx context.Context, stage domain.ProtocolStage, config *domain.TaskProtocolConfig) (*domain.ProtocolStageResult, error) {
	return s.executeStageWithRetry(ctx, stage, config)
}

// RequestCorrectionGuidance requests AI-generated correction guidance for errors
func (s *TaskProtocolServiceImpl) RequestCorrectionGuidance(ctx context.Context, error *domain.ErrorDetails, taskContext *domain.TaskContext) (*domain.CorrectionGuidance, error) {
	if s.aiService == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Create a prompt for correction guidance
	// prompt := s.buildCorrectionPrompt(error, taskContext)

	// Request AI guidance (simplified implementation)
	// In a real implementation, this would call the AI service with the prompt
	guidance := &domain.CorrectionGuidance{
		Error:       error,
		Steps:       make([]*domain.CorrectionStep, 0),
		Explanation: "AI-generated correction guidance would be provided here",
		Confidence:  0.8,
	}

	return guidance, nil
}

// Helper methods

func (s *TaskProtocolServiceImpl) determineOverallSuccess(stages []*domain.ProtocolStageResult) bool {
	for _, stage := range stages {
		// Critical stages that must pass
		if (stage.Stage == domain.StageBuilding || stage.Stage == domain.StageTesting) && !stage.Success {
			return false
		}
	}
	return true
}

func (s *TaskProtocolServiceImpl) hasLintingErrors(report interface{}) bool {
	// Simplified implementation - check if report indicates errors
	// In a real implementation, this would parse the static analysis report
	return false
}

func (s *TaskProtocolServiceImpl) parseErrorDetails(err error, stage domain.ProtocolStage) *domain.ErrorDetails {
	return &domain.ErrorDetails{
		Stage:     stage,
		ErrorType: domain.ErrorTypeCompilation, // Default type
		Message:   err.Error(),
		Tool:      string(stage),
		Severity:  "error",
	}
}

func (s *TaskProtocolServiceImpl) buildCorrectionPrompt(error *domain.ErrorDetails, context *domain.TaskContext) string {
	return fmt.Sprintf("Error in %s stage: %s. Please provide correction guidance.", error.Stage, error.Message)
}

func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
