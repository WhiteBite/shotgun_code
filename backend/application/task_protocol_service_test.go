package application

import (
	"context"
	"shotgun_code/domain"
	"shotgun_code/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestTaskProtocolService_ExecuteProtocol tests the main protocol execution workflow
func TestTaskProtocolService_ExecuteProtocol(t *testing.T) {
	tests := []struct {
		name           string
		config         *domain.TaskProtocolConfig
		mockSetup      func(*testMocks)
		expectedResult func(*domain.TaskProtocolResult) bool
		expectedError  string
	}{
		{
			name: "successful_protocol_execution_all_stages",
			config: &domain.TaskProtocolConfig{
				ProjectPath: "/test/project",
				Languages:   []string{"go", "typescript"},
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
			},
			mockSetup: func(mocks *testMocks) {
				// Setup successful execution for all stages
				mocks.staticAnalyzer.On("AnalyzeProject", mock.Anything, "/test/project", []string{"go", "typescript"}).Return(&domain.StaticAnalysisReport{}, nil)
				mocks.buildService.On("ValidateProject", mock.Anything, "/test/project", []string{"go", "typescript"}).Return(&domain.ProjectValidationResult{Success: true}, nil)
				mocks.testService.On("RunSmokeTests", mock.Anything, "/test/project", "go").Return([]*domain.TestResult{}, nil)
				mocks.testService.On("RunSmokeTests", mock.Anything, "/test/project", "typescript").Return([]*domain.TestResult{}, nil)
				mocks.testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})
				mocks.guardrailService.On("ValidateTask", mock.Anything, mock.Anything, mock.Anything).Return(&domain.TaskValidationResult{Valid: true}, nil)
			},
			expectedResult: func(result *domain.TaskProtocolResult) bool {
				return result.Success &&
					len(result.Stages) == 4 &&
					result.CorrectionCycles == 0 &&
					result.FinalError == ""
			},
		},
		{
			name: "protocol_with_failing_build_stage",
			config: &domain.TaskProtocolConfig{
				ProjectPath: "/test/project",
				Languages:   []string{"go"},
				EnabledStages: []domain.ProtocolStage{
					domain.StageLinting,
					domain.StageBuilding,
					domain.StageTesting,
				},
				MaxRetries: 1,
				FailFast:   true,
				SelfCorrection: domain.SelfCorrectionConfig{
					Enabled: false,
				},
			},
			mockSetup: func(mocks *testMocks) {
				// Linting succeeds
				mocks.staticAnalyzer.On("AnalyzeProject", mock.Anything, "/test/project", []string{"go"}).Return(&domain.StaticAnalysisReport{}, nil)
				// Building fails
				mocks.buildService.On("ValidateProject", mock.Anything, "/test/project", []string{"go"}).Return(&domain.ProjectValidationResult{Success: false}, nil)
			},
			expectedResult: func(result *domain.TaskProtocolResult) bool {
				return !result.Success &&
					len(result.Stages) >= 2 && // Should have linting and building stages
					result.FinalError != ""
			},
		},
		{
			name: "protocol_with_self_correction",
			config: &domain.TaskProtocolConfig{
				ProjectPath: "/test/project",
				Languages:   []string{"go"},
				EnabledStages: []domain.ProtocolStage{
					domain.StageBuilding,
				},
				MaxRetries: 2,
				FailFast:   false,
				SelfCorrection: domain.SelfCorrectionConfig{
					Enabled:      true,
					MaxAttempts:  3,
					AIAssistance: true,
				},
			},
			mockSetup: func(mocks *testMocks) {
				// First attempt fails, second succeeds after correction
				mocks.buildService.On("ValidateProject", mock.Anything, "/test/project", []string{"go"}).Return(&domain.ProjectValidationResult{Success: false}, assert.AnError).Once()
				mocks.buildService.On("ValidateProject", mock.Anything, "/test/project", []string{"go"}).Return(&domain.ProjectValidationResult{Success: true}, nil).Once()

				// Error analysis and correction
				mocks.errorAnalyzer.On("AnalyzeError", mock.Anything, domain.StageBuilding).Return(&domain.ErrorDetails{
					Stage:     domain.StageBuilding,
					ErrorType: domain.ErrorTypeCompilation,
					Message:   "compilation failed",
				}, nil)
				mocks.errorAnalyzer.On("SuggestCorrections", mock.Anything).Return([]*domain.CorrectionStep{
					{
						Action:      domain.ActionFixSyntax,
						Target:      "main.go",
						Description: "Fix syntax error",
					},
				}, nil)
				mocks.correctionEngine.On("ApplyCorrections", mock.Anything, mock.Anything, "/test/project").Return(&domain.CorrectionResult{
					Success: true,
					Message: "Corrections applied",
				}, nil)
			},
			expectedResult: func(result *domain.TaskProtocolResult) bool {
				return result.Success &&
					result.CorrectionCycles > 0 &&
					len(result.Stages) == 1 &&
					result.Stages[0].Attempts > 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mocks := createTestMocks()
			tt.mockSetup(mocks)

			// Create concrete services where needed
			verificationPipeline := &VerificationPipelineService{}
			aiService := &IntelligentAIService{}

			service := NewTaskProtocolService(
				mocks.logger,
				verificationPipeline,
				mocks.staticAnalyzer,
				mocks.testService,
				mocks.buildService,
				mocks.guardrailService,
				aiService,
				mocks.errorAnalyzer,
				mocks.correctionEngine,
			)

			// Execute
			result, err := service.ExecuteProtocol(context.Background(), tt.config)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, tt.expectedResult(result), "Result validation failed")
			}

			// Verify mocks
			mocks.assertExpectations(t)
		})
	}
}

