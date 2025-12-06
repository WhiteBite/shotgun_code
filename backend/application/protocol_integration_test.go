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

// TestTaskProtocolIntegration_EndToEnd tests the complete integration workflow
func TestTaskProtocolIntegration_EndToEnd(t *testing.T) {
	// Setup - create all required services
	logger := &TestLogger{}

	// Create mock services using testutils for interfaces
	staticAnalyzer := &testutils.MockStaticAnalyzerService{}
	testService := &testutils.MockTestService{}
	buildService := &testutils.MockBuildService{}
	guardrailService := &testutils.MockGuardrailService{}
	fileSystemProvider := &MockFileSystemProvider{}

	// Setup mock expectations
	staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)
	buildService.On("ValidateProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.ProjectValidationResult{Success: true}, nil)
	testService.On("RunSmokeTests", mock.Anything, mock.Anything, mock.Anything).Return([]*domain.TestResult{}, nil)
	testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})
	guardrailService.On("ValidateTask", mock.Anything, mock.Anything, mock.Anything).Return(&domain.TaskValidationResult{Valid: true}, nil)

	// Create concrete services for types that expect concrete implementations
	verificationPipeline := &VerificationPipelineService{}
	taskflowService := &testutils.MockTaskflowService{}

	// Create protocol services
	errorAnalyzer := NewErrorAnalyzer(logger)
	correctionEngine := NewCorrectionEngine(logger, fileSystemProvider)

	taskProtocolService := NewTaskProtocolService(
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

	// Create configuration service
	configService := NewTaskProtocolConfigService(logger, fileSystemProvider)

	taskflowService.On("ExecuteTaskflow").Return(nil)
	taskflowService.On("LoadTasks").Return([]domain.Task{
		{ID: "task-1", Name: "Test Task", State: domain.TaskStateTodo},
	}, nil)
	taskflowService.On("GetTaskStatus", mock.Anything).Return(&domain.TaskStatus{
		TaskID: "task-1",
		State:  domain.TaskStateDone,
	}, nil)
	taskflowService.On("UpdateTaskStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create integration service
	integration := NewTaskflowProtocolIntegration(
		logger,
		taskflowService,
		taskProtocolService,
		configService,
		&IntelligentAIService{}, // Use concrete type
	)

	// Test AI code validation workflow
	t.Run("AI_Code_Validation_Workflow", func(t *testing.T) {
		request := &AICodeValidationRequest{
			ProjectPath:  "/test/project",
			Context:      "Go code with potential issues",
			Languages:    []string{"go"},
			ChangedFiles: []string{"main.go", "utils.go"},
		}

		result, err := integration.ValidateAIGeneratedCode(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Stages)
	})

	// Test taskflow integration
	t.Run("Taskflow_Integration_Workflow", func(t *testing.T) {
		options := &TaskflowProtocolOptions{
			FailFast:   false,
			ForceRerun: true,
			Parallel:   false,
		}

		result, err := integration.ExecuteTaskflowWithProtocol(context.Background(), options)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.TaskResults)
	})

	// Test protocol configuration creation
	t.Run("Protocol_Configuration_Creation", func(t *testing.T) {
		aiRequest := &AICodeGenerationRequest{
			ProjectPath: "/test/project",
			Context:     "package main\nimport \"fmt\"\nfunc main() {}",
			Languages:   []string{"go"},
			Task:        "Create simple Go program",
		}

		config, err := integration.CreateTaskProtocolForAIGeneration(context.Background(), aiRequest)

		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "/test/project", config.ProjectPath)
		assert.Contains(t, config.Languages, "go")
		assert.True(t, config.SelfCorrection.Enabled)
		assert.True(t, config.SelfCorrection.AIAssistance)
	})
}

