package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"strings"
	"time"
)

// TaskflowProtocolIntegration integrates Task Protocol with Taskflow services
type TaskflowProtocolIntegration struct {
	log                 domain.Logger
	taskflowService     domain.TaskflowService
	taskProtocolService domain.TaskProtocolService
	configService       *TaskProtocolConfigService
	aiService           *IntelligentAIService
}

// NewTaskflowProtocolIntegration creates a new integration service
func NewTaskflowProtocolIntegration(
	log domain.Logger,
	taskflowService domain.TaskflowService,
	taskProtocolService domain.TaskProtocolService,
	configService *TaskProtocolConfigService,
	aiService *IntelligentAIService,
) *TaskflowProtocolIntegration {
	return &TaskflowProtocolIntegration{
		log:                 log,
		taskflowService:     taskflowService,
		taskProtocolService: taskProtocolService,
		configService:       configService,
		aiService:           aiService,
	}
}

// ExecuteTaskWithProtocol executes a taskflow task with protocol verification
func (t *TaskflowProtocolIntegration) ExecuteTaskWithProtocol(ctx context.Context, taskID string) (*domain.TaskProtocolResult, error) {
	t.log.Info(fmt.Sprintf("Executing task %s with protocol verification", taskID))

	// Get task details from taskflow
	task, err := t.getTaskDetails(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task details: %w", err)
	}

	// Update task status to in progress
	if err := t.taskflowService.UpdateTaskStatus(taskID, domain.TaskStateRunning, "Starting protocol verification"); err != nil {
		t.log.Warning(fmt.Sprintf("Failed to update task status: %v", err))
	}

	// Create protocol configuration for the task
	protocolConfig, err := t.createProtocolConfigForTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to create protocol config: %w", err)
	}

	// Execute the task protocol
	result, err := t.taskProtocolService.ExecuteProtocol(ctx, protocolConfig)
	if err != nil {
		// Update task status to failed
		t.taskflowService.UpdateTaskStatus(taskID, domain.TaskStateFailed, fmt.Sprintf("Protocol verification failed: %v", err))
		return result, fmt.Errorf("protocol execution failed: %w", err)
	}

	// Update task status based on protocol result
	if result.Success {
		if err := t.taskflowService.UpdateTaskStatus(taskID, domain.TaskStateDone, "Protocol verification completed successfully"); err != nil {
			t.log.Warning(fmt.Sprintf("Failed to update task status: %v", err))
		}
	} else {
		if err := t.taskflowService.UpdateTaskStatus(taskID, domain.TaskStateFailed, fmt.Sprintf("Protocol verification failed: %s", result.FinalError)); err != nil {
			t.log.Warning(fmt.Sprintf("Failed to update task status: %v", err))
		}
	}

	// Store protocol result metadata in task
	t.storeProtocolResultInTask(taskID, result)

	t.log.Info(fmt.Sprintf("Task %s protocol execution completed with success: %t", taskID, result.Success))
	return result, nil
}

// ExecuteTaskflowWithProtocol executes an entire taskflow with protocol verification
func (t *TaskflowProtocolIntegration) ExecuteTaskflowWithProtocol(ctx context.Context, options *TaskflowProtocolOptions) (*TaskflowProtocolResult, error) {
	t.log.Info("Executing taskflow with protocol verification")

	result := &TaskflowProtocolResult{
		StartedAt:   time.Now(),
		TaskResults: make(map[string]*domain.TaskProtocolResult),
		Success:     true,
	}

	// Load tasks from taskflow
	tasks, err := t.taskflowService.LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Execute tasks in dependency order
	for _, task := range tasks {
		if t.shouldExecuteTask(task, options) {
			t.log.Info(fmt.Sprintf("Executing task: %s", task.ID))

			taskResult, err := t.ExecuteTaskWithProtocol(ctx, task.ID)
			if err != nil {
				t.log.Error(fmt.Sprintf("Task %s failed: %v", task.ID, err))
				result.Success = false
				result.FailedTasks = append(result.FailedTasks, task.ID)

				if options.FailFast {
					result.CompletedAt = time.Now()
					return result, fmt.Errorf("taskflow execution failed at task %s: %w", task.ID, err)
				}
			} else {
				result.TaskResults[task.ID] = taskResult
				if !taskResult.Success {
					result.Success = false
					result.FailedTasks = append(result.FailedTasks, task.ID)
				}
			}
		}
	}

	result.CompletedAt = time.Now()
	t.log.Info(fmt.Sprintf("Taskflow execution completed with success: %t", result.Success))
	return result, nil
}

