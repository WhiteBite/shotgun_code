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
	log              domain.Logger
	config           domain.TaskflowConfig
	tasks            map[string]domain.Task
	statuses         map[string]*domain.TaskStatus
	mu               sync.RWMutex
	planPath         string
	statusPath       string
	planner          *RouterPlannerService
	routerLlmService *RouterLLMService
	guardrails       domain.GuardrailService
	repo             domain.TaskflowRepository
	gitRepo          domain.GitRepository
}

// NewTaskflowService создает новый сервис taskflow
func NewTaskflowService(log domain.Logger, planner *RouterPlannerService, routerLlmService *RouterLLMService, guardrails domain.GuardrailService, repo domain.TaskflowRepository, gitRepo domain.GitRepository) domain.TaskflowService {
	service := &TaskflowServiceImpl{
		log:              log,
		tasks:            make(map[string]domain.Task),
		statuses:         make(map[string]*domain.TaskStatus),
		planPath:         "tasks/plan.yaml",
		statusPath:       "tasks/status.json",
		planner:          planner,
		routerLlmService: routerLlmService,
		guardrails:       guardrails,
		repo:             repo,
		gitRepo:          gitRepo,
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
	if strings.Contains(task.ID, taskTypeScaffold) {
		taskType = taskTypeScaffold
	} else if strings.Contains(task.ID, taskTypeDepsFix) {
		taskType = taskTypeDepsFix
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
	tasks := make([]domain.Task, 0, len(plan.Tasks))
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

// getTaskType determines task type from taskID
func getTaskType(taskID string) string {
	if strings.Contains(taskID, "scaffold") {
		return "scaffold"
	}
	if strings.Contains(taskID, "deps_fix") {
		return "deps_fix"
	}
	return "regular"
}

// checkDependencies verifies all dependencies are completed
func (s *TaskflowServiceImpl) checkDependencies(taskID string) error {
	deps, err := s.GetTaskDependencies(taskID)
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %w", err)
	}
	for _, dep := range deps {
		if dep.State != domain.TaskStateDone {
			return fmt.Errorf("dependency %s is not completed (state: %s)", dep.ID, dep.State)
		}
	}
	return nil
}

// validateWithGuardrails validates task with guardrails
func (s *TaskflowServiceImpl) validateWithGuardrails(taskID string) error {
	if s.guardrails == nil {
		return nil
	}

	taskType := getTaskType(taskID)
	if taskType == "scaffold" || taskType == "deps_fix" {
		if err := s.guardrails.EnableEphemeralMode(taskID, taskType, 5*time.Minute); err != nil {
			s.log.Warning(fmt.Sprintf("Failed to enable ephemeral mode for task %s: %v", taskID, err))
		}
	}

	validationResult, err := s.guardrails.ValidateTask(taskID, []string{}, 0)
	if err != nil {
		s.log.Error(fmt.Sprintf("Guardrail validation failed for task %s: %v", taskID, err))
		return fmt.Errorf("guardrail validation failed: %w", err)
	}
	if !validationResult.Valid {
		s.log.Error(fmt.Sprintf("Task %s failed guardrail validation: %s", taskID, validationResult.Error))
		return fmt.Errorf("task validation failed: %s", validationResult.Error)
	}
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

	if err := s.checkDependencies(taskID); err != nil {
		return err
	}
	if err := s.validateWithGuardrails(taskID); err != nil {
		return err
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

	pipeline, err := s.planner.CreatePipeline(context.Background(), task, nil)
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
func (s *TaskflowServiceImpl) CancelAutonomousTask(ctx context.Context, taskID string) error {
	s.log.Info(fmt.Sprintf("Cancelling autonomous task: %s", taskID))

	s.mu.Lock()
	defer s.mu.Unlock()

	status, exists := s.statuses[taskID]
	if !exists {
		return domain.NewTaskNotFoundError(taskID)
	}

	// Check if task can be cancelled
	if status.State == domain.TaskStateDone {
		return domain.NewInvalidTaskStateError(taskID, string(status.State), "cancellable")
	}

	// Обновляем статус на отмененный
	status.State = domain.TaskStateFailed
	status.Message = "Task cancelled by user"

	// Сохраняем статус
	if err := s.saveStatuses(); err != nil {
		return domain.NewInternalError("Failed to save task status after cancellation", err)
	}

	s.log.Info(fmt.Sprintf("Task %s cancelled successfully", taskID))
	return nil
}

// GetAutonomousTaskStatus получает статус автономной задачи
func (s *TaskflowServiceImpl) GetAutonomousTaskStatus(ctx context.Context, taskID string) (*domain.AutonomousTaskStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.statuses[taskID]
	if !exists {
		return nil, domain.NewTaskNotFoundError(taskID)
	}

	// Конвертируем в AutonomousTaskStatus
	autonomousStatus := &domain.AutonomousTaskStatus{
		TaskId:                 taskID,
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
func (s *TaskflowServiceImpl) GetTaskLogs(ctx context.Context, taskID string) ([]domain.LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Проверяем существование задачи
	status, exists := s.statuses[taskID]
	if !exists {
		return nil, domain.NewTaskNotFoundError(taskID)
	}

	// Создаем базовые логи на основе статуса задачи
	var logs []domain.LogEntry

	// Лог создания задачи
	if status.StartedAt != nil {
		logs = append(logs, domain.LogEntry{
			ID:        fmt.Sprintf("%s-created", taskID),
			TaskID:    taskID,
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
		ID:        fmt.Sprintf("%s-status", taskID),
		TaskID:    taskID,
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
			ID:        fmt.Sprintf("%s-error", taskID),
			TaskID:    taskID,
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
			ID:        fmt.Sprintf("%s-completed", taskID),
			TaskID:    taskID,
			Level:     "INFO",
			Message:   "Task completed",
			Timestamp: *status.CompletedAt,
			Metadata: map[string]interface{}{
				"event":    "task_completed",
				"duration": status.Duration.String(),
			},
		})
	}

	s.log.Debug(fmt.Sprintf("Retrieved %d log entries for task %s", len(logs), taskID))
	return logs, nil
}

// changeTaskState is a helper for PauseTask and ResumeTask
func (s *TaskflowServiceImpl) changeTaskState(taskID string, fromState, toState domain.TaskState, message, action string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	status, exists := s.statuses[taskID]
	if !exists {
		return domain.NewTaskNotFoundError(taskID)
	}

	if status.State != fromState {
		return domain.NewInvalidTaskStateError(taskID, string(status.State), string(fromState))
	}

	status.State = toState
	status.Message = message
	status.UpdatedAt = time.Now()

	if err := s.saveStatuses(); err != nil {
		return domain.NewInternalError("Failed to save task status after "+action, err)
	}

	s.log.Info(fmt.Sprintf("Task %s %s successfully", taskID, action))
	return nil
}

// PauseTask приостанавливает выполнение задачи
func (s *TaskflowServiceImpl) PauseTask(ctx context.Context, taskID string) error {
	return s.changeTaskState(taskID, domain.TaskStateTodo, domain.TaskStateBlocked, "Task paused by user", "paused")
}

// ResumeTask возобновляет выполнение приостановленной задачи
func (s *TaskflowServiceImpl) ResumeTask(ctx context.Context, taskID string) error {
	return s.changeTaskState(taskID, domain.TaskStateBlocked, domain.TaskStateTodo, "Task resumed by user", "resumed")
}

// planAutonomousTask creates execution plan for autonomous task
func (s *TaskflowServiceImpl) planAutonomousTask(ctx context.Context, request domain.AutonomousTaskRequest, status *domain.AutonomousTaskStatus) (*TaskPipeline, domain.Task, error) {
	s.updateAutonomousTaskStatus(status.TaskId, "running", "Planning task...", 10.0)
	s.log.Info(fmt.Sprintf("[Task %s] Generating execution plan for: %s", status.TaskId, request.Task))

	contextPack, err := s.buildContextForTask(ctx, request)
	if err != nil {
		s.log.Error(fmt.Sprintf("[Task %s] Failed to build context: %v", status.TaskId, err))
		return nil, domain.Task{}, fmt.Errorf("failed to build context for task: %w", err)
	}

	planningTask := domain.Task{
		ID: status.TaskId, Name: "Autonomous Planning Task",
		Metadata: map[string]interface{}{"original_request": request.Task, "sla_policy": request.SlaPolicy, "project_path": request.ProjectPath},
	}

	var policy *PipelinePolicy
	if llmResponse, err := s.routerLlmService.CreatePipelineWithLLM(ctx, planningTask, contextPack); err == nil && llmResponse != nil && !llmResponse.FallbackUsed {
		policy = llmResponse.Policy
		s.log.Info(fmt.Sprintf("[Task %s] Using LLM-defined policy.", status.TaskId))
	} else {
		s.log.Info(fmt.Sprintf("[Task %s] Using heuristic policy.", status.TaskId))
	}

	basePipeline, err := s.planner.CreatePipeline(ctx, planningTask, policy)
	if err != nil {
		s.log.Error(fmt.Sprintf("[Task %s] Failed to create execution plan: %v", status.TaskId, err))
		return nil, domain.Task{}, fmt.Errorf("failed to create execution plan: %w", err)
	}

	s.log.Info(fmt.Sprintf("[Task %s] Execution plan generated with %d steps.", status.TaskId, len(basePipeline.Steps)))
	s.updateAutonomousTaskStatus(status.TaskId, "running", "Execution plan created. Starting execution...", 20.0)
	return basePipeline, planningTask, nil
}

// finishAutonomousTask completes the task and generates report
func (s *TaskflowServiceImpl) finishAutonomousTask(request domain.AutonomousTaskRequest, status *domain.AutonomousTaskStatus) {
	s.updateAutonomousTaskStatus(status.TaskId, "running", "Generating final report...", 95.0)
	if diff, err := s.gitRepo.GenerateDiff(request.ProjectPath); err != nil {
		s.log.Error(fmt.Sprintf("[Task %s] Failed to generate git diff: %v", status.TaskId, err))
	} else {
		s.log.Info(fmt.Sprintf("[Task %s] Git Diff:\n%s", status.TaskId, diff))
	}
	s.updateAutonomousTaskStatus(status.TaskId, "completed", "Task completed successfully", 100.0)
	s.log.Info(fmt.Sprintf("[Task %s] Autonomous task finished.", status.TaskId))
}

// attemptRepair attempts to repair a failed pipeline step
func (s *TaskflowServiceImpl) attemptRepair(ctx context.Context, planningTask domain.Task, pipeline *TaskPipeline, status *domain.AutonomousTaskStatus, attempt int) error {
	s.updateAutonomousTaskStatus(status.TaskId, "running", "Execution failed. Attempting self-correction...", 80.0+float64(attempt)*5)

	failedStep := s.findFailedStep(pipeline)
	if failedStep == nil {
		return fmt.Errorf("pipeline failed but no failed step found. Final status: %s", pipeline.Status)
	}

	s.log.Info(fmt.Sprintf("[Task %s] Found failed step: %s. Attempting to repair.", status.TaskId, failedStep.Name))
	repairPipeline, err := s.createRepairPipeline(ctx, planningTask, failedStep)
	if err != nil {
		return fmt.Errorf("failed to create repair pipeline: %w", err)
	}

	if err := s.planner.ExecutePipeline(ctx, repairPipeline); err != nil {
		return fmt.Errorf("repair pipeline execution failed: %w", err)
	}
	if repairPipeline.Status != PipelineStatusCompleted {
		return fmt.Errorf("repair pipeline did not complete successfully. Final status: %s", repairPipeline.Status)
	}
	s.log.Info(fmt.Sprintf("[Task %s] Repair successful. Retrying main pipeline.", status.TaskId))
	return nil
}

// executeAutonomousTask выполняет автономную задачу с циклом самоисправления
func (s *TaskflowServiceImpl) executeAutonomousTask(ctx context.Context, request domain.AutonomousTaskRequest, status *domain.AutonomousTaskStatus) error {
	basePipeline, planningTask, err := s.planAutonomousTask(ctx, request, status)
	if err != nil {
		return err
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		s.log.Info(fmt.Sprintf("[Task %s] Starting pipeline execution, attempt %d/%d.", status.TaskId, i+1, maxRetries))
		currentPipeline := *basePipeline

		if err := s.planner.ExecutePipeline(ctx, &currentPipeline); err == nil && currentPipeline.Status == PipelineStatusCompleted {
			s.finishAutonomousTask(request, status)
			return nil
		}

		s.log.Error(fmt.Sprintf("[Task %s] Pipeline execution failed", status.TaskId))
		if err := s.attemptRepair(ctx, planningTask, &currentPipeline, status, i); err != nil {
			return err
		}
	}
	return fmt.Errorf("task failed after %d repair attempts", maxRetries)
}

// findFailedStep находит первый проваленный шаг в пайплайне
func (s *TaskflowServiceImpl) findFailedStep(pipeline *TaskPipeline) *TaskPipelineStep {
	for _, step := range pipeline.Steps {
		if step.Status == StepStatusFailed {
			return step
		}
	}
	return nil
}

// createRepairPipeline создает пайплайн для исправления ошибки
func (s *TaskflowServiceImpl) createRepairPipeline(ctx context.Context, task domain.Task, failedStep *TaskPipelineStep) (*TaskPipeline, error) {
	repairStep := &TaskPipelineStep{
		ID:        fmt.Sprintf("%s-repair-%d", task.ID, time.Now().Unix()),
		Name:      fmt.Sprintf("Repair for: %s", failedStep.Name),
		Type:      StepTypeRepair,
		Status:    StepStatusPending,
		Priority:  1,
		DependsOn: []string{},
		Config: map[string]interface{}{
			"task_id":           task.ID,
			"project_path":      task.Metadata["project_path"],
			"error_output":      failedStep.Error, // Передаем ошибку из проваленного шага
			"repair_strategies": []string{"auto_fix"},
			"max_attempts":      1,
		},
	}

	repairPipeline := &TaskPipeline{
		TaskID:    fmt.Sprintf("%s-repair", task.ID),
		Steps:     []*TaskPipelineStep{repairStep},
		Status:    PipelineStatusPending,
		CreatedAt: time.Now(),
		Policy: &PipelinePolicy{ // Простая политика для исправления
			FailFast: true,
		},
	}

	return repairPipeline, nil
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
func (s *TaskflowServiceImpl) notifyTaskFailure(taskID, message string) {
	// This could emit an event to the frontend via the event bus
	s.log.Error(fmt.Sprintf("Task %s failed: %s", taskID, message))
}

// updateAutonomousTaskStatus обновляет статус автономной задачи
func (s *TaskflowServiceImpl) updateAutonomousTaskStatus(taskID, status, message string, progress float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskStatus, exists := s.statuses[taskID]
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

// buildContextForTask собирает контекст для задачи, включая все файлы проекта
func (s *TaskflowServiceImpl) buildContextForTask(ctx context.Context, request domain.AutonomousTaskRequest) (map[string]interface{}, error) {
	s.log.Info(fmt.Sprintf("[Task %s] Building context for autonomous task...", request.Task))

	// Получаем все файлы в проекте
	allFiles, err := s.gitRepo.GetAllFiles(request.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get all files: %w", err)
	}

	// Формируем context pack
	contextPack := map[string]interface{}{
		"project_path":     request.ProjectPath,
		"all_files":        allFiles,
		"task_description": request.Task,
	}

	s.log.Info(fmt.Sprintf("[Task %s] Context built with %d files.", request.Task, len(allFiles)))
	return contextPack, nil
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