// TestGoProtocolImplementation_Integration tests Go-specific protocol integration
func TestGoProtocolImplementation_Integration(t *testing.T) {
	logger := &TestLogger{}
	staticAnalyzer := &testutils.MockStaticAnalyzerService{}
	buildService := &testutils.MockBuildService{}
	testService := &testutils.MockTestService{}

	// Setup mock expectations
	staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)
	buildService.On("ValidateProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.ProjectValidationResult{Success: true}, nil)
	testService.On("RunSmokeTests", mock.Anything, mock.Anything, mock.Anything).Return([]*domain.TestResult{}, nil)
	testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})

	goProtocol := NewGoProtocolImplementation(logger, staticAnalyzer, buildService, testService)

	config := &domain.TaskProtocolConfig{
		ProjectPath: "/test/go/project",
		Languages:   []string{"go"},
	}

	t.Run("Go_Linting_Stage", func(t *testing.T) {
		err := goProtocol.ExecuteLintingStage(context.Background(), config)
		assert.NoError(t, err)
	})

	t.Run("Go_Building_Stage", func(t *testing.T) {
		err := goProtocol.ExecuteBuildingStage(context.Background(), config)
		assert.NoError(t, err)
	})

	t.Run("Go_Testing_Stage", func(t *testing.T) {
		err := goProtocol.ExecuteTestingStage(context.Background(), config)
		assert.NoError(t, err)
	})

	t.Run("Go_Error_Analysis", func(t *testing.T) {
		errorOutput := "main.go:10:5: undefined: fmt"
		details, err := goProtocol.AnalyzeGoError(errorOutput)

		assert.NoError(t, err)
		assert.NotNil(t, details)
		assert.Equal(t, domain.ErrorTypeCompilation, details.ErrorType)
		assert.Equal(t, "go", details.Tool)
		assert.NotEmpty(t, details.Suggestions)
	})

	t.Run("Go_Correction_Suggestions", func(t *testing.T) {
		errorDetails := &domain.ErrorDetails{
			Stage:      domain.StageBuilding,
			ErrorType:  domain.ErrorTypeCompilation,
			Message:    "undefined: fmt",
			SourceFile: "main.go",
		}

		corrections, err := goProtocol.SuggestGoCorrections(errorDetails)

		assert.NoError(t, err)
		assert.NotEmpty(t, corrections)
		assert.Equal(t, domain.ActionFixImport, corrections[0].Action)
	})
}

// TestTypeScriptProtocolImplementation_Integration tests TypeScript-specific protocol integration
func TestTypeScriptProtocolImplementation_Integration(t *testing.T) {
	logger := &TestLogger{}
	staticAnalyzer := &testutils.MockStaticAnalyzerService{}
	buildService := &testutils.MockBuildService{}
	testService := &testutils.MockTestService{}

	// Setup mock expectations
	staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)
	buildService.On("ValidateProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.ProjectValidationResult{Success: true}, nil)
	testService.On("RunSmokeTests", mock.Anything, mock.Anything, mock.Anything).Return([]*domain.TestResult{}, nil)
	testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})

	tsProtocol := NewTypeScriptProtocolImplementation(logger, staticAnalyzer, buildService, testService)

	config := &domain.TaskProtocolConfig{
		ProjectPath: "/test/ts/project",
		Languages:   []string{"typescript"},
	}

	t.Run("TypeScript_Linting_Stage", func(t *testing.T) {
		err := tsProtocol.ExecuteLintingStage(context.Background(), config)
		assert.NoError(t, err)
	})

	t.Run("TypeScript_Building_Stage", func(t *testing.T) {
		err := tsProtocol.ExecuteBuildingStage(context.Background(), config)
		assert.NoError(t, err)
	})

	t.Run("TypeScript_Testing_Stage", func(t *testing.T) {
		err := tsProtocol.ExecuteTestingStage(context.Background(), config)
		assert.NoError(t, err)
	})

	t.Run("TypeScript_Error_Analysis", func(t *testing.T) {
		errorOutput := "src/main.ts:25:10: error TS2304: Cannot find name 'SomeType'"
		details, err := tsProtocol.AnalyzeTypeScriptError(errorOutput)

		assert.NoError(t, err)
		assert.NotNil(t, details)
		assert.Equal(t, domain.ErrorTypeImport, details.ErrorType)
		assert.Equal(t, "tsc", details.Tool)
		assert.NotEmpty(t, details.Suggestions)
	})

	t.Run("TypeScript_Correction_Suggestions", func(t *testing.T) {
		errorDetails := &domain.ErrorDetails{
			Stage:      domain.StageBuilding,
			ErrorType:  domain.ErrorTypeImport,
			Message:    "TS2304: Cannot find name 'SomeType'",
			SourceFile: "src/main.ts",
		}

		corrections, err := tsProtocol.SuggestTypeScriptCorrections(errorDetails)

		assert.NoError(t, err)
		assert.NotEmpty(t, corrections)
		assert.Equal(t, domain.ActionFixImport, corrections[0].Action)
	})
}