// TestTaskProtocolService_ValidateStage tests individual stage validation
func TestTaskProtocolService_ValidateStage(t *testing.T) {
	tests := []struct {
		name            string
		stage           domain.ProtocolStage
		config          *domain.TaskProtocolConfig
		mockSetup       func(*testMocks)
		expectedSuccess bool
		expectedError   string
	}{
		{
			name:  "linting_stage_success",
			stage: domain.StageLinting,
			config: &domain.TaskProtocolConfig{
				ProjectPath: "/test/project",
				Languages:   []string{"go"},
			},
			mockSetup: func(mocks *testMocks) {
				mocks.staticAnalyzer.On("AnalyzeProject", mock.Anything, "/test/project", []string{"go"}).Return(&domain.StaticAnalysisReport{}, nil)
			},
			expectedSuccess: true,
		},
		{
			name:  "building_stage_success",
			stage: domain.StageBuilding,
			config: &domain.TaskProtocolConfig{
				ProjectPath: "/test/project",
				Languages:   []string{"go"},
			},
			mockSetup: func(mocks *testMocks) {
				mocks.buildService.On("ValidateProject", mock.Anything, "/test/project", []string{"go"}).Return(&domain.ProjectValidationResult{Success: true}, nil)
			},
			expectedSuccess: true,
		},
		{
			name:  "testing_stage_success",
			stage: domain.StageTesting,
			config: &domain.TaskProtocolConfig{
				ProjectPath: "/test/project",
				Languages:   []string{"go"},
			},
			mockSetup: func(mocks *testMocks) {
				mocks.testService.On("RunSmokeTests", mock.Anything, "/test/project", "go").Return([]*domain.TestResult{}, nil)
				mocks.testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})
			},
			expectedSuccess: true,
		},
		{
			name:  "guardrails_stage_success",
			stage: domain.StageGuardrails,
			config: &domain.TaskProtocolConfig{
				ProjectPath: "/test/project",
				Languages:   []string{"go"},
			},
			mockSetup: func(mocks *testMocks) {
				mocks.guardrailService.On("ValidateTask", mock.Anything, mock.Anything, mock.Anything).Return(&domain.TaskValidationResult{Valid: true}, nil)
			},
			expectedSuccess: true,
		},
		{
			name:  "unsupported_stage",
			stage: domain.ProtocolStage("unsupported"),
			config: &domain.TaskProtocolConfig{
				ProjectPath: "/test/project",
				Languages:   []string{"go"},
			},
			mockSetup: func(mocks *testMocks) {
				// No setup needed
			},
			expectedSuccess: false,
			expectedError:   "unsupported stage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mocks := createTestMocks()
			tt.mockSetup(mocks)

			// Create concrete services where needed
			verificationPipeline := &VerificationPipelineService{}
			aiService := &IntelligentAIService{}

			service := NewTaskProtocolService(
				mocks.logger,
				verificationPipeline,
				mocks.staticAnalyzer,
				mocks.testService,
				mocks.buildService,
				mocks.guardrailService,
				aiService,
				mocks.errorAnalyzer,
				mocks.correctionEngine,
			)

			// Execute
			result, err := service.ValidateStage(context.Background(), tt.stage, tt.config)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSuccess, result.Success)
				assert.Equal(t, tt.stage, result.Stage)
			}

			// Verify mocks
			mocks.assertExpectations(t)
		})
	}
}

