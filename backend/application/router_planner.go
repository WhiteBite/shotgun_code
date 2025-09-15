package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"sort"
	"strings"
	"time"
)

// RouterPlannerService предоставляет эвристический планировщик TPL (DAG) без LLM
type RouterPlannerService struct {
	log domain.Logger
}

// NewRouterPlannerService создает новый сервис планировщика
func NewRouterPlannerService(log domain.Logger) *RouterPlannerService {
	return &RouterPlannerService{
		log: log,
	}
}

// TaskPipelineStep представляет шаг в пайплайне задачи
type TaskPipelineStep struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Type        TaskPipelineStepType    `json:"type"`
	Status      TaskPipelineStepStatus  `json:"status"`
	Priority    int                     `json:"priority"`
	DependsOn   []string                `json:"depends_on"`
	Config      map[string]interface{}  `json:"config"`
	StartedAt   *time.Time              `json:"started_at,omitempty"`
	CompletedAt *time.Time              `json:"completed_at,omitempty"`
	Duration    time.Duration           `json:"duration,omitempty"`
	Error       string                  `json:"error,omitempty"`
	Result      *TaskPipelineStepResult `json:"result,omitempty"`
}

// TaskPipelineStepType определяет тип шага пайплайна
type TaskPipelineStepType string

const (
	StepTypeRetrieve TaskPipelineStepType = "retrieve"
	StepTypeASTSynth TaskPipelineStepType = "ast_synth"
	StepTypeCompile  TaskPipelineStepType = "compile"
	StepTypeTest     TaskPipelineStepType = "test"
	StepTypeStatic   TaskPipelineStepType = "static"
	StepTypeRepair   TaskPipelineStepType = "repair"
	StepTypeFormat   TaskPipelineStepType = "format"
	StepTypeValidate TaskPipelineStepType = "validate"
)

// TaskPipelineStepStatus определяет статус шага пайплайна
type TaskPipelineStepStatus string

const (
	StepStatusPending   TaskPipelineStepStatus = "pending"
	StepStatusRunning   TaskPipelineStepStatus = "running"
	StepStatusCompleted TaskPipelineStepStatus = "completed"
	StepStatusFailed    TaskPipelineStepStatus = "failed"
	StepStatusSkipped   TaskPipelineStepStatus = "skipped"
)