// TestProtocolConfigurationManagement tests configuration management integration
func TestProtocolConfigurationManagement(t *testing.T) {
	logger := &TestLogger{}
	fileSystemProvider := &MockFileSystemProvider{}
	configService := NewTaskProtocolConfigService(logger, fileSystemProvider)

	t.Run("Load_Default_Configuration", func(t *testing.T) {
		config, err := configService.LoadConfiguration("config/task_protocol.yaml")

		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.NotEmpty(t, config.Languages)
		assert.NotEmpty(t, config.EnabledStages)
		assert.True(t, config.SelfCorrection.Enabled)
	})

	t.Run("Project_Specific_Configuration", func(t *testing.T) {
		config, err := configService.GetConfigurationForProject(
			context.Background(),
			"/test/project",
			[]string{"go", "typescript"},
		)

		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "/test/project", config.ProjectPath)
		assert.Contains(t, config.Languages, "go")
		assert.Contains(t, config.Languages, "typescript")
	})

	t.Run("Configuration_Validation", func(t *testing.T) {
		config := &domain.TaskProtocolConfig{
			ProjectPath: "/test/project",
			Languages:   []string{"go"},
			EnabledStages: []domain.ProtocolStage{
				domain.StageLinting,
				domain.StageBuilding,
			},
			MaxRetries: 3,
			SelfCorrection: domain.SelfCorrectionConfig{
				Enabled:     true,
				MaxAttempts: 5,
			},
		}

		err := configService.ValidateConfiguration(config)
		assert.NoError(t, err)
	})

	t.Run("Stage_Configuration", func(t *testing.T) {
		stageConfig, err := configService.GetStageConfiguration(domain.StageLinting, "go")

		assert.NoError(t, err)
		assert.NotNil(t, stageConfig)
		assert.NotEmpty(t, stageConfig["tools"])
	})
}