// TestTaskProtocolService_RequestCorrectionGuidance tests AI-powered correction guidance
func TestTaskProtocolService_RequestCorrectionGuidance(t *testing.T) {
	tests := []struct {
		name        string
		error       *domain.ErrorDetails
		context     *domain.TaskContext
		mockSetup   func(*testMocks)
		expectError bool
	}{
		{
			name: "successful_guidance_request",
			error: &domain.ErrorDetails{
				Stage:     domain.StageBuilding,
				ErrorType: domain.ErrorTypeCompilation,
				Message:   "compilation failed",
			},
			context: &domain.TaskContext{
				ProjectPath: "/test/project",
			},
			mockSetup: func(mocks *testMocks) {
				// Mock AI service to return guidance
			},
			expectError: false,
		},
		{
			name: "no_ai_service_available",
			error: &domain.ErrorDetails{
				Stage:     domain.StageBuilding,
				ErrorType: domain.ErrorTypeCompilation,
				Message:   "compilation failed",
			},
			context: &domain.TaskContext{
				ProjectPath: "/test/project",
			},
			mockSetup: func(mocks *testMocks) {
				// Set AI service to nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mocks := createTestMocks()
			tt.mockSetup(mocks)

			// Create concrete services where needed
			verificationPipeline := &VerificationPipelineService{}
			var aiService *IntelligentAIService
			if !tt.expectError {
				aiService = &IntelligentAIService{}
			}

			service := NewTaskProtocolService(
				mocks.logger,
				verificationPipeline,
				mocks.staticAnalyzer,
				mocks.testService,
				mocks.buildService,
				mocks.guardrailService,
				aiService,
				mocks.errorAnalyzer,
				mocks.correctionEngine,
			)

			// Execute
			guidance, err := service.RequestCorrectionGuidance(context.Background(), tt.error, tt.context)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, guidance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, guidance)
				assert.Equal(t, tt.error, guidance.Error)
			}
		})
	}
}

// Test helpers and mocks

type testMocks struct {
	logger               *testutils.MockLogger
	verificationPipeline *testutils.MockVerificationPipelineService
	staticAnalyzer       *testutils.MockStaticAnalyzerService
	testService          *testutils.MockTestService
	buildService         *testutils.MockBuildService
	guardrailService     *testutils.MockGuardrailService
	aiService            *testutils.MockIntelligentAIService
	errorAnalyzer        *MockErrorAnalyzer
	correctionEngine     *MockCorrectionEngine
}

func createTestMocks() *testMocks {
	return &testMocks{
		logger:               testutils.NewMockLogger(),
		verificationPipeline: &testutils.MockVerificationPipelineService{},
		staticAnalyzer:       &testutils.MockStaticAnalyzerService{},
		testService:          &testutils.MockTestService{},
		buildService:         &testutils.MockBuildService{},
		guardrailService:     &testutils.MockGuardrailService{},
		aiService:            &testutils.MockIntelligentAIService{},
		errorAnalyzer:        &MockErrorAnalyzer{},
		correctionEngine:     &MockCorrectionEngine{},
	}
}

func (m *testMocks) assertExpectations(t *testing.T) {
	m.verificationPipeline.AssertExpectations(t)
	m.staticAnalyzer.AssertExpectations(t)
	m.testService.AssertExpectations(t)
	m.buildService.AssertExpectations(t)
	m.guardrailService.AssertExpectations(t)
	m.errorAnalyzer.AssertExpectations(t)
	m.correctionEngine.AssertExpectations(t)
}

// Mock implementations

type MockErrorAnalyzer struct {
	mock.Mock
}

func (m *MockErrorAnalyzer) AnalyzeError(errorOutput string, stage domain.ProtocolStage) (*domain.ErrorDetails, error) {
	args := m.Called(errorOutput, stage)
	return args.Get(0).(*domain.ErrorDetails), args.Error(1)
}

