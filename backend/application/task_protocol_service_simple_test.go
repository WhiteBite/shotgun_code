package application

import (
	"context"
	"shotgun_code/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTaskProtocolService_ExecuteProtocol tests basic protocol execution
func TestTaskProtocolService_ExecuteProtocol_Basic(t *testing.T) {
	// Create a minimal test setup
	logger := &TestLogger{}
	verificationPipeline := &MockVerificationPipelineService{}
	staticAnalyzer := &MockStaticAnalyzerService{}
	testService := &MockTestService{}
	buildService := &MockBuildService{}
	guardrailService := &MockGuardrailService{}
	aiService := &MockIntelligentAIService{}
	errorAnalyzer := NewErrorAnalyzer(logger)
	fileSystemProvider := &MockFileSystemProvider{}
	correctionEngine := NewCorrectionEngine(logger, fileSystemProvider)

	service := NewTaskProtocolService(
		logger,
		verificationPipeline,
		staticAnalyzer,
		testService,
		buildService,
		guardrailService,
		aiService,
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
			Enabled:     false,
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
	verificationPipeline := &MockVerificationPipelineService{}
	staticAnalyzer := &MockStaticAnalyzerService{}
	testService := &MockTestService{}
	buildService := &MockBuildService{}
	guardrailService := &MockGuardrailService{}
	aiService := &MockIntelligentAIService{}
	errorAnalyzer := NewErrorAnalyzer(logger)
	fileSystemProvider := &MockFileSystemProvider{}
	correctionEngine := NewCorrectionEngine(logger, fileSystemProvider)

	service := NewTaskProtocolService(
		logger,
		verificationPipeline,
		staticAnalyzer,
		testService,
		buildService,
		guardrailService,
		aiService,
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

type MockVerificationPipelineService struct{}

type MockStaticAnalyzerService struct{}

func (m *MockStaticAnalyzerService) AnalyzeProject(ctx context.Context, projectPath string, languages []string) (interface{}, error) {
	return nil, nil
}

type MockTestService struct{}

func (m *MockTestService) RunSmokeTests(ctx context.Context, projectPath, language string) ([]*domain.TestResult, error) {
	return []*domain.TestResult{}, nil
}

func (m *MockTestService) ValidateTestResults(results []*domain.TestResult) *domain.TestValidationResult {
	return &domain.TestValidationResult{Success: true}
}

type MockBuildService struct{}

func (m *MockBuildService) ValidateProject(ctx context.Context, projectPath string, languages []string) (*domain.ProjectValidationResult, error) {
	return &domain.ProjectValidationResult{Success: true}, nil
}

func (m *MockBuildService) DetectLanguages(ctx context.Context, projectPath string) ([]string, error) {
	return []string{"go"}, nil
}

func (m *MockBuildService) GetSupportedLanguages() []string {
	return []string{"go", "typescript", "javascript"}
}

type MockGuardrailService struct{}

func (m *MockGuardrailService) ValidateTask(taskID string, files []string, linesChanged int64) (*domain.TaskValidationResult, error) {
	return &domain.TaskValidationResult{Valid: true}, nil
}

type MockIntelligentAIService struct{}

type MockFileSystemProvider struct{}

func (m *MockFileSystemProvider) ReadFile(filename string) ([]byte, error) {
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
	verificationPipeline := &MockVerificationPipelineService{}
	staticAnalyzer := &MockStaticAnalyzerService{}
	testService := &MockTestService{}
	buildService := &MockBuildService{}
	guardrailService := &MockGuardrailService{}
	aiService := &MockIntelligentAIService{}
	errorAnalyzer := NewErrorAnalyzer(logger)
	fileSystemProvider := &MockFileSystemProvider{}
	correctionEngine := NewCorrectionEngine(logger, fileSystemProvider)

	service := NewTaskProtocolService(
		logger,
		verificationPipeline,
		staticAnalyzer,
		testService,
		buildService,
		guardrailService,
		aiService,
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
			Enabled:     true,
			MaxAttempts: 3,
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