// TestProtocolReportingAndMetrics tests reporting and metrics integration
func TestProtocolReportingAndMetrics(t *testing.T) {
	logger := &TestLogger{}
	fileSystemProvider := &MockFileSystemProvider{}
	fileSystemWriter := &MockFileSystemWriter{}

	// Create services
	staticAnalyzer := &testutils.MockStaticAnalyzerService{}
	testService := &testutils.MockTestService{}
	buildService := &testutils.MockBuildService{}
	guardrailService := &testutils.MockGuardrailService{}
	errorAnalyzer := NewErrorAnalyzer(logger)
	correctionEngine := NewCorrectionEngine(logger, fileSystemProvider)

	// Setup mock expectations
	staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)
	buildService.On("ValidateProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.ProjectValidationResult{Success: true}, nil)
	testService.On("RunSmokeTests", mock.Anything, mock.Anything, mock.Anything).Return([]*domain.TestResult{}, nil)
	testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})
	guardrailService.On("ValidateTask", mock.Anything, mock.Anything, mock.Anything).Return(&domain.TaskValidationResult{Valid: true}, nil)

	// Create real services that require concrete implementations
	formatterService := NewFormatterService(logger, &MockCommandRunner{})

	taskProtocolService := NewTaskProtocolService(
		logger,
		nil, // verification pipeline not needed for this test
		staticAnalyzer,
		testService,
		buildService,
		guardrailService,
		&IntelligentAIService{}, // Use concrete type
		errorAnalyzer,
		correctionEngine,
	)

	verificationPipeline := NewVerificationPipelineService(
		logger,
		buildService,
		testService,
		staticAnalyzer,
		formatterService,
		fileSystemWriter,
		taskProtocolService,
	)

	t.Run("Protocol_Report_Generation", func(t *testing.T) {
		config := &domain.TaskProtocolConfig{
			ProjectPath: "/test/project",
			Languages:   []string{"go"},
			EnabledStages: []domain.ProtocolStage{
				domain.StageLinting,
				domain.StageBuilding,
			},
			MaxRetries: 1,
		}

		result, err := taskProtocolService.ExecuteProtocol(context.Background(), config)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.TaskID)
		assert.NotZero(t, result.StartedAt)
		assert.NotZero(t, result.CompletedAt)
		assert.Equal(t, 2, len(result.Stages))
	})

	t.Run("Verification_Pipeline_Integration", func(t *testing.T) {
		config := &domain.TaskProtocolConfig{
			ProjectPath: "/test/project",
			Languages:   []string{"go"},
			EnabledStages: []domain.ProtocolStage{
				domain.StageLinting,
			},
		}

		result, err := verificationPipeline.RunTaskProtocolVerification(context.Background(), config)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
	})

	t.Run("Protocol_Configuration_Creation", func(t *testing.T) {
		verifyConfig := &domain.VerificationConfig{
			ProjectPath: "/test/project",
			Languages:   []string{"go", "typescript"},
		}

		protocolConfig := verificationPipeline.CreateTaskProtocolConfig(verifyConfig)

		assert.NotNil(t, protocolConfig)
		assert.Equal(t, verifyConfig.ProjectPath, protocolConfig.ProjectPath)
		assert.Equal(t, verifyConfig.Languages, protocolConfig.Languages)
		assert.True(t, protocolConfig.SelfCorrection.Enabled)
		assert.NotEmpty(t, protocolConfig.EnabledStages)
	})
}

// Performance and stress tests