func (m *MockErrorAnalyzer) SuggestCorrections(error *domain.ErrorDetails) ([]*domain.CorrectionStep, error) {
	args := m.Called(error)
	return args.Get(0).([]*domain.CorrectionStep), args.Error(1)
}

func (m *MockErrorAnalyzer) ClassifyErrorType(errorOutput string) domain.ErrorType {
	args := m.Called(errorOutput)
	return args.Get(0).(domain.ErrorType)
}

type MockCorrectionEngine struct {
	mock.Mock
}

func (m *MockCorrectionEngine) ApplyCorrection(ctx context.Context, step *domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	args := m.Called(ctx, step, projectPath)
	return args.Get(0).(*domain.CorrectionResult), args.Error(1)
}

func (m *MockCorrectionEngine) ApplyCorrections(ctx context.Context, steps []*domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	args := m.Called(ctx, steps, projectPath)
	return args.Get(0).(*domain.CorrectionResult), args.Error(1)
}

func (m *MockCorrectionEngine) CanHandle(error *domain.ErrorDetails) bool {
	args := m.Called(error)
	return args.Bool(0)
}

// Performance and edge case tests

// TestTaskProtocolService_Performance tests protocol execution performance
func TestTaskProtocolService_Performance(t *testing.T) {
	// Test that protocol execution completes within reasonable time
	mocks := createTestMocks()
	mocks.staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)
	mocks.buildService.On("ValidateProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.ProjectValidationResult{Success: true}, nil)
	mocks.testService.On("RunSmokeTests", mock.Anything, mock.Anything, mock.Anything).Return([]*domain.TestResult{}, nil)
	mocks.testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})
	mocks.guardrailService.On("ValidateTask", mock.Anything, mock.Anything, mock.Anything).Return(&domain.TaskValidationResult{Valid: true}, nil)

	// Create concrete services where needed
	verificationPipeline := &VerificationPipelineService{}
	aiService := &IntelligentAIService{}

	service := NewTaskProtocolService(
		mocks.logger,
		verificationPipeline,
		mocks.staticAnalyzer,
		mocks.testService,
		mocks.buildService,
		mocks.guardrailService,
		aiService,
		mocks.errorAnalyzer,
		mocks.correctionEngine,
	)

	config := &domain.TaskProtocolConfig{
		ProjectPath: "/test/project",
		Languages:   []string{"go", "typescript", "javascript"},
		EnabledStages: []domain.ProtocolStage{
			domain.StageLinting,
			domain.StageBuilding,
			domain.StageTesting,
			domain.StageGuardrails,
		},
		MaxRetries: 1,
		FailFast:   false,
	}

	// Test execution time
	start := time.Now()
	result, err := service.ExecuteProtocol(context.Background(), config)
	duration := time.Since(start)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Less(t, duration, 5*time.Second, "Protocol execution should complete quickly in test environment")

	t.Logf("Protocol execution completed in %v", duration)
}

// TestTaskProtocolService_ConcurrentExecution tests concurrent protocol execution
func TestTaskProtocolService_ConcurrentExecution(t *testing.T) {
	// Test that multiple protocol executions can run concurrently
	mocks := createTestMocks()
	mocks.staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)
	mocks.buildService.On("ValidateProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.ProjectValidationResult{Success: true}, nil)
	mocks.testService.On("RunSmokeTests", mock.Anything, mock.Anything, mock.Anything).Return([]*domain.TestResult{}, nil)
	mocks.testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})
	mocks.guardrailService.On("ValidateTask", mock.Anything, mock.Anything, mock.Anything).Return(&domain.TaskValidationResult{Valid: true}, nil)

	// Create concrete services where needed
	verificationPipeline := &VerificationPipelineService{}
	aiService := &IntelligentAIService{}

	service := NewTaskProtocolService(
		mocks.logger,
		verificationPipeline,
		mocks.staticAnalyzer,
		mocks.testService,
		mocks.buildService,
		mocks.guardrailService,
		aiService,
		mocks.errorAnalyzer,
		mocks.correctionEngine,
	)

	config := &domain.TaskProtocolConfig{
		ProjectPath: "/test/project",
		Languages:   []string{"go"},
		EnabledStages: []domain.ProtocolStage{
			domain.StageLinting,
		},
		MaxRetries: 1,
		FailFast:   false,
	}

	// Run multiple executions concurrently
	const numGoroutines = 5
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := service.ExecuteProtocol(context.Background(), config)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}