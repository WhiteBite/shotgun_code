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

// TestTaskProtocolService_ExecuteProtocol tests basic protocol execution
func TestTaskProtocolService_ExecuteProtocol_Basic(t *testing.T) {
	// Create a minimal test setup
	logger := &TestLogger{}
	verificationPipeline := &VerificationPipelineService{}
	staticAnalyzer := &testutils.MockStaticAnalyzerService{}
	testService := &testutils.MockTestService{}
	buildService := &testutils.MockBuildService{}
	guardrailService := &testutils.MockGuardrailService{}
	errorAnalyzer := NewErrorAnalyzer(logger)
	fileSystemProvider := &MockFileSystemProvider{}
	correctionEngine := NewCorrectionEngine(logger, fileSystemProvider)

	// Setup mock expectations
	staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)

	service := NewTaskProtocolService(
		logger,
		verificationPipeline,
		staticAnalyzer,
		testService,
		buildService,
		guardrailService,
		&IntelligentAIService{}, // Use concrete type
		errorAnalyzer,
		correctionEngine,
	)

	config := &domain.TaskProtocolConfig{
		ProjectPath: "/test/project",
		Languages:   []string{"go"},
		EnabledStages: []domain.ProtocolStage{
			domain.StageLinting,
		},
		MaxRetries: 1,
		FailFast:   false,
		SelfCorrection: domain.SelfCorrectionConfig{
			Enabled: false,
		},
	}

	// Execute
	result, err := service.ExecuteProtocol(context.Background(), config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result.Stages))
}

// TestTaskProtocolService_ValidateStage tests individual stage validation
func TestTaskProtocolService_ValidateStage_Linting(t *testing.T) {
	// Create a minimal test setup
	logger := &TestLogger{}
	verificationPipeline := &VerificationPipelineService{}
	staticAnalyzer := &testutils.MockStaticAnalyzerService{}
	testService := &testutils.MockTestService{}
	buildService := &testutils.MockBuildService{}
	guardrailService := &testutils.MockGuardrailService{}
	errorAnalyzer := NewErrorAnalyzer(logger)
	fileSystemProvider := &MockFileSystemProvider{}
	correctionEngine := NewCorrectionEngine(logger, fileSystemProvider)

	// Setup mock expectations
	staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)

	service := NewTaskProtocolService(
		logger,
		verificationPipeline,
		staticAnalyzer,
		testService,
		buildService,
		guardrailService,
		&IntelligentAIService{}, // Use concrete type
		errorAnalyzer,
		correctionEngine,
	)

	config := &domain.TaskProtocolConfig{
		ProjectPath: "/test/project",
		Languages:   []string{"go"},
	}

	// Execute
	result, err := service.ValidateStage(context.Background(), domain.StageLinting, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.StageLinting, result.Stage)
}

// Test helpers and mock implementations

type TestLogger struct{}

func (l *TestLogger) Info(message string)    {}
func (l *TestLogger) Warning(message string) {}
func (l *TestLogger) Error(message string)   {}
func (l *TestLogger) Debug(message string)   {}
func (l *TestLogger) Fatal(message string)   {}

type MockFileSystemProvider struct {
	ReadFileContent string
}

func (m *MockFileSystemProvider) ReadFile(filename string) ([]byte, error) {
	if m.ReadFileContent != "" {
		return []byte(m.ReadFileContent), nil
	}
	return []byte("mock content"), nil
}

func (m *MockFileSystemProvider) WriteFile(filename string, data []byte, perm int) error {
	return nil
}

func (m *MockFileSystemProvider) MkdirAll(path string, perm int) error {
	return nil
}

// Integration test
func TestTaskProtocolService_Integration(t *testing.T) {
	// This test demonstrates the integration between components
	logger := &TestLogger{}
	verificationPipeline := &VerificationPipelineService{}
	staticAnalyzer := &testutils.MockStaticAnalyzerService{}
	testService := &testutils.MockTestService{}
	buildService := &testutils.MockBuildService{}
	guardrailService := &testutils.MockGuardrailService{}
	errorAnalyzer := NewErrorAnalyzer(logger)
	fileSystemProvider := &MockFileSystemProvider{}
	correctionEngine := NewCorrectionEngine(logger, fileSystemProvider)

	// Setup mock expectations
	staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)
	buildService.On("ValidateProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.ProjectValidationResult{Success: true}, nil)
	testService.On("RunSmokeTests", mock.Anything, mock.Anything, mock.Anything).Return([]*domain.TestResult{}, nil)
	testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})
	guardrailService.On("ValidateTask", mock.Anything, mock.Anything, mock.Anything).Return(&domain.TaskValidationResult{Valid: true}, nil)

	service := NewTaskProtocolService(
		logger,
		verificationPipeline,
		staticAnalyzer,
		testService,
		buildService,
		guardrailService,
		&IntelligentAIService{}, // Use concrete type
		errorAnalyzer,
		correctionEngine,
	)

	config := &domain.TaskProtocolConfig{
		ProjectPath: "/test/project",
		Languages:   []string{"go", "typescript"},
		EnabledStages: []domain.ProtocolStage{
			domain.StageLinting,
			domain.StageBuilding,
			domain.StageTesting,
			domain.StageGuardrails,
		},
		MaxRetries: 2,
		FailFast:   false,
		SelfCorrection: domain.SelfCorrectionConfig{
			Enabled:      true,
			MaxAttempts:  3,
			AIAssistance: true,
		},
		Timeouts: map[string]time.Duration{
			"linting":    2 * time.Minute,
			"building":   5 * time.Minute,
			"testing":    10 * time.Minute,
			"guardrails": 1 * time.Minute,
		},
	}

	// Execute the full protocol
	start := time.Now()
	result, err := service.ExecuteProtocol(context.Background(), config)
	duration := time.Since(start)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, 4, len(result.Stages))
	assert.Less(t, duration, 10*time.Second, "Should complete quickly in test environment")

	// Verify all stages were executed
	stageNames := make(map[domain.ProtocolStage]bool)
	for _, stage := range result.Stages {
		stageNames[stage.Stage] = true
	}

	assert.True(t, stageNames[domain.StageLinting], "Linting stage should be executed")
	assert.True(t, stageNames[domain.StageBuilding], "Building stage should be executed")
	assert.True(t, stageNames[domain.StageTesting], "Testing stage should be executed")
	assert.True(t, stageNames[domain.StageGuardrails], "Guardrails stage should be executed")

	t.Logf("Protocol execution completed in %v with %d stages", duration, len(result.Stages))
}