func TestProtocolPerformance(t *testing.T) {
	logger := &TestLogger{}
	fileSystemProvider := &MockFileSystemProvider{}

	// Create services
	staticAnalyzer := &testutils.MockStaticAnalyzerService{}
	testService := &testutils.MockTestService{}
	buildService := &testutils.MockBuildService{}
	guardrailService := &testutils.MockGuardrailService{}
	errorAnalyzer := NewErrorAnalyzer(logger)
	correctionEngine := NewCorrectionEngine(logger, fileSystemProvider)

	// Setup mock expectations
	staticAnalyzer.On("AnalyzeProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.StaticAnalysisReport{}, nil)
	buildService.On("ValidateProject", mock.Anything, mock.Anything, mock.Anything).Return(&domain.ProjectValidationResult{Success: true}, nil)
	testService.On("RunSmokeTests", mock.Anything, mock.Anything, mock.Anything).Return([]*domain.TestResult{}, nil)
	testService.On("ValidateTestResults", mock.Anything).Return(&domain.TestValidationResult{Success: true})
	guardrailService.On("ValidateTask", mock.Anything, mock.Anything, mock.Anything).Return(&domain.TaskValidationResult{Valid: true}, nil)

	taskProtocolService := NewTaskProtocolService(
		logger,
		nil,
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
		Languages:   []string{"go", "typescript", "javascript"},
		EnabledStages: []domain.ProtocolStage{
			domain.StageLinting,
			domain.StageBuilding,
			domain.StageTesting,
			domain.StageGuardrails,
		},
		MaxRetries: 2,
	}

	t.Run("Protocol_Execution_Performance", func(t *testing.T) {
		start := time.Now()
		result, err := taskProtocolService.ExecuteProtocol(context.Background(), config)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Less(t, duration, 10*time.Second, "Protocol should complete within reasonable time")

		t.Logf("Protocol execution took %v for %d languages and %d stages",
			duration, len(config.Languages), len(config.EnabledStages))
	})

	t.Run("Concurrent_Protocol_Execution", func(t *testing.T) {
		const numGoroutines = 3
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := taskProtocolService.ExecuteProtocol(context.Background(), config)
				results <- err
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})
}

// Mock services for integration testing

type MockTaskflowService struct{}

func (m *MockTaskflowService) LoadTasks() ([]domain.Task, error) {
	return []domain.Task{
		{
			ID:    "task-1",
			Name:  "Test Task 1",
			State: domain.TaskStateTodo,
			Metadata: map[string]interface{}{
				"projectPath": "/test/project",
				"languages":   []string{"go"},
			},
		},
		{
			ID:    "task-2",
			Name:  "Test Task 2",
			State: domain.TaskStateTodo,
			Metadata: map[string]interface{}{
				"projectPath": "/test/project",
				"languages":   []string{"typescript"},
			},
		},
	}, nil
}

func (m *MockTaskflowService) GetTaskStatus(taskID string) (*domain.TaskStatus, error) {
	return &domain.TaskStatus{
		TaskID: taskID,
		State:  domain.TaskStateTodo,
	}, nil
}

func (m *MockTaskflowService) UpdateTaskStatus(taskID string, state domain.TaskState, message string) error {
	return nil
}

func (m *MockTaskflowService) ExecuteTask(taskID string) error {
	return nil
}

func (m *MockTaskflowService) GetTaskDependencies(taskID string) ([]domain.Task, error) {
	return []domain.Task{}, nil
}

func (m *MockTaskflowService) ExecuteTaskflow() error {
	return nil
}

func (m *MockTaskflowService) GetReadyTasks() ([]domain.Task, error) {
	return []domain.Task{}, nil
}

func (m *MockTaskflowService) ValidateTaskflow() error {
	return nil
}

func (m *MockTaskflowService) GetTaskflowProgress() (float64, error) {
	return 0.0, nil
}

func (m *MockTaskflowService) ResetTaskflow() error {
	return nil
}

func (m *MockTaskflowService) StartAutonomousTask(ctx context.Context, request domain.AutonomousTaskRequest) (*domain.AutonomousTaskResponse, error) {
	return &domain.AutonomousTaskResponse{}, nil
}

func (m *MockTaskflowService) CancelAutonomousTask(ctx context.Context, taskId string) error {
	return nil
}

func (m *MockTaskflowService) GetAutonomousTaskStatus(ctx context.Context, taskId string) (*domain.AutonomousTaskStatus, error) {
	return &domain.AutonomousTaskStatus{}, nil
}

func (m *MockTaskflowService) ListAutonomousTasks(ctx context.Context, projectPath string) ([]domain.AutonomousTask, error) {
	return []domain.AutonomousTask{}, nil
}

func (m *MockTaskflowService) GetTaskLogs(ctx context.Context, taskId string) ([]domain.LogEntry, error) {
	return []domain.LogEntry{}, nil
}

func (m *MockTaskflowService) PauseTask(ctx context.Context, taskId string) error {
	return nil
}

func (m *MockTaskflowService) ResumeTask(ctx context.Context, taskId string) error {
	return nil
}

type MockFormatterService struct{}

func (m *MockFormatterService) FormatProject(ctx context.Context, projectPath, language string) (*domain.FormatResult, error) {
	return &domain.FormatResult{
		Language:   language,
		Success:    true,
		FilesCount: 5,
	}, nil
}

type MockFileSystemWriter struct{}

func (m *MockFileSystemWriter) WriteFile(filename string, data []byte, perm int) error {
	return nil
}

func (m *MockFileSystemWriter) MkdirAll(path string, perm int) error {
	return nil
}

func (m *MockFileSystemWriter) Remove(name string) error {
	return nil
}

func (m *MockFileSystemWriter) RemoveAll(path string) error {
	return nil
}

type MockCommandRunner struct{}

func (m *MockCommandRunner) RunCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return []byte("mock output"), nil
}

func (m *MockCommandRunner) RunCommandInDir(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	return []byte("mock output"), nil
}