// TaskPipelineStepResult содержит результат выполнения шага
type TaskPipelineStepResult struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Metrics     map[string]float64     `json:"metrics,omitempty"`
	Artifacts   []string               `json:"artifacts,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// TaskPipeline представляет полный пайплайн для задачи
type TaskPipeline struct {
	TaskID      string              `json:"task_id"`
	Steps       []*TaskPipelineStep `json:"steps"`
	Status      TaskPipelineStatus  `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	StartedAt   *time.Time          `json:"started_at,omitempty"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	Duration    time.Duration       `json:"duration,omitempty"`
	Error       string              `json:"error,omitempty"`
	Policy      *PipelinePolicy     `json:"policy"`
}

// TaskPipelineStatus определяет статус пайплайна
type TaskPipelineStatus string

const (
	PipelineStatusPending   TaskPipelineStatus = "pending"
	PipelineStatusRunning   TaskPipelineStatus = "running"
	PipelineStatusCompleted TaskPipelineStatus = "completed"
	PipelineStatusFailed    TaskPipelineStatus = "failed"
)

// PipelinePolicy определяет политику выполнения пайплайна
type PipelinePolicy struct {
	EnableRetrieve bool          `json:"enable_retrieve"`
	EnableASTSynth bool          `json:"enable_ast_synth"`
	EnableCompile  bool          `json:"enable_compile"`
	EnableTest     bool          `json:"enable_test"`
	EnableStatic   bool          `json:"enable_static"`
	EnableRepair   bool          `json:"enable_repair"`
	EnableFormat   bool          `json:"enable_format"`
	EnableValidate bool          `json:"enable_validate"`
	FailFast       bool          `json:"fail_fast"`
	RetryFailed    bool          `json:"retry_failed"`
	MaxRetries     int           `json:"max_retries"`
	ParallelSteps  bool          `json:"parallel_steps"`
	Timeout        time.Duration `json:"timeout"`
}

// CreatePipeline создает пайплайн для задачи
func (r *RouterPlannerService) CreatePipeline(ctx context.Context, task domain.Task) (*TaskPipeline, error) {
	r.log.Info(fmt.Sprintf("Creating pipeline for task: %s", task.ID))

	// Определяем политику на основе типа задачи
	policy := r.determinePolicy(task)

	// Создаем шаги пайплайна
	steps := r.createPipelineSteps(task, policy)

	// Сортируем шаги по приоритету и зависимостям
	r.sortPipelineSteps(steps)

	pipeline := &TaskPipeline{
		TaskID:    task.ID,
		Steps:     steps,
		Status:    PipelineStatusPending,
		CreatedAt: time.Now(),
		Policy:    policy,
	}

	r.log.Info(fmt.Sprintf("Created pipeline with %d steps for task: %s", len(steps), task.ID))
	return pipeline, nil
}

// ExecutePipeline выполняет пайплайн
func (r *RouterPlannerService) ExecutePipeline(ctx context.Context, pipeline *TaskPipeline) error {
	r.log.Info(fmt.Sprintf("Executing pipeline for task: %s", pipeline.TaskID))

	now := time.Now()
	pipeline.StartedAt = &now
	pipeline.Status = PipelineStatusRunning

	// Выполняем шаги последовательно или параллельно
	if pipeline.Policy.ParallelSteps {
		return r.executePipelineParallel(ctx, pipeline)
	}
	return r.executePipelineSequential(ctx, pipeline)
}

// determinePolicy определяет политику выполнения на основе задачи
func (r *RouterPlannerService) determinePolicy(task domain.Task) *PipelinePolicy {
	policy := &PipelinePolicy{
		EnableRetrieve: true,
		EnableASTSynth: true,
		EnableCompile:  true,
		EnableTest:     true,
		EnableStatic:   true,
		EnableRepair:   true,
		EnableFormat:   true,
		EnableValidate: true,
		FailFast:       true,
		RetryFailed:    true,
		MaxRetries:     3,
		ParallelSteps:  false,
		Timeout:        30 * time.Minute,
	}

	// Настройки на основе типа задачи
	switch {
	case strings.Contains(task.ID, "ark-100"):
		// Правила и конфигурация - только валидация
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-110"):
		// Индексатор - retrieve + ast_synth
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-120"):
		// Context Pack - retrieve + ast_synth + validate
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-130"):
		// Apply Engine - полный пайплайн
		policy.EnableRepair = true
	case strings.Contains(task.ID, "ark-140"):
		// Derived Diff - retrieve + ast_synth + validate
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-150"):
		// Verification Pipeline - полный пайплайн
		policy.EnableRepair = true
	case strings.Contains(task.ID, "ark-160"):
		// Targeted Tests - retrieve + test
		policy.EnableASTSynth = false
		policy.EnableCompile = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-170"):
		// Static Analyzers - retrieve + static
		policy.EnableASTSynth = false
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-180"):
		// Router Planner - только планирование
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
		policy.EnableFormat = false
	case strings.Contains(task.ID, "ark-999"):
		// Final Gate - полный пайплайн с валидацией
		policy.EnableRepair = true
		policy.FailFast = true
	}

	return policy
}

// createPipelineSteps создает шаги пайплайна
func (r *RouterPlannerService) createPipelineSteps(task domain.Task, policy *PipelinePolicy) []*TaskPipelineStep {
	var steps []*TaskPipelineStep
	stepID := 1

	// Retrieve step
	if policy.EnableRetrieve {
		steps = append(steps, &TaskPipelineStep{
			ID:        fmt.Sprintf("%s-step-%d", task.ID, stepID),
			Name:      "Retrieve Context",
			Type:      StepTypeRetrieve,
			Status:    StepStatusPending,
			Priority:  1,
			DependsOn: []string{},
			Config: map[string]interface{}{
				"task_id":      task.ID,
				"step_file":    task.StepFile,
				"budgets":      task.Budgets,
				"dependencies": task.DependsOn,
			},
		})
		stepID++
	}

	// AST Synthesis step
	if policy.EnableASTSynth {
		dependsOn := []string{}
		if policy.EnableRetrieve {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, 1))
		}
		steps = append(steps, &TaskPipelineStep{
			ID:        fmt.Sprintf("%s-step-%d", task.ID, stepID),
			Name:      "AST Synthesis",
			Type:      StepTypeASTSynth,
			Status:    StepStatusPending,
			Priority:  2,
			DependsOn: dependsOn,
			Config: map[string]interface{}{
				"task_id":      task.ID,
				"symbol_graph": true,
				"context_pack": true,
			},
		})
		stepID++
	}

	// Compile step
	if policy.EnableCompile {
		dependsOn := []string{}
		if policy.EnableASTSynth {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableRetrieve {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, 1))
		}
		steps = append(steps, &TaskPipelineStep{
			ID:        fmt.Sprintf("%s-step-%d", task.ID, stepID),
			Name:      "Compile",
			Type:      StepTypeCompile,
			Status:    StepStatusPending,
			Priority:  3,
			DependsOn: dependsOn,
			Config: map[string]interface{}{
				"task_id":    task.ID,
				"language":   "go",
				"build_mode": "debug",
			},
		})
		stepID++
	}

	// Test step
	if policy.EnableTest {
		dependsOn := []string{}
		if policy.EnableCompile {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableASTSynth {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableRetrieve {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, 1))
		}
		steps = append(steps, &TaskPipelineStep{
			ID:        fmt.Sprintf("%s-step-%d", task.ID, stepID),
			Name:      "Test",
			Type:      StepTypeTest,
			Status:    StepStatusPending,
			Priority:  4,
			DependsOn: dependsOn,
			Config: map[string]interface{}{
				"task_id":   task.ID,
				"test_mode": "targeted",
				"coverage":  true,
			},
		})
		stepID++
	}

	// Static Analysis step
	if policy.EnableStatic {
		dependsOn := []string{}
		if policy.EnableTest {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableCompile {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableASTSynth {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableRetrieve {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, 1))
		}
		steps = append(steps, &TaskPipelineStep{
			ID:        fmt.Sprintf("%s-step-%d", task.ID, stepID),
			Name:      "Static Analysis",
			Type:      StepTypeStatic,
			Status:    StepStatusPending,
			Priority:  5,
			DependsOn: dependsOn,
			Config: map[string]interface{}{
				"task_id":       task.ID,
				"analyzers":     []string{"staticcheck", "go vet", "eslint"},
				"fail_on_error": false,
			},
		})
		stepID++
	}

	// Format step
	if policy.EnableFormat {
		dependsOn := []string{}
		if policy.EnableStatic {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableTest {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableCompile {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableASTSynth {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableRetrieve {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, 1))
		}
		steps = append(steps, &TaskPipelineStep{
			ID:        fmt.Sprintf("%s-step-%d", task.ID, stepID),
			Name:      "Format",
			Type:      StepTypeFormat,
			Status:    StepStatusPending,
			Priority:  6,
			DependsOn: dependsOn,
			Config: map[string]interface{}{
				"task_id":    task.ID,
				"formatters": []string{"gofmt", "goimports", "prettier"},
			},
		})
		stepID++
	}

	// Validate step
	if policy.EnableValidate {
		dependsOn := []string{}
		if policy.EnableFormat {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableStatic {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableTest {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableCompile {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableASTSynth {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableRetrieve {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, 1))
		}
		steps = append(steps, &TaskPipelineStep{
			ID:        fmt.Sprintf("%s-step-%d", task.ID, stepID),
			Name:      "Validate",
			Type:      StepTypeValidate,
			Status:    StepStatusPending,
			Priority:  7,
			DependsOn: dependsOn,
			Config: map[string]interface{}{
				"task_id":          task.ID,
				"validation_rules": []string{"syntax", "semantics", "policy"},
			},
		})
		stepID++
	}

	// Repair step (если включен и есть ошибки)
	if policy.EnableRepair {
		dependsOn := []string{}
		if policy.EnableValidate {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableFormat {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableStatic {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableTest {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableCompile {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableASTSynth {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, stepID-1))
		} else if policy.EnableRetrieve {
			dependsOn = append(dependsOn, fmt.Sprintf("%s-step-%d", task.ID, 1))
		}
		steps = append(steps, &TaskPipelineStep{
			ID:        fmt.Sprintf("%s-step-%d", task.ID, stepID),
			Name:      "Repair",
			Type:      StepTypeRepair,
			Status:    StepStatusPending,
			Priority:  8,
			DependsOn: dependsOn,
			Config: map[string]interface{}{
				"task_id":           task.ID,
				"repair_strategies": []string{"auto_fix", "suggestions", "rollback"},
				"max_attempts":      policy.MaxRetries,
			},
		})
	}

	return steps
}

// sortPipelineSteps сортирует шаги по приоритету и зависимостям
func (r *RouterPlannerService) sortPipelineSteps(steps []*TaskPipelineStep) {
	sort.Slice(steps, func(i, j int) bool {
		// Сначала по приоритету
		if steps[i].Priority != steps[j].Priority {
			return steps[i].Priority < steps[j].Priority
		}
		// Затем по количеству зависимостей
		return len(steps[i].DependsOn) < len(steps[j].DependsOn)
	})
}

// executePipelineSequential выполняет пайплайн последовательно
func (r *RouterPlannerService) executePipelineSequential(ctx context.Context, pipeline *TaskPipeline) error {
	for _, step := range pipeline.Steps {
		if err := r.executeStep(ctx, step); err != nil {
			if pipeline.Policy.FailFast {
				pipeline.Status = PipelineStatusFailed
				pipeline.Error = err.Error()
				now := time.Now()
				pipeline.CompletedAt = &now
				pipeline.Duration = now.Sub(*pipeline.StartedAt)
				return fmt.Errorf("step %s failed: %w", step.ID, err)
			}
			// Продолжаем выполнение других шагов
		}
	}

	pipeline.Status = PipelineStatusCompleted
	now := time.Now()
	pipeline.CompletedAt = &now
	pipeline.Duration = now.Sub(*pipeline.StartedAt)

	r.log.Info(fmt.Sprintf("Pipeline completed for task: %s", pipeline.TaskID))
	return nil
}

// executePipelineParallel выполняет пайплайн параллельно
func (r *RouterPlannerService) executePipelineParallel(ctx context.Context, pipeline *TaskPipeline) error {
	// Упрощенная реализация параллельного выполнения
	// В реальной реализации нужно учитывать зависимости между шагами
	return r.executePipelineSequential(ctx, pipeline)
}

// executeStep выполняет один шаг пайплайна
func (r *RouterPlannerService) executeStep(ctx context.Context, step *TaskPipelineStep) error {
	r.log.Info(fmt.Sprintf("Executing step: %s (%s)", step.Name, step.ID))

	now := time.Now()
	step.StartedAt = &now
	step.Status = StepStatusRunning

	// Выполняем шаг в зависимости от типа
	var err error
	switch step.Type {
	case StepTypeRetrieve:
		err = r.executeRetrieveStep(ctx, step)
	case StepTypeASTSynth:
		err = r.executeASTSynthStep(ctx, step)
	case StepTypeCompile:
		err = r.executeCompileStep(ctx, step)
	case StepTypeTest:
		err = r.executeTestStep(ctx, step)
	case StepTypeStatic:
		err = r.executeStaticStep(ctx, step)
	case StepTypeFormat:
		err = r.executeFormatStep(ctx, step)
	case StepTypeValidate:
		err = r.executeValidateStep(ctx, step)
	case StepTypeRepair:
		err = r.executeRepairStep(ctx, step)
	default:
		err = fmt.Errorf("unknown step type: %s", step.Type)
	}

	// Обновляем статус шага
	completedAt := time.Now()
	step.CompletedAt = &completedAt
	step.Duration = completedAt.Sub(*step.StartedAt)

	if err != nil {
		step.Status = StepStatusFailed
		step.Error = err.Error()
		r.log.Error(fmt.Sprintf("Step %s failed: %v", step.ID, err))
	} else {
		step.Status = StepStatusCompleted
		step.Result = &TaskPipelineStepResult{
			Success: true,
			Message: fmt.Sprintf("Step %s completed successfully", step.Name),
		}
		r.log.Info(fmt.Sprintf("Step %s completed successfully", step.ID))
	}

	return err
}

// executeRetrieveStep выполняет шаг извлечения контекста
func (r *RouterPlannerService) executeRetrieveStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага retrieve
	time.Sleep(100 * time.Millisecond)
	return nil
}

// executeASTSynthStep выполняет шаг синтеза AST
func (r *RouterPlannerService) executeASTSynthStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага ast_synth
	time.Sleep(200 * time.Millisecond)
	return nil
}

// executeCompileStep выполняет шаг компиляции
func (r *RouterPlannerService) executeCompileStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага compile
	time.Sleep(300 * time.Millisecond)
	return nil
}

// executeTestStep выполняет шаг тестирования
func (r *RouterPlannerService) executeTestStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага test
	time.Sleep(400 * time.Millisecond)
	return nil
}

// executeStaticStep выполняет шаг статического анализа
func (r *RouterPlannerService) executeStaticStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага static
	time.Sleep(250 * time.Millisecond)
	return nil
}

// executeFormatStep выполняет шаг форматирования
func (r *RouterPlannerService) executeFormatStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага format
	time.Sleep(150 * time.Millisecond)
	return nil
}

// executeValidateStep выполняет шаг валидации
func (r *RouterPlannerService) executeValidateStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага validate
	time.Sleep(100 * time.Millisecond)
	return nil
}

// executeRepairStep выполняет шаг исправления
func (r *RouterPlannerService) executeRepairStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага repair
	time.Sleep(500 * time.Millisecond)
	return nil
}

// GetPipelineStatus возвращает статус пайплайна
func (r *RouterPlannerService) GetPipelineStatus(pipeline *TaskPipeline) map[string]interface{} {
	completed := 0
	failed := 0
	pending := 0
	running := 0

	for _, step := range pipeline.Steps {
		switch step.Status {
		case StepStatusCompleted:
			completed++
		case StepStatusFailed:
			failed++
		case StepStatusPending:
			pending++
		case StepStatusRunning:
			running++
		}
	}

	return map[string]interface{}{
		"task_id":     pipeline.TaskID,
		"status":      pipeline.Status,
		"total_steps": len(pipeline.Steps),
		"completed":   completed,
		"failed":      failed,
		"pending":     pending,
		"running":     running,
		"progress":    float64(completed) / float64(len(pipeline.Steps)),
		"duration":    pipeline.Duration,
		"error":       pipeline.Error,
	}
}
