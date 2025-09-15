package application

import (
	"context"
	"fmt"
	"os"
	"shotgun_code/domain"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// TaskflowServiceImpl реализует TaskflowService и TaskTypeProvider
type TaskflowServiceImpl struct {
	log        domain.Logger
	config     domain.TaskflowConfig
	tasks      map[string]domain.Task
	statuses   map[string]*domain.TaskStatus
	mu         sync.RWMutex
	planPath   string
	statusPath string
	planner    *RouterPlannerService
	guardrails domain.GuardrailService
	repo       domain.TaskflowRepository
}

// NewTaskflowService создает новый сервис taskflow
func NewTaskflowService(log domain.Logger, planner *RouterPlannerService, guardrails domain.GuardrailService, repo domain.TaskflowRepository) domain.TaskflowService {
	service := &TaskflowServiceImpl{
		log:        log,
		tasks:      make(map[string]domain.Task),
		statuses:   make(map[string]*domain.TaskStatus),
		planPath:   "tasks/plan.yaml",
		statusPath: "tasks/status.json",
		planner:    planner,
		guardrails: guardrails,
		repo:       repo,
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

	// Загружаем задачи при инициализации
	if _, err := service.LoadTasks(); err != nil {
		log.Warning(fmt.Sprintf("Failed to load tasks: %v", err))
	}

	return service
}

// GetTaskType возвращает тип задачи по ID (реализация TaskTypeProvider)
func (s *TaskflowServiceImpl) GetTaskType(taskID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return "", fmt.Errorf("task not found: %s", taskID)
	}

	// Определяем тип задачи по ID или другим характеристикам
	taskType := "regular"
	if strings.Contains(task.ID, "scaffold") {
		taskType = "scaffold"
	} else if strings.Contains(task.ID, "deps_fix") {
		taskType = "deps_fix"
	}

	return taskType, nil
}

// LoadTasks загружает задачи из plan.yaml
func (s *TaskflowServiceImpl) LoadTasks() ([]domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Читаем plan.yaml
	data, err := os.ReadFile(s.planPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	// Парсим YAML
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

	// Загружаем текущие статусы
	statuses, err := s.loadStatuses()
	if err != nil {
		s.log.Warning(fmt.Sprintf("Failed to load statuses: %v", err))
	}

	// Создаем задачи
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

// GetTaskStatus возвращает статус задачи
func (s *TaskflowServiceImpl) GetTaskStatus(taskID string) (*domain.TaskStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.statuses[taskID]
	if !exists {
		return nil, fmt.Errorf("task status not found: %s", taskID)
	}

	return status, nil
}

// UpdateTaskStatus обновляет статус задачи
func (s *TaskflowServiceImpl) UpdateTaskStatus(taskID string, state domain.TaskState, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Обновляем статус
	status, exists := s.statuses[taskID]
	if !exists {
		status = &domain.TaskStatus{
			TaskID: taskID,
		}
		s.statuses[taskID] = status
	}

	status.State = state
	status.Message = message

	// Обновляем время начала/завершения
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

	// Сохраняем в файл
	if err := s.saveStatuses(); err != nil {
		return fmt.Errorf("failed to save statuses: %w", err)
	}

	s.log.Info(fmt.Sprintf("Updated task %s status to %s: %s", taskID, state, message))
	return nil
}

// ExecuteTask выполняет задачу
func (s *TaskflowServiceImpl) ExecuteTask(taskID string) error {
	s.mu.Lock()
	task, exists := s.tasks[taskID]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("task not found: %s", taskID)
	}
	s.mu.Unlock()

	// Проверяем зависимости
	deps, err := s.GetTaskDependencies(taskID)
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %w", err)
	}

	for _, dep := range deps {
		if dep.State != domain.TaskStateDone {
			return fmt.Errorf("dependency %s is not completed (state: %s)", dep.ID, dep.State)
		}
	}

	// Проверяем guardrails перед выполнением
	if s.guardrails != nil {
		// Включаем ephemeral mode для scaffold/deps_fix задач
		taskType := "regular"
		if strings.Contains(taskID, "scaffold") {
			taskType = "scaffold"
		} else if strings.Contains(taskID, "deps_fix") {
			taskType = "deps_fix"
		}
		
		if taskType == "scaffold" || taskType == "deps_fix" {
			if err := s.guardrails.EnableEphemeralMode(taskID, taskType, 5*time.Minute); err != nil {
				s.log.Warning(fmt.Sprintf("Failed to enable ephemeral mode for task %s: %v", taskID, err))
			}
		}

		// Валидируем задачу (пока без файлов, так как они еще не изменены)
		validationResult, err := s.guardrails.ValidateTask(taskID, []string{}, 0)
		if err != nil {
			s.log.Error(fmt.Sprintf("Guardrail validation failed for task %s: %v", taskID, err))
			return fmt.Errorf("guardrail validation failed: %w", err)
		}

		if !validationResult.Valid {
			s.log.Error(fmt.Sprintf("Task %s failed guardrail validation: %s", taskID, validationResult.Error))
			return fmt.Errorf("task validation failed: %s", validationResult.Error)
		}
	}

	// Обновляем статус на выполнение
	now := time.Now()
	status := &domain.TaskStatus{
		TaskID:    taskID,
		State:     domain.TaskStateTodo,
		Progress:  0.0,
		Message:   "Starting task execution",
		StartedAt: &now,
	}

	s.mu.Lock()
	s.statuses[taskID] = status
	s.mu.Unlock()

	// Создаем и выполняем пайплайн через планировщик
	s.log.Info(fmt.Sprintf("Creating pipeline for task: %s", taskID))

	pipeline, err := s.planner.CreatePipeline(context.Background(), task)
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	// Выполняем пайплайн
	s.log.Info(fmt.Sprintf("Executing pipeline for task: %s", taskID))
	if err := s.planner.ExecutePipeline(context.Background(), pipeline); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	// Получаем статус пайплайна
	pipelineStatus := s.planner.GetPipelineStatus(pipeline)

	// Обновляем прогресс и статус задачи
	progress := pipelineStatus["progress"].(float64)
	status.Progress = progress
	status.Message = fmt.Sprintf("Pipeline completed with %d steps", len(pipeline.Steps))

	if pipeline.Status == PipelineStatusCompleted {
		status.State = domain.TaskStateDone
		status.Message = "Task completed successfully via pipeline"
	} else {
		status.State = domain.TaskStateFailed
		status.Message = fmt.Sprintf("Pipeline failed: %s", pipeline.Error)
	}

	// Отключаем ephemeral mode после завершения
	if s.guardrails != nil && (strings.Contains(taskID, "scaffold") || strings.Contains(taskID, "deps_fix")) {
		s.guardrails.DisableEphemeralMode()
	}

	// Сохраняем статус
	return s.UpdateTaskStatus(taskID, status.State, status.Message)
}

// ExecuteTaskflow выполняет весь taskflow
func (s *TaskflowServiceImpl) ExecuteTaskflow() error {
	s.log.Info("Starting taskflow execution")

	for {
		// Получаем готовые задачи
		readyTasks, err := s.GetReadyTasks()
		if err != nil {
			return fmt.Errorf("failed to get ready tasks: %w", err)
		}

		if len(readyTasks) == 0 {
			s.log.Info("No more tasks to execute")
			break
		}

		// Выполняем задачи
		for _, task := range readyTasks {
			if err := s.ExecuteTask(task.ID); err != nil {
				s.log.Error(fmt.Sprintf("Failed to execute task %s: %v", task.ID, err))
				if err := s.UpdateTaskStatus(task.ID, domain.TaskStateFailed, err.Error()); err != nil {
					s.log.Error(fmt.Sprintf("Failed to update task status: %v", err))
				}
			}
		}
	}

	s.log.Info("Taskflow execution completed")
	return nil
}

// GetReadyTasks возвращает готовые к выполнению задачи
func (s *TaskflowServiceImpl) GetReadyTasks() ([]domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var readyTasks []domain.Task

	for _, task := range s.tasks {
		if task.State != domain.TaskStateTodo {
			continue
		}

		// Проверяем зависимости
		ready := true
		for _, depID := range task.DependsOn {
			dep, exists := s.tasks[depID]
			if !exists {
				ready = false
				break
			}
			if dep.State != domain.TaskStateDone {
				ready = false
				break
			}
		}

		if ready {
			readyTasks = append(readyTasks, task)
		}
	}

	return readyTasks, nil
}

// GetTaskDependencies возвращает зависимости задачи
func (s *TaskflowServiceImpl) GetTaskDependencies(taskID string) ([]domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	var deps []domain.Task
	for _, depID := range task.DependsOn {
		if dep, exists := s.tasks[depID]; exists {
			deps = append(deps, dep)
		}
	}

	return deps, nil
}

// ValidateTaskflow проверяет корректность taskflow
func (s *TaskflowServiceImpl) ValidateTaskflow() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Проверяем циклические зависимости
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for taskID := range s.tasks {
		if !visited[taskID] {
			if s.hasCycle(taskID, visited, recStack) {
				return fmt.Errorf("circular dependency detected")
			}
		}
	}

	// Проверяем существование файлов
	for _, task := range s.tasks {
		if task.StepFile != "" {
			if _, err := os.Stat(task.StepFile); os.IsNotExist(err) {
				return fmt.Errorf("step file not found: %s", task.StepFile)
			}
		}
	}

	return nil
}