// CreateTaskProtocolForAIGeneration creates a protocol for AI-generated code validation
func (t *TaskflowProtocolIntegration) CreateTaskProtocolForAIGeneration(ctx context.Context, aiRequest *AICodeGenerationRequest) (*domain.TaskProtocolConfig, error) {
	t.log.Info("Creating task protocol for AI code generation")

	// Detect languages from the AI request context
	languages, err := t.detectLanguagesFromContext(aiRequest.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to detect languages: %w", err)
	}

	// Create base protocol configuration
	config, err := t.configService.GetConfigurationForProject(ctx, aiRequest.ProjectPath, languages)
	if err != nil {
		return nil, fmt.Errorf("failed to get protocol configuration: %w", err)
	}

	// Customize configuration for AI generation
	config = t.customizeConfigForAIGeneration(config, aiRequest)

	return config, nil
}

// ValidateAIGeneratedCode validates AI-generated code using the task protocol
func (t *TaskflowProtocolIntegration) ValidateAIGeneratedCode(ctx context.Context, request *AICodeValidationRequest) (*domain.TaskProtocolResult, error) {
	t.log.Info("Validating AI-generated code with task protocol")

	// Create protocol configuration for validation
	protocolConfig, err := t.CreateTaskProtocolForAIGeneration(ctx, &AICodeGenerationRequest{
		ProjectPath: request.ProjectPath,
		Context:     request.Context,
		Languages:   request.Languages,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create protocol config: %w", err)
	}

	// Execute validation protocol
	result, err := t.taskProtocolService.ExecuteProtocol(ctx, protocolConfig)
	if err != nil {
		return result, fmt.Errorf("validation protocol failed: %w", err)
	}

	// If validation failed and AI correction is enabled, attempt correction
	if !result.Success && protocolConfig.SelfCorrection.Enabled && protocolConfig.SelfCorrection.AIAssistance {
		t.log.Info("Attempting AI-assisted correction")

		correctionResult, err := t.attemptAICorrection(ctx, result, request)
		if err != nil {
			t.log.Warning(fmt.Sprintf("AI correction failed: %v", err))
		} else if correctionResult != nil {
			result = correctionResult
		}
	}

	return result, nil
}

// Helper methods

func (t *TaskflowProtocolIntegration) getTaskDetails(taskID string) (*domain.Task, error) {
	// In a real implementation, this would retrieve task details from the taskflow service
	// For now, return a basic task structure
	return &domain.Task{
		ID:        taskID,
		Name:      fmt.Sprintf("Task %s", taskID),
		State:     domain.TaskStateTodo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}, nil
}

func (t *TaskflowProtocolIntegration) createProtocolConfigForTask(ctx context.Context, task *domain.Task) (*domain.TaskProtocolConfig, error) {
	// Extract project path from task metadata
	projectPath := "/default/project/path"
	if path, exists := task.Metadata["projectPath"]; exists {
		if pathStr, ok := path.(string); ok {
			projectPath = pathStr
		}
	}

	// Extract languages from task metadata
	languages := []string{"go"}
	if langs, exists := task.Metadata["languages"]; exists {
		if langSlice, ok := langs.([]string); ok {
			languages = langSlice
		}
	}

	// Get configuration for the project
	config, err := t.configService.GetConfigurationForProject(ctx, projectPath, languages)
	if err != nil {
		return nil, err
	}

	// Customize configuration based on task requirements
	if task.Budgets.MaxFiles > 0 {
		// Adjust configuration based on task budgets
		config.MaxRetries = 5 // Increase retries for complex tasks
	}

	return config, nil
}

func (t *TaskflowProtocolIntegration) shouldExecuteTask(task domain.Task, options *TaskflowProtocolOptions) bool {
	// Check if task should be executed based on options
	if options.TaskFilter != "" && task.ID != options.TaskFilter {
		return false
	}

	// Skip already completed tasks unless forced
	if task.State == domain.TaskStateDone && !options.ForceRerun {
		return false
	}

	return true
}

func (t *TaskflowProtocolIntegration) storeProtocolResultInTask(taskID string, result *domain.TaskProtocolResult) {
	// Store protocol result metadata in the task
	// In a real implementation, this would update the task in the taskflow repository
	metadata := map[string]interface{}{
		"protocolResult": map[string]interface{}{
			"success":          result.Success,
			"correctionCycles": result.CorrectionCycles,
			"stagesExecuted":   len(result.Stages),
			"executionTime":    result.CompletedAt.Sub(result.StartedAt).String(),
		},
	}

	t.log.Debug(fmt.Sprintf("Storing protocol result metadata for task %s: %+v", taskID, metadata))
}

func (t *TaskflowProtocolIntegration) detectLanguagesFromContext(context string) ([]string, error) {
	// Simple language detection based on context keywords
	languages := make([]string, 0)

	if containsGoKeywords(context) {
		languages = append(languages, "go")
	}
	if containsTypeScriptKeywords(context) {
		languages = append(languages, "typescript")
	}
	if containsJavaScriptKeywords(context) {
		languages = append(languages, "javascript")
	}

	if len(languages) == 0 {
		// Default to Go if no languages detected
		languages = append(languages, "go")
	}

	return languages, nil
}

func (t *TaskflowProtocolIntegration) customizeConfigForAIGeneration(config *domain.TaskProtocolConfig, request *AICodeGenerationRequest) *domain.TaskProtocolConfig {
	// Enable self-correction for AI-generated code
	config.SelfCorrection.Enabled = true
	config.SelfCorrection.AIAssistance = true
	config.SelfCorrection.MaxAttempts = 5

	// Increase retry attempts for AI generation
	config.MaxRetries = 3

	// Enable all verification stages for AI-generated code
	config.EnabledStages = []domain.ProtocolStage{
		domain.StageLinting,
		domain.StageBuilding,
		domain.StageTesting,
		domain.StageGuardrails,
	}

	return config
}

func (t *TaskflowProtocolIntegration) attemptAICorrection(ctx context.Context, result *domain.TaskProtocolResult, request *AICodeValidationRequest) (*domain.TaskProtocolResult, error) {
	// Find the first failed stage for correction
	var failedStage *domain.ProtocolStageResult
	for _, stage := range result.Stages {
		if !stage.Success {
			failedStage = stage
			break
		}
	}

	if failedStage == nil || failedStage.ErrorDetails == nil {
		return nil, fmt.Errorf("no failed stage with error details found")
	}

	// Request AI guidance for correction
	taskContext := &domain.TaskContext{
		ProjectPath: request.ProjectPath,
		Languages:   request.Languages,
		Files:       request.ChangedFiles,
	}

	guidance, err := t.taskProtocolService.RequestCorrectionGuidance(ctx, failedStage.ErrorDetails, taskContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI correction guidance: %w", err)
	}

	t.log.Info(fmt.Sprintf("Received AI correction guidance: %s", guidance.Explanation))

	// In a real implementation, this would apply the AI suggestions and re-run the protocol
	// For now, we'll simulate a successful correction
	return result, nil
}

// Helper functions for language detection

func containsGoKeywords(context string) bool {
	goKeywords := []string{"package", "func", "import", "go.mod", "go.sum", ".go"}
	for _, keyword := range goKeywords {
		if strings.Contains(context, keyword) {
			return true
		}
	}
	return false
}

func containsTypeScriptKeywords(context string) bool {
	tsKeywords := []string{"interface", "type", "typescript", ".ts", ".tsx", "tsconfig.json"}
	for _, keyword := range tsKeywords {
		if strings.Contains(context, keyword) {
			return true
		}
	}
	return false
}

func containsJavaScriptKeywords(context string) bool {
	jsKeywords := []string{"javascript", ".js", ".jsx", "package.json", "node_modules"}
	for _, keyword := range jsKeywords {
		if strings.Contains(context, keyword) {
			return true
		}
	}
	return false
}

// Supporting types

type TaskflowProtocolOptions struct {
	TaskFilter string
	FailFast   bool
	ForceRerun bool
	Parallel   bool
}

type TaskflowProtocolResult struct {
	StartedAt   time.Time                             `json:"startedAt"`
	CompletedAt time.Time                             `json:"completedAt"`
	Success     bool                                  `json:"success"`
	TaskResults map[string]*domain.TaskProtocolResult `json:"taskResults"`
	FailedTasks []string                              `json:"failedTasks"`
}

type AICodeGenerationRequest struct {
	ProjectPath string   `json:"projectPath"`
	Context     string   `json:"context"`
	Languages   []string `json:"languages"`
	Task        string   `json:"task"`
}

type AICodeValidationRequest struct {
	ProjectPath  string   `json:"projectPath"`
	Context      string   `json:"context"`
	Languages    []string `json:"languages"`
	ChangedFiles []string `json:"changedFiles"`
}
