package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"shotgun_code/domain"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// TaskflowService implements domain.TaskflowService for workflow management
type TaskflowService struct {
	log        domain.Logger
	config     domain.TaskflowConfig
	tasks      map[string]domain.Task
	statuses   map[string]*domain.TaskStatus
	mu         sync.RWMutex
	planPath   string
	statusPath string
	guardrails domain.GuardrailService
}

// NewTaskflowService creates a new taskflow service
func NewTaskflowService(log domain.Logger, guardrails domain.GuardrailService) *TaskflowService {
	service := &TaskflowService{
		log:        log,
		tasks:      make(map[string]domain.Task),
		statuses:   make(map[string]*domain.TaskStatus),
		planPath:   "tasks/plan.yaml",
		statusPath: "tasks/status.json",
		guardrails: guardrails,
		config: domain.TaskflowConfig{
			AutoStart:     true,
			MaxConcurrent: 3,
			RetryAttempts: 3,
			RetryDelay:    5 * time.Second,
			Timeout:       30 * time.Minute,
			EnableLogging: true,
			EnableMetrics: true,
		},
	}

	// Load tasks during initialization
	if _, err := service.LoadTasks(); err != nil {
		log.Warning(fmt.Sprintf("Failed to load tasks: %v", err))
	}

	return service
}

// LoadTasks loads tasks from plan.yaml
func (s *TaskflowService) LoadTasks() ([]domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read plan.yaml
	data, err := os.ReadFile(s.planPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	// Parse YAML
	var plan struct {
		Version int `yaml:"version"`
		Tasks   []struct {
			ID        string   `yaml:"id"`
			Name      string   `yaml:"name"`
			DependsOn []string `yaml:"dependsOn"`
			StepFile  string   `yaml:"stepFile"`
			Budgets   struct {
				MaxFiles        int `yaml:"maxFiles"`
				MaxChangedLines int `yaml:"maxChangedLines"`
			} `yaml:"budgets"`
			Status string `yaml:"status"`
		} `yaml:"tasks"`
	}

	if err := yaml.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan file: %w", err)
	}

	// Load current statuses
	statuses, err := s.loadStatuses()
	if err != nil {
		s.log.Warning(fmt.Sprintf("Failed to load statuses: %v", err))
	}

	// Create tasks
	var tasks []domain.Task
	for _, taskData := range plan.Tasks {
		state := domain.TaskStateTodo
		if status, exists := statuses[taskData.ID]; exists {
			state = status
		}

		task := domain.Task{
			ID:        taskData.ID,
			Name:      taskData.Name,
			State:     state,
			DependsOn: taskData.DependsOn,
			StepFile:  taskData.StepFile,
			Budgets: domain.TaskBudgets{
				MaxFiles:        taskData.Budgets.MaxFiles,
				MaxChangedLines: taskData.Budgets.MaxChangedLines,
			},
			Status:    taskData.Status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Metadata:  make(map[string]interface{}),
		}

		s.tasks[task.ID] = task
		tasks = append(tasks, task)
	}

	s.log.Info(fmt.Sprintf("Loaded %d tasks from plan", len(tasks)))
	return tasks, nil
}

// GetTaskStatus returns task status
func (s *TaskflowService) GetTaskStatus(taskID string) (*domain.TaskStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.statuses[taskID]
	if !exists {
		return nil, fmt.Errorf("task status not found: %s", taskID)
	}

	return status, nil
}

// UpdateTaskStatus updates task status
func (s *TaskflowService) UpdateTaskStatus(taskID string, state domain.TaskState, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update status
	status, exists := s.statuses[taskID]
	if !exists {
		status = &domain.TaskStatus{
			TaskID: taskID,
		}
		s.statuses[taskID] = status
	}

	status.State = state
	status.Message = message

	// Update start/completion time
	switch state {
	case domain.TaskStateDone, domain.TaskStateFailed, domain.TaskStateBlocked:
		if status.CompletedAt == nil {
			now := time.Now()
			status.CompletedAt = &now
			if status.StartedAt != nil {
				status.Duration = now.Sub(*status.StartedAt)
			}
		}
	}

	// Save to file
	if err := s.saveStatuses(); err != nil {
		return fmt.Errorf("failed to save statuses: %w", err)
	}

	s.log.Info(fmt.Sprintf("Updated task %s status to %s: %s", taskID, state, message))
	return nil
}

// ExecuteTask executes a task
func (s *TaskflowService) ExecuteTask(taskID string) error {
	s.mu.Lock()
	task, exists := s.tasks[taskID]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("task not found: %s", taskID)
	}
	s.mu.Unlock()

	// Check dependencies
	deps, err := s.GetTaskDependencies(taskID)
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %w", err)
	}

	for _, dep := range deps {
		if dep.State != domain.TaskStateDone {
			return fmt.Errorf("dependency %s is not completed (state: %s)", dep.ID, dep.State)
		}
	}

	// Update status to running
	if err := s.UpdateTaskStatus(taskID, domain.TaskStateRunning, "Task started"); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Execute task logic (simplified for migration)
	s.log.Info(fmt.Sprintf("Executing task: %s", task.Name))

	// Update status to done
	if err := s.UpdateTaskStatus(taskID, domain.TaskStateDone, "Task completed successfully"); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// GetTaskDependencies returns task dependencies
func (s *TaskflowService) GetTaskDependencies(taskID string) ([]domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	var dependencies []domain.Task
	for _, depID := range task.DependsOn {
		if dep, exists := s.tasks[depID]; exists {
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, nil
}

// loadStatuses loads task statuses from file
func (s *TaskflowService) loadStatuses() (map[string]domain.TaskState, error) {
	data, err := os.ReadFile(s.statusPath)
	if err != nil {
		return make(map[string]domain.TaskState), nil // Return empty map if file doesn't exist
	}

	var statuses map[string]domain.TaskState
	if err := json.Unmarshal(data, &statuses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal statuses: %w", err)
	}

	return statuses, nil
}

// saveStatuses saves task statuses to file
func (s *TaskflowService) saveStatuses() error {
	statuses := make(map[string]domain.TaskState)
	for id, status := range s.statuses {
		statuses[id] = status.State
	}

	data, err := json.Marshal(statuses)
	if err != nil {
		return fmt.Errorf("failed to marshal statuses: %w", err)
	}

	return os.WriteFile(s.statusPath, data, 0644)
}