// GetTaskflowProgress возвращает прогресс выполнения
func (s *TaskflowServiceImpl) GetTaskflowProgress() (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.tasks) == 0 {
		return 0.0, nil
	}

	completed := 0
	for _, task := range s.tasks {
		if task.State == domain.TaskStateDone {
			completed++
		}
	}

	return float64(completed) / float64(len(s.tasks)), nil
}

// ResetTaskflow сбрасывает taskflow
func (s *TaskflowServiceImpl) ResetTaskflow() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Сбрасываем все статусы
	for taskID := range s.tasks {
		s.statuses[taskID] = &domain.TaskStatus{
			TaskID: taskID,
			State:  domain.TaskStateTodo,
		}
	}

	// Сохраняем в файл
	if err := s.saveStatuses(); err != nil {
		return fmt.Errorf("failed to save statuses: %w", err)
	}

	s.log.Info("Taskflow reset completed")
	return nil
}

// StartAutonomousTask запускает автономную задачу
func (s *TaskflowServiceImpl) StartAutonomousTask(ctx context.Context, request domain.AutonomousTaskRequest) (*domain.AutonomousTaskResponse, error) {
	s.log.Info(fmt.Sprintf("Starting autonomous task: %s", request.Task))

	// Input validation with structured errors
	if err := s.validateAutonomousTaskRequest(request); err != nil {
		return nil, domain.NewValidationError("Invalid autonomous task request", map[string]interface{}{
			"task":        request.Task,
			"projectPath": request.ProjectPath,
			"slaPolicy":   request.SlaPolicy,
		})
	}

	// Генерируем уникальный ID задачи
	taskID := fmt.Sprintf("autonomous_%d", time.Now().Unix())

	// Check for existing running tasks
	if s.hasRunningTasks() {
		return nil, domain.NewInvalidTaskStateError(taskID, "new", "no_running_tasks")
	}

	// Создаем статус задачи
	status := &domain.AutonomousTaskStatus{
		TaskId:                 taskID,
		Status:                 "pending",
		CurrentStep:            "initializing",
		Progress:               0.0,
		EstimatedTimeRemaining: 300, // 5 минут по умолчанию
		StartedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	// Attempt to create task with error recovery
	if err := s.createTaskStatus(taskID, request); err != nil {
		return nil, domain.NewInternalError("Failed to create task status", err)
	}

	// Safe goroutine execution with panic recovery
	go s.safeExecuteAutonomousTask(ctx, request, status)

	response := &domain.AutonomousTaskResponse{
		TaskId:  taskID,
		Status:  "accepted",
		Message: "Task accepted for processing",
	}

	return response, nil
}

// CancelAutonomousTask отменяет автономную задачу
func (s *TaskflowServiceImpl) CancelAutonomousTask(ctx context.Context, taskId string) error {
	s.log.Info(fmt.Sprintf("Cancelling autonomous task: %s", taskId))

	s.mu.Lock()
	defer s.mu.Unlock()

	status, exists := s.statuses[taskId]
	if !exists {
		return domain.NewTaskNotFoundError(taskId)
	}

	// Check if task can be cancelled
	if status.State == domain.TaskStateDone {
		return domain.NewInvalidTaskStateError(taskId, string(status.State), "cancellable")
	}

	// Обновляем статус на отмененный
	status.State = domain.TaskStateFailed
	status.Message = "Task cancelled by user"

	// Сохраняем статус
	if err := s.saveStatuses(); err != nil {
		return domain.NewInternalError("Failed to save task status after cancellation", err)
	}

	s.log.Info(fmt.Sprintf("Task %s cancelled successfully", taskId))
	return nil
}

// GetAutonomousTaskStatus получает статус автономной задачи
func (s *TaskflowServiceImpl) GetAutonomousTaskStatus(ctx context.Context, taskId string) (*domain.AutonomousTaskStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.statuses[taskId]
	if !exists {
		return nil, domain.NewTaskNotFoundError(taskId)
	}

	// Конвертируем в AutonomousTaskStatus
	autonomousStatus := &domain.AutonomousTaskStatus{
		TaskId:                 taskId,
		Status:                 string(status.State),
		CurrentStep:            status.Message,
		Progress:               status.Progress * 100, // Конвертируем в проценты
		EstimatedTimeRemaining: 0,                     // TODO: Реализовать расчет
		StartedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Error:                  status.Error,
	}

	if status.StartedAt != nil {
		autonomousStatus.StartedAt = *status.StartedAt
	}

	return autonomousStatus, nil
}

// ListAutonomousTasks возвращает список автономных задач
func (s *TaskflowServiceImpl) ListAutonomousTasks(ctx context.Context, projectPath string) ([]domain.AutonomousTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var autonomousTasks []domain.AutonomousTask

	// Фильтруем задачи по проектному пути и выбираем только автономные
	for taskID, status := range s.statuses {
		if strings.HasPrefix(taskID, "autonomous_") {
			// Создаем AutonomousTask из статуса
			task := domain.AutonomousTask{
				ID:          taskID,
				Name:        "Autonomous Task",
				Description: status.Message,
				Status:      string(status.State),
				ProjectPath: projectPath,
				Progress:    status.Progress,
				CreatedAt:   time.Now(), // TODO: Store actual creation time
				UpdatedAt:   time.Now(),
				Error:       status.Error,
			}

			if status.StartedAt != nil {
				task.CreatedAt = *status.StartedAt
			}

			if status.CompletedAt != nil {
				task.CompletedAt = status.CompletedAt
			}

			autonomousTasks = append(autonomousTasks, task)
		}
	}

	s.log.Info(fmt.Sprintf("Found %d autonomous tasks for project %s", len(autonomousTasks), projectPath))
	return autonomousTasks, nil
}

// GetTaskLogs возвращает логи для задачи
func (s *TaskflowServiceImpl) GetTaskLogs(ctx context.Context, taskId string) ([]domain.LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Проверяем существование задачи
	status, exists := s.statuses[taskId]
	if !exists {
		return nil, domain.NewTaskNotFoundError(taskId)
	}

	// Создаем базовые логи на основе статуса задачи
	var logs []domain.LogEntry

	// Лог создания задачи
	if status.StartedAt != nil {
		logs = append(logs, domain.LogEntry{
			ID:        fmt.Sprintf("%s-created", taskId),
			TaskID:    taskId,
			Level:     "INFO",
			Message:   "Task created",
			Timestamp: *status.StartedAt,
			Metadata: map[string]interface{}{
				"event": "task_created",
			},
		})
	}

	// Лог текущего статуса
	logs = append(logs, domain.LogEntry{
		ID:        fmt.Sprintf("%s-status", taskId),
		TaskID:    taskId,
		Level:     "INFO",
		Message:   fmt.Sprintf("Task status: %s - %s", status.State, status.Message),
		Timestamp: status.UpdatedAt,
		Metadata: map[string]interface{}{
			"state":    string(status.State),
			"progress": status.Progress,
		},
	})

	// Лог ошибки если есть
	if status.Error != "" {
		logs = append(logs, domain.LogEntry{
			ID:        fmt.Sprintf("%s-error", taskId),
			TaskID:    taskId,
			Level:     "ERROR",
			Message:   status.Error,
			Timestamp: status.UpdatedAt,
			Metadata: map[string]interface{}{
				"event": "task_error",
			},
		})
	}

	// Лог завершения если задача завершена
	if status.CompletedAt != nil {
		logs = append(logs, domain.LogEntry{
			ID:        fmt.Sprintf("%s-completed", taskId),
			TaskID:    taskId,
			Level:     "INFO",
			Message:   "Task completed",
			Timestamp: *status.CompletedAt,
			Metadata: map[string]interface{}{
				"event":    "task_completed",
				"duration": status.Duration.String(),
			},
		})
	}

	s.log.Debug(fmt.Sprintf("Retrieved %d log entries for task %s", len(logs), taskId))
	return logs, nil
}

// PauseTask приостанавливает выполнение задачи
func (s *TaskflowServiceImpl) PauseTask(ctx context.Context, taskId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	status, exists := s.statuses[taskId]
	if !exists {
		return domain.NewTaskNotFoundError(taskId)
	}

	// Проверяем, можно ли приостановить задачу
	if status.State != domain.TaskStateTodo {
		return domain.NewInvalidTaskStateError(taskId, string(status.State), "running")
	}

	// Обновляем статус на заблокированный (используем как паузу)
	status.State = domain.TaskStateBlocked
	status.Message = "Task paused by user"
	status.UpdatedAt = time.Now()

	// Сохраняем статус
	if err := s.saveStatuses(); err != nil {
		return domain.NewInternalError("Failed to save task status after pause", err)
	}

	s.log.Info(fmt.Sprintf("Task %s paused successfully", taskId))
	return nil
}

// ResumeTask возобновляет выполнение приостановленной задачи
func (s *TaskflowServiceImpl) ResumeTask(ctx context.Context, taskId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	status, exists := s.statuses[taskId]
	if !exists {
		return domain.NewTaskNotFoundError(taskId)
	}

	// Проверяем, можно ли возобновить задачу
	if status.State != domain.TaskStateBlocked {
		return domain.NewInvalidTaskStateError(taskId, string(status.State), "paused")
	}

	// Обновляем статус обратно на выполнение
	status.State = domain.TaskStateTodo
	status.Message = "Task resumed by user"
	status.UpdatedAt = time.Now()

	// Сохраняем статус
	if err := s.saveStatuses(); err != nil {
		return domain.NewInternalError("Failed to save task status after resume", err)
	}

	s.log.Info(fmt.Sprintf("Task %s resumed successfully", taskId))
	return nil
}

// executeAutonomousTask выполняет автономную задачу
func (s *TaskflowServiceImpl) executeAutonomousTask(ctx context.Context, request domain.AutonomousTaskRequest, status *domain.AutonomousTaskStatus) error {
	// Обновляем статус на выполнение
	s.updateAutonomousTaskStatus(status.TaskId, "running", "Starting task execution", 10.0)

	// Здесь должна быть логика выполнения автономной задачи
	// Пока что просто симулируем выполнение
	time.Sleep(2 * time.Second)
	s.updateAutonomousTaskStatus(status.TaskId, "running", "Analyzing project structure", 30.0)

	time.Sleep(2 * time.Second)
	s.updateAutonomousTaskStatus(status.TaskId, "running", "Generating code changes", 60.0)

	time.Sleep(2 * time.Second)
	s.updateAutonomousTaskStatus(status.TaskId, "running", "Applying changes", 80.0)

	time.Sleep(1 * time.Second)
	s.updateAutonomousTaskStatus(status.TaskId, "completed", "Task completed successfully", 100.0)
	
	return nil
}

// safeExecuteAutonomousTask executes autonomous task with comprehensive error recovery
func (s *TaskflowServiceImpl) safeExecuteAutonomousTask(ctx context.Context, request domain.AutonomousTaskRequest, status *domain.AutonomousTaskStatus) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Error(fmt.Sprintf("PANIC in autonomous task execution: %v", r))
			s.updateAutonomousTaskStatus(status.TaskId, "failed", 
				fmt.Sprintf("Task execution panicked: %v", r), 100.0)
			
			// Send panic notification to frontend
			s.notifyTaskFailure(status.TaskId, fmt.Sprintf("Internal error: %v", r))
		}
	}()

	if err := s.executeAutonomousTask(ctx, request, status); err != nil {
		s.log.Error(fmt.Sprintf("Autonomous task execution failed: %v", err))
		s.updateAutonomousTaskStatus(status.TaskId, "failed", err.Error(), 100.0)
		s.notifyTaskFailure(status.TaskId, err.Error())
	}
}

// validateAutonomousTaskRequest validates the autonomous task request
func (s *TaskflowServiceImpl) validateAutonomousTaskRequest(request domain.AutonomousTaskRequest) error {
	if request.Task == "" {
		return fmt.Errorf("task description cannot be empty")
	}
	
	if request.ProjectPath == "" {
		return fmt.Errorf("project path cannot be empty")
	}
	
	if request.SlaPolicy == "" {
		return fmt.Errorf("SLA policy cannot be empty")
	}
	
	validSLAPolicies := []string{"lite", "standard", "strict"}
	validSLA := false
	for _, policy := range validSLAPolicies {
		if request.SlaPolicy == policy {
			validSLA = true
			break
		}
	}
	
	if !validSLA {
		return fmt.Errorf("invalid SLA policy: %s, must be one of: %v", request.SlaPolicy, validSLAPolicies)
	}
	
	return nil
}

// hasRunningTasks checks if there are any currently running tasks
func (s *TaskflowServiceImpl) hasRunningTasks() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, status := range s.statuses {
		if status.State == domain.TaskStateTodo {
			return true
		}
	}
	return false
}

// createTaskStatus creates and stores a new task status
func (s *TaskflowServiceImpl) createTaskStatus(taskID string, request domain.AutonomousTaskRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.statuses[taskID] = &domain.TaskStatus{
		TaskID: taskID,
		State:  domain.TaskStateTodo,
	}
	
	return s.saveStatuses()
}

// notifyTaskFailure sends failure notification to frontend
func (s *TaskflowServiceImpl) notifyTaskFailure(taskId, message string) {
	// This could emit an event to the frontend via the event bus
	s.log.Error(fmt.Sprintf("Task %s failed: %s", taskId, message))
}

// updateAutonomousTaskStatus обновляет статус автономной задачи
func (s *TaskflowServiceImpl) updateAutonomousTaskStatus(taskId, status, message string, progress float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskStatus, exists := s.statuses[taskId]
	if !exists {
		return
	}

	// Обновляем статус
	switch status {
	case "completed":
		taskStatus.State = domain.TaskStateDone
	case "failed":
		taskStatus.State = domain.TaskStateFailed
	default:
		taskStatus.State = domain.TaskStateTodo
	}

	taskStatus.Message = message
	taskStatus.Progress = progress / 100.0 // Конвертируем обратно в 0-1
	taskStatus.UpdatedAt = time.Now()

	// Сохраняем статус
	if err := s.saveStatuses(); err != nil {
		s.log.Error(fmt.Sprintf("Failed to save task status: %v", err))
	}
}

// loadStatuses загружает статусы из репозитория
func (s *TaskflowServiceImpl) loadStatuses() (map[string]domain.TaskState, error) {
	return s.repo.LoadStatuses()
}

// saveStatuses сохраняет статусы в репозиторий
func (s *TaskflowServiceImpl) saveStatuses() error {
	statuses := make(map[string]domain.TaskState)
	for taskID, status := range s.statuses {
		statuses[taskID] = status.State
	}
	return s.repo.SaveStatuses(statuses)
}

// hasCycle проверяет наличие циклических зависимостей
func (s *TaskflowServiceImpl) hasCycle(taskID string, visited, recStack map[string]bool) bool {
	visited[taskID] = true
	recStack[taskID] = true

	task, exists := s.tasks[taskID]
	if !exists {
		return false
	}

	for _, depID := range task.DependsOn {
		if !visited[depID] {
			if s.hasCycle(depID, visited, recStack) {
				return true
			}
		} else if recStack[depID] {
			return true
		}
	}

	recStack[taskID] = false
	return false